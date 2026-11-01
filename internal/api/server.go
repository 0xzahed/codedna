// Package api provides the REST and Server-Sent Events (SSE) server for CodeDNA,
// providing endpoints for graph visualization, change impact, architecture drift,
// runtime tracing, AI codebase intelligence, and embedded web dashboard delivery.
package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/codedna/codedna/internal/ai"
	"github.com/codedna/codedna/internal/analysis"
	"github.com/codedna/codedna/internal/architecture"
	"github.com/codedna/codedna/internal/engine"
	"github.com/codedna/codedna/internal/impact"
	"github.com/codedna/codedna/internal/model"
	"github.com/codedna/codedna/internal/report"
	"github.com/codedna/codedna/internal/runtime"
	"github.com/codedna/codedna/internal/store"
	"github.com/codedna/codedna/internal/watcher"
)

// Server coordinates the API handlers, SSE broadcasters, and web dashboard.
type Server struct {
	store      *store.Store
	aiClient   *ai.Client
	webFS      http.FileSystem
	sseMu      sync.RWMutex
	sseClients map[chan string]bool
	watcher    *watcher.Watcher
	lastScanID int64
	repoRoot   string
	mux        *http.ServeMux
}

// Config specifies server options.
type Config struct {
	Store    *store.Store
	AIConfig ai.Config
	WebFS    http.FileSystem
	RepoRoot string
}

// NewServer initializes a CodeDNA API server.
func NewServer(cfg Config) *Server {
	if cfg.RepoRoot == "" {
		cfg.RepoRoot = "."
	}
	absRoot, _ := filepath.Abs(cfg.RepoRoot)

	s := &Server{
		store:      cfg.Store,
		aiClient:   ai.NewClient(cfg.AIConfig),
		webFS:      cfg.WebFS,
		sseClients: make(map[chan string]bool),
		repoRoot:   absRoot,
		mux:        http.NewServeMux(),
	}

	s.registerRoutes()
	return s
}

// Handler returns the HTTP handler with standard CORS and logging middleware.
func (s *Server) Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers for local development & API access
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		s.mux.ServeHTTP(w, r)
	})
}

func (s *Server) registerRoutes() {
	// API routes
	s.mux.HandleFunc("/api/health", s.handleHealth)
	s.mux.HandleFunc("/api/repos", s.handleRepos)
	s.mux.HandleFunc("/api/repos/scan", s.handleScan)
	s.mux.HandleFunc("/api/scans/", s.handleScanSubroutes)
	s.mux.HandleFunc("/api/runtime/events", s.handleRuntimeEvents)
	s.mux.HandleFunc("/api/ai/query", s.handleAIQuery)
	s.mux.HandleFunc("/api/ai/context", s.handleAIContext)
	s.mux.HandleFunc("/api/ai/business-logic", s.handleAIBusinessLogic)
	s.mux.HandleFunc("/api/events", s.handleSSE)

	// Web dashboard fallback / static files
	if s.webFS != nil {
		fileServer := http.FileServer(s.webFS)
		s.mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			if strings.HasPrefix(r.URL.Path, "/api/") {
				http.NotFound(w, r)
				return
			}

			// Clean incoming path
			cleanPath := r.URL.Path

			// 1. Try path directly
			if f, err := s.webFS.Open(cleanPath); err == nil {
				defer f.Close()
				if fi, err := f.Stat(); err == nil {
					if !fi.IsDir() {
						fileServer.ServeHTTP(w, r)
						return
					}
					// If directory, check if it has index.html
					if idxF, err := s.webFS.Open(filepath.Join(cleanPath, "index.html")); err == nil {
						idxF.Close()
						fileServer.ServeHTTP(w, r)
						return
					}
				}
			}

			// 2. Try path + /index.html
			if idxF, err := s.webFS.Open(filepath.Join(cleanPath, "index.html")); err == nil {
				idxF.Close()
				r.URL.Path = filepath.Join(cleanPath, "index.html")
				fileServer.ServeHTTP(w, r)
				return
			}

			// 3. Try path + .html
			if htmlF, err := s.webFS.Open(cleanPath + ".html"); err == nil {
				htmlF.Close()
				r.URL.Path = cleanPath + ".html"
				fileServer.ServeHTTP(w, r)
				return
			}

			// 4. SPA fallback: serve root index.html
			r.URL.Path = "/"
			fileServer.ServeHTTP(w, r)
		})
	} else {
		// Fallback landing page if no embedded web assets
		s.mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/" {
				w.Header().Set("Content-Type", "text/html; charset=utf-8")
				w.Write([]byte(`<!DOCTYPE html><html><head><title>CodeDNA 2.0</title><style>body{background:#121214;color:#e4e4e7;font-family:sans-serif;padding:40px;line-height:1.6;}h1{color:#60a5fa;}code{background:#27272a;padding:2px 6px;border-radius:4px;}</style></head><body><h1>CodeDNA 2.0 API Server</h1><p>Understand Your Code. See Its DNA.</p><p>API endpoints active at <code>/api/*</code>.</p></body></html>`))
				return
			}
			http.NotFound(w, r)
		})
	}
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"status":    "healthy",
		"service":   "codedna",
		"version":   "2.0.0",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"repoRoot":  s.repoRoot,
	})
}

