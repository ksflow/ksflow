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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

const (
	KUName = "test-ku"
)

var (
	kuNamespacedName = types.NamespacedName{Name: KUName, Namespace: testNamespace}
	ku               = &ksfv1.KafkaUser{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "ksflow.io/v1alpha1",
			Kind:       "KafkaUser",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      KUName,
			Namespace: testNamespace,
		},
		Spec: ksfv1.KafkaUserSpec{},
	}
)

func TestKafkaUserReconcile(t *testing.T) {
	t.Run("create a KafkaUser", func(t *testing.T) {
		ctx := context.Background()
		assert.NoError(t, testK8sClient.Create(ctx, ku.DeepCopy()))
		createdKU := &ksfv1.KafkaUser{}
		assert.Eventually(t, func() bool {
			return testK8sClient.Get(ctx, kuNamespacedName, createdKU) == nil && createdKU.Status.Phase != ksfv1.KsflowPhaseUnknown
		}, testWaitDuration, testWaitInterval)
		assert.Equal(t, ksfv1.KsflowPhaseAvailable, createdKU.Status.Phase)
		assert.Equal(t, "CN=test-ku.default.svc,OU=TEST,O=Marketing,L=Charlottesville,ST=Va,C=US", createdKU.Status.UserName)
	})
}
