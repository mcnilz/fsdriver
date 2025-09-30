//go:build linux

package main

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "github.com/example/fsdriver/proto"
)

type grpcClient struct {
	conn   *grpc.ClientConn
	client pb.FileSystemServiceClient
	mu     sync.RWMutex
}

func newGRPCClient(addr string) (*grpcClient, error) {
	log.Printf("Connecting to server at %s", addr)
	conn, err := grpc.Dial(addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),                 // Wait for connection to be ready
		grpc.WithTimeout(10*time.Second), // Connection timeout
	)
	if err != nil {
		log.Printf("Failed to connect to server: %v", err)
		return nil, fmt.Errorf("dial server: %w", err)
	}
	log.Printf("gRPC connection established to %s", addr)

	log.Printf("Successfully connected to server at %s", addr)
	return &grpcClient{
		conn:   conn,
		client: pb.NewFileSystemServiceClient(conn),
	}, nil
}

func (c *grpcClient) Close() error {
	return c.conn.Close()
}

// TestConnection performs a simple test to verify the server is reachable and responsive
func (c *grpcClient) TestConnection(ctx context.Context) error {
	c.mu.RLock()
	client := c.client
	c.mu.RUnlock()

	// Try to stat the root directory as a simple connectivity test
	resp, err := client.Stat(ctx, &pb.StatRequest{Path: "."})
	if err != nil {
		return fmt.Errorf("connection test failed: %w", err)
	}

	// Check if we got a valid response
	switch result := resp.Result.(type) {
	case *pb.StatResponse_Info:
		if !result.Info.IsDir {
			return fmt.Errorf("server root is not a directory")
		}
		log.Printf("Connection test successful - server root is accessible (name: %s)", result.Info.Name)
	case *pb.StatResponse_Error:
		return fmt.Errorf("server returned error: %d - %s", result.Error.Code, result.Error.Message)
	default:
		return fmt.Errorf("unexpected server response")
	}

	return nil
}

func (c *grpcClient) Stat(ctx context.Context, path string) (*pb.FileInfo, error) {
	c.mu.RLock()
	client := c.client
	c.mu.RUnlock()

	resp, err := client.Stat(ctx, &pb.StatRequest{Path: path})
	if err != nil {
		return nil, err
	}

	switch result := resp.Result.(type) {
	case *pb.StatResponse_Info:
		return result.Info, nil
	case *pb.StatResponse_Error:
		return nil, fmt.Errorf("stat error %d: %s", result.Error.Code, result.Error.Message)
	default:
		return nil, fmt.Errorf("unexpected stat response")
	}
}

func (c *grpcClient) ReadDir(ctx context.Context, path string, offset, limit int32) ([]*pb.FileInfo, bool, error) {
	c.mu.RLock()
	client := c.client
	c.mu.RUnlock()

	log.Printf("gRPC ReadDir call: path=%s, offset=%d, limit=%d", path, offset, limit)

	resp, err := client.ReadDir(ctx, &pb.ReadDirRequest{
		Path:   path,
		Offset: offset,
		Limit:  limit,
	})
	if err != nil {
		log.Printf("gRPC ReadDir call failed: %v", err)
		return nil, false, err
	}

	if resp.Error != nil {
		log.Printf("gRPC ReadDir server error: %d - %s", resp.Error.Code, resp.Error.Message)
		return nil, false, fmt.Errorf("readdir error %d: %s", resp.Error.Code, resp.Error.Message)
	}

	log.Printf("gRPC ReadDir call successful: %d entries, hasMore=%v", len(resp.Entries), resp.HasMore)
	return resp.Entries, resp.HasMore, nil
}

func (c *grpcClient) Open(ctx context.Context, path string, flags int32) (int32, error) {
	c.mu.RLock()
	client := c.client
	c.mu.RUnlock()

	resp, err := client.Open(ctx, &pb.OpenRequest{Path: path, Flags: flags})
	if err != nil {
		return 0, err
	}

	switch result := resp.Result.(type) {
	case *pb.OpenResponse_Handle:
		return result.Handle, nil
	case *pb.OpenResponse_Error:
		return 0, fmt.Errorf("open error %d: %s", result.Error.Code, result.Error.Message)
	default:
		return 0, fmt.Errorf("unexpected open response")
	}
}

func (c *grpcClient) Read(ctx context.Context, handle int32, offset int64, size int32) ([]byte, error) {
	c.mu.RLock()
	client := c.client
	c.mu.RUnlock()

	resp, err := client.Read(ctx, &pb.ReadRequest{
		Handle: handle,
		Offset: offset,
		Size:   size,
	})
	if err != nil {
		return nil, err
	}

	switch result := resp.Result.(type) {
	case *pb.ReadResponse_Data:
		return result.Data, nil
	case *pb.ReadResponse_Error:
		return nil, fmt.Errorf("read error %d: %s", result.Error.Code, result.Error.Message)
	default:
		return nil, fmt.Errorf("unexpected read response")
	}
}

func (c *grpcClient) CloseHandle(ctx context.Context, handle int32) error {
	c.mu.RLock()
	client := c.client
	c.mu.RUnlock()

	resp, err := client.Close(ctx, &pb.CloseRequest{Handle: handle})
	if err != nil {
		return err
	}

	if resp.Error != nil {
		return fmt.Errorf("close error %d: %s", resp.Error.Code, resp.Error.Message)
	}

	return nil
}
