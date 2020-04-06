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
	"encoding/json"
	"errors"
	"fmt"

	"strconv"

	api "github.com/atomix/api/proto/atomix/controller"
	"github.com/atomix/kubernetes-controller/pkg/apis/cloud/v1beta2"
	storage "github.com/atomix/redis-storage/pkg/apis/v1beta1"
	"github.com/gogo/protobuf/jsonpb"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	policyv1beta1 "k8s.io/api/policy/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const (
	apiPort      = 5678
	protocolPort = 5679
)

// getClusterResourceName returns the given resource name for the given cluster
func getClusterResourceName(store *storage.RedisStorage, resource string) string {
	return fmt.Sprintf("%s-%s", store.Name, resource)
}

// GetClusterName returns the cluster name for the given cluster
func GetClusterName(database *v1beta2.Database, cluster int) string {
	return fmt.Sprintf("%s-%d", database.Name, cluster)
}

// GetClusterNamespacedName returns the NamespacedName for the given cluster
func GetClusterNamespacedName(database *v1beta2.Database, cluster int) types.NamespacedName {
	return types.NamespacedName{
		Name:      GetClusterName(database, cluster),
		Namespace: database.Namespace,
	}
}

// NewCluster returns the configuration for the given cluster
func NewCluster(database *v1beta2.Database, cluster int) *v1beta2.Cluster {
	meta := database.Spec.Template.ObjectMeta
	meta.Name = GetClusterName(database, cluster)
	meta.Namespace = database.Namespace
	if meta.Labels == nil {
		meta.Labels = make(map[string]string)
	}
	for key, value := range newClusterLabels(database, cluster) {
		meta.Labels[key] = value
	}
	meta.Annotations = newClusterAnnotations(database, cluster)
	return &v1beta2.Cluster{
		ObjectMeta: meta,
		Spec:       database.Spec.Template.Spec,
	}
}

// GetClusterLabelsForDatabase returns the labels for the clusters in the given database
func GetClusterLabelsForDatabase(database *v1beta2.Database) map[string]string {
	return map[string]string{
		appKey:      atomixApp,
		typeKey:     clusterType,
		databaseKey: database.Name,
	}
}

// GetClusterLabels returns the labels for the given cluster
func GetClusterLabels(store *storage.RedisStorage) map[string]string {
	labels := make(map[string]string)
	if value, ok := store.Labels[appKey]; ok {
		labels[appKey] = value
	}
	if value, ok := store.Labels[typeKey]; ok {
		labels[typeKey] = value
	}
	if value, ok := store.Labels[databaseKey]; ok {
		labels[databaseKey] = value
	}
	if value, ok := store.Labels[clusterKey]; ok {
		labels[clusterKey] = value
	}
	return labels
}

// newClusterLabels returns a new labels map containing the cluster app
func newClusterLabels(database *v1beta2.Database, cluster int) map[string]string {
	labels := GetClusterLabelsForDatabase(database)
	labels[clusterKey] = fmt.Sprint(cluster)
	return labels
}

// newClusterAnnotations returns annotations for the given cluster
func newClusterAnnotations(database *v1beta2.Database, cluster int) map[string]string {
	return map[string]string{
		controllerAnnotation: GetQualifiedControllerName(),
		typeAnnotation:       clusterType,
		databaseAnnotation:   database.Name,
		clusterAnnotation:    fmt.Sprint(cluster),
	}
}

// GetDatabaseFromClusterAnnotations returns the database name from the given cluster annotations
func GetDatabaseFromClusterAnnotations(store *storage.RedisStorage) (string, error) {
	database, ok := store.Annotations[databaseAnnotation]
	if !ok {
		return "", errors.New("cluster missing database annotation")
	}
	return database, nil
}

