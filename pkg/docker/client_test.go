package docker

import (
	"context"
	"io"
	"testing"
	"time"

	"github.com/dgunzy/go-container-orchestrator/config"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient(t *testing.T) {
	client, err := NewClient()
	if err != nil {
		t.Fatalf("Error creating client: %s", err)
	}
	if client == nil {
		t.Fatalf("Client is nil")
	}
}

func TestListContainers(t *testing.T) {
	client, err := NewClient()
	if err != nil {
		t.Fatalf("Error creating client: %s", err)
	}
	containers, err := client.ListContainers(context.Background())
	if err != nil {
		t.Fatalf("Error listing containers: %s", err)
	}
	t.Logf("Found %d containers", len(containers))
	for _, container := range containers {
		t.Logf("Container Image: %s, Container Name %s", container.Image, container.Names)
	}
}

func TestCreateAndRemoveContainer(t *testing.T) {
	client, err := NewClient()
	if err != nil {
		t.Fatalf("Error creating client: %s", err)
	}
	ctx := context.Background()

	response, err := client.CreateContainer(ctx, &container.Config{
		Image: "go-capstone-app",
	}, nil, &network.NetworkingConfig{}, "test-container")

	if err != nil {
		t.Fatalf("Error creating container: %s", err)

	}

	err = client.RemoveContainer(ctx, response.ID, container.RemoveOptions{})
	if err != nil {
		t.Fatalf("Error removing container: %s", err)
	}

}

func TestPullContainers(t *testing.T) {
	getEnv, err := config.LoadEnv()
	if err != nil {
		t.Fatalf("Error loading env: %s", err)
	}
	client, err := NewClient()
	if err != nil {
		t.Fatalf("Error creating client: %s", err)
	}
	ctx := context.Background()

	ctx, cancel := context.WithTimeout(ctx, 40*time.Second)
	defer cancel()

	userName := getEnv("TEST_DOCKER_USERNAME")
	password := getEnv("TEST_DOCKER_PASSWORD")
	imageName := getEnv("TEST_DOCKER_IMAGE")
	repoName := getEnv("TEST_DOCKER_REPO")
	// registryName := getEnv("TEST_DOCKER_REGISTRY")
	// For non docker hub images this needs to be set to the registry name

	if userName == "" || password == "" {
		t.Fatalf("Username or password not set")

	}

	fullImageName := repoName + ":" + imageName
	err = client.PullImageFromPrivateRegistry(ctx, fullImageName, userName, password)

	if err != nil {
		t.Fatalf("Error pulling image: %s", err)
	}

	err = client.RemoveImage(ctx, fullImageName)
	if err != nil {
		t.Fatalf("Error removing image: %s", err)
	}

}

func TestPullAndRunContainer(t *testing.T) {
	getEnv, err := config.LoadEnv()
	if err != nil {
		t.Fatalf("Error loading env: %s", err)
	}
	client, err := NewClient()
	if err != nil {
		t.Fatalf("Error creating client: %s", err)
	}
	ctx := context.Background()

	ctx, cancel := context.WithTimeout(ctx, 40*time.Second)
	defer cancel()

	userName := getEnv("TEST_DOCKER_USERNAME")
	password := getEnv("TEST_DOCKER_PASSWORD")
	imageName := getEnv("TEST_DOCKER_IMAGE")
	repoName := getEnv("TEST_DOCKER_REPO")
	// registryName := getEnv("TEST_DOCKER_REGISTRY")
	// For non docker hub images this needs to be set to the registry name

	fullImageName := repoName + ":" + imageName
	err = client.PullImageFromPrivateRegistry(ctx, fullImageName, userName, password)

	if err != nil {
		t.Fatalf("Error pulling image: %s", err)
	}

	response, err := client.CreateContainer(ctx, &container.Config{
		Image: fullImageName,
	}, nil, &network.NetworkingConfig{}, "test-container")

	if err != nil {
		t.Fatalf("Error creating container: %s", err)

	}

	err = client.RemoveContainer(ctx, response.ID, container.RemoveOptions{})
	if err != nil {
		t.Fatalf("Error removing container: %s", err)
	}
	err = client.RemoveImage(ctx, fullImageName)
	if err != nil {
		t.Fatalf("Error removing image: %s", err)
	}
}

const testImage = "alpine:latest"

func setupTest(t *testing.T) *DockerClient {
	client, err := NewClient()
	require.NoError(t, err)

	ctx := context.Background()

	// Pull the test image
	reader, err := client.client.ImagePull(ctx, testImage, image.PullOptions{})
	require.NoError(t, err)
	_, err = io.Copy(io.Discard, reader)
	require.NoError(t, err)
	reader.Close()

	return client
}

func TestStartStopRestartContainer(t *testing.T) {
	client := setupTest(t)
	ctx := context.Background()

	t.Log("Creating test container")
	resp, err := client.CreateContainer(ctx, &container.Config{
		Image: testImage,
		Cmd:   []string{"sleep", "30"},
	}, nil, nil, "test-container")
	require.NoError(t, err)
	defer func() {
		t.Log("Removing test container")
		err := client.RemoveContainer(ctx, resp.ID, container.RemoveOptions{Force: true})
		if err != nil {
			t.Logf("Error removing container: %v", err)
		}
	}()

	// Test StartContainer
	t.Log("Starting container")
	err = client.StartContainer(ctx, resp.ID)
	assert.NoError(t, err)

	// Check if container is running
	t.Log("Checking if container is running")
	state, err := client.HealthCheck(ctx, resp.ID)
	assert.NoError(t, err)
	t.Logf("Container state: %s", state.Status)
	assert.Equal(t, "running", state.Status)

	// Test StopContainer
	t.Log("Stopping container")
	timeout := 10
	err = client.StopContainer(ctx, resp.ID, &timeout)
	assert.NoError(t, err)

	// Check if container is stopped
	t.Log("Checking if container is stopped")
	state, err = client.HealthCheck(ctx, resp.ID)
	assert.NoError(t, err)
	t.Logf("Container state: %s", state.Status)
	assert.Equal(t, "exited", state.Status)

	// Test RestartContainer
	t.Log("Restarting container")
	err = client.RestartContainer(ctx, resp.ID, &timeout)
	assert.NoError(t, err)

	// Check if container is running again
	t.Log("Checking if container is running after restart")
	state, err = client.HealthCheck(ctx, resp.ID)
	assert.NoError(t, err)
	t.Logf("Container state: %s", state.Status)
	assert.Equal(t, "running", state.Status)
}

func TestGetContainerLogs(t *testing.T) {
	client := setupTest(t)
	ctx := context.Background()

	// Create and start a test container that outputs logs
	resp, err := client.CreateContainer(ctx, &container.Config{
		Image: testImage,
		Cmd:   []string{"echo", "test log message"},
	}, nil, nil, "test-log-container")
	require.NoError(t, err)
	defer client.RemoveContainer(ctx, resp.ID, container.RemoveOptions{Force: true})

	err = client.StartContainer(ctx, resp.ID)
	require.NoError(t, err)

	// Wait for the container to finish
	time.Sleep(2 * time.Second)

	// Test GetContainerLogs
	logs, err := client.GetContainerLogs(ctx, resp.ID)
	assert.NoError(t, err)
	defer logs.Close()

	// Read and check logs
	logContent, err := io.ReadAll(logs)
	assert.NoError(t, err)
	assert.Contains(t, string(logContent), "test log message")
}

func TestExecuteContainerCommand(t *testing.T) {
	client := setupTest(t)
	ctx := context.Background()

	// Create and start a test container
	resp, err := client.CreateContainer(ctx, &container.Config{
		Image: testImage,
		Cmd:   []string{"sleep", "30"},
	}, nil, nil, "test-exec-container")
	require.NoError(t, err)
	defer client.RemoveContainer(ctx, resp.ID, container.RemoveOptions{Force: true})

	err = client.StartContainer(ctx, resp.ID)
	require.NoError(t, err)

	// Test ExecuteContainerCommand
	execResp, err := client.ExecuteContainerCommand(ctx, resp.ID, []string{"echo", "hello world"})
	assert.NoError(t, err)
	assert.NotEmpty(t, execResp.ID)

	// You might want to add more assertions here to check the output of the executed command
}

func TestHealthCheck(t *testing.T) {
	client := setupTest(t)
	ctx := context.Background()

	// Create and start a test container with a health check
	resp, err := client.CreateContainer(ctx, &container.Config{
		Image: testImage,
		Cmd:   []string{"sleep", "30"},
		Healthcheck: &container.HealthConfig{
			Test:     []string{"CMD", "echo", "healthy"},
			Interval: time.Duration(1) * time.Second,
		},
	}, nil, nil, "test-health-container")
	require.NoError(t, err)
	defer client.RemoveContainer(ctx, resp.ID, container.RemoveOptions{Force: true})

	err = client.StartContainer(ctx, resp.ID)
	require.NoError(t, err)

	// Wait for health check to run
	time.Sleep(2 * time.Second)

	// Test HealthCheck
	state, err := client.HealthCheck(ctx, resp.ID)
	assert.NoError(t, err)
	assert.NotNil(t, state.Health)
	assert.Equal(t, "healthy", state.Health.Status)
}
