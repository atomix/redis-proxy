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
	"github.com/atomix/redis-proxy/pkg/controller/v1beta1/util/k8s"
	"github.com/onosproject/onos-lib-go/pkg/logging"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrlmgr "sigs.k8s.io/controller-runtime/pkg/manager"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logging.GetLogger("controller", "cluster", "reconciler")

// Add adds redis controller
func Add(mgr ctrlmgr.Manager) error {
	r := &Reconciler{
		client: mgr.GetClient(),
		scheme: mgr.GetScheme(),
		config: mgr.GetConfig(),
	}

	// Create a new redis controller
	c, err := controller.New("redis-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource Cluster
	err = c.Watch(&source.Kind{Type: &v1beta1.Cluster{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// Watch for changes to secondary resource StatefulSets for readiness checks
	err = c.Watch(&source.Kind{Type: &appsv1.StatefulSet{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &v1beta1.Cluster{},
	})
	if err != nil {
		return err
	}
	return nil
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
	log.Info("Reconciling Cluster")

	// Fetch the Cluster instance
	cluster := &v1beta1.Cluster{}
	err := r.client.Get(context.TODO(), request.NamespacedName, cluster)
	if err != nil {
		if errors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		return reconcile.Result{Requeue: true}, err
	}

	v1beta1.SetClusterDefaults(cluster)

	// Reconcile the redis database backend
	err = r.reconcileRedisBackend(cluster)
	if err != nil {
		return reconcile.Result{}, err
	}

	// Reconcile the redis proxy
	if cluster.Spec.Proxy != nil {
		err = r.reconcileRedisProxy(cluster)
		if err != nil {
			return reconcile.Result{}, err
		}
	}

	// Reconcile the cluster status
	if err := r.reconcileStatus(cluster); err != nil {
		return reconcile.Result{}, err
	}
	return reconcile.Result{}, nil
}

func (r *Reconciler) reconcileRedisProxy(cluster *v1beta1.Cluster) error {
	err := r.reconcileProxyConfigMap(cluster)
	if err != nil {
		return err
	}

	// Reconcile the redis proxy Deployment
	err = r.reconcileProxyDeployment(cluster)
	if err != nil {
		return err
	}

	// Reconcile the redis proxy Service
	err = r.reconcileProxyService(cluster)
	if err != nil {
		return err
	}
	return nil
}

func (r *Reconciler) reconcileRedisBackend(cluster *v1beta1.Cluster) error {
	// Reconcile the cluster config map
	err := r.reconcileBackendConfigMap(cluster)
	if err != nil {
		return err
	}

	// Reconcile the pod disruption budget
	err = r.reconcileBackendDisruptionBudget(cluster)
	if err != nil {
		return err
	}

	// Reconcile the StatefulSet
	err = r.reconcileBackendStatefulSet(cluster)
	if err != nil {
		return err
	}

	// Reconcile the headless cluster service
	err = r.reconcileHeadlessService(cluster)
	if err != nil {
		return err
	}

	return nil
}

func (r *Reconciler) reconcileStatus(cluster *v1beta1.Cluster) error {
	set := &appsv1.StatefulSet{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: k8s.GetClusterStatefulSetName(cluster), Namespace: cluster.Namespace}, set)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil
		}
		return err
	}

	// Update the backend replicas status
	if set.Status.ReadyReplicas != cluster.Status.Backend.ReadyReplicas {
		log.Info("Updating Backend status", "Name", cluster.Name, "Namespace", cluster.Namespace, "ReadyReplicas", set.Status.ReadyReplicas)
		cluster.Status.Backend.ReadyReplicas = set.Status.ReadyReplicas
		err = r.client.Status().Update(context.TODO(), cluster)
		if err != nil {
			return err
		}
		return nil
	}

	// Update the proxy status
	if cluster.Spec.Proxy != nil {
		deployment := &appsv1.Deployment{}
		err := r.client.Get(context.TODO(), types.NamespacedName{Name: k8s.GetProxyDeploymentName(cluster), Namespace: cluster.Namespace}, deployment)
		if err != nil {
			if errors.IsNotFound(err) {
				return nil
			}
			return err
		}

		if cluster.Status.Proxy == nil {
			cluster.Status.Proxy = &v1beta1.ProxyStatus{
				Ready: false,
			}
			err = r.client.Status().Update(context.TODO(), cluster)
			if err != nil {
				return err
			}
			return nil
		} else if deployment.Status.ReadyReplicas > 0 != cluster.Status.Proxy.Ready {
			cluster.Status.Proxy.Ready = deployment.Status.ReadyReplicas > 0
			log.Info("Updating Proxy status", "Name", cluster.Name, "Namespace", cluster.Namespace, "Ready", cluster.Status.Proxy.Ready)
			err = r.client.Status().Update(context.TODO(), cluster)
			if err != nil {
				return err
			}
			return nil
		}
	}

	return nil
}
