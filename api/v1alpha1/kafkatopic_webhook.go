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

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
var kafkatopiclog = logf.Log.WithName("kafkatopic-resource")

func (r *KafkaTopic) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

//+kubebuilder:webhook:path=/validate-ksflow-io-v1alpha1-kafkatopic,mutating=false,failurePolicy=fail,sideEffects=None,groups=ksflow.io,resources=kafkatopics,verbs=update,versions=v1alpha1,name=vkafkatopic.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &KafkaTopic{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *KafkaTopic) ValidateCreate() error {
	kafkatopiclog.Info("validate create", "name", r.Name)
	// unused, update the kubebuilder annotation to enable
	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *KafkaTopic) ValidateUpdate(old runtime.Object) error {
	kafkatopiclog.Info("validate update", "name", r.Name)

	var allErrs field.ErrorList

	oldKafkaTopic, _ := old.(*KafkaTopic)
	if r.Spec.ReplicationFactor != oldKafkaTopic.Spec.ReplicationFactor {
		replicationFactorFieldPath := field.NewPath("spec").Child("replicationFactor")
		errStr := fmt.Sprintf("modifying this field is not yet supported")
		allErrs = append(allErrs, field.Invalid(replicationFactorFieldPath, r.Spec.ReplicationFactor, errStr))
	}
	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *KafkaTopic) ValidateDelete() error {
	kafkatopiclog.Info("validate delete", "name", r.Name)
	// unused, update the kubebuilder annotation to enable
	return nil
}