// GetClusterIDFromClusterAnnotations returns the cluster ID from the given cluster annotations
func GetClusterIDFromClusterAnnotations(store *storage.RedisStorage) (int32, error) {
	idstr, ok := store.Annotations[clusterAnnotation]
	if !ok {
		return 1, nil
	}

	id, err := strconv.ParseInt(idstr, 0, 32)
	if err != nil {
		return 0, err
	}
	return int32(id), nil
}

// GetClusterServiceName returns the given cluster's service name
func GetClusterServiceName(store *storage.RedisStorage) string {
	return store.Name
}

// getPodName returns the name of the pod for the given pod ID
func getPodName(store *storage.RedisStorage, pod int) string {
	return fmt.Sprintf("%s-%d", store.Name, pod)
}

// getPodDNSName returns the fully qualified DNS name for the given pod ID
func getPodDNSName(store *storage.RedisStorage, pod int) string {
	return fmt.Sprintf("%s-%d.%s.%s.svc.cluster.local", store.Name, pod, GetClusterHeadlessServiceName(store), store.Namespace)
}

// GetClusterServiceNamespacedName returns the given cluster's NamespacedName
func GetClusterServiceNamespacedName(store *storage.RedisStorage) types.NamespacedName {
	return types.NamespacedName{
		Name:      store.Name,
		Namespace: store.Namespace,
	}
}

// GetClusterHeadlessServiceName returns the headless service name for the given cluster
func GetClusterHeadlessServiceName(store *storage.RedisStorage) string {
	return getClusterResourceName(store, headlessServiceSuffix)
}

// GetClusterDisruptionBudgetName returns the pod disruption budget name for the given cluster
func GetClusterDisruptionBudgetName(store *storage.RedisStorage) string {
	return getClusterResourceName(store, disruptionBudgetSuffix)
}

// GetClusterConfigMapName returns the ConfigMap name for the given cluster
func GetClusterConfigMapName(store *storage.RedisStorage) string {
	return getClusterResourceName(store, configSuffix)
}

// GetClusterStatefulSetName returns the StatefulSet name for the given cluster
func GetClusterStatefulSetName(store *storage.RedisStorage) string {
	return store.Name
}

// NewClusterConfigMap returns a new ConfigMap for initializing Atomix clusters
func NewClusterConfigMap(store *storage.RedisStorage, config interface{}) (*corev1.ConfigMap, error) {
	clusterConfig, err := newNodeConfigString(store)
	if err != nil {
		return nil, err
	}

	protocolConfig, err := newProtocolConfigString(config)
	if err != nil {
		return nil, err
	}

	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      GetClusterConfigMapName(store),
			Namespace: store.Namespace,
			Labels:    store.Labels,
		},
		Data: map[string]string{
			clusterConfigFile:  clusterConfig,
			protocolConfigFile: protocolConfig,
		},
	}, nil
}

// newNodeConfigString creates a node configuration string for the given cluster
func newNodeConfigString(store *storage.RedisStorage) (string, error) {
	members := make([]*api.MemberConfig, store.Spec.Backend.Replicas)
	for i := 0; i < int(store.Spec.Backend.Replicas); i++ {
		members[i] = &api.MemberConfig{
			ID:           getPodName(store, i),
			Host:         getPodDNSName(store, i),
			ProtocolPort: protocolPort,
			APIPort:      apiPort,
		}
	}

	config := &api.ClusterConfig{
		Members: members,
	}

	marshaller := jsonpb.Marshaler{}
	return marshaller.MarshalToString(config)
}

