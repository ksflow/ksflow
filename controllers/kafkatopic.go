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
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	KafkaTopicFinalizerName = "kafka-topic.ksflow.io/finalizer"
)

func handleDeletionAndFinalizers(ctx context.Context,
	meta *metav1.ObjectMeta,
	reclaimPolicy ksfv1.KafkaTopicReclaimPolicy,
	o client.Object,
	topicName string,
	kadmClient *kadm.Client) (finalizersUpdated bool, err error) {

	if meta.DeletionTimestamp.IsZero() {
		if !controllerutil.ContainsFinalizer(o, KafkaTopicFinalizerName) {
			return controllerutil.AddFinalizer(o, KafkaTopicFinalizerName), nil
		}
	} else {
		if controllerutil.ContainsFinalizer(o, KafkaTopicFinalizerName) {
			if reclaimPolicy == ksfv1.KafkaTopicReclaimPolicyDelete {
				if err = deleteTopicFromKafka(ctx, topicName, kadmClient); err != nil {
					return false, err
				}
			}
			var exists bool
			exists, err = topicExists(ctx, topicName, kadmClient)
			if err != nil {
				return false, err
			}
			if exists {
				// ref: return err so that it uses exponential backoff (ref: https://github.com/kubernetes-sigs/controller-runtime/issues/808#issuecomment-639845414)
				return false, errors.New("waiting for topic to finish deleting")
			}
			return controllerutil.RemoveFinalizer(o, KafkaTopicFinalizerName), nil
		}
	}
	return false, nil
}

func createOrUpdateTopic(ctx context.Context,
	desired *ksfv1.KafkaTopicInClusterConfiguration,
	topicName string,
	kadmClient *kadm.Client) error {

	var ticc *ksfv1.KafkaTopicInClusterConfiguration
	ticc, err := getTopicInClusterConfiguration(ctx, topicName, kadmClient)
	if err != nil {
		return err
	}
	if ticc != nil {
		if err = updateTopicInKafka(ctx, desired, ticc, topicName, kadmClient); err != nil {
			return err
		}
	} else {
		if err = createTopicInKafka(ctx, desired, topicName, kadmClient); err != nil {
			return err
		}
	}
	return nil
}

// topicExists checks if the given topic exists in the Kafka cluster
func topicExists(ctx context.Context, topicName string, kadmClient *kadm.Client) (bool, error) {
	allTopicDetails, err := kadmClient.ListTopics(ctx)
	if err != nil {
		return false, err
	}
	td, ok := allTopicDetails[topicName]
	if !ok || td.Err == kerr.UnknownTopicOrPartition {
		return false, nil
	}
	if td.Err != nil {
		return false, td.Err
	}
	return true, nil
}

// getTopicInClusterConfiguration retrieves the current observed state for the given topicName by making any necessary calls to Kafka
func getTopicInClusterConfiguration(ctx context.Context, topicName string, kadmClient *kadm.Client) (*ksfv1.KafkaTopicInClusterConfiguration, error) {
	ktc := ksfv1.KafkaTopicInClusterConfiguration{}
	allTopicDetails, err := kadmClient.ListTopics(ctx)
	if err != nil {
		return nil, err
	}
	td, ok := allTopicDetails[topicName]
	if !ok || td.Err == kerr.UnknownTopicOrPartition {
		return nil, nil
	}
	if td.Err != nil {
		return nil, td.Err
	}

	rf := int16(td.Partitions.NumReplicas())
	ktc.ReplicationFactor = &rf
	ktc.Partitions = pointer.Int32(int32(len(td.Partitions)))

	rcs, err := kadmClient.DescribeTopicConfigs(ctx, topicName)
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
func updateTopicInKafka(ctx context.Context,
	desired *ksfv1.KafkaTopicInClusterConfiguration,
	observed *ksfv1.KafkaTopicInClusterConfiguration,
	topicName string,
	kadmClient *kadm.Client) error {
	logger := log.FromContext(ctx)

	// Set Partitions
	desiredPartitions := negone32IfNil(desired.Partitions)
	observedPartitions := negone32IfNil(observed.Partitions)
	if desiredPartitions != observedPartitions {
		updatePartitionsResponses, err := kadmClient.UpdatePartitions(ctx, int(desiredPartitions), topicName)
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
		alterConfigsResponses, err := kadmClient.AlterTopicConfigs(ctx, alterConfigs, topicName)
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

	// Set ReplicationFactor
	desiredReplicationFactor := negone16IfNil(desired.ReplicationFactor)
	observedReplicationFactor := negone16IfNil(observed.ReplicationFactor)
	if desiredReplicationFactor != observedReplicationFactor {
		logger.Error(fmt.Errorf("cannot change replicationFactor from %d to %d", observedReplicationFactor, desiredReplicationFactor), "updating replication factor is not yet supported")
	}

	return nil
}

// createTopicInKafka creates the specified kafka topic using the provided Kafka client
func createTopicInKafka(ctx context.Context, kts *ksfv1.KafkaTopicInClusterConfiguration, topicName string, kadmClient *kadm.Client) error {
	responses, err := kadmClient.CreateTopics(ctx, negone32IfNil(kts.Partitions), negone16IfNil(kts.ReplicationFactor), kts.Configs, topicName)
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
func deleteTopicFromKafka(ctx context.Context, topicName string, kadmClient *kadm.Client) error {
	responses, err := kadmClient.DeleteTopics(ctx, topicName)
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

func negone32IfNil(p *int32) int32 {
	np := int32(-1)
	if p != nil {
		np = *p
	}
	return np
}

func negone16IfNil(p *int16) int16 {
	np := int16(-1)
	if p != nil {
		np = *p
	}
	return np
}
