package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"

	pb "github.com/example/fsdriver/proto"
)

// loggingInterceptor logs client connections and method calls
func loggingInterceptor(share string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// Get client peer information
		p, ok := peer.FromContext(ctx)
		clientAddr := "unknown"
		if ok {
			clientAddr = p.Addr.String()
		}

		// Log the method call
		logx.Info("gRPC method called",
			"method", info.FullMethod,
			"client_addr", clientAddr,
			"share", share)

		// Call the actual handler
		resp, err := handler(ctx, req)

		// Log any errors
		if err != nil {
			logx.Error("gRPC method error",
				"method", info.FullMethod,
				"client_addr", clientAddr,
				"error", err)
		}

		return resp, err
	}
}

// streamLoggingInterceptor logs client connections for streaming methods
func streamLoggingInterceptor(share string) grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		// Get client peer information
		p, ok := peer.FromContext(ss.Context())
		clientAddr := "unknown"
		if ok {
			clientAddr = p.Addr.String()
		}

		// Log the stream start
		logx.Info("gRPC stream started",
			"method", info.FullMethod,
			"client_addr", clientAddr,
			"share", share)

		// Call the actual handler
		err := handler(srv, ss)

		// Log stream end
		if err != nil {
			logx.Error("gRPC stream error",
				"method", info.FullMethod,
				"client_addr", clientAddr,
				"error", err)
		} else {
			logx.Info("gRPC stream ended",
				"method", info.FullMethod,
				"client_addr", clientAddr)
		}

		return err
	}
}

func main() {
	var share string
	var addr string

	flag.StringVar(&share, "share", "", "Windows directory to share (root)")
	flag.StringVar(&addr, "addr", "127.0.0.1:50051", "listen address")
	flag.Parse()

	if share == "" {
		logx.Error("missing required flag", "flag", "--share")
		os.Exit(2)
	}

	// Validate share path exists and is a directory
	info, err := os.Stat(share)
	if err != nil {
		logx.Error("share path invalid", "error", err)
		os.Exit(2)
	}
	if !info.IsDir() {
		logx.Error("share must be a directory", "share", share)
		os.Exit(2)
	}

	// Debug: List directory contents
	entries, err := os.ReadDir(share)
	if err != nil {
		logx.Error("failed to read share directory", "error", err)
		os.Exit(2)
	}
	logx.Info("share directory contents", "path", share, "count", len(entries))
	for i, entry := range entries {
		if i < 10 { // Show first 10 entries
			logx.Info("entry", "name", entry.Name(), "is_dir", entry.IsDir())
		}
	}
	if len(entries) > 10 {
		logx.Info("... and more entries", "total", len(entries))
	}

	lis, err := net.Listen("tcp", addr)
	if err != nil {
		logx.Error("failed to listen", "addr", addr, "error", err)
		os.Exit(1)
	}

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(loggingInterceptor(share)),
		grpc.StreamInterceptor(streamLoggingInterceptor(share)),
	)
	srv, err := NewFileSystemServer(share)
	if err != nil {
		logx.Error("failed to initialize server", "error", err)
		os.Exit(1)
	}
	pb.RegisterFileSystemServiceServer(grpcServer, srv)

	logx.Info("fsdriver server listening", "addr", addr, "share", share)

	// Show all available network interfaces
	interfaces, err := net.Interfaces()
	if err == nil {
		logx.Info("available network interfaces:")
		for _, iface := range interfaces {
			addrs, err := iface.Addrs()
			if err == nil && len(addrs) > 0 {
				logx.Info("interface", "name", iface.Name, "addresses", addrs)
			}
		}
	}

	logx.Info("server ready to accept connections")
	if err := grpcServer.Serve(lis); err != nil {
		fmt.Fprintf(os.Stderr, "server error: %v\n", err)
		os.Exit(1)
	}
}
