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

package kafkatopic

import (
	"context"
	"fmt"

	"github.com/twmb/franz-go/pkg/kadm"
	"k8s.io/apimachinery/pkg/api/equality"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	ksfv1 "github.com/ksflow/ksflow/api/v1alpha1"
)

type KafkaTopicReconciler struct {
	client.Client
	Scheme                 *runtime.Scheme
	KafkaConnectionConfig  ksfv1.KafkaConnectionConfig
	KafkaTopicSpecDefaults ksfv1.KafkaTopicSpec
}

//+kubebuilder:rbac:groups=ksflow.io,resources=kafkatopics,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=ksflow.io,resources=kafkatopics/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=ksflow.io,resources=kafkatopics/finalizers,verbs=update

func (r *KafkaTopicReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	var kt ksfv1.KafkaTopic
	if err := r.Get(ctx, req.NamespacedName, &kt); err != nil {
		logger.Error(err, "unable to get KafkaTopic")
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

	// Reconcile, update spec w/finalizers, update status, return
	ktCopy := kt.DeepCopy()
	ktCopySpec, err := ktCopy.Spec.WithDefaultsFrom(&r.KafkaTopicSpecDefaults)
	if err != nil {
		logger.Error(err, "unable to set defaults on Kafka Topic")
		return ctrl.Result{}, err
	}
	err = doReconcile(&ktCopy.ObjectMeta, &ktCopy.Status, ktCopySpec, ktCopy, ktCopy.FullTopicName(), kadmClient)
	if !equality.Semantic.DeepEqual(kt.Finalizers, ktCopy.Finalizers) {
		if specErr := r.Client.Update(ctx, ktCopy); specErr != nil {
			if err != nil {
				err = fmt.Errorf("failed while updating spec: %v: %v", specErr, err)
			} else {
				err = fmt.Errorf("failed to update spec: %v", specErr)
			}
		}
	}
	kt.Status.LastUpdated = metav1.Now()
	if statusErr := r.Client.Status().Update(ctx, ktCopy); statusErr != nil {
		if err != nil {
			err = fmt.Errorf("failed while updating status: %v: %v", statusErr, err)
		} else {
			err = fmt.Errorf("failed to update status: %v", statusErr)
		}
	}
	return ctrl.Result{}, err
}

func (r *KafkaTopicReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&ksfv1.KafkaTopic{}).
		Complete(r)
}
