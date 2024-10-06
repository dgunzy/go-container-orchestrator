package cli

import (
	"context"

	"github.com/dgunzy/go-container-orchestrator/internal/container"
)

var updateCmd = NewCommand(
	"update",
	"Update a container or image, only the container name and image is required, all other fields are optional",
	runUpdate,
)

func init() {
	rootCmd.AddCommand(updateCmd.Command)
	createCmd.Flags().String("domain", "", "Domain name for the container")
	createCmd.Flags().String("image", "", "Image name for the container")
	createCmd.Flags().String("name", "", "Name for the container")
	createCmd.Flags().String("port", "", "Container port")
	createCmd.Flags().String("username", "", "Registry username")
	createCmd.Flags().String("password", "", "Registry password")
}

func runUpdate(cmd *Command, args []string) {
	//
	config := &container.ContainerConfig{
		DomainName:       cmd.Flags().Lookup("domain").Value.String(),
		ImageName:        cmd.Flags().Lookup("image").Value.String(),
		ContainerName:    cmd.Flags().Lookup("name").Value.String(),
		ContainerPort:    cmd.Flags().Lookup("port").Value.String(),
		RegistryUsername: cmd.Flags().Lookup("username").Value.String(),
		RegistryPassword: cmd.Flags().Lookup("password").Value.String(),
	}

	err := cmd.CM.UpdateExistingContainer(context.Background(), config)
	if err != nil {
		cmd.CM.Logger.Error("Error updating container: %v", err)
		return
	}

}
