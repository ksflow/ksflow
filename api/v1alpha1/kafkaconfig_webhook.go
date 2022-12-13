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
var kafkaconfiglog = logf.Log.WithName("kafkaconfig-resource")

func (r *KafkaConfig) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

//+kubebuilder:webhook:path=/validate-ksflow-io-v1alpha1-kafkaconfig,mutating=false,failurePolicy=fail,sideEffects=None,groups=ksflow.io,resources=kafkaconfigs,verbs=update,versions=v1alpha1,name=vkafkaconfig.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &KafkaConfig{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *KafkaConfig) ValidateCreate() error {
	kafkaconfiglog.Info("validate create", "name", r.Name)
	// unused, update the kubebuilder annotation to enable
	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *KafkaConfig) ValidateUpdate(old runtime.Object) error {
	kafkaconfiglog.Info("validate update", "name", r.Name)

	var allErrs field.ErrorList

	oldKafkaConfig, _ := old.(*KafkaConfig)
	if r.Spec.TopicPrefix != oldKafkaConfig.Spec.TopicPrefix {
		topicPrefixFieldPath := field.NewPath("spec").Child("topicPrefix")
		errStr := fmt.Sprintf("field is immutable")
		allErrs = append(allErrs, field.Invalid(topicPrefixFieldPath, r.Spec.TopicPrefix, errStr))
	}
	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *KafkaConfig) ValidateDelete() error {
	kafkaconfiglog.Info("validate delete", "name", r.Name)
	// unused, update the kubebuilder annotation to enable
	return nil
}
