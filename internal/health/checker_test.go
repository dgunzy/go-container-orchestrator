package health_test

import (
	"context"
	"testing"
	"time"

	"github.com/dgunzy/go-container-orchestrator/internal/container"
	"github.com/dgunzy/go-container-orchestrator/internal/health"
	"github.com/dgunzy/go-container-orchestrator/tests"
	docker_container "github.com/docker/docker/api/types/container"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testImage = "alpine:latest"

func TestHealthyContainer(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cm := tests.InitTestConfig()
	defer tests.CleanupTestResources(cm.DockerClient)

	config := &container.ContainerConfig{
		DomainName:    "test.example.com",
		ImageName:     testImage,
		ContainerName: "test-create-container",
		Cmd:           []string{"tail", "-f", "/dev/null"},
	}

	err := cm.CreateNewContainer(ctx, config)
	require.NoError(t, err, "Error creating container")

	// Verify the container was created and is running
	containers, err := cm.DockerClient.ListContainers(ctx)
	require.NoError(t, err, "Error listing containers")

	var found bool
	var containerID string
	for _, c := range containers {
		if c.Names[0] == "/"+config.ContainerName {
			found = true
			assert.Equal(t, "running", c.State, "Container is not running")
			containerID = c.ID
			break
		}
	}
	assert.True(t, found, "Created container was not found")

	healthChecker := health.NewHealthChecker(cm.DockerClient, cm.Db, 1*time.Second)

	healthCheckerCtx, healthCheckerCancel := context.WithTimeout(ctx, 5*time.Second)
	defer healthCheckerCancel()

	go healthChecker.Start(healthCheckerCtx)

	// Wait for the health checker to run a few times
	select {
	case <-healthCheckerCtx.Done():
	case <-ctx.Done():
		t.Fatal("Test timed out")
	}

	// Clean up
	err = cm.DockerClient.StopContainer(ctx, containerID, nil)
	assert.NoError(t, err, "Error stopping container")
	err = cm.DockerClient.RemoveContainer(ctx, containerID, docker_container.RemoveOptions{Force: true})
	assert.NoError(t, err, "Error removing container")
	err = cm.DockerClient.RemoveImage(ctx, testImage)
	assert.NoError(t, err, "Error removing image")
}

func TestContainerKill(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()

	cm := tests.InitTestConfig()
	defer tests.CleanupTestResources(cm.DockerClient)
	config := &container.ContainerConfig{
		DomainName:    "test.example.com",
		ImageName:     testImage,
		ContainerName: "test-create-container",
		Cmd:           []string{"tail", "-f", "/dev/null"},
	}

	err := cm.CreateNewContainer(ctx, config)
	require.NoError(t, err, "Error creating container")

	// Verify the container was created and is running
	containers, err := cm.DockerClient.ListContainers(ctx)
	require.NoError(t, err, "Error listing containers")

	var found bool
	var containerID string
	for _, c := range containers {
		if c.Names[0] == "/"+config.ContainerName {
			found = true
			assert.Equal(t, "running", c.State, "Container is not running")
			containerID = c.ID
			break
		}
	}
	require.True(t, found, "Created container was not found")

	err = cm.DockerClient.StopContainer(ctx, containerID, nil)
	require.NoError(t, err, "Error stopping container")
	t.Log("Container stopped by test")

	// Wait a bit to ensure the container is stopped
	time.Sleep(2 * time.Second)

	// Verify the container is stopped
	state, err := cm.DockerClient.HealthCheck(ctx, containerID)
	require.NoError(t, err, "Error checking container health")
	assert.NotEqual(t, "running", state.Status, "Container should not be running")

	healthChecker := health.NewHealthChecker(cm.DockerClient, cm.Db, 3*time.Second) // Increased interval
	healthCheckerCtx, healthCheckerCancel := context.WithTimeout(ctx, 20*time.Second)
	defer healthCheckerCancel()

	go healthChecker.Start(healthCheckerCtx)

	// Wait for the health checker to run a few times
	time.Sleep(10 * time.Second)

	// Verify that the container has been restarted
	state, err = cm.DockerClient.HealthCheck(ctx, containerID)
	require.NoError(t, err, "Error checking container health")
	assert.Equal(t, "running", state.Status, "Container should have been restarted and running")

	// Clean up
	err = cm.DockerClient.StopContainer(ctx, containerID, nil)
	assert.NoError(t, err, "Error stopping container")
	err = cm.DockerClient.RemoveContainer(ctx, containerID, docker_container.RemoveOptions{Force: true})
	assert.NoError(t, err, "Error removing container")
	err = cm.DockerClient.RemoveImage(ctx, testImage)
	assert.NoError(t, err, "Error removing image")
}
