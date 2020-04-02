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

package cluster

import (
	"context"

	"github.com/atomix/kubernetes-controller/pkg/apis/cloud/v1beta1"
	"github.com/atomix/kubernetes-controller/pkg/controller/v1beta1/util/k8s"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
)

func (r *Reconciler) reconcileProxyConfigMap(cluster *v1beta1.Cluster) error {
	cm := &corev1.ConfigMap{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: k8s.GetProxyConfigMapName(cluster), Namespace: cluster.Namespace}, cm)
	if err != nil && errors.IsNotFound(err) {
		err = r.addProxyConfigMap(cluster)
	}
	return err
}

func (r *Reconciler) reconcileProxyDeployment(cluster *v1beta1.Cluster) error {
	dep := &appsv1.Deployment{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: k8s.GetProxyDeploymentName(cluster), Namespace: cluster.Namespace}, dep)
	if err != nil && errors.IsNotFound(err) {
		err = r.addProxyDeployment(cluster)
	}
	return err
}

func (r *Reconciler) reconcileProxyService(cluster *v1beta1.Cluster) error {
	service := &corev1.Service{}
	err := r.client.Get(context.TODO(), k8s.GetClusterServiceNamespacedName(cluster), service)
	if err != nil && errors.IsNotFound(err) {
		err = r.addProxyService(cluster)
	}
	return err
}

func (r *Reconciler) addProxyConfigMap(cluster *v1beta1.Cluster) error {
	panic("Implement me")
}

func (r *Reconciler) addProxyDeployment(cluster *v1beta1.Cluster) error {
	panic("Implement me")
}

func (r *Reconciler) addProxyService(cluster *v1beta1.Cluster) error {
	panic("Implement me")
}
