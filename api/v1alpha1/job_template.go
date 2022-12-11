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

// File modified from https://github.com/numaproj/numaflow/blob/cc44875b2a2e01694ceca1fb085c3423bd330a38/pkg/apis/numaflow/v1alpha1/job_template.go

package v1alpha1

type JobTemplate struct {
	// +optional
	AbstractPodTemplate `json:",inline" protobuf:"bytes,1,opt,name=abstractPodTemplate"`
	// +optional
	ContainerTemplate *ContainerTemplate `json:"containerTemplate,omitempty"`
	// TTLSecondsAfterFinished limits the lifetime of a Job that has finished
	// execution (either Complete or Failed). If this field is set,
	// ttlSecondsAfterFinished after the Job finishes, it is eligible to be
	// automatically deleted. When the Job is being deleted, its lifecycle
	// guarantees (e.g. finalizers) will be honored. If this field is unset,
	// the Job won't be automatically deleted. If this field is set to zero,
	// the Job becomes eligible to be deleted immediately after it finishes.
	// +optional
	TTLSecondsAfterFinished *int32 `json:"ttlSecondsAfterFinished,omitempty"`
	// BackoffLimit specifies the number of retries before marking this job failed.
	// More info: https://kubernetes.io/docs/concepts/workloads/controllers/job/#pod-backoff-failure-policy
	// +optional
	BackoffLimit *int32 `json:"backoffLimit,omitempty"`
}
