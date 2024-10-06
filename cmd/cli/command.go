package cli

import (
	"github.com/dgunzy/go-container-orchestrator/internal/container"
	"github.com/spf13/cobra"
)

type Command struct {
	*cobra.Command
	CM *container.ContainerManager
}

func NewCommand(use string, short string, run func(cmd *Command, args []string)) *Command {
	c := &Command{
		Command: &cobra.Command{
			Use:   use,
			Short: short,
		},
	}
	c.Run = func(cmd *cobra.Command, args []string) {
		run(c, args)
	}
	return c
}
