package tests

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/dgunzy/go-container-orchestrator/internal/container"
	"github.com/dgunzy/go-container-orchestrator/internal/database"
	"github.com/dgunzy/go-container-orchestrator/internal/health"
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
	err = logging.Setup("../../container_manager_logs")
	if err != nil {
		fmt.Printf("Failed to set up logging: %v\n", err)
		os.Exit(1)
	}

	logger := logging.GetLogger()

	dockerClient, err := internal_client.NewClient()
	if err != nil {
		fmt.Printf("Error creating Docker client: %v\n", err)
		os.Exit(1)
	}

	db, err := database.NewDatabase(":memory:")
	if err != nil {
		fmt.Printf("Error creating database: %v\n", err)
		os.Exit(1)
	}
	healthChecker := health.NewHealthChecker(dockerClient, db, 5*time.Minute, logger)

	cm, err := container.NewContainerManager(dockerClient, ":memory:", logger, healthChecker)
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
