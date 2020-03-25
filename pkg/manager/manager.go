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

package manager

import (
	"github.com/atomix/redis-proxy/pkg/redisclient"
	"github.com/gomodule/redigo/redis"
	"github.com/onosproject/onos-lib-go/pkg/logging"
	"google.golang.org/grpc"
)

var mgr Manager

var log = logging.GetLogger("manager")

// NewManager creates a new redis proxy manager
func NewManager(redisEndPoint string, opts []grpc.DialOption) (*Manager, error) {
	log.Info("Initialize redis proxy manager")
	mgr = Manager{
		redisPool: redisclient.NewPool(redisEndPoint),
		sessions:  make(map[int64]*redis.Conn),
	}
	return &mgr, nil
}

// Manager redis proxy manager
type Manager struct {
	redisPool *redis.Pool
	sessions  map[int64]*redis.Conn
}

// RemoveSession remove a session from list of sessions
func (m *Manager) RemoveSession(sessionID int64) {
	delete(m.sessions, sessionID)
}

// GetSession returns a connection based on a given session ID
func (m *Manager) GetSession(sessionID int64) *redis.Conn {
	if conn, ok := m.sessions[sessionID]; ok {
		return conn
	}
	return nil
}

// AddSession adds a session to list of sessions
func (m *Manager) AddSession(sessionID int64, conn *redis.Conn) {
	m.sessions[sessionID] = conn
}

// GetRedisPool get redis pool
func (m *Manager) GetRedisPool() *redis.Pool {
	return m.redisPool
}

// GetManager returns the manager
func GetManager() *Manager {
	return &mgr
}
