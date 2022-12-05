/*
Copyright 2022.

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

type KafkaTopicSpec struct {
	// Partitions is the number of partitions in the topic.  Cannot be reduced.
	Partitions int32 `json:"partitions"`
	// ReplicationFactor is the number of replicas for each of the topic's partitions.
	ReplicationFactor int16 `json:"replicationFactor"`
	// Config is the configuration for the topic, see: https://kafka.apache.org/documentation/#topicconfigs
	Config *string `json:"config,omitempty"`
}

type KafkaTopicStatus struct {
	Phase   ExternalResourcePhase `json:"phase,omitempty"`
	Message string                `json:"message,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

type KafkaTopic struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   KafkaTopicSpec   `json:"spec,omitempty"`
	Status KafkaTopicStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

type KafkaTopicList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []KafkaTopic `json:"items"`
}

func init() {
	SchemeBuilder.Register(&KafkaTopic{}, &KafkaTopicList{})
}
