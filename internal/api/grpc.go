package api

import (
	"context"

	"github.com/heysubinoy/pyazdb/api/proto"
	"github.com/heysubinoy/pyazdb/pkg/kv"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// GRPCServer implements the proto.KVServiceServer interface.
// It wraps a kv.Store and exposes it over gRPC.
type GRPCServer struct {
	proto.UnimplementedKVServiceServer
	Store kv.Store
}

// NewGRPCServer creates a new gRPC server with the given store.
func NewGRPCServer(store kv.Store) *GRPCServer {
	return &GRPCServer{
		Store: store,
	}
}

// Get retrieves a value by key.
func (s *GRPCServer) Get(ctx context.Context, req *proto.GetRequest) (*proto.GetResponse, error) {
	if req.Key == "" {
		return nil, status.Error(codes.InvalidArgument, "key is required")
	}

	value, found := s.Store.Get(req.Key)
	return &proto.GetResponse{
		Value: value,
		Found: found,
	}, nil
}

// Set stores a key-value pair.
func (s *GRPCServer) Set(ctx context.Context, req *proto.SetRequest) (*proto.SetResponse, error) {
	if req.Key == "" {
		return nil, status.Error(codes.InvalidArgument, "key is required")
	}

	if err := s.Store.Set(req.Key, req.Value); err != nil {
		return nil, status.Error(codes.Internal, "failed to set key")
	}

	return &proto.SetResponse{
		Success: true,
	}, nil
}

// Delete removes a key from the store.
func (s *GRPCServer) Delete(ctx context.Context, req *proto.DeleteRequest) (*proto.DeleteResponse, error) {
	if req.Key == "" {
		return nil, status.Error(codes.InvalidArgument, "key is required")
	}

	if err := s.Store.Delete(req.Key); err != nil {
		return nil, status.Error(codes.Internal, "failed to delete key")
	}

	return &proto.DeleteResponse{
		Success: true,
	}, nil
}
