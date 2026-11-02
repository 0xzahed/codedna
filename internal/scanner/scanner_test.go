package scanner

import (
	"os"
	"path/filepath"
	"testing"
)

func TestScannerBasic(t *testing.T) {
	tmpDir := t.TempDir()

	// Create sample project structure
	err := os.MkdirAll(filepath.Join(tmpDir, "src", "api"), 0o755)
	if err != nil {
		t.Fatal(err)
	}
	err = os.MkdirAll(filepath.Join(tmpDir, "node_modules", "badpkg"), 0o755)
	if err != nil {
		t.Fatal(err)
	}

	// Create Go file
	err = os.WriteFile(filepath.Join(tmpDir, "src", "api", "server.go"), []byte("package api\n\nfunc Start() {}\n"), 0o644)
	if err != nil {
		t.Fatal(err)
	}

	// Create Python file
	err = os.WriteFile(filepath.Join(tmpDir, "main.py"), []byte("def run():\n    print('hello')\n"), 0o644)
	if err != nil {
		t.Fatal(err)
	}

	// Create ignored file in node_modules
	err = os.WriteFile(filepath.Join(tmpDir, "node_modules", "badpkg", "index.js"), []byte("console.log('skip');"), 0o644)
	if err != nil {
		t.Fatal(err)
	}

	// Create package.json
	pkgJSON := `{
  "name": "sample-project",
  "dependencies": {
    "react": "^18.2.0",
    "express": "^4.18.2"
  }
}`
	err = os.WriteFile(filepath.Join(tmpDir, "package.json"), []byte(pkgJSON), 0o644)
	if err != nil {
		t.Fatal(err)
	}

	// Create requirements.txt
	reqs := "fastapi>=0.100.0\nuvicorn\nrequests==2.31.0\n"
	err = os.WriteFile(filepath.Join(tmpDir, "requirements.txt"), []byte(reqs), 0o644)
	if err != nil {
		t.Fatal(err)
	}

	sc := New()
	res, err := sc.Scan(tmpDir, nil)
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	if res.Name == "" {
		t.Error("Expected repository name")
	}

	// Check languages
	if res.Languages["go"] != 1 {
		t.Errorf("Expected 1 Go file, got %d", res.Languages["go"])
	}
	if res.Languages["python"] != 1 {
		t.Errorf("Expected 1 Python file, got %d", res.Languages["python"])
	}
	if res.Languages["javascript"] != 0 {
		t.Errorf("Expected node_modules JS file to be skipped, got %d", res.Languages["javascript"])
	}

	// Check dependencies
	jsDeps := res.Deps["javascript"]
	if len(jsDeps) < 2 {
		t.Errorf("Expected at least 2 JS deps, got %v", jsDeps)
	}

	pyDeps := res.Deps["python"]
	if len(pyDeps) < 3 {
		t.Errorf("Expected at least 3 Py deps, got %v", pyDeps)
	}

	// Check frameworks
	hasExpress := false
	hasFastAPI := false
	for _, fw := range res.Frameworks {
		if fw == "Express" {
			hasExpress = true
		}
		if fw == "FastAPI" {
			hasFastAPI = true
		}
	}
	if !hasExpress {
		t.Errorf("Expected Express framework to be detected, got %v", res.Frameworks)
	}
	if !hasFastAPI {
		t.Errorf("Expected FastAPI framework to be detected, got %v", res.Frameworks)
	}
}

func TestGitDetection(t *testing.T) {
	// Detect git on current repo
	meta, err := detectGit(".")
	if err != nil {
		t.Fatalf("detectGit failed: %v", err)
	}
	if !meta.IsGitRepo {
		t.Error("Expected current directory to be detected as git repo")
	}
	if meta.Branch != "main" {
		t.Errorf("Expected branch main, got %s", meta.Branch)
	}
}

func TestResolveTarget(t *testing.T) {
	// 1. Current dir
	p, err := ResolveTarget(".")
	if err != nil {
		t.Fatalf("ResolveTarget(.) failed: %v", err)
	}
	if p == "" {
		t.Error("Expected non-empty resolved path")
	}

	// 2. Non-existent path returns error
	_, err = ResolveTarget("/non/existent/path/for/sure/codedna")
	if err == nil {
		t.Error("Expected error for non-existent path")
	}

	// 3. Temp dir
	tmp := t.TempDir()
	p, err = ResolveTarget(tmp)
	if err != nil {
		t.Fatalf("ResolveTarget(tmp) failed: %v", err)
	}
	if p != tmp {
		t.Errorf("Expected %s, got %s", tmp, p)
	}

	// 4. Home directory protection
	if home, herr := os.UserHomeDir(); herr == nil && home != "" {
		_, err = ResolveTarget(home)
		if err == nil {
			t.Error("Expected error when resolving home directory directly")
		}
	}
}

func TestGitignoreAndBinaryFiltering(t *testing.T) {
	tmpDir := t.TempDir()

	// Create files
	_ = os.WriteFile(filepath.Join(tmpDir, ".gitignore"), []byte("ignored_dir/\n*.secret.go\n"), 0o644)
	_ = os.MkdirAll(filepath.Join(tmpDir, "ignored_dir"), 0o755)
	_ = os.WriteFile(filepath.Join(tmpDir, "ignored_dir", "bad.go"), []byte("package bad\n"), 0o644)
	_ = os.WriteFile(filepath.Join(tmpDir, "app.secret.go"), []byte("package secret\n"), 0o644)
	_ = os.WriteFile(filepath.Join(tmpDir, "main.go"), []byte("package main\n"), 0o644)
	_ = os.WriteFile(filepath.Join(tmpDir, "app.exe"), []byte("MZbinarydata"), 0o644)

	sc := New()
	res, err := sc.Scan(tmpDir, nil)
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	for _, f := range res.Files {
		if filepath.Base(f.Path) == "bad.go" {
			t.Errorf("Expected bad.go in ignored_dir to be skipped by .gitignore")
		}
		if filepath.Base(f.Path) == "app.secret.go" {
			t.Errorf("Expected app.secret.go to be skipped by .gitignore")
		}
		if filepath.Base(f.Path) == "app.exe" {
			t.Errorf("Expected app.exe to be skipped by binary filter")
		}
	}
}
