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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	crv1 "sigs.k8s.io/controller-runtime/pkg/config/v1alpha1"
)

// NewClient retrieves a Kafka client from the KafkaConfig
func (kcc *KafkaConnectionConfig) NewClient() (*kgo.Client, error) {
	return kgo.NewClient(kgo.SeedBrokers(kcc.BootstrapServers...))
}

type KafkaConnectionConfig struct {
	// BootstrapServers provides the list of initial Kafka brokers to connect to
	BootstrapServers []string `json:"bootstrapServers"`

	// PrincipalTemplate contains a Go Template that maps KafkaUser information to a Principal
	// Currently {{ .Name }} and {{ .Namespace }} are available within the template
	// Defaults to "User:ANONYMOUS"
	// +optional
	PrincipalTemplate *string `json:"principalTemplate"`
}

//+kubebuilder:object:root=true

// KsflowConfig is the configuration to run the controller manager, typically mounted in from ConfigMap
type KsflowConfig struct {
	metav1.TypeMeta `json:",inline"`

	crv1.ControllerManagerConfigurationSpec `json:",inline"`

	// KafkaConnectionConfig provides values for connecting to the Kafka cluster
	KafkaConnectionConfig KafkaConnectionConfig `json:"kafkaConnection"`

	// KafkaTopicDefaults provides default values for any KafkaTopic
	KafkaTopicDefaultsConfig KafkaTopicSpec `json:"kafkaTopicDefaults,omitempty"`
}

func init() {
	SchemeBuilder.Register(&KsflowConfig{})
}
