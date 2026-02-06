// ABOUTME: Unix implementation of WithLock using syscall.Flock (LOCK_EX).
// ABOUTME: Provides exclusive file locking for serializing writes on Unix systems.

//go:build !windows

package mdstore

import (
	"os"
	"path/filepath"
	"syscall"
)

// WithLock acquires an exclusive file lock on <dir>/.lock, executes fn, then releases.
// Only serializes writes — reads don't need locking.
func WithLock(dir string, fn func() error) error {
	lockPath := filepath.Join(dir, ".lock")

	if err := EnsureDir(dir); err != nil {
		return err
	}

	f, err := os.OpenFile(lockPath, os.O_CREATE|os.O_RDWR, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()

	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX); err != nil {
		return err
	}
	defer syscall.Flock(int(f.Fd()), syscall.LOCK_UN)

	return fn()
}
