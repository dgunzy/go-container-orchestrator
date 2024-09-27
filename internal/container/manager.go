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

func NewContainerManager(dockerClient *docker.DockerClient) *ContainerManager {
	logger := logging.GetLogger()
	database, err := database.NewDatabase("../../db.sqlite")
	if err != nil {
		logger.Errorf("Error creating database: %s", err)
	}
	return &ContainerManager{
		dockerClient: dockerClient,
		portFinder:   newPortFinder(),
		db:           database,
		logger:       logger,
	}
}

func (cm *ContainerManager) CreateNewContainer(ctx context.Context, config *ContainerConfig) error {
	cm.logger.Infof("Creating new container: %s", config.ContainerName)
	err := cm.dockerClient.PullImageFromPrivateRegistry(ctx,
		config.ImageName,
		config.RegistryUsername,
		config.RegistryPassword)

	if err != nil {
		cm.logger.Errorf("Error pulling image: %s", err)
		return fmt.Errorf("error pulling image: %s", err)
	}
	newHostPort, err := cm.portFinder.findAvailablePort()
	if err != nil {
		cm.logger.Errorf("Error finding available port: %s", err)
		return fmt.Errorf("error finding available port: %s", err)
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
		return fmt.Errorf("error creating container: %s", err)
	}

	cm.logger.Infof("Created container with ID: %s", response.ID)

	err = cm.dockerClient.StartContainer(ctx, response.ID)
	if err != nil {
		cm.logger.Errorf("Error starting container: %s", err)
		return fmt.Errorf("error starting container: %s", err)
	}

	cm.logger.Info("Container started successfully")
	//TODO add nginx config to expose this internal port to proxy
	return nil
}

func (cm *ContainerManager) UpdateExistingContainer(ctx context.Context, config *ContainerConfig) error {
	cm.logger.Infof("Updating existing container: %s", config.ContainerName)
	// Generate a timestamp for the new container name
	timestamp := time.Now().Format("20060102150405")
	newContainerName := fmt.Sprintf("%s_%s", config.ContainerName, timestamp)

	// Step 1: Pull the new image
	err := cm.dockerClient.PullImageFromPrivateRegistry(ctx,
		config.ImageName,
		config.RegistryUsername,
		config.RegistryPassword)
	if err != nil {
		cm.logger.Errorf("Error pulling updated image: %s", err)
		return fmt.Errorf("error pulling updated image: %s", err)
	}

	// Step 2: Find an available host port
	newHostPort, err := cm.portFinder.findAvailablePort()
	if err != nil {
		cm.logger.Errorf("Error finding available port: %s", err)
		return fmt.Errorf("error finding available port: %s", err)
	}

	// Step 3: Create a new container with the updated image and new port
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
		return fmt.Errorf("error creating new container: %s", err)
	}

	// Step 4: Start the new container
	err = cm.dockerClient.StartContainer(ctx, response.ID)
	if err != nil {
		cm.logger.Errorf("Error starting new container: %s", err)
		return fmt.Errorf("error starting new container: %s", err)
	}
	cm.logger.Infof("New container started: %s", newContainerName)

	// make sure new container has started here
	time.Sleep(2 * time.Second)

	// Step 5: Verify the new container is healthy (implement health check logic)
	// TODO: Implement health check
	cm.logger.Info("TODO: Implement health check")

	// Step 6: Update Nginx configuration to point to the new container
	// TODO: Update Nginx config to use newHostPort
	cm.logger.Info("TODO: Update Nginx configuration")

	err = cm.dockerClient.StopContainer(ctx, config.ContainerName, nil)
	if err != nil {
		cm.logger.Errorf("Error stopping old container: %s", err)
	}

	err = cm.dockerClient.RemoveContainer(ctx, config.ContainerName, container.RemoveOptions{Force: true})
	if err != nil {
		cm.logger.Errorf("Error removing old container: %s", err)
	}

	// Step 9: Update the database with the new container information
	// TODO: Update database with newContainerName and newHostPort
	cm.logger.Info("TODO: Update database with new container information")

	cm.logger.Infof("Container update completed: %s", newContainerName)
	return nil
}

// TODO add function to remove container from db and

// Todo add function to load the containers and ports from db and start them
