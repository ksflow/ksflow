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
	"os"
	"path"
	"path/filepath"
	"runtime"
	"testing"

	"k8s.io/utils/pointer"
	ctrl "sigs.k8s.io/controller-runtime"

	ksfv1 "github.com/ksflow/ksflow/api/v1alpha1"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	// +kubebuilder:scaffold:imports
)

var testK8sClient client.Client
var testEnv *envtest.Environment
var testCtx context.Context
var testKafkaContainerWrapper *TestContainerWrapper
var testCancel context.CancelFunc

func currTestDir() string {
	_, currTestFilename, _, _ := runtime.Caller(0)
	return path.Dir(currTestFilename)
}

func TestMain(m *testing.M) {
	if err := setup(); err != nil {
		teardown()
		os.Exit(1)
	} else {
		code := m.Run()
		teardown()
		os.Exit(code)
	}
}

func setup() error {
	testCtx, testCancel = context.WithCancel(context.TODO())

	testEnv = &envtest.Environment{
		CRDDirectoryPaths:     []string{filepath.Join("..", "config", "crd", "bases")},
		ErrorIfCRDPathMissing: true,
	}

	testCfg, err := testEnv.Start()
	if err != nil {
		return err
	}

	testKafkaContainerWrapper = &TestContainerWrapper{}
	if err = testKafkaContainerWrapper.RunKafka(); err != nil {
		return err
	}

	if err = ksfv1.AddToScheme(scheme.Scheme); err != nil {
		return err
	}

	// +kubebuilder:scaffold:scheme

	testK8sClient, err = client.New(testCfg, client.Options{Scheme: scheme.Scheme})
	if err != nil {
		return err
	}

	k8sManager, err := ctrl.NewManager(testCfg, ctrl.Options{Scheme: scheme.Scheme})
	if err != nil {
		return err
	}

	kafkaConnectionConfig := ksfv1.KafkaConnectionConfig{
		BootstrapServers: testKafkaContainerWrapper.GetAddresses(),
		KafkaTLSConfig: ksfv1.KafkaTLSConfig{
			CertFilePath: path.Join(currTestDir(), "testdata", "certs", "test-ksflow-controller.crt"),
			KeyFilePath:  path.Join(currTestDir(), "testdata", "certs", "test-ksflow-controller.key"),
			CAFilePath:   path.Join(currTestDir(), "testdata", "certs", "test-root-ca.crt"),
		},
	}

	rfi16 := int16(1)
	kafkaTopicDefaultsConfig := ksfv1.KafkaTopicSpec{
		KafkaTopicInClusterConfiguration: ksfv1.KafkaTopicInClusterConfiguration{
			Partitions:        pointer.Int32(2),
			ReplicationFactor: &rfi16,
			Configs:           nil,
		},
	}

	err = (&KafkaTopicReconciler{
		Client:                 k8sManager.GetClient(),
		Scheme:                 k8sManager.GetScheme(),
		KafkaConnectionConfig:  kafkaConnectionConfig,
		KafkaTopicSpecDefaults: kafkaTopicDefaultsConfig,
	}).SetupWithManager(k8sManager)
	if err != nil {
		return err
	}

	go func() {
		if err = k8sManager.Start(testCtx); err != nil {
			panic(err)
		}
	}()
	return nil
}

func teardown() {
	testCancel()
	_ = testEnv.Stop()
	_ = testKafkaContainerWrapper.CleanUp()
}
