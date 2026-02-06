// ABOUTME: Atomic file operations for safe concurrent writes.
// ABOUTME: Provides AtomicWrite (tmp+rename) and EnsureDir helpers.
package mdstore

import (
	"os"
	"path/filepath"
)

// AtomicWrite writes data to path atomically via tmp file + rename.
// Creates parent directories if they don't exist.
func AtomicWrite(path string, data []byte) error {
	dir := filepath.Dir(path)
	if err := EnsureDir(dir); err != nil {
		return err
	}

	tmp, err := os.CreateTemp(dir, ".tmp-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()

	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return err
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpName)
		return err
	}

	return os.Rename(tmpName, path)
}

// EnsureDir creates a directory and all parents if they don't exist.
func EnsureDir(path string) error {
	return os.MkdirAll(path, 0o755)
}
