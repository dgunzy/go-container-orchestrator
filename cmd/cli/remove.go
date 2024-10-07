package cli

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
)

func (cli *CLI) newRemoveCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove <container-name>",
		Short: "Remove a container or container and its image",
		Run:   cli.runRemove,
	}

	cmd.Flags().Bool("full", false, "Remove both the container and its image")

	return cmd
}

func (cli *CLI) runRemove(cmd *cobra.Command, args []string) {
	if len(args) < 1 {
		cli.cm.Logger.Error("Container name is required")
		fmt.Println("Usage: remove <container-name> [--full]")
		return
	}

	containerName := args[0]
	fullRemove, _ := cmd.Flags().GetBool("full")

	containers, err := cli.cm.Db.GetContainersByPartialName(containerName)
	if err != nil {
		cli.cm.Logger.Error("Error fetching container: %v", err)
		return
	}

	if len(containers) == 0 {
		cli.cm.Logger.Error("No containers found matching '%s'", containerName)
		return
	}

	if len(containers) > 1 {
		fmt.Printf("Multiple containers found matching '%s':\n", containerName)
		for _, c := range containers {
			cli.cm.Logger.Info("- %s (ID: %s)\n", c.ContainerName, c.ContainerID)
		}
		cli.cm.Logger.Info("Please specify the exact container name.")
		return
	}

	container := containers[0]

	if fullRemove {
		err = cli.cm.RemoveContainerAndImage(context.Background(), container.ContainerID)
		if err != nil {
			cli.cm.Logger.Error("Error removing container and image: %v", err)
			return
		}
		cli.cm.Logger.Info("Successfully removed container '%s' and its image.\n", container.ContainerName)
	} else {
		err = cli.cm.RemoveContainer(context.Background(), container.ContainerID)
		if err != nil {
			cli.cm.Logger.Error("Error removing container: %v", err)
			return
		}
		cli.cm.Logger.Info("Successfully removed container '%s'.\n", container.ContainerName)
	}
}
