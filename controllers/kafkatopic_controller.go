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

	ksfv1 "github.com/ksflow/ksflow/api/v1alpha1"
	"github.com/twmb/franz-go/pkg/kadm"
	"github.com/twmb/franz-go/pkg/kerr"
	"github.com/twmb/franz-go/pkg/kgo"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/pointer"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// KafkaTopicReconciler reconciles a KafkaTopic object
type KafkaTopicReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=ksflow.io,resources=kafkatopics,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=ksflow.io,resources=kafkatopics/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=ksflow.io,resources=kafkatopics/finalizers,verbs=update
// +kubebuilder:rbac:groups=ksflow.io,resources=kafkaconfigs,verbs=get;watch;list

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *KafkaTopicReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// Get KafkaTopic
	var kt ksfv1.KafkaTopic
	if err := r.Get(ctx, req.NamespacedName, &kt); err != nil {
		logger.Error(err, "unable to get KafkaTopic")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Get KafkaConfig
	var kc ksfv1.KafkaConfig
	if err := r.Get(ctx, types.NamespacedName{Name: KafkaConfigName}, &kc); err != nil {
		reason := fmt.Sprintf("unabled to get %q KafkaConfig", KafkaConfigName)
		return r.updateStatus(ctx, &kt, err, ksfv1.KafkaTopicPhaseFailed, reason)
	}

	// Create kafka client
	kgoClient, err := kgo.NewClient(
		kgo.SeedBrokers(kc.BootstrapServers()...),
	)
	if err != nil {
		logger.Error(err, "unable to create kafka client")
		return ctrl.Result{}, err
	}
	defer kgoClient.Close()
	kadmClient := kadm.NewClient(kgoClient)
	kt.Status.FullTopicName = kt.RawTopicName(&kc)

	// Topic delete logic
	if kt.ObjectMeta.DeletionTimestamp.IsZero() {
		if !controllerutil.ContainsFinalizer(&kt, KafkaTopicFinalizerName) {
			controllerutil.AddFinalizer(&kt, KafkaTopicFinalizerName)
			if err := r.Update(ctx, &kt); err != nil {
				return r.updateStatus(ctx, &kt, err, ksfv1.KafkaTopicPhaseFailed, "failed to add finalizer")
			}
		}
	} else {
		if controllerutil.ContainsFinalizer(&kt, KafkaTopicFinalizerName) {
			if *kt.Spec.ReclaimPolicy == ksfv1.KafkaTopicReclaimPolicyDelete {
				if err := r.deleteTopicFromKafka(ctx, kt.Status.FullTopicName, kadmClient); err != nil {
					return r.updateStatus(ctx, &kt, err, ksfv1.KafkaTopicPhaseFailed, "failed to delete topic from kafka")
				}
			}

			controllerutil.RemoveFinalizer(&kt, KafkaTopicFinalizerName)
			if err := r.Update(ctx, &kt); err != nil {
				return r.updateStatus(ctx, &kt, err, ksfv1.KafkaTopicPhaseFailed, "failed to remove finalizer")
			}
		}
		return ctrl.Result{}, nil
	}

	// Get observed state
	ktc, err := r.getKafkaTopicConfigFromKafka(ctx, kt.Status.FullTopicName, kadmClient)
	if err != nil {
		return r.updateStatus(ctx, &kt, err, ksfv1.KafkaTopicPhaseFailed, "failed to query kafka for existing topic")
	}

	if ktc != nil {
		// Topic update logic
		kt.Status.KafkaTopicConfig = *ktc
		err = r.updateTopicInKafka(ctx, &kt, kadmClient)
		if err != nil {
			return r.updateStatus(ctx, &kt, err, ksfv1.KafkaTopicPhaseFailed, "failed to update topic")
		}
	} else {
		// Topic create logic
		kt.Status.KafkaTopicConfig = ksfv1.KafkaTopicConfig{}
		err = r.createTopicInKafka(ctx, &kt, kadmClient)
		if err != nil {
			return r.updateStatus(ctx, &kt, err, ksfv1.KafkaTopicPhaseFailed, "failed to create topic")
		}
	}

	return r.updateStatus(ctx, &kt, nil, ksfv1.KafkaTopicPhaseAvailable, "")
}

func (r *KafkaTopicReconciler) updateStatus(ctx context.Context, kt *ksfv1.KafkaTopic, err error,
	phase ksfv1.KafkaTopicPhase, reason string) (ctrl.Result, error) {

	logger := log.FromContext(ctx)
	if err != nil {
		logger.Error(err, reason)
	}
	kt.Status.Phase = phase
	kt.Status.Reason = reason
	if uerr := r.Status().Update(ctx, kt); uerr != nil {
		logger.Error(uerr, "unable to update KafkaTopic status")
		return ctrl.Result{}, uerr
	}
	kt.Status.LastUpdated = metav1.Now()
	return ctrl.Result{}, err
}

