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

	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/atomix/kubernetes-controller/pkg/apis/cloud/v1beta1"
	"github.com/atomix/redis-proxy/pkg/controller/v1beta1/util/k8s"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	policyv1beta1 "k8s.io/api/policy/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
)

func (r *Reconciler) addBackendHeadlessService(cluster *v1beta1.Cluster) error {
	panic("Implement me")
}

func (r *Reconciler) addBackendDisruptionBudget(cluster *v1beta1.Cluster) error {
	panic("Implement me")
}

func (r *Reconciler) addBackendStatefulSet(cluster *v1beta1.Cluster) error {
	panic("Implement me")
}

func (r *Reconciler) addBackendConfigMap(cluster *v1beta1.Cluster) error {
	log.Info("Creating backend ConfigMap", "Name", cluster.Name, "Namespace", cluster.Namespace)
	var config interface{}
	if cluster.Spec.Backend.Protocol != nil {
		config = cluster.Spec.Backend.Protocol.Custom
	} else {
		config = ""
	}
	cm, err := k8s.NewClusterConfigMap(cluster, config)
	if err != nil {
		return err
	}
	if err := controllerutil.SetControllerReference(cluster, cm, r.scheme); err != nil {
		return err
	}
	return r.client.Create(context.TODO(), cm)
}

func (r *Reconciler) reconcileHeadlessService(cluster *v1beta1.Cluster) error {
	service := &corev1.Service{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: k8s.GetClusterHeadlessServiceName(cluster), Namespace: cluster.Namespace}, service)
	if err != nil && errors.IsNotFound(err) {
		err = r.addBackendHeadlessService(cluster)
	}
	return err
}

func (r *Reconciler) reconcileBackendStatefulSet(cluster *v1beta1.Cluster) error {
	set := &appsv1.StatefulSet{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: k8s.GetClusterStatefulSetName(cluster), Namespace: cluster.Namespace}, set)
	if err != nil && errors.IsNotFound(err) {
		err = r.addBackendStatefulSet(cluster)
	}
	return err
}

func (r *Reconciler) reconcileBackendConfigMap(cluster *v1beta1.Cluster) error {
	cm := &corev1.ConfigMap{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: k8s.GetClusterConfigMapName(cluster), Namespace: cluster.Namespace}, cm)
	if err != nil && errors.IsNotFound(err) {
		err = r.addBackendConfigMap(cluster)
	}
	return err
}

func (r *Reconciler) reconcileBackendDisruptionBudget(cluster *v1beta1.Cluster) error {
	budget := &policyv1beta1.PodDisruptionBudget{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: k8s.GetClusterDisruptionBudgetName(cluster), Namespace: cluster.Namespace}, budget)
	if err != nil && errors.IsNotFound(err) {
		err = r.addBackendDisruptionBudget(cluster)
	}
	return err
}
