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
	"errors"

	"github.com/twmb/franz-go/pkg/kgo"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	ksfv1 "github.com/ksflow/ksflow/api/v1alpha1"
)

// KafkaConfigReconciler reconciles a KafkaConfig object
type KafkaConfigReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=ksflow.io,resources=kafkaconfigs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=ksflow.io,resources=kafkaconfigs/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=ksflow.io,resources=kafkaconfigs/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *KafkaConfigReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// Get Object
	var kc ksfv1.KafkaConfig
	if err := r.Get(ctx, req.NamespacedName, &kc); err != nil {
		logger.Error(err, "unable to get KafkaConfig")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Create kafka client
	kafkaClient, err := kgo.NewClient(
		kgo.SeedBrokers(kc.BootstrapServers()...),
	)
	if err != nil {
		logger.Error(err, "unable to create kafka client")
		return ctrl.Result{}, err
	}
	defer kafkaClient.Close()

	// Delete logic
	if kc.ObjectMeta.DeletionTimestamp.IsZero() {
		if !controllerutil.ContainsFinalizer(&kc, KafkaConfigFinalizerName) {
			controllerutil.AddFinalizer(&kc, KafkaConfigFinalizerName)
			if err := r.Update(ctx, &kc); err != nil {
				return r.updateStatus(ctx, &kc, err, ksfv1.KafkaConfigPhaseFailed, "failed to add finalizer", true)
			}
		}
	} else {
		if controllerutil.ContainsFinalizer(&kc, KafkaConfigFinalizerName) {
			var ktl ksfv1.KafkaTopicList
			if ktlErr := r.List(ctx, &ktl); ktlErr != nil {
				return r.updateStatus(ctx, &kc, err, ksfv1.KafkaConfigPhaseFailed, "unabled to list KafkaTopics", true)
			}
			if len(ktl.Items) > 0 {
				// ref: return err so that it uses exponential backoff (ref: https://github.com/kubernetes-sigs/controller-runtime/issues/808#issuecomment-639845414)
				reason := "waiting for KafkaTopics to be deleted"
				return r.updateStatus(ctx, &kc, errors.New(reason), ksfv1.KafkaConfigPhaseDeleting, reason, false)
			}

			controllerutil.RemoveFinalizer(&kc, KafkaConfigFinalizerName)
			if err := r.Update(ctx, &kc); err != nil {
				return r.updateStatus(ctx, &kc, err, ksfv1.KafkaConfigPhaseFailed, "failed to remove finalizer", true)
			}
		}
		return ctrl.Result{}, nil
	}

	// Check connection to brokers
	connErr := kafkaClient.Ping(ctx)
	if connErr != nil {
		return r.updateStatus(ctx, &kc, nil, ksfv1.KafkaConfigPhaseFailed, "unable to request api versions from brokers", true)
	}

	return r.updateStatus(ctx, &kc, nil, ksfv1.KafkaConfigPhaseAvailable, "", true)
}

func (r *KafkaConfigReconciler) updateStatus(ctx context.Context, kc *ksfv1.KafkaConfig, err error,
	phase ksfv1.KafkaConfigPhase, reason string, logerr bool) (ctrl.Result, error) {

	logger := log.FromContext(ctx)
	if err != nil && logerr {
		logger.Error(err, reason)
	}
	kc.Status.Phase = phase
	kc.Status.Reason = reason
	kc.Status.LastUpdated = metav1.Now()
	if err = r.Status().Update(ctx, kc); err != nil {
		logger.Error(err, "unable to update KafkaConfig status")
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *KafkaConfigReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&ksfv1.KafkaConfig{}).
		Complete(r)
}
