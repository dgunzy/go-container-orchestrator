package cli

import (
	"context"
	"fmt"
	"os"

	pb "github.com/dgunzy/go-container-orchestrator/pkg/proto"
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
		fmt.Println("Container name is required")
		fmt.Println("Usage: remove <container-name> [--full]")
		return
	}
	containerName := args[0]
	fullRemove, _ := cmd.Flags().GetBool("full")

	resp, err := cli.client.client.RemoveContainer(context.Background(), &pb.RemoveContainerRequest{
		ContainerName: containerName,
		RemoveImage:   fullRemove,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error removing container: %v\n", err)
		return
	}
	if resp.Success {
		fmt.Printf("Successfully removed container '%s'", containerName)
		if fullRemove {
			fmt.Print(" and its image")
		}
		fmt.Println(".")
	} else {
		fmt.Println("Failed to remove the container.")
	}
}
