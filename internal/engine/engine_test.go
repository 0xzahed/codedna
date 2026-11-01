package engine

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/codedna/codedna/internal/store"
)

func TestEnginePipeline(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	st, err := store.Open(dbPath)
	if err != nil {
		t.Fatalf("Failed to open store: %v", err)
	}
	defer st.Close()

	// Create test repository files
	repoDir := filepath.Join(tmpDir, "repo")
	err = os.MkdirAll(filepath.Join(repoDir, "src"), 0o755)
	if err != nil {
		t.Fatal(err)
	}

	mainGo := `package main

import "fmt"

func Process() {
	fmt.Println("processing")
	RunStep()
}

func RunStep() {
	fmt.Println("step")
}

func main() {
	Process()
}
`
	err = os.WriteFile(filepath.Join(repoDir, "src", "main.go"), []byte(mainGo), 0o644)
	if err != nil {
		t.Fatal(err)
	}

	res, err := Run(repoDir, st, nil)
	if err != nil {
		t.Fatalf("Pipeline Run failed: %v", err)
	}

	if res.RepoID <= 0 || res.ScanID <= 0 {
		t.Errorf("Expected valid RepoID and ScanID, got %d, %d", res.RepoID, res.ScanID)
	}

	if res.Stats.FilesParsed != 1 {
		t.Errorf("Expected 1 file parsed, got %d", res.Stats.FilesParsed)
	}

	if len(res.Symbols) < 3 { // Process, RunStep, main
		t.Errorf("Expected at least 3 symbols, got %d", len(res.Symbols))
	}

	if len(res.Rels) == 0 {
		t.Errorf("Expected relationships to be resolved")
	}

	// Verify DB storage
	loadedSyms, err := st.LoadSymbols(res.ScanID)
	if err != nil {
		t.Fatalf("Failed to load symbols from DB: %v", err)
	}
	if len(loadedSyms) != len(res.Symbols) {
		t.Errorf("Symbol count mismatch in DB: %d vs %d", len(loadedSyms), len(res.Symbols))
	}
}
