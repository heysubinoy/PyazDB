package main

import (
	"log"
	"net"
	"net/http"

	"os"
	"path/filepath"

	"github.com/hashicorp/raft"
	raftboltdb "github.com/hashicorp/raft-boltdb"
	"github.com/heysubinoy/pyazdb/api/proto"
	"github.com/heysubinoy/pyazdb/internal/api"
	"github.com/heysubinoy/pyazdb/internal/store"
	"google.golang.org/grpc"
)

func setupRaft(memStore *store.MemStore) *store.RaftStore {
	dir := "./pyaz"
	os.MkdirAll(dir, 0700)

	config := raft.DefaultConfig()
	config.LocalID = raft.ServerID("node1")

	logStore, err := raftboltdb.NewBoltStore(filepath.Join(dir, "raft-log.bolt"))
	if err != nil {
		log.Fatalf("Failed to create log store: %v", err)
	}
	stableStore, err := raftboltdb.NewBoltStore(filepath.Join(dir, "raft-stable.bolt"))
	if err != nil {
		log.Fatalf("Failed to create stable store: %v", err)
	}
	snapshots, err := raft.NewFileSnapshotStore(dir, 1, os.Stdout)
	if err != nil {
		log.Fatalf("Failed to create snapshot store: %v", err)
	}

	_, inmemTransport := raft.NewInmemTransport(raft.ServerAddress("127.0.0.1:12000"))

	raftStore := store.NewRaftStore(memStore, nil)
	raftNode, _ := raft.NewRaft(config, raftStore, logStore, stableStore, snapshots, inmemTransport)

	// Set the raft instance in raftStore after creation
	// (to avoid cyclic dependency during construction)
	raftStoreWithNode := store.NewRaftStore(memStore, raftNode)

	cfg := raft.Configuration{
		Servers: []raft.Server{
			{
				ID:      config.LocalID,
				Address: inmemTransport.LocalAddr(),
			},
		},
	}
	raftNode.BootstrapCluster(cfg)

	return raftStoreWithNode
}

func main() {
	memStore := store.NewMemStore()
	raftStore := setupRaft(memStore)

	go func() {
		grpcAddr := ":9090"
		lis, err := net.Listen("tcp", grpcAddr)
		if err != nil {
			log.Fatalf("Failed to listen on %s: %v", grpcAddr, err)
		}

		grpcServer := grpc.NewServer()
		kvServer := api.NewGRPCServer(raftStore)
		proto.RegisterKVServiceServer(grpcServer, kvServer)

		log.Printf("gRPC server listening on %s", grpcAddr)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("Failed to serve gRPC: %v", err)
		}
	}()

	srv := api.NewServer(raftStore)
	mux := http.NewServeMux()
	srv.RegisterRoutes(mux)

	httpAddr := ":8080"
	log.Printf("HTTP server listening on %s", httpAddr)

	if err := http.ListenAndServe(httpAddr, mux); err != nil {
		log.Fatal(err)
	}
}
