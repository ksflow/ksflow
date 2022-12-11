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

var _ = Describe("ClusterKafkaConfig controller", func() {

	const (
		CKCCName      = "test-ckc"
		CKCCNamespace = "default"

		CKCCSecurityProtocol = "PLAINTEXT"

		timeout  = time.Second * 10
		duration = time.Second * 10
		interval = time.Millisecond * 250
	)

	Context("When updating ClusterKafkaConfig bootstrap.servers", func() {
		It("Should update the ClusterKafkaConfig connection Status", func() {
			By("By creating a new ClusterKafkaConfig")
			ctx := context.Background()
			ckc := &ksfv1.ClusterKafkaConfig{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "ksflow.io/v1alpha1",
					Kind:       "ClusterKafkaConfig",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      CKCCName,
					Namespace: CKCCNamespace,
				},
				Spec: ksfv1.ClusterKafkaConfigSpec{
					TopicPrefix: "io.ksflow.test",
					Configs: ksfv1.KafkaConfigs{
						BootstrapServers: strings.Join(testPostgresContainerWrapper.GetAddresses(), ","),
						SecurityProtocol: CKCCSecurityProtocol,
					},
				},
			}
			Expect(testK8sClient.Create(ctx, ckc)).Should(Succeed())

			CKCCNamespacedName := types.NamespacedName{Name: CKCCName, Namespace: CKCCNamespace}
			createdCKCC := &ksfv1.ClusterKafkaConfig{}
			Eventually(func() bool {
				err := testK8sClient.Get(ctx, CKCCNamespacedName, createdCKCC)
				if err != nil {
					return false
				}
				return true
			}, timeout, interval).Should(BeTrue())
			Expect(createdCKCC.Spec.TopicPrefix).Should(Equal("io.ksflow.test"))

			By("By checking that the ClusterKafkaConfig has a Failed phase")
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
			Expect(testK8sClient.Patch(ctx, ckc, crclient.RawPatch(types.MergePatchType, []byte(patchStr)))).Should(Succeed())

			By("By checking that the ClusterKafkaConfig has an Available phase")
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
