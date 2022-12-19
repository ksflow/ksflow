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
	"path"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/ksflow/ksflow/controllers/kafkatopic"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	ctrl "sigs.k8s.io/controller-runtime"

	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	ksfv1 "github.com/ksflow/ksflow/api/v1alpha1"
	// +kubebuilder:scaffold:imports
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

var testCfg *rest.Config
var testK8sClient client.Client
var testEnv *envtest.Environment
var testCtx context.Context
var testKafkaContainerWrapper *TestContainerWrapper
var testPostgresContainerWrapper *TestContainerWrapper
var testCancel context.CancelFunc

func TestAPIs(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecs(t, "Controller Suite")
}

func currTestDir() string {
	_, currTestFilename, _, _ := runtime.Caller(0)
	return path.Dir(currTestFilename)
}

var _ = BeforeSuite(func() {
	logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))

	testCtx, testCancel = context.WithCancel(context.TODO())

	By("bootstrapping test environment")
	testEnv = &envtest.Environment{
		CRDDirectoryPaths:     []string{filepath.Join("..", "config", "crd", "bases")},
		ErrorIfCRDPathMissing: true,
	}

	var err error
	// cfg is defined in this file globally.
	testCfg, err = testEnv.Start()
	Expect(err).NotTo(HaveOccurred())
	Expect(testCfg).NotTo(BeNil())

	testKafkaContainerWrapper = &TestContainerWrapper{}
	err = testKafkaContainerWrapper.RunKafka()
	Expect(err).NotTo(HaveOccurred())
	kafkaBootstrapServers := testKafkaContainerWrapper.GetAddresses()
	Expect(kafkaBootstrapServers).NotTo(BeEmpty())

	testPostgresContainerWrapper = &TestContainerWrapper{}
	err = testPostgresContainerWrapper.RunPostgres()
	Expect(err).NotTo(HaveOccurred())
	postgresAddr := testPostgresContainerWrapper.GetAddress()
	Expect(postgresAddr).NotTo(BeEmpty())

	err = ksfv1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	// +kubebuilder:scaffold:scheme

	testK8sClient, err = client.New(testCfg, client.Options{Scheme: scheme.Scheme})
	Expect(err).NotTo(HaveOccurred())
	Expect(testK8sClient).NotTo(BeNil())

	k8sManager, err := ctrl.NewManager(testCfg, ctrl.Options{Scheme: scheme.Scheme})
	Expect(err).ToNot(HaveOccurred())

	kafkaConfig := ksfv1.KafkaConfig{
		BootstrapServers: testKafkaContainerWrapper.GetAddresses(),
		KafkaTLSConfig: ksfv1.KafkaTLSConfig{
			CertFilePath: path.Join(currTestDir(), "testdata", "certs", "test-ksflow-controller.crt"),
			KeyFilePath:  path.Join(currTestDir(), "testdata", "certs", "test-ksflow-controller.key"),
			CAFilePath:   path.Join(currTestDir(), "testdata", "certs", "test-root-ca.crt"),
		},
	}

	err = (&kafkatopic.KafkaTopicReconciler{
		Client:      k8sManager.GetClient(),
		Scheme:      k8sManager.GetScheme(),
		KafkaConfig: kafkaConfig,
	}).SetupWithManager(k8sManager)
	Expect(err).ToNot(HaveOccurred())

	err = (&kafkatopic.ClusterKafkaTopicReconciler{
		Client:      k8sManager.GetClient(),
		Scheme:      k8sManager.GetScheme(),
		KafkaConfig: kafkaConfig,
	}).SetupWithManager(k8sManager)
	Expect(err).ToNot(HaveOccurred())

	go func() {
		defer GinkgoRecover()
		err = k8sManager.Start(testCtx)
		Expect(err).ToNot(HaveOccurred(), "failed to run manager")
	}()

})

var _ = AfterSuite(func() {
	testCancel()
	By("tearing down the test environment")
	err := testEnv.Stop()
	Expect(err).NotTo(HaveOccurred())

	err = testKafkaContainerWrapper.CleanUp()
	Expect(err).NotTo(HaveOccurred())

	err = testPostgresContainerWrapper.CleanUp()
	Expect(err).NotTo(HaveOccurred())
})
