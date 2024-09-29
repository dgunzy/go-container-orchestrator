package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/dgunzy/go-container-orchestrator/internal/container"
	"github.com/dgunzy/go-container-orchestrator/internal/logging"
	"github.com/dgunzy/go-container-orchestrator/pkg/docker"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testImage  = "alpine:latest"
	testImage2 = "alpine:3.12"
	serverAddr = "localhost:8080"
)

func TestAPIServer(t *testing.T) {
	err := logging.Setup("./container_manager_logs")
	require.NoError(t, err, "Failed to set up logging")
	defer logging.CloseGlobalLogger()

	logger := logging.GetLogger()

	// Setup
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	dockerClient, err := docker.NewClient()
	require.NoError(t, err, "Error creating Docker client")

	cm, err := container.NewContainerManager(dockerClient, ":memory:", logger)
	require.NoError(t, err, "Error creating ContainerManager")
	err = cm.Db.InitSchema()
	require.NoError(t, err, "Error initializing database schema")

	server := NewServer(cm, logger)

	// Start server
	go func() {
		err := server.Start(ctx, serverAddr)
		assert.NoError(t, err)
	}()

	// Wait for server to start
	time.Sleep(2 * time.Second)

	// Test creating a new container
	t.Run("CreateNewContainer", func(t *testing.T) {
		payload := WebhookPayload{
			Action:        "create",
			ContainerName: "test-api-container",
			ImageName:     testImage,
			DomainName:    "test.example.com",
			ContainerPort: "8080",
		}

		resp, err := sendWebhook(payload)
		require.NoError(t, err, "Error sending webhook")
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		// Verify container was created
		containers, err := cm.Db.ListContainers()
		require.NoError(t, err, "Error listing containers")

		var found bool
		for _, c := range containers {
			if c.ContainerName == payload.ContainerName {
				found = true
				break
			}
		}
		assert.True(t, found, "Created container was not found")
	})

	// Test updating the container
	t.Run("UpdateExistingContainer", func(t *testing.T) {
		payload := WebhookPayload{
			Action:        "update",
			ContainerName: "test-api-container",
			ImageName:     testImage2,
			DomainName:    "test.example.com",
			ContainerPort: "8080",
		}

		resp, err := sendWebhook(payload)
		require.NoError(t, err, "Error sending webhook")
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		// Verify container was updated
		containers, err := cm.Db.ListContainers()
		require.NoError(t, err, "Error listing containers")

		var found bool
		for _, c := range containers {
			if c.ContainerName == payload.ContainerName && c.ImageName == payload.ImageName {
				found = true
				break
			}
		}
		assert.True(t, found, "Updated container was not found")
	})

	// Cleanup
	t.Cleanup(func() {
		cancel() // Stop the server

		// Use the ContainerManager to list and remove containers and images
		containers, err := cm.Db.ListContainers()
		require.NoError(t, err, "Error listing containers from database")

		for _, c := range containers {
			if c.ContainerName == "test-api-container" {
				err = cm.RemoveContainerAndImage(context.Background(), c.ContainerID)
				assert.NoError(t, err, "Error removing container and image")
				break
			}
		}
	})
}

func sendWebhook(payload WebhookPayload) (*http.Response, error) {
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("error marshalling payload: %w", err)
	}

	resp, err := http.Post(fmt.Sprintf("http://%s/webhook", serverAddr), "application/json", bytes.NewBuffer(jsonPayload))
	if err != nil {
		return nil, fmt.Errorf("error sending webhook: %w", err)
	}

	// Read and log the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response body: %v\n", err)
	} else {
		fmt.Printf("Response body: %s\n", string(body))
	}
	resp.Body.Close()

	return resp, nil
}
