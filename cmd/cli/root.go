package cli

import (
	"fmt"
	"os"
	"time"

	"github.com/dgunzy/go-container-orchestrator/internal/container"
	"github.com/dgunzy/go-container-orchestrator/internal/database"
	"github.com/dgunzy/go-container-orchestrator/internal/health"
	"github.com/dgunzy/go-container-orchestrator/internal/logging"
	internal_client "github.com/dgunzy/go-container-orchestrator/pkg/docker"
	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
)

func NewCLI() (*CLI, error) {
	cm, err := initializeContainerManager()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize container manager: %w", err)
	}

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

func initializeContainerManager() (*container.ContainerManager, error) {
	if err := godotenv.Load(); err != nil {
		fmt.Println("Warning: Error loading .env file")
	}
	// Get the log path from the environment variable, or use the default value
	logPath := os.Getenv("LOG_PATH")
	if logPath == "" {
		logPath = "./container_manager_logs"
	}

	err := logging.Setup(logPath)
	if err != nil {
		return nil, fmt.Errorf("failed to set up logging: %w", err)
	}
	logger := logging.GetLogger()

	dockerClient, err := internal_client.NewClient()
	if err != nil {
		return nil, fmt.Errorf("error creating Docker client: %w", err)
	}
	// change this for testing!
	DBPath := os.Getenv("DB_PATH")
	if DBPath == "" {
		logger.Warn("No DB_PATH environment variable set, using in-memory database")
		DBPath = "./container_manager.db"
	}

	db, err := database.NewDatabase(DBPath)
	if err != nil {
		return nil, fmt.Errorf("error creating database: %w", err)
	}

	healthChecker := health.NewHealthChecker(dockerClient, db, 5*time.Minute, logger)

	cm, err := container.NewContainerManager(dockerClient, DBPath, logger, healthChecker)
	if err != nil {
		return nil, fmt.Errorf("error creating ContainerManager: %w", err)
	}

	if err := cm.Db.InitSchema(); err != nil {
		return nil, fmt.Errorf("error initializing database schema: %w", err)
	}

	return cm, nil
}

func Execute() {
	cli, err := NewCLI()
	if err != nil {
		fmt.Printf("Error initializing CLI: %v\n", err)
		os.Exit(1)
	}

	if err := cli.Run(); err != nil {
		fmt.Printf("Error executing CLI: %v\n", err)
		os.Exit(1)
	}
}