func (s *Server) handleRepos(w http.ResponseWriter, r *http.Request) {
	repos, err := s.store.LoadAllRepos()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load repos: "+err.Error())
		return
	}
	writeJSON(w, http.StatusOK, repos)
}

func (s *Server) handleScan(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	targetPath := s.repoRoot
	var body struct {
		Path string `json:"path"`
	}
	if r.Body != nil {
		json.NewDecoder(r.Body).Decode(&body)
		if body.Path != "" {
			targetPath = body.Path
		}
	}

	progressCb := func(done, total int, path string) {
		s.broadcastSSE(fmt.Sprintf(`{"type":"scan_progress","done":%d,"total":%d,"file":%q}`, done, total, path))
	}

	res, err := engine.Run(targetPath, s.store, progressCb)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "pipeline execution failed: "+err.Error())
		return
	}

	s.lastScanID = res.ScanID
	s.broadcastSSE(fmt.Sprintf(`{"type":"scan_completed","scanId":%d,"repoId":%d}`, res.ScanID, res.RepoID))

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"scanId": res.ScanID,
		"repoId": res.RepoID,
		"repo":   res.Repo,
		"stats":  res.Stats,
	})
}

func (s *Server) handleScanSubroutes(w http.ResponseWriter, r *http.Request) {
	// Parse /api/scans/:id/...
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/scans/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		http.NotFound(w, r)
		return
	}

	scanID, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid scan id")
		return
	}

	sub := ""
	if len(parts) > 1 {
		sub = parts[1]
	}

	switch sub {
	case "":
		s.handleScanDetail(w, r, scanID)
	case "graph":
		s.handleScanGraph(w, r, scanID)
	case "symbols":
		s.handleScanSymbols(w, r, scanID)
	case "files":
		s.handleScanFiles(w, r, scanID)
	case "file":
		s.handleScanFileContent(w, r, scanID)
	case "issues":
		s.handleScanIssues(w, r, scanID)
	case "dependencies":
		s.handleScanDependencies(w, r, scanID)
	case "impact":
		s.handleScanImpact(w, r, scanID)
	case "architecture":
		s.handleScanArchitecture(w, r, scanID)
	case "runtime":
		s.handleScanRuntime(w, r, scanID)
	case "report":
		s.handleScanReport(w, r, scanID)
	default:
		http.NotFound(w, r)
	}
}

func (s *Server) handleScanDetail(w http.ResponseWriter, r *http.Request, scanID int64) {
	syms, err := s.store.LoadSymbols(scanID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"scanId":      scanID,
		"symbolCount": len(syms),
	})
}

