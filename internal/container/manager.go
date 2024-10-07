package container

import (
	"context"
	"fmt"
	"time"

	"github.com/dgunzy/go-container-orchestrator/internal/database"
	"github.com/dgunzy/go-container-orchestrator/internal/health"
	"github.com/dgunzy/go-container-orchestrator/internal/logging"
	"github.com/dgunzy/go-container-orchestrator/pkg/docker"
	"github.com/docker/docker/api/types/strslice"
)

type ContainerManager struct {
	DockerClient  *docker.DockerClient
	portFinder    *portFinder
	Db            *database.Database
	Logger        *logging.Logger
	HealthChecker *health.HealthChecker
	// TODO add nginx
}

type ContainerConfig struct {
	DomainName       string
	ImageName        string
	ContainerName    string
	ContainerID      string
	ContainerPort    string
	HostPort         string
	RegistryUsername string
	RegistryPassword string
	Cmd              strslice.StrSlice
	Status           string
}

func NewContainerManager(DockerClient *docker.DockerClient, DbPath string, logger *logging.Logger, healthChecker *health.HealthChecker) (*ContainerManager, error) {
	Db, err := database.NewDatabase(DbPath)
	if err != nil {
		return nil, fmt.Errorf("error creating database: %w", err)
	}
	return &ContainerManager{
		DockerClient:  DockerClient,
		portFinder:    newPortFinder(),
		Db:            Db,
		Logger:        logger,
		HealthChecker: healthChecker,
	}, nil
}

func (cm *ContainerManager) CreateNewContainer(ctx context.Context, config *ContainerConfig) error {
	cm.Logger.Info("Creating new container: %s", config.ContainerName)

	if err := cm.pullImage(ctx, config); err != nil {
		return err
	}

	newHostPort, err := cm.portFinder.findAvailablePort()
	if err != nil {
		return fmt.Errorf("error finding available port: %w", err)
	}

	containerInfo, err := cm.createAndStartContainer(ctx, config, newHostPort)
	if err != nil {
		return err
	}

	if err := cm.Db.AddContainer(*containerInfo); err != nil {
		cm.Logger.Error("Error saving container info to database: %s", err)
		return fmt.Errorf("error saving container info to database: %w", err)
	}

	//TODO add nginx config to expose this internal port to proxy
	return nil
}

// the config can just contain the container name and the new image name
func (cm *ContainerManager) UpdateExistingContainer(ctx context.Context, config *ContainerConfig) error {
	cm.Logger.Info("Updating existing container: %s", config.ContainerName)

	oldContainerInfo, err := cm.getContainerInfo(config)
	if err != nil {
		return fmt.Errorf("error getting old container info: %w", err)
	}

	if err := cm.pullImage(ctx, config); err != nil {
		return err
	}

	newContainerInfo, err := cm.createAndStartNewContainer(ctx, config)
	if err != nil {
		return err
	}

	// TODO: Implement health check
	cm.Logger.Info("TODO: Implement health check")

	// TODO: Update Nginx config to use newHostPort
	cm.Logger.Info("TODO: Update Nginx configuration")

	if err := cm.stopAndRemoveContainer(ctx, oldContainerInfo.ContainerID); err != nil {
		cm.Logger.Error("Error stopping/removing old container: %s", err)
		// Continue with the update process even if this fails
	}

	if err := cm.updateDatabase(oldContainerInfo, newContainerInfo); err != nil {
		return err
	}

	cm.Logger.Info("Container update completed: %s", newContainerInfo.ContainerName)
	return nil
}

func (cm *ContainerManager) RemoveContainer(ctx context.Context, containerID string) error {
	cm.Logger.Info("Removing container: %s", containerID)

	if err := cm.stopAndRemoveContainer(ctx, containerID); err != nil {
		return err
	}

	if err := cm.Db.DeleteContainer(containerID); err != nil {
		cm.Logger.Error("Error removing container info from database: %s", err)
		return fmt.Errorf("error removing container info from database: %w", err)
	}

	cm.Logger.Info("Container removed successfully: %s", containerID)
	return nil
}

