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
	"fmt"

	"github.com/twmb/franz-go/pkg/kadm"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	ksfv1 "github.com/ksflow/ksflow/api/v1alpha1"
)

// KafkaServiceReconciler reconciles a KafkaService object
type KafkaServiceReconciler struct {
	client.Client
	Scheme                *runtime.Scheme
	KafkaConnectionConfig ksfv1.KafkaConnectionConfig
}

//+kubebuilder:rbac:groups=ksflow.io,resources=kafkaservices,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=ksflow.io,resources=kafkaservices/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=ksflow.io,resources=kafkaservices/finalizers,verbs=update
//+kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch;create;update;patch;delete

func (r *KafkaServiceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// Get KafkaService
	var ks ksfv1.KafkaService
	if err := r.Get(ctx, req.NamespacedName, &ks); err != nil {
		logger.Error(err, "unable to get KafkaService")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Create Kafka client
	kgoClient, err := r.KafkaConnectionConfig.NewClient()
	if err != nil {
		logger.Error(err, "unable to create Kafka client")
		return ctrl.Result{}, err
	}
	defer kgoClient.Close()
	kadmClient := kadm.NewClient(kgoClient)

	// Create a copy of KafkaClient with defaults applied
	ksCopy := ks.DeepCopy()

	// Reconcile
	err = reconcileService(ksCopy, kadmClient)

	// Update in-cluster spec w/finalizers
	if !equality.Semantic.DeepEqual(ks.Finalizers, ksCopy.Finalizers) {
		ks.SetFinalizers(ksCopy.Finalizers)
		if specErr := r.Client.Update(ctx, &ks); specErr != nil {
			if err != nil {
				err = fmt.Errorf("failed while updating spec: %v: %v", specErr, err)
			} else {
				err = fmt.Errorf("failed to update spec: %v", specErr)
			}
		}
	}

	// Update in-cluster status
	ksCopy.Status.DeepCopyInto(&ks.Status)
	if statusErr := r.Client.Status().Update(ctx, &ks); statusErr != nil {
		if err != nil {
			err = fmt.Errorf("failed while updating status: %v: %v", statusErr, err)
		} else {
			err = fmt.Errorf("failed to update status: %v", statusErr)
		}
	}

	return ctrl.Result{}, err
}

func reconcileService(kafkaService *ksfv1.KafkaService, kadmClient *kadm.Client) error {
	kafkaService.Status.LastUpdated = metav1.Now()
	kafkaService.Status.Phase = ksfv1.KsflowPhaseUnknown
	kafkaService.Status.Reason = ""

	errs := validation.IsDNS1035Label(kafkaService.Name)
	if len(errs) > 0 {
		kafkaService.Status.Phase = ksfv1.KsflowPhaseError
		kafkaService.Status.Reason = fmt.Sprintf("invalid KafkaService name: %q", errs[0])
		return fmt.Errorf(kafkaService.Status.Reason)
	}

	// Service deletion & finalizers
	if !kafkaService.DeletionTimestamp.IsZero() {
		kafkaService.Status.Phase = ksfv1.KsflowPhaseDeleting
	}
	ret, err := handleServiceDeletionAndFinalizers(kafkaService, kadmClient)
	if err != nil {
		kafkaService.Status.Phase = ksfv1.KsflowPhaseError
		kafkaService.Status.Reason = err.Error()
		return err
	}
	if ret {
		return nil
	}

	// Service create or update
	// TODO: create ACLs
	// TODO: create secret with ssl and bootstrap.servers

	// Update status
	// TODO: update status

	kafkaService.Status.Phase = ksfv1.KsflowPhaseAvailable
	return nil
}

// handleServiceDeletionAndFinalizers updates finalizers if necessary and handles deletion of kafka services
// returns false if processing should continue, true if we should finish reconcile
func handleServiceDeletionAndFinalizers(kafkaService *ksfv1.KafkaService, kadmClient *kadm.Client) (bool, error) {
	if kafkaService.DeletionTimestamp.IsZero() {
		if !controllerutil.ContainsFinalizer(kafkaService, KafkaServiceFinalizerName) {
			controllerutil.AddFinalizer(kafkaService, KafkaServiceFinalizerName)
		}
	} else {
		// The object is being deleted
		if controllerutil.ContainsFinalizer(kafkaService, KafkaServiceFinalizerName) {
			// our finalizer is present, so lets handle any external dependency
			if err := deleteACLsForPrincipalFromKafka(kafkaService.KafkaPrincipal(), kadmClient); err != nil {
				return true, err
			}
			exists, err := aclsForPrincipalExist(kafkaService.KafkaPrincipal(), kadmClient)
			if err != nil {
				return true, err
			}
			if exists {
				// ref: return err so that it uses exponential backoff (ref: https://github.com/kubernetes-sigs/controller-runtime/issues/808#issuecomment-639845414)
				return true, errors.New("waiting for acls to finish deleting")
			}
			controllerutil.RemoveFinalizer(kafkaService, KafkaServiceFinalizerName)
		}
		// Stop reconciliation as the item is being deleted
		return true, nil
	}
	return false, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *KafkaServiceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&ksfv1.KafkaService{}).
		Owns(&v1.Secret{}).
		Complete(r)
}
