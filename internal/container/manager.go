package container

import (
	"context"
	"fmt"
	"time"

	"github.com/dgunzy/go-container-orchestrator/internal/database"
	"github.com/dgunzy/go-container-orchestrator/internal/logging"
	"github.com/dgunzy/go-container-orchestrator/pkg/docker"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/strslice"
	"github.com/docker/go-connections/nat"
)

type ContainerManager struct {
	DockerClient *docker.DockerClient
	PortFinder   *portFinder
	Db           *database.Database
	Logger       *logging.Logger
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
}

func NewContainerManager(DockerClient *docker.DockerClient, DbPath string, logger *logging.Logger) (*ContainerManager, error) {

	Db, err := database.NewDatabase(DbPath)
	if err != nil {
		return nil, fmt.Errorf("error creating database: %w", err)
	}
	return &ContainerManager{
		DockerClient: DockerClient,
		PortFinder:   newPortFinder(),
		Db:           Db,
		Logger:       logger,
	}, nil
}

func (cm *ContainerManager) CreateNewContainer(ctx context.Context, config *ContainerConfig) error {
	cm.Logger.Info("Creating new container: %s", config.ContainerName)
	err := cm.DockerClient.PullImageFromPrivateRegistry(ctx,
		config.ImageName,
		config.RegistryUsername,
		config.RegistryPassword)

	if err != nil {
		cm.Logger.Error("Error pulling image: %s", err)
		return fmt.Errorf("error pulling image: %w", err)
	}

	newHostPort, err := cm.PortFinder.findAvailablePort()
	if err != nil {
		cm.Logger.Error("Error finding available port: %s", err)
		return fmt.Errorf("error finding available port: %w", err)
	}

	containerConfig := &container.Config{
		Image:      config.ImageName,
		Domainname: config.DomainName,
		Cmd:        config.Cmd,
	}
	hostConfig := &container.HostConfig{
		PortBindings: nat.PortMap{
			nat.Port(config.ContainerPort): []nat.PortBinding{
				{
					HostIP:   "0.0.0.0",
					HostPort: newHostPort,
				},
			},
		},
	}
	response, err := cm.DockerClient.CreateContainer(ctx, containerConfig, hostConfig, &network.NetworkingConfig{}, config.ContainerName)
	if err != nil {
		cm.Logger.Error("Error creating container: %s", err)
		return fmt.Errorf("error creating container: %w", err)
	}

	cm.Logger.Info("Created container with ID: %s", response.ID)

	err = cm.DockerClient.StartContainer(ctx, response.ID)
	if err != nil {
		cm.Logger.Error("Error starting container: %s", err)
		return fmt.Errorf("error starting container: %w", err)
	}

	cm.Logger.Info("Container started successfully")

	// Save container info to database
	containerInfo := database.ContainerInfo{
		ContainerID:   response.ID,
		ContainerName: config.ContainerName,
		ImageName:     config.ImageName,
		DomainName:    config.DomainName,
		HostPort:      newHostPort,
		ContainerPort: config.ContainerPort,
		Status:        "running",
	}
	err = cm.Db.AddContainer(containerInfo)
	if err != nil {
		cm.Logger.Error("Error saving container info to database: %s", err)
		return fmt.Errorf("error saving container info to database: %w", err)
	}

	//TODO add nginx config to expose this internal port to proxy
	return nil
}

func (cm *ContainerManager) UpdateExistingContainer(ctx context.Context, config *ContainerConfig) error {
	cm.Logger.Info("Updating existing container: %s", config.ContainerID)
	timestamp := time.Now().Format("20060102150405")
	newContainerName := fmt.Sprintf("%s_%s", config.ContainerName, timestamp)

	err := cm.DockerClient.PullImageFromPrivateRegistry(ctx,
		config.ImageName,
		config.RegistryUsername,
		config.RegistryPassword)
	if err != nil {
		cm.Logger.Error("Error pulling updated image: %s", err)
		return fmt.Errorf("error pulling updated image: %w", err)
	}

	newHostPort, err := cm.PortFinder.findAvailablePort()
	if err != nil {
		cm.Logger.Error("Error finding available port: %s", err)
		return fmt.Errorf("error finding available port: %w", err)
	}

	containerConfig := &container.Config{
		Image:      config.ImageName,
		Domainname: config.DomainName,
		Cmd:        config.Cmd,
	}
	hostConfig := &container.HostConfig{
		PortBindings: nat.PortMap{
			nat.Port(config.ContainerPort): []nat.PortBinding{
				{
					HostIP:   "0.0.0.0",
					HostPort: newHostPort,
				},
			},
		},
	}
	response, err := cm.DockerClient.CreateContainer(ctx, containerConfig, hostConfig, &network.NetworkingConfig{}, newContainerName)
	if err != nil {
		cm.Logger.Error("Error creating new container: %s", err)
		return fmt.Errorf("error creating new container: %w", err)
	}

	err = cm.DockerClient.StartContainer(ctx, response.ID)
	if err != nil {
		cm.Logger.Error("Error starting new container: %s", err)
		return fmt.Errorf("error starting new container: %w", err)
	}
	cm.Logger.Info("New container started: %s", newContainerName)

	time.Sleep(2 * time.Second)

	// TODO: Implement health check
	cm.Logger.Info("TODO: Implement health check")

	// TODO: Update Nginx config to use newHostPort
	cm.Logger.Info("TODO: Update Nginx configuration")

	oldContainerInfo, err := cm.Db.GetContainer(config.ContainerID)
	if err != nil {
		cm.Logger.Error("Error getting old container info: %s", err)
		return fmt.Errorf("error getting old container info: %w", err)
	}

	err = cm.DockerClient.StopContainer(ctx, oldContainerInfo.ContainerID, nil)
	if err != nil {
		cm.Logger.Error("Error stopping old container: %s", err)
	}

	err = cm.DockerClient.RemoveContainer(ctx, oldContainerInfo.ContainerID, container.RemoveOptions{Force: true})
	if err != nil {
		cm.Logger.Error("Error removing old container: %s", err)
	}

	// Update database with new container information
	newContainerInfo := database.ContainerInfo{
		ContainerID:   response.ID,
		ContainerName: newContainerName,
		ImageName:     config.ImageName,
		DomainName:    config.DomainName,
		HostPort:      newHostPort,
		ContainerPort: config.ContainerPort,
		Status:        "running",
	}
	err = cm.Db.AddContainer(newContainerInfo)
	if err != nil {
		cm.Logger.Error("Error saving new container info to database: %s", err)
		return fmt.Errorf("error saving new container info to database: %w", err)
	}

	err = cm.Db.DeleteContainer(oldContainerInfo.ContainerID)
	if err != nil {
		cm.Logger.Error("Error removing old container info from database: %s", err)
		return fmt.Errorf("error removing old container info from database: %w", err)
	}

	cm.Logger.Info("Container update completed: %s", newContainerName)
	return nil
}

