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

package controllers

import (
	"context"
	"testing"

	ksfv1 "github.com/ksflow/ksflow/api/v1alpha1"
	"github.com/stretchr/testify/assert"
	"github.com/twmb/franz-go/pkg/sr"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

const (
	KSName = "test-ks"
)

var (
	ksNamespacedName = types.NamespacedName{Name: KSName, Namespace: testNamespace}
	ksSchema         = `{"type":"record","namespace":"test","name":"Test","fields":[{"name":"Foo","type":"string"}]}`
	ks               = &ksfv1.KafkaSchema{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "ksflow.io/v1alpha1",
			Kind:       "KafkaSchema",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      KSName,
			Namespace: testNamespace,
		},
		Spec: ksfv1.KafkaSchemaSpec{
			KafkaSubjectInClusterConfiguration: ksfv1.KafkaSubjectInClusterConfiguration{
				Schema:             ksSchema,
				Type:               ksfv1.KafkaSchemaTypeAvro,
				CompatibilityLevel: ksfv1.KafkaSchemaCompatibilityLevelNone,
				Mode:               ksfv1.KafkaSchemaModeReadWrite,
				References:         []sr.SchemaReference{},
			},
		},
	}
)

func TestKafkaSchemaReconcile(t *testing.T) {
	t.Run("create and update KafkaSchema", func(t *testing.T) {
		ctx := context.Background()
		assert.NoError(t, testK8sClient.Create(ctx, ks.DeepCopy()))
		createdKS := &ksfv1.KafkaSchema{}
		assert.Eventually(t, func() bool {
			return testK8sClient.Get(ctx, ksNamespacedName, createdKS) == nil && createdKS.Status.Phase != ksfv1.KsflowPhaseUnknown
		}, testWaitDuration, testWaitInterval)
		assert.Equal(t, ksfv1.KsflowPhaseAvailable, createdKS.Status.Phase)

		ksCopy := ks.DeepCopy()
		ksCopy.Spec.Schema = `{"type":"record","namespace":"test","name":"Test","fields":[{"name":"Bar","type":"string"}]}`
		ksCopy.SetResourceVersion(createdKS.GetResourceVersion())
		assert.NoError(t, testK8sClient.Update(ctx, ksCopy))
		updatedKS := &ksfv1.KafkaSchema{}
		assert.Eventually(t, func() bool {
			return testK8sClient.Get(ctx, ksNamespacedName, updatedKS) == nil && updatedKS.Status.SchemaCount != nil && *updatedKS.Status.SchemaCount == 2
		}, testWaitDuration, testWaitInterval)
		assert.Equal(t, ksCopy.Spec.Schema, updatedKS.Status.Schema)
	})
}
