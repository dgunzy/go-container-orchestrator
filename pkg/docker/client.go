package docker

import (
	"context"

	"github.com/docker/docker/api/types"

	"github.com/docker/docker/api/types/container"

	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
)

type DockerClient struct {
	client *client.Client
}

func NewClient() (*DockerClient, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return nil, err
	}
	return &DockerClient{client: cli}, nil
}

func (d *DockerClient) ListContainers(ctx context.Context) ([]types.Container, error) {
	return d.client.ContainerList(ctx, container.ListOptions{})
}

func (d *DockerClient) CreateContainer(ctx context.Context,
	config *container.Config,
	hostConfig *container.HostConfig,
	networkingConfig *network.NetworkingConfig,
	containerName string) (container.CreateResponse, error) {
	return d.client.ContainerCreate(ctx, config, hostConfig, networkingConfig, nil, containerName)
}

func (d *DockerClient) RemoveContainer(ctx context.Context, containerID string, options container.RemoveOptions) error {
	return d.client.ContainerRemove(ctx, containerID, options)
}
