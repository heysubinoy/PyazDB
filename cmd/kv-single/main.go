package main

import (
	"log"
	"net"
	"net/http"

	"github.com/heysubinoy/pyazdb/api/proto"
	"github.com/heysubinoy/pyazdb/internal/api"
	"github.com/heysubinoy/pyazdb/internal/store"
	"google.golang.org/grpc"
)

func main() {
	// Create the in-memory store
	memStore := store.NewMemStore()

	// Start gRPC server in a goroutine
	go func() {
		grpcAddr := ":9090"
		lis, err := net.Listen("tcp", grpcAddr)
		if err != nil {
			log.Fatalf("Failed to listen on %s: %v", grpcAddr, err)
		}

		grpcServer := grpc.NewServer()
		kvServer := api.NewGRPCServer(memStore)
		proto.RegisterKVServiceServer(grpcServer, kvServer)

		log.Printf("gRPC server listening on %s", grpcAddr)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("Failed to serve gRPC: %v", err)
		}
	}()

	// Create the HTTP server with the store
	srv := api.NewServer(memStore)

	// Register routes
	mux := http.NewServeMux()
	srv.RegisterRoutes(mux)

	// Start the HTTP server
	httpAddr := ":8080"
	log.Printf("HTTP server listening on %s", httpAddr)

	if err := http.ListenAndServe(httpAddr, mux); err != nil {
		log.Fatal(err)
	}
}