func (cm *ContainerManager) LoadAndStartContainers(ctx context.Context, Cmd strslice.StrSlice) error {
	cm.Logger.Info("Loading and starting containers from database")

	containers, err := cm.Db.ListContainers()
	if err != nil {
		cm.Logger.Error("Error listing containers from database: %s", err)
		return fmt.Errorf("error listing containers from database: %w", err)
	}
	if containers == nil {
		cm.Logger.Warn("No containers to load")
		return nil
	}

	for _, containerInfo := range containers {
		config := &ContainerConfig{
			ImageName:     containerInfo.ImageName,
			DomainName:    containerInfo.DomainName,
			ContainerName: containerInfo.ContainerName,
			ContainerPort: containerInfo.ContainerPort,
			Cmd:           Cmd,
		}

		newContainerInfo, err := cm.createAndStartContainer(ctx, config, containerInfo.HostPort)
		if err != nil {
			cm.Logger.Error("Error creating/starting container %s: %s", containerInfo.ContainerName, err)
			continue
		}

		if err := cm.Db.DeleteContainer(containerInfo.ContainerID); err != nil {
			cm.Logger.Error("Error removing old container info from database: %s", err)
			continue
		}

		if err := cm.Db.AddContainer(*newContainerInfo); err != nil {
			cm.Logger.Error("Error saving new container info to database: %s", err)
			return fmt.Errorf("error saving new container info to database: %w", err)
		}
	}

	cm.Logger.Info("Finished loading and starting containers")
	return nil
}

func (cm *ContainerManager) RemoveContainerAndImage(ctx context.Context, containerID string) error {
	cm.Logger.Info("Removing container and image: %s", containerID)

	containerInfo, err := cm.Db.GetContainer(containerID)
	if err != nil {
		cm.Logger.Error("Error getting container info: %s", err)
		return fmt.Errorf("error getting container info: %w", err)
	}

	if err := cm.RemoveContainer(ctx, containerID); err != nil {
		return err
	}

	if err := cm.DockerClient.RemoveImage(ctx, containerInfo.ImageName); err != nil {
		cm.Logger.Error("Error removing image: %s", err)
		return fmt.Errorf("error removing image: %w", err)
	}

	cm.Logger.Info("Container and image removed successfully container: %s image: %s", containerID, containerInfo.ImageName)
	return nil
}

func (cm *ContainerManager) ListContainers() ([]ContainerConfig, error) {
	dbContainers, err := cm.Db.ListContainers()
	if err != nil {
		cm.Logger.Error("Error listing containers from database: %s", err)
		return nil, fmt.Errorf("error listing containers from database: %w", err)
	}
	var containers []ContainerConfig
	for _, c := range dbContainers {
		containers = append(containers, ContainerConfig{
			DomainName:    c.DomainName,
			ImageName:     c.ImageName,
			ContainerName: c.ContainerName,
			ContainerID:   c.ContainerID,
			ContainerPort: c.ContainerPort,
			HostPort:      c.HostPort,
			Status:        cm.ContainerStatus(c.ContainerID),
		})
	}
	return containers, nil
}

func (cm *ContainerManager) ContainerStatus(containerId string) string {
	state, err := cm.DockerClient.HealthCheck(context.Background(), containerId)
	if err != nil {
		cm.Logger.Error("Error checking container %s: %s", containerId, err)
		return ""
	}
	return state.Status
}
func (cm *ContainerManager) RunAsDaemon(ctx context.Context) error {
	cm.Logger.Info("Starting ContainerManager daemon")

	// Initialize the HealthChecker if it's not already initialized
	if cm.HealthChecker == nil {
		cm.HealthChecker = health.NewHealthChecker(cm.DockerClient, cm.Db, 1*time.Minute, cm.Logger)
	}

	// Start the health checker in a separate goroutine
	healthCheckerCtx, healthCheckerCancel := context.WithCancel(ctx)
	defer healthCheckerCancel()
	go cm.HealthChecker.Start(healthCheckerCtx)

	// Main daemon loop
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			cm.Logger.Info("Shutting down ContainerManager daemon")
			return nil
		case <-ticker.C:
			// Perform periodic tasks here, if any
			cm.Logger.Debug("ContainerManager daemon is running")
		}
	}
}
