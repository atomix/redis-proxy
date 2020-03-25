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

	"github.com/atomix/api/proto/atomix/headers"
	api "github.com/atomix/api/proto/atomix/map"
	"github.com/atomix/redis-proxy/pkg/atomix/commands"
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
	// TODO It should be implemented

	log.Info("Received CreateRequest %+v", request)
	responseHeader := &headers.ResponseHeader{
		SessionID: request.Header.SessionID,
		Status:    headers.ResponseStatus_OK,
	}

	response := &api.CreateResponse{
		Header: responseHeader,
	}
	log.Info("Sending CreateResponse %+v", response)
	return response, nil
}

// Close closes a session
func (s *Server) Close(ctx context.Context, request *api.CloseRequest) (*api.CloseResponse, error) {
	// TODO It should be implemented
	log.Info("Received CloseRequest %+v", request)
	responseHeader := &headers.ResponseHeader{
		SessionID: request.Header.SessionID,
		Status:    headers.ResponseStatus_OK,
	}

	response := &api.CloseResponse{
		Header: responseHeader,
	}
	log.Info("Sending CloseResponse %+v", response)
	return response, nil
}

// Size gets the number of entries in the map
func (s *Server) Size(ctx context.Context, request *api.SizeRequest) (*api.SizeResponse, error) {
	log.Info("Received SizeRequest:", request)
	size, err := redis.Int(s.DoCommand(request.Header, commands.HLEN, request.Header.Name.Name))
	if err != nil {
		return nil, err
	}
	responseHeader := &headers.ResponseHeader{
		SessionID: request.Header.SessionID,
		Status:    headers.ResponseStatus_OK,
	}

	response := &api.SizeResponse{
		Header: responseHeader,
		Size_:  int32(size),
	}
	log.Info("Sending SizeResponse:", response)
	return response, nil
}

// Exists checks whether the map contains a key
func (s *Server) Exists(ctx context.Context, request *api.ExistsRequest) (*api.ExistsResponse, error) {
	log.Info("Received ExistsRequest:", request)
	containsKey, err := redis.Bool(s.DoCommand(request.Header, commands.HEXISTS, request.Header.Name.Name, request.Key))
	if err != nil {
		return nil, err
	}
	responseHeader := &headers.ResponseHeader{
		SessionID: request.Header.SessionID,
		Status:    headers.ResponseStatus_OK,
	}

	response := &api.ExistsResponse{
		Header:      responseHeader,
		ContainsKey: containsKey,
	}
	log.Info("Sending ExistsResponse:", response)
	return response, nil
}

// Put puts a key/value pair into the map
func (s *Server) Put(ctx context.Context, request *api.PutRequest) (*api.PutResponse, error) {
	log.Info("Received PutRequest:", request)
	_, err := s.DoCommand(request.Header, commands.HSET, request.Header.Name.Name, request.Key, request.Value)
	if err != nil {
		return nil, err
	}
	responseHeader := &headers.ResponseHeader{
		SessionID: request.Header.SessionID,
		Status:    headers.ResponseStatus_OK,
	}

	response := &api.PutResponse{
		Header: responseHeader,
	}
	log.Info("Sending PutResponse:", response)
	return response, nil
}

// Replace replaces a key/value pair in the map
func (s *Server) Replace(ctx context.Context, request *api.ReplaceRequest) (*api.ReplaceResponse, error) {
	log.Info("Received ReplaceRequest:", request)
	_, err := s.DoCommand(request.Header, commands.HSET, request.Header.Name.Name, request.Key, request.NewValue)
	if err != nil {
		return nil, err
	}

	responseHeader := &headers.ResponseHeader{
		SessionID: request.Header.SessionID,
		Status:    headers.ResponseStatus_OK,
	}

	response := &api.ReplaceResponse{
		Header: responseHeader,
	}
	log.Info("Sending ReplaceResponse:", response)
	return response, nil
}

