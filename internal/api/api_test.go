package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/dgunzy/go-container-orchestrator/tests"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testImage  = "alpine:latest"
	testImage2 = "alpine:3.12"
	serverAddr = "localhost:8080"
)

func TestAPIServer(t *testing.T) {
	cm := tests.InitTestConfig()
	defer tests.CleanupTestResources(cm.DockerClient)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	server := NewServer(&cm, cm.Logger)

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
		time.Sleep(2 * time.Second)
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

		// Wait a bit for the update to complete
		time.Sleep(2 * time.Second)

		// Verify container was updated
		containers, err := cm.Db.ListContainers()
		require.NoError(t, err, "Error listing containers")

		var found bool
		for _, c := range containers {
			t.Logf("Container: Name=%s, Image=%s", c.ContainerName, c.ImageName)
			if strings.HasPrefix(c.ContainerName, payload.ContainerName) && c.ImageName == payload.ImageName {
				found = true
				break
			}
		}
		assert.True(t, found, "Updated container was not found")

		if !found {
			t.Logf("Expected container with name prefix '%s' and image '%s' not found", payload.ContainerName, payload.ImageName)
			t.Logf("Available containers:")
			for _, c := range containers {
				t.Logf("  - Name: %s, Image: %s", c.ContainerName, c.ImageName)
			}
		}
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
