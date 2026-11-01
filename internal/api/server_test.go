package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/codedna/codedna/internal/ai"
	"github.com/codedna/codedna/internal/model"
	"github.com/codedna/codedna/internal/store"
)

func setupTestServer(t *testing.T) (*Server, string, *store.Store) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	st, err := store.Open(dbPath)
	if err != nil {
		t.Fatalf("Failed to open store: %v", err)
	}

	// Create test files in tmpDir
	testGo := `package main

func Execute() {
	println("test")
}
`
	os.WriteFile(filepath.Join(tmpDir, "main.go"), []byte(testGo), 0o644)

	srv := NewServer(Config{
		Store:    st,
		AIConfig: ai.Config{Provider: "offline"},
		RepoRoot: tmpDir,
	})

	return srv, tmpDir, st
}

func TestHealthEndpoint(t *testing.T) {
	srv, _, st := setupTestServer(t)
	defer st.Close()

	req := httptest.NewRequest("GET", "/api/health", nil)
	w := httptest.NewRecorder()

	srv.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected 200 OK, got %d", w.Code)
	}

	var res map[string]interface{}
	json.NewDecoder(w.Body).Decode(&res)
	if res["status"] != "healthy" {
		t.Errorf("Expected healthy status, got %v", res["status"])
	}
}

func TestScanAndGraphEndpoints(t *testing.T) {
	srv, _, st := setupTestServer(t)
	defer st.Close()

	// 1. Trigger scan
	scanReq := httptest.NewRequest("POST", "/api/repos/scan", bytes.NewReader([]byte("{}")))
	w := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, scanReq)

	if w.Code != http.StatusOK {
		t.Fatalf("Scan failed with code %d: %s", w.Code, w.Body.String())
	}

	// 2. Query graph
	graphReq := httptest.NewRequest("GET", "/api/scans/1/graph", nil)
	wGraph := httptest.NewRecorder()
	srv.Handler().ServeHTTP(wGraph, graphReq)

	if wGraph.Code != http.StatusOK {
		t.Fatalf("Graph query failed with code %d: %s", wGraph.Code, wGraph.Body.String())
	}

	var graph model.GraphData
	json.NewDecoder(wGraph.Body).Decode(&graph)
	if len(graph.Nodes) == 0 {
		t.Errorf("Expected graph nodes in response")
	}

	// 3. Query symbols
	symReq := httptest.NewRequest("GET", "/api/scans/1/symbols?q=Execute", nil)
	wSym := httptest.NewRecorder()
	srv.Handler().ServeHTTP(wSym, symReq)

	var syms []model.Symbol
	json.NewDecoder(wSym.Body).Decode(&syms)
	if len(syms) == 0 || syms[0].Name != "Execute" {
		t.Errorf("Expected Execute symbol, got %+v", syms)
	}

	// 4. Safe file content endpoint
	fileReq := httptest.NewRequest("GET", "/api/scans/1/file?path=main.go", nil)
	wFile := httptest.NewRecorder()
	srv.Handler().ServeHTTP(wFile, fileReq)
	if wFile.Code != http.StatusOK {
		t.Errorf("Expected 200 OK for file content, got %d", wFile.Code)
	}

	// Path traversal protection test
	badReq := httptest.NewRequest("GET", "/api/scans/1/file?path=../../etc/passwd", nil)
	wBad := httptest.NewRecorder()
	srv.Handler().ServeHTTP(wBad, badReq)
	if wBad.Code != http.StatusForbidden {
		t.Errorf("Expected 403 Forbidden for path traversal, got %d", wBad.Code)
	}
}

func TestRuntimeEventsAndReport(t *testing.T) {
	srv, _, st := setupTestServer(t)
	defer st.Close()

	// Ingest runtime trace
	events := []model.RuntimeTraceEvent{
		{
			RunID:        "test-run",
			FunctionName: "Execute",
			File:         "main.go",
			Line:         3,
			DurationUs:   500,
			TimestampNs:  1000,
		},
	}
	body, _ := json.Marshal(events)
	rtReq := httptest.NewRequest("POST", "/api/runtime/events", bytes.NewReader(body))
	wRt := httptest.NewRecorder()
	srv.Handler().ServeHTTP(wRt, rtReq)

	if wRt.Code != http.StatusOK {
		t.Fatalf("Runtime ingest failed: %s", wRt.Body.String())
	}

	// Query runtime summary
	summaryReq := httptest.NewRequest("GET", "/api/scans/1/runtime", nil)
	wSum := httptest.NewRecorder()
	srv.Handler().ServeHTTP(wSum, summaryReq)

	var summary model.RuntimeSummary
	json.NewDecoder(wSum.Body).Decode(&summary)
	if summary.TotalEvents != 1 {
		t.Errorf("Expected 1 runtime event, got %d", summary.TotalEvents)
	}

	// Query Markdown report
	repReq := httptest.NewRequest("GET", "/api/scans/1/report?format=markdown", nil)
	wRep := httptest.NewRecorder()
	srv.Handler().ServeHTTP(wRep, repReq)

	if wRep.Code != http.StatusOK {
		t.Fatalf("Report failed: %s", wRep.Body.String())
	}
	if !bytes.Contains(wRep.Body.Bytes(), []byte("Software Intelligence Report")) {
		t.Errorf("Unexpected report content")
	}
}
