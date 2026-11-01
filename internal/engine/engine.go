// Package engine coordinates the end-to-end CodeDNA intelligence pipeline:
// scanning, parsing, static analysis, graph generation, issue detection, and persistence.
package engine

import (
	"fmt"
	"runtime"
	"sync"
	"time"

	"github.com/codedna/codedna/internal/analysis"
	"github.com/codedna/codedna/internal/architecture"
	"github.com/codedna/codedna/internal/model"
	"github.com/codedna/codedna/internal/parser"
	"github.com/codedna/codedna/internal/scanner"
	"github.com/codedna/codedna/internal/store"
)

// PipelineResult encapsulates the results of a complete codebase analysis.
type PipelineResult struct {
	RepoID       int64
	ScanID       int64
	Repo         model.RepoInfo
	Scan         model.Scan
	Stats        *model.RepoScanStats
	Files        []model.FileNode
	Symbols      []model.Symbol
	Rels         []model.Relationship
	Issues       []model.Issue
	Architecture model.ArchHealthReport
	Index        *analysis.Index
	FileGraph    *analysis.FileGraph
}

// Run executes the complete CodeDNA scan and analysis pipeline.
func Run(repoRoot string, st *store.Store, progress scanner.ScanProgress) (*PipelineResult, error) {
	startTime := time.Now()

	// 1. Scan directory & extract metadata
	sc := scanner.New()
	scanRes, err := sc.Scan(repoRoot, progress)
	if err != nil {
		return nil, fmt.Errorf("scan failed: %w", err)
	}

	// 2. Persist / retrieve repo record
	repoID, err := st.UpsertRepo(scanRes.Name, scanRes.Root, scanRes.Git)
	if err != nil {
		return nil, fmt.Errorf("upsert repo: %w", err)
	}

	// 3. Begin scan run
	scanID, err := st.BeginScan(repoID)
	if err != nil {
		return nil, fmt.Errorf("begin scan: %w", err)
	}

	// 4. Parse code files using Tree-sitter in parallel
	canParse := scanner.ParsedLanguages()
	var parseableFiles []model.FileNode
	for _, f := range scanRes.Files {
		if canParse[f.Language] {
			parseableFiles = append(parseableFiles, f)
		}
	}

	totalParseable := len(parseableFiles)
	numWorkers := runtime.NumCPU()
	if numWorkers > 16 {
		numWorkers = 16
	}
	if numWorkers < 2 {
		numWorkers = 2
	}

	type parseTask struct {
		index int
		file  model.FileNode
	}
	type parseOutput struct {
		index int
		res   *model.FileParseResult
		err   error
	}

	tasks := make(chan parseTask, totalParseable)
	results := make(chan parseOutput, totalParseable)

	var wg sync.WaitGroup
	for w := 0; w < numWorkers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for t := range tasks {
				res, err := parser.Parse(scanRes.Root, t.file.Path)
				results <- parseOutput{index: t.index, res: res, err: err}
			}
		}()
	}

	for i, f := range parseableFiles {
		tasks <- parseTask{index: i, file: f}
	}
	close(tasks)

	go func() {
		wg.Wait()
		close(results)
	}()

	parseResults := make([]*model.FileParseResult, 0, totalParseable)
	var parseErrorsCount int
	completedCount := 0

	for out := range results {
		completedCount++
		if out.err != nil {
			parseErrorsCount++
		} else if out.res != nil {
			if len(out.res.ParseErrors) > 0 {
				parseErrorsCount += len(out.res.ParseErrors)
			}
			parseResults = append(parseResults, out.res)
		}

		if progress != nil && (completedCount%10 == 0 || completedCount == totalParseable) {
			filePath := ""
			if out.res != nil {
				filePath = out.res.File
			}
			progress(completedCount, totalParseable, filePath)
		}
	}

	// 5. Build symbol index & resolve relationships
	idx := analysis.Build(parseResults, scanRes.Files)
	rels := idx.Resolve(parseResults)

	// 6. Construct module graph & detect issues
	fileGraph := idx.ModuleGraph(parseResults)
	cycles := fileGraph.Cycles()
	deadCode := idx.DeadCodeCandidates()
	hotspots := idx.Hotspots(30)
	entrypoints := idx.EntryPoints(scanRes.Files)

	var allIssues []model.Issue
	allIssues = append(allIssues, cycles...)
	allIssues = append(allIssues, deadCode...)
	allIssues = append(allIssues, hotspots...)
	allIssues = append(allIssues, entrypoints...)

	// 7. Evaluate Architecture Boundaries & Rules
	archHealth, err := architecture.EvaluateRepo(scanRes.Root, idx, fileGraph)
	if err == nil {
		for _, v := range archHealth.Violations {
			allIssues = append(allIssues, model.Issue{
				Kind:     model.IssueLayerViolation,
				Severity: v.Severity,
				Title:    "Layer violation: " + v.RuleName,
				File:     v.FromFile,
				Detail: map[string]interface{}{
					"toFile":     v.ToFile,
					"importPath": v.ImportPath,
					"reason":     v.Reason,
					"line":       v.Line,
				},
			})
		}
	}

	// 8. Persist results to SQLite
	if err := st.SaveFiles(scanID, scanRes.Files); err != nil {
		return nil, fmt.Errorf("save files: %w", err)
	}
	if err := st.SaveSymbols(scanID, idx.Symbols); err != nil {
		return nil, fmt.Errorf("save symbols: %w", err)
	}
	if err := st.SaveRelationships(scanID, rels); err != nil {
		return nil, fmt.Errorf("save relationships: %w", err)
	}
	if err := st.SaveIssues(scanID, allIssues); err != nil {
		return nil, fmt.Errorf("save issues: %w", err)
	}

	// 9. Build aggregate scan statistics
	durationStr := time.Since(startTime).Round(time.Millisecond).String()
	stats := &model.RepoScanStats{
		FilesScanned:    len(scanRes.Files) + scanRes.SkippedFiles,
		FilesParsed:     len(parseResults),
		FilesSkipped:    scanRes.SkippedFiles,
		ParseErrors:     parseErrorsCount,
		SymbolsFound:    len(idx.Symbols),
		Relationships:   len(rels),
		ResolvedCalls:   idx.Resolved,
		UnresolvedCalls: idx.Unresolved,
		LOC:             scanRes.TotalLOC,
		Languages:       scanRes.Languages,
		LOCByLanguage:   scanRes.LOCByLang,
		Frameworks:      scanRes.Frameworks,
		DirectDeps:      scanRes.Deps,
		Duration:        durationStr,
	}

	if err := st.FinishScan(scanID, "complete", stats); err != nil {
		return nil, fmt.Errorf("finish scan: %w", err)
	}

	repoInfo := model.RepoInfo{
		ID:        repoID,
		Name:      scanRes.Name,
		Root:      scanRes.Root,
		CreatedAt: time.Now(),
		Git:       scanRes.Git,
	}

	scanInfo := model.Scan{
		ID:        scanID,
		RepoID:    repoID,
		StartedAt: startTime,
		Status:    "complete",
		Stats:     stats,
	}

	return &PipelineResult{
		RepoID:       repoID,
		ScanID:       scanID,
		Repo:         repoInfo,
		Scan:         scanInfo,
		Stats:        stats,
		Files:        scanRes.Files,
		Symbols:      idx.Symbols,
		Rels:         rels,
		Issues:       allIssues,
		Architecture: archHealth,
		Index:        idx,
		FileGraph:    fileGraph,
	}, nil
}
