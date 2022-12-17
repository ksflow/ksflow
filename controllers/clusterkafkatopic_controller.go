/*
Copyright 2022 The Ksflow Authors.

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

	"github.com/twmb/franz-go/pkg/kadm"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	ksfv1 "github.com/ksflow/ksflow/api/v1alpha1"
)

// ClusterKafkaTopicReconciler reconciles a ClusterKafkaTopic object
type ClusterKafkaTopicReconciler struct {
	client.Client
	Scheme      *runtime.Scheme
	KafkaConfig ksfv1.KafkaConfig
}

//+kubebuilder:rbac:groups=ksflow.io,resources=clusterkafkatopics,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=ksflow.io,resources=clusterkafkatopics/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=ksflow.io,resources=clusterkafkatopics/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *ClusterKafkaTopicReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// Get ClusterKafkaTopic
	var kt ksfv1.ClusterKafkaTopic
	if err := r.Get(ctx, req.NamespacedName, &kt); err != nil {
		logger.Error(err, "unable to get ClusterKafkaTopic")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Create Kafka client
	kgoClient, err := r.KafkaConfig.NewClient()
	if err != nil {
		logger.Error(err, "unable to create Kafka client")
		return ctrl.Result{}, err
	}
	defer kgoClient.Close()
	kadmClient := kadm.NewClient(kgoClient)

	// TODO(user): your logic here

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ClusterKafkaTopicReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&ksfv1.ClusterKafkaTopic{}).
		Complete(r)
}