// getKafkaTopicConfigFromKafka attempts to retrieve the state from Kafka
func (r *KafkaTopicReconciler) getKafkaTopicConfigFromKafka(ctx context.Context, topic string, kadmClient *kadm.Client) (*ksfv1.KafkaTopicConfig, error) {
	ktc := ksfv1.KafkaTopicConfig{}
	allTopicDetails, err := kadmClient.ListTopics(ctx)
	if err != nil {
		return nil, err
	}
	td, ok := allTopicDetails[topic]
	if !ok || td.Err == kerr.UnknownTopicOrPartition {
		return nil, nil
	}
	if td.Err != nil {
		return nil, td.Err
	}

	rf := int16(td.Partitions.NumReplicas())
	ktc.ReplicationFactor = &rf
	ktc.Partitions = pointer.Int32(int32(len(td.Partitions)))

	rcs, err := kadmClient.DescribeTopicConfigs(ctx, topic)
	if err != nil {
		return nil, err
	}
	rc, err := rcs.On(topic, nil)
	if rc.Err != nil {
		return nil, rc.Err
	}
	rcMap := map[string]*string{}
	for _, c := range rc.Configs {
		rcMap[c.Key] = c.Value
	}
	ktc.Configs = rcMap

	return &ktc, nil
}

func (r *KafkaTopicReconciler) updateTopicInKafka(ctx context.Context, kt *ksfv1.KafkaTopic, kadmClient *kadm.Client) error {
	// Set Partitions
	if kt.Spec.Partitions != nil && kt.Spec.Partitions != kt.Status.Partitions {
		updatePartitionsResponses, err := kadmClient.UpdatePartitions(ctx, int(*kt.Spec.Partitions), kt.Status.FullTopicName)
		if err != nil {
			return err
		}
		updatePartitionsResponse, err := updatePartitionsResponses.On(kt.Status.FullTopicName, nil)
		if updatePartitionsResponse.Err != nil {
			return updatePartitionsResponse.Err
		}
	}

	// Set Configs
	alterConfigs := []kadm.AlterConfig{}
	for k, v := range kt.Spec.Configs {
		if v != kt.Status.Configs[k] {
			alterConfigs = append(alterConfigs, kadm.AlterConfig{Op: kadm.SetConfig, Name: k, Value: v})
		}
	}
	for k, v := range kt.Status.Configs {
		if v != nil && kt.Spec.Configs[k] == nil {
			alterConfigs = append(alterConfigs, kadm.AlterConfig{Op: kadm.DeleteConfig, Name: k, Value: v})
		}
	}
	alterConfigsResponses, err := kadmClient.AlterTopicConfigs(ctx, alterConfigs, kt.Status.FullTopicName)
	if err != nil {
		return err
	}
	alterConfigsResponse, err := alterConfigsResponses.On(kt.Status.FullTopicName, nil)
	if err != nil {
		return err
	}
	if alterConfigsResponse.Err != nil {
		return err
	}

	// Set ReplicationFactor

	//TODO: implement
	kadmClient.ListPartitionReassignments()

	return nil
}

func (r *KafkaTopicReconciler) createTopicInKafka(ctx context.Context, kt *ksfv1.KafkaTopic, kadmClient *kadm.Client) error {
	partitions := int32(-1)
	if kt.Spec.Partitions != nil {
		partitions = *kt.Spec.Partitions
	}
	replicationFactor := int16(-1)
	if kt.Spec.ReplicationFactor != nil {
		replicationFactor = *kt.Spec.ReplicationFactor
	}
	responses, err := kadmClient.CreateTopics(ctx, partitions, replicationFactor, kt.Spec.Configs, kt.Status.FullTopicName)
	if err != nil {
		return err
	}
	response, err := responses.On(kt.Status.FullTopicName, nil)
	if err != nil {
		return err
	}
	if response.Err != nil {
		return err
	}
	return nil
}

func (r *KafkaTopicReconciler) deleteTopicFromKafka(ctx context.Context, topic string, kadmClient *kadm.Client) error {
	logger := log.FromContext(ctx)

	responses, err := kadmClient.DeleteTopics(ctx, topic)
	if err != nil {
		return err
	}
	response, err := responses.On(topic, nil)
	if err != nil {
		return err
	}
	if response.Err != nil {
		if response.Err.Error() == kerr.UnknownTopicOrPartition.Error() {
			logger.Error(err, "unable to delete topic", "topic", topic)
		} else {
			return response.Err
		}
	}
	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *KafkaTopicReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&ksfv1.KafkaTopic{}).
		Complete(r)
}
