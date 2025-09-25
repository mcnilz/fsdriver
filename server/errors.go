package main

import (
    "errors"
    "os"

    pb "github.com/example/fsdriver/proto"
)

// errno maps generic errors to POSIX-like codes; extend with Win32 specifics later.
func errno(err error) *pb.Error {
    if errors.Is(err, os.ErrNotExist) {
        return &pb.Error{Code: 2, Message: err.Error()} // ENOENT
    }
    if errors.Is(err, os.ErrPermission) {
        return &pb.Error{Code: 13, Message: err.Error()} // EACCES
    }
    return &pb.Error{Code: 5, Message: err.Error()} // EIO default
}


