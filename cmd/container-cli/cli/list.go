package cli

import (
	"context"
	"fmt"
	"os"

	pb "github.com/dgunzy/go-container-orchestrator/pkg/proto"
	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

func (cli *CLI) newListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all containers and their status",
		Run:   cli.runList,
	}
}

func (cli *CLI) runList(cmd *cobra.Command, args []string) {
	resp, err := cli.client.client.ListContainers(context.Background(), &pb.ListContainersRequest{})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error listing containers: %v\n", err)
		return
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Name", "ID", "Image", "Domain", "Port", "Status"})
	// ... (rest of the table setup)

	for _, container := range resp.Containers {
		status := container.Status
		if status == "running" {
			status = color.GreenString(status)
		} else {
			status = color.RedString(status)
		}
		table.Append([]string{
			container.ContainerName,
			container.ContainerId[:12],
			container.ImageName,
			container.DomainName,
			container.ContainerPort + ":" + container.HostPort,
			status,
		})
	}
	table.Render()
}
