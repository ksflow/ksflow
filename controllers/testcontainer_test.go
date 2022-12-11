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
		Image:        fmt.Sprintf("%s:%s", "docker.vectorized.io/vectorized/redpanda", "v22.3.5"),
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
