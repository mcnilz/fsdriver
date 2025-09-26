package main

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"sync"

	pb "github.com/example/fsdriver/proto"
)

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
	return normalizeWithinRoot(s.root, rel)
}

func (s *fileSystemServer) toFileInfo(fi os.FileInfo) *pb.FileInfo {
	mode := uint32(fi.Mode().Perm())
	ts := fi.ModTime().Unix()
	return &pb.FileInfo{
		Name:       fi.Name(),
		IsDir:      fi.IsDir(),
		Size:       fi.Size(),
		ModTime:    ts,
		AccessTime: ts,
		ChangeTime: ts,
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
	logx.Info("ReadDir request", "path", req.Path, "offset", req.Offset, "limit", req.Limit)
	abs, err := s.confine(req.Path)
	if err != nil {
		logx.Error("ReadDir path confinement failed", "path", req.Path, "error", err)
		return &pb.ReadDirResponse{Error: errno(err)}, nil
	}
	logx.Info("ReadDir confined path", "original", req.Path, "absolute", abs)
	f, err := os.Open(abs)
	if err != nil {
		logx.Error("ReadDir open failed", "path", abs, "error", err)
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
	logx.Info("ReadDir response", "entries_returned", len(out), "total_entries", len(entries), "has_more", hasMore)
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
	hid := s.registerHandle(abs, f)
	return &pb.OpenResponse{Result: &pb.OpenResponse_Handle{Handle: hid}}, nil
}

func (s *fileSystemServer) Read(ctx context.Context, req *pb.ReadRequest) (*pb.ReadResponse, error) {
	h := s.getHandle(req.Handle)
	if h == nil {
		return &pb.ReadResponse{Result: &pb.ReadResponse_Error{Error: &pb.Error{Code: int32(2), Message: "bad handle"}}}, nil
	}
	if req.Offset < 0 || req.Size < 0 {
		return &pb.ReadResponse{Result: &pb.ReadResponse_Error{Error: &pb.Error{Code: int32(22), Message: "invalid offset/size"}}}, nil
	}
	if _, err := h.file.Seek(req.Offset, io.SeekStart); err != nil {
		return &pb.ReadResponse{Result: &pb.ReadResponse_Error{Error: errno(err)}}, nil
	}
	buf := make([]byte, req.Size)
	n, err := io.ReadFull(h.file, buf)
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
	h := s.takeHandle(req.Handle)
	if h == nil {
		return &pb.CloseResponse{Error: &pb.Error{Code: int32(2), Message: "bad handle"}}, nil
	}
	_ = h.file.Close()
	return &pb.CloseResponse{}, nil
}
