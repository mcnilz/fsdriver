//go:build !linux

package main

import "fmt"

func mountRemote(share, mountpoint, addr string, readOnly bool) error {
    return fmt.Errorf("FUSE mount only supported on linux builds")
}


