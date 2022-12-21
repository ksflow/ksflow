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

	ksfv1 "github.com/ksflow/ksflow/api/v1alpha1"
	"github.com/twmb/franz-go/pkg/kadm"
	"github.com/twmb/franz-go/pkg/kerr"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/validation"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const (
	FinalizerName = "kafka-topic.ksflow.io/finalizer"
)

// doReconcile handles reconciliation of KafkaTopic and ClusterKafkaTopics
func doReconcile(
	meta *metav1.ObjectMeta,
	status *ksfv1.KafkaTopicStatus,
	spec *ksfv1.KafkaTopicSpec,
	o client.Object,
	topicName string,
	kadmClient *kadm.Client) error {

	status.LastUpdated = metav1.Now()
	status.Phase = ksfv1.KafkaTopicPhaseUnknown
	status.Reason = ""

	errs := validation.IsDNS1035Label(meta.Name)
	if len(errs) > 0 {
		status.Phase = ksfv1.KafkaTopicPhaseError
		status.Reason = fmt.Sprintf("invalid KafkaTopic name: %q", errs[0])
		return fmt.Errorf(status.Reason)
	}
	if spec.ReclaimPolicy == nil {
		status.Reason = "reclaim policy is required and no defaults found in controller config"
		return errors.New(status.Reason)
	}
	status.ReclaimPolicy = spec.ReclaimPolicy
	if spec.Partitions == nil {
		status.Reason = "topic partitions is required and no defaults found in controller config"
		return errors.New(status.Reason)
	}
	if spec.ReplicationFactor == nil {
		status.Reason = "topic replication factor is required and no defaults found in controller config"
		return errors.New(status.Reason)
	}

	// Topic deletion & finalizers
	if !meta.DeletionTimestamp.IsZero() {
		status.Phase = ksfv1.KafkaTopicPhaseDeleting
	}
	ret, err := handleDeletionAndFinalizers(meta, *status.ReclaimPolicy, o, topicName, kadmClient)
	if err != nil {
		status.Phase = ksfv1.KafkaTopicPhaseError
		status.Reason = err.Error()
		return err
	}
	if ret {
		return nil
	}

	// Topic create or update
	if err = createOrUpdateTopic(&spec.KafkaTopicInClusterConfiguration, topicName, kadmClient); err != nil {
		return err
	}
	if err != nil {
		status.Phase = ksfv1.KafkaTopicPhaseError
		status.Reason = err.Error()
		return err
	}

	// Update status
	var ticc *ksfv1.KafkaTopicInClusterConfiguration
	ticc, err = getTopicInClusterConfiguration(topicName, kadmClient)
	if err != nil {
		status.Phase = ksfv1.KafkaTopicPhaseError
		status.Reason = err.Error()
		return err
	}
	if ticc != nil {
		status.KafkaTopicInClusterConfiguration = *ticc
	}

	err = topicIsUpToDate(spec.KafkaTopicInClusterConfiguration, status.KafkaTopicInClusterConfiguration)
	if err != nil {
		status.Phase = ksfv1.KafkaTopicPhaseUpdating
		status.Reason = err.Error()
		return err
	}

	status.Phase = ksfv1.KafkaTopicPhaseAvailable
	return nil
}

// topicIsUpToDate returns an error indicating why the topic is not up-to-date, or nil if it is up-to-date
// assumes validation has already run and spec partitions/replicationFactor are non-nil
// for now it just checks the ones set in the spec. TODO: do better checking for unspecified values in spec.
func topicIsUpToDate(specKTICC ksfv1.KafkaTopicInClusterConfiguration, statusKTICC ksfv1.KafkaTopicInClusterConfiguration) error {
	if statusKTICC.Partitions == nil {
		return fmt.Errorf(`spec partitions %d does not match status partitions "nil"`, *specKTICC.Partitions)
	}
	if *specKTICC.Partitions != *statusKTICC.Partitions {
		return fmt.Errorf(`spec partitions %d does not equal status partitions %d`, *specKTICC.Partitions, *statusKTICC.Partitions)
	}
	if statusKTICC.ReplicationFactor == nil {
		return fmt.Errorf(`spec replicationFactor %d does not match status replicationFactor "nil"`, *specKTICC.ReplicationFactor)
	}
	if *specKTICC.ReplicationFactor != *statusKTICC.ReplicationFactor {
		return fmt.Errorf(`spec replicationFactor %d does not equal status replicationFactor %d`, *specKTICC.ReplicationFactor, *statusKTICC.ReplicationFactor)
	}
	// there may be some default configs that are set for Kafka topics, but not in spec.  for now just checking the configs we explicitly set.
	for k, v := range specKTICC.Configs {
		if v != nil {
			if statusKTICC.Configs[k] == nil {
				return fmt.Errorf(`spec %q config value %q does not equal status config "nil"`, k, *v)
			}
			if *statusKTICC.Configs[k] != *v {
				return fmt.Errorf(`spec %q config value %q does not equal status config %q`, k, *v, *statusKTICC.Configs[k])
			}
		}
	}
	return nil
}

