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
	"bytes"
	"text/template"

	"github.com/twmb/franz-go/pkg/sr"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

// FinalName is the actual Kafka schema name used on the Kafka cluster
func (ku *KafkaSchema) FinalName(tpl *template.Template) (string, error) {
	var tplBytes bytes.Buffer
	if err := tpl.Execute(&tplBytes, types.NamespacedName{Namespace: ku.Namespace, Name: ku.Name}); err != nil {
		return "", err
	}
	return tplBytes.String(), nil
}

// KafkaSubjectInClusterConfiguration is the part of the KafkaSchemaSpec that maps directly to a real SR subject
type KafkaSubjectInClusterConfiguration struct {
	Schema string `json:"schema"`

	// +kubebuilder:default:=AVRO

	// +optional
	Type KafkaSchemaType `json:"type,omitempty"`

	// +kubebuilder:default:=BACKWARD

	// +optional
	CompatibilityLevel KafkaSchemaCompatibilityLevel `json:"compatibilityLevel,omitempty"`

	// +kubebuilder:default:=READWRITE

	// +optional
	Mode KafkaSchemaMode `json:"mode,omitempty"`

	// +optional
	References []sr.SchemaReference `json:"references,omitempty"`
}

// KafkaSchemaSpec defines the desired state of KafkaSchema
type KafkaSchemaSpec struct {
	KafkaSubjectInClusterConfiguration `json:",inline"`
}

// KafkaSchemaStatus defines the observed state of KafkaSchema
type KafkaSchemaStatus struct {
	KafkaSubjectInClusterConfiguration `json:",inline"`

	SubjectName string      `json:"subjectName,omitempty"`
	SchemaCount *int32      `json:"schemaCount,omitempty"`
	Phase       KsflowPhase `json:"phase,omitempty"`
	Reason      string      `json:"reason,omitempty"`
	LastUpdated metav1.Time `json:"lastUpdated,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName=ks
// +kubebuilder:printcolumn:name="Subject",type=string,JSONPath=`.status.subjectName`
// +kubebuilder:printcolumn:name="Versions",type=string,JSONPath=`.status.schemaCount`
// +kubebuilder:printcolumn:name="Type",type=string,JSONPath=`.status.type`
// +kubebuilder:printcolumn:name="Compatibility",type=string,JSONPath=`.status.compatibilityLevel`
// +kubebuilder:printcolumn:name="Mode",type=string,JSONPath=`.status.mode`
// +kubebuilder:printcolumn:name="Status",type=string,JSONPath=`.status.phase`
// +kubebuilder:printcolumn:name="Reason",type=string,JSONPath=`.status.reason`
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=`.metadata.creationTimestamp`

// KafkaSchema is the Schema for the kafkaschemas API
type KafkaSchema struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   KafkaSchemaSpec   `json:"spec,omitempty"`
	Status KafkaSchemaStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// KafkaSchemaList contains a list of KafkaSchema
type KafkaSchemaList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []KafkaSchema `json:"items"`
}

func init() {
	SchemeBuilder.Register(&KafkaSchema{}, &KafkaSchemaList{})
}
