package main

import (
    "flag"
    "fmt"
    "os"
)

func main() {
    var share string
    var mountpoint string
    var addr string
    var readOnly bool

    flag.StringVar(&share, "share", "", "Share name or path exposed by server")
    flag.StringVar(&mountpoint, "mountpoint", "", "Mount point (Linux)")
    flag.StringVar(&addr, "addr", "127.0.0.1:50051", "server address")
    flag.BoolVar(&readOnly, "ro", true, "mount read-only")
    
    flag.Usage = func() {
        fmt.Fprintf(os.Stderr, "Usage: %s [options]\n\n", os.Args[0])
        fmt.Fprintf(os.Stderr, "Options:\n")
        flag.PrintDefaults()
        fmt.Fprintf(os.Stderr, "\nExamples:\n")
        fmt.Fprintf(os.Stderr, "  %s --share test --mountpoint /mnt/fsdriver/test\n", os.Args[0])
        fmt.Fprintf(os.Stderr, "  %s --share test --mountpoint /mnt/fsdriver/test --addr 127.0.0.1:50055\n", os.Args[0])
    }
    
    flag.Parse()

    // Validate required parameters
    if share == "" {
        fmt.Fprintln(os.Stderr, "Error: --share is required")
        flag.Usage()
        os.Exit(2)
    }
    
    if mountpoint == "" {
        fmt.Fprintln(os.Stderr, "Error: --mountpoint is required")
        flag.Usage()
        os.Exit(2)
    }
    
    // Validate mountpoint exists and is a directory
    if info, err := os.Stat(mountpoint); err != nil {
        fmt.Fprintf(os.Stderr, "Error: mountpoint '%s' does not exist: %v\n", mountpoint, err)
        os.Exit(2)
    } else if !info.IsDir() {
        fmt.Fprintf(os.Stderr, "Error: mountpoint '%s' is not a directory\n", mountpoint)
        os.Exit(2)
    }

    if err := mountRemote(share, mountpoint, addr, readOnly); err != nil {
        fmt.Fprintf(os.Stderr, "Mount error: %v\n", err)
        os.Exit(1)
    }
}


