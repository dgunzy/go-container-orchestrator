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
	dockerClient *docker.DockerClient
	portFinder   *portFinder
	db           *database.Database
	logger       *logging.Logger
	// TODO add nginx
}

type ContainerConfig struct {
	DomainName       string
	ImageName        string
	ContainerName    string
	ContainerPort    string
	HostPort         string
	RegistryUsername string
	RegistryPassword string
	Cmd              strslice.StrSlice
}

func NewContainerManager(dockerClient *docker.DockerClient, dbPath string) (*ContainerManager, error) {
	logger := logging.GetLogger()
	db, err := database.NewDatabase(dbPath)
	if err != nil {
		return nil, fmt.Errorf("error creating database: %w", err)
	}
	return &ContainerManager{
		dockerClient: dockerClient,
		portFinder:   newPortFinder(),
		db:           db,
		logger:       logger,
	}, nil
}

func (cm *ContainerManager) CreateNewContainer(ctx context.Context, config *ContainerConfig) error {
	cm.logger.Infof("Creating new container: %s", config.ContainerName)
	err := cm.dockerClient.PullImageFromPrivateRegistry(ctx,
		config.ImageName,
		config.RegistryUsername,
		config.RegistryPassword)

	if err != nil {
		cm.logger.Errorf("Error pulling image: %s", err)
		return fmt.Errorf("error pulling image: %w", err)
	}

	newHostPort, err := cm.portFinder.findAvailablePort()
	if err != nil {
		cm.logger.Errorf("Error finding available port: %s", err)
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
	response, err := cm.dockerClient.CreateContainer(ctx, containerConfig, hostConfig, &network.NetworkingConfig{}, config.ContainerName)
	if err != nil {
		cm.logger.Errorf("Error creating container: %s", err)
		return fmt.Errorf("error creating container: %w", err)
	}

	cm.logger.Infof("Created container with ID: %s", response.ID)

	err = cm.dockerClient.StartContainer(ctx, response.ID)
	if err != nil {
		cm.logger.Errorf("Error starting container: %s", err)
		return fmt.Errorf("error starting container: %w", err)
	}

	cm.logger.Info("Container started successfully")

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
	err = cm.db.AddContainer(containerInfo)
	if err != nil {
		cm.logger.Errorf("Error saving container info to database: %s", err)
		return fmt.Errorf("error saving container info to database: %w", err)
	}

	//TODO add nginx config to expose this internal port to proxy
	return nil
}

func (cm *ContainerManager) UpdateExistingContainer(ctx context.Context, config *ContainerConfig) error {
	cm.logger.Infof("Updating existing container: %s", config.ContainerName)
	timestamp := time.Now().Format("20060102150405")
	newContainerName := fmt.Sprintf("%s_%s", config.ContainerName, timestamp)

	err := cm.dockerClient.PullImageFromPrivateRegistry(ctx,
		config.ImageName,
		config.RegistryUsername,
		config.RegistryPassword)
	if err != nil {
		cm.logger.Errorf("Error pulling updated image: %s", err)
		return fmt.Errorf("error pulling updated image: %w", err)
	}

	newHostPort, err := cm.portFinder.findAvailablePort()
	if err != nil {
		cm.logger.Errorf("Error finding available port: %s", err)
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
	response, err := cm.dockerClient.CreateContainer(ctx, containerConfig, hostConfig, &network.NetworkingConfig{}, newContainerName)
	if err != nil {
		cm.logger.Errorf("Error creating new container: %s", err)
		return fmt.Errorf("error creating new container: %w", err)
	}

	err = cm.dockerClient.StartContainer(ctx, response.ID)
	if err != nil {
		cm.logger.Errorf("Error starting new container: %s", err)
		return fmt.Errorf("error starting new container: %w", err)
	}
	cm.logger.Infof("New container started: %s", newContainerName)

	time.Sleep(2 * time.Second)

	// TODO: Implement health check
	cm.logger.Info("TODO: Implement health check")

	// TODO: Update Nginx config to use newHostPort
	cm.logger.Info("TODO: Update Nginx configuration")

	oldContainerInfo, err := cm.db.GetContainer(config.ContainerName)
	if err != nil {
		cm.logger.Errorf("Error getting old container info: %s", err)
		return fmt.Errorf("error getting old container info: %w", err)
	}

	err = cm.dockerClient.StopContainer(ctx, oldContainerInfo.ContainerID, nil)
	if err != nil {
		cm.logger.Errorf("Error stopping old container: %s", err)
	}

	err = cm.dockerClient.RemoveContainer(ctx, oldContainerInfo.ContainerID, container.RemoveOptions{Force: true})
	if err != nil {
		cm.logger.Errorf("Error removing old container: %s", err)
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
	err = cm.db.AddContainer(newContainerInfo)
	if err != nil {
		cm.logger.Errorf("Error saving new container info to database: %s", err)
		return fmt.Errorf("error saving new container info to database: %w", err)
	}

	err = cm.db.DeleteContainer(oldContainerInfo.ContainerID)
	if err != nil {
		cm.logger.Errorf("Error removing old container info from database: %s", err)
		return fmt.Errorf("error removing old container info from database: %w", err)
	}

	cm.logger.Infof("Container update completed: %s", newContainerName)
	return nil
}

func (cm *ContainerManager) RemoveContainer(ctx context.Context, containerID string) error {
	cm.logger.Infof("Removing container: %s", containerID)

	err := cm.dockerClient.StopContainer(ctx, containerID, nil)
	if err != nil {
		cm.logger.Errorf("Error stopping container: %s", err)
		// Continue with removal even if stop fails
	}

	err = cm.dockerClient.RemoveContainer(ctx, containerID, container.RemoveOptions{Force: true})
	if err != nil {
		cm.logger.Errorf("Error removing container: %s", err)
		return fmt.Errorf("error removing container: %w", err)
	}

	err = cm.db.DeleteContainer(containerID)
	if err != nil {
		cm.logger.Errorf("Error removing container info from database: %s", err)
		return fmt.Errorf("error removing container info from database: %w", err)
	}

	cm.logger.Infof("Container removed successfully: %s", containerID)
	return nil
}

func (cm *ContainerManager) LoadAndStartContainers(ctx context.Context) error {
	cm.logger.Info("Loading and starting containers from database")

	containers, err := cm.db.ListContainers()
	if err != nil {
		cm.logger.Errorf("Error listing containers from database: %s", err)
		return fmt.Errorf("error listing containers from database: %w", err)
	}

	for _, containerInfo := range containers {
		containerConfig := &container.Config{
			Image:      containerInfo.ImageName,
			Domainname: containerInfo.DomainName,
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

		response, err := cm.dockerClient.CreateContainer(ctx, containerConfig, hostConfig, &network.NetworkingConfig{}, containerInfo.ContainerName)
		if err != nil {
			cm.logger.Errorf("Error creating container %s: %s", containerInfo.ContainerName, err)
			continue
		}

		err = cm.dockerClient.StartContainer(ctx, response.ID)
		if err != nil {
			cm.logger.Errorf("Error starting container %s: %s", containerInfo.ContainerName, err)
			continue
		}

		cm.logger.Infof("Container started successfully: %s", containerInfo.ContainerName)

		// Update container ID in database if it has changed
		if response.ID != containerInfo.ContainerID {
			containerInfo.ContainerID = response.ID
			err = cm.db.UpdateContainerStatus(containerInfo.ContainerID, "running")
			if err != nil {
				cm.logger.Errorf("Error updating container status in database: %s", err)
			}
		}
	}

	cm.logger.Info("Finished loading and starting containers")
	return nil
}
