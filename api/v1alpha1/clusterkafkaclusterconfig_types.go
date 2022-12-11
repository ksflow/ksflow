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
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +kubebuilder:validation:Enum=PLAINTEXT;SSL

type KafkaSecurityProtocol string

const (
	KafkaSecurityProtocolPlaintext KafkaSecurityProtocol = "PLAINTEXT"
	KafkaSecurityProtocolSSL       KafkaSecurityProtocol = "SSL"
)

// KafkaClusterConfigs defines the Kafka configs for connecting to the Kafka Cluster
// configs match what kafka clients use to connect, however they are camelCase instead of period-separated.
type KafkaClusterConfigs struct {

	// +kubebuilder:validation:Pattern=`^([^,:]+):(\d+)(,([^,:]+):(\d+))*$`

	// The "bootstrap.servers" kafka config for connecting to the cluster. For example "host1:port1,host2:port2".
	BootstrapServers string `json:"bootstrapServers"`

	// The "security.protocol" kafka config for connecting to the cluster, must be "PLAINTEXT" or "SSL".
	SecurityProtocol KafkaSecurityProtocol `json:"securityProtocol"`
}

// ClusterKafkaClusterConfigSpec defines the desired state of ClusterKafkaClusterConfig
type ClusterKafkaClusterConfigSpec struct {

	// +kubebuilder:validation:MaxLength=125
	// +kubebuilder:validation:Pattern=`^[a-z]([-a-z0-9\.]*[a-z])?$`

	// Prefix assigned to every kafka topic, resulting a topic format matching: "<my-kafkaCluster-topicPrefix>.<my-namespace>.<my-kafkaTopic-name>"
	TopicPrefix string `json:"topicPrefix"`

	// Kafka Configs for connecting to the cluster.
	Configs KafkaClusterConfigs `json:"configs"`
}

// ClusterKafkaClusterConfigStatus defines the observed state of ClusterKafkaClusterConfig
type ClusterKafkaClusterConfigStatus struct {
	Phase       ClusterKafkaClusterConfigPhase `json:"phase,omitempty"`
	Message     string                         `json:"message,omitempty"`
	LastUpdated metav1.Time                    `json:"lastUpdated,omitempty"`
}

//+kubebuilder:validation:Enum="";Available;Failed;Deleting

// ClusterKafkaClusterConfigPhase defines the phase of the ClusterKafkaClusterConfig
type ClusterKafkaClusterConfigPhase string

const (
	ClusterKafkaClusterConfigPhaseUnknown   ClusterKafkaClusterConfigPhase = ""
	ClusterKafkaClusterConfigPhaseAvailable ClusterKafkaClusterConfigPhase = "Available"
	ClusterKafkaClusterConfigPhaseFailed    ClusterKafkaClusterConfigPhase = "Failed"
	ClusterKafkaClusterConfigPhaseDeleting  ClusterKafkaClusterConfigPhase = "Deleting"
)

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:scope=Cluster,shortName=ckcc
//+kubebuilder:printcolumn:name="Prefix",type=string,JSONPath=`.spec.topicPrefix`
//+kubebuilder:printcolumn:name="Protocol",type=string,JSONPath=`.spec.configs.securityProtocol`
//+kubebuilder:printcolumn:name="Status",type=string,JSONPath=`.status.phase`
//+kubebuilder:printcolumn:name="Message",type=string,JSONPath=`.status.message`
//+kubebuilder:printcolumn:name="Age",type="date",JSONPath=`.metadata.creationTimestamp`

// ClusterKafkaClusterConfig is the Schema for the clusterkafkaclusterconfigs API
type ClusterKafkaClusterConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ClusterKafkaClusterConfigSpec   `json:"spec,omitempty"`
	Status ClusterKafkaClusterConfigStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ClusterKafkaClusterConfigList contains a list of ClusterKafkaClusterConfig
type ClusterKafkaClusterConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ClusterKafkaClusterConfig `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ClusterKafkaClusterConfig{}, &ClusterKafkaClusterConfigList{})
}

// BootstrapServers returns the list of bootstrap servers for the Kafka cluster
func (c *ClusterKafkaClusterConfig) BootstrapServers() []string {
	return strings.Split(c.Spec.Configs.BootstrapServers, ",")
}