func (s *Server) handleScanGraph(w http.ResponseWriter, r *http.Request, scanID int64) {
	files, err := s.store.LoadFiles(scanID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load files: "+err.Error())
		return
	}
	syms, err := s.store.LoadSymbols(scanID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load symbols: "+err.Error())
		return
	}
	rels, err := s.store.LoadRelationships(scanID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load relationships: "+err.Error())
		return
	}

	// Filter options
	filterType := r.URL.Query().Get("type")
	searchQuery := strings.ToLower(r.URL.Query().Get("q"))

	var nodes []model.GraphNode
	var edges []model.GraphEdge

	nodeMap := make(map[string]bool)

	// File nodes
	for _, f := range files {
		if filterType != "" && filterType != "file" {
			continue
		}
		if searchQuery != "" && !strings.Contains(strings.ToLower(f.Path), searchQuery) {
			continue
		}
		nodes = append(nodes, model.GraphNode{
			ID:         "file:" + f.Path,
			Type:       "file",
			Label:      filepath.Base(f.Path),
			File:       f.Path,
			Language:   f.Language,
			Complexity: f.LOC,
			Data: map[string]interface{}{
				"size": f.Size,
				"loc":  f.LOC,
				"tags": f.Tags,
			},
		})
		nodeMap["file:"+f.Path] = true
	}

	// Symbol nodes (functions, classes, methods)
	for _, sym := range syms {
		if filterType != "" && string(sym.Kind) != filterType {
			continue
		}
		if searchQuery != "" && !strings.Contains(strings.ToLower(sym.Name), searchQuery) && !strings.Contains(strings.ToLower(sym.File), searchQuery) {
			continue
		}

		nodeType := string(sym.Kind)
		nodes = append(nodes, model.GraphNode{
			ID:         sym.ID,
			Type:       nodeType,
			Label:      sym.Name,
			File:       sym.File,
			Line:       sym.Line,
			Complexity: sym.Complexity,
			Exported:   sym.Exported,
			Language:   sym.Language,
			Data: map[string]interface{}{
				"qualifiedName": sym.QualifiedName,
				"endLine":       sym.EndLine,
				"parentClass":   sym.ParentClassName,
			},
		})
		nodeMap[sym.ID] = true
	}

	// Graph edges
	for i, rel := range rels {
		source := rel.FromID
		target := rel.ToID
		if target == "" {
			continue
		}

		// Only include edges whose endpoints are in node selection when filtered
		if filterType != "" || searchQuery != "" {
			if !nodeMap[source] && !nodeMap[target] {
				continue
			}
		}

		edges = append(edges, model.GraphEdge{
			ID:         fmt.Sprintf("e-%d", i),
			Source:     source,
			Target:     target,
			Kind:       string(rel.Kind),
			Confidence: rel.Confidence,
			Method:     string(rel.Method),
			Locations:  rel.Locations,
		})
	}

	// Include runtime edges if present
	traces, _ := s.store.LoadRuntimeTraces(scanID)
	if len(traces) > 0 {
		rtSummary := runtime.IngestTraces(traces, rels)
		for _, re := range rtSummary.RuntimeEdges {
			edges = append(edges, re)
		}
	}

	writeJSON(w, http.StatusOK, model.GraphData{
		Nodes: nodes,
		Edges: edges,
	})
}

func (s *Server) handleScanSymbols(w http.ResponseWriter, r *http.Request, scanID int64) {
	syms, err := s.store.LoadSymbols(scanID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	q := strings.ToLower(r.URL.Query().Get("q"))
	kind := r.URL.Query().Get("kind")
	file := r.URL.Query().Get("file")

	var filtered []model.Symbol
	for _, s := range syms {
		if q != "" && !strings.Contains(strings.ToLower(s.Name), q) && !strings.Contains(strings.ToLower(s.QualifiedName), q) {
			continue
		}
		if kind != "" && string(s.Kind) != kind {
			continue
		}
		if file != "" && s.File != file {
			continue
		}
		filtered = append(filtered, s)
	}

	writeJSON(w, http.StatusOK, filtered)
}

func (s *Server) handleScanFiles(w http.ResponseWriter, r *http.Request, scanID int64) {
	files, err := s.store.LoadFiles(scanID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, files)
}

func (s *Server) handleScanFileContent(w http.ResponseWriter, r *http.Request, scanID int64) {
	reqPath := r.URL.Query().Get("path")
	if reqPath == "" {
		writeError(w, http.StatusBadRequest, "missing path parameter")
		return
	}

	// Security: Path traversal protection
	cleaned := filepath.Clean(reqPath)
	if strings.HasPrefix(cleaned, "..") || filepath.IsAbs(cleaned) {
		writeError(w, http.StatusForbidden, "invalid file path (path traversal prevented)")
		return
	}

	fullPath := filepath.Join(s.repoRoot, cleaned)
	if !strings.HasPrefix(fullPath, s.repoRoot) {
		writeError(w, http.StatusForbidden, "access denied: path outside repository root")
		return
	}

	content, err := os.ReadFile(fullPath)
	if err != nil {
		writeError(w, http.StatusNotFound, "file not found: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"path":    cleaned,
		"content": string(content),
		"size":    len(content),
	})
}

func (s *Server) handleScanIssues(w http.ResponseWriter, r *http.Request, scanID int64) {
	issues, err := s.store.LoadIssues(scanID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, issues)
}

func (s *Server) handleScanDependencies(w http.ResponseWriter, r *http.Request, scanID int64) {
	// Discover dependencies directly or load from latest stats
	stats, _, err := s.store.LastStats(1)
	if err != nil || stats == nil || len(stats.DirectDeps) == 0 {
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"dependencies": map[string][]string{},
		})
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"dependencies": stats.DirectDeps,
		"frameworks":   stats.Frameworks,
	})
}

