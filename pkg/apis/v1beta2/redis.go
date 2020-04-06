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

package v1beta2

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// RedisStorageName is the name of the redis protocol
const RedisStorageName = "RedisStorage"

// RedisStorageSpec defines the desired state of RedisStorage
type RedisStorageSpec struct {
	// Backend is the redis backend
	Backend Backend `json:"backend,omitempty"`

	// Proxy is the redis proxy
	Proxy Proxy `json:"proxy,omitempty"`

	RedisStorageStatus RedisStorageStatus `json:"redis,omitempty"`
}

// RedisStorageStatus defines the observed state of RedisStorage
type RedisStorageStatus struct {

	// Proxy is the proxy status
	Proxy *ProxyStatus `json:"proxy,omitempty"`

	// Backend is the backend status
	Backend BackendStatus `json:"backend,omitempty"`
}

// +kubebuilder:object:root=true

// RedisStorage is the Schema for the redisstoragess API
type RedisStorage struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RedisStorageSpec   `json:"spec,omitempty"`
	Status RedisStorageStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// RedisStorageList contains a list of RedisStorage
type RedisStorageList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []RedisStorage `json:"items"`
}

func init() {
	SchemeBuilder.Register(&RedisStorage{}, &RedisStorageList{})
}
