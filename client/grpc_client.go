//go:build linux

package main

import (
	"context"
	"fmt"
	"sync"

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
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("dial server: %w", err)
	}
	return &grpcClient{
		conn:   conn,
		client: pb.NewFileSystemServiceClient(conn),
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
