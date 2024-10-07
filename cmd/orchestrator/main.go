package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dgunzy/go-container-orchestrator/cmd/cli"
	"github.com/dgunzy/go-container-orchestrator/internal/logging"
)

func main() {
	logPath := os.Getenv("LOG_PATH")
	if logPath == "" {
		logPath = "./container_manager_logs"
	}
	// Set up logging
	if err := logging.Setup(logPath); err != nil {
		fmt.Printf("Failed to set up logging: %v\n", err)
		os.Exit(1)
	}
	logger := logging.GetLogger()

	// Initialize the CLI
	c, err := cli.NewCLI()
	if err != nil {
		logger.Error("Failed to initialize CLI: %v", err)
		os.Exit(1)
	}

	// Check if we're running in daemon mode
	if len(os.Args) > 1 && os.Args[1] == "serve" {
		runContainerManager(c)
	} else {
		// Run CLI commands
		if err := c.Run(); err != nil {
			logger.Error("CLI command failed: %v", err)
			os.Exit(1)
		}
	}
}

func runContainerManager(c *cli.CLI) {
	logger := logging.GetLogger()
	logger.Info("Starting container manager in daemon mode")

	// Create a context that we can cancel
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Set up signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Run the container manager in a goroutine
	errChan := make(chan error, 1)
	go func() {
		errChan <- c.GetContainerManager().RunAsDaemon(ctx)
	}()

	// Wait for termination signal or error
	select {
	case sig := <-sigChan:
		logger.Info("Received termination signal: %v. Shutting down...", sig)
	case err := <-errChan:
		if err != nil {
			logger.Error("Container manager failed: %v", err)
		} else {
			logger.Info("Container manager stopped")
		}
	}

	// Cancel the context to stop the container manager
	cancel()

	// Allow some time for graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	select {
	case <-errChan:
		logger.Info("Container manager stopped gracefully")
	case <-shutdownCtx.Done():
		logger.Warn("Graceful shutdown timed out")
	}
}
