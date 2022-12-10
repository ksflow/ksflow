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

// +kubebuilder:validation:Enum=PLAINTEXT;SSL

type KafkaSecurityProtocol string

const (
	KafkaSecurityProtocolPlaintext KafkaSecurityProtocol = "PLAINTEXT"
	KafkaSecurityProtocolSSL       KafkaSecurityProtocol = "SSL"
)

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
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:scope=Cluster,shortName=ckcc
//+kubebuilder:printcolumn:name="Topic Prefix",type=string,JSONPath=`.spec.topicPrefix`
//+kubebuilder:printcolumn:name="Security Protocol",type=string,JSONPath=`.spec.securityProtocol`
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
