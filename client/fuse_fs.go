//go:build linux

package main

import (
	"context"
	"io"
	"log"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"

	pb "github.com/example/fsdriver/proto"
)

type fuseFS struct {
	fs.Inode
	client *grpcClient
	share  string
	path   string // Current path for this node
}

func newFuseFS(client *grpcClient, share string) *fuseFS {
	return &fuseFS{
		client: client,
		share:  share,
		path:   "", // Root path
	}
}

func (f *fuseFS) Getattr(ctx context.Context, fh fs.FileHandle, out *fuse.Attr) syscall.Errno {
	path := f.getPath(ctx)
	log.Printf("Getattr: path=%s", path)
	if path == "" {
		log.Printf("Getattr: empty path, returning ENOENT")
		return syscall.ENOENT
	}

	info, err := f.client.Stat(ctx, path)
	if err != nil {
		log.Printf("Getattr: Stat failed for path=%s, error=%v", path, err)
		return f.mapError(err)
	}

	log.Printf("Getattr: Stat success for path=%s, name=%s, isDir=%v", path, info.Name, info.IsDir)
	f.fillAttr(info, out)
	return 0
}

func (f *fuseFS) Open(ctx context.Context, fh fs.FileHandle, flags uint32) (fs.FileHandle, uint32, syscall.Errno) {
	path := f.getPath(ctx)
	if path == "" {
		return nil, 0, syscall.ENOENT
	}

	handle, err := f.client.Open(ctx, path, int32(flags))
	if err != nil {
		return nil, 0, f.mapError(err)
	}

	return &fuseFile{client: f.client, handle: handle}, 0, 0
}

func (f *fuseFS) ReadDir(ctx context.Context) (fs.DirStream, syscall.Errno) {
	path := f.getPath(ctx)
	log.Printf("ReadDir: path=%s", path)
	if path == "" {
		log.Printf("ReadDir: empty path, returning ENOENT")
		return nil, syscall.ENOENT
	}

	// For root directory, use empty path to get share contents
	requestPath := ""
	if f.path != "" {
		requestPath = f.path
	}
	log.Printf("ReadDir: calling gRPC ReadDir with requestPath=%s", requestPath)
	
	entries, _, err := f.client.ReadDir(ctx, requestPath, 0, 0)
	if err != nil {
		log.Printf("ReadDir: gRPC call failed, requestPath=%s, error=%v", requestPath, err)
		return nil, f.mapError(err)
	}
	log.Printf("ReadDir: gRPC success, found %d entries", len(entries))

	dirEntries := make([]fuse.DirEntry, 0, len(entries))
	for _, info := range entries {
		mode := uint32(0o644)
		if info.IsDir {
			mode = 0o755
		}
		if info.IsSymlink {
			mode |= syscall.S_IFLNK
		} else if info.IsDir {
			mode |= syscall.S_IFDIR
		} else {
			mode |= syscall.S_IFREG
		}

		dirEntries = append(dirEntries, fuse.DirEntry{
			Name: info.Name,
			Mode: mode,
		})
	}

	return fs.NewListDirStream(dirEntries), 0
}

func (f *fuseFS) Lookup(ctx context.Context, name string, out *fuse.EntryOut) (*fs.Inode, syscall.Errno) {
	path := f.getPath(ctx)
	log.Printf("Lookup: path=%s, name=%s", path, name)
	if path == "" {
		log.Printf("Lookup: empty path, returning ENOENT")
		return nil, syscall.ENOENT
	}

	childPath := filepath.Join(path, name)
	log.Printf("Lookup: calling Stat for childPath=%s", childPath)
	info, err := f.client.Stat(ctx, childPath)
	if err != nil {
		log.Printf("Lookup: Stat failed for childPath=%s, error=%v", childPath, err)
		return nil, f.mapError(err)
	}
	log.Printf("Lookup: Stat success for childPath=%s, name=%s, isDir=%v", childPath, info.Name, info.IsDir)

	child := f.NewInode(ctx, &fuseFS{client: f.client, share: f.share, path: childPath}, fs.StableAttr{
		Mode: f.modeFromInfo(info),
		Ino:  f.hashIno(childPath),
	})

	f.fillAttr(info, &out.Attr)
	return child, 0
}

func (f *fuseFS) getPath(ctx context.Context) string {
	// Return the current path for this node
	if f.path == "" {
		return f.share // Root directory
	}
	return f.path
}

func (f *fuseFS) fillAttr(info *pb.FileInfo, out *fuse.Attr) {
	out.Size = uint64(info.Size)
	out.Mode = f.modeFromInfo(info)
	out.Uid = info.Uid
	out.Gid = info.Gid
	out.Atime = uint64(info.AccessTime)
	out.Mtime = uint64(info.ModTime)
	out.Ctime = uint64(info.ChangeTime)
	out.Nlink = 1
}

func (f *fuseFS) modeFromInfo(info *pb.FileInfo) uint32 {
	mode := info.Mode
	if info.IsDir {
		mode |= syscall.S_IFDIR
	} else if info.IsSymlink {
		mode |= syscall.S_IFLNK
	} else {
		mode |= syscall.S_IFREG
	}
	return mode
}

func (f *fuseFS) hashIno(path string) uint64 {
	// Simple hash for inode number
	hash := uint64(0)
	for _, c := range path {
		hash = hash*31 + uint64(c)
	}
	return hash
}

func (f *fuseFS) mapError(err error) syscall.Errno {
	// Map gRPC errors to syscall errors
	if err == nil {
		return 0
	}
	// For now, return generic I/O error
	// TODO: Parse gRPC error details for better mapping
	return syscall.EIO
}

type fuseFile struct {
	client *grpcClient
	handle int32
	mu     sync.Mutex
}

func (f *fuseFile) Read(ctx context.Context, dest []byte, off int64) (fuse.ReadResult, syscall.Errno) {
	f.mu.Lock()
	defer f.mu.Unlock()

	data, err := f.client.Read(ctx, f.handle, off, int32(len(dest)))
	if err != nil {
		return nil, f.mapError(err)
	}

	return fuse.ReadResultData(data), 0
}

func (f *fuseFile) Release(ctx context.Context) syscall.Errno {
	err := f.client.CloseHandle(ctx, f.handle)
	if err != nil {
		log.Printf("close handle error: %v", err)
		return f.mapError(err)
	}
	return 0
}

func (f *fuseFile) mapError(err error) syscall.Errno {
	if err == nil {
		return 0
	}
	// TODO: Better error mapping
	return syscall.EIO
}

// Prevent unused import warnings
var _ = time.Now
var _ = io.EOF
