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
	"github.com/twmb/franz-go/pkg/kmsg"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type KafkaACLSpec struct {
	// User is the name of the KafkaUser for which permissions are applied (i.e. "my-namespace.my-user", "my-user")
	User string `json:"user"`
	// Topic is the name of the KafkaTopic in this namespace for the ACL (i.e. "my-topic", "*", "my-")
	Topic string `json:"topic"`
	// Operation is the operation (i.e. "read", "write", "describe")
	Operation kmsg.ACLOperation `json:"operation"`
	// PermissionType is the permission type (i.e. "ALLOW", "DENY")
	PermissionType kmsg.ACLPermissionType `json:"permissionType"`
	// PatternType is the pattern type (i.e. "any", "literal", "prefixed")
	PatternType kmsg.ACLResourcePatternType `json:"patternType"`
}

type KafkaACLStatus struct {
	Phase     ExternalResourcePhase `json:"phase,omitempty"`
	Message   string                `json:"message,omitempty"`
	Principal string                `json:"principal,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

type KafkaACL struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   KafkaACLSpec   `json:"spec,omitempty"`
	Status KafkaACLStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

type KafkaACLList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []KafkaACL `json:"items"`
}

func init() {
	SchemeBuilder.Register(&KafkaACL{}, &KafkaACLList{})
}
