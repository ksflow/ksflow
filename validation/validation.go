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

package validation

import (
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/api/validation"
	"k8s.io/apimachinery/pkg/types"
)

// ValidateRFC1035LabelName follows https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#rfc-1035-label-names
// topic names are formed from 3 of these (i.e. "my-controller-prefix.my-namespace.my-topic"), which will satisfy topic name requirements
// (ref: https://github.com/apache/kafka/blob/3.3.1/clients/src/main/java/org/apache/kafka/common/internals/Topic.java#L33)
// while also not allowing "_" as they collide with periods (kafka recommends only using one of the other)
// Consider using RFC 1123 instead of 1035 since it matches what namespace names can be.  Easy to change, hard to change back, keeping restrictive for now.
func ValidateRFC1035LabelName(s string) error {
	namespacedNameValidationErrors := validation.NameIsDNS1035Label(s, false)
	if len(namespacedNameValidationErrors) > 0 {
		return fmt.Errorf("name is not a valid dns label, %v", namespacedNameValidationErrors[0])
	}
	return nil
}

// ParseRFC1035NamespacedName parses a resource name into a NamespacedName
func ParseRFC1035NamespacedName(resourceName string, defaultNamespace string) (*types.NamespacedName, error) {
	rns := strings.SplitN(resourceName, ".", 2)
	if err := ValidateRFC1035LabelName(rns[0]); err != nil {
		return nil, err
	}
	if len(rns) == 1 {
		return &types.NamespacedName{Namespace: defaultNamespace, Name: rns[0]}, nil
	}
	if err := ValidateRFC1035LabelName(rns[1]); err != nil {
		return nil, err
	}
	return &types.NamespacedName{Namespace: rns[0], Name: rns[1]}, nil
}

// ValidateRFC1035NamespacedName validates the resource name is valid (i.e. "my-resource" or "my-namespace.my-resource")
func ValidateRFC1035NamespacedName(fqrn string) error {
	_, err := ParseRFC1035NamespacedName(fqrn, "default")
	return err
}