func (s *Server) handleScanImpact(w http.ResponseWriter, r *http.Request, scanID int64) {
	baseRef := r.URL.Query().Get("base")
	if baseRef == "" {
		baseRef = "main"
	}

	diffs, err := impact.ParseGitDiff(s.repoRoot, baseRef)
	if err != nil {
		// Try diffing against empty base
		diffs, _ = impact.ParseGitDiff(s.repoRoot, "")
	}

	syms, _ := s.store.LoadSymbols(scanID)
	rels, _ := s.store.LoadRelationships(scanID)
	files, _ := s.store.LoadFiles(scanID)

	idx := &analysis.Index{
		Symbols: syms,
		Rel:     rels,
	}

	report := impact.AnalyzeImpact(s.repoRoot, baseRef, diffs, idx, files)
	s.store.SaveImpactReport(scanID, report)

	writeJSON(w, http.StatusOK, report)
}

func (s *Server) handleScanArchitecture(w http.ResponseWriter, r *http.Request, scanID int64) {
	rules, err := architecture.LoadRules(s.repoRoot)
	if err != nil {
		rules = architecture.DefaultRules()
	}

	syms, _ := s.store.LoadSymbols(scanID)
	rels, _ := s.store.LoadRelationships(scanID)
	issues, _ := s.store.LoadIssues(scanID)

	var cycles []model.Issue
	for _, iss := range issues {
		if iss.Kind == model.IssueCircularDep {
			cycles = append(cycles, iss)
		}
	}

	report := architecture.CheckArchitecture(rules, rels, syms, cycles)
	writeJSON(w, http.StatusOK, report)
}

func (s *Server) handleRuntimeEvents(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeError(w, http.StatusBadRequest, "cannot read request body")
		return
	}

	var events []model.RuntimeTraceEvent
	// Try array first, then single event
	if err := json.Unmarshal(body, &events); err != nil {
		var single model.RuntimeTraceEvent
		if err2 := json.Unmarshal(body, &single); err2 != nil {
			writeError(w, http.StatusBadRequest, "invalid runtime trace payload")
			return
		}
		events = append(events, single)
	}

	scanID := s.lastScanID
	if scanID <= 0 {
		scanID = 1
	}

	if err := s.store.SaveRuntimeTraces(scanID, events); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to persist trace events: "+err.Error())
		return
	}

	s.broadcastSSE(fmt.Sprintf(`{"type":"runtime_events_ingested","count":%d}`, len(events)))
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"status":   "recorded",
		"received": len(events),
	})
}

