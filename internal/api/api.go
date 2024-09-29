package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/dgunzy/go-container-orchestrator/internal/container"
	"github.com/dgunzy/go-container-orchestrator/internal/logging"
)

type WebhookPayload struct {
	Action        string  `json:"action"`
	ContainerName string  `json:"container_name"`
	ImageName     string  `json:"image_name"`
	DomainName    string  `json:"domain_name"`
	ContainerPort string  `json:"container_port"`
	Username      *string `json:"username,omitempty"`
	Password      *string `json:"password,omitempty"`
}

type Server struct {
	containerManager *container.ContainerManager
	logger           *logging.Logger
}

func NewServer(cm *container.ContainerManager, logger *logging.Logger) *Server {
	return &Server{
		containerManager: cm,
		logger:           logger,
	}
}

func (s *Server) Start(ctx context.Context, addr string) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/webhook", s.handleWebhook)

	server := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	go func() {
		<-ctx.Done()
		s.logger.Info("Shutting down server...")
		if err := server.Shutdown(context.Background()); err != nil {
			s.logger.Error("Server shutdown error: %v", err)
		}
	}()

	s.logger.Info("Starting webhook server on %s", addr)
	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		return err
	}

	return nil
}
func (s *Server) handleWebhook(w http.ResponseWriter, r *http.Request) {
	s.logger.Info("Received webhook request")
	if r.Method != http.MethodPost {
		s.logger.Error("Method not allowed: %s", r.Method)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var payload WebhookPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		s.logger.Error("Invalid JSON payload: %v", err)
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	s.logger.Info("Received payload: %+v", payload)

	config := &container.ContainerConfig{
		DomainName:    payload.DomainName,
		ImageName:     payload.ImageName,
		ContainerName: payload.ContainerName,
		ContainerPort: payload.ContainerPort,
	}

	// Only set username and password if they are provided
	if payload.Username != nil {
		config.RegistryUsername = *payload.Username
	}
	if payload.Password != nil {
		config.RegistryPassword = *payload.Password
	}

	var err error
	var respMsg string

	switch payload.Action {
	case "create":
		s.logger.Info("Creating new container")
		err = s.containerManager.CreateNewContainer(r.Context(), config)
		respMsg = "Container created successfully"
	case "update":
		s.logger.Info("Updating existing container")
		err = s.containerManager.UpdateExistingContainer(r.Context(), config)
		respMsg = "Container updated successfully"
	default:
		s.logger.Error("Invalid action: %s", payload.Action)
		http.Error(w, "Invalid action", http.StatusBadRequest)
		return
	}

	if err != nil {
		s.logger.Error("Failed to %s container: %v", payload.Action, err)
		http.Error(w, fmt.Sprintf("Internal server error: %v", err), http.StatusInternalServerError)
		return
	}

	s.logger.Info(respMsg)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(respMsg))
}
