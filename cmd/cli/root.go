package cli

import (
	"fmt"
	"os"

	"github.com/dgunzy/go-container-orchestrator/internal/container"
	"github.com/dgunzy/go-container-orchestrator/internal/logging"
	"github.com/dgunzy/go-container-orchestrator/pkg/docker"
	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
)

var rootCmd = NewCommand(
	"container-orchestrator",
	"A container orchestrator CLI",
	func(cmd *Command, args []string) {
		// Root command action
	},
)

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
}

func initConfig() {
	// Initialize logging
	if err := logging.Setup("../../container_manager_logs"); err != nil {
		fmt.Printf("Error setting up logging: %v\n", err)
		os.Exit(1)
	}

	logger := logging.GetLogger()

	// Initialize Docker client
	dockerClient, err := docker.NewClient()
	if err != nil {
		logger.Error("Error creating Docker client: %v", err)
		os.Exit(1)
	}

	if godotenv.Load() != nil {
		logger.Warn("Error loading .env file")
	}

	DBPath := os.Getenv("DB_PATH")
	if DBPath == "" {
		logger.Warn("No DB_PATH environment variable set, using in-memory database")
		DBPath = ":memory:"
	}
	// Initialize ContainerManager
	cm, err := container.NewContainerManager(dockerClient, DBPath, logger)
	if err != nil {
		logger.Error("Error creating ContainerManager: %v", err)
		os.Exit(1)
	}

	// Set the ContainerManager for all commands
	rootCmd.CM = cm
}
