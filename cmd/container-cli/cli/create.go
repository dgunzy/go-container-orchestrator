package cli

import (
	"context"
	"fmt"
	"os"

	pb "github.com/dgunzy/go-container-orchestrator/pkg/proto"
	"github.com/spf13/cobra"
)

func (cli *CLI) newCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new container",
		Run:   cli.runCreate,
	}

	cmd.Flags().String("domain", "", "Domain name for the container")
	cmd.Flags().String("image", "", "Image name for the container")
	cmd.Flags().String("name", "", "Name for the container")
	cmd.Flags().String("port", "", "Container port")
	cmd.Flags().String("username", "", "Registry username")
	cmd.Flags().String("password", "", "Registry password")

	return cmd
}

func (cli *CLI) runCreate(cmd *cobra.Command, args []string) {
	config := &pb.ContainerConfig{
		DomainName:       cmd.Flag("domain").Value.String(),
		ImageName:        cmd.Flag("image").Value.String(),
		ContainerName:    cmd.Flag("name").Value.String(),
		ContainerPort:    cmd.Flag("port").Value.String(),
		RegistryUsername: cmd.Flag("username").Value.String(),
		RegistryPassword: cmd.Flag("password").Value.String(),
	}
	resp, err := cli.client.client.CreateContainer(context.Background(), &pb.CreateContainerRequest{Config: config})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating container: %v\n", err)
		return
	}
	fmt.Printf("Container created successfully with ID: %s\n", resp.ContainerId)
}
