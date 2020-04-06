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

	storage "github.com/atomix/redis-storage/pkg/apis/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/atomix/redis-storage/pkg/controller/v1beta1/util/k8s"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	policyv1beta1 "k8s.io/api/policy/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
)

func (r *Reconciler) reconcileRedisBackend(store *storage.RedisStorage) error {
	// Reconcile the cluster config map
	err := r.reconcileBackendConfigMap(store)
	if err != nil {
		return err
	}

	// Reconcile the pod disruption budget
	err = r.reconcileBackendDisruptionBudget(store)
	if err != nil {
		return err
	}

	// Reconcile the StatefulSet
	err = r.reconcileBackendStatefulSet(store)
	if err != nil {
		return err
	}

	// Reconcile the headless cluster service
	err = r.reconcileHeadlessService(store)
	if err != nil {
		return err
	}

	return nil
}

func (r *Reconciler) addBackendHeadlessService(store *storage.RedisStorage) error {
	log.Info("Creating headless backend service", "Name", store.Name, "Namespace", store.Namespace)
	service := k8s.NewClusterHeadlessService(store)
	if err := controllerutil.SetControllerReference(store, service, r.scheme); err != nil {
		return err
	}
	return r.client.Create(context.TODO(), service)
}

func (r *Reconciler) addBackendDisruptionBudget(store *storage.RedisStorage) error {
	log.Info("Creating pod disruption budget", "Name", store.Name, "Namespace", store.Namespace)
	budget := k8s.NewClusterDisruptionBudget(store)
	if err := controllerutil.SetControllerReference(store, budget, r.scheme); err != nil {
		return err
	}
	return r.client.Create(context.TODO(), budget)
}

func (r *Reconciler) addBackendStatefulSet(store *storage.RedisStorage) error {
	log.Info("Creating backend replicas", "Name:", store.Name, "Namespace:", store.Namespace)
	set, err := k8s.NewBackendStatefulSet(store)
	if err != nil {
		return err
	}
	if err := controllerutil.SetControllerReference(store, set, r.scheme); err != nil {
		log.Info(err)
		return err
	}

	return r.client.Create(context.TODO(), set)
}

func (r *Reconciler) addBackendConfigMap(store *storage.RedisStorage) error {
	log.Info("Creating backend ConfigMap", "Name:", store.Name, "Namespace:", store.Namespace)
	var config interface{}

	cm, err := k8s.NewClusterConfigMap(store, config)
	if err != nil {
		return err
	}
	if err := controllerutil.SetControllerReference(store, cm, r.scheme); err != nil {
		return err
	}
	return r.client.Create(context.TODO(), cm)
}

func (r *Reconciler) reconcileHeadlessService(store *storage.RedisStorage) error {
	service := &corev1.Service{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: k8s.GetClusterHeadlessServiceName(store), Namespace: store.Namespace}, service)
	if err != nil && errors.IsNotFound(err) {
		err = r.addBackendHeadlessService(store)
	}
	return err
}

func (r *Reconciler) reconcileBackendStatefulSet(store *storage.RedisStorage) error {
	set := &appsv1.StatefulSet{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: k8s.GetClusterStatefulSetName(store), Namespace: store.Namespace}, set)
	if err != nil && errors.IsNotFound(err) {
		err = r.addBackendStatefulSet(store)
	}
	return err
}

func (r *Reconciler) reconcileBackendConfigMap(store *storage.RedisStorage) error {
	cm := &corev1.ConfigMap{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: k8s.GetClusterConfigMapName(store), Namespace: store.Namespace}, cm)
	if err != nil && errors.IsNotFound(err) {
		err = r.addBackendConfigMap(store)
	}
	return err
}

func (r *Reconciler) reconcileBackendDisruptionBudget(store *storage.RedisStorage) error {
	budget := &policyv1beta1.PodDisruptionBudget{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: k8s.GetClusterDisruptionBudgetName(store), Namespace: store.Namespace}, budget)
	if err != nil && errors.IsNotFound(err) {
		err = r.addBackendDisruptionBudget(store)
	}
	return err
}
