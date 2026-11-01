// Package watcher monitors a directory for file modifications and triggers rescan events.
package watcher

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/codedna/codedna/internal/scanner"
)

// ChangeCallback is invoked when files are modified.
type ChangeCallback func(changedFiles []string)

// Watcher polls file modification times with debouncing and safe ignore filtering.
type Watcher struct {
	root     string
	interval time.Duration
	cb       ChangeCallback
	ignore   map[string]bool
	mu       sync.Mutex
	mtimes   map[string]time.Time
}

// New creates a Watcher.
func New(root string, interval time.Duration, cb ChangeCallback) *Watcher {
	if interval <= 0 {
		interval = 500 * time.Millisecond
	}
	return &Watcher{
		root:     root,
		interval: interval,
		cb:       cb,
		ignore:   scanner.DefaultIgnore,
		mtimes:   make(map[string]time.Time),
	}
}

// Start begins watching until context is canceled.
func (w *Watcher) Start(ctx context.Context) {
	// Initial snapshot of mtimes
	w.scanSnapshot()

	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			changed := w.checkChanges()
			if len(changed) > 0 && w.cb != nil {
				w.cb(changed)
			}
		}
	}
}

func (w *Watcher) scanSnapshot() {
	w.mu.Lock()
	defer w.mu.Unlock()

	filepath.WalkDir(w.root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		rel, _ := filepath.Rel(w.root, path)
		parts := strings.Split(rel, string(filepath.Separator))
		for _, p := range parts {
			if w.ignore[p] || strings.HasPrefix(p, ".codedna") {
				if d.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
		}
		if d.IsDir() {
			return nil
		}
		if fi, err := d.Info(); err == nil {
			w.mtimes[rel] = fi.ModTime()
		}
		return nil
	})
}

func (w *Watcher) checkChanges() []string {
	w.mu.Lock()
	defer w.mu.Unlock()

	var changed []string
	currentFiles := make(map[string]bool)

	filepath.WalkDir(w.root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		rel, _ := filepath.Rel(w.root, path)
		parts := strings.Split(rel, string(filepath.Separator))
		for _, p := range parts {
			if w.ignore[p] || strings.HasPrefix(p, ".codedna") {
				if d.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
		}
		if d.IsDir() {
			return nil
		}

		currentFiles[rel] = true
		fi, err := d.Info()
		if err != nil {
			return nil
		}

		oldTime, exists := w.mtimes[rel]
		if !exists || fi.ModTime().After(oldTime) {
			changed = append(changed, rel)
			w.mtimes[rel] = fi.ModTime()
		}
		return nil
	})

	// Check for deleted files
	for file := range w.mtimes {
		if !currentFiles[file] {
			changed = append(changed, file)
			delete(w.mtimes, file)
		}
	}

	return changed
}
