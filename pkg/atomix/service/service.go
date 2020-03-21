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

package service

import (
	"context"

	"github.com/atomix/redis-proxy/pkg/manager"
	"github.com/gomodule/redigo/redis"
	"github.com/onosproject/onos-lib-go/pkg/logging"

	"github.com/atomix/api/proto/atomix/headers"
)

// Server redis proxy server
type Server struct {
}

var log = logging.GetLogger("atomix", "service")

// DoCommand performs a redis command
func (s *Server) DoCommand(header *headers.RequestHeader, commandName string, args ...interface{}) (interface{}, error) {
	mgr := manager.GetManager()
	conn := *mgr.GetSession(int64(header.SessionID))
	sessionID, err := redis.Int64(conn.Do("CLIENT", "ID"))
	log.Info("Client ID", sessionID, err)
	response, err := conn.Do(commandName, args...)
	return response, err
}

// DoCreateService creates a service
func (s *Server) DoCreateService(ctx context.Context) {

}
