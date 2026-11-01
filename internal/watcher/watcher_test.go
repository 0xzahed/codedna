package watcher

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestWatcherDetectsChanges(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "app.py")
	os.WriteFile(filePath, []byte("print(1)\n"), 0o644)

	var changedList []string
	changeCh := make(chan bool, 5)

	w := New(tmpDir, 50*time.Millisecond, func(changed []string) {
		changedList = append(changedList, changed...)
		changeCh <- true
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go w.Start(ctx)

	// Wait for watcher to initialize
	time.Sleep(100 * time.Millisecond)

	// Modify file
	time.Sleep(20 * time.Millisecond)
	os.WriteFile(filePath, []byte("print(2)\n"), 0o644)

	select {
	case <-changeCh:
		if len(changedList) == 0 {
			t.Fatal("Expected changed files")
		}
	case <-time.After(1 * time.Second):
		t.Fatal("Watcher timed out waiting for file modification event")
	}
}
