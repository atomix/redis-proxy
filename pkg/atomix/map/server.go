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

package _map //nolint:golint

import (
	"context"

	api "github.com/atomix/api/proto/atomix/map"
	"github.com/atomix/redis-proxy/pkg/atomix/service"
	"github.com/atomix/redis-proxy/pkg/server"
	"github.com/gomodule/redigo/redis"
	"github.com/onosproject/onos-lib-go/pkg/logging"
	"google.golang.org/grpc"
)

var log = logging.GetLogger("redis", "map")

// NewService returns a new Service
func NewService() (server.Service, error) {
	return &Service{}, nil
}

// Service is an implementation of map api service.
type Service struct {
	server.Service
}

// Register registers the map service
func (s Service) Register(r *grpc.Server) {
	api.RegisterMapServiceServer(r, newServer())
}

func newServer() api.MapServiceServer {
	return &Server{
		Server: &service.Server{},
	}
}

// Server is an implementation of MapServiceServer for the map primitive
type Server struct {
	api.MapServiceServer
	*service.Server
}

// Create opens a new session
func (s *Server) Create(ctx context.Context, request *api.CreateRequest) (*api.CreateResponse, error) {
	log.Info("Received CreateRequest %+v", request)
	s.DoCreateService(ctx)

	response := &api.CreateResponse{}
	log.Info("Sending CreateResponse %+v", response)
	return response, nil
}

// Close closes a session
/*func (s *Server) Close(ctx context.Context, request *api.CloseRequest) (*api.CloseResponse, error) {
	log.Info("Received CloseRequest %+v", request)
	if request.Delete {
	}


	response := &api.CloseResponse{}
	log.Info("Sending CloseResponse %+v", response)
	return response, nil
}*/

// Size gets the number of entries in the map
func (s *Server) Size(ctx context.Context, request *api.SizeRequest) (*api.SizeResponse, error) {
	log.Info("Received SizeRequest %+v", request)
	size, err := redis.Int(s.DoCommand(request.Header, HLEN, request.Header.Name.Name))
	if err != nil {
		return nil, err
	}

	response := &api.SizeResponse{
		Size_: int32(size),
	}
	log.Info("Sending SizeResponse %+v", response)
	return response, nil
}

// Exists checks whether the map contains a key
func (s *Server) Exists(ctx context.Context, request *api.ExistsRequest) (*api.ExistsResponse, error) {
	log.Info("Received ExistsRequest %+v", request)
	containsKey, err := redis.Bool(s.DoCommand(request.Header, HEXISTS, request.Header.Name.Name, request.Key))
	if err != nil {
		return nil, err
	}

	response := &api.ExistsResponse{
		ContainsKey: containsKey,
	}
	log.Info("Sending ExistsResponse %+v", response)
	return response, nil
}

// Put puts a key/value pair into the map
func (s *Server) Put(ctx context.Context, request *api.PutRequest) (*api.PutResponse, error) {
	log.Info("Received PutRequest %+v", request)

	log.Info(request.Header.SessionID)
	_, err := s.DoCommand(request.Header, HSET, request.Header.Name.Name, request.Key, request.Value)
	if err != nil {
		return nil, err
	}

	response := &api.PutResponse{}
	log.Info("Sending PutResponse %+v", response)
	return response, nil
}

// Replace replaces a key/value pair in the map
func (s *Server) Replace(ctx context.Context, request *api.ReplaceRequest) (*api.ReplaceResponse, error) {
	log.Info("Received ReplaceRequest %+v", request)
	_, err := s.DoCommand(request.Header, HSET, request.Header.Name.Name, request.Key, request.NewValue)
	if err != nil {
		return nil, err
	}

	response := &api.ReplaceResponse{}
	log.Info("Sending ReplaceResponse %+v", response)
	return response, nil
}

// Get gets the value of a key
func (s *Server) Get(ctx context.Context, request *api.GetRequest) (*api.GetResponse, error) {
	log.Info("Received GetRequest %+v", request)
	value, err := redis.Bytes(s.DoCommand(request.Header, HGET, request.Header.Name.Name, request.Key))
	if err != nil {
		return nil, err
	}

	if value == nil {
		return nil, err
	}

	response := &api.GetResponse{
		Value: value,
	}
	log.Info("Sending GetRequest %+v", response)
	return response, nil
}

