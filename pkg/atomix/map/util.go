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
	"github.com/gomodule/redigo/redis"
)

func (m *Service) getMapEntryValue(key string) (MapEntryValue, bool, error) {
	conn := m.redisPool.Get()
	defer conn.Close()
	mapValue, err := conn.Do(HGET, m.mapName, key)
	if mapValue == nil {
		return MapEntryValue{}, false, nil
	}
	if err != nil {
		return MapEntryValue{}, false, err
	}

	mapEntryValue := MapEntryValue{}
	err = mapEntryValue.Unmarshal(mapValue.([]byte))
	if err != nil {
		return MapEntryValue{}, false, err
	}
	return mapEntryValue, true, nil
}

func (m *Service) getMapEntries() map[string]*MapEntryValue {
	conn := m.redisPool.Get()
	defer conn.Close()
	entries, err := redis.StringMap(conn.Do(HGETALL, m.mapName))
	if err != nil {
		return nil
	}

	mapEntryValues := make(map[string]*MapEntryValue, len(entries))
	for key := range entries {
		mapEntryValue, _, _ := m.getMapEntryValue(key)
		mapEntryValues[key] = &mapEntryValue
	}

	return mapEntryValues

}
