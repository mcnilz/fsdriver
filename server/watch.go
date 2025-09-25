package main

import (
    "context"
    "io/fs"
    "path/filepath"
    "strings"
    "sync"
    "time"

    "github.com/fsnotify/fsnotify"

    pb "github.com/example/fsdriver/proto"
)

// Watch implements the bidirectional stream for change notifications.
func (s *fileSystemServer) Watch(stream pb.FileSystemService_WatchServer) error {
    // One watcher per stream
    watcher, err := fsnotify.NewWatcher()
    if err != nil {
        return err
    }
    defer watcher.Close()

    ctx, cancel := context.WithCancel(stream.Context())
    defer cancel()

    var mu sync.Mutex
    watched := make(map[string]struct{})

    addPath := func(rel string, recursive bool) error {
        abs, err := s.confine(rel)
        if err != nil {
            return err
        }
        // Add directory itself
        if _, ok := watched[abs]; !ok {
            if err := watcher.Add(abs); err != nil {
                return err
            }
            watched[abs] = struct{}{}
        }
        if !recursive {
            return nil
        }
        // Walk subdirectories using WalkDir
        return filepath.WalkDir(abs, func(p string, d fs.DirEntry, walkErr error) error {
            if walkErr != nil {
                return nil
            }
            if !d.IsDir() {
                return nil
            }
            if _, ok := watched[p]; ok {
                return nil
            }
            if err := watcher.Add(p); err == nil {
                watched[p] = struct{}{}
            }
            return nil
        })
    }

    // Receiver goroutine: accept subscription requests
    recvErrCh := make(chan error, 1)
    go func() {
        defer close(recvErrCh)
        for {
            req, err := stream.Recv()
            if err != nil {
                recvErrCh <- err
                return
            }
            mu.Lock()
            e := addPath(req.Path, req.Recursive)
            mu.Unlock()
            if e != nil {
                // Report an ATTRIB with error context using old_path field for details
                _ = stream.Send(&pb.WatchEvent{
                    Path:      sanitizeRel(req.Path),
                    Type:      pb.WatchEventType_ATTRIB,
                    OldPath:   "error: " + e.Error(),
                    Timestamp: time.Now().Unix(),
                })
            }
        }
    }()

    // Event loop
    for {
        select {
        case <-ctx.Done():
            return ctx.Err()
        case err := <-recvErrCh:
            return err
        case ev := <-watcher.Events:
            // Map event
            evtType := mapFsnotifyEvent(ev)
            if evtType == pb.WatchEventType_UNKNOWN {
                continue
            }
            // Compute path relative to share root
            rel := ev.Name
            if strings.HasPrefix(rel, s.root) {
                if r, err := filepath.Rel(s.root, rel); err == nil {
                    rel = filepath.ToSlash(r)
                }
            }
            _ = stream.Send(&pb.WatchEvent{
                Path:      rel,
                Type:      evtType,
                Timestamp: time.Now().Unix(),
            })
            // On create of directory when recursive, attempt to add watcher
            if (ev.Op&fsnotify.Create) != 0 {
                // Best-effort add
                mu.Lock()
                _ = addPath(rel, true)
                mu.Unlock()
            }
        case err := <-watcher.Errors:
            // Surface watcher errors as ATTRIB with details
            _ = stream.Send(&pb.WatchEvent{
                Path:      "",
                Type:      pb.WatchEventType_ATTRIB,
                OldPath:   "watch-error: " + err.Error(),
                Timestamp: time.Now().Unix(),
            })
        }
    }
}

func mapFsnotifyEvent(ev fsnotify.Event) pb.WatchEventType {
    switch {
    case ev.Op&fsnotify.Create != 0:
        return pb.WatchEventType_CREATE
    case ev.Op&fsnotify.Remove != 0:
        return pb.WatchEventType_DELETE
    case ev.Op&fsnotify.Rename != 0:
        return pb.WatchEventType_RENAME
    case ev.Op&(fsnotify.Write|fsnotify.Chmod) != 0:
        return pb.WatchEventType_MODIFY
    default:
        return pb.WatchEventType_UNKNOWN
    }
}

func sanitizeRel(p string) string {
    return filepath.ToSlash(filepath.Clean(p))
}