// newProtocolConfigString creates a protocol configuration string for the given cluster and protocol
func newProtocolConfigString(config interface{}) (string, error) {
	bytes, err := json.Marshal(config)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// NewClusterDisruptionBudget returns a new pod disruption budget for the cluster group cluster
func NewClusterDisruptionBudget(store *storage.RedisStorage) *policyv1beta1.PodDisruptionBudget {
	minAvailable := intstr.FromInt(int(store.Spec.Backend.Replicas)/2 + 1)
	return &policyv1beta1.PodDisruptionBudget{
		ObjectMeta: metav1.ObjectMeta{
			Name:      GetClusterDisruptionBudgetName(store),
			Namespace: store.Namespace,
		},
		Spec: policyv1beta1.PodDisruptionBudgetSpec{
			MinAvailable: &minAvailable,
		},
	}
}

// NewClusterService returns a new service for a cluster
func NewClusterService(store *storage.RedisStorage) *corev1.Service {
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      GetClusterServiceName(store),
			Namespace: store.Namespace,
			Labels:    store.Labels,
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Name: "api",
					Port: apiPort,
				},
			},
			Selector: GetClusterLabels(store),
		},
	}
}

// NewClusterHeadlessService returns a new headless service for a cluster group
func NewClusterHeadlessService(store *storage.RedisStorage) *corev1.Service {
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      GetClusterHeadlessServiceName(store),
			Namespace: store.Namespace,
			Labels:    store.Labels,
			Annotations: map[string]string{
				"service.alpha.kubernetes.io/tolerate-unready-endpoints": "true",
			},
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Name: "api",
					Port: apiPort,
				},
				{
					Name: "protocol",
					Port: protocolPort,
				},
			},
			PublishNotReadyAddresses: true,
			ClusterIP:                "None",
			Selector:                 GetClusterLabels(store),
		},
	}
}

// NewBackendStatefulSet returns a new StatefulSet for a cluster group
func NewBackendStatefulSet(store *storage.RedisStorage) (*appsv1.StatefulSet, error) {
	volumeClaims := []corev1.PersistentVolumeClaim{}
	volumes := []corev1.Volume{
		newConfigVolume(GetClusterConfigMapName(store)),
	}
	if store.Spec.Backend.VolumeClaim != nil {
		volumeClaim := *store.Spec.Backend.VolumeClaim
		volumeClaim.Name = dataVolume
		volumeClaims = append(volumeClaims, volumeClaim)
	} else {
		volumes = append(volumes, newDataVolume())
	}

	image := store.Spec.Backend.Image
	pullPolicy := store.Spec.Backend.ImagePullPolicy
	livenessProbe := store.Spec.Backend.LivenessProbe
	readinessProbe := store.Spec.Backend.ReadinessProbe

	if pullPolicy == "" {
		pullPolicy = corev1.PullIfNotPresent
	}

	apiContainerPort := corev1.ContainerPort{
		Name:          "api",
		ContainerPort: 5678,
	}
	protocolContainerPort := corev1.ContainerPort{
		Name:          "protocol",
		ContainerPort: 5679,
	}

	containerBuilder := NewContainer()
	container := containerBuilder.SetImage(image).
		SetName("atomix").
		SetPullPolicy(pullPolicy).
		SetArgs(store.Spec.Backend.Args...).
		SetEnv(store.Spec.Backend.Env).
		SetPorts([]corev1.ContainerPort{apiContainerPort, protocolContainerPort}).
		SetReadinessProbe(readinessProbe).
		SetLivenessProbe(livenessProbe).
		SetVolumeMounts([]corev1.VolumeMount{newDataVolumeMount(), newConfigVolumeMount()}).
		Build()

	return &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      GetClusterStatefulSetName(store),
			Namespace: store.Namespace,
			Labels:    store.Labels,
		},
		Spec: appsv1.StatefulSetSpec{
			ServiceName: GetClusterHeadlessServiceName(store),
			Replicas:    &store.Spec.Backend.Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: GetClusterLabels(store),
			},
			UpdateStrategy: appsv1.StatefulSetUpdateStrategy{
				Type: appsv1.RollingUpdateStatefulSetStrategyType,
			},
			PodManagementPolicy:  appsv1.ParallelPodManagement,
			VolumeClaimTemplates: volumeClaims,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: store.Labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						newContainer(container),
					},
					Volumes: volumes,
				},
			},
		},
	}, nil
}