// handleDeletionAndFinalizers updates finalizers if necessary and handles deletion of kafka topics
// returns false if processing should continue, true if we should finish reconcile
func handleDeletionAndFinalizers(
	meta *metav1.ObjectMeta,
	reclaimPolicy ksfv1.KafkaTopicReclaimPolicy,
	o client.Object,
	topicName string,
	kadmClient *kadm.Client) (bool, error) {

	if meta.DeletionTimestamp.IsZero() {
		if !controllerutil.ContainsFinalizer(o, FinalizerName) {
			controllerutil.AddFinalizer(o, FinalizerName)
		}
	} else {
		// The object is being deleted
		if controllerutil.ContainsFinalizer(o, FinalizerName) {
			// our finalizer is present, so lets handle any external dependency
			if reclaimPolicy == ksfv1.KafkaTopicReclaimPolicyDelete {
				if err := deleteTopicFromKafka(topicName, kadmClient); err != nil {
					return true, err
				}
			}
			exists, err := topicExists(topicName, kadmClient)
			if err != nil {
				return true, err
			}
			if exists {
				// ref: return err so that it uses exponential backoff (ref: https://github.com/kubernetes-sigs/controller-runtime/issues/808#issuecomment-639845414)
				return true, errors.New("waiting for topic to finish deleting")
			}
			controllerutil.RemoveFinalizer(o, FinalizerName)
		}
		// Stop reconciliation as the item is being deleted
		return true, nil
	}
	return false, nil
}

func createOrUpdateTopic(
	desired *ksfv1.KafkaTopicInClusterConfiguration,
	topicName string,
	kadmClient *kadm.Client) error {

	var ticc *ksfv1.KafkaTopicInClusterConfiguration
	ticc, err := getTopicInClusterConfiguration(topicName, kadmClient)
	if err != nil {
		return err
	}
	if ticc != nil {
		if err = updateTopicInKafka(desired, ticc, topicName, kadmClient); err != nil {
			return err
		}
	} else {
		if err = createTopicInKafka(desired, topicName, kadmClient); err != nil {
			return err
		}
	}
	return nil
}

// topicExists returns true if the topic exists
func topicExists(topicName string, kadmClient *kadm.Client) (bool, error) {
	tds, err := kadmClient.ListTopics(context.Background())
	if err != nil {
		return false, err
	}
	td, ok := tds[topicName]
	return ok, td.Err
}

// getTopicInClusterConfiguration retrieves the current observed state for the given topicName by making any necessary calls to Kafka
func getTopicInClusterConfiguration(topicName string, kadmClient *kadm.Client) (*ksfv1.KafkaTopicInClusterConfiguration, error) {
	ktc := ksfv1.KafkaTopicInClusterConfiguration{}
	tds, err := kadmClient.ListTopics(context.Background())
	if err != nil {
		return nil, err
	}
	td, ok := tds[topicName]
	if !ok || td.Err == kerr.UnknownTopicOrPartition {
		return nil, nil
	}
	if td.Err != nil {
		return nil, td.Err
	}

	rf := int16(td.Partitions.NumReplicas())
	ktc.ReplicationFactor = &rf
	ktc.Partitions = pointer.Int32(int32(len(td.Partitions)))

	rcs, err := kadmClient.DescribeTopicConfigs(context.Background(), topicName)
	if err != nil {
		return nil, err
	}
	rc, err := rcs.On(topicName, nil)
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

// updateTopicInKafka compares the desired state (coming from spec) and the observed state (coming from the status),
// making any necessary calls to Kafka to bring them closer together.
func updateTopicInKafka(
	desired *ksfv1.KafkaTopicInClusterConfiguration,
	observed *ksfv1.KafkaTopicInClusterConfiguration,
	topicName string,
	kadmClient *kadm.Client) error {

	// Set ReplicationFactor
	if *desired.ReplicationFactor != *observed.ReplicationFactor {
		return fmt.Errorf("cannot change replicationFactor from %d to %d, updating replication factor is not yet supported", *observed.ReplicationFactor, *desired.ReplicationFactor)
	}

	// Set Partitions
	if *desired.Partitions != *observed.Partitions {
		updatePartitionsResponses, err := kadmClient.UpdatePartitions(context.Background(), int(*desired.Partitions), topicName)
		if err != nil {
			return err
		}
		updatePartitionsResponse, err := updatePartitionsResponses.On(topicName, nil)
		if updatePartitionsResponse.Err != nil {
			return updatePartitionsResponse.Err
		}
	}

	// Set Configs
	alterConfigs := []kadm.AlterConfig{}
	for k, v := range desired.Configs {
		if v != observed.Configs[k] {
			alterConfigs = append(alterConfigs, kadm.AlterConfig{Op: kadm.SetConfig, Name: k, Value: v})
		}
	}
	for k, v := range observed.Configs {
		if v != nil && desired.Configs[k] == nil {
			alterConfigs = append(alterConfigs, kadm.AlterConfig{Op: kadm.DeleteConfig, Name: k, Value: v})
		}
	}
	if len(alterConfigs) != 0 {
		alterConfigsResponses, err := kadmClient.AlterTopicConfigs(context.Background(), alterConfigs, topicName)
		if err != nil {
			return err
		}
		alterConfigsResponse, err := alterConfigsResponses.On(topicName, nil)
		if err != nil {
			return err
		}
		if alterConfigsResponse.Err != nil {
			return err
		}
	}

	return nil
}

// createTopicInKafka creates the specified kafka topic using the provided Kafka client
func createTopicInKafka(kts *ksfv1.KafkaTopicInClusterConfiguration, topicName string, kadmClient *kadm.Client) error {
	responses, err := kadmClient.CreateTopics(context.Background(), *kts.Partitions, *kts.ReplicationFactor, kts.Configs, topicName)
	if err != nil {
		return err
	}
	response, err := responses.On(topicName, nil)
	if err != nil {
		return err
	}
	if response.Err != nil {
		return err
	}
	return nil
}

// deleteTopicFromKafka deletes the specified kafka topic using the provided Kafka client
func deleteTopicFromKafka(topicName string, kadmClient *kadm.Client) error {
	responses, err := kadmClient.DeleteTopics(context.Background(), topicName)
	if err != nil {
		return err
	}
	response, err := responses.On(topicName, nil)
	if err != nil {
		return err
	}
	if response.Err != nil && response.Err != kerr.UnknownTopicOrPartition {
		return response.Err
	}
	return nil
}
