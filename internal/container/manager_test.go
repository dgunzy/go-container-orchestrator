package container

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/dgunzy/go-container-orchestrator/pkg/docker"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testImage = "alpine:latest"
const testImage2 = "alpine:3.12"

func TestCreateNewContainer(t *testing.T) {
	ctx := context.Background()
	dockerClient, err := docker.NewClient()
	require.NoError(t, err, "Error creating Docker client")

	cm, err := NewContainerManager(dockerClient, ":memory:")
	require.NoError(t, err, "Error creating ContainerManager")

	err = cm.db.InitSchema()
	require.NoError(t, err, "Error initializing database schema")

	config := &ContainerConfig{
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

	// Clean up
	err = dockerClient.StopContainer(ctx, containerID, nil)
	assert.NoError(t, err, "Error stopping container")
	err = dockerClient.RemoveContainer(ctx, containerID, container.RemoveOptions{Force: true})
	assert.NoError(t, err, "Error removing container")
	err = dockerClient.RemoveImage(ctx, testImage)
	assert.NoError(t, err, "Error removing image")
}

func TestUpdateExistingContainer(t *testing.T) {
	ctx := context.Background()
	dockerClient, err := docker.NewClient()
	require.NoError(t, err, "Error creating Docker client")

	cm, err := NewContainerManager(dockerClient, ":memory:")
	require.NoError(t, err, "Error creating ContainerManager")

	err = cm.db.InitSchema()
	require.NoError(t, err, "Error initializing database schema")

	// Create an initial container
	initialConfig := &ContainerConfig{
		DomainName:    "test.example.com",
		ImageName:     testImage,
		ContainerName: "test-update-container",
		Cmd:           []string{"tail", "-f", "/dev/null"},
	}
	err = cm.CreateNewContainer(ctx, initialConfig)
	require.NoError(t, err, "Error creating initial container")

	// Wait for the container to start
	time.Sleep(5 * time.Second)

	// Get the container ID from Docker
	containers, err := dockerClient.ListContainers(ctx)
	require.NoError(t, err, "Error listing containers")
	var initialContainerID string
	for _, c := range containers {
		if c.Names[0] == "/"+initialConfig.ContainerName {
			initialContainerID = c.ID
			break
		}
	}
	require.NotEmpty(t, initialContainerID, "Initial container ID not found")

	// Update the container
	updateConfig := &ContainerConfig{
		DomainName:    "test.example.com",
		ImageName:     testImage2,
		ContainerName: "test-update-container",
		ContainerID:   initialContainerID,
		Cmd:           []string{"tail", "-f", "/dev/null"},
	}
	err = cm.UpdateExistingContainer(ctx, updateConfig)
	assert.NoError(t, err, "Error updating container")

	// Wait for the update to complete
	time.Sleep(5 * time.Second)

	// Verify the update
	containers, err = dockerClient.ListContainers(ctx)
	require.NoError(t, err, "Error listing containers")

	var updatedContainer *types.Container
	for _, c := range containers {
		if strings.HasPrefix(c.Names[0], "/"+updateConfig.ContainerName) && c.Image == updateConfig.ImageName {
			updatedContainer = &c
			break
		}
	}

	require.NotNil(t, updatedContainer, "Updated container not found")
	assert.Equal(t, "running", updatedContainer.State, "Updated container is not running")
	assert.NotEqual(t, initialContainerID, updatedContainer.ID, "Container ID should have changed")

	// Clean up
	err = dockerClient.StopContainer(ctx, updatedContainer.ID, nil)
	assert.NoError(t, err, "Error stopping updated container")
	err = dockerClient.RemoveContainer(ctx, updatedContainer.ID, container.RemoveOptions{Force: true})
	assert.NoError(t, err, "Error removing updated container")

	_ = dockerClient.RemoveImage(ctx, testImage)
	_ = dockerClient.RemoveImage(ctx, testImage2)
}

func TestLoadAndStartContainers(t *testing.T) {
	ctx := context.Background()
	dockerClient, err := docker.NewClient()
	require.NoError(t, err, "Error creating Docker client")

	cm, err := NewContainerManager(dockerClient, ":memory:")
	require.NoError(t, err, "Error creating ContainerManager")

	err = cm.db.InitSchema()
	require.NoError(t, err, "Error initializing database schema")

	// Create multiple containers
	containers := []ContainerConfig{
		{
			DomainName:    "test1.example.com",
			ImageName:     testImage,
			ContainerName: "test-load-container-1",
			ContainerPort: "8080",
			Cmd:           []string{"tail", "-f", "/dev/null"},
		},
		{
			DomainName:    "test2.example.com",
			ImageName:     testImage2,
			ContainerName: "test-load-container-2",
			ContainerPort: "8081",
			Cmd:           []string{"tail", "-f", "/dev/null"},
		},
	}

	createdContainers := make(map[string]string) // map[containerName]containerID

	for _, config := range containers {
		err := cm.CreateNewContainer(ctx, &config)
		require.NoError(t, err, "Error creating container: %s", config.ContainerName)

		// Get the container ID
		dockerContainers, err := dockerClient.ListContainers(ctx)
		require.NoError(t, err, "Error listing containers")
		for _, c := range dockerContainers {
			if c.Names[0] == "/"+config.ContainerName {
				createdContainers[config.ContainerName] = c.ID
				break
			}
		}
	}

	// Wait for containers to start
	time.Sleep(3 * time.Second)

	// Stop all containers
	for _, containerID := range createdContainers {
		err = dockerClient.StopContainer(ctx, containerID, nil)
		require.NoError(t, err, "Error stopping container: %s", containerID)
		err = dockerClient.RemoveContainer(ctx, containerID, container.RemoveOptions{Force: true})
		require.NoError(t, err, "Error removing container: %s", containerID)
	}

	// Wait for containers to stop
	time.Sleep(3 * time.Second)

	// Use LoadAndStartContainers to bring them back up
	err = cm.LoadAndStartContainers(ctx, []string{"tail", "-f", "/dev/null"})

	require.NoError(t, err, "Error loading and starting containers")

	// Wait for containers to start
	time.Sleep(3 * time.Second)

	// Verify that all containers are running
	runningContainers, err := dockerClient.ListContainers(ctx)
	require.NoError(t, err, "Error listing containers")

	assert.Len(t, runningContainers, len(containers), "Not all containers are running")

	// Query the database for the latest container information
	dbContainers, err := cm.db.ListContainers()
	require.NoError(t, err, "Error listing containers from database")
	assert.Len(t, dbContainers, len(containers), "Not all containers are in the database")

	// Check if all containers in the database are running
	for _, dbContainer := range dbContainers {
		containerState, err := dockerClient.HealthCheck(ctx, dbContainer.ContainerID)
		require.NoError(t, err, "Error checking container state: %s", dbContainer.ContainerName)
		assert.Equal(t, "running", containerState.Status, "Container %s is not running", dbContainer.ContainerName)
	}

	// Clean up
	for _, dbContainer := range dbContainers {
		err = dockerClient.StopContainer(ctx, dbContainer.ContainerID, nil)
		assert.NoError(t, err, "Error stopping container: %s", dbContainer.ContainerName)
		err = dockerClient.RemoveContainer(ctx, dbContainer.ContainerID, container.RemoveOptions{Force: true})
		assert.NoError(t, err, "Error removing container: %s", dbContainer.ContainerName)
	}

	// Remove test images
	_ = dockerClient.RemoveImage(ctx, testImage)
	_ = dockerClient.RemoveImage(ctx, testImage2)

}
