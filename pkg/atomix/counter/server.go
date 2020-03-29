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

package counter

import (
	"context"

	"github.com/gomodule/redigo/redis"

	"github.com/atomix/redis-proxy/pkg/atomix/commands"

	api "github.com/atomix/api/proto/atomix/counter"
	"github.com/atomix/api/proto/atomix/headers"
	"github.com/atomix/redis-proxy/pkg/atomix/service"
	"github.com/atomix/redis-proxy/pkg/server"
	"github.com/onosproject/onos-lib-go/pkg/logging"
	"google.golang.org/grpc"
)

var log = logging.GetLogger("redis", "counter")

// NewService returns a new Service
func NewService() (server.Service, error) {
	return &Service{}, nil
}

// Service is an implementation of counter api service.
type Service struct {
	server.Service
}

// Register registers the counter service
func (s Service) Register(r *grpc.Server) {
	api.RegisterCounterServiceServer(r, newServer())
}

func newServer() api.CounterServiceServer {
	return &Server{
		Server: &service.Server{},
	}
}

// Server is an implementation of CounterServiceServer for the counter primitive
type Server struct {
	api.CounterServiceServer
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

// Set sets the current value of the counter
func (s *Server) Set(ctx context.Context, request *api.SetRequest) (*api.SetResponse, error) {
	log.Info("Received SetRequest ", request)
	_, err := s.DoCommand(request.Header, commands.SET, request.Header.Name.Name, request.Value)
	if err != nil {
		return nil, err
	}
	responseHeader := &headers.ResponseHeader{
		SessionID: request.Header.SessionID,
		Status:    headers.ResponseStatus_OK,
	}
	response := &api.SetResponse{
		Header: responseHeader,
	}
	log.Info("Sent SetResponse", response)
	return response, nil

}

// Get gets the current value of the counter
func (s *Server) Get(ctx context.Context, request *api.GetRequest) (*api.GetResponse, error) {
	log.Info("Received GetRequest:", request)
	value, err := redis.Int64(s.DoCommand(request.Header, commands.GET, request.Header.Name.Name))
	if err != nil {
		return nil, err
	}
	responseHeader := &headers.ResponseHeader{
		SessionID: request.Header.SessionID,
		Status:    headers.ResponseStatus_OK,
	}
	response := &api.GetResponse{
		Header: responseHeader,
		Value:  value,
	}
	log.Info("Sent GetResponse:", response)
	return response, nil
}

// Increment increments the value of the counter by a delta
func (s *Server) Increment(ctx context.Context, request *api.IncrementRequest) (*api.IncrementResponse, error) {
	log.Info("Received IncrementRequest:", request)
	prevValue, err := redis.Int64(s.DoCommand(request.Header, commands.GET, request.Header.Name.Name))
	if err != nil {
		return nil, err
	}
	_, err = s.DoCommand(request.Header, commands.INCRBY, request.Header.Name.Name, request.Delta)
	if err != nil {
		return nil, err
	}

	nextValue, err := redis.Int64(s.DoCommand(request.Header, commands.GET, request.Header.Name.Name))
	if err != nil {
		return nil, err
	}
	responseHeader := &headers.ResponseHeader{
		SessionID: request.Header.SessionID,
		Status:    headers.ResponseStatus_OK,
	}
	response := &api.IncrementResponse{
		Header:        responseHeader,
		PreviousValue: prevValue,
		NextValue:     nextValue,
	}

	log.Info("Sent IncrementResponse", response)
	return response, nil

}

// Decrement decrements the value of the counter by a delta
func (s *Server) Decrement(ctx context.Context, request *api.DecrementRequest) (*api.DecrementResponse, error) {
	log.Info("Received DecrementRequest:", request)
	prevValue, err := redis.Int64(s.DoCommand(request.Header, commands.GET, request.Header.Name.Name))
	if err != nil {
		return nil, err
	}
	_, err = s.DoCommand(request.Header, commands.DECRBY, request.Header.Name.Name, request.Delta)
	if err != nil {
		return nil, err
	}

	nextValue, err := redis.Int64(s.DoCommand(request.Header, commands.GET, request.Header.Name.Name))
	if err != nil {
		return nil, err
	}
	responseHeader := &headers.ResponseHeader{
		SessionID: request.Header.SessionID,
		Status:    headers.ResponseStatus_OK,
	}
	response := &api.DecrementResponse{
		Header:        responseHeader,
		PreviousValue: prevValue,
		NextValue:     nextValue,
	}

	log.Info("Sent DecrementResponse", response)
	return response, nil
}

// CheckAndSet updates the value of the counter conditionally
func (s *Server) CheckAndSet(ctx context.Context, request *api.CheckAndSetRequest) (*api.CheckAndSetResponse, error) {
	log.Info("Received CheckAndSetRequest", request)
	_, err := s.DoLuaScript(request.Header, checkAndSetScript, 1, request.Header.Name.Name, request.Expect, request.Update)
	if err != nil {
		responseHeader := &headers.ResponseHeader{
			SessionID: request.Header.SessionID,
			Status:    headers.ResponseStatus_ERROR,
		}
		response := &api.CheckAndSetResponse{
			Header:    responseHeader,
			Succeeded: false,
		}

		log.Info("Sent CAS response", response)
		return response, err
	}

	responseHeader := &headers.ResponseHeader{
		SessionID: request.Header.SessionID,
		Status:    headers.ResponseStatus_OK,
	}
	response := &api.CheckAndSetResponse{
		Header:    responseHeader,
		Succeeded: true,
	}

	log.Info("Sent CAS response", response)
	return response, nil
}

// Close closes a session
func (s *Server) Close(ctx context.Context, request *api.CloseRequest) (*api.CloseResponse, error) {
	// TODO It should be implemented
	log.Info("Received CloseRequest ", request)
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
