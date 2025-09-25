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
    flag.Parse()

    if share == "" || mountpoint == "" {
        fmt.Fprintln(os.Stderr, "usage: fsdriver-mount --share <name> --mountpoint </path> [--addr host:port] [--ro]")
        os.Exit(2)
    }

    if err := mountRemote(share, mountpoint, addr, readOnly); err != nil {
        fmt.Fprintln(os.Stderr, "mount error:", err)
        os.Exit(1)
    }
}