// Get gets the value of a key
func (s *Server) Get(ctx context.Context, request *api.GetRequest) (*api.GetResponse, error) {
	log.Info("Received GetRequest:", request)
	value, err := redis.Bytes(s.DoCommand(request.Header, commands.HGET, request.Header.Name.Name, request.Key))
	if err != nil {
		return nil, err
	}

	if value == nil {
		return nil, err
	}
	responseHeader := &headers.ResponseHeader{
		SessionID: request.Header.SessionID,
		Status:    headers.ResponseStatus_OK,
	}

	response := &api.GetResponse{
		Header:  responseHeader,
		Value:   value,
		Version: int64(request.Header.SessionID),
	}
	log.Info("Sending GetRequest:", response)
	return response, nil
}

// Remove removes a key from the map
func (s *Server) Remove(ctx context.Context, request *api.RemoveRequest) (*api.RemoveResponse, error) {
	log.Info("Received RemoveRequest:", request)
	_, err := s.DoCommand(request.Header, commands.HDEL, request.Header.Name.Name, request.Key)
	if err != nil {
		return nil, err
	}

	responseHeader := &headers.ResponseHeader{
		SessionID: request.Header.SessionID,
		Status:    headers.ResponseStatus_OK,
	}

	response := &api.RemoveResponse{
		Header: responseHeader,
	}
	log.Info("Sending RemoveRequest:", response)
	return response, nil
}

// Clear removes all keys from the map
func (s *Server) Clear(ctx context.Context, request *api.ClearRequest) (*api.ClearResponse, error) {
	log.Info("Received ClearRequest:", request)

	keys, err := redis.Strings(s.DoCommand(request.Header, commands.HKEYS, request.Header.Name.Name))
	if err != nil {
		return nil, err
	}

	for _, key := range keys {
		_, err = s.DoCommand(request.Header, commands.HDEL, request.Header.Name.Name, key)
		if err != nil {
			return nil, err
		}
	}
	responseHeader := &headers.ResponseHeader{
		SessionID: request.Header.SessionID,
		Status:    headers.ResponseStatus_OK,
	}

	response := &api.ClearResponse{
		Header: responseHeader,
	}
	log.Info("Sending ClearResponse:", response)
	return response, nil
}

// Events listens for map change events
func (s *Server) Events(request *api.EventRequest, srv api.MapService_EventsServer) error {
	panic("Implement me")
}

// Entries lists all entries currently in the map
func (s *Server) Entries(request *api.EntriesRequest, srv api.MapService_EntriesServer) error {
	entries, err := redis.StringMap(s.DoCommand(request.Header, commands.HGETALL, request.Header.Name.Name))
	if err != nil {
		return err
	}
	// Opening entry response
	responseHeader := &headers.ResponseHeader{
		SessionID: request.Header.SessionID,
		Status:    headers.ResponseStatus_OK,
		Type:      headers.ResponseType_OPEN_STREAM,
	}

	entryResponse := &api.EntriesResponse{
		Header: responseHeader,
	}

	err = srv.Send(entryResponse)
	if err != nil {
		return err
	}

	// Map entry responses
	for key, value := range entries {
		responseHeader = &headers.ResponseHeader{
			SessionID: request.Header.SessionID,
			Status:    headers.ResponseStatus_OK,
			Type:      headers.ResponseType_RESPONSE,
		}
		entryResponse := &api.EntriesResponse{
			Header:  responseHeader,
			Key:     key,
			Value:   []byte(value),
			Version: int64(request.Header.SessionID),
		}

		err = srv.Send(entryResponse)
		if err != nil {
			return err
		}

	}

	// Closing entry response
	responseHeader = &headers.ResponseHeader{
		SessionID: request.Header.SessionID,
		Status:    headers.ResponseStatus_OK,
		Type:      headers.ResponseType_CLOSE_STREAM,
	}
	entryResponse = &api.EntriesResponse{
		Header: responseHeader,
	}

	err = srv.Send(entryResponse)
	if err != nil {
		return err
	}

	log.Info("Finished EntriesRequest:", request)
	return nil
}
