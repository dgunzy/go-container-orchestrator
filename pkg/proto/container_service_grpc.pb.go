package proto

import (
	context "context"

	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// ContainerServiceClient is the client API for ContainerService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type ContainerServiceClient interface {
	CreateContainer(ctx context.Context, in *CreateContainerRequest, opts ...grpc.CallOption) (*CreateContainerResponse, error)
	ListContainers(ctx context.Context, in *ListContainersRequest, opts ...grpc.CallOption) (*ListContainersResponse, error)
	UpdateContainer(ctx context.Context, in *UpdateContainerRequest, opts ...grpc.CallOption) (*UpdateContainerResponse, error)
	RemoveContainer(ctx context.Context, in *RemoveContainerRequest, opts ...grpc.CallOption) (*RemoveContainerResponse, error)
}

type containerServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewContainerServiceClient(cc grpc.ClientConnInterface) ContainerServiceClient {
	return &containerServiceClient{cc}
}

func (c *containerServiceClient) CreateContainer(ctx context.Context, in *CreateContainerRequest, opts ...grpc.CallOption) (*CreateContainerResponse, error) {
	out := new(CreateContainerResponse)
	err := c.cc.Invoke(ctx, "/containerservice.ContainerService/CreateContainer", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *containerServiceClient) ListContainers(ctx context.Context, in *ListContainersRequest, opts ...grpc.CallOption) (*ListContainersResponse, error) {
	out := new(ListContainersResponse)
	err := c.cc.Invoke(ctx, "/containerservice.ContainerService/ListContainers", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *containerServiceClient) UpdateContainer(ctx context.Context, in *UpdateContainerRequest, opts ...grpc.CallOption) (*UpdateContainerResponse, error) {
	out := new(UpdateContainerResponse)
	err := c.cc.Invoke(ctx, "/containerservice.ContainerService/UpdateContainer", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *containerServiceClient) RemoveContainer(ctx context.Context, in *RemoveContainerRequest, opts ...grpc.CallOption) (*RemoveContainerResponse, error) {
	out := new(RemoveContainerResponse)
	err := c.cc.Invoke(ctx, "/containerservice.ContainerService/RemoveContainer", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ContainerServiceServer is the server API for ContainerService service.
// All implementations must embed UnimplementedContainerServiceServer
// for forward compatibility
type ContainerServiceServer interface {
	CreateContainer(context.Context, *CreateContainerRequest) (*CreateContainerResponse, error)
	ListContainers(context.Context, *ListContainersRequest) (*ListContainersResponse, error)
	UpdateContainer(context.Context, *UpdateContainerRequest) (*UpdateContainerResponse, error)
	RemoveContainer(context.Context, *RemoveContainerRequest) (*RemoveContainerResponse, error)
	mustEmbedUnimplementedContainerServiceServer()
}

// UnimplementedContainerServiceServer must be embedded to have forward compatible implementations.
type UnimplementedContainerServiceServer struct {
}

func (UnimplementedContainerServiceServer) CreateContainer(context.Context, *CreateContainerRequest) (*CreateContainerResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateContainer not implemented")
}
func (UnimplementedContainerServiceServer) ListContainers(context.Context, *ListContainersRequest) (*ListContainersResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListContainers not implemented")
}
func (UnimplementedContainerServiceServer) UpdateContainer(context.Context, *UpdateContainerRequest) (*UpdateContainerResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateContainer not implemented")
}
func (UnimplementedContainerServiceServer) RemoveContainer(context.Context, *RemoveContainerRequest) (*RemoveContainerResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RemoveContainer not implemented")
}
func (UnimplementedContainerServiceServer) mustEmbedUnimplementedContainerServiceServer() {}

// UnsafeContainerServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to ContainerServiceServer will
// result in compilation errors.
type UnsafeContainerServiceServer interface {
	mustEmbedUnimplementedContainerServiceServer()
}

func RegisterContainerServiceServer(s grpc.ServiceRegistrar, srv ContainerServiceServer) {
	s.RegisterService(&ContainerService_ServiceDesc, srv)
}

func _ContainerService_CreateContainer_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreateContainerRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ContainerServiceServer).CreateContainer(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/containerservice.ContainerService/CreateContainer",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ContainerServiceServer).CreateContainer(ctx, req.(*CreateContainerRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ContainerService_ListContainers_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListContainersRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ContainerServiceServer).ListContainers(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/containerservice.ContainerService/ListContainers",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ContainerServiceServer).ListContainers(ctx, req.(*ListContainersRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ContainerService_UpdateContainer_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpdateContainerRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ContainerServiceServer).UpdateContainer(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/containerservice.ContainerService/UpdateContainer",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ContainerServiceServer).UpdateContainer(ctx, req.(*UpdateContainerRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ContainerService_RemoveContainer_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RemoveContainerRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ContainerServiceServer).RemoveContainer(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/containerservice.ContainerService/RemoveContainer",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ContainerServiceServer).RemoveContainer(ctx, req.(*RemoveContainerRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// ContainerService_ServiceDesc is the grpc.ServiceDesc for ContainerService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var ContainerService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "containerservice.ContainerService",
	HandlerType: (*ContainerServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "CreateContainer",
			Handler:    _ContainerService_CreateContainer_Handler,
		},
		{
			MethodName: "ListContainers",
			Handler:    _ContainerService_ListContainers_Handler,
		},
		{
			MethodName: "UpdateContainer",
			Handler:    _ContainerService_UpdateContainer_Handler,
		},
		{
			MethodName: "RemoveContainer",
			Handler:    _ContainerService_RemoveContainer_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "pkg/proto/container_service.proto",
}
