package cli

import (
	"github.com/dgunzy/go-container-orchestrator/internal/container"
	"github.com/spf13/cobra"
)

type CLI struct {
	rootCmd *cobra.Command
	cm      *container.ContainerManager
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
func (cli *CLI) ExecuteWithArgs(args []string) error {
	cli.rootCmd.SetArgs(args)
	return cli.rootCmd.Execute()
}

// Expose ContainerManager for testing
func (cli *CLI) GetContainerManager() *container.ContainerManager {
	return cli.cm
}
