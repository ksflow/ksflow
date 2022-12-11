/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"

	"github.com/twmb/franz-go/pkg/kgo"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	ksfv1 "github.com/ksflow/ksflow/api/v1alpha1"
)

// ClusterKafkaConfigReconciler reconciles a ClusterKafkaConfig object
type ClusterKafkaConfigReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=ksflow.io,resources=clusterkafkaconfigs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=ksflow.io,resources=clusterkafkaconfigs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=ksflow.io,resources=clusterkafkaconfigs/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop for the request
func (r *ClusterKafkaConfigReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// Get Object
	var ckc ksfv1.ClusterKafkaConfig
	if err := r.Get(ctx, req.NamespacedName, &ckc); err != nil {
		logger.Error(err, "unable to get ClusterKafkaConfig")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Create kafka client
	kafkaClient, err := kgo.NewClient(
		kgo.SeedBrokers(ckc.BootstrapServers()...),
	)
	if err != nil {
		logger.Error(err, "unable to create kafka client")
		return ctrl.Result{}, err
	}
	defer kafkaClient.Close()

	// Check connection to brokers
	connErr := kafkaClient.Ping(ctx)
	if connErr != nil {
		ckc.Status.Phase = ksfv1.ClusterKafkaConfigPhaseFailed
		ckc.Status.Message = "unable to request api versions from brokers"
	} else {
		ckc.Status.Phase = ksfv1.ClusterKafkaConfigPhaseAvailable
		ckc.Status.Message = ""
	}
	ckc.Status.LastUpdated = metav1.Now()

	// Update status
	if err = r.Status().Update(ctx, &ckc); err != nil {
		logger.Error(err, "unable to update ClusterKafkaConfig status")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ClusterKafkaConfigReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&ksfv1.ClusterKafkaConfig{}).
		Complete(r)
}
