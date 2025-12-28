package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/heysubinoy/pyazdb/api/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	// Connect to gRPC server
	conn, err := grpc.NewClient("localhost:9091", grpc.WithTransportCredentials(insecure.NewCredentials()))
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
}
