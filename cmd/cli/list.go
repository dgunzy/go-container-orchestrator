package cli

import (
	"os"

	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
)

var listCmd = NewCommand(
	"list",
	"List all containers and their status",
	runList,
)

func init() {
	rootCmd.AddCommand(listCmd.Command)
}

func runList(cmd *Command, args []string) {
	containers, err := cmd.CM.ListContainers()
	if err != nil {
		cmd.CM.Logger.Error("Error listing containers: %v", err)
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
