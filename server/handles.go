package main

import (
    "os"
)

type fileHandle struct {
    id      int32
    absPath string
    file    *os.File
}

func (s *fileSystemServer) registerHandle(absPath string, f *os.File) int32 {
    s.mu.Lock()
    defer s.mu.Unlock()
    s.nextHandleID++
    id := s.nextHandleID
    s.handles[id] = &fileHandle{id: id, absPath: absPath, file: f}
    return id
}

func (s *fileSystemServer) getHandle(id int32) *fileHandle {
    s.mu.Lock()
    defer s.mu.Unlock()
    return s.handles[id]
}

func (s *fileSystemServer) takeHandle(id int32) *fileHandle {
    s.mu.Lock()
    defer s.mu.Unlock()
    h := s.handles[id]
    if h != nil {
        delete(s.handles, id)
    }
    return h
}


