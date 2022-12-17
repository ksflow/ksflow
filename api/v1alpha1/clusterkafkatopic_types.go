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

// FullTopicName is the actual Kafka topic name used on the Kafka cluster
func (ckt *ClusterKafkaTopic) FullTopicName() string {
	return ckt.Name
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,shortName=ckt
// +kubebuilder:printcolumn:name="Partitions",type=string,JSONPath=`.spec.partitions`
// +kubebuilder:printcolumn:name="Replicas",type=string,JSONPath=`.spec.replicationFactor`
// +kubebuilder:printcolumn:name="ReclaimPolicy",type=string,JSONPath=`.spec.reclaimPolicy`
// +kubebuilder:printcolumn:name="Status",type=string,JSONPath=`.status.phase`
// +kubebuilder:printcolumn:name="Reason",type=string,JSONPath=`.status.reason`
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=`.metadata.creationTimestamp`

// ClusterKafkaTopic is the Schema for the clusterkafkatopics API
type ClusterKafkaTopic struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   KafkaTopicSpec   `json:"spec,omitempty"`
	Status KafkaTopicStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ClusterKafkaTopicList contains a list of ClusterKafkaTopic
type ClusterKafkaTopicList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ClusterKafkaTopic `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ClusterKafkaTopic{}, &ClusterKafkaTopicList{})
}
