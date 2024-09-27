package container

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/dgunzy/go-container-orchestrator/pkg/docker"
	"github.com/docker/docker/api/types/container"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testImage = "alpine:latest"
const testImage2 = "alpine:3.12"

func TestCreateNewContainer(t *testing.T) {
	// cm, dockerClient := setupTest(t)
	ctx := context.Background()
	dockerClient, err := docker.NewClient()
	if err != nil {
		t.Fatalf("Error creating client: %s", err)
	}

	cm, err := NewContainerManager(dockerClient, ":memory:")
	if err != nil {
		t.Fatalf("Error creating client: %s", err)
	}

	cm.db.InitSchema()

	config := &ContainerConfig{
		DomainName:    "test.example.com",
		ImageName:     testImage,
		ContainerName: "test-create-container",
		Cmd:           []string{"tail", "-f", "/dev/null"},
	}

	err = cm.CreateNewContainer(ctx, config)
	if err != nil {
		t.Fatalf("Error creating container: %s", err)
	}
	assert.NoError(t, err)

	// Verify the container was created and is running
	containers, err := dockerClient.ListContainers(ctx)
	assert.NoError(t, err)

	var found bool
	for _, c := range containers {
		if c.Names[0] == "/"+config.ContainerName {
			found = true
			assert.Equal(t, "running", c.State)
			break
		}
	}
	assert.True(t, found, "Created container was not found or not running")

	// Clean up
	err = dockerClient.StopContainer(ctx, config.ContainerName, nil)
	assert.NoError(t, err)
	err = dockerClient.RemoveContainer(ctx, config.ContainerName, container.RemoveOptions{})
	assert.NoError(t, err)
	err = dockerClient.RemoveImage(ctx, testImage)
	assert.NoError(t, err)
}

func TestUpdateExistingContainer(t *testing.T) {
	ctx := context.Background()
	dockerClient, err := docker.NewClient()
	if err != nil {
		t.Fatalf("Error creating client: %s", err)
	}
	cm, err := NewContainerManager(dockerClient, ":memory:")

	if err != nil {
		t.Fatalf("Error creating client: %s", err)
	}
	cm.db.InitSchema()

	// First, create an initial container
	initialConfig := &ContainerConfig{
		DomainName:    "test.example.com",
		ImageName:     testImage,
		ContainerName: "test-update-container",
		Cmd:           []string{"tail", "-f", "/dev/null"},
	}
	err = cm.CreateNewContainer(ctx, initialConfig)
	require.NoError(t, err)

	// Wait for the container to start
	time.Sleep(5 * time.Second)

	// Now, update the container
	updateConfig := &ContainerConfig{
		DomainName:    "test.example.com",
		ImageName:     testImage2,
		ContainerName: "test-update-container",
		Cmd:           []string{"tail", "-f", "/dev/null"},
	}
	err = cm.UpdateExistingContainer(ctx, updateConfig)
	assert.NoError(t, err)

	// Wait for the update to complete
	time.Sleep(5 * time.Second)

	// List all containers and log them
	containers, err := dockerClient.ListContainers(ctx)
	assert.NoError(t, err)

	fmt.Println("--- All running containers ---")
	for _, c := range containers {
		fmt.Printf("ID: %s, Names: %v, Image: %s, State: %s\n", c.ID[:12], c.Names, c.Image, c.State)
	}
	fmt.Println("------------------------------")

	// Check for the new container
	var foundNew bool
	var newContainerID string
	expectedImageName := updateConfig.ImageName // This should be testImage2
	for _, c := range containers {
		if c.Image == expectedImageName && c.State == "running" {
			foundNew = true
			newContainerID = c.ID
			break
		}
	}

	if !foundNew {
		t.Errorf("New container with image %s was not found or not running", expectedImageName)
	} else {
		fmt.Printf("Found new container: ID %s, Image %s\n", newContainerID[:12], expectedImageName)
	}

	// Clean up
	if newContainerID != "" {
		err = dockerClient.StopContainer(ctx, newContainerID, nil)
		assert.NoError(t, err)
		err = dockerClient.RemoveContainer(ctx, newContainerID, container.RemoveOptions{Force: true})
		assert.NoError(t, err)
	}

	// Remove all containers with names starting with "test-update-container"
	containers, _ = dockerClient.ListContainers(ctx)
	for _, c := range containers {
		if len(c.Names) > 0 && strings.HasPrefix(c.Names[0], "/test-update-container") {
			dockerClient.StopContainer(ctx, c.ID, nil)
			dockerClient.RemoveContainer(ctx, c.ID, container.RemoveOptions{Force: true})
		}
	}

	// Wait a bit before trying to remove images
	time.Sleep(2 * time.Second)

	// Try to remove images, but don't fail the test if we can't
	_ = dockerClient.RemoveImage(ctx, testImage)
	_ = dockerClient.RemoveImage(ctx, testImage2)
}
