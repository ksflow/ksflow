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

// KafkaTopicSpec defines the desired state of KafkaTopic
type KafkaTopicSpec struct {
	// +kubebuilder:default=-1

	// Partitions is the number of partitions in the topic.
	Partitions int32 `json:"partitions"`

	// +kubebuilder:default=-1

	// ReplicationFactor is the number of replicas for each of the topic's partitions.
	ReplicationFactor int16 `json:"replicationFactor"`

	// Configs are the configs for the topic, see: https://kafka.apache.org/documentation/#topicconfigs
	Configs *KafkaTopicConfigs `json:"configs,omitempty"`

	// JobTemplate is used to customize the Job that runs to create/modify/delete the topic
	// +optional
	JobTemplate *JobTemplate `json:"jobTemplate,omitempty"`
}

// KafkaTopicStatus defines the observed state of KafkaTopic
type KafkaTopicStatus struct {
	Phase       KafkaTopicPhase `json:"phase,omitempty"`
	Message     string          `json:"message,omitempty"`
	LastUpdated metav1.Time     `json:"lastUpdated,omitempty"`
}

// +kubebuilder:validation:Enum="";Creating;Available;Failed;Deleting

// KafkaTopicPhase defines the phase of the KafkaTopic
type KafkaTopicPhase string

const (
	KafkaTopicPhaseUnknown   ClusterKafkaConfigPhase = ""
	KafkaTopicPhaseCreating  ClusterKafkaConfigPhase = "Creating"
	KafkaTopicPhaseDeleting  ClusterKafkaConfigPhase = "Deleting"
	KafkaTopicPhaseAvailable ClusterKafkaConfigPhase = "Available"
	KafkaTopicPhaseFailed    ClusterKafkaConfigPhase = "Failed"
)

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName=kt
// +kubebuilder:printcolumn:name="Partitions",type=string,JSONPath=`.spec.partitions`
// +kubebuilder:printcolumn:name="Replicas",type=string,JSONPath=`.spec.replicationFactor`
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
