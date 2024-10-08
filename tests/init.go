package tests

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/dgunzy/go-container-orchestrator/internal/container"
	"github.com/dgunzy/go-container-orchestrator/internal/logging"

	internal_client "github.com/dgunzy/go-container-orchestrator/pkg/docker"
	docker_container "github.com/docker/docker/api/types/container"
)

// returns an initialized struct of type DockerClient
func InitTestConfig() container.ContainerManager {
	// Remove old logs
	err := os.RemoveAll("../../container_manager_logs")
	if err != nil {
		fmt.Printf("Failed to remove old logs: %v\n", err)
	}

	cm, err := container.NewContainerManager()
	if err != nil {
		fmt.Printf("Error creating ContainerManager: %v\n", err)
		os.Exit(1)
	}
	cm.Db.InitSchema()

	return *cm
}

func CleanupTestResources(dockerClient *internal_client.DockerClient) {
	ctx := context.Background()

	// Remove test containers
	containers, err := dockerClient.ListContainers(ctx)
	if err == nil {
		for _, container := range containers {
			if strings.Contains(container.Names[0], "test") {
				_ = dockerClient.RemoveContainer(ctx, container.ID, docker_container.RemoveOptions{Force: true})
			}
		}
	}

	// Remove Alpine images
	images, err := dockerClient.ListImages(ctx)
	if err == nil {
		for _, image := range images {
			for _, tag := range image.RepoTags {
				if strings.HasPrefix(tag, "alpine:") {
					_ = dockerClient.RemoveImage(ctx, tag)
				}
			}
		}
	}
	logging.CloseGlobalLogger()

}
