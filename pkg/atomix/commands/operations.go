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

package commands

// Redis commands
const (
	// Delete a key
	DEL = "DEL"
	// Set the value of an element in a list by its index
	LSET = "LSET"
	// Get an element from the list by its index
	LINDEX = "LINDEX"
	// Get the length of a list
	LLEN = "LLEN"
	// Append one or more elements to the end of a list
	RPUSH = "RPUSH"

	// Decrements the number stored at key by decrement
	DECRBY = "DECRBY"
	// Increments the number stored at key by increment
	INCRBY = "INCRBY"
	// Get the value of key
	GET = "GET"
	// Set the value of a key
	SET = "SET"

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

	// PING redis server
	PING = "PING"
)
