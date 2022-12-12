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

// ClusterKafkaConfigSpec defines the desired state of ClusterKafkaConfig
type ClusterKafkaConfigSpec struct {

	// +kubebuilder:validation:MaxLength=125
	// +kubebuilder:validation:Pattern=`^[a-z]([-a-z0-9\.]*[a-z])?$`

	// Prefix assigned to every kafka topic, resulting a topic format matching: "<my-kafkaCluster-topicPrefix>.<my-namespace>.<my-kafkaTopic-name>"
	TopicPrefix string `json:"topicPrefix"`

	// Kafka Configs for connecting to the cluster.
	Configs KafkaConfigs `json:"configs"`
}

// ClusterKafkaConfigStatus defines the observed state of ClusterKafkaConfig
type ClusterKafkaConfigStatus struct {
	Phase       ClusterKafkaConfigPhase `json:"phase,omitempty"`
	Message     string                  `json:"message,omitempty"`
	LastUpdated metav1.Time             `json:"lastUpdated,omitempty"`
}

// +kubebuilder:validation:Enum="";Available;Failed;Deleting

// ClusterKafkaConfigPhase defines the phase of the ClusterKafkaConfig
type ClusterKafkaConfigPhase string

const (
	ClusterKafkaConfigPhaseUnknown   ClusterKafkaConfigPhase = ""
	ClusterKafkaConfigPhaseAvailable ClusterKafkaConfigPhase = "Available"
	ClusterKafkaConfigPhaseFailed    ClusterKafkaConfigPhase = "Failed"
)

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,shortName=ckc
// +kubebuilder:printcolumn:name="Prefix",type=string,JSONPath=`.spec.topicPrefix`
// +kubebuilder:printcolumn:name="Protocol",type=string,JSONPath=`.spec.configs.security\.protocol`
// +kubebuilder:printcolumn:name="Status",type=string,JSONPath=`.status.phase`
// +kubebuilder:printcolumn:name="Message",type=string,JSONPath=`.status.message`
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=`.metadata.creationTimestamp`

// ClusterKafkaConfig is the Schema for the clusterkafkaconfigs API
type ClusterKafkaConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ClusterKafkaConfigSpec   `json:"spec,omitempty"`
	Status ClusterKafkaConfigStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ClusterKafkaConfigList contains a list of ClusterKafkaConfig
type ClusterKafkaConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ClusterKafkaConfig `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ClusterKafkaConfig{}, &ClusterKafkaConfigList{})
}

// BootstrapServers returns the list of bootstrap servers for the Kafka cluster
func (c *ClusterKafkaConfig) BootstrapServers() []string {
	return strings.Split(c.Spec.Configs.BootstrapServers, ",")
}
