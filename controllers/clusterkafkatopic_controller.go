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
	"github.com/twmb/franz-go/pkg/kgo"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	ksfv1 "github.com/ksflow/ksflow/api/v1alpha1"
)

type ClusterKafkaTopicReconciler struct {
	client.Client
	Scheme      *runtime.Scheme
	KafkaConfig ksfv1.KafkaConfig
}

//+kubebuilder:rbac:groups=ksflow.io,resources=clusterkafkatopics,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=ksflow.io,resources=clusterkafkatopics/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=ksflow.io,resources=clusterkafkatopics/finalizers,verbs=update

func (r *ClusterKafkaTopicReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// Get ClusterKafkaTopic
	var kt ksfv1.ClusterKafkaTopic
	if err := r.Get(ctx, req.NamespacedName, &kt); err != nil {
		logger.Error(err, "unable to get ClusterKafkaTopic")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	return ctrl.Result{}, r.doReconcile(ctx, &kt)
}

// same as KafkaTopicReconciler's doReconcile... use generics?
func (r *ClusterKafkaTopicReconciler) doReconcile(ctx context.Context, kt *ksfv1.ClusterKafkaTopic) (err error) {
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
	if kt.ObjectMeta.DeletionTimestamp.IsZero() {
		if !controllerutil.ContainsFinalizer(kt, KafkaTopicFinalizerName) {
			controllerutil.AddFinalizer(kt, KafkaTopicFinalizerName)
			if err = r.Update(ctx, kt); err != nil {
				return err
			}
		}
	} else {
		if controllerutil.ContainsFinalizer(kt, KafkaTopicFinalizerName) {
			if *kt.Spec.ReclaimPolicy == ksfv1.KafkaTopicReclaimPolicyDelete {
				if err = deleteTopicFromKafka(ctx, kt.FullTopicName(), kadmClient); err != nil {
					return err
				}
			}
			var exists bool
			exists, err = topicExists(ctx, kt.FullTopicName(), kadmClient)
			if err != nil {
				return err
			}
			if exists {
				// ref: return err so that it uses exponential backoff (ref: https://github.com/kubernetes-sigs/controller-runtime/issues/808#issuecomment-639845414)
				return errors.New("waiting for topic to finish deleting")
			}
			controllerutil.RemoveFinalizer(kt, KafkaTopicFinalizerName)
			if err = r.Update(ctx, kt); err != nil {
				return err
			}
		}
		return nil
	}

	return createOrUpdateTopic(ctx, &kt.Spec.KafkaTopicInClusterConfiguration, kt.FullTopicName(), kadmClient)
}

func (r *ClusterKafkaTopicReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&ksfv1.ClusterKafkaTopic{}).
		Complete(r)
}
