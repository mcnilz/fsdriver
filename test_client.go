package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "github.com/example/fsdriver/proto"
)

func main() {
	addr := "127.0.0.1:50052"
	if len(os.Args) > 1 {
		addr = os.Args[1]
	}

	log.Printf("Testing connection to %s", addr)

	conn, err := grpc.Dial(addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
		grpc.WithTimeout(10*time.Second),
	)
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	client := pb.NewFileSystemServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Test ReadDir
	log.Printf("Testing ReadDir...")
	resp, err := client.ReadDir(ctx, &pb.ReadDirRequest{Path: "", Offset: 0, Limit: 10})
	if err != nil {
		log.Fatalf("ReadDir failed: %v", err)
	}

	log.Printf("Success! Found %d entries:", len(resp.Entries))
	for i, entry := range resp.Entries {
		if i < 5 { // Show first 5
			log.Printf("  %s (dir: %v, size: %d)", entry.Name, entry.IsDir, entry.Size)
		}
	}
	if len(resp.Entries) > 5 {
		log.Printf("  ... and %d more", len(resp.Entries)-5)
	}
}
