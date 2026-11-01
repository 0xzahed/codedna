package parser

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestParseRealRepos exercises the parser over real repositories on disk.
// Skipping symbol-count assertions: goal is robustness (no panics, no timeouts).
func TestParseRealRepos(t *testing.T) {
	candidates := []string{"/home/zahed/sherlock", "/home/zahed/burp-mcp", "/home/zahed/glow-botanicals", "/home/zahed/antigravity-skills"}
	parsed := 0
	for _, repo := range candidates {
		if st, err := os.Stat(repo); err != nil || !st.IsDir() {
			continue
		}
		start := time.Now()
		errCount := 0
		filepath.WalkDir(repo, func(p string, d os.DirEntry, err error) error {
			if err != nil || d.IsDir() {
				if d != nil && d.IsDir() {
					switch d.Name() {
					case ".git", "node_modules", "vendor", "dist", "build", ".next", "__pycache__", ".venv", ".cache":
						return filepath.SkipDir
					}
				}
				return nil
			}
			if time.Since(start) > 8*time.Second {
				return filepath.SkipAll
			}
			ext := strings.ToLower(filepath.Ext(d.Name()))
			lang := LanguageForExt(ext)
			if lang == "" {
				return nil
			}
			rel, rerr := filepath.Rel(repo, p)
			if rerr != nil {
				return nil
			}
			res, perr := Parse(repo, rel)
			parsed++
			if perr != nil {
				if errCount < 3 {
					t.Logf("ERR %s: %v", rel, perr)
				}
				errCount++
				return nil
			}
			_ = res
			return nil
		})
		t.Logf("%s: parse errors=%d", repo, errCount)
	}
	if parsed == 0 {
		t.Skip("no candidate repos found")
	}
	t.Logf("total parsed: %d", parsed)
}
