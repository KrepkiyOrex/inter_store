// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v5.27.3
// source: inventory.proto

package inventory

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

// InventoryServiceClient is the client API for InventoryService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type InventoryServiceClient interface {
	AddInventory(ctx context.Context, in *AddInventoryRequest, opts ...grpc.CallOption) (*AddInventoryResponse, error)
	RemoveInventory(ctx context.Context, in *RemoveInventoryRequest, opts ...grpc.CallOption) (*RemoveInventoryResponse, error)
	GetInventory(ctx context.Context, in *GetInventoryRequest, opts ...grpc.CallOption) (*GetInventoryResponse, error)
}

type inventoryServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewInventoryServiceClient(cc grpc.ClientConnInterface) InventoryServiceClient {
	return &inventoryServiceClient{cc}
}

func (c *inventoryServiceClient) AddInventory(ctx context.Context, in *AddInventoryRequest, opts ...grpc.CallOption) (*AddInventoryResponse, error) {
	out := new(AddInventoryResponse)
	err := c.cc.Invoke(ctx, "/inventory.InventoryService/AddInventory", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *inventoryServiceClient) RemoveInventory(ctx context.Context, in *RemoveInventoryRequest, opts ...grpc.CallOption) (*RemoveInventoryResponse, error) {
	out := new(RemoveInventoryResponse)
	err := c.cc.Invoke(ctx, "/inventory.InventoryService/RemoveInventory", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *inventoryServiceClient) GetInventory(ctx context.Context, in *GetInventoryRequest, opts ...grpc.CallOption) (*GetInventoryResponse, error) {
	out := new(GetInventoryResponse)
	err := c.cc.Invoke(ctx, "/inventory.InventoryService/GetInventory", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// InventoryServiceServer is the server API for InventoryService service.
// All implementations must embed UnimplementedInventoryServiceServer
// for forward compatibility
type InventoryServiceServer interface {
	AddInventory(context.Context, *AddInventoryRequest) (*AddInventoryResponse, error)
	RemoveInventory(context.Context, *RemoveInventoryRequest) (*RemoveInventoryResponse, error)
	GetInventory(context.Context, *GetInventoryRequest) (*GetInventoryResponse, error)
	mustEmbedUnimplementedInventoryServiceServer()
}

// UnimplementedInventoryServiceServer must be embedded to have forward compatible implementations.
type UnimplementedInventoryServiceServer struct {
}

func (UnimplementedInventoryServiceServer) AddInventory(context.Context, *AddInventoryRequest) (*AddInventoryResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method AddInventory not implemented")
}
func (UnimplementedInventoryServiceServer) RemoveInventory(context.Context, *RemoveInventoryRequest) (*RemoveInventoryResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RemoveInventory not implemented")
}
func (UnimplementedInventoryServiceServer) GetInventory(context.Context, *GetInventoryRequest) (*GetInventoryResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetInventory not implemented")
}
func (UnimplementedInventoryServiceServer) mustEmbedUnimplementedInventoryServiceServer() {}

// UnsafeInventoryServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to InventoryServiceServer will
// result in compilation errors.
type UnsafeInventoryServiceServer interface {
	mustEmbedUnimplementedInventoryServiceServer()
}

func RegisterInventoryServiceServer(s grpc.ServiceRegistrar, srv InventoryServiceServer) {
	s.RegisterService(&InventoryService_ServiceDesc, srv)
}

func _InventoryService_AddInventory_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(AddInventoryRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(InventoryServiceServer).AddInventory(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/inventory.InventoryService/AddInventory",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(InventoryServiceServer).AddInventory(ctx, req.(*AddInventoryRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _InventoryService_RemoveInventory_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RemoveInventoryRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(InventoryServiceServer).RemoveInventory(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/inventory.InventoryService/RemoveInventory",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(InventoryServiceServer).RemoveInventory(ctx, req.(*RemoveInventoryRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _InventoryService_GetInventory_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetInventoryRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(InventoryServiceServer).GetInventory(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/inventory.InventoryService/GetInventory",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(InventoryServiceServer).GetInventory(ctx, req.(*GetInventoryRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// InventoryService_ServiceDesc is the grpc.ServiceDesc for InventoryService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var InventoryService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "inventory.InventoryService",
	HandlerType: (*InventoryServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "AddInventory",
			Handler:    _InventoryService_AddInventory_Handler,
		},
		{
			MethodName: "RemoveInventory",
			Handler:    _InventoryService_RemoveInventory_Handler,
		},
		{
			MethodName: "GetInventory",
			Handler:    _InventoryService_GetInventory_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "inventory.proto",
}
