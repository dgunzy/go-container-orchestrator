package health

import (
	"context"
	"sync"
	"time"

	"github.com/dgunzy/go-container-orchestrator/internal/database"
	"github.com/dgunzy/go-container-orchestrator/internal/logging"
	"github.com/dgunzy/go-container-orchestrator/pkg/docker"
	"github.com/docker/docker/api/types"
)

type HealthChecker struct {
	dockerClient *docker.DockerClient
	db           *database.Database
	interval     time.Duration
	logger       *logging.Logger
}

func NewHealthChecker(dockerClient *docker.DockerClient, db *database.Database, interval time.Duration) *HealthChecker {
	return &HealthChecker{
		dockerClient: dockerClient,
		db:           db,
		interval:     interval,
		logger:       logging.GetLogger(),
	}
}

func (hc *HealthChecker) Start(ctx context.Context) {
	ticker := time.NewTicker(hc.interval)

	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			hc.checkContainers(ctx)
		}
	}
}

func (hc *HealthChecker) checkContainers(ctx context.Context) {

	containers, err := hc.db.ListContainers()
	if err != nil {
		hc.logger.Error("Error listing containers: %s", err)
		return
	}
	var wg sync.WaitGroup
	for _, c := range containers {
		wg.Add(1)

		go func(container database.ContainerInfo) {
			defer wg.Done()
			if err := hc.checkContainer(ctx, &container); err != nil {
				hc.logger.Error("Error checking container %s: %s", container.ContainerName, err)
			}
		}(c)
		// Add a delay between each container check
		time.Sleep(2 * time.Second)
	}
	wg.Wait()
}

func (hc *HealthChecker) checkContainer(ctx context.Context, container *database.ContainerInfo) error {
	context, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()
	// Check container is running
	state, err := hc.dockerClient.HealthCheck(context, container.ContainerID)
	if err != nil {
		hc.logger.Error("Error checking container %s: %s", container.ContainerName, err)
		return err
	}
	if state.Status != "running" {
		hc.logger.Warn("Container %s is not running, starting.. ", container.ContainerName)
		return hc.dockerClient.StartContainer(ctx, container.ContainerID)
	}
	if state.Health != nil && state.Health.Status != types.Healthy {
		hc.logger.Warn("Container %s is unhealthy. Attempting to restart...", container.ContainerName)
		return hc.dockerClient.RestartContainer(ctx, container.ContainerID, nil)
	}

	hc.logger.Info("Container %s is healthy and running", container.ContainerName)
	return nil
}
