package docker

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/docker/docker/api/types/image"

	"github.com/docker/docker/client"
)

type DockerClient struct {
	client *client.Client
}

type PullOptions struct {
	RegistryAuth string
}

type AuthConfig struct {
	Username string
	Password string
}

func NewClient() (*DockerClient, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return nil, err
	}
	return &DockerClient{client: cli}, nil
}

// PullImageFromPrivateRegistry pulls an image from a private registry but not if the exact image exits locally
func (d *DockerClient) PullImageFromPrivateRegistry(ctx context.Context, fullImageName, username, password string) error {

	exists, err := d.ImageExists(ctx, fullImageName)
	if err != nil {
		return fmt.Errorf("error checking if image exists: %s", err)
	}

	if exists {
		return errors.New("image already exists with same tag, not pulling again")
	}

	auth := AuthConfig{
		Username: username,
		Password: password,
	}
	encodedJSON, err := json.Marshal(auth)
	if err != nil {
		return fmt.Errorf("error encoding auth config")
	}
	authStr := base64.URLEncoding.EncodeToString(encodedJSON)

	out, err := d.client.ImagePull(ctx, fullImageName, image.PullOptions{RegistryAuth: authStr})

	if err != nil {
		return fmt.Errorf("error pulling image: %s", err)
	}

	defer out.Close()
	_, err = io.Copy(os.Stdout, out)
	if err != nil {
		return fmt.Errorf("failed to write pull output: %v", err)
	}

	return nil
}

func (d *DockerClient) ImageExists(ctx context.Context, imageName string) (bool, error) {
	_, _, err := d.client.ImageInspectWithRaw(ctx, imageName)
	if err != nil {
		if client.IsErrNotFound(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
