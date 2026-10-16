package config

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// lockPoll is how often Lock tries again to lock a configuration file that
// another epx process has locked.
const lockPoll = 50 * time.Millisecond

// Lock locks the configuration file against every other epx process that
// changes it, waiting until the one that holds the lock releases it or ctx
// ends; unlock releases it. The lock is held on a file of its own,
// epx/config.lock, as Save replaces the configuration file whole. The
// system releases it when the process ends, however it ends.
func (s Store) Lock(ctx context.Context) (unlock func(), err error) {
	err = os.MkdirAll(s.dir, 0o700)
	if err != nil {
		return nil, err
	}
	lock, err := os.OpenFile(filepath.Join(s.dir, "config.lock"), os.O_RDWR|os.O_CREATE, 0o600)
	if err != nil {
		return nil, err
	}
	err = waitForLock(ctx, lock)
	if err != nil {
		_ = lock.Close() // only locked
		return nil, err
	}
	return func() {
		_ = unlockFile(lock) // closing the file releases the lock too
		_ = lock.Close()
	}, nil
}

// waitForLock locks file, trying again every lockPoll until ctx ends.
func waitForLock(ctx context.Context, file *os.File) error {
	retry := time.NewTicker(lockPoll)
	defer retry.Stop()
	for {
		locked, err := tryLock(file)
		if locked || err != nil {
			return err
		}
		select {
		case <-ctx.Done():
			return fmt.Errorf("waiting for another epx to finish changing profiles: %w", context.Cause(ctx))
		case <-retry.C:
		}
	}
}
