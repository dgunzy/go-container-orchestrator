package container

import (
	"context"
	"fmt"
	"time"

	"github.com/dgunzy/go-container-orchestrator/internal/database"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/go-connections/nat"
)

func (cm *ContainerManager) getContainerInfo(config *ContainerConfig) (*database.ContainerInfo, error) {
	if config.ContainerID != "" {
		return cm.Db.GetContainer(config.ContainerID)
	}
	containers, err := cm.Db.GetContainersByPartialName(config.ContainerName)
	if err != nil || len(containers) == 0 || len(containers) > 1 {
		return nil, fmt.Errorf("error getting container info: %w", err)
	}
	return &containers[0], nil
}

func (cm *ContainerManager) pullImage(ctx context.Context, config *ContainerConfig) error {
	err := cm.DockerClient.PullImageFromPrivateRegistry(ctx,
		config.ImageName,
		config.RegistryUsername,
		config.RegistryPassword)
	if err != nil {
		cm.Logger.Error("Error pulling image: %s", err)
		return fmt.Errorf("error pulling image: %w", err)
	}
	return nil
}

func (cm *ContainerManager) createAndStartContainer(ctx context.Context, config *ContainerConfig, hostPort string) (*database.ContainerInfo, error) {
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
					HostPort: hostPort,
				},
			},
		},
	}
	response, err := cm.DockerClient.CreateContainer(ctx, containerConfig, hostConfig, &network.NetworkingConfig{}, config.ContainerName)
	if err != nil {
		cm.Logger.Error("Error creating container: %s", err)
		return nil, fmt.Errorf("error creating container: %w", err)
	}

	cm.Logger.Info("Created container with ID: %s", response.ID)

	err = cm.DockerClient.StartContainer(ctx, response.ID)
	if err != nil {
		cm.Logger.Error("Error starting container: %s", err)
		return nil, fmt.Errorf("error starting container: %w", err)
	}

	cm.Logger.Info("Container started successfully")

	return &database.ContainerInfo{
		ContainerID:   response.ID,
		ContainerName: config.ContainerName,
		ImageName:     config.ImageName,
		DomainName:    config.DomainName,
		HostPort:      hostPort,
		ContainerPort: config.ContainerPort,
		Status:        "running",
	}, nil
}

func (cm *ContainerManager) createAndStartNewContainer(ctx context.Context, config *ContainerConfig) (*database.ContainerInfo, error) {
	newHostPort, err := cm.portFinder.findAvailablePort()
	if err != nil {
		return nil, fmt.Errorf("error finding available port: %w", err)
	}

	newContainerName := fmt.Sprintf("%s_%s", config.ContainerName, time.Now().Format("20060102150405"))
	config.ContainerName = newContainerName

	return cm.createAndStartContainer(ctx, config, newHostPort)
}

func (cm *ContainerManager) stopAndRemoveContainer(ctx context.Context, containerID string) error {
	if err := cm.DockerClient.StopContainer(ctx, containerID, nil); err != nil {
		cm.Logger.Error("Error stopping container: %s", err)
		// Continue with removal even if stop fails
	}

	if err := cm.DockerClient.RemoveContainer(ctx, containerID, container.RemoveOptions{Force: true}); err != nil {
		cm.Logger.Error("Error removing container: %s", err)
		return fmt.Errorf("error removing container: %w", err)
	}

	return nil
}

func (cm *ContainerManager) updateDatabase(oldInfo *database.ContainerInfo, newInfo *database.ContainerInfo) error {
	if err := cm.Db.AddContainer(*newInfo); err != nil {
		return fmt.Errorf("error saving new container info to database: %w", err)
	}

	if err := cm.Db.DeleteContainer(oldInfo.ContainerID); err != nil {
		return fmt.Errorf("error removing old container info from database: %w", err)
	}

	return nil
}
