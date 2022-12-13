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

func (kt *KafkaTopic) RawTopicName(kc *KafkaConfig) string {
	return kc.Spec.TopicPrefix + "." + kt.Namespace + "." + kt.Name
}

// KafkaTopicConfig is the part of the KafkaTopicSpec that maps directly to a real Kafka topic
type KafkaTopicConfig struct {
	// +kubebuilder:validation:Minimum=1

	// Partitions is the number of partitions in the topic.
	// +optional
	Partitions *int32 `json:"partitions,omitempty"`

	// +kubebuilder:validation:Minimum=1

	// ReplicationFactor is the number of replicas for each of the topic's partitions.
	// +optional
	ReplicationFactor *int16 `json:"replicationFactor,omitempty"`

	// Configs are the configs for the topic, see: https://kafka.apache.org/documentation/#topicconfigs
	// All values are specified as strings, since support for floating-point varies across languages
	Configs map[string]*string `json:"configs,omitempty"`
}

// KafkaTopicSpec defines the desired state of KafkaTopic
type KafkaTopicSpec struct {

	// +optional
	KafkaTopicConfig `json:",inline"`

	// +kubebuilder:default=Delete

	// ReclaimPolicy defines what should happen to the underlying kafka topic if the KafkaTopic is deleted.
	// +optional
	ReclaimPolicy *KafkaTopicReclaimPolicy `json:"reclaimPolicy,omitempty"`
}

// KafkaTopicStatus defines the observed state of KafkaTopic
type KafkaTopicStatus struct {
	KafkaTopicConfig `json:",inline"`
	KafkaConfigs     KafkaConfigs    `json:"kafkaConfigs,omitempty"`
	Phase            KafkaTopicPhase `json:"phase,omitempty"`
	Reason           string          `json:"reason,omitempty"`
	FullTopicName    string          `json:"fullTopicName,omitempty"`
	LastUpdated      metav1.Time     `json:"lastUpdated,omitempty"`
}

// +kubebuilder:validation:Enum="";Creating;Available;Failed;Deleting

// KafkaTopicPhase defines the phase of the KafkaTopic
type KafkaTopicPhase string

const (
	KafkaTopicPhaseUnknown   KafkaTopicPhase = ""
	KafkaTopicPhaseCreating  KafkaTopicPhase = "Creating"
	KafkaTopicPhaseDeleting  KafkaTopicPhase = "Deleting"
	KafkaTopicPhaseAvailable KafkaTopicPhase = "Available"
	KafkaTopicPhaseFailed    KafkaTopicPhase = "Failed"
)

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName=kt
// +kubebuilder:printcolumn:name="Partitions",type=string,JSONPath=`.spec.partitions`
// +kubebuilder:printcolumn:name="Replicas",type=string,JSONPath=`.spec.replicationFactor`
// +kubebuilder:printcolumn:name="ReclaimPolicy",type=string,JSONPath=`.spec.reclaimPolicy`
// +kubebuilder:printcolumn:name="TopicName",type=string,JSONPath=`.status.fullTopicName`,priority=10
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
