package cli

import (
	"os"

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
	containers, err := cli.cm.ListContainers()
	if err != nil {
		cli.cm.Logger.Error("Error listing containers: %v", err)
		return
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Name", "ID", "Image", "Domain", "Port", "Status"})
	table.SetBorder(false)
	table.SetAutoWrapText(false)
	table.SetAutoFormatHeaders(true)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetCenterSeparator("")
	table.SetColumnSeparator("")
	table.SetRowSeparator("")
	table.SetHeaderLine(false)
	table.SetTablePadding("\t")
	table.SetNoWhiteSpace(true)

	for _, container := range containers {
		status := container.Status
		if status == "running" {
			status = color.GreenString(status)
		} else {
			status = color.RedString(status)
		}

		table.Append([]string{
			container.ContainerName,
			container.ContainerID[:12],
			container.ImageName,
			container.DomainName,
			container.ContainerPort + ":" + container.HostPort,
			status,
		})
	}

	table.Render()
}
