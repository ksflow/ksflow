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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type KafkaAppSpec struct {
	//TODO
	Foo string `json:"foo"`
}

type KafkaAppStatus struct {
	Phase     ExternalResourcePhase `json:"phase,omitempty"`
	Message   string                `json:"message,omitempty"`
	Principal string                `json:"principal,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

type KafkaApp struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   KafkaAppSpec   `json:"spec,omitempty"`
	Status KafkaAppStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

type KafkaAppList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []KafkaApp `json:"items"`
}

func init() {
	SchemeBuilder.Register(&KafkaApp{}, &KafkaAppList{})
}
