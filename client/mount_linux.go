//go:build linux

package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
)

func mountRemote(share, mountpoint, addr string, readOnly bool) error {
	log.Printf("Starting mount: share=%s, mountpoint=%s, addr=%s, ro=%v", share, mountpoint, addr, readOnly)

	// Create gRPC client
	client, err := newGRPCClient(addr)
	if err != nil {
		return fmt.Errorf("connect to server: %w", err)
	}
	defer client.Close()

	log.Printf("gRPC client created successfully")

	// Test connection before proceeding with mount
	log.Printf("Testing connection to server...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.TestConnection(ctx); err != nil {
		log.Printf("Connection test failed: %v", err)
		return fmt.Errorf("server connection test failed - please ensure server is running and accessible: %w", err)
	}

	log.Printf("Connection test passed - server is reachable and responsive")
	log.Printf("Proceeding with FUSE mount...")

	// Create FUSE filesystem
	fuseFS := newFuseFS(client, share)

	// Mount options
	entryTimeout := 1 * time.Second
	attrTimeout := 1 * time.Second
	opts := &fs.Options{
		MountOptions: fuse.MountOptions{
			Debug: true, // Enable debug logging
			// Disable ReadDirPlus to force use of ReadDir
			DisableReadDirPlus: true,
		},
		EntryTimeout: &entryTimeout,
		AttrTimeout:  &attrTimeout,
	}

	// Mount the filesystem
	server, err := fs.Mount(mountpoint, fuseFS, opts)
	if err != nil {
		return fmt.Errorf("mount failed: %w", err)
	}

	log.Printf("FUSE filesystem mounted at %s (share=%s, server=%s)", mountpoint, share, addr)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigCh
		log.Println("Received signal, unmounting...")
		_ = server.Unmount()
	}()

	// Serve the filesystem
	server.Wait()

	return nil
}
