package main

import (
	"context"
	"log"
	"net"

	"github.com/dgunzy/go-container-orchestrator/internal/container"
	pb "github.com/dgunzy/go-container-orchestrator/pkg/proto"
	"google.golang.org/grpc"
)

type server struct {
	pb.UnimplementedContainerServiceServer
	cm *container.ContainerManager
}

func (s *server) CreateContainer(ctx context.Context, req *pb.CreateContainerRequest) (*pb.CreateContainerResponse, error) {
	config := &container.ContainerConfig{
		DomainName:       req.Config.DomainName,
		ImageName:        req.Config.ImageName,
		ContainerName:    req.Config.ContainerName,
		ContainerPort:    req.Config.ContainerPort,
		RegistryUsername: req.Config.RegistryUsername,
		RegistryPassword: req.Config.RegistryPassword,
	}

	err := s.cm.CreateNewContainer(ctx, config)
	if err != nil {
		s.cm.Logger.Error("Error creating container: %v", err)
		return nil, err
	}

	return &pb.CreateContainerResponse{ContainerId: config.ContainerID}, nil
}

func (s *server) ListContainers(ctx context.Context, req *pb.ListContainersRequest) (*pb.ListContainersResponse, error) {
	containers, err := s.cm.ListContainers()
	if err != nil {
		s.cm.Logger.Error("Error creating container: %v", err)
		return nil, err
	}

	var pbContainers []*pb.ContainerConfig
	for _, c := range containers {
		pbContainers = append(pbContainers, &pb.ContainerConfig{
			DomainName:    c.DomainName,
			ImageName:     c.ImageName,
			ContainerName: c.ContainerName,
			ContainerId:   c.ContainerID,
			ContainerPort: c.ContainerPort,
			HostPort:      c.HostPort,
			Status:        c.Status,
		})
	}

	return &pb.ListContainersResponse{Containers: pbContainers}, nil
}

func (s *server) UpdateContainer(ctx context.Context, req *pb.UpdateContainerRequest) (*pb.UpdateContainerResponse, error) {
	config := &container.ContainerConfig{
		DomainName:       req.Config.DomainName,
		ImageName:        req.Config.ImageName,
		ContainerName:    req.Config.ContainerName,
		ContainerPort:    req.Config.ContainerPort,
		RegistryUsername: req.Config.RegistryUsername,
		RegistryPassword: req.Config.RegistryPassword,
	}

	err := s.cm.UpdateExistingContainer(ctx, config)
	if err != nil {
		s.cm.Logger.Error("Error creating container: %v", err)
		return nil, err
	}

	return &pb.UpdateContainerResponse{Success: true}, nil
}

func (s *server) RemoveContainer(ctx context.Context, req *pb.RemoveContainerRequest) (*pb.RemoveContainerResponse, error) {
	err := s.cm.RemoveContainer(ctx, req.ContainerName)
	if err != nil {
		s.cm.Logger.Error("Error creating container: %v", err)
		return nil, err
	}

	return &pb.RemoveContainerResponse{Success: true}, nil
}

func main() {
	cm, err := container.NewContainerManager()
	if err != nil {
		log.Fatalf("Failed to create ContainerManager: %v", err)
	}

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterContainerServiceServer(s, &server{cm: cm})

	log.Printf("Server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
