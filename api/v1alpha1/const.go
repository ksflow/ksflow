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

import "github.com/twmb/franz-go/pkg/sr"

// +kubebuilder:validation:Enum="";Updating;Deleting;Available;Error

// KsflowPhase defines the phase
type KsflowPhase string

const (
	KsflowPhaseUnknown   KsflowPhase = ""
	KsflowPhaseCreating  KsflowPhase = "Creating"
	KsflowPhaseUpdating  KsflowPhase = "Updating"
	KsflowPhaseDeleting  KsflowPhase = "Deleting"
	KsflowPhaseAvailable KsflowPhase = "Available"
	KsflowPhaseError     KsflowPhase = "Error"
)

// +kubebuilder:validation:Enum=AVRO;JSON;PROTOBUF

// KafkaSchemaType defines the schema type
type KafkaSchemaType string

const (
	KafkaSchemaTypeAvro     KafkaSchemaType = "AVRO"
	KafkaSchemaTypeJson     KafkaSchemaType = "JSON"
	KafkaSchemaTypeProtobuf KafkaSchemaType = "PROTOBUF"
)

func (kst KafkaSchemaType) ToFranz() sr.SchemaType {
	switch kst {
	case KafkaSchemaTypeJson:
		return sr.TypeJSON
	case KafkaSchemaTypeProtobuf:
		return sr.TypeProtobuf
	default:
		return sr.TypeAvro
	}
}

// +kubebuilder:validation:Enum=BACKWARD;BACKWARD_TRANSITIVE;FORWARD;FORWARD_TRANSITIVE;FULL;FULL_TRANSITIVE;NONE

// KafkaSchemaCompatibilityLevel defines the compatibility type
type KafkaSchemaCompatibilityLevel string

const (
	KafkaSchemaCompatibilityLevelBackward           KafkaSchemaCompatibilityLevel = "BACKWARD"
	KafkaSchemaCompatibilityLevelBackwardTransitive KafkaSchemaCompatibilityLevel = "BACKWARD_TRANSITIVE"
	KafkaSchemaCompatibilityLevelForward            KafkaSchemaCompatibilityLevel = "FORWARD"
	KafkaSchemaCompatibilityLevelForwardTransitive  KafkaSchemaCompatibilityLevel = "FORWARD_TRANSITIVE"
	KafkaSchemaCompatibilityLevelFull               KafkaSchemaCompatibilityLevel = "FULL"
	KafkaSchemaCompatibilityLevelFullTransitive     KafkaSchemaCompatibilityLevel = "FULL_TRANSITIVE"
	KafkaSchemaCompatibilityLevelNone               KafkaSchemaCompatibilityLevel = "NONE"
)

func (kscl KafkaSchemaCompatibilityLevel) ToFranz() sr.CompatibilityLevel {
	switch kscl {
	case KafkaSchemaCompatibilityLevelBackwardTransitive:
		return sr.CompatBackwardTransitive
	case KafkaSchemaCompatibilityLevelForward:
		return sr.CompatForward
	case KafkaSchemaCompatibilityLevelForwardTransitive:
		return sr.CompatForwardTransitive
	case KafkaSchemaCompatibilityLevelFull:
		return sr.CompatFull
	case KafkaSchemaCompatibilityLevelFullTransitive:
		return sr.CompatFullTransitive
	case KafkaSchemaCompatibilityLevelNone:
		return sr.CompatNone
	default:
		return sr.CompatBackward
	}
}

// +kubebuilder:validation:Enum=IMPORT;READONLY;READWRITE

// KafkaSchemaMode defines the mode
type KafkaSchemaMode string

const (
	KafkaSchemaModeImport    KafkaSchemaMode = "IMPORT"
	KafkaSchemaModeReadOnly  KafkaSchemaMode = "READONLY"
	KafkaSchemaModeReadWrite KafkaSchemaMode = "READWRITE"
)

func (ksm KafkaSchemaMode) ToFranz() sr.Mode {
	switch ksm {
	case KafkaSchemaModeImport:
		return sr.ModeImport
	case KafkaSchemaModeReadOnly:
		return sr.ModeReadOnly
	default:
		return sr.ModeReadWrite
	}
}
