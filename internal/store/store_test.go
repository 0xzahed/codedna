package store

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/codedna/codedna/internal/model"
)

func TestStoreOperations(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test_codedna.db")

	s, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	defer s.Close()

	// Upsert repo
	git := model.GitMeta{
		IsGitRepo: true,
		Branch:    "main",
		Commit:    "abcdef123456",
	}
	repoID, err := s.UpsertRepo("test-repo", "/home/test/repo", git)
	if err != nil {
		t.Fatalf("UpsertRepo failed: %v", err)
	}
	if repoID <= 0 {
		t.Errorf("Expected positive repoID, got %d", repoID)
	}

	// Begin scan
	scanID, err := s.BeginScan(repoID)
	if err != nil {
		t.Fatalf("BeginScan failed: %v", err)
	}

	// Save files
	files := []model.FileNode{
		{
			Path:     "main.go",
			Language: "go",
			Size:     1024,
			LOC:      50,
			Tags:     []string{"entrypoint"},
		},
	}
	if err := s.SaveFiles(scanID, files); err != nil {
		t.Fatalf("SaveFiles failed: %v", err)
	}

	// Save symbols
	symbols := []model.Symbol{
		{
			ID:            "main.go:main",
			Name:          "main",
			QualifiedName: "main",
			Kind:          model.SymbolFunction,
			File:          "main.go",
			Language:      "go",
			Line:          10,
			Complexity:    2,
			Exported:      false,
		},
	}
	if err := s.SaveSymbols(scanID, symbols); err != nil {
		t.Fatalf("SaveSymbols failed: %v", err)
	}

	// Save relationships
	rels := []model.Relationship{
		{
			Kind:       model.RelCalls,
			FromID:     "main.go:main",
			ToID:       "fmt.Println",
			ToName:     "Println",
			Confidence: 0.9,
			Method:     model.MethodExactUnique,
			Locations:  []model.Ref{{File: "main.go", Line: 12, Col: 4}},
		},
	}
	if err := s.SaveRelationships(scanID, rels); err != nil {
		t.Fatalf("SaveRelationships failed: %v", err)
	}

	// Save issues
	issues := []model.Issue{
		{
			Kind:     model.IssueComplexity,
			Severity: model.SevLow,
			Title:    "Notice",
			File:     "main.go",
			Detail:   map[string]interface{}{"msg": "all good"},
		},
	}
	if err := s.SaveIssues(scanID, issues); err != nil {
		t.Fatalf("SaveIssues failed: %v", err)
	}

	// Finish scan
	now := time.Now()
	stats := &model.RepoScanStats{
		FilesScanned:  1,
		FilesParsed:   1,
		SymbolsFound:  1,
		Relationships: 1,
		LOC:           50,
		Languages:     map[string]int{"go": 1},
	}
	if err := s.FinishScan(scanID, "complete", stats); err != nil {
		t.Fatalf("FinishScan failed: %v", err)
	}
	_ = now

	// Query back
	loadedSyms, err := s.LoadSymbols(scanID)
	if err != nil {
		t.Fatalf("LoadSymbols failed: %v", err)
	}
	if len(loadedSyms) != 1 || loadedSyms[0].Name != "main" {
		t.Errorf("Unexpected loaded symbols: %+v", loadedSyms)
	}

	loadedRels, err := s.LoadRelationships(scanID)
	if err != nil {
		t.Fatalf("LoadRelationships failed: %v", err)
	}
	if len(loadedRels) != 1 || loadedRels[0].ToName != "Println" {
		t.Errorf("Unexpected loaded rels: %+v", loadedRels)
	}

	loadedIssues, err := s.LoadIssues(scanID)
	if err != nil {
		t.Fatalf("LoadIssues failed: %v", err)
	}
	if len(loadedIssues) != 1 {
		t.Errorf("Unexpected loaded issues: %+v", loadedIssues)
	}

	loadedFiles, err := s.LoadFiles(scanID)
	if err != nil {
		t.Fatalf("LoadFiles failed: %v", err)
	}
	if len(loadedFiles) != 1 || loadedFiles[0].Path != "main.go" {
		t.Errorf("Unexpected loaded files: %+v", loadedFiles)
	}
}
