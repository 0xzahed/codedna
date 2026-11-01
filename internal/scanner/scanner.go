// Package scanner walks a repository, detects languages, frameworks and
// dependencies, and extracts git metadata. It never executes repository code.
package scanner

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/codedna/codedna/internal/model"
)

// DefaultIgnore directories (matched against a path segment).
var DefaultIgnore = map[string]bool{
	".git": true, "node_modules": true, "vendor": true, "dist": true,
	"build": true, "target": true, ".venv": true, "venv": true,
	"__pycache__": true, ".idea": true, ".vscode": true, ".cache": true,
	"coverage": true, ".next": true, ".nuxt": true, ".tox": true,
	".mypy_cache": true, ".pytest_cache": true, ".gradle": true,
	"out": true, "bin": false, ".terraform": true, "site-packages": true,
	"static_cdn": true, "static_serve": true, "staticfiles": true,
}

// language detection mapping: extension -> canonical language name.
var extLangs = map[string]string{
	".go": "go", ".py": "python", ".pyi": "python", ".js": "javascript",
	".mjs": "javascript", ".cjs": "javascript", ".jsx": "javascript",
	".ts": "typescript", ".mts": "typescript", ".cts": "typescript", ".tsx": "typescript",
	".java": "java", ".php": "php",
}

// ParsedLanguages are the languages this build of CodeDNA can parse with Tree-sitter.
func ParsedLanguages() map[string]bool {
	return map[string]bool{"go": true, "python": true, "javascript": true, "typescript": true, "java": true, "php": true}
}

// File language classification. Returns lang, isTest, accessErr-nil if skip.
type walkStats struct {
	skipped int
}

// Scanner holds scan configuration.
type Scanner struct {
	Ignore map[string]bool
	// MaxFileSize halts reading larger files (bytes); protects memory.
	MaxFileSize int64
}

// New creates a Scanner with default ignores.
func New() *Scanner {
	return &Scanner{Ignore: DefaultIgnore, MaxFileSize: 2 << 20} // 2 MiB
}

// shouldSkip reports whether a relative path segment list should be ignored.
func (s *Scanner) shouldSkip(parts []string) bool {
	for _, p := range parts {
		if strings.HasPrefix(p, ".codedna") || strings.HasPrefix(p, ".codedna-") {
			return true
		}
		if s.Ignore[p] {
			return true
		}
	}
	// node_modules nested: any segment in Ignore set already covered above.
	return false
}

// NodePackage uses composition to count things.
type ScanProgress func(done int, total int, path string)

// ScanResult is everything the scanner discovers about a repository.
type ScanResult struct {
	Root      string             `json:"root"`
	Name      string             `json:"name"`
	Git       model.GitMeta      `json:"git"`
	Files     []model.FileNode   `json:"files"`
	Deps      map[string][]string `json:"dependencies"`
	Frameworks []string          `json:"frameworks"`
	Languages map[string]int     `json:"languages"`
	LOCByLang map[string]int     `json:"locByLanguage"`
	TotalLOC  int                `json:"totalLoc"`
	SkippedFiles int              `json:"skippedFiles"`
	CfgFiles  map[string]string  `json:"-"` // manifest file paths per ecosystem
}

// ResolveTarget resolves a local folder path (supporting ~/) or a remote GitHub/Git repository URL.
// Remote repositories are cloned into ~/.codedna/repos/<slug> with depth 1.
func ResolveTarget(target string) (string, error) {
	target = strings.TrimSpace(target)
	target = strings.Trim(target, "\"'`")
	target = strings.TrimSpace(target)
	if target == "" {
		target = "."
	}

	isRemote := strings.HasPrefix(target, "https://") ||
		strings.HasPrefix(target, "http://") ||
		strings.HasPrefix(target, "git@") ||
		strings.HasPrefix(target, "github.com/") ||
		strings.HasPrefix(target, "gitlab.com/")

	if isRemote {
		cloneURL := target
		if strings.HasPrefix(target, "github.com/") || strings.HasPrefix(target, "gitlab.com/") {
			cloneURL = "https://" + target
		}

		// Clean slug for directory name, e.g. "https://github.com/gin-gonic/gin.git" -> "gin-gonic_gin"
		u := cloneURL
		u = strings.TrimSuffix(u, ".git")
		u = strings.TrimPrefix(u, "https://")
		u = strings.TrimPrefix(u, "http://")
		u = strings.TrimPrefix(u, "git@")
		u = strings.ReplaceAll(u, ":", "/")
		parts := strings.Split(u, "/")
		slug := "repo"
		if len(parts) >= 2 {
			slug = parts[len(parts)-2] + "_" + parts[len(parts)-1]
		} else if len(parts) == 1 {
			slug = parts[0]
		}

		home, err := os.UserHomeDir()
		if err != nil {
			home = os.TempDir()
		}
		destDir := filepath.Join(home, ".codedna", "repos", slug)

		if fi, err := os.Stat(destDir); err == nil && fi.IsDir() {
			// Already cloned, try fast pull or reuse existing
			cmd := exec.Command("git", "-C", destDir, "pull", "--ff-only")
			_ = cmd.Run()
			return destDir, nil
		}

		if err := os.MkdirAll(filepath.Dir(destDir), 0o755); err != nil {
			return "", fmt.Errorf("create repos directory: %w", err)
		}

		cmd := exec.Command("git", "clone", "--depth", "1", cloneURL, destDir)
		var errBuf strings.Builder
		cmd.Stderr = &errBuf
		if err := cmd.Run(); err != nil {
			return "", fmt.Errorf("git clone failed for %s: %s", cloneURL, errBuf.String())
		}
		return destDir, nil
	}

	// Expand ~/ if provided
	if strings.HasPrefix(target, "~/") {
		home, err := os.UserHomeDir()
		if err == nil {
			target = filepath.Join(home, target[2:])
		}
	}

	absRoot, err := filepath.Abs(target)
	if err != nil {
		return "", fmt.Errorf("resolve path: %w", err)
	}

	info, err := os.Stat(absRoot)
	if err != nil {
		return "", fmt.Errorf("target path does not exist: %w", err)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("target path %s is not a directory", absRoot)
	}

	return absRoot, nil
}

