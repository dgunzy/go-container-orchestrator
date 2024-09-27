package main

import (
	"fmt"
	"io"
	"os"

	"github.com/dgunzy/go-container-orchestrator/internal/logging"
)

func main() {
	logFile, err := os.OpenFile("container_manager.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		fmt.Printf("Error opening log file: %s", err)
	}
	defer logFile.Close()
	logger := logging.GetLogger()
	logger.SetOutput(io.MultiWriter(os.Stdout, logFile))
	fmt.Println("Container time")
}
