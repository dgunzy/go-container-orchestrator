package database

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDatabaseOperations(t *testing.T) {
	db, err := NewDatabase(":memory:") // Use in-memory SQLite database for testing
	require.NoError(t, err, "Error creating database")
	defer db.Close()

	err = db.InitSchema()
	require.NoError(t, err, "Error initializing database schema")

	t.Run("AddAndGetContainer", func(t *testing.T) {
		containerInfo := ContainerInfo{
			ContainerID:   "mock-id-1",
			ContainerName: "mock-container-1",
			ImageName:     "mock-image:latest",
			DomainName:    "mock.example.com",
			HostPort:      "8080",
			ContainerPort: "80",
			Status:        "running",
			CreatedAt:     time.Now(),
		}

		err := db.AddContainer(containerInfo)
		assert.NoError(t, err, "Error adding container to database")

		savedInfo, err := db.GetContainer(containerInfo.ContainerID)
		assert.NoError(t, err, "Error getting container from database")
		assert.Equal(t, containerInfo.ContainerName, savedInfo.ContainerName)
		assert.Equal(t, containerInfo.ImageName, savedInfo.ImageName)
		assert.Equal(t, containerInfo.Status, savedInfo.Status)

		// Test getting container by name
		savedInfoByName, err := db.GetContainerByName(containerInfo.ContainerName)
		assert.NoError(t, err, "Error getting container by name from database")
		assert.Equal(t, containerInfo.ContainerID, savedInfoByName.ContainerID)
	})

	t.Run("UpdateContainerStatus", func(t *testing.T) {
		containerInfo := ContainerInfo{
			ContainerID:   "mock-id-2",
			ContainerName: "mock-container-2",
			ImageName:     "mock-image:latest",
			DomainName:    "mock2.example.com",
			HostPort:      "8081",
			ContainerPort: "80",
			Status:        "running",
			CreatedAt:     time.Now(),
		}

		err := db.AddContainer(containerInfo)
		require.NoError(t, err, "Error adding container to database")

		newStatus := "stopped"
		err = db.UpdateContainerStatus(containerInfo.ContainerID, newStatus)
		assert.NoError(t, err, "Error updating container status")

		updatedInfo, err := db.GetContainer(containerInfo.ContainerID)
		assert.NoError(t, err, "Error getting updated container info")
		assert.Equal(t, newStatus, updatedInfo.Status)
	})

	t.Run("ListContainers", func(t *testing.T) {
		// Add a few containers
		for i := 3; i <= 5; i++ {
			containerInfo := ContainerInfo{
				ContainerID:   fmt.Sprintf("mock-id-%d", i),
				ContainerName: fmt.Sprintf("mock-container-%d", i),
				ImageName:     "mock-image:latest",
				DomainName:    fmt.Sprintf("mock%d.example.com", i),
				HostPort:      fmt.Sprint(8080 + i),
				ContainerPort: "80",
				Status:        "running",
				CreatedAt:     time.Now(),
			}

			err := db.AddContainer(containerInfo)
			require.NoError(t, err, "Error adding container to database")
		}

		containers, err := db.ListContainers()
		assert.NoError(t, err, "Error listing containers")
		assert.Len(t, containers, 5, "Expected 5 containers in the list")
	})

	t.Run("DeleteContainer", func(t *testing.T) {
		containerInfo := ContainerInfo{
			ContainerID:   "mock-id-6",
			ContainerName: "mock-container-6",
			ImageName:     "mock-image:latest",
			DomainName:    "mock6.example.com",
			HostPort:      "8086",
			ContainerPort: "80",
			Status:        "running",
			CreatedAt:     time.Now(),
		}

		err := db.AddContainer(containerInfo)
		require.NoError(t, err, "Error adding container to database")

		err = db.DeleteContainer(containerInfo.ContainerID)
		assert.NoError(t, err, "Error deleting container from database")

		_, err = db.GetContainer(containerInfo.ContainerID)
		assert.Error(t, err, "Expected error when getting deleted container")
	})

	t.Run("AddInvalidContainer", func(t *testing.T) {
		invalidContainer := ContainerInfo{
			ContainerID: "", // Invalid: empty ContainerID
			ImageName:   "mock-image:latest",
			Status:      "running",
		}

		err := db.AddContainer(invalidContainer)
		assert.Error(t, err, "Expected error when adding invalid container")
	})

	t.Run("UpdateNonExistentContainer", func(t *testing.T) {
		err := db.UpdateContainerStatus("non-existent-id", "stopped")
		assert.Error(t, err, "Expected error when updating non-existent container")
	})

	t.Run("GetNonExistentContainer", func(t *testing.T) {
		_, err := db.GetContainer("non-existent-id")
		assert.Error(t, err, "Expected error when getting non-existent container")
	})
}
