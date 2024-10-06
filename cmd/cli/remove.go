package cli

import (
	"context"
	"fmt"
)

var removeCmd = NewCommand(
	"remove <container-name>",
	"Remove a container or container and its image",
	runRemove,
)

func init() {
	rootCmd.AddCommand(removeCmd.Command)
	removeCmd.Flags().Bool("full", false, "Remove both the container and its image")
}

func runRemove(cmd *Command, args []string) {
	if len(args) < 1 {
		cmd.CM.Logger.Error("Container name is required")
		fmt.Println("Usage: remove <container-name> [--full]")
		return
	}

	containerName := args[0]
	fullRemove, _ := cmd.Flags().GetBool("full")

	containers, err := cmd.CM.Db.GetContainersByPartialName(containerName)
	if err != nil {
		cmd.CM.Logger.Error("Error fetching container: %v", err)
		return
	}

	if len(containers) == 0 {
		cmd.CM.Logger.Error("No containers found matching '%s'", containerName)
		return
	}

	if len(containers) > 1 {
		fmt.Printf("Multiple containers found matching '%s':\n", containerName)
		for _, c := range containers {
			cmd.CM.Logger.Info("- %s (ID: %s)\n", c.ContainerName, c.ContainerID)
		}
		cmd.CM.Logger.Info("Please specify the exact container name.")
		return
	}

	container := containers[0]

	if fullRemove {
		err = cmd.CM.RemoveContainerAndImage(context.Background(), container.ContainerID)
		if err != nil {
			cmd.CM.Logger.Error("Error removing container and image: %v", err)
			return
		}
		cmd.CM.Logger.Info("Successfully removed container '%s' and its image.\n", container.ContainerName)
	} else {
		err = cmd.CM.RemoveContainer(context.Background(), container.ContainerID)
		if err != nil {
			cmd.CM.Logger.Error("Error removing container: %v", err)
			return
		}
		cmd.CM.Logger.Info("Successfully removed container '%s'.\n", container.ContainerName)
	}
}
