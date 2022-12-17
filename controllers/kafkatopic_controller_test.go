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
	"strings"
	"time"

	ksfv1 "github.com/ksflow/ksflow/api/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	crclient "sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("KafkaConfig controller", func() {

	const (
		KTName      = "test-kt"
		KTNamespace = "default"

		timeout  = time.Second * 10
		duration = time.Second * 10
		interval = time.Millisecond * 250
	)

	Context("When creating empty KafkaTopic", func() {
		It("Should update the topic in Kafka", func() {
			By("By creating a new KafkaTopic")
			ctx := context.Background()
			kc := &ksfv1.KafkaTopic{
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
			Expect(testK8sClient.Create(ctx, kc)).Should(Succeed())

			KTNamespacedName := types.NamespacedName{Name: KTName, Namespace: KTNamespace}
			createdKT := &ksfv1.KafkaTopic{}
			Eventually(func() bool {
				err := testK8sClient.Get(ctx, KTNamespacedName, createdKT)
				if err != nil {
					return false
				}
				return true
			}, timeout, interval).Should(BeTrue())
			Expect(createdKT.Spec.ReclaimPolicy).Should(Equal(ksfv1.KafkaTopicReclaimPolicyDelete))

			By("By starting the Kafka broker")

			// add label to trigger reconcile
			patchStr := fmt.Sprintf(`{"spec": {"configs": {"bootstrap.servers": "%s"}}}`, strings.Join(testKafkaContainerWrapper.GetAddresses(), ","))
			Expect(testK8sClient.Patch(ctx, kc, crclient.RawPatch(types.MergePatchType, []byte(patchStr)))).Should(Succeed())

			By("By checking that the KafkaTopic has an Available phase")
			Eventually(func() (string, error) {
				err := testK8sClient.Get(ctx, KTNamespacedName, createdKT)
				if err != nil {
					return "", err
				}
				return string(createdKT.Status.Phase), nil
			}, duration, interval).Should(Equal(ksfv1.KafkaTopicPhaseAvailable))
			// TODO: check kafka cluster for topic
		})
	})
})