// Package impact analyzes the blast radius and consequences of code changes
// using Git diffs, AST symbol mappings, and dependency graph reachability.
package impact

import (
	"bufio"
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/codedna/codedna/internal/analysis"
	"github.com/codedna/codedna/internal/model"
)

// FileDiff holds changed lines for a file.
type FileDiff struct {
	Path         string
	Status       string // "modified", "added", "deleted"
	ChangedLines []int
}

var hunkHeaderRe = regexp.MustCompile(`^@@ -[0-9]+(?:,[0-9]+)? \+([0-9]+)(?:,([0-9]+))? @@`)

// ParseGitDiff runs git diff and extracts modified files and line numbers.
func ParseGitDiff(repoRoot string, baseRef string) ([]FileDiff, error) {
	var args []string
	if baseRef != "" {
		args = []string{"-C", repoRoot, "diff", "--unified=0", baseRef}
	} else {
		// Default to working tree diff vs HEAD
		args = []string{"-C", repoRoot, "diff", "--unified=0", "HEAD"}
	}

	cmd := exec.Command("git", args...)
	cmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0")
	out, err := cmd.CombinedOutput()
	if err != nil {
		// If HEAD diff fails (e.g. initial empty repo), fallback to diffing cached or untracked
		cmd = exec.Command("git", "-C", repoRoot, "diff", "--unified=0")
		out, err = cmd.CombinedOutput()
		if err != nil {
			return nil, fmt.Errorf("git diff failed: %s (%w)", string(out), err)
		}
	}

	return ParseDiffOutput(string(out)), nil
}

// ParseDiffOutput parses unified diff text into FileDiff structs.
func ParseDiffOutput(diffText string) []FileDiff {
	var results []FileDiff
	var current *FileDiff

	scanner := bufio.NewScanner(strings.NewReader(diffText))
	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, "diff --git ") {
			if current != nil {
				results = append(results, *current)
			}
			parts := strings.Fields(line)
			filePath := ""
			if len(parts) >= 4 {
				filePath = strings.TrimPrefix(parts[3], "b/")
			}
			current = &FileDiff{
				Path:   filePath,
				Status: "modified",
			}
			continue
		}

		if current == nil {
			continue
		}

		if strings.HasPrefix(line, "new file mode") {
			current.Status = "added"
		} else if strings.HasPrefix(line, "deleted file mode") {
			current.Status = "deleted"
		}

		if strings.HasPrefix(line, "@@") {
			m := hunkHeaderRe.FindStringSubmatch(line)
			if len(m) >= 2 {
				startLine, _ := strconv.Atoi(m[1])
				count := 1
				if len(m) >= 3 && m[2] != "" {
					count, _ = strconv.Atoi(m[2])
				}
				for i := 0; i < count; i++ {
					current.ChangedLines = append(current.ChangedLines, startLine+i)
				}
			}
		}
	}

	if current != nil {
		results = append(results, *current)
	}

	return results
}

// AnalyzeImpact computes affected symbols, upstream/downstream dependents,
// impacted APIs, affected tests, and an overall risk score.
func AnalyzeImpact(repoRoot string, baseRef string, diffs []FileDiff, idx *analysis.Index, files []model.FileNode) *model.ImpactReport {
	report := &model.ImpactReport{
		BaseRef:   baseRef,
		TargetRef: "HEAD",
		CreatedAt: time.Now(),
	}

	if len(diffs) == 0 {
		report.Summary = "No code changes detected against base ref."
		report.RiskScore = 0
		report.RiskLevel = model.SevInfo
		return report
	}

	fileMap := make(map[string]model.FileNode)
	for _, f := range files {
		fileMap[f.Path] = f
	}

	changedFilePaths := []string{}
	changedSymbols := []model.ChangedSymbol{}
	changedSymIDs := make(map[string]bool)
	changedFileSet := make(map[string]bool)

	for _, d := range diffs {
		changedFilePaths = append(changedFilePaths, d.Path)
		changedFileSet[d.Path] = true

		// Find symbols overlapping changed lines
		for _, sym := range idx.Symbols {
			if sym.File != d.Path {
				continue
			}

			// Check if any changed line falls within symbol range
			overlaps := false
			for _, line := range d.ChangedLines {
				if line >= sym.Line && (sym.EndLine == 0 || line <= sym.EndLine) {
					overlaps = true
					break
				}
			}

			if overlaps {
				changedSymbols = append(changedSymbols, model.ChangedSymbol{
					Symbol:       sym,
					ChangeType:   d.Status,
					ChangedLines: d.ChangedLines,
				})
				changedSymIDs[sym.ID] = true
			}
		}
	}

	report.ChangedFiles = changedFilePaths
	report.ChangedSymbols = changedSymbols

	// Upstream analysis: who calls or imports changed symbols / files?
	upstreamAffected := make(map[string]bool)
	downstreamCalls := make(map[string]bool)

	for _, rel := range idx.Rel {
		// Downstream: Changed symbol calls ToID
		if changedSymIDs[rel.FromID] && rel.ToID != "" {
			downstreamCalls[rel.ToID] = true
		}
		// Upstream: FromID calls changed symbol ToID
		if changedSymIDs[rel.ToID] {
			upstreamAffected[rel.FromID] = true
			if sym := idx.Symbol(rel.FromID); sym != nil {
				upstreamAffected[sym.File] = true
			}
		}
	}

	// Also propagate upstream dependencies through file graph
	// Build reverse file dependency map
	for _, rel := range idx.Rel {
		fromSym := idx.Symbol(rel.FromID)
		toSym := idx.Symbol(rel.ToID)
		if fromSym != nil && toSym != nil {
			if changedFileSet[toSym.File] && fromSym.File != toSym.File {
				upstreamAffected[fromSym.File] = true
			}
		}
	}

	upstreamList := []string{}
	for item := range upstreamAffected {
		upstreamList = append(upstreamList, item)
	}
	report.UpstreamAffected = upstreamList

	downstreamList := []string{}
	for item := range downstreamCalls {
		downstreamList = append(downstreamList, item)
	}
	report.DownstreamCalls = downstreamList

	// Identify affected API endpoints
	apiEndpoints := identifyAffectedAPIs(changedSymbols, upstreamList, idx)
	report.AffectedEndpoints = apiEndpoints

	// Identify potentially affected tests
	affectedTests := identifyAffectedTests(changedFilePaths, upstreamList, files)
	report.AffectedTests = affectedTests

	// Calculate Risk Score (0-100)
	score, level := calculateRiskScore(changedSymbols, upstreamList, apiEndpoints, affectedTests)
	report.RiskScore = score
	report.RiskLevel = level

	report.Summary = fmt.Sprintf(
		"Impact Analysis: %d changed files, %d modified symbols. Blast radius: %d upstream dependents, %d downstream calls, %d affected APIs, %d relevant tests.",
		len(changedFilePaths), len(changedSymbols), len(upstreamList), len(downstreamList), len(apiEndpoints), len(affectedTests),
	)

	return report
}

func identifyAffectedAPIs(changedSyms []model.ChangedSymbol, upstream []string, idx *analysis.Index) []string {
	apiSet := make(map[string]bool)

	checkAPI := func(name, file string) bool {
		lowerName := strings.ToLower(name)
		lowerFile := strings.ToLower(file)
		isAPIFile := strings.Contains(lowerFile, "api") || strings.Contains(lowerFile, "controller") ||
			strings.Contains(lowerFile, "handler") || strings.Contains(lowerFile, "route") || strings.Contains(lowerFile, "endpoint")
		isAPIName := strings.HasPrefix(lowerName, "get") || strings.HasPrefix(lowerName, "post") ||
			strings.HasPrefix(lowerName, "put") || strings.HasPrefix(lowerName, "delete") ||
			strings.Contains(lowerName, "handler") || strings.Contains(lowerName, "servehttp")
		return isAPIFile || isAPIName
	}

	for _, cs := range changedSyms {
		if checkAPI(cs.Symbol.Name, cs.Symbol.File) {
			apiSet[cs.Symbol.QualifiedName+" ("+cs.Symbol.File+")"] = true
		}
	}

	for _, u := range upstream {
		if sym := idx.Symbol(u); sym != nil {
			if checkAPI(sym.Name, sym.File) {
				apiSet[sym.QualifiedName+" ("+sym.File+")"] = true
			}
		}
	}

	apis := []string{}
	for a := range apiSet {
		apis = append(apis, a)
	}
	return apis
}

func identifyAffectedTests(changedFiles []string, upstream []string, files []model.FileNode) []string {
	testSet := make(map[string]bool)

	isTest := func(p string) bool {
		lp := strings.ToLower(p)
		base := strings.ToLower(filepath.Base(p))
		return strings.HasSuffix(base, "_test.go") || strings.HasSuffix(base, "_test.py") ||
			strings.HasSuffix(base, ".test.ts") || strings.HasSuffix(base, ".test.js") ||
			strings.HasSuffix(base, ".spec.ts") || strings.HasSuffix(base, ".spec.js") ||
			strings.Contains(lp, "/test/") || strings.Contains(lp, "/tests/")
	}

	for _, f := range files {
		if !isTest(f.Path) {
			continue
		}

		// Direct test file match
		for _, cf := range changedFiles {
			cfBase := strings.TrimSuffix(filepath.Base(cf), filepath.Ext(cf))
			if strings.Contains(f.Path, cfBase) {
				testSet[f.Path] = true
			}
		}

		// Upstream test match
		for _, u := range upstream {
			uBase := strings.TrimSuffix(filepath.Base(u), filepath.Ext(u))
			if strings.Contains(f.Path, uBase) {
				testSet[f.Path] = true
			}
		}
	}

	tests := []string{}
	for t := range testSet {
		tests = append(tests, t)
	}
	return tests
}

func calculateRiskScore(changedSyms []model.ChangedSymbol, upstream []string, apis []string, tests []string) (float64, model.Severity) {
	// Base weight components:
	// 1. Complexity of changed symbols
	maxComplexity := 0
	totalComplexity := 0
	for _, cs := range changedSyms {
		totalComplexity += cs.Symbol.Complexity
		if cs.Symbol.Complexity > maxComplexity {
			maxComplexity = cs.Symbol.Complexity
		}
	}

	// Blast radius: upstream callers
	blastRadius := len(upstream)

	// API exposure: affected endpoints add significant risk
	apiCount := len(apis)

	// Test mitigation: presence of tests mitigates risk
	testCount := len(tests)

	rawScore := (float64(totalComplexity) * 2.5) + (float64(blastRadius) * 5.0) + (float64(apiCount) * 12.0)
	if testCount > 0 {
		mitigation := math.Min(float64(testCount)*4.0, 30.0)
		rawScore = math.Max(0, rawScore-mitigation)
	}

	score := math.Min(100.0, rawScore)
	score = math.Round(score*10) / 10

	var level model.Severity
	if score >= 60.0 {
		level = model.SevHigh
	} else if score >= 25.0 {
		level = model.SevMedium
	} else if score > 0 {
		level = model.SevLow
	} else {
		level = model.SevInfo
	}

	return score, level
}
