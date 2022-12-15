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
	crv1 "sigs.k8s.io/controller-runtime/pkg/config/v1alpha1"
)

type KafkaTLSConfig struct {
	// CertFilePath is the path to the file holding the client-side TLS certificate to use
	CertFilePath string `json:"cert"`

	// KeyFilePath is the path to the file holding the client’s private key
	KeyFilePath string `json:"key"`

	// CAFilePath is the path to the file containing certificate authority certificates to use
	// in verifying a presented server certificate
	CAFilePath string `json:"ca"`
}

type KafkaConfig struct {
	// BootstrapServers provides the list of initial Kafka brokers to connect to
	BootstrapServers []string `json:"bootstrapServers"`

	// KafkaTLSConfig provides certificates for connecting to Kafka over mutual TLS
	KafkaTLSConfig KafkaTLSConfig `json:"tls"`
}

//+kubebuilder:object:root=true

// KsflowConfig is the configuration to run the controller manager, typically mounted in from ConfigMap
type KsflowConfig struct {
	metav1.TypeMeta `json:",inline"`

	crv1.ControllerManagerConfigurationSpec `json:",inline"`

	KafkaConfig KafkaConfig `json:"kafka"`
}

func init() {
	SchemeBuilder.Register(&KsflowConfig{})
}
