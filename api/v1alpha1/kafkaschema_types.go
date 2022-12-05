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
	"github.com/twmb/franz-go/pkg/sr"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type KafkaSchemaReference struct {
	Name    string `json:"name"`
	Subject string `json:"subject"`
	Version int    `json:"version"`
}

type KafkaSchemaSpec struct {
	// Schema is the actual unescaped text of a schema.
	Schema string `json:"schema"`

	// Format is the format of a schema.
	Format sr.SchemaType `json:"format"`

	// References declares other schemas this schema references.
	References []KafkaSchemaReference `json:"references,omitempty"`

	// CompatibilityLevel is the compatibility level requirement among the schema versions.
	CompatibilityLevel *sr.CompatibilityLevel `json:"compatibilityLevel,omitempty"`
}

type KafkaSchemaStatus struct {
	Phase         ExternalResourcePhase `json:"phase,omitempty"`
	Message       string                `json:"message,omitempty"`
	SchemaVersion *int                  `json:"schemaVersion,omitempty"`
	SchemaID      *int                  `json:"schemaID,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

type KafkaSchema struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   KafkaSchemaSpec   `json:"spec,omitempty"`
	Status KafkaSchemaStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

type KafkaSchemaList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []KafkaSchema `json:"items"`
}

func init() {
	SchemeBuilder.Register(&KafkaSchema{}, &KafkaSchemaList{})
}