func (s *Server) handleScanRuntime(w http.ResponseWriter, r *http.Request, scanID int64) {
	traces, err := s.store.LoadRuntimeTraces(scanID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	rels, _ := s.store.LoadRelationships(scanID)
	summary := runtime.IngestTraces(traces, rels)
	writeJSON(w, http.StatusOK, summary)
}

func (s *Server) handleAIQuery(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req model.AIQueryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid query request")
		return
	}

	scanID := s.lastScanID
	if scanID <= 0 {
		scanID = 1
	}

	syms, _ := s.store.LoadSymbols(scanID)
	rels, _ := s.store.LoadRelationships(scanID)
	stats, _, _ := s.store.LastStats(1)

	resp, err := s.aiClient.QueryCodebase(r.Context(), req, syms, rels, stats)
	if err != nil {
		if err == ai.ErrConsentRequired {
			writeJSON(w, http.StatusForbidden, map[string]interface{}{
				"error":   "Privacy protection: remote AI calls require explicit user consent.",
				"consent": false,
			})
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

func (s *Server) handleAIContext(w http.ResponseWriter, r *http.Request) {
	format := r.URL.Query().Get("format")
	if format == "" {
		format = "markdown"
	}

	scanID := s.lastScanID
	if scanID <= 0 {
		scanID = 1
	}

	files, _ := s.store.LoadFiles(scanID)
	syms, _ := s.store.LoadSymbols(scanID)
	rels, _ := s.store.LoadRelationships(scanID)
	stats, _, _ := s.store.LastStats(1)

	repoName := filepath.Base(s.repoRoot)
	ctxText := ai.GenerateContext(format, repoName, files, syms, rels, stats)

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(ctxText))
}

func (s *Server) handleAIBusinessLogic(w http.ResponseWriter, r *http.Request) {
	scanID := s.lastScanID
	if scanID <= 0 {
		scanID = 1
	}

	files, _ := s.store.LoadFiles(scanID)
	syms, _ := s.store.LoadSymbols(scanID)

	modules := ai.MapBusinessLogic(files, syms)
	writeJSON(w, http.StatusOK, modules)
}

func (s *Server) handleScanReport(w http.ResponseWriter, r *http.Request, scanID int64) {
	format := r.URL.Query().Get("format")
	if format == "" {
		format = "markdown"
	}

	repo, _ := s.store.LoadRepo(1)
	if repo == nil {
		repo = &model.RepoInfo{Name: filepath.Base(s.repoRoot), Root: s.repoRoot}
	}

	stats, _, _ := s.store.LastStats(1)
	files, _ := s.store.LoadFiles(scanID)
	syms, _ := s.store.LoadSymbols(scanID)
	rels, _ := s.store.LoadRelationships(scanID)
	issues, _ := s.store.LoadIssues(scanID)
	traces, _ := s.store.LoadRuntimeTraces(scanID)

	rules, _ := architecture.LoadRules(s.repoRoot)
	archReport := architecture.CheckArchitecture(rules, rels, syms, nil)
	rtSummary := runtime.IngestTraces(traces, rels)

	repData := report.CompleteReportData{
		Repo:         *repo,
		Scan:         model.Scan{ID: scanID, Status: "complete"},
		Stats:        stats,
		Files:        files,
		TopSymbols:   syms,
		Issues:       issues,
		Architecture: &archReport,
		Runtime:      rtSummary,
		GeneratedAt:  time.Now(),
	}
	repData.HealthScore = report.CalculateHealthScore(stats, issues, &archReport)

	if format == "json" {
		js, err := report.GenerateJSON(repData)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(js)
		return
	}

	md := report.GenerateMarkdown(repData)
	w.Header().Set("Content-Type", "text/markdown; charset=utf-8")
	w.Write([]byte(md))
}

func (s *Server) handleSSE(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	clientChan := make(chan string, 10)
	s.sseMu.Lock()
	s.sseClients[clientChan] = true
	s.sseMu.Unlock()

	defer func() {
		s.sseMu.Lock()
		delete(s.sseClients, clientChan)
		close(clientChan)
		s.sseMu.Unlock()
	}()

	// Send initial ping
	fmt.Fprintf(w, "data: {\"type\":\"connected\"}\n\n")
	flusher.Flush()

	for {
		select {
		case <-r.Context().Done():
			return
		case msg, ok := <-clientChan:
			if !ok {
				return
			}
			fmt.Fprintf(w, "data: %s\n\n", msg)
			flusher.Flush()
		}
	}
}

func (s *Server) broadcastSSE(msg string) {
	s.sseMu.RLock()
	defer s.sseMu.RUnlock()

	for ch := range s.sseClients {
		select {
		case ch <- msg:
		default:
		}
	}
}

// StartWatcher attaches directory watching with SSE notifications.
func (s *Server) StartWatcher(ctx context.Context) {
	s.watcher = watcher.New(s.repoRoot, 1*time.Second, func(changed []string) {
		s.broadcastSSE(fmt.Sprintf(`{"type":"files_changed","files":%s}`, toJSONString(changed)))
	})
	go s.watcher.Start(ctx)
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func toJSONString(v interface{}) string {
	b, _ := json.Marshal(v)
	return string(b)
}
