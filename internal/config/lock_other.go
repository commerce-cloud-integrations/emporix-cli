//go:build !(darwin || dragonfly || freebsd || linux || netbsd || openbsd || windows)

package config

import (
	"errors"
	"fmt"
	"os"
)

// tryLock fails: epx locks files only where the system offers it a lock
// that ends with the process (flock, or LockFileEx on Windows).
func tryLock(*os.File) (bool, error) {
	return false, fmt.Errorf("locking the configuration file: %w", errors.ErrUnsupported)
}

// unlockFile releases no lock, as tryLock takes none.
func unlockFile(*os.File) error { return nil }
