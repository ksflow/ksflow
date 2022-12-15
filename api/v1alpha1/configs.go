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

import v1 "k8s.io/api/core/v1"

// ConfigsSource represents a source for the value of configs.
type ConfigsSource struct {
	// Selects the key of a ConfigMap containing the configs for the topic in property file format, see: https://kafka.apache.org/documentation/#topicconfigs
	// Same behavior as in a container, see: https://github.com/kubernetes/website/blob/snapshot-initial-v1.26/content/en/examples/pods/pod-single-configmap-env-variable.yaml#L14
	// +optional
	ConfigMapKeyRef *v1.ConfigMapKeySelector `json:"configMapKeyRef,omitempty"`
}
