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

	// 5. Query using 'latest' alias
	latestGraphReq := httptest.NewRequest("GET", "/api/scans/latest/graph", nil)
	wLatest := httptest.NewRecorder()
	srv.Handler().ServeHTTP(wLatest, latestGraphReq)
	if wLatest.Code != http.StatusOK {
		t.Fatalf("Query using 'latest' alias failed with code %d: %s", wLatest.Code, wLatest.Body.String())
	}

	// 6. Query export using 'latest' alias
	exportReq := httptest.NewRequest("GET", "/api/scans/latest/export?format=mermaid", nil)
	wExport := httptest.NewRecorder()
	srv.Handler().ServeHTTP(wExport, exportReq)
	if wExport.Code != http.StatusOK {
		t.Fatalf("Export with 'latest' alias failed with code %d: %s", wExport.Code, wExport.Body.String())
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

func TestArchitectureRulesEndpoints(t *testing.T) {
	srv, _, st := setupTestServer(t)
	defer st.Close()

	// Trigger initial scan to set up repo and scan #1
	scanReq := httptest.NewRequest("POST", "/api/repos/scan", bytes.NewReader([]byte("{}")))
	wInit := httptest.NewRecorder()
	srv.Handler().ServeHTTP(wInit, scanReq)
	if wInit.Code != http.StatusOK {
		t.Fatalf("Initial scan failed: %d %s", wInit.Code, wInit.Body.String())
	}

	// 1. GET architecture report
	req := httptest.NewRequest("GET", "/api/scans/1/architecture", nil)
	w := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("GET /architecture failed: %d %s", w.Code, w.Body.String())
	}
	var rep model.ArchHealthReport
	if err := json.NewDecoder(w.Body).Decode(&rep); err != nil {
		t.Fatalf("Decode failed: %v", err)
	}
	initRuleCount := rep.RuleCount

	// 2. POST new custom architecture rule
	newRule := model.ArchRule{
		ID:          "custom-db-protect",
		Name:        "Direct DB Isolation",
		FromPattern: "**/web/**",
		ToPattern:   "**/db/**",
		Forbidden:   true,
		Severity:    model.SevHigh,
		Description: "Web routes cannot directly reach DB",
	}
	body, _ := json.Marshal(newRule)
	reqPost := httptest.NewRequest("POST", "/api/scans/1/architecture/rules", bytes.NewReader(body))
	wPost := httptest.NewRecorder()
	srv.Handler().ServeHTTP(wPost, reqPost)
	if wPost.Code != http.StatusOK {
		t.Fatalf("POST /architecture/rules failed: %d %s", wPost.Code, wPost.Body.String())
	}
	var repAfterAdd model.ArchHealthReport
	json.NewDecoder(wPost.Body).Decode(&repAfterAdd)
	if repAfterAdd.RuleCount != initRuleCount+1 {
		t.Errorf("Expected %d rules after addition, got %d", initRuleCount+1, repAfterAdd.RuleCount)
	}

	// 3. DELETE custom architecture rule
	reqDel := httptest.NewRequest("DELETE", "/api/scans/1/architecture/rules/custom-db-protect", nil)
	wDel := httptest.NewRecorder()
	srv.Handler().ServeHTTP(wDel, reqDel)
	if wDel.Code != http.StatusOK {
		t.Fatalf("DELETE /architecture/rules/custom-db-protect failed: %d %s", wDel.Code, wDel.Body.String())
	}
	var repAfterDel model.ArchHealthReport
	json.NewDecoder(wDel.Body).Decode(&repAfterDel)
	if repAfterDel.RuleCount != initRuleCount {
		t.Errorf("Expected %d rules after delete, got %d", initRuleCount, repAfterDel.RuleCount)
	}

	// 4. Reset rules
	reqReset := httptest.NewRequest("POST", "/api/scans/1/architecture/rules/reset", nil)
	wReset := httptest.NewRecorder()
	srv.Handler().ServeHTTP(wReset, reqReset)
	if wReset.Code != http.StatusOK {
		t.Fatalf("POST /architecture/rules/reset failed: %d %s", wReset.Code, wReset.Body.String())
	}
}
