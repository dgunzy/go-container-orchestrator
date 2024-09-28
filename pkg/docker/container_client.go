package docker

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
)

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

func (d *DockerClient) RemoveImage(ctx context.Context, imageName string) error {
	_, err := d.client.ImageRemove(ctx, imageName, image.RemoveOptions{})

	return err
}

func (d *DockerClient) StartContainer(ctx context.Context, containerID string) error {
	return d.client.ContainerStart(ctx, containerID, container.StartOptions{})
}

func (d *DockerClient) StopContainer(ctx context.Context, containerID string, timeout *int) error {
	return d.client.ContainerStop(ctx, containerID, container.StopOptions{Timeout: timeout})
}

func (d *DockerClient) RestartContainer(ctx context.Context, containerID string, timeout *int) error {
	return d.client.ContainerRestart(ctx, containerID, container.StopOptions{Timeout: timeout})
}

func (d *DockerClient) GetContainerLogs(ctx context.Context, containerID string) (io.ReadCloser, error) {
	return d.client.ContainerLogs(ctx, containerID, container.LogsOptions{ShowStdout: true, ShowStderr: true})
}

func (d *DockerClient) ExecuteContainerCommand(ctx context.Context, containerID string, command []string) (types.IDResponse, error) {
	return d.client.ContainerExecCreate(ctx, containerID, container.ExecOptions{
		Cmd:          command,
		AttachStdout: true,
		AttachStderr: true,
	})
}

func (d *DockerClient) HealthCheck(ctx context.Context, containerID string) (types.ContainerState, error) {
	if d == nil || d.client == nil {
		return types.ContainerState{}, errors.New("DockerClient or its client is nil")
	}

	inspect, err := d.client.ContainerInspect(ctx, containerID)
	if err != nil {
		return types.ContainerState{}, fmt.Errorf("failed to inspect container: %w", err)
	}

	if inspect.State == nil {
		return types.ContainerState{}, errors.New("container state is nil")
	}

	return *inspect.State, nil
}
