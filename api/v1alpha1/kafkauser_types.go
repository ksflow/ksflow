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
	"k8s.io/apimachinery/pkg/types"
)

func (ku *KafkaUser) CertificateSigningRequestName() string {
	// uses periods, since they aren't allowed for namespaces or for the name of kafka users
	return fmt.Sprintf("ksflow.ku.%s.%s", ku.Namespace, ku.Name)
}
func (ku *KafkaUser) CertificateSigningRequestNamespacedName() types.NamespacedName {
	return types.NamespacedName{Name: ku.CertificateSigningRequestName()}
}

func (ku *KafkaUser) SecretName() string {
	if ku.Spec.SecretName != nil {
		return *ku.Spec.SecretName
	}
	return fmt.Sprintf("ksflow-ku-%s", ku.Name)
}
func (ku *KafkaUser) SecretNamespacedName() types.NamespacedName {
	return types.NamespacedName{Namespace: ku.Namespace, Name: ku.SecretName()}
}

// KafkaUserSpec defines the desired state of KafkaUser
type KafkaUserSpec struct {
	// Secret name in which to store credentials, otherwise defaults to "ksflow-ku-<KAFKA_USER_NAME>".
	// +optional
	SecretName *string `json:"secretName,omitempty"`
}

// KafkaUserStatus defines the observed state of KafkaUser
type KafkaUserStatus struct {
	Phase       KsflowPhase `json:"phase,omitempty"`
	Reason      string      `json:"reason,omitempty"`
	LastUpdated metav1.Time `json:"lastUpdated,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName=ku
// +kubebuilder:printcolumn:name="Status",type=string,JSONPath=`.status.phase`
// +kubebuilder:printcolumn:name="Reason",type=string,JSONPath=`.status.reason`
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=`.metadata.creationTimestamp`

// KafkaUser is the Schema for the kafkausers API
type KafkaUser struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   KafkaUserSpec   `json:"spec,omitempty"`
	Status KafkaUserStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// KafkaUserList contains a list of KafkaUser
type KafkaUserList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []KafkaUser `json:"items"`
}

func init() {
	SchemeBuilder.Register(&KafkaUser{}, &KafkaUserList{})
}
