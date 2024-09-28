package main

import (
	"fmt"
	"os"

	"github.com/dgunzy/go-container-orchestrator/internal/logging"
)

func main() {
	err := logging.Setup("./container_manager_logs")
	if err != nil {
		fmt.Printf("Failed to set up logging: %v\n", err)
		os.Exit(1)
	}
	defer logging.CloseGlobalLogger()

	logger := logging.GetLogger()
	logger.Info("Application started")
	fmt.Println("Container time")
}
