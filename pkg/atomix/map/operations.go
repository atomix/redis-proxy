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

const (
	opPut     = "Put"
	opReplace = "Replace"
	opRemove  = "Remove"
	opGet     = "Get"
	opExists  = "Exists"
	opSize    = "Size"
	opClear   = "Clear"
	opEvents  = "Events"
	opEntries = "Entries"
)

// Redis commands
const (
	// Get the value of a hash field
	HGET = "HGET"
	// Get the values of all the given hash fields
	HMGET = "HMGET"
	// Set the string value of a hash field
	HSET = "HSET"
	// Delete one or more hash fields
	HDEL = "HDEL"
	// Get all fields and values in a hash
	HGETALL = "HGETALL"
	// Get number of fields in a hash
	HLEN = "HLEN"
	// Determine if a hash field exists
	HEXISTS = "HEXISTS"
	// Get all of the fields in a hash
	HKEYS = "HKEYS"
)
