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
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Printf("Failed to connect to server: %v", err)
		return nil, fmt.Errorf("dial server: %w", err)
	}
	log.Printf("gRPC connection established to %s", addr)
	
	// Test the connection with a simple call
	client := pb.NewFileSystemServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	// Try to stat the root directory to test connection
	_, err = client.Stat(ctx, &pb.StatRequest{Path: ""})
	if err != nil {
		log.Printf("Connection test failed: %v", err)
		conn.Close()
		return nil, fmt.Errorf("connection test failed: %w", err)
	}
	
	log.Printf("Successfully connected and tested server at %s", addr)
	return &grpcClient{
		conn:   conn,
		client: client,
	}, nil
}

func (c *grpcClient) Close() error {
	return c.conn.Close()
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
	
	resp, err := client.ReadDir(ctx, &pb.ReadDirRequest{
		Path:   path,
		Offset: offset,
		Limit:  limit,
	})
	if err != nil {
		return nil, false, err
	}
	
	if resp.Error != nil {
		return nil, false, fmt.Errorf("readdir error %d: %s", resp.Error.Code, resp.Error.Message)
	}
	
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