// Scan walks root and produces a ScanResult. It does not parse code.
func (s *Scanner) Scan(root string, progress ScanProgress) (*ScanResult, error) {
	absRoot, err := ResolveTarget(root)
	if err != nil {
		return nil, err
	}

	res := &ScanResult{
		Root:       absRoot,
		Name:       filepath.Base(absRoot),
		Deps:       map[string][]string{},
		Frameworks: []string{},
		Languages:  map[string]int{},
		LOCByLang:  map[string]int{},
		CfgFiles:   map[string]string{},
	}

	git, err := detectGit(absRoot)
	if err != nil {
		return nil, err
	}
	res.Git = git

	count := 0
	err = filepath.WalkDir(absRoot, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil // unreadable entries don't abort the scan
		}
		rel, _ := filepath.Rel(absRoot, path)
		parts := strings.Split(rel, string(filepath.Separator))
		if d.IsDir() {
			if s.shouldSkip(parts) {
				return filepath.SkipDir
			}
			return nil
		}
		if s.shouldSkip(parts) {
			return nil
		}
		base := filepath.Base(path)
		if strings.HasSuffix(base, ".min.js") || strings.HasSuffix(base, ".min.css") {
			return nil
		}
		ext := strings.ToLower(filepath.Ext(base))
		size := int64(0)
		if fi, fierr := d.Info(); fierr == nil {
			size = fi.Size()
		}

		// Language classification by extension first.
		lang, isLang := extLangs[ext]
		isConfig := base == "go.mod" || base == "go.sum" || base == "requirements.txt" ||
			base == "pyproject.toml" || base == "setup.py" || base == "package.json" ||
			base == "composer.json" || base == "pom.xml" || base == "build.gradle" || base == "Cargo.toml"
		if !isLang && !isConfig {
			return nil
		}
		if size > s.MaxFileSize {
			res.SkippedFiles++
			return nil
		}

		// Tags: isTest quick heuristics.
		tags := []string{}
		lowerRel := strings.ToLower(rel)
		if strings.Contains(lowerRel, "_test") || strings.Contains(lowerRel, "test") &&
			(strings.HasPrefix(base, "test_") || strings.HasSuffix(lowerRel, "_test.go") ||
				strings.HasSuffix(base, "_test.py") || strings.HasSuffix(base, ".test.js") ||
				strings.HasSuffix(base, ".test.ts") || strings.HasSuffix(base, ".spec.js") ||
				strings.HasSuffix(base, ".spec.ts") || strings.Contains(lowerRel, "/tests/") ||
				strings.Contains(lowerRel, "/test/")) {
			tags = append(tags, "test")
		}

		loc := 0
		if isLang {
			f, ferr := os.Open(path)
			if ferr == nil {
				loc = countLines(f)
				f.Close()
			}
		}
		fn := model.FileNode{
			Path:     filepath.ToSlash(rel),
			Size:     size,
			LOC:      loc,
			Tags:     tags,
		}
		if isConfig {
			lang = manifestEcosystem(base)
			fn.Tags = append(fn.Tags, "manifest")
			res.CfgFiles[base] = fn.Path
		} else {
			res.Languages[lang]++
			res.LOCByLang[lang] += loc
			res.TotalLOC += loc
		}
		fn.Language = lang
		res.Files = append(res.Files, fn)

		count++
		if progress != nil && count%25 == 0 {
			progress(count, 0, fn.Path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	sort.Slice(res.Files, func(i, j int) bool { return res.Files[i].Path < res.Files[j].Path })

	res.Deps = extractDependencies(absRoot, res.CfgFiles)
	res.Frameworks = detectFrameworks(absRoot, res.CfgFiles, res.Languages)

	return res, nil
}

func countLines(f *os.File) int {
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 64*1024), 1<<20)
	n := 0
	for sc.Scan() {
		n++
	}
	return n
}

func manifestEcosystem(base string) string {
	switch base {
	case "go.mod", "go.sum":
		return "go"
	case "requirements.txt", "pyproject.toml", "setup.py":
		return "python"
	case "package.json":
		return "javascript"
	case "composer.json":
		return "php"
	case "pom.xml", "build.gradle":
		return "java"
	}
	return "config"
}
