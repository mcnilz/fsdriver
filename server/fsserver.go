package main

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	pb "github.com/example/fsdriver/proto"
)

type fileHandle struct {
	id      int32
	absPath string
	f       *os.File
}

type fileSystemServer struct {
	pb.UnimplementedFileSystemServiceServer
	root         string
	mu           sync.Mutex
	nextHandleID int32
	handles      map[int32]*fileHandle
}

func NewFileSystemServer(root string) (*fileSystemServer, error) {
	abs, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}
	return &fileSystemServer{root: abs, handles: make(map[int32]*fileHandle)}, nil
}

func (s *fileSystemServer) confine(rel string) (string, error) {
	// Normalize and ensure path stays within root
	cleaned := filepath.Clean(rel)
	joined := filepath.Join(s.root, cleaned)
	abs, err := filepath.Abs(joined)
	if err != nil {
		return "", err
	}
	if abs != s.root && !isSubpath(abs, s.root) {
		return "", errors.New("path escapes root")
	}
	return abs, nil
}

func isSubpath(path string, root string) bool {
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return false
	}
	return rel != ".." && !startsWithDotDot(rel)
}

func startsWithDotDot(rel string) bool {
	return rel == ".." || len(rel) > 3 && rel[:3] == "..\\" || len(rel) > 3 && rel[:3] == "../"
}

func (s *fileSystemServer) toFileInfo(fi os.FileInfo) *pb.FileInfo {
	mode := uint32(fi.Mode().Perm())
	return &pb.FileInfo{
		Name:       fi.Name(),
		IsDir:      fi.IsDir(),
		Size:       fi.Size(),
		ModTime:    fi.ModTime().Unix(),
		AccessTime: fi.ModTime().Unix(),
		ChangeTime: fi.ModTime().Unix(),
		Mode:       mode,
		Uid:        0,
		Gid:        0,
		IsSymlink:  fi.Mode()&os.ModeSymlink != 0,
	}
}

func (s *fileSystemServer) Stat(ctx context.Context, req *pb.StatRequest) (*pb.StatResponse, error) {
	abs, err := s.confine(req.Path)
	if err != nil {
		return &pb.StatResponse{Result: &pb.StatResponse_Error{Error: errno(err)}}, nil
	}
	fi, err := os.Lstat(abs)
	if err != nil {
		return &pb.StatResponse{Result: &pb.StatResponse_Error{Error: errno(err)}}, nil
	}
	return &pb.StatResponse{Result: &pb.StatResponse_Info{Info: s.toFileInfo(fi)}}, nil
}

func (s *fileSystemServer) ReadDir(ctx context.Context, req *pb.ReadDirRequest) (*pb.ReadDirResponse, error) {
	abs, err := s.confine(req.Path)
	if err != nil {
		return &pb.ReadDirResponse{Error: errno(err)}, nil
	}
	f, err := os.Open(abs)
	if err != nil {
		return &pb.ReadDirResponse{Error: errno(err)}, nil
	}
	defer f.Close()

	// Pagination
	offset := int(req.Offset)
	limit := int(req.Limit)
	if limit < 0 {
		limit = 0
	}
	entries, err := f.Readdir(0)
	if err != nil && err != io.EOF {
		return &pb.ReadDirResponse{Error: errno(err)}, nil
	}
	var out []*pb.FileInfo
	for i := offset; i < len(entries); i++ {
		if limit > 0 && len(out) >= limit {
			break
		}
		out = append(out, s.toFileInfo(entries[i]))
	}
	hasMore := (offset+len(out) < len(entries))
	return &pb.ReadDirResponse{Entries: out, HasMore: hasMore}, nil
}

func (s *fileSystemServer) Open(ctx context.Context, req *pb.OpenRequest) (*pb.OpenResponse, error) {
	abs, err := s.confine(req.Path)
	if err != nil {
		return &pb.OpenResponse{Result: &pb.OpenResponse_Error{Error: errno(err)}}, nil
	}
	f, err := os.Open(abs)
	if err != nil {
		return &pb.OpenResponse{Result: &pb.OpenResponse_Error{Error: errno(err)}}, nil
	}
	s.mu.Lock()
	s.nextHandleID++
	hid := s.nextHandleID
	s.handles[hid] = &fileHandle{id: hid, absPath: abs, f: f}
	s.mu.Unlock()
	return &pb.OpenResponse{Result: &pb.OpenResponse_Handle{Handle: hid}}, nil
}

func (s *fileSystemServer) Read(ctx context.Context, req *pb.ReadRequest) (*pb.ReadResponse, error) {
	s.mu.Lock()
	h := s.handles[req.Handle]
	s.mu.Unlock()
	if h == nil {
		return &pb.ReadResponse{Result: &pb.ReadResponse_Error{Error: &pb.Error{Code: int32(2), Message: "bad handle"}}}, nil
	}
	if req.Offset < 0 || req.Size < 0 {
		return &pb.ReadResponse{Result: &pb.ReadResponse_Error{Error: &pb.Error{Code: int32(22), Message: "invalid offset/size"}}}, nil
	}
	if _, err := h.f.Seek(req.Offset, io.SeekStart); err != nil {
		return &pb.ReadResponse{Result: &pb.ReadResponse_Error{Error: errno(err)}}, nil
	}
	buf := make([]byte, req.Size)
	n, err := io.ReadFull(h.f, buf)
	if err == io.ErrUnexpectedEOF || err == io.EOF {
		// partial read at EOF is fine
		return &pb.ReadResponse{Result: &pb.ReadResponse_Data{Data: buf[:n]}}, nil
	}
	if err != nil {
		return &pb.ReadResponse{Result: &pb.ReadResponse_Error{Error: errno(err)}}, nil
	}
	return &pb.ReadResponse{Result: &pb.ReadResponse_Data{Data: buf[:n]}}, nil
}

func (s *fileSystemServer) Close(ctx context.Context, req *pb.CloseRequest) (*pb.CloseResponse, error) {
	s.mu.Lock()
	h := s.handles[req.Handle]
	if h != nil {
		delete(s.handles, req.Handle)
	}
	s.mu.Unlock()
	if h == nil {
		return &pb.CloseResponse{Error: &pb.Error{Code: int32(2), Message: "bad handle"}}, nil
	}
	_ = h.f.Close()
	return &pb.CloseResponse{}, nil
}

// errno maps generic errors to POSIX-like codes; improve with Win32 codes later.
func errno(err error) *pb.Error {
	if errors.Is(err, os.ErrNotExist) {
		return &pb.Error{Code: 2, Message: err.Error()} // ENOENT
	}
	if errors.Is(err, os.ErrPermission) {
		return &pb.Error{Code: 13, Message: err.Error()} // EACCES
	}
	// default EIO
	return &pb.Error{Code: 5, Message: err.Error()}
}

// prevent unused imports warnings during early development
var _ = time.Now
