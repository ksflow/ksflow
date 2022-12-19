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
	"os"
	"path"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

type TestContainerWrapper struct {
	testContainer     testcontainers.Container
	testContainerPort int
}

func (t *TestContainerWrapper) RunKafka() error {
	req := testcontainers.ContainerRequest{
		Image:        fmt.Sprintf("%s:%s", "bitnami/kafka", "3.3.1-debian-11-r19"),
		ExposedPorts: []string{"9092/tcp", "9093/tcp"},
		Env: map[string]string{
			"BITNAMI_DEBUG":                            "yes",
			"KAFKA_ENABLE_KRAFT":                       "yes",
			"KAFKA_BROKER_ID":                          "1",
			"KAFKA_CFG_PROCESS_ROLES":                  "broker,controller",
			"KAFKA_CFG_CONTROLLER_LISTENER_NAMES":      "CONTROLLER",
			"KAFKA_CFG_LISTENER_SECURITY_PROTOCOL_MAP": "SSL:SSL,CONTROLLER:SSL",
			"KAFKA_CFG_ADVERTISED_LISTENERS":           "SSL://localhost:9092",
			"KAFKA_CFG_EARLY_START_LISTENERS":          "CONTROLLER",
			"KAFKA_CFG_SUPER_USERS":                    "User:CN=localhost,OU=Some Unit,O=Widgets Inc,L=Columbus,ST=Ohio,C=US;User:CN=ksflow-test,OU=Some Unit,O=Widgets Inc,L=Columbus,ST=Ohio,C=US",
			"KAFKA_CFG_LISTENERS":                      "SSL://:9092,CONTROLLER://:9093",
			"KAFKA_CFG_INTER_BROKER_LISTENER_NAME":     "SSL",
			"KAFKA_CFG_CONTROLLER_QUORUM_VOTERS":       "1@localhost:9093",
			"KAFKA_CFG_SSL_KEYSTORE_TYPE":              "PKCS12",
			"KAFKA_CFG_SSL_KEYSTORE_LOCATION":          "/opt/bitnami/kafka/config/certs/kafka.keystore.p12",
			"KAFKA_CFG_SSL_KEYSTORE_PASSWORD":          "password",
			"KAFKA_CFG_SSL_KEY_PASSWORD":               "password",
			"KAFKA_CFG_SSL_TRUSTSTORE_TYPE":            "PKCS12",
			"KAFKA_CFG_SSL_TRUSTSTORE_LOCATION":        "/opt/bitnami/kafka/config/certs/kafka.truststore.p12",
			"KAFKA_CFG_SSL_TRUSTSTORE_PASSWORD":        "password",
			"KAFKA_CFG_SSL_ENABLED_PROTOCOLS":          "TLSv1.2",
			"KAFKA_CFG_SSL_PROTOCOL":                   "TLSv1.2",
			"KAFKA_CFG_SSL_CLIENT_AUTH":                "required",
			"KAFKA_CFG_AUTO_CREATE_TOPICS_ENABLE":      "false",
			"KAFKA_CFG_ALLOW_EVERYONE_IF_NO_ACL_FOUND": "false",
			"KAFKA_CFG_AUTHORIZER_CLASS_NAME":          "org.apache.kafka.metadata.authorizer.StandardAuthorizer",
		},
		//WaitingFor: wait.ForLog("INFO [KafkaRaftServer nodeId=1] Kafka Server started (kafka.server.KafkaRaftServer)"),
		Cmd: []string{"/opt/bitnami/scripts/kafka/test-kafka-run.sh"},
		//Entrypoint: []string{"/bin/sh", "-c", "tail -f /dev/null"},
		AutoRemove: false,
		Files: []testcontainers.ContainerFile{
			{
				HostFilePath:      path.Join(currTestDir(), "testdata", "scripts", "test-kafka-run.sh"),
				ContainerFilePath: "/opt/bitnami/scripts/kafka/test-kafka-run.sh",
				FileMode:          493, // 755
			},
			{
				HostFilePath:      path.Join(currTestDir(), "testdata", "certs", "test-kafka.keystore.jks"),
				ContainerFilePath: "/kafka.keystore.jks",
				FileMode:          420, // 644
			},
			{
				HostFilePath:      path.Join(currTestDir(), "testdata", "certs", "test-kafka.truststore.jks"),
				ContainerFilePath: "/kafka.truststore.jks",
				FileMode:          420, // 644
			},
			{
				HostFilePath:      path.Join(currTestDir(), "testdata", "certs", "test-kafka.keystore.p12"),
				ContainerFilePath: "/kafka.keystore.p12",
				FileMode:          420, // 644
			},
			{
				HostFilePath:      path.Join(currTestDir(), "testdata", "certs", "test-kafka.truststore.p12"),
				ContainerFilePath: "/kafka.truststore.p12",
				FileMode:          420, // 644
			},
		},
	}

	container, err := testcontainers.GenericContainer(context.Background(), testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return fmt.Errorf("could not create container: %w", err)
	}

	mPort, err := container.MappedPort(context.Background(), "9092")
	if err != nil {
		return fmt.Errorf("could not get mapped port from the container: %w", err)
	}

	// set KAFKA_CFG_ADVERTISED_LISTENERS to the port that testcontainers decided to use
	// ref: https://franklinlindemberg.medium.com/how-to-use-kafka-with-testcontainers-in-golang-applications-9266c738c879
	kafkaStartFile, err := os.CreateTemp("", "testcontainers_start.sh")
	if err != nil {
		panic(err)
	}
	defer os.Remove(kafkaStartFile.Name())
	if _, err = kafkaStartFile.WriteString("#!/bin/bash \n"); err != nil {
		return err
	}
	if _, err = kafkaStartFile.WriteString(fmt.Sprintf("export KAFKA_CFG_ADVERTISED_LISTENERS='SSL://localhost:%v'\n", mPort.Int())); err != nil {
		return err
	}
	if err = container.CopyFileToContainer(context.Background(), kafkaStartFile.Name(), "/testcontainers_start.sh", 493); err != nil {
		return err
	}

	t.testContainer = container
	t.testContainerPort = mPort.Int()

	// consider checking port or logs to verify things start up before returning if it is a problem
	return nil
}

func (t *TestContainerWrapper) RunPostgres() error {
	req := testcontainers.ContainerRequest{
		Image:        fmt.Sprintf("%s:%s", "postgres", "12"),
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_USER":     "postgres",
			"POSTGRES_PASSWORD": "postgres",
			"POSTGRES_DB":       "postgres",
		},
		WaitingFor: wait.ForListeningPort("5432/tcp"),
		AutoRemove: true,
	}

	container, err := testcontainers.GenericContainer(context.Background(), testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return fmt.Errorf("could not create container: %w", err)
	}

	mPort, err := container.MappedPort(context.Background(), "5432")
	if err != nil {
		return fmt.Errorf("could not get mapped port from the container: %w", err)
	}

	t.testContainer = container
	t.testContainerPort = mPort.Int()

	return nil
}

func (t *TestContainerWrapper) CleanUp() error {
	ctx, cancelFunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelFunc()

	return t.testContainer.Terminate(ctx)
}

func (t *TestContainerWrapper) GetAddresses() []string {
	return []string{t.GetAddress()}
}

func (t *TestContainerWrapper) GetAddress() string {
	return fmt.Sprintf("localhost:%d", t.testContainerPort)
}
