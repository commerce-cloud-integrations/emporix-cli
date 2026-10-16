package config

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

func TestLockKeepsOtherChangesWaiting(t *testing.T) {
	store := In(t.TempDir())
	unlock, err := store.Lock(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	// Another lock waits, until its context ends.
	ctx, cancel := context.WithTimeout(context.Background(), 3*lockPoll)
	defer cancel()
	if _, waited := store.Lock(ctx); !errors.Is(waited, context.DeadlineExceeded) {
		t.Fatalf("second lock: %v; want it to wait until its deadline", waited)
	}
	// One waiting gets the lock once it is released.
	locked := make(chan error, 1)
	go func() {
		again, failed := store.Lock(context.Background())
		if failed == nil {
			again()
		}
		locked <- failed
	}()
	select {
	case early := <-locked:
		t.Fatalf("lock while held: %v", early)
	case <-time.After(2 * lockPoll):
	}
	unlock()
	if late := <-locked; late != nil {
		t.Fatalf("lock once released: %v", late)
	}
	info, err := os.Stat(filepath.Join(filepath.Dir(store.Path()), "config.lock"))
	if err != nil || (info.Mode().Perm()&0o077 != 0 && runtime.GOOS != "windows") {
		t.Errorf("lock file %v, %v; want one only the user may read", info, err)
	}
}

func TestLockFailsOnAnUnusableDirectory(t *testing.T) {
	blocked := filepath.Join(t.TempDir(), "file")
	if err := os.WriteFile(blocked, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := In(blocked).Lock(context.Background()); err == nil {
		t.Error("lock below a file succeeded")
	}
	if runtime.GOOS == "windows" {
		return // a read-only directory takes new files on Windows
	}
	store := In(t.TempDir())
	readOnly := filepath.Dir(store.Path())
	if err := os.Mkdir(readOnly, 0o500); err != nil {
		t.Fatal(err)
	}
	if _, err := store.Lock(context.Background()); err == nil {
		t.Error("lock in a directory without the lock file, where none can be made, succeeded")
	}
}
