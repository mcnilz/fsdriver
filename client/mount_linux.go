//go:build linux

package main

import (
    "fmt"
)

func mountRemote(share, mountpoint, addr string, readOnly bool) error {
    // Placeholder: implement go-fuse filesystem and gRPC wiring in next steps
    return fmt.Errorf("FUSE mount not yet implemented: share=%s mount=%s addr=%s ro=%v", share, mountpoint, addr, readOnly)
}


