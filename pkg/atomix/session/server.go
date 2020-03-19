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

	"github.com/atomix/redis-proxy/pkg/redisclient"

	api "github.com/atomix/api/proto/atomix/session"
	"github.com/atomix/redis-proxy/pkg/atomix/server"
	service "github.com/atomix/redis-proxy/pkg/server"
	"github.com/gomodule/redigo/redis"
	"google.golang.org/grpc"
)

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
	return &Server{
		Server: &server.Server{
			RedisPool: redisclient.NewPool("localhost:6379"),
		},
	}
}

// Server is an implementation of SessionServiceServer for session management
type Server struct {
	*server.Server
	RedisPool *redis.Pool
}

// OpenSession opens a new redis session
func (*Server) OpenSession(ctx context.Context, request *api.OpenSessionRequest) (*api.OpenSessionResponse, error) {
	panic("implement me")
}

// KeepAlive keeps a session alive
func (*Server) KeepAlive(ctx context.Context, request *api.KeepAliveRequest) (*api.KeepAliveResponse, error) {
	panic("implement me")
}

// CloseSession closes a new redis session
func (*Server) CloseSession(ctx context.Context, request *api.CloseSessionRequest) (*api.CloseSessionResponse, error) {
	panic("implement me")
}
