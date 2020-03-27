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

package list

import (
	"context"

	"github.com/gomodule/redigo/redis"

	"github.com/atomix/redis-proxy/pkg/atomix/commands"

	"github.com/atomix/api/proto/atomix/headers"
	api "github.com/atomix/api/proto/atomix/list"
	"github.com/atomix/redis-proxy/pkg/atomix/service"
	"github.com/atomix/redis-proxy/pkg/server"
	"github.com/onosproject/onos-lib-go/pkg/logging"
	"google.golang.org/grpc"
)

var log = logging.GetLogger("redis", "list")

// NewService returns a new Service
func NewService() (server.Service, error) {
	return &Service{}, nil
}

// Service is an implementation of list api service.
type Service struct {
	server.Service
}

// Register registers the list service
func (s Service) Register(r *grpc.Server) {
	api.RegisterListServiceServer(r, newServer())
}

func newServer() api.ListServiceServer {
	return &Server{
		Server: &service.Server{},
	}
}

// Server is an implementation of ListServiceServer for the list primitive
type Server struct {
	api.ListServiceServer
	*service.Server
}

// Create opens a new session
func (s *Server) Create(ctx context.Context, request *api.CreateRequest) (*api.CreateResponse, error) {
	log.Info("Received CreateRequest ", request)
	responseHeader := &headers.ResponseHeader{
		SessionID: request.Header.SessionID,
		Status:    headers.ResponseStatus_OK,
	}

	response := &api.CreateResponse{
		Header: responseHeader,
	}
	log.Info("Sending CreateResponse ", response)
	return response, nil
}

// Close closes a session
func (s *Server) Close(ctx context.Context, request *api.CloseRequest) (*api.CloseResponse, error) {
	// TODO It should be implemented
	log.Info("Received CloseRequest:", request)
	responseHeader := &headers.ResponseHeader{
		SessionID: request.Header.SessionID,
		Status:    headers.ResponseStatus_OK,
	}

	response := &api.CloseResponse{
		Header: responseHeader,
	}
	log.Info("Sending CloseResponse:", response)
	return response, nil
}

// Size gets the number of elements in the list
func (s *Server) Size(ctx context.Context, request *api.SizeRequest) (*api.SizeResponse, error) {
	log.Info("Received SizeRequest:", request)
	size, err := redis.Int64(s.DoCommand(request.Header, commands.LLEN, request.Header.Name.Name))
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
	log.Info("Sent AppendResponse", response)
	return response, nil
}

// Contains checks whether the list contains a value
func (s *Server) Contains(ctx context.Context, request *api.ContainsRequest) (*api.ContainsResponse, error) {
	panic("Implement me")
}

// Append adds a value to the end of the list
func (s *Server) Append(ctx context.Context, request *api.AppendRequest) (*api.AppendResponse, error) {
	log.Info("Received AppendRequest:", request)
	_, err := s.DoCommand(request.Header, commands.RPUSH, request.Header.Name.Name, request.Value)
	if err != nil {
		return nil, err
	}
	responseHeader := &headers.ResponseHeader{
		SessionID: request.Header.SessionID,
		Status:    headers.ResponseStatus_OK,
	}
	response := &api.AppendResponse{
		Header: responseHeader,
		Status: api.ResponseStatus_OK,
	}
	log.Info("Sent AppendResponse", response)
	return response, nil
}

// Insert inserts a value at a specific index
func (s *Server) Insert(ctx context.Context, request *api.InsertRequest) (*api.InsertResponse, error) {
	log.Info("Received InsertRequest:", request)

	err := s.DoLuaScript(request.Header, insertScript, 1, "BEFORE", request.Value)
	if err != nil {
		log.Info(err)
	}

	responseHeader := &headers.ResponseHeader{
		SessionID: request.Header.SessionID,
		Status:    headers.ResponseStatus_OK,
	}
	response := &api.InsertResponse{
		Header: responseHeader,
		Status: api.ResponseStatus_OK,
	}
	log.Info("Sent InsertResponse", response)
	return response, nil

}

// Set sets the value at a specific index
func (s *Server) Set(ctx context.Context, request *api.SetRequest) (*api.SetResponse, error) {
	log.Info("Received SetRequest:", request)
	_, err := s.DoCommand(request.Header, commands.LSET, request.Header.Name.Name, request.Index, request.Value)
	if err != nil {
		return nil, err
	}
	responseHeader := &headers.ResponseHeader{
		SessionID: request.Header.SessionID,
		Status:    headers.ResponseStatus_OK,
	}
	response := &api.SetResponse{
		Header: responseHeader,
		Status: api.ResponseStatus_OK,
	}
	log.Info("Sent SetResponse", response)
	return response, nil
}

// Get gets the value at a specific index
func (s *Server) Get(ctx context.Context, request *api.GetRequest) (*api.GetResponse, error) {
	log.Info("Received GetRequest:", request)
	value, err := redis.String(s.DoCommand(request.Header, commands.LINDEX, request.Header.Name.Name, request.Index))
	if err != nil {
		return nil, err
	}
	responseHeader := &headers.ResponseHeader{
		SessionID: request.Header.SessionID,
		Status:    headers.ResponseStatus_OK,
	}
	response := &api.GetResponse{
		Header: responseHeader,
		Status: api.ResponseStatus_OK,
		Value:  value,
	}
	return response, nil
}

// Remove removes an index from the list
func (s *Server) Remove(ctx context.Context, request *api.RemoveRequest) (*api.RemoveResponse, error) {
	panic("Implement me")
}

// Clear removes all indexes from the list
func (s *Server) Clear(ctx context.Context, request *api.ClearRequest) (*api.ClearResponse, error) {
	log.Info("Received ClearRequest:", request)
	_, err := s.DoCommand(request.Header, commands.DEL, request.Header.Name.Name)
	if err != nil {
		return nil, err
	}
	responseHeader := &headers.ResponseHeader{
		SessionID: request.Header.SessionID,
		Status:    headers.ResponseStatus_OK,
	}
	response := &api.ClearResponse{
		Header: responseHeader,
	}
	log.Info("Sent ClearResponse", response)
	return response, nil

}

// Events listens for list change events
func (s *Server) Events(request *api.EventRequest, srv api.ListService_EventsServer) error {
	panic("Implement me")
}

// Iterate lists all the value in the list
func (s *Server) Iterate(request *api.IterateRequest, srv api.ListService_IterateServer) error {
	panic("Implement me")
}
