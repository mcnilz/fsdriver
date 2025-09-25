package main

import (
    "flag"
    "fmt"
    "log"
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
        log.Fatal("--share is required")
    }

    // Validate share path exists and is a directory
    info, err := os.Stat(share)
    if err != nil {
        log.Fatalf("share path invalid: %v", err)
    }
    if !info.IsDir() {
        log.Fatalf("share must be a directory: %s", share)
    }

    lis, err := net.Listen("tcp", addr)
    if err != nil {
        log.Fatalf("failed to listen on %s: %v", addr, err)
    }

    grpcServer := grpc.NewServer()
    srv, err := NewFileSystemServer(share)
    if err != nil {
        log.Fatalf("failed to initialize server: %v", err)
    }
    pb.RegisterFileSystemServiceServer(grpcServer, srv)

    log.Printf("fsdriver server listening on %s (share=%s)", addr, share)
    if err := grpcServer.Serve(lis); err != nil {
        fmt.Fprintf(os.Stderr, "server error: %v\n", err)
        os.Exit(1)
    }
}


