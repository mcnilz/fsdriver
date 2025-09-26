//go:build linux

package main

import (
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
	// Create gRPC client
	client, err := newGRPCClient(addr)
	if err != nil {
		return fmt.Errorf("connect to server: %w", err)
	}
	defer client.Close()

	// Create FUSE filesystem
	fuseFS := newFuseFS(client, share)

	// Mount options
	entryTimeout := 1 * time.Second
	attrTimeout := 1 * time.Second
	opts := &fs.Options{
		MountOptions: fuse.MountOptions{
			Debug: false,
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
