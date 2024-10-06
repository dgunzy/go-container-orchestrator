package database

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/dgunzy/go-container-orchestrator/internal/logging"
	_ "github.com/mattn/go-sqlite3"
)

type ContainerInfo struct {
	ID            int
	ContainerID   string
	ContainerName string
	ImageName     string
	DomainName    string
	HostPort      string
	ContainerPort string
	Status        string
}

type Database struct {
	db     *sql.DB
	logger *logging.Logger
}

func NewDatabase(dbPath string) (*Database, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	return &Database{
		db:     db,
		logger: logging.GetLogger(),
	}, nil
}

func (d *Database) Close() error {
	d.logger.Info("Closing database connection")
	return d.db.Close()
}

func (d *Database) InitSchema() error {
	d.logger.Info("Initializing database schema")
	_, err := d.db.Exec(`
		CREATE TABLE IF NOT EXISTS containers (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			container_id TEXT NOT NULL,
			container_name TEXT NOT NULL,
			image_name TEXT NOT NULL,
			domain_name TEXT NOT NULL,
			host_port INTEGER NOT NULL,
			container_port INTEGER NOT NULL,
			status TEXT NOT NULL
		);
	`)
	if err != nil {
		return fmt.Errorf("failed to initialize schema: %w", err)
	}
	return nil
}

func (d *Database) AddContainer(info ContainerInfo) error {
	if err := validateContainerInfo(info); err != nil {
		return fmt.Errorf("invalid container info: %w", err)
	}

	d.logger.Info("Adding container: %s", info.ContainerName)
	_, err := d.db.Exec(`
		INSERT INTO containers (container_id, container_name, image_name, domain_name, host_port, container_port, status)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, info.ContainerID, info.ContainerName, info.ImageName, info.DomainName, info.HostPort, info.ContainerPort, info.Status)
	if err != nil {
		return fmt.Errorf("failed to add container: %w", err)
	}
	return nil
}

func (d *Database) UpdateContainerStatus(containerID, status string) error {
	d.logger.Info("Updating status for container %s to %s", containerID, status)
	result, err := d.db.Exec("UPDATE containers SET status = ? WHERE container_id = ?", status, containerID)
	if err != nil {
		return fmt.Errorf("failed to update container status: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return errors.New("no container found with the given ID")
	}
	return nil
}

func (d *Database) GetContainer(containerID string) (*ContainerInfo, error) {
	d.logger.Info("Fetching container: %s", containerID)
	var info ContainerInfo
	err := d.db.QueryRow("SELECT * FROM containers WHERE container_id = ?", containerID).Scan(
		&info.ID, &info.ContainerID, &info.ContainerName, &info.ImageName,
		&info.DomainName, &info.HostPort, &info.ContainerPort, &info.Status)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("no container found with ID %s", containerID)
		}
		return nil, fmt.Errorf("failed to get container: %w", err)
	}
	return &info, nil
}

func (d *Database) GetContainersByPartialName(partialName string) ([]ContainerInfo, error) {
	d.logger.Info("Fetching containers by partial name: %s", partialName)
	var containers []ContainerInfo

	// Use LIKE with % wildcards for partial matching
	query := "SELECT * FROM containers WHERE container_name LIKE ?"
	rows, err := d.db.Query(query, "%"+partialName+"%")
	if err != nil {
		return nil, fmt.Errorf("failed to query containers by name: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var info ContainerInfo
		err := rows.Scan(
			&info.ID, &info.ContainerID, &info.ContainerName, &info.ImageName,
			&info.DomainName, &info.HostPort, &info.ContainerPort, &info.Status)
		if err != nil {
			return nil, fmt.Errorf("failed to scan container row: %w", err)
		}
		containers = append(containers, info)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating container rows: %w", err)
	}

	if len(containers) == 0 {
		return nil, fmt.Errorf("no containers found with name containing '%s'", partialName)
	}

	return containers, nil
}

func (d *Database) ListContainers() ([]ContainerInfo, error) {
	d.logger.Info("Listing all containers")
	rows, err := d.db.Query("SELECT * FROM containers")
	if err != nil {
		return nil, fmt.Errorf("failed to query containers: %w", err)
	}
	defer rows.Close()
	var containers []ContainerInfo
	for rows.Next() {
		var info ContainerInfo
		err := rows.Scan(&info.ID, &info.ContainerID, &info.ContainerName, &info.ImageName,
			&info.DomainName, &info.HostPort, &info.ContainerPort, &info.Status)
		if err != nil {
			return nil, fmt.Errorf("failed to scan container row: %w", err)
		}
		d.logger.Info("Found container: %+v", info)
		containers = append(containers, info)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating container rows: %w", err)
	}
	return containers, nil
}

func (d *Database) DeleteContainer(containerID string) error {
	d.logger.Info("Deleting container: %s", containerID)
	result, err := d.db.Exec("DELETE FROM containers WHERE container_id = ?", containerID)
	if err != nil {
		return fmt.Errorf("failed to delete container: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return errors.New("no container found with the given ID")
	}
	return nil
}

func validateContainerInfo(info ContainerInfo) error {
	if info.ContainerID == "" {
		return errors.New("container ID cannot be empty")
	}
	if info.ContainerName == "" {
		return errors.New("container name cannot be empty")
	}
	if info.ImageName == "" {
		return errors.New("image name cannot be empty")
	}
	if info.DomainName == "" {
		return errors.New("domain name cannot be empty")
	}

	if info.Status == "" {
		return errors.New("status cannot be empty")
	}
	return nil
}
