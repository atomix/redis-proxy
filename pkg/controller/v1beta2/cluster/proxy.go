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
	"fmt"

	"strconv"

	storage "github.com/atomix/redis-proxy/pkg/apis/v1beta2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/atomix/redis-proxy/pkg/controller/v1beta2/util/k8s"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
)

func (r *Reconciler) reconcileRedisProxy(store *storage.RedisStorage) error {
	err := r.reconcileProxyConfigMap(store)
	if err != nil {
		return err
	}

	// Reconcile the redis proxy Deployment
	err = r.reconcileProxyDeployment(store)
	if err != nil {
		return err
	}

	// Reconcile the redis proxy Service
	err = r.reconcileProxyService(store)
	if err != nil {
		return err
	}
	return nil
}

func (r *Reconciler) reconcileProxyConfigMap(store *storage.RedisStorage) error {
	cm := &corev1.ConfigMap{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: k8s.GetProxyConfigMapName(store), Namespace: store.Namespace}, cm)
	if err != nil && errors.IsNotFound(err) {
		err = r.addProxyConfigMap(store)
	}
	return err
}

func (r *Reconciler) reconcileProxyDeployment(store *storage.RedisStorage) error {
	dep := &appsv1.Deployment{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: k8s.GetProxyDeploymentName(store), Namespace: store.Namespace}, dep)
	if err != nil && errors.IsNotFound(err) {
		err = r.addProxyDeployment(store)
	}
	return err
}

func (r *Reconciler) reconcileProxyService(store *storage.RedisStorage) error {
	service := &corev1.Service{}
	err := r.client.Get(context.TODO(), k8s.GetClusterServiceNamespacedName(store), service)
	if err != nil && errors.IsNotFound(err) {
		err = r.addProxyService(store)
	}
	return err
}

func (r *Reconciler) addProxyConfigMap(store *storage.RedisStorage) error {
	log.Info("Creating proxy ConfigMap", "Name", store.Name, "Namespace", store.Namespace)
	cm := &corev1.ConfigMap{}
	name := types.NamespacedName{
		Name:      fmt.Sprintf("%s-session", store.Name),
		Namespace: store.Namespace,
	}
	if err := r.client.Get(context.TODO(), name, cm); err != nil {
		if errors.IsNotFound(err) {
			cm = &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      name.Name,
					Namespace: name.Namespace,
				},
				Data: map[string]string{
					"session": "0",
				},
			}
			if err := controllerutil.SetControllerReference(store, cm, r.scheme); err != nil {
				return err
			}
			if err := r.client.Create(context.TODO(), cm); err != nil {
				return err
			}
			if err := r.client.Get(context.TODO(), name, cm); err != nil {
				return err
			}
		} else {
			return err
		}
	}

	session, err := strconv.Atoi(cm.Data["session"])
	if err != nil {
		return err
	}
	session++
	cm.Data["session"] = fmt.Sprintf("%d", session)
	if err := r.client.Update(context.TODO(), cm); err != nil {
		return err
	}

	config := map[string]interface{}{
		"sessionId": fmt.Sprintf("%d", session),
	}
	cm, err = k8s.NewProxyConfigMap(store, config)
	if err != nil {
		return err
	}
	if err := controllerutil.SetControllerReference(store, cm, r.scheme); err != nil {
		return err
	}
	return r.client.Create(context.TODO(), cm)
}

func (r *Reconciler) addProxyDeployment(store *storage.RedisStorage) error {
	log.Info("Creating proxy Deployment", "Name:", store.Name, "Namespace:", store.Namespace)
	dep, err := k8s.NewProxyDeployment(store)
	if err != nil {
		return err
	}
	if err := controllerutil.SetControllerReference(store, dep, r.scheme); err != nil {
		return err
	}
	return r.client.Create(context.TODO(), dep)
}

func (r *Reconciler) addProxyService(store *storage.RedisStorage) error {
	log.Info("Creating proxy service", "Name", store.Name, "Namespace", store.Namespace)
	service := k8s.NewProxyService(store)
	if err := controllerutil.SetControllerReference(store, service, r.scheme); err != nil {
		return err
	}
	return r.client.Create(context.TODO(), service)
}
