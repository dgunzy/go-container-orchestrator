package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/dgunzy/go-container-orchestrator/cmd/cli"
	"github.com/dgunzy/go-container-orchestrator/internal/container"
	"github.com/dgunzy/go-container-orchestrator/internal/logging"
)

func main() {
	if err := logging.Setup("./container_manager_logs"); err != nil {
		fmt.Printf("Failed to set up logging: %v\n", err)
		os.Exit(1)
	}
	logger := logging.GetLogger()

	cm, err := container.NewContainerManager()
	if err != nil {
		logger.Error("Failed to initialize ContainerManager: %v", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Set up signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	var wg sync.WaitGroup

	// Start ContainerManager
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := cm.RunAsDaemon(ctx); err != nil {
			logger.Error("Container manager failed: %v", err)
		}
	}()

	// Start CLI
	wg.Add(1)
	go func() {
		defer wg.Done()
		c, err := cli.NewCLI(cm)
		if err != nil {
			logger.Error("Failed to initialize CLI: %v", err)
			return
		}
		if err := c.RunInteractive(ctx); err != nil {
			logger.Error("CLI failed: %v", err)
		}
	}()

	// Wait for termination signal
	<-sigChan
	logger.Info("Received termination signal. Shutting down...")
	cancel()

	// Wait for goroutines to finish
	wg.Wait()
	logger.Info("Shutdown complete")
}
