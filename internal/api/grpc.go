package api

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/hashicorp/raft"
	"github.com/heysubinoy/pyazdb/api/proto"
	"github.com/heysubinoy/pyazdb/pkg/kv"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

// GRPCServer implements the proto.KVServiceServer interface.
// It wraps a kv.Store and exposes it over gRPC.
type GRPCServer struct {
	proto.UnimplementedKVServiceServer
	Store     kv.Store
	Raft      *raft.Raft
	GRPCPort  string
	MandiAddr string
}

// NewGRPCServer creates a new gRPC server with the given store.
func NewGRPCServer(store kv.Store, raftNode *raft.Raft, grpcPort, mandiAddr string) *GRPCServer {
	return &GRPCServer{
		Store:     store,
		Raft:      raftNode,
		GRPCPort:  grpcPort,
		MandiAddr: mandiAddr,
	}
}

// Get retrieves a value by key.
func (s *GRPCServer) Get(ctx context.Context, req *proto.GetRequest) (*proto.GetResponse, error) {
	if req.Key == "" {
		return nil, status.Error(codes.InvalidArgument, "key is required")
	}
	if s.Raft != nil && s.Raft.State() != raft.Leader {
		// Automatically forward to leader
		leaderAddr := s.getLeaderGRPCAddr()
		if leaderAddr == "" {
			return nil, status.Error(codes.Unavailable, "Not leader and no leader known")
		}
		conn, err := grpc.Dial(leaderAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			return nil, status.Errorf(codes.Unavailable, "Cannot connect to leader: %v", err)
		}
		defer conn.Close()
		client := proto.NewKVServiceClient(conn)
		return client.Get(ctx, req)
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
	if s.Raft != nil && s.Raft.State() != raft.Leader {
		// Automatically forward to leader
		leaderAddr := s.getLeaderGRPCAddr()
		if leaderAddr == "" {
			return nil, status.Error(codes.Unavailable, "Not leader and no leader known")
		}
		conn, err := grpc.Dial(leaderAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			return nil, status.Errorf(codes.Unavailable, "Cannot connect to leader: %v", err)
		}
		defer conn.Close()
		client := proto.NewKVServiceClient(conn)
		return client.Set(ctx, req)
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
	if s.Raft != nil && s.Raft.State() != raft.Leader {
		// Automatically forward to leader
		leaderAddr := s.getLeaderGRPCAddr()
		if leaderAddr == "" {
			return nil, status.Error(codes.Unavailable, "Not leader and no leader known")
		}
		conn, err := grpc.Dial(leaderAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			return nil, status.Errorf(codes.Unavailable, "Cannot connect to leader: %v", err)
		}
		defer conn.Close()
		client := proto.NewKVServiceClient(conn)
		return client.Delete(ctx, req)
	}
	if err := s.Store.Delete(req.Key); err != nil {
		return nil, status.Error(codes.Internal, "failed to delete key")
	}
	return &proto.DeleteResponse{
		Success: true,
	}, nil
}

// getLeaderGRPCAddr queries mandi to get the leader's gRPC address
func (s *GRPCServer) getLeaderGRPCAddr() string {
	if s.MandiAddr == "" {
		return ""
	}

	resp, err := http.Get(s.MandiAddr + "/leader")
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return ""
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return ""
	}

	var leaderInfo struct {
		GRPCAddr string `json:"grpc_addr"`
	}

	if err := json.Unmarshal(body, &leaderInfo); err != nil {
		return ""
	}

	return leaderInfo.GRPCAddr
}
