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

	"github.com/atomix/kubernetes-controller/pkg/apis/cloud/v1beta2"

	"github.com/atomix/kubernetes-controller/pkg/controller/v1beta2/storage"
	"k8s.io/apimachinery/pkg/runtime/schema"

	storagev1beta1 "github.com/atomix/redis-storage/pkg/apis/v1beta1"
	"github.com/atomix/redis-storage/pkg/controller/v1beta1/util/k8s"
	"github.com/onosproject/onos-lib-go/pkg/logging"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrlmgr "sigs.k8s.io/controller-runtime/pkg/manager"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var log = logging.GetLogger("controller", "cluster", "reconciler")

// Add adds redis controller
func Add(mgr ctrlmgr.Manager) error {
	r := &Reconciler{
		client: mgr.GetClient(),
		scheme: mgr.GetScheme(),
		config: mgr.GetConfig(),
	}

	gvk := schema.GroupVersionKind{
		Group:   "storage.cloud.atomix.io",
		Version: "v1beta1",
		Kind:    "RedisStorage",
	}
	return storage.AddClusterReconciler(mgr, r, gvk)

}

var _ reconcile.Reconciler = &Reconciler{}

// Reconciler reconciles a cluster object
type Reconciler struct {
	client client.Client
	scheme *runtime.Scheme
	config *rest.Config
}

// Reconcile reconciles controller objects
func (r *Reconciler) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	log.Info("Reconciling Redis Storage")

	cluster := &v1beta2.Cluster{}
	err := r.client.Get(context.TODO(), request.NamespacedName, cluster)
	if err != nil {
		if errors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		return reconcile.Result{Requeue: true}, err
	}

	name := types.NamespacedName{
		Namespace: cluster.Spec.Storage.Namespace,
		Name:      cluster.Spec.Storage.Name,
	}

	store := &storagev1beta1.RedisStorage{}
	err = r.client.Get(context.TODO(), name, store)
	if err != nil {
		return reconcile.Result{}, err
	}

	// Reconcile the redis database backend
	err = r.reconcileRedisBackend(store)
	if err != nil {
		log.Info(err)
		return reconcile.Result{}, err
	}

	// Reconcile the redis proxy
	err = r.reconcileRedisProxy(store)
	if err != nil {
		return reconcile.Result{}, err
	}

	// Reconcile the cluster status
	if err := r.reconcileStatus(store); err != nil {
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

func (r *Reconciler) reconcileStatus(proto *storagev1beta1.RedisStorage) error {
	set := &appsv1.StatefulSet{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: k8s.GetClusterStatefulSetName(proto), Namespace: proto.Namespace}, set)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil
		}
		return err
	}

	// Update the backend replicas status
	if set.Status.ReadyReplicas != proto.Status.Backend.ReadyReplicas {
		log.Info("Updating Backend status", "Name", proto.Name, "Namespace", proto.Namespace, "ReadyReplicas", set.Status.ReadyReplicas)
		proto.Status.Backend.ReadyReplicas = set.Status.ReadyReplicas
		err = r.client.Status().Update(context.TODO(), proto)
		if err != nil {
			return err
		}
		return nil
	}

	deployment := &appsv1.Deployment{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: k8s.GetProxyDeploymentName(proto), Namespace: proto.Namespace}, deployment)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil
		}
		return err
	}

	if proto.Status.Proxy == nil {
		proto.Status.Proxy = &storagev1beta1.ProxyStatus{
			Ready: true,
		}
		err = r.client.Status().Update(context.TODO(), proto)
		if err != nil {
			return err
		}
		return nil
	} else if deployment.Status.ReadyReplicas > 0 != proto.Status.Proxy.Ready {
		proto.Status.Proxy.Ready = deployment.Status.ReadyReplicas > 0
		log.Info("Updating Proxy status", "Name", proto.Name, "Namespace", proto.Namespace, "Ready", proto.Status.Proxy.Ready)
		err = r.client.Status().Update(context.TODO(), proto)
		if err != nil {
			return err
		}
		return nil
	}

	return nil
}
