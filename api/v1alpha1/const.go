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

// +kubebuilder:validation:Enum="";Updating;Deleting;Available;Error

// KafkaTopicPhase defines the phase of the KafkaTopic
type KafkaTopicPhase string

const (
	KafkaTopicPhaseUnknown   KafkaTopicPhase = ""
	KafkaTopicPhaseUpdating  KafkaTopicPhase = "Updating"
	KafkaTopicPhaseDeleting  KafkaTopicPhase = "Deleting"
	KafkaTopicPhaseAvailable KafkaTopicPhase = "Available"
	KafkaTopicPhaseError     KafkaTopicPhase = "Error"
)

// +kubebuilder:validation:Enum=Delete;Retain

type KafkaTopicReclaimPolicy string

const (
	KafkaTopicReclaimPolicyDelete KafkaTopicReclaimPolicy = "Delete"
	KafkaTopicReclaimPolicyRetain KafkaTopicReclaimPolicy = "Retain"
)
