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
	"github.com/twmb/franz-go/pkg/kgo"
	"github.com/twmb/franz-go/pkg/sr"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	crv1 "sigs.k8s.io/controller-runtime/pkg/config/v1alpha1"
)

// NewClient retrieves a Kafka client from the KafkaConnectionConfig
func (kcc *KafkaConnectionConfig) NewClient() (*kgo.Client, error) {
	return kgo.NewClient(kgo.SeedBrokers(kcc.BootstrapServers...))
}

// NewClient retrieves a Schema Registry client from the SchemaRegistryConnectionConfig
func (srcc *SchemaRegistryConnectionConfig) NewClient() (*sr.Client, error) {
	return sr.NewClient(sr.URLs(srcc.URLs...))
}

type KafkaConnectionConfig struct {
	// BootstrapServers provides the list of initial Kafka brokers to connect to
	BootstrapServers []string `json:"bootstrapServers"`
}

type SchemaRegistryConnectionConfig struct {
	// URLs provides the endpoint URLs for the schema registry
	URLs []string `json:"urls"`
}

type KafkaTopicConfig struct {
	// KafkaConnectionConfig is config required to create a kafka client that connects to the Kafka cluster
	KafkaConnectionConfig KafkaConnectionConfig `json:"kafkaConnection"`

	// NameTemplate contains a Go Template that maps a KafkaTopic to a kafka topic.
	// Currently {{ .Name }} and {{ .Namespace }} are available within the template
	// and are expected to be used to create a unique topic for the KafkaTopic.
	// i.e. "{{ .Namespace }}.{{ .Name }}"
	// ref: https://kafka.apache.org/documentation/#multitenancy-topic-naming
	NameTemplate string `json:"nameTemplate"`

	// KafkaTopicDefaults provides default values for any KafkaTopic
	KafkaTopicDefaultsConfig KafkaTopicSpec `json:"defaults,omitempty"`
}

type KafkaSchemaConfig struct {
	// SchemaRegistryConnectionConfig is config required to create a schema registry client
	SchemaRegistryConnectionConfig SchemaRegistryConnectionConfig `json:"schemaRegistryConnection"`

	// NameTemplate contains a Go Template that maps a KafkaSchema to a kafka schema.
	// Currently {{ .Name }} and {{ .Namespace }} are available within the template
	// and are expected to be used to create a unique schema for the KafkaSchema.
	// i.e. "{{ .Namespace }}.{{ .Name }}"
	NameTemplate string `json:"nameTemplate"`
}

type KafkaUserConfig struct {
	// NameTemplate contains a Go Template that maps a KafkaUser to a kafka user.
	// Currently {{ .Name }} and {{ .Namespace }} are available within the template
	// and are expected to be used to create a unique user for the KafkaUser.
	// The principal type is assumed to be "User:", and should NOT be included in the nameTemplate
	// i.e. "CN={{ .Name }}.{{ .Namespace }}.svc,OU=TEST,O=Marketing,L=Charlottesville,ST=Va,C=US"
	NameTemplate string `json:"nameTemplate"`
}

//+kubebuilder:object:root=true

// KsflowConfig is the configuration to run the controller manager, typically mounted in from ConfigMap
type KsflowConfig struct {
	metav1.TypeMeta `json:",inline"`

	crv1.ControllerManagerConfigurationSpec `json:",inline"`

	// KafkaTopicConfig contains the configuration necessary to run the KafkaTopic controller
	KafkaTopicConfig KafkaTopicConfig `json:"kafkaTopic"`

	// KafkaUserConfig contains the configuration necessary to run the KafkaUser controller
	KafkaUserConfig KafkaUserConfig `json:"kafkaUser"`

	// KafkaSchemaConfig contains the configuration necessary to run the KafkaSchema controller
	KafkaSchemaConfig KafkaSchemaConfig `json:"kafkaSchema"`
}

func init() {
	SchemeBuilder.Register(&KsflowConfig{})
}
