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
	"fmt"
	"testing"
	"time"

	ksfv1 "github.com/ksflow/ksflow/api/v1alpha1"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	crclient "sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	KTName      = "test-kt"
	KTNamespace = "default"

	duration = time.Second * 10
	interval = time.Millisecond * 250
)

var (
	ktNamespacedName = types.NamespacedName{Name: KTName, Namespace: KTNamespace}
	kc               = &ksfv1.KafkaTopic{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "ksflow.io/v1alpha1",
			Kind:       "KafkaTopic",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      KTName,
			Namespace: KTNamespace,
		},
		Spec: ksfv1.KafkaTopicSpec{},
	}
)

func TestReconcile(t *testing.T) {
	t.Run("create and update KafkaTopic", func(t *testing.T) {
		ctx := context.Background()
		assert.NoError(t, testK8sClient.Create(ctx, kc.DeepCopy()))
		createdKT := &ksfv1.KafkaTopic{}
		assert.Eventually(t, func() bool {
			return testK8sClient.Get(ctx, ktNamespacedName, createdKT) == nil && createdKT.Status.ReclaimPolicy != nil
		}, duration, interval)
		assert.Equal(t, ksfv1.KafkaTopicReclaimPolicyDelete, *createdKT.Status.ReclaimPolicy)
		assert.Equal(t, ksfv1.KafkaTopicPhaseAvailable, createdKT.Status.Phase)

		retentionBytes := "1073741824"
		patchStr := fmt.Sprintf(`{"spec": {"configs": {"retention.bytes": "%s"}}}`, retentionBytes)
		assert.NoError(t, testK8sClient.Patch(ctx, kc.DeepCopy(), crclient.RawPatch(types.MergePatchType, []byte(patchStr))))
		updatedKT := &ksfv1.KafkaTopic{}
		assert.Eventually(t, func() bool {
			return testK8sClient.Get(ctx, ktNamespacedName, updatedKT) == nil && updatedKT.Status.Configs["retention.bytes"] != nil
		}, duration, interval)
		assert.Equal(t, retentionBytes, *updatedKT.Status.Configs["retention.bytes"])
	})
}
