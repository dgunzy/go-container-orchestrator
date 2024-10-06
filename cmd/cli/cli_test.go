package cli_test

import (
	"bytes"
	"context"
	"os"
	"testing"
	"time"

	"github.com/dgunzy/go-container-orchestrator/cmd/cli"
	"github.com/dgunzy/go-container-orchestrator/internal/container"
	"github.com/dgunzy/go-container-orchestrator/internal/logging"
	"github.com/dgunzy/go-container-orchestrator/tests"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var cm *container.ContainerManager
var rootCmd *cli.Command

func TestMain(m *testing.M) {
	cm := tests.InitTestConfig()
	defer tests.CleanupTestResources(cm.DockerClient)
	// Initialize CLI root command
	rootCmd = cli.NewCommand(cm)

	// Run tests
	exitCode := m.Run()

	// Teardown code

	if err := logging.CloseGlobalLogger(); err != nil {
		panic(err)
	}

	os.Exit(exitCode)
}

func TestCLICommands(t *testing.T) {
	// Helper function to execute CLI commands
	executeCommand := func(args ...string) (string, error) {
		buf := new(bytes.Buffer)
		rootCmd.SetOut(buf)
		rootCmd.SetErr(buf)
		rootCmd.SetArgs(args)
		err := rootCmd.Execute()
		return buf.String(), err
	}

	// Test cases
	t.Run("Create Container", func(t *testing.T) {
		output, err := executeCommand("create", "--name", "test-container", "--image", "alpine:latest", "--domain", "test.com", "--port", "8080")
		assert.NoError(t, err)
		assert.Contains(t, output, "Container created successfully")

		// Verify container was actually created
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		containers, err := cm.DockerClient.ListContainers(ctx)
		require.NoError(t, err)
		found := false
		for _, c := range containers {
			if c.Names[0] == "/test-container" {
				found = true
				break
			}
		}
		assert.True(t, found, "Created container not found in Docker")
	})

	t.Run("List Containers", func(t *testing.T) {
		output, err := executeCommand("list")
		assert.NoError(t, err)
		assert.Contains(t, output, "test-container")
		assert.Contains(t, output, "alpine:latest")
	})

	t.Run("Update Container", func(t *testing.T) {
		output, err := executeCommand("update", "--name", "test-container", "--image", "alpine:3.14")
		assert.NoError(t, err)
		assert.Contains(t, output, "Container updated successfully")

		// Verify container was actually updated
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		containers, err := cm.DockerClient.ListContainers(ctx)
		require.NoError(t, err)
		found := false
		for _, c := range containers {
			if c.Names[0] == "/test-container" && c.Image == "alpine:3.14" {
				found = true
				break
			}
		}
		assert.True(t, found, "Updated container not found in Docker")
	})

	t.Run("Remove Container", func(t *testing.T) {
		output, err := executeCommand("remove", "test-container")
		assert.NoError(t, err)
		assert.Contains(t, output, "Successfully removed container")

		// Verify container was actually removed
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		containers, err := cm.DockerClient.ListContainers(ctx)
		require.NoError(t, err)
		found := false
		for _, c := range containers {
			if c.Names[0] == "/test-container" {
				found = true
				break
			}
		}
		assert.False(t, found, "Removed container still found in Docker")
	})
}
