package main

import (
    "flag"
    "fmt"
    "net"
    "os"

    "google.golang.org/grpc"

    pb "github.com/example/fsdriver/proto"
)

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

    lis, err := net.Listen("tcp", addr)
    if err != nil {
        logx.Error("failed to listen", "addr", addr, "error", err)
        os.Exit(1)
    }

    grpcServer := grpc.NewServer()
    srv, err := NewFileSystemServer(share)
    if err != nil {
        logx.Error("failed to initialize server", "error", err)
        os.Exit(1)
    }
    pb.RegisterFileSystemServiceServer(grpcServer, srv)

    logx.Info("fsdriver server listening", "addr", addr, "share", share)
    if err := grpcServer.Serve(lis); err != nil {
        fmt.Fprintf(os.Stderr, "server error: %v\n", err)
        os.Exit(1)
    }
}


