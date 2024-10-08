package main

import (
	"fmt"
	"os"

	cli "github.com/dgunzy/go-container-orchestrator/cmd/container-cli/cli"
)

func main() {
	// You might want to make this configurable
	serverAddress := "localhost:50051"

	c, err := cli.NewCLI(serverAddress)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create CLI: %v\n", err)
		os.Exit(1)
	}
	defer c.Close()

	if err := c.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
