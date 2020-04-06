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

package k8s

import (
	"fmt"

	storage "github.com/atomix/redis-storage/pkg/apis/v1beta1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// getProxyName returns the name of the given cluster's proxy
func getProxyName(store *storage.RedisStorage) string {
	return fmt.Sprintf("%s-proxy", store.Name)
}

// GetProxyLabels returns the labels for the given partition
func GetProxyLabels(store *storage.RedisStorage) map[string]string {
	labels := make(map[string]string)
	if value, ok := store.Labels[appKey]; ok {
		labels[appKey] = value
	}
	if value, ok := store.Labels[databaseKey]; ok {
		labels[databaseKey] = value
	}
	if value, ok := store.Labels[clusterKey]; ok {
		labels[clusterKey] = value
	}
	labels[typeKey] = proxyType
	return labels
}

// GetProxyConfigMapName returns the ConfigMap name for the given proxy and partition
func GetProxyConfigMapName(store *storage.RedisStorage) string {
	return getClusterResourceName(store, getProxyName(store))
}

// GetProxyDeploymentName returns the StatefulSet name for the given partition
func GetProxyDeploymentName(store *storage.RedisStorage) string {
	return getClusterResourceName(store, getProxyName(store))
}

// NewProxyConfigMap returns a new ConfigMap for initializing a proxy
func NewProxyConfigMap(store *storage.RedisStorage, config map[string]interface{}) (*corev1.ConfigMap, error) {
	protocolConfig, err := newProtocolConfigString(config)
	if err != nil {
		return nil, err
	}

	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      GetProxyConfigMapName(store),
			Namespace: store.Namespace,
			Labels:    store.Labels,
		},
		Data: map[string]string{
			protocolConfigFile: protocolConfig,
		},
	}, nil
}

// NewProxyService returns a new service for a partition proxy
func NewProxyService(store *storage.RedisStorage) *corev1.Service {
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      GetClusterServiceName(store),
			Namespace: store.Namespace,
			Labels:    GetProxyLabels(store),
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Name: "api",
					Port: 5678,
				},
			},
			Selector: GetProxyLabels(store),
		},
	}
}

// NewProxyDeployment returns a new Deployment for a partition proxy
func NewProxyDeployment(store *storage.RedisStorage) (*appsv1.Deployment, error) {
	var one int32 = 1
	image := store.Spec.Proxy.Image
	pullPolicy := store.Spec.Proxy.ImagePullPolicy
	livenessProbe := store.Spec.Proxy.LivenessProbe
	readinessProbe := store.Spec.Proxy.ReadinessProbe

	if pullPolicy == "" {
		pullPolicy = corev1.PullIfNotPresent
	}
	containerBuilder := NewContainer()
	apiContainerPort := corev1.ContainerPort{
		Name:          "api",
		ContainerPort: 5678,
	}
	protocolContainerPort := corev1.ContainerPort{
		Name:          "protocol",
		ContainerPort: 5679,
	}

	container := containerBuilder.SetImage(image).
		SetName("atomix").
		SetPullPolicy(pullPolicy).
		SetArgs(store.Spec.Proxy.Args...).
		SetEnv(store.Spec.Proxy.Env).
		SetPorts([]corev1.ContainerPort{apiContainerPort, protocolContainerPort}).
		SetReadinessProbe(readinessProbe).
		SetLivenessProbe(livenessProbe).
		SetVolumeMounts([]corev1.VolumeMount{newConfigVolumeMount()}).
		Build()

	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      GetProxyDeploymentName(store),
			Namespace: store.Namespace,
			Labels:    GetProxyLabels(store),
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &one,
			Selector: &metav1.LabelSelector{
				MatchLabels: GetProxyLabels(store),
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: GetProxyLabels(store),
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						newContainer(container),
					},
					Volumes: []corev1.Volume{
						newConfigVolume(GetProxyConfigMapName(store)),
					},
				},
			},
		},
	}, nil
}
