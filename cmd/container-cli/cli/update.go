package cli

import (
	"context"
	"fmt"
	"os"

	pb "github.com/dgunzy/go-container-orchestrator/pkg/proto"
	"github.com/spf13/cobra"
)

func (cli *CLI) newUpdateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update a container or image, only the container name and image is required, all other fields are optional",
		Run:   cli.runUpdate,
	}

	cmd.Flags().String("domain", "", "Domain name for the container")
	cmd.Flags().String("image", "", "Image name for the container")
	cmd.Flags().String("name", "", "Name for the container")
	cmd.Flags().String("port", "", "Container port")
	cmd.Flags().String("username", "", "Registry username")
	cmd.Flags().String("password", "", "Registry password")

	return cmd
}

func (cli *CLI) runUpdate(cmd *cobra.Command, args []string) {
	config := &pb.ContainerConfig{
		DomainName:       cmd.Flag("domain").Value.String(),
		ImageName:        cmd.Flag("image").Value.String(),
		ContainerName:    cmd.Flag("name").Value.String(),
		ContainerPort:    cmd.Flag("port").Value.String(),
		RegistryUsername: cmd.Flag("username").Value.String(),
		RegistryPassword: cmd.Flag("password").Value.String(),
	}
	resp, err := cli.client.client.UpdateContainer(context.Background(), &pb.UpdateContainerRequest{Config: config})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error updating container: %v\n", err)
		return
	}
	if resp.Success {
		fmt.Println("Container updated successfully")
	} else {
		fmt.Println("Failed to update the container")
	}
}
