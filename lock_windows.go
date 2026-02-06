// ABOUTME: Windows implementation of WithLock using O_CREATE|O_EXCL retry loop.
// ABOUTME: Provides exclusive file locking with stale lock detection for Windows systems.

//go:build windows

package mdstore

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const (
	lockRetryInterval = 50 * time.Millisecond
	lockTimeout       = 10 * time.Second
	staleLockAge      = 30 * time.Second
)

// WithLock acquires an exclusive file lock on <dir>/.lock, executes fn, then releases.
// Uses O_CREATE|O_EXCL retry loop with stale lock detection on Windows.
func WithLock(dir string, fn func() error) error {
	lockPath := filepath.Join(dir, ".lock")

	if err := EnsureDir(dir); err != nil {
		return err
	}

	deadline := time.Now().Add(lockTimeout)

	for {
		f, err := os.OpenFile(lockPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
		if err == nil {
			// Lock acquired
			f.Close()
			defer os.Remove(lockPath)
			return fn()
		}

		// Check for stale lock
		info, statErr := os.Stat(lockPath)
		if statErr == nil && time.Since(info.ModTime()) > staleLockAge {
			os.Remove(lockPath)
			continue
		}

		if time.Now().After(deadline) {
			return fmt.Errorf("mdstore: lock timeout after %v on %s", lockTimeout, lockPath)
		}

		time.Sleep(lockRetryInterval)
	}
}
