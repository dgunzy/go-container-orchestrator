package docker

import (
	"context"
	"testing"

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
