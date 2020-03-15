// Copyright 2020-present Open Networking Foundation.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package _map

import (
	"bytes"
	"io"

	rc "github.com/atomix/redis-proxy/pkg/redis_client"

	"github.com/atomix/go-framework/pkg/atomix/node"
	"github.com/atomix/go-framework/pkg/atomix/service"
	"github.com/atomix/go-framework/pkg/atomix/stream"
	"github.com/atomix/go-framework/pkg/atomix/util"
	"github.com/golang/protobuf/proto"
	"github.com/gomodule/redigo/redis"
)

func init() {
	node.RegisterService(service.ServiceType_MAP, newService)
}

var redisServer = "localhost:6379"

// newService returns a new Service
func newService(scheduler service.Scheduler, context service.Context) service.Service {
	pool := rc.NewPool(redisServer)
	service := &Service{
		ManagedService: service.NewManagedService(service.ServiceType_MAP, scheduler, context),
		mapName:        "entries",
		redisPool:      pool,
		timers:         make(map[string]service.Timer),
		listeners:      make(map[uint64]map[uint64]listener),
	}
	service.init()
	return service
}

// Service is a state machine for a map primitive
type Service struct {
	*service.ManagedService
	redisPool *redis.Pool
	mapName   string
	timers    map[string]service.Timer
	listeners map[uint64]map[uint64]listener
}

// init initializes the map service
func (m *Service) init() {
	m.Executor.RegisterUnaryOperation(opPut, m.Put)
	m.Executor.RegisterUnaryOperation(opReplace, m.Replace)
	m.Executor.RegisterUnaryOperation(opRemove, m.Remove)
	m.Executor.RegisterUnaryOperation(opGet, m.Get)
	m.Executor.RegisterUnaryOperation(opExists, m.Exists)
	m.Executor.RegisterUnaryOperation(opSize, m.Size)
	m.Executor.RegisterUnaryOperation(opClear, m.Clear)
	m.Executor.RegisterStreamOperation(opEvents, m.Events)
	m.Executor.RegisterStreamOperation(opEntries, m.Entries)
}

// Backup takes a snapshot of the service
func (m *Service) Backup(writer io.Writer) error {
	conn := m.redisPool.Get()
	defer conn.Close()
	listeners := make([]*Listener, 0)
	for sessionID, sessionListeners := range m.listeners {
		for streamID, sessionListener := range sessionListeners {
			listeners = append(listeners, &Listener{
				SessionId: sessionID,
				StreamId:  streamID,
				Key:       sessionListener.key,
			})
		}
	}

	if err := util.WriteVarInt(writer, len(listeners)); err != nil {
		return err
	}
	if err := util.WriteSlice(writer, listeners, proto.Marshal); err != nil {
		return err
	}

	entries := m.getMapEntries()

	return util.WriteMap(writer, entries, func(key string, value *MapEntryValue) ([]byte, error) {
		return proto.Marshal(&MapEntry{
			Key:   key,
			Value: value,
		})
	})
}

// Restore restores the service from a snapshot
func (m *Service) Restore(reader io.Reader) error {
	length, err := util.ReadVarInt(reader)
	if err != nil {
		return err
	}

	listeners := make([]*Listener, length)
	err = util.ReadSlice(reader, listeners, func(data []byte) (*Listener, error) {
		listener := &Listener{}
		if err := proto.Unmarshal(data, listener); err != nil {
			return nil, err
		}
		return listener, nil
	})
	if err != nil {
		return err
	}

	m.listeners = make(map[uint64]map[uint64]listener)
	for _, snapshotListener := range listeners {
		sessionListeners, ok := m.listeners[snapshotListener.SessionId]
		if !ok {
			sessionListeners = make(map[uint64]listener)
			m.listeners[snapshotListener.SessionId] = sessionListeners
		}
		sessionListeners[snapshotListener.StreamId] = listener{
			key:    snapshotListener.Key,
			stream: m.Sessions()[snapshotListener.SessionId].Stream(snapshotListener.StreamId),
		}
	}

	entries := make(map[string]*MapEntryValue)
	err = util.ReadMap(reader, entries, func(data []byte) (string, *MapEntryValue, error) {
		entry := &MapEntry{}
		if err := proto.Unmarshal(data, entry); err != nil {
			return "", nil, err
		}
		return entry.Key, entry.Value, nil
	})
	if err != nil {
		return err
	}
	//m.entries = entries
	return nil
}