// Remove removes a key from the map
func (s *Server) Remove(ctx context.Context, request *api.RemoveRequest) (*api.RemoveResponse, error) {
	log.Info("Received RemoveRequest %+v", request)
	_, err := s.DoCommand(request.Header, HDEL, request.Header.Name.Name, request.Key)
	if err != nil {
		return nil, err
	}

	response := &api.RemoveResponse{}
	log.Info("Sending RemoveRequest %+v", response)
	return response, nil
}

// Clear removes all keys from the map
func (s *Server) Clear(ctx context.Context, request *api.ClearRequest) (*api.ClearResponse, error) {
	log.Info("Received ClearRequest %+v", request)

	keys, err := redis.Strings(s.DoCommand(request.Header, HKEYS, request.Header.Name.Name))
	if err != nil {
		return nil, err
	}

	for _, key := range keys {
		_, err = s.DoCommand(request.Header, HDEL, request.Header.Name.Name, key)
		if err != nil {
			return nil, err
		}
	}

	response := &api.ClearResponse{}
	log.Info("Sending ClearResponse %+v", response)
	return response, nil
}

// Events listens for map change events
/*func (s *Server) Events(request *api.EventRequest, srv api.MapService_EventsServer) error {
	log.Tracef("Received EventRequest %+v", request)
	in, err := proto.Marshal(&ListenRequest{
		Replay: request.Replay,
		Key:    request.Key,
	})
	if err != nil {
		return err
	}

	stream := streams.NewBufferedStream()
	if err := s.DoCommandStream(srv.Context(), opEvents, in, request.Header, stream); err != nil {
		return err
	}

	for {
		result, ok := stream.Receive()
		if !ok {
			break
		}

		if result.Failed() {
			return result.Error
		}

		response := &ListenResponse{}
		output := result.Value.(server.SessionOutput)
		if err = proto.Unmarshal(output.Value.([]byte), response); err != nil {
			return err
		}

		var eventResponse *api.EventResponse
		switch output.Header.Type {
		case headers.ResponseType_OPEN_STREAM:
			eventResponse = &api.EventResponse{
				Header: output.Header,
			}
		case headers.ResponseType_CLOSE_STREAM:
			eventResponse = &api.EventResponse{
				Header: output.Header,
			}
		default:
			eventResponse = &api.EventResponse{
				Header:  output.Header,
				Type:    getEventType(response.Type),
				Key:     response.Key,
				Value:   response.Value,
				Version: int64(response.Version),
				Created: response.Created,
				Updated: response.Updated,
			}
		}

		log.Tracef("Sending EventResponse %+v", eventResponse)
		if err = srv.Send(eventResponse); err != nil {
			return err
		}
	}
	log.Tracef("Finished EventRequest %+v", request)
	return nil
}*/

// Entries lists all entries currently in the map
func (s *Server) Entries(request *api.EntriesRequest, srv api.MapService_EntriesServer) error {
	log.Info("Received EntriesRequest %+v", request)
	entries, err := redis.StringMap(s.DoCommand(request.Header, HGETALL, request.Header.Name.Name))
	if err != nil {
		return err
	}

	for key, value := range entries {
		entryResponse := &api.EntriesResponse{
			Key:   key,
			Value: []byte(value),
		}

		err = srv.Send(entryResponse)
		if err != nil {
			return err
		}

	}

	log.Info("Finished EntriesRequest %+v", request)
	return nil
}

/*func getResponseStatus(status UpdateStatus) api.ResponseStatus {
	switch status {
	case UpdateStatus_OK:
		return api.ResponseStatus_OK
	case UpdateStatus_NOOP:
		return api.ResponseStatus_NOOP
	case UpdateStatus_PRECONDITION_FAILED:
		return api.ResponseStatus_PRECONDITION_FAILED
	case UpdateStatus_WRITE_LOCK:
		return api.ResponseStatus_WRITE_LOCK
	}
	return api.ResponseStatus_OK
}

func getEventType(eventType ListenResponse_Type) api.EventResponse_Type {
	switch eventType {
	case ListenResponse_NONE:
		return api.EventResponse_NONE
	case ListenResponse_INSERTED:
		return api.EventResponse_INSERTED
	case ListenResponse_UPDATED:
		return api.EventResponse_UPDATED
	case ListenResponse_REMOVED:
		return api.EventResponse_REMOVED
	}
	return api.EventResponse_NONE
}*/
