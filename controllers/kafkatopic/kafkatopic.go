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
	"errors"
	"fmt"
	"strconv"

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
	FinalizerName = "kafka-topic.ksflow.io/finalizer"
)

// doReconcile handles reconciliation of KafkaTopic and ClusterKafkaTopics
func doReconcile(ctx context.Context,
	meta *metav1.ObjectMeta,
	status *ksfv1.KafkaTopicStatus,
	spec *ksfv1.KafkaTopicSpec,
	reclaimPolicy ksfv1.KafkaTopicReclaimPolicy,
	o client.Object,
	topicName string,
	kadmClient *kadm.Client) error {

	status.Phase = ksfv1.KafkaTopicPhaseUnknown

	// Topic deletion & finalizers
	if !meta.DeletionTimestamp.IsZero() {
		status.Phase = ksfv1.KafkaTopicPhaseDeleting
	}
	ret, err := handleDeletionAndFinalizers(ctx, meta, reclaimPolicy, o, topicName, kadmClient)
	if err != nil {
		status.Phase = ksfv1.KafkaTopicPhaseError
		return err
	}
	if ret {
		return nil
	}

	// Topic create or update
	if err = createOrUpdateTopic(ctx, &spec.KafkaTopicInClusterConfiguration, topicName, kadmClient); err != nil {
		return err
	}
	if err != nil {
		status.Phase = ksfv1.KafkaTopicPhaseError
		return err
	}

	// Update status
	var ticc *ksfv1.KafkaTopicInClusterConfiguration
	ticc, err = getTopicInClusterConfiguration(ctx, topicName, kadmClient)
	if err != nil {
		status.Phase = ksfv1.KafkaTopicPhaseError
		return err
	}
	if ticc != nil {
		status.KafkaTopicInClusterConfiguration = *ticc
	}

	status.Phase = ksfv1.KafkaTopicPhaseAvailable

	return nil
}

// handleDeletionAndFinalizers updates finalizers if necessary and handles deletion of kafka topics
// returns false if processing should continue, true if we should finish reconcile
func handleDeletionAndFinalizers(ctx context.Context,
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
				if err := deleteTopicFromKafka(ctx, topicName, kadmClient); err != nil {
					return true, err
				}
			}
			exists, err := topicExists(ctx, topicName, kadmClient)
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

// getBrokerConfigs retrieves the Kafka broker configs from the Kafka cluster, ref: https://kafka.apache.org/documentation/#brokerconfigs
func getBrokerConfigs(ctx context.Context, kadmClient *kadm.Client) ([]kadm.Config, error) {
	allBrokerResourceConfigs, err := kadmClient.DescribeBrokerConfigs(ctx)
	if err != nil {
		return nil, err
	}
	if len(allBrokerResourceConfigs) != 1 {
		return nil, fmt.Errorf("expected to retrieve exactly 1 cluster-level broker config, but received %d", len(allBrokerResourceConfigs))
	}
	brokerResourceConfig := allBrokerResourceConfigs[0]
	if brokerResourceConfig.Err != nil {
		return nil, fmt.Errorf("received error from broker configs, err %w", brokerResourceConfig.Err)
	}
	return brokerResourceConfig.Configs, nil
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

// getDefaultTopicPartitionsAndReplicas returns the default partitions and replication-factor from the Kafka cluster
func getDefaultTopicPartitionsAndReplicas(ctx context.Context, kadmClient *kadm.Client) (*int32, *int16, error) {
	logger := log.FromContext(ctx)
	var defaultTopicPartitions *int32
	var defaultTopicReplicationFactor *int16
	bcs, err := getBrokerConfigs(ctx, kadmClient)
	if err != nil {
		return nil, nil, fmt.Errorf("error retrieving broker configs, err: %w", err)
	}
	logger.Info("config len:", "len", len(bcs))
	for _, c := range bcs {
		if c.Key == "default.replication.factor" && c.Value != nil {
			i, err := strconv.Atoi(*c.Value)
			if err != nil {
				return nil, nil, fmt.Errorf("unable to retrieve default replication factor from cluster, err: %w", err)
			}
			i16 := int16(i)
			defaultTopicReplicationFactor = &i16
		} else if c.Key == "num.partitions" && c.Value != nil {
			i, err := strconv.Atoi(*c.Value)
			if err != nil {
				return nil, nil, fmt.Errorf("unable to retrieve default partitions from cluster, err: %w", err)
			}
			defaultTopicPartitions = pointer.Int32(int32(i))
		}
	}
	if defaultTopicReplicationFactor == nil {
		return nil, nil, fmt.Errorf("default.replication.factor config missing from broker configs")
	}
	if defaultTopicPartitions == nil {
		return nil, nil, fmt.Errorf("num.partitions config missing from broker configs")
	}
	return defaultTopicPartitions, defaultTopicReplicationFactor, nil
}

// updateTopicInKafka compares the desired state (coming from spec) and the observed state (coming from the status),
// making any necessary calls to Kafka to bring them closer together.
func updateTopicInKafka(ctx context.Context,
	desired *ksfv1.KafkaTopicInClusterConfiguration,
	observed *ksfv1.KafkaTopicInClusterConfiguration,
	topicName string,
	kadmClient *kadm.Client) error {
	logger := log.FromContext(ctx)

	// Retrieve default topic partitions and replicationFactor
	defaultPartitions, defaultReplicationFactor, err := getDefaultTopicPartitionsAndReplicas(ctx, kadmClient)
	if err != nil {
		return err
	}

	// Set Partitions
	desiredPartitions := *defaultPartitions
	observedPartitions := *defaultPartitions
	if desired.Partitions != nil {
		desiredPartitions = *desired.Partitions
	}
	if observed.Partitions != nil {
		observedPartitions = *observed.Partitions
	}
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
	desiredReplicationFactor := *defaultReplicationFactor
	observedReplicationFactor := *defaultReplicationFactor
	if desired.ReplicationFactor != nil {
		desiredReplicationFactor = *desired.ReplicationFactor
	}
	if observed.ReplicationFactor != nil {
		observedReplicationFactor = *observed.ReplicationFactor
	}
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
