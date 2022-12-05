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
	cfg "sigs.k8s.io/controller-runtime/pkg/config/v1alpha1"
)

type KafkaTopicConfig struct {
	//TODO
}
type KafkaProducerConfig struct {
	//TODO
}
type KafkaConsumerConfig struct {
	//TODO
}
type KafkaAdminConfig struct {
	//TODO
}
type KafkaSchemaRegistryConfig struct {
	SchemaRegistryURL string `json:"schema.registry.url,omitempty"`
}

type GlobalKafkaUserConfig struct {
	// GlobalKafkaUserCertificateTemplate is a Go Template for defining a cert-manager Certificate from a KafkaUser
	// "{{ .User.Namespace }}" and "{{ .User.Name }}" should be used to generate the certificate
	GlobalKafkaUserCertificateTemplate string `json:"certificateTemplate"`
}

type GlobalKafkaACLConfig struct {
	// GlobalKafkaACLPrincipalTemplate is a Go Template for defining the ACL principal from a KafkaACL's user
	// "{{ .User.Namespace }}" and "{{ .User.Name }}" should be used to generate the principal
	GlobalKafkaACLPrincipalTemplate string `json:"principalTemplate"`
}

type KafkaConfigs struct {
	// Topic is configuration for a KafkaTopic
	// ref: https://kafka.apache.org/documentation/#topicconfigs
	Topic KafkaTopicConfig `json:"topic"`

	// Producer is configuration for a Kafka producer
	// ref: https://kafka.apache.org/documentation/#producerconfigs
	Producer KafkaProducerConfig `json:"producer"`

	// Consumer is configuration for a Kafka consumer
	// ref: https://kafka.apache.org/documentation/#consumerconfigs
	Consumer KafkaConsumerConfig `json:"consumer"`

	// Admin is configuration for an Admin client
	// ref: https://kafka.apache.org/documentation/#adminclientconfigs
	Admin KafkaAdminConfig `json:"admin"`

	// SchemaRegistry is configuration for schema registry
	// ref: https://github.com/confluentinc/schema-registry/blob/3ef56384414e182bd8b724c4216d2d70e6eb7a24/schema-serializer/src/main/java/io/confluent/kafka/serializers/AbstractKafkaSchemaSerDeConfig.java
	SchemaRegistry KafkaSchemaRegistryConfig `json:"schemaRegistry"`
}

type GlobalKafkaConfig struct {
	// TopicPrefix is the prefix added to every topic specified in Kubernetes when applying to Kafka.
	TopicPrefix string `json:"topicPrefix"`

	// Configs are configurations for kafka-related things (i.e. clients, topics, etc)
	// ref: https://kafka.apache.org/documentation/#configuration
	Configs KafkaConfigs `json:"configs"`
}

//+kubebuilder:object:root=true

type KsflowConfig struct {
	metav1.TypeMeta `json:",inline"`

	cfg.ControllerManagerConfigurationSpec `json:",inline"`

	// GlobalKafkaConfig are Kafka-related settings.
	GlobalKafkaConfig GlobalKafkaConfig `json:"kafka"`
}

func init() {
	SchemeBuilder.Register(&KsflowConfig{})
}
