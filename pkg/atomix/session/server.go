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

package session

import (
	"context"

	"github.com/atomix/redis-proxy/pkg/atomix/commands"

	"github.com/atomix/redis-proxy/pkg/manager"

	"github.com/gomodule/redigo/redis"

	"github.com/onosproject/onos-lib-go/pkg/logging"

	"github.com/atomix/api/proto/atomix/headers"
	api "github.com/atomix/api/proto/atomix/session"
	service "github.com/atomix/redis-proxy/pkg/server"
	"google.golang.org/grpc"
)

var log = logging.GetLogger("redis", "session")

// NewService returns a new Service
func NewService() (service.Service, error) {
	return &Service{}, nil
}

// Register registers the session service
func (s Service) Register(r *grpc.Server) {
	api.RegisterSessionServiceServer(r, newServer())
}

// Service is an implementation of map api service.
type Service struct {
	service.Service
}

func newServer() api.SessionServiceServer {
	return &Server{}
}

// Server is an implementation of SessionServiceServer for session management
type Server struct {
	api.SessionServiceServer
}

// OpenSession opens a new redis session
func (s *Server) OpenSession(ctx context.Context, request *api.OpenSessionRequest) (*api.OpenSessionResponse, error) {
	log.Info("Open a new session")
	mgr := manager.GetManager()
	conn := mgr.GetRedisPool().Get()
	sessionID, err := redis.Int64(conn.Do("CLIENT", "ID"))
	if err != nil {
		return &api.OpenSessionResponse{}, err
	}
	mgr.AddSession(sessionID, &conn)
	header := headers.ResponseHeader{
		SessionID: uint64(sessionID),
		Index:     uint64(sessionID),
	}
	response := api.OpenSessionResponse{
		Header: &header,
	}

	return &response, nil
}

// KeepAlive keeps a session alive
func (s *Server) KeepAlive(ctx context.Context, request *api.KeepAliveRequest) (*api.KeepAliveResponse, error) {
	mgr := manager.GetManager()
	conn := *mgr.GetSession(int64(request.Header.GetSessionID()))
	_, err := conn.Do(commands.PING)
	if err != nil {
		return &api.KeepAliveResponse{}, nil
	}

	header := headers.ResponseHeader{
		SessionID: uint64(request.Header.SessionID),
		Index:     uint64(request.Header.SessionID),
	}

	response := &api.KeepAliveResponse{
		Header: &header,
	}

	return response, nil
}

// CloseSession closes a new redis session
func (s *Server) CloseSession(ctx context.Context, request *api.CloseSessionRequest) (*api.CloseSessionResponse, error) {
	mgr := manager.GetManager()
	pool := mgr.GetRedisPool()
	conn := pool.Get()
	defer conn.Close()
	_, err := conn.Do("CLIENT", "KILL", "ID", request.Header.SessionID)
	if err != nil {
		return &api.CloseSessionResponse{}, err
	}
	header := headers.ResponseHeader{
		SessionID: uint64(request.Header.SessionID),
		Index:     uint64(request.Header.SessionID),
	}
	mgr.RemoveSession(int64(request.Header.SessionID))
	response := api.CloseSessionResponse{
		Header: &header,
	}

	return &response, nil

}
