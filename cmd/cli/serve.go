package cli

import (
	"context"

	"github.com/spf13/cobra"
)

func (cli *CLI) newServeCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "serve",
		Short: "Run the container manager in daemon mode",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Implement continuous running logic here
			ctx := context.Background()
			return cli.cm.RunAsDaemon(ctx)
		},
	}
}
