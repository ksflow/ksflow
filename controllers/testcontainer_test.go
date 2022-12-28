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
		Image:        fmt.Sprintf("%s:%s", "docker.vectorized.io/vectorized/redpanda", "v22.3.9"),
		ExposedPorts: []string{"9092/tcp"},
		Cmd:          []string{"redpanda", "start"},
		WaitingFor:   wait.ForLog("Successfully started Redpanda!"),
		AutoRemove:   true,
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
