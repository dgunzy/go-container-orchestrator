package cli_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/dgunzy/go-container-orchestrator/cmd/cli"
	"github.com/dgunzy/go-container-orchestrator/tests"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCLIIntegration(t *testing.T) {
	// Set up test environment
	os.Setenv("DB_PATH", ":memory:")
	os.Setenv("LOG_PATH", "../../test_logs")

	// Initialize CLI
	c, err := cli.NewCLI()
	require.NoError(t, err, "Failed to initialize CLI")

	defer tests.CleanupTestResources(c.GetContainerManager().DockerClient)

	// Start the container manager process
	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		err := c.ExecuteWithArgs([]string{"serve"})
		assert.NoError(t, err, "Failed to run serve command")
	}()

	// Wait for the container manager to start
	time.Sleep(5 * time.Second)

	// Run tests while the container manager is running
	t.Run("CLIOperations", func(t *testing.T) {
		// Test create command
		t.Run("CreateContainer", func(t *testing.T) {
			err := c.ExecuteWithArgs([]string{"create", "--domain", "test.com", "--image", "alpine:latest", "--name", "test-container", "--port", "80"})
			assert.NoError(t, err, "Failed to create container")
		})

		// Allow some time for the container to be created
		time.Sleep(2 * time.Second)

		// Test list command
		t.Run("ListContainers", func(t *testing.T) {
			err := c.ExecuteWithArgs([]string{"list"})
			assert.NoError(t, err, "Failed to list containers")
			// TODO: Add assertion to check if the created container is in the list
		})

		// Test update command
		t.Run("UpdateContainer", func(t *testing.T) {
			err := c.ExecuteWithArgs([]string{"update", "--name", "test-container", "--image", "alpine:3.12"})
			assert.NoError(t, err, "Failed to update container")
		})

		// Allow some time for the container to be updated
		time.Sleep(2 * time.Second)

		// Test list again to verify update
		t.Run("ListContainersAfterUpdate", func(t *testing.T) {
			err := c.ExecuteWithArgs([]string{"list"})
			assert.NoError(t, err, "Failed to list containers after update")
			// TODO: Add assertion to check if the container image has been updated
		})

		// Test remove command
		t.Run("RemoveContainer", func(t *testing.T) {
			err := c.ExecuteWithArgs([]string{"remove", "test-container"})
			assert.NoError(t, err, "Failed to remove container")
		})

		// Allow some time for the container to be removed
		time.Sleep(2 * time.Second)

		// Test list one last time to verify removal
		t.Run("ListContainersAfterRemoval", func(t *testing.T) {
			err := c.ExecuteWithArgs([]string{"list"})
			assert.NoError(t, err, "Failed to list containers after removal")
			// TODO: Add assertion to check if the container has been removed from the list
		})
	})

	// Stop the container manager
	cancel()

	// Wait for the container manager to stop
	time.Sleep(2 * time.Second)
}
