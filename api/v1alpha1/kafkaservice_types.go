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
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (ks *KafkaService) KafkaPrincipal() string {
	//TODO: impl, maybe in config provide a go template that gives access to the KafkaTopicNamespacedName?
	// i.e. "User:CN={{ .Name }}.{{ .Namespace }}.ks.cluster.local,OU=Software,O=dseapy,L=Charlottesville,ST=Virginia,C=US"
	return fmt.Sprintf("User:CN=%s.%s.ks.cluster.local", ks.Name, ks.Namespace)
}

// SecretName provides the name of the kubernetes secret to store the Service config (ssl, bootstrapServers, etc.)
func (ks *KafkaService) SecretName() string {
	return fmt.Sprintf("ksflow-ks-%s", ks.Name)
}

// KafkaTopicNamespacedName provides the namespace and name for the topic
type KafkaTopicNamespacedName struct {
	// Name of the KafkaTopic
	Name string `json:"name"`

	// Namespace of the KafkaTopic, if omitted will use same namespace as the KafkaService
	// +optional
	Namespace string `json:"namespace,omitempty"`
}

// KafkaServiceSpec defines the desired state of KafkaService
type KafkaServiceSpec struct {
	// InputTopics is the list of kafka topics that this KafkaService can read from
	// +optional
	InputTopics []KafkaTopicNamespacedName `json:"inputTopics,omitempty"`

	// OutputTopics is the list of kafka topics that this KafkaService can write to
	// +optional
	OutputTopics []KafkaTopicNamespacedName `json:"outputTopics,omitempty"`
}

// KafkaServiceStatus defines the observed state of KafkaService
type KafkaServiceStatus struct {
	Phase       KsflowPhase `json:"phase,omitempty"`
	Reason      string      `json:"reason,omitempty"`
	LastUpdated metav1.Time `json:"lastUpdated,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName=ks
// +kubebuilder:printcolumn:name="Status",type=string,JSONPath=`.status.phase`
// +kubebuilder:printcolumn:name="Reason",type=string,JSONPath=`.status.reason`
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=`.metadata.creationTimestamp`

// KafkaService is the Schema for the kafkaservices API
type KafkaService struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   KafkaServiceSpec   `json:"spec,omitempty"`
	Status KafkaServiceStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// KafkaServiceList contains a list of KafkaService
type KafkaServiceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []KafkaService `json:"items"`
}

func init() {
	SchemeBuilder.Register(&KafkaService{}, &KafkaServiceList{})
}