func (cm *ContainerManager) RemoveContainer(ctx context.Context, containerID string) error {
	cm.Logger.Info("Removing container: %s", containerID)

	err := cm.DockerClient.StopContainer(ctx, containerID, nil)
	if err != nil {
		cm.Logger.Error("Error stopping container: %s", err)
		// Continue with removal even if stop fails
	}

	err = cm.DockerClient.RemoveContainer(ctx, containerID, container.RemoveOptions{Force: true})
	if err != nil {
		cm.Logger.Error("Error removing container: %s", err)
		return fmt.Errorf("error removing container: %w", err)
	}

	err = cm.Db.DeleteContainer(containerID)
	if err != nil {
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
		containerConfig := &container.Config{
			Image:      containerInfo.ImageName,
			Domainname: containerInfo.DomainName,
			Cmd:        Cmd,
		}
		hostConfig := &container.HostConfig{
			PortBindings: nat.PortMap{
				nat.Port(containerInfo.ContainerPort): []nat.PortBinding{
					{
						HostIP:   "0.0.0.0",
						HostPort: containerInfo.HostPort,
					},
				},
			},
		}

		response, err := cm.DockerClient.CreateContainer(ctx, containerConfig, hostConfig, &network.NetworkingConfig{}, containerInfo.ContainerName)
		if err != nil {
			cm.Logger.Error("Error creating container %s: %s", containerInfo.ContainerID, err)
			continue
		}

		err = cm.DockerClient.StartContainer(ctx, response.ID)

		if err != nil {
			cm.Logger.Error("Error starting container %s: %s", containerInfo.ContainerID, err)
			continue
		}

		cm.Logger.Info("Container started successfully: %s", containerInfo.ContainerID)

		err = cm.Db.DeleteContainer(containerInfo.ContainerID)
		if err != nil {
			cm.Logger.Error("Error removing container info from database: %s", err)
			continue
		}

		containerInfo := database.ContainerInfo{
			ContainerID:   response.ID,
			ContainerName: containerInfo.ContainerName,
			ImageName:     containerInfo.ImageName,
			DomainName:    containerInfo.DomainName,
			HostPort:      containerInfo.HostPort,
			ContainerPort: containerInfo.ContainerPort,
			Status:        "running",
		}
		err = cm.Db.AddContainer(containerInfo)
		if err != nil {
			cm.Logger.Error("Error saving new container info to database: %s", err)
			return fmt.Errorf("error saving new container info to database: %w", err)
		}
	}

	cm.Logger.Info("Finished loading and starting containers")
	return nil
}

func (cm *ContainerManager) RemoveContainerAndImage(ctx context.Context, containerID string) error {
	cm.Logger.Info("Removing container: %s", containerID)

	// Get container info for image removal
	containerInfo, err := cm.Db.GetContainer(containerID)
	if err != nil {
		cm.Logger.Error("Error getting container info: %s", err)
		return fmt.Errorf("error getting container info: %w", err)
	}

	err = cm.DockerClient.StopContainer(ctx, containerID, nil)
	if err != nil {
		cm.Logger.Error("Error stopping container: %s", err)
		// Continue with removal even if stop fails
	}

	err = cm.DockerClient.RemoveContainer(ctx, containerID, container.RemoveOptions{Force: true})
	if err != nil {
		cm.Logger.Error("Error removing container: %s", err)
		return fmt.Errorf("error removing container: %w", err)
	}

	err = cm.Db.DeleteContainer(containerID)
	if err != nil {
		cm.Logger.Error("Error removing container info from database: %s", err)
		return fmt.Errorf("error removing container info from database: %w", err)
	}

	err = cm.DockerClient.RemoveImage(ctx, containerInfo.ImageName)
	if err != nil {
		cm.Logger.Error("Error removing image: %s", err)
		return fmt.Errorf("error removing image: %w", err)
	}

	cm.Logger.Info("Container and image removed successfully container: %s image: %s", containerID, containerInfo.ImageName)
	return nil
}
