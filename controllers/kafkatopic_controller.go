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
	"fmt"

	"github.com/twmb/franz-go/pkg/kadm"
	"github.com/twmb/franz-go/pkg/kgo"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	ksfv1 "github.com/ksflow/ksflow/api/v1alpha1"
)

type KafkaTopicReconciler struct {
	client.Client
	Scheme      *runtime.Scheme
	KafkaConfig ksfv1.KafkaConfig
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

	return ctrl.Result{}, r.doReconcile(ctx, &kt)
}

func (r *KafkaTopicReconciler) doReconcile(ctx context.Context, kt *ksfv1.KafkaTopic) (err error) {
	// Update status regardless of outcome
	defer func() {
		if statusErr := r.updateStatus(ctx, kt.Name); statusErr != nil {
			if err != nil {
				err = fmt.Errorf("failed while updating status: %v: %v", statusErr, err)
			} else {
				err = fmt.Errorf("failed to update status: %v", statusErr)
			}
		}
	}()

	// Create Kafka client
	var kgoClient *kgo.Client
	kgoClient, err = r.KafkaConfig.NewClient()
	if err != nil {
		return err
	}
	defer kgoClient.Close()
	kadmClient := kadm.NewClient(kgoClient)

	// Topic deletion & finalizers
	needsUpdate, err := handleDeletionAndFinalizers(ctx, &kt.ObjectMeta, *kt.Spec.ReclaimPolicy, kt, kt.FullTopicName(), kadmClient)
	if err != nil {
		return err
	}
	if !kt.ObjectMeta.DeletionTimestamp.IsZero() {
		return nil
	}
	if needsUpdate {
		if err = r.Update(ctx, kt); err != nil {
			return err
		}
	}

	return createOrUpdateTopic(ctx, &kt.Spec.KafkaTopicInClusterConfiguration, kt.FullTopicName(), kadmClient)
}

func (r *KafkaTopicReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&ksfv1.KafkaTopic{}).
		Complete(r)
}
