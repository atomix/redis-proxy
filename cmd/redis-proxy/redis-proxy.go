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

package main

import (
	"flag"

	"github.com/atomix/redis-proxy/pkg/atomix/session"

	_map "github.com/atomix/redis-proxy/pkg/atomix/map"
	service "github.com/atomix/redis-proxy/pkg/server"
	"github.com/onosproject/onos-lib-go/pkg/logging"
)

var log = logging.GetLogger("main")

func main() {
	caPath := flag.String("caPath", "", "path to CA certificate")
	keyPath := flag.String("keyPath", "", "path to client private key")
	certPath := flag.String("certPath", "", "path to client certificate")
	port := flag.Int("port", 5150, "redis proxy port")
	//redisAddress := flag.String("redis-address", "redis:6379", "redis server address")
	flag.Parse()
	err := startServer(*caPath, *keyPath, *certPath, *port)
	if err != nil {
		log.Fatal("Unable to start onos-ric ", err)
	}

}

// Creates gRPC server and registers various services; then serves.
func startServer(caPath string, keyPath string, certPath string, port int) error {
	s := service.NewServer(service.NewServerConfig(caPath, keyPath, certPath, int16(port), true))
	s.AddService(_map.Service{})
	s.AddService(session.Service{})

	return s.Serve(func(started string) {
		log.Info("Started Redis Proxy Server", started)
	})
}
