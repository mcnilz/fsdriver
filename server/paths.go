package main

import (
    "errors"
    "path/filepath"
)

func normalizeWithinRoot(root string, rel string) (string, error) {
    cleaned := filepath.Clean(rel)
    joined := filepath.Join(root, cleaned)
    abs, err := filepath.Abs(joined)
    if err != nil {
        return "", err
    }
    if abs != root && !isSubpath(abs, root) {
        return "", errors.New("path escapes root")
    }
    return abs, nil
}

func isSubpath(path string, root string) bool {
    rel, err := filepath.Rel(root, path)
    if err != nil {
        return false
    }
    if rel == "." {
        return true
    }
    return rel != ".." && !startsWithDotDot(rel)
}

func startsWithDotDot(rel string) bool {
    return rel == ".." || (len(rel) >= 3 && (rel[:3] == "..\\" || rel[:3] == "../"))
}


