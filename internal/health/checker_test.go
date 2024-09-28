package health

import (
	"context"
	"testing"
	"time"

	"github.com/dgunzy/go-container-orchestrator/internal/container"
	"github.com/dgunzy/go-container-orchestrator/pkg/docker"
	docker_container "github.com/docker/docker/api/types/container"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testImage = "alpine:latest"

func TestHealthyContainer(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	dockerClient, err := docker.NewClient()
	require.NoError(t, err, "Error creating Docker client")

	cm, err := container.NewContainerManager(dockerClient, ":memory:")
	require.NoError(t, err, "Error creating ContainerManager")

	err = cm.Db.InitSchema()
	require.NoError(t, err, "Error initializing database schema")

	config := &container.ContainerConfig{
		DomainName:    "test.example.com",
		ImageName:     testImage,
		ContainerName: "test-create-container",
		Cmd:           []string{"tail", "-f", "/dev/null"},
	}

	err = cm.CreateNewContainer(ctx, config)
	require.NoError(t, err, "Error creating container")

	// Verify the container was created and is running
	containers, err := dockerClient.ListContainers(ctx)
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

	healthChecker := NewHealthChecker(dockerClient, cm.Db, 1*time.Second)

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
	err = dockerClient.StopContainer(ctx, containerID, nil)
	assert.NoError(t, err, "Error stopping container")
	err = dockerClient.RemoveContainer(ctx, containerID, docker_container.RemoveOptions{Force: true})
	assert.NoError(t, err, "Error removing container")
	err = dockerClient.RemoveImage(ctx, testImage)
	assert.NoError(t, err, "Error removing image")
}
