/*

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
		CKCCName      = "test-kc"
		CKCCNamespace = "default"

		CKCCSecurityProtocol = "PLAINTEXT"

		timeout  = time.Second * 10
		duration = time.Second * 10
		interval = time.Millisecond * 250
	)

	Context("When updating KafkaConfig bootstrap.servers", func() {
		It("Should update the KafkaConfig connection Status", func() {
			By("By creating a new KafkaConfig")
			ctx := context.Background()
			kc := &ksfv1.KafkaConfig{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "ksflow.io/v1alpha1",
					Kind:       "KafkaConfig",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      CKCCName,
					Namespace: CKCCNamespace,
				},
				Spec: ksfv1.KafkaConfigSpec{
					TopicPrefix: "io.ksflow.test",
					Configs: ksfv1.KafkaConfigs{
						BootstrapServers: strings.Join(testPostgresContainerWrapper.GetAddresses(), ","),
						SecurityProtocol: CKCCSecurityProtocol,
					},
				},
			}
			Expect(testK8sClient.Create(ctx, kc)).Should(Succeed())

			CKCCNamespacedName := types.NamespacedName{Name: CKCCName, Namespace: CKCCNamespace}
			createdCKCC := &ksfv1.KafkaConfig{}
			Eventually(func() bool {
				err := testK8sClient.Get(ctx, CKCCNamespacedName, createdCKCC)
				if err != nil {
					return false
				}
				return true
			}, timeout, interval).Should(BeTrue())
			Expect(createdCKCC.Spec.TopicPrefix).Should(Equal("io.ksflow.test"))

			By("By checking that the KafkaConfig has a Failed phase")
			Eventually(func() (string, error) {
				err := testK8sClient.Get(ctx, CKCCNamespacedName, createdCKCC)
				if err != nil {
					return "", err
				}
				return string(createdCKCC.Status.Phase), nil
			}, duration, interval).Should(Equal("Failed"))

			By("By starting the Kafka broker")

			// add label to trigger reconcile
			patchStr := fmt.Sprintf(`{"spec": {"configs": {"bootstrap.servers": "%s"}}}`, strings.Join(testKafkaContainerWrapper.GetAddresses(), ","))
			Expect(testK8sClient.Patch(ctx, kc, crclient.RawPatch(types.MergePatchType, []byte(patchStr)))).Should(Succeed())

			By("By checking that the KafkaConfig has an Available phase")
			Eventually(func() (string, error) {
				err := testK8sClient.Get(ctx, CKCCNamespacedName, createdCKCC)
				if err != nil {
					return "", err
				}
				return string(createdCKCC.Status.Phase), nil
			}, duration, interval).Should(Equal("Available"))
		})
	})
})
