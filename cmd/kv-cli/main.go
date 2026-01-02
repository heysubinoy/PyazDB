package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/heysubinoy/pyazdb/api/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type LeaderInfo struct {
	ID       string `json:"id"`
	Addr     string `json:"addr"`
	HTTPAddr string `json:"http_addr"`
	GRPCAddr string `json:"grpc_addr"`
	Term     uint64 `json:"term"`
}

func getLeaderGRPCAddr(mandiAddr string) (string, error) {
	resp, err := http.Get(mandiAddr + "/leader")
	if err != nil {
		return "", fmt.Errorf("failed to query mandi: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("no leader available (status: %d)", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	var leader LeaderInfo
	if err := json.Unmarshal(body, &leader); err != nil {
		return "", fmt.Errorf("failed to parse leader info: %w", err)
	}

	if leader.GRPCAddr == "" {
		return "", fmt.Errorf("leader gRPC address not available")
	}

	// If the address starts with ":", it's missing a host - use localhost
	grpcAddr := leader.GRPCAddr
	if len(grpcAddr) > 0 && grpcAddr[0] == ':' {
		grpcAddr = "localhost" + grpcAddr
	}

	return grpcAddr, nil
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	// Get mandi address from environment or use default
	mandiAddr := os.Getenv("MANDI_ADDR")
	if mandiAddr == "" {
		mandiAddr = "http://127.0.0.1:7000"
	}

	// Discover leader from mandi
	leaderAddr, err := getLeaderGRPCAddr(mandiAddr)
	if err != nil {
		log.Fatalf("Failed to discover leader: %v", err)
	}

	fmt.Printf("Connecting to leader at %s\n", leaderAddr)

	// Connect to gRPC server using passthrough resolver for direct address connection
	conn, err := grpc.NewClient("passthrough:///"+leaderAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	client := proto.NewKVServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	command := os.Args[1]

	switch command {
	case "get":
		if len(os.Args) < 3 {
			fmt.Println("Usage: kv-cli get <key>")
			os.Exit(1)
		}
		handleGet(ctx, client, os.Args[2])

	case "set":
		if len(os.Args) < 4 {
			fmt.Println("Usage: kv-cli set <key> <value>")
			os.Exit(1)
		}
		handleSet(ctx, client, os.Args[2], os.Args[3])

	case "delete":
		if len(os.Args) < 3 {
			fmt.Println("Usage: kv-cli delete <key>")
			os.Exit(1)
		}
		handleDelete(ctx, client, os.Args[2])

	default:
		fmt.Printf("Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func handleGet(ctx context.Context, client proto.KVServiceClient, key string) {
	resp, err := client.Get(ctx, &proto.GetRequest{Key: key})
	if err != nil {
		log.Fatalf("Get failed: %v", err)
	}

	if resp.Found {
		fmt.Println(resp.Value)
	} else {
		fmt.Printf("Key '%s' not found\n", key)
		os.Exit(1)
	}
}

func handleSet(ctx context.Context, client proto.KVServiceClient, key, value string) {
	resp, err := client.Set(ctx, &proto.SetRequest{
		Key:   key,
		Value: value,
	})
	if err != nil {
		log.Fatalf("Set failed: %v", err)
	}

	if resp.Success {
		fmt.Printf("Set '%s' = '%s'\n", key, value)
	}
}

func handleDelete(ctx context.Context, client proto.KVServiceClient, key string) {
	resp, err := client.Delete(ctx, &proto.DeleteRequest{Key: key})
	if err != nil {
		log.Fatalf("Delete failed: %v", err)
	}

	if resp.Success {
		fmt.Printf("Deleted '%s'\n", key)
	}
}

func printUsage() {
	fmt.Println("Usage:")
	fmt.Println("  kv-cli get <key>")
	fmt.Println("  kv-cli set <key> <value>")
	fmt.Println("  kv-cli delete <key>")
	fmt.Println("")
	fmt.Println("Environment variables:")
	fmt.Println("  MANDI_ADDR - Mandi discovery service address (default: http://127.0.0.1:7000)")
}
