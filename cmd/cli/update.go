package cli

import (
	"context"

	"github.com/dgunzy/go-container-orchestrator/internal/container"
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
	config := &container.ContainerConfig{
		DomainName:       cmd.Flag("domain").Value.String(),
		ImageName:        cmd.Flag("image").Value.String(),
		ContainerName:    cmd.Flag("name").Value.String(),
		ContainerPort:    cmd.Flag("port").Value.String(),
		RegistryUsername: cmd.Flag("username").Value.String(),
		RegistryPassword: cmd.Flag("password").Value.String(),
	}

	err := cli.cm.UpdateExistingContainer(context.Background(), config)
	if err != nil {
		cli.cm.Logger.Error("Error updating container: %v", err)
		return
	}
	cli.cm.Logger.Info("Container updated successfully")
}
