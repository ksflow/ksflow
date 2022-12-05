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

package kafkaacl

import (
	"fmt"

	"github.com/dseapy/api/v1alpha1"
	"github.com/dseapy/validation"
	"github.com/twmb/franz-go/pkg/kmsg"
)

// Validate ensures the KafkaACL is valid
func Validate(ka *v1alpha1.KafkaACL) error {
	if err := validation.ValidateRFC1035LabelName(ka.Name); err != nil {
		return fmt.Errorf(`invalid name: %v`, err)
	}
	if ka.Spec.Operation != kmsg.ACLOperationRead &&
		ka.Spec.Operation != kmsg.ACLOperationWrite &&
		ka.Spec.Operation != kmsg.ACLOperationDescribe {
		return fmt.Errorf(`invalid spec: only "read", "write", and "describe" ACLs are allowed`)
	}
	if ka.Spec.PermissionType != kmsg.ACLPermissionTypeAllow &&
		ka.Spec.PermissionType != kmsg.ACLPermissionTypeDeny {
		return fmt.Errorf(`invalid spec: only "ALLOW" and "DENY" ACLs are allowed`)
	}
	if ka.Spec.PatternType != kmsg.ACLResourcePatternTypeAny &&
		ka.Spec.PatternType != kmsg.ACLResourcePatternTypeLiteral &&
		ka.Spec.PatternType != kmsg.ACLResourcePatternTypePrefixed {
		return fmt.Errorf(`invalid spec: only "any", "literal" and "prefixed" ACLs are allowed`)
	}
	if err := validation.ValidateRFC1035LabelName(ka.Spec.Topic); err != nil {
		return fmt.Errorf(`invalid spec: %v`, err)
	}
	if err := validation.ValidateRFC1035NamespacedName(ka.Spec.User); err != nil {
		return fmt.Errorf(`invalid spec %v`, err)
	}
	return nil
}
