package docker

import (
	"context"
	"testing"
	"time"

	"github.com/dgunzy/go-container-orchestrator/config"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
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

	if userName == "" || password == "" {
		t.Fatalf("Username or password not set")

	}

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
