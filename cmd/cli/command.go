package cli

import (
	"github.com/dgunzy/go-container-orchestrator/internal/container"
	"github.com/spf13/cobra"
)

type CLI struct {
	rootCmd *cobra.Command
	cm      *container.ContainerManager
}

func NewCLI(cm *container.ContainerManager) *CLI {
	cli := &CLI{
		rootCmd: &cobra.Command{
			Use:   "container-orchestrator",
			Short: "A container orchestrator CLI",
		},
		cm: cm,
	}
	cli.initCommands()
	return cli
}

func (cli *CLI) Run() error {
	return cli.rootCmd.Execute()
}

func (cli *CLI) initCommands() {
	cli.rootCmd.AddCommand(
		cli.newCreateCommand(),
		cli.newListCommand(),
		cli.newRemoveCommand(),
		cli.newUpdateCommand(),
		cli.newServeCommand(),
	)
}
