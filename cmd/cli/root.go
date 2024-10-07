package cli

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/dgunzy/go-container-orchestrator/internal/container"
	"github.com/spf13/cobra"
)

type CLI struct {
	rootCmd *cobra.Command
	cm      *container.ContainerManager
}

func NewCLI(cm *container.ContainerManager) (*CLI, error) {

	cli := &CLI{
		rootCmd: &cobra.Command{
			Use:   "container-orchestrator",
			Short: "A container orchestrator CLI",
		},
		cm: cm,
	}

	cli.initCommands()
	return cli, nil
}
func (cli *CLI) RunInteractive(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			fmt.Print("container-orchestrator> ")
			var input string
			_, err := fmt.Scanln(&input)
			if err != nil {
				if err.Error() == "EOF" {
					return nil
				}
				fmt.Fprintln(os.Stderr, "Error reading input:", err)
				continue
			}

			args := strings.Fields(input)
			if len(args) == 0 {
				continue
			}

			cli.rootCmd.SetArgs(args)
			if err := cli.rootCmd.Execute(); err != nil {
				fmt.Fprintln(os.Stderr, "Error executing command:", err)
			}
		}
	}
}
func (cli *CLI) ExecuteWithArgs(args []string) error {
	cli.rootCmd.SetArgs(args)
	return cli.rootCmd.Execute()
}

func (cli *CLI) GetContainerManager() *container.ContainerManager {
	return cli.cm
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
