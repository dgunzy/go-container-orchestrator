package cli

import (
	"github.com/spf13/cobra"
)

type CLI struct {
	rootCmd *cobra.Command
	client  *Client
}

func NewCLI(address string) (*CLI, error) {
	c, err := NewClient(address)
	if err != nil {
		return nil, err
	}
	cli := &CLI{
		rootCmd: &cobra.Command{
			Use:   "container-orchestrator",
			Short: "A container orchestrator CLI",
		},
		client: c,
	}
	cli.initCommands()
	return cli, nil
}

func (cli *CLI) Run() error {
	return cli.rootCmd.Execute()
}

func (cli *CLI) Close() error {
	return cli.client.Close()
}

func (cli *CLI) initCommands() {
	cli.rootCmd.AddCommand(
		cli.newCreateCommand(),
		cli.newListCommand(),
		cli.newRemoveCommand(),
		cli.newUpdateCommand(),
		// cli.newServeCommand(),
	)
}
