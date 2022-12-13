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
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +kubebuilder:validation:Enum=PLAINTEXT;SSL

type KafkaSecurityProtocol string

const (
	KafkaSecurityProtocolPlaintext KafkaSecurityProtocol = "PLAINTEXT"
	KafkaSecurityProtocolSSL       KafkaSecurityProtocol = "SSL"
)

// KafkaConfigs defines the Kafka configs for connecting to the Kafka Cluster
// configs match what kafka clients use to connect.
type KafkaConfigs struct {

	// +kubebuilder:validation:Pattern=`^([^,:]+):(\d+)(,([^,:]+):(\d+))*$`

	// The "bootstrap.servers" kafka config for connecting to the cluster. For example "host1:port1,host2:port2".
	BootstrapServers string `json:"bootstrap.servers"`

	// The "security.protocol" kafka config for connecting to the cluster, must be "PLAINTEXT" or "SSL".
	SecurityProtocol KafkaSecurityProtocol `json:"security.protocol"`
}

// KafkaConfigSpec defines the desired state of KafkaConfig
type KafkaConfigSpec struct {

	// +kubebuilder:validation:MaxLength=125
	// +kubebuilder:validation:Pattern=`^[a-z]([-a-z0-9\.]*[a-z])?$`

	// Prefix assigned to every kafka topic, resulting a topic format matching: "<my-kafkaCluster-topicPrefix>.<my-namespace>.<my-kafkaTopic-name>"
	TopicPrefix string `json:"topicPrefix"`

	// Kafka Configs for connecting to the cluster.
	Configs KafkaConfigs `json:"configs"`
}

// KafkaConfigStatus defines the observed state of KafkaConfig
type KafkaConfigStatus struct {
	Phase       KafkaConfigPhase `json:"phase,omitempty"`
	Reason      string           `json:"reason,omitempty"`
	LastUpdated metav1.Time      `json:"lastUpdated,omitempty"`
}

// +kubebuilder:validation:Enum="";Available;Failed;Deleting

// KafkaConfigPhase defines the phase of the KafkaConfig
type KafkaConfigPhase string

const (
	KafkaConfigPhaseUnknown   KafkaConfigPhase = ""
	KafkaConfigPhaseAvailable KafkaConfigPhase = "Available"
	KafkaConfigPhaseFailed    KafkaConfigPhase = "Failed"
	KafkaConfigPhaseDeleting  KafkaConfigPhase = "Deleting"
)

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,shortName=kc
// +kubebuilder:printcolumn:name="Prefix",type=string,JSONPath=`.spec.topicPrefix`
// +kubebuilder:printcolumn:name="Protocol",type=string,JSONPath=`.spec.configs.security\.protocol`
// +kubebuilder:printcolumn:name="Status",type=string,JSONPath=`.status.phase`
// +kubebuilder:printcolumn:name="Reason",type=string,JSONPath=`.status.reason`
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=`.metadata.creationTimestamp`

// KafkaConfig is the Schema for the kafkaconfigs API
type KafkaConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   KafkaConfigSpec   `json:"spec,omitempty"`
	Status KafkaConfigStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// KafkaConfigList contains a list of KafkaConfig
type KafkaConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []KafkaConfig `json:"items"`
}

func init() {
	SchemeBuilder.Register(&KafkaConfig{}, &KafkaConfigList{})
}

// BootstrapServers returns the list of bootstrap servers for the Kafka cluster
func (c *KafkaConfig) BootstrapServers() []string {
	return strings.Split(c.Spec.Configs.BootstrapServers, ",")
}