// Put puts a key/value pair in the map
func (m *Service) Put(value []byte) ([]byte, error) {
	m.getMapEntries()
	conn := m.redisPool.Get()
	defer conn.Close()
	request := &PutRequest{}
	if err := proto.Unmarshal(value, request); err != nil {
		return nil, err
	}

	//oldValue := m.entries[request.Key]
	mapOldValue, err := conn.Do(HGET, m.mapName, request.Key)

	if err != nil {
		return nil, err
	}

	if mapOldValue == nil {
		// If the version is positive then reject the request.
		if !request.IfEmpty && request.Version > 0 {
			return proto.Marshal(&PutResponse{
				Status: UpdateStatus_PRECONDITION_FAILED,
			})
		}

		// Create a new entry value and set it in the map.
		newValue := &MapEntryValue{
			Value:   request.Value,
			Version: m.Context.Index(),
			TTL:     request.TTL,
			Created: m.Context.Timestamp(),
			Updated: m.Context.Timestamp(),
		}

		newValueBytes, err := newValue.Marshal()
		if err != nil {
			return nil, err
		}
		//m.entries[request.Key] = newValue
		_, err = conn.Do(HSET, m.mapName, request.Key, newValueBytes)
		if err != nil {
			return nil, err
		}
		// Schedule the timeout for the value if necessary.
		m.scheduleTTL(request.Key, newValue)

		// Publish an event to listener streams.
		m.sendEvent(&ListenResponse{
			Type:    ListenResponse_INSERTED,
			Key:     request.Key,
			Value:   newValue.Value,
			Version: newValue.Version,
			Created: newValue.Created,
			Updated: newValue.Updated,
		})

		return proto.Marshal(&PutResponse{
			Status: UpdateStatus_OK,
		})
	}

	oldValue := MapEntryValue{}
	err = oldValue.Unmarshal(mapOldValue.([]byte))
	if err != nil {
		return nil, err
	}

	// If the version is -1 then reject the request.
	// If the version is positive then compare the version to the current version.
	if request.IfEmpty || (!request.IfEmpty && request.Version > 0 && request.Version != oldValue.Version) {
		return proto.Marshal(&PutResponse{
			Status:          UpdateStatus_PRECONDITION_FAILED,
			PreviousValue:   oldValue.Value,
			PreviousVersion: oldValue.Version,
		})
	}

	// If the value is equal to the current value, return a no-op.
	if bytes.Equal(oldValue.Value, request.Value) {
		return proto.Marshal(&PutResponse{
			Status:          UpdateStatus_NOOP,
			PreviousValue:   oldValue.Value,
			PreviousVersion: oldValue.Version,
		})
	}

	// Create a new entry value and set it in the map.
	newValue := &MapEntryValue{
		Value:   request.Value,
		Version: m.Context.Index(),
		TTL:     request.TTL,
		Created: oldValue.Created,
		Updated: m.Context.Timestamp(),
	}

	newValueBytes, _ := newValue.Marshal()
	_, err = conn.Do(HSET, m.mapName, request.Key, newValueBytes)
	if err != nil {
		return nil, err
	}
	//m.entries[request.Key] = newValue

	// Schedule the timeout for the value if necessary.
	m.scheduleTTL(request.Key, newValue)

	// Publish an event to listener streams.
	m.sendEvent(&ListenResponse{
		Type:    ListenResponse_UPDATED,
		Key:     request.Key,
		Value:   newValue.Value,
		Version: newValue.Version,
		Created: newValue.Created,
		Updated: newValue.Updated,
	})

	return proto.Marshal(&PutResponse{
		Status:          UpdateStatus_OK,
		PreviousValue:   oldValue.Value,
		PreviousVersion: oldValue.Version,
		Created:         newValue.Created,
		Updated:         newValue.Updated,
	})

}

