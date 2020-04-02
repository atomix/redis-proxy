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

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/atomix/kubernetes-controller/pkg/apis/cloud/v1beta1"
	"github.com/atomix/redis-proxy/pkg/controller/v1beta1/util/k8s"
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
	log.Info("Creating proxy ConfigMap", "Name", cluster.Name, "Namespace", cluster.Namespace)
	cm := &corev1.ConfigMap{}
	name := types.NamespacedName{
		Name:      fmt.Sprintf("%s-session", cluster.Name),
		Namespace: cluster.Namespace,
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
			if err := controllerutil.SetControllerReference(cluster, cm, r.scheme); err != nil {
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
	cm, err = k8s.NewProxyConfigMap(cluster, config)
	if err != nil {
		return err
	}
	if err := controllerutil.SetControllerReference(cluster, cm, r.scheme); err != nil {
		return err
	}
	return r.client.Create(context.TODO(), cm)
}

func (r *Reconciler) addProxyDeployment(cluster *v1beta1.Cluster) error {
	log.Info("Creating proxy Deployment", "Name", cluster.Name, "Namespace", cluster.Namespace)
	dep, err := k8s.NewProxyDeployment(cluster)
	if err != nil {
		return err
	}
	if err := controllerutil.SetControllerReference(cluster, dep, r.scheme); err != nil {
		return err
	}
	return r.client.Create(context.TODO(), dep)
}

func (r *Reconciler) addProxyService(cluster *v1beta1.Cluster) error {
	log.Info("Creating proxy service", "Name", cluster.Name, "Namespace", cluster.Namespace)
	service := k8s.NewProxyService(cluster)
	if err := controllerutil.SetControllerReference(cluster, service, r.scheme); err != nil {
		return err
	}
	return r.client.Create(context.TODO(), service)
}
