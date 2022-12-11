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

// File modified from https://github.com/numaproj/numaflow/blob/cc44875b2a2e01694ceca1fb085c3423bd330a38/pkg/apis/numaflow/v1alpha1/container_template.go

package v1alpha1

import corev1 "k8s.io/api/core/v1"

// ContainerTemplate defines customized spec for a container
type ContainerTemplate struct {
	Resources       corev1.ResourceRequirements `json:"resources,omitempty"`
	ImagePullPolicy corev1.PullPolicy           `json:"imagePullPolicy,omitempty"`
	SecurityContext *corev1.SecurityContext     `json:"securityContext,omitempty"`
	Env             []corev1.EnvVar             `json:"env,omitempty"`
}
