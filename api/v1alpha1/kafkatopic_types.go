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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +kubebuilder:validation:Enum=Delete;Retain

type KafkaTopicReclaimPolicy string

const (
	KafkaTopicReclaimPolicyDelete KafkaTopicReclaimPolicy = "Delete"
	KafkaTopicReclaimPolicyRetain KafkaTopicReclaimPolicy = "Retain"
)

// +kubebuilder:validation:Enum="";Creating;Deleting;Available;Failed

// KafkaTopicPhase defines the phase of the KafkaTopic
type KafkaTopicPhase string

const (
	KafkaTopicPhaseUnknown   KafkaTopicPhase = ""
	KafkaTopicPhaseCreating  KafkaTopicPhase = "Creating"
	KafkaTopicPhaseDeleting  KafkaTopicPhase = "Deleting"
	KafkaTopicPhaseAvailable KafkaTopicPhase = "Available"
	KafkaTopicPhaseFailed    KafkaTopicPhase = "Failed"
)

// FullTopicName is the actual Kafka topic name used on the Kafka cluster
func (kt *KafkaTopic) FullTopicName() string {
	return kt.Namespace + "." + kt.Name
}

// KafkaTopicSpec defines the desired state of KafkaTopic
type KafkaTopicSpec struct {
	// +kubebuilder:default:=Delete

	// ReclaimPolicy defines what should happen to the underlying kafka topic if the KafkaTopic is deleted.
	// +optional
	ReclaimPolicy *KafkaTopicReclaimPolicy `json:"reclaimPolicy,omitempty"`

	// +kubebuilder:validation:Minimum=1

	// Partitions is the number of partitions in the topic.
	// +optional
	Partitions *int32 `json:"partitions,omitempty"`

	// +kubebuilder:validation:Minimum=1

	// ReplicationFactor is the number of replicas for each of the topic's partitions.
	// +optional
	ReplicationFactor *int16 `json:"replicationFactor,omitempty"`

	// Configs contains the configs for the topic, see: https://kafka.apache.org/documentation/#topicconfigs
	// All values are specified as strings
	// +optional
	Configs map[string]*string `json:"configs,omitempty"`

	// +kubebuilder:default:={configMapKeyRef:{name: ksflow-kafka-configs, key: topic.properties, optional: true}}

	// ConfigsFrom is the source of the configs' value. Expects a properties file containing the configs
	// Values defined in Configs with a duplicate key will take precedence.
	// +optional
	ConfigsFrom *ConfigsSource `json:"configsFrom,omitempty"`
}

// KafkaTopicStatus defines the observed state of KafkaTopic
type KafkaTopicStatus struct {
	Phase       KafkaTopicPhase `json:"phase,omitempty"`
	Reason      string          `json:"reason,omitempty"`
	LastUpdated metav1.Time     `json:"lastUpdated,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName=kt
// +kubebuilder:printcolumn:name="Partitions",type=string,JSONPath=`.spec.partitions`
// +kubebuilder:printcolumn:name="Replicas",type=string,JSONPath=`.spec.replicationFactor`
// +kubebuilder:printcolumn:name="ReclaimPolicy",type=string,JSONPath=`.spec.reclaimPolicy`
// +kubebuilder:printcolumn:name="Status",type=string,JSONPath=`.status.phase`
// +kubebuilder:printcolumn:name="Reason",type=string,JSONPath=`.status.reason`
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=`.metadata.creationTimestamp`

// KafkaTopic is the Schema for the kafkatopics API
type KafkaTopic struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   KafkaTopicSpec   `json:"spec,omitempty"`
	Status KafkaTopicStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// KafkaTopicList contains a list of KafkaTopic
type KafkaTopicList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []KafkaTopic `json:"items"`
}

func init() {
	SchemeBuilder.Register(&KafkaTopic{}, &KafkaTopicList{})
}