// Replace replaces a key/value pair in the map
func (m *Service) Replace(value []byte) ([]byte, error) {
	conn := m.redisPool.Get()
	defer conn.Close()
	request := &ReplaceRequest{}
	if err := proto.Unmarshal(value, request); err != nil {
		return nil, err
	}

	oldValue, err, exist := m.getMapEntryValue(request.Key)
	if err != nil {
		return nil, err
	}
	if !exist {
		return proto.Marshal(&ReplaceResponse{
			Status: UpdateStatus_PRECONDITION_FAILED,
		})
	}

	// If the version was specified and does not match the entry version, fail the replace.
	if request.PreviousVersion != 0 && request.PreviousVersion != oldValue.Version {
		return proto.Marshal(&ReplaceResponse{
			Status: UpdateStatus_PRECONDITION_FAILED,
		})
	}

	// If the value was specified and does not match the entry value, fail the replace.
	if len(request.PreviousValue) != 0 && bytes.Equal(request.PreviousValue, oldValue.Value) {
		return proto.Marshal(&ReplaceResponse{
			Status: UpdateStatus_PRECONDITION_FAILED,
		})
	}

	// If we've made it this far, update the entry.
	// Create a new entry value and set it in the map.
	newValue := &MapEntryValue{
		Value:   request.NewValue,
		Version: m.Context.Index(),
		TTL:     request.TTL,
		Created: oldValue.Created,
		Updated: m.Context.Timestamp(),
	}

	newValueBytes, _ := newValue.Marshal()
	_, err = conn.Do(HSET, m.mapName, request.Key, newValueBytes)
	if err != nil {
		return nil, err
	}

	// Schedule the timeout for the value if necessary.
	m.scheduleTTL(request.Key, newValue)

	// Publish an event to listener streams.
	m.sendEvent(&ListenResponse{
		Type:    ListenResponse_UPDATED,
		Key:     request.Key,
		Value:   newValue.Value,
		Version: newValue.Version,
		Created: newValue.Created,
		Updated: newValue.Updated,
	})

	return proto.Marshal(&ReplaceResponse{
		Status:          UpdateStatus_OK,
		PreviousValue:   oldValue.Value,
		PreviousVersion: oldValue.Version,
		NewVersion:      newValue.Version,
		Created:         newValue.Created,
		Updated:         newValue.Updated,
	})
}

// Remove removes a key/value pair from the map
func (m *Service) Remove(bytes []byte) ([]byte, error) {
	conn := m.redisPool.Get()
	defer conn.Close()

	request := &RemoveRequest{}
	if err := proto.Unmarshal(bytes, request); err != nil {
		return nil, err
	}

	value, err, exist := m.getMapEntryValue(request.Key)
	if err != nil {
		return nil, err
	}
	if !exist {
		return proto.Marshal(&RemoveResponse{
			Status: UpdateStatus_NOOP,
		})
	}

	// If the request version is set, verify that the request version matches the entry version.
	if request.Version > 0 && request.Version != value.Version {
		return proto.Marshal(&RemoveResponse{
			Status: UpdateStatus_PRECONDITION_FAILED,
		})
	}

	// Delete the entry from the map.
	_, err = conn.Do(HDEL, m.mapName, request.Key)
	if err != nil {
		return nil, err
	}

	// Cancel any TTLs.
	m.cancelTTL(request.Key)

	// Publish an event to listener streams.
	m.sendEvent(&ListenResponse{
		Type:    ListenResponse_REMOVED,
		Key:     request.Key,
		Value:   value.Value,
		Version: value.Version,
		Created: value.Created,
		Updated: value.Updated,
	})

	return proto.Marshal(&RemoveResponse{
		Status:          UpdateStatus_OK,
		PreviousValue:   value.Value,
		PreviousVersion: value.Version,
		Created:         value.Created,
		Updated:         value.Updated,
	})
}

// Get gets a value from the map
func (m *Service) Get(bytes []byte) ([]byte, error) {
	request := &GetRequest{}
	if err := proto.Unmarshal(bytes, request); err != nil {
		return nil, err
	}

	value, err, exist := m.getMapEntryValue(request.Key)
	if err != nil {
		return nil, err
	}

	if !exist {
		return proto.Marshal(&GetResponse{})
	}

	return proto.Marshal(&GetResponse{
		Value:   value.Value,
		Version: value.Version,
		Created: value.Created,
		Updated: value.Updated,
	})
}

