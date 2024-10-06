package database_test

import (
	"fmt"
	"testing"

	"github.com/dgunzy/go-container-orchestrator/internal/database"
	"github.com/dgunzy/go-container-orchestrator/tests"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDatabaseOperations(t *testing.T) {
	cm := tests.InitTestConfig()
	defer tests.CleanupTestResources(cm.DockerClient)
	db := *cm.Db

	t.Run("AddAndGetContainer", func(t *testing.T) {
		containerInfo := database.ContainerInfo{
			ContainerID:   "mock-id-1",
			ContainerName: "mock-container-1",
			ImageName:     "mock-image:latest",
			DomainName:    "mock.example.com",
			HostPort:      "8080",
			ContainerPort: "80",
			Status:        "running",
		}

		err := db.AddContainer(containerInfo)
		assert.NoError(t, err, "Error adding container to database")

		savedInfo, err := db.GetContainer(containerInfo.ContainerID)
		assert.NoError(t, err, "Error getting container from database")
		assert.Equal(t, containerInfo.ContainerName, savedInfo.ContainerName)
		assert.Equal(t, containerInfo.ImageName, savedInfo.ImageName)
		assert.Equal(t, containerInfo.Status, savedInfo.Status)

		// Test getting container by name
		savedInfoByName, err := db.GetContainersByPartialName(containerInfo.ContainerName)
		assert.NoError(t, err, "Error getting container by name from database")
		assert.Equal(t, containerInfo.ContainerID, savedInfoByName[0].ContainerID)
	})

	t.Run("UpdateContainerStatus", func(t *testing.T) {
		containerInfo := database.ContainerInfo{
			ContainerID:   "mock-id-2",
			ContainerName: "mock-container-2",
			ImageName:     "mock-image:latest",
			DomainName:    "mock2.example.com",
			HostPort:      "8081",
			ContainerPort: "80",
			Status:        "running",
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
			containerInfo := database.ContainerInfo{
				ContainerID:   fmt.Sprintf("mock-id-%d", i),
				ContainerName: fmt.Sprintf("mock-container-%d", i),
				ImageName:     "mock-image:latest",
				DomainName:    fmt.Sprintf("mock%d.example.com", i),
				HostPort:      fmt.Sprint(8080 + i),
				ContainerPort: "80",
				Status:        "running",
			}

			err := db.AddContainer(containerInfo)
			require.NoError(t, err, "Error adding container to database")
		}

		containers, err := db.ListContainers()
		assert.NoError(t, err, "Error listing containers")
		assert.Len(t, containers, 5, "Expected 5 containers in the list")
	})

	t.Run("DeleteContainer", func(t *testing.T) {
		containerInfo := database.ContainerInfo{
			ContainerID:   "mock-id-6",
			ContainerName: "mock-container-6",
			ImageName:     "mock-image:latest",
			DomainName:    "mock6.example.com",
			HostPort:      "8086",
			ContainerPort: "80",
			Status:        "running",
		}

		err := db.AddContainer(containerInfo)
		require.NoError(t, err, "Error adding container to database")

		err = db.DeleteContainer(containerInfo.ContainerID)
		assert.NoError(t, err, "Error deleting container from database")

		_, err = db.GetContainer(containerInfo.ContainerID)
		assert.Error(t, err, "Expected error when getting deleted container")
	})

	t.Run("AddInvalidContainer", func(t *testing.T) {
		invalidContainer := database.ContainerInfo{
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
	t.Run("GetContainersByPartialName", func(t *testing.T) {
		// Add containers with different names for testing partial search
		testContainers := []database.ContainerInfo{
			{ContainerID: "id-web-1", ContainerName: "web-server-1", ImageName: "nginx:latest", DomainName: "web1.example.com", HostPort: "8091", ContainerPort: "80", Status: "running"},
			{ContainerID: "id-web-2", ContainerName: "web-server-2", ImageName: "nginx:latest", DomainName: "web2.example.com", HostPort: "8092", ContainerPort: "80", Status: "running"},
			{ContainerID: "id-db", ContainerName: "db-server", ImageName: "postgres:latest", DomainName: "db.example.com", HostPort: "5432", ContainerPort: "5432", Status: "running"},
			{ContainerID: "id-cache", ContainerName: "cache-server", ImageName: "redis:latest", DomainName: "cache.example.com", HostPort: "6379", ContainerPort: "6379", Status: "running"},
		}

		for _, c := range testContainers {
			err := db.AddContainer(c)
			require.NoError(t, err, "Error adding test container to database")
		}

		// Test cases for partial name search
		testCases := []struct {
			partialName     string
			expectedCount   int
			expectedError   bool
			expectedNameIds []string
		}{
			{"web", 2, false, []string{"id-web-1", "id-web-2"}},
			{"server", 4, false, []string{"id-web-1", "id-web-2", "id-db", "id-cache"}},
			{"db", 1, false, []string{"id-db"}},
			{"cache", 1, false, []string{"id-cache"}},
			{"nonexistent", 0, true, nil},
		}

		for _, tc := range testCases {
			t.Run(fmt.Sprintf("PartialName_%s", tc.partialName), func(t *testing.T) {
				results, err := db.GetContainersByPartialName(tc.partialName)

				if tc.expectedError {
					assert.Error(t, err, "Expected error for non-existent partial name")
				} else {
					assert.NoError(t, err, "Unexpected error in partial name search")
				}

				assert.Len(t, results, tc.expectedCount, "Unexpected number of results")

				if tc.expectedNameIds != nil {
					resultIDs := make([]string, len(results))
					for i, r := range results {
						resultIDs[i] = r.ContainerID
					}
					assert.ElementsMatch(t, tc.expectedNameIds, resultIDs, "Mismatched container IDs in results")
				}
			})
		}
	})
}
