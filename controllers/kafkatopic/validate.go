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

package kafkatopic

import (
	"fmt"

	"github.com/ksflow/ksflow/api/v1alpha1"
	"github.com/ksflow/ksflow/validation"
)

// Validate ensures the KafkaTopic is valid
func Validate(kt *v1alpha1.KafkaTopic) error {
	if err := validation.ValidateRFC1035LabelName(kt.Name); err != nil {
		return fmt.Errorf(`invalid name: %v`, err)
	}
	if kt.Spec.ReplicationFactor < 1 {
		return fmt.Errorf(`invalid spec: replicationFactor must be at least 1, was %d`, kt.Spec.ReplicationFactor)
	}
	if kt.Spec.Partitions < 1 {
		return fmt.Errorf(`invalid spec: partitions must be at least 1, was %d`, kt.Spec.Partitions)
	}
	return nil
}