// Exists checks if the map contains a key
func (m *Service) Exists(bytes []byte) ([]byte, error) {
	conn := m.redisPool.Get()
	defer conn.Close()
	request := &ContainsKeyRequest{}
	if err := proto.Unmarshal(bytes, request); err != nil {
		return nil, err
	}

	exist, err := redis.Bool(conn.Do(HEXISTS, request.Key))
	if err != nil {
		return nil, err
	}
	return proto.Marshal(&ContainsKeyResponse{
		ContainsKey: exist,
	})
}

// Size returns the size of the map
func (m *Service) Size(bytes []byte) ([]byte, error) {
	conn := m.redisPool.Get()
	defer conn.Close()
	mapLen, _ := redis.Int(conn.Do(HLEN, m.mapName))
	return proto.Marshal(&SizeResponse{
		Size_: int32(mapLen),
	})
}

// Clear removes all entries from the map
func (m *Service) Clear(value []byte) ([]byte, error) {
	conn := m.redisPool.Get()
	defer conn.Close()

	keys, err := redis.Strings(conn.Do(HKEYS, m.mapName))
	if err != nil {
		return nil, err
	}

	for _, keyValue := range keys {
		_, err := conn.Do(HDEL, m.mapName, keyValue)
		if err != nil {
			return nil, err
		}
	}
	return proto.Marshal(&ClearResponse{})
}

// Events sends change events to the client
func (m *Service) Events(bytes []byte, stream stream.WriteStream) {
	conn := m.redisPool.Get()
	defer conn.Close()
	request := &ListenRequest{}
	if err := proto.Unmarshal(bytes, request); err != nil {
		stream.Error(err)
		stream.Close()
		return
	}

	// Create and populate the listener
	l := listener{
		key:    request.Key,
		stream: stream,
	}
	listeners, ok := m.listeners[m.Session().ID]
	if !ok {
		listeners = make(map[uint64]listener)
		m.listeners[m.Session().ID] = listeners
	}
	listeners[m.Session().StreamID()] = l

	entries := m.getMapEntries()

	// If replay was requested, send existing entries
	if request.Replay {
		for key, value := range entries {
			stream.Result(proto.Marshal(&ListenResponse{
				Type:    ListenResponse_NONE,
				Key:     key,
				Value:   value.Value,
				Version: value.Version,
				Created: value.Created,
				Updated: value.Updated,
			}))
		}
	}
}

// Entries returns a stream of entries to the client
func (m *Service) Entries(value []byte, stream stream.WriteStream) {
	defer stream.Close()
	entries := m.getMapEntries()
	for key, entry := range entries {
		stream.Result(proto.Marshal(&EntriesResponse{
			Key:     key,
			Value:   entry.Value,
			Version: entry.Version,
			Created: entry.Created,
			Updated: entry.Updated,
		}))
	}
}

func (m *Service) scheduleTTL(key string, value *MapEntryValue) {
	entries := m.getMapEntries()
	m.cancelTTL(key)
	if value.TTL != nil {
		m.timers[key] = m.Scheduler.ScheduleOnce(value.Created.Add(*value.TTL).Sub(m.Context.Timestamp()), func() {
			delete(entries, key)
			m.sendEvent(&ListenResponse{
				Type:    ListenResponse_REMOVED,
				Key:     key,
				Value:   value.Value,
				Version: uint64(value.Version),
				Created: value.Created,
				Updated: value.Updated,
			})
		})
	}
}

func (m *Service) cancelTTL(key string) {
	timer, ok := m.timers[key]
	if ok {
		timer.Cancel()
	}
}

func (m *Service) sendEvent(event *ListenResponse) {
	bytes, _ := proto.Marshal(event)
	for sessionID, listeners := range m.listeners {
		session := m.SessionOf(sessionID)
		if session != nil {
			for _, listener := range listeners {
				if listener.key != "" {
					if event.Key == listener.key {
						listener.stream.Value(bytes)
					}
				} else {
					listener.stream.Value(bytes)
				}
			}
		}
	}
}

type listener struct {
	key    string
	stream stream.WriteStream
}