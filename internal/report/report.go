// Package report generates comprehensive Markdown and JSON software intelligence reports.
package report

import (
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/codedna/codedna/internal/model"
)

// CompleteReportData gathers all intelligence dimensions for reporting.
type CompleteReportData struct {
	Repo         model.RepoInfo           `json:"repo"`
	Scan         model.Scan               `json:"scan"`
	Stats        *model.RepoScanStats     `json:"stats,omitempty"`
	Files        []model.FileNode         `json:"files,omitempty"`
	TopSymbols   []model.Symbol           `json:"topSymbols,omitempty"`
	Issues       []model.Issue            `json:"issues,omitempty"`
	Architecture *model.ArchHealthReport  `json:"architecture,omitempty"`
	Impact       *model.ImpactReport      `json:"impact,omitempty"`
	Runtime      *model.RuntimeSummary    `json:"runtime,omitempty"`
	HealthScore  float64                  `json:"healthScore"`
	GeneratedAt  time.Time                `json:"generatedAt"`
}

// CalculateHealthScore computes an overall codebase health score (0-100).
func CalculateHealthScore(stats *model.RepoScanStats, issues []model.Issue, arch *model.ArchHealthReport) float64 {
	score := 100.0

	// Deduct for circular dependencies
	for _, iss := range issues {
		if iss.Kind == model.IssueCircularDep {
			score -= 10.0
		} else if iss.Kind == model.IssueComplexity {
			switch iss.Severity {
			case model.SevHigh:
				score -= 3.0
			case model.SevMedium:
				score -= 1.5
			}
		} else if iss.Kind == model.IssueDeadCode {
			score -= 0.5
		}
	}

	// Architecture violations
	if arch != nil {
		for _, v := range arch.Violations {
			switch v.Severity {
			case model.SevHigh:
				score -= 5.0
			case model.SevMedium:
				score -= 2.5
			default:
				score -= 1.0
			}
		}
	}

	score = math.Max(0.0, math.Min(100.0, score))
	return math.Round(score*10) / 10
}

// GenerateMarkdown creates an executive Markdown report.
func GenerateMarkdown(data CompleteReportData) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("# CodeDNA Software Intelligence Report: %s\n\n", data.Repo.Name))
	sb.WriteString(fmt.Sprintf("> Generated on %s by CodeDNA 2.0 Engine\n\n", data.GeneratedAt.Format(time.RFC1123)))

	// Executive Summary Table
	sb.WriteString("## 1. Executive Summary\n\n")
	sb.WriteString(fmt.Sprintf("- **Overall Codebase Health Score**: `%.1f / 100`\n", data.HealthScore))
	if data.Repo.Git.IsGitRepo {
		sb.WriteString(fmt.Sprintf("- **Git Branch**: `%s` (commit `%s`)\n", data.Repo.Git.Branch, data.Repo.Git.Commit))
		if data.Repo.Git.LastCommitAt != nil {
			sb.WriteString(fmt.Sprintf("- **Last Commit**: %s by %s\n", data.Repo.Git.LastCommitAt.Format("2006-01-02 15:04:05"), data.Repo.Git.AuthorName))
		}
	}
	if data.Stats != nil {
		sb.WriteString(fmt.Sprintf("- **Total LOC**: `%d` across `%d` parsed files\n", data.Stats.LOC, data.Stats.FilesParsed))
		sb.WriteString(fmt.Sprintf("- **Discovered Code Symbols**: `%d`\n", data.Stats.SymbolsFound))
		sb.WriteString(fmt.Sprintf("- **Resolved Call Relationships**: `%d` (resolved: %d, unresolved: %d)\n", data.Stats.Relationships, data.Stats.ResolvedCalls, data.Stats.UnresolvedCalls))
		if len(data.Stats.Frameworks) > 0 {
			sb.WriteString(fmt.Sprintf("- **Detected Frameworks**: %s\n", strings.Join(data.Stats.Frameworks, ", ")))
		}
	}
	sb.WriteString("\n")

	// Language Breakdown
	if data.Stats != nil && len(data.Stats.Languages) > 0 {
		sb.WriteString("## 2. Language Breakdown\n\n")
		sb.WriteString("| Language | Files | Lines of Code |\n")
		sb.WriteString("| :--- | :--- | :--- |\n")
		for lang, cnt := range data.Stats.Languages {
			loc := data.Stats.LOCByLanguage[lang]
			sb.WriteString(fmt.Sprintf("| **%s** | %d | %d |\n", lang, cnt, loc))
		}
		sb.WriteString("\n")
	}

	// Circular Dependencies
	var cycles []model.Issue
	for _, iss := range data.Issues {
		if iss.Kind == model.IssueCircularDep {
			cycles = append(cycles, iss)
		}
	}
	sb.WriteString("## 3. Circular Dependencies\n\n")
	if len(cycles) == 0 {
		sb.WriteString("No circular dependencies detected between modules.\n\n")
	} else {
		sb.WriteString(fmt.Sprintf("Found **%d circular dependency chain(s)**:\n\n", len(cycles)))
		for i, c := range cycles {
			sb.WriteString(fmt.Sprintf("%d. **%s**\n", i+1, c.Title))
			if dt, ok := c.Detail.(map[string]interface{}); ok {
				if chain, ok := dt["chain"].(string); ok {
					sb.WriteString(fmt.Sprintf("   - Cycle: `%s`\n", chain))
				}
			}
		}
		sb.WriteString("\n")
	}

	// Complexity Hotspots
	var hotspots []model.Issue
	for _, iss := range data.Issues {
		if iss.Kind == model.IssueComplexity {
			hotspots = append(hotspots, iss)
		}
	}
	sb.WriteString("## 4. Complexity Hotspots\n\n")
	if len(hotspots) == 0 {
		sb.WriteString("No extreme cyclomatic complexity hotspots detected (threshold >= 15).\n\n")
	} else {
		sb.WriteString("| Severity | Symbol | File | Complexity |\n")
		sb.WriteString("| :--- | :--- | :--- | :--- |\n")
		for _, h := range hotspots {
			comp := 0
			if dt, ok := h.Detail.(map[string]interface{}); ok {
				if cVal, ok := dt["complexity"].(int); ok {
					comp = cVal
				}
			}
			sb.WriteString(fmt.Sprintf("| %s | `%s` | `%s` | %d |\n", h.Severity, h.Title, h.File, comp))
		}
		sb.WriteString("\n")
	}

	// Architecture Health
	if data.Architecture != nil {
		sb.WriteString("## 5. Architectural Health & Layer Boundaries\n\n")
		sb.WriteString(fmt.Sprintf("- **Layer Health Score**: `%.1f / 100`\n", data.Architecture.LayerHealthScore))
		sb.WriteString(fmt.Sprintf("- **Enforced Architecture Rules**: `%d`\n", data.Architecture.RuleCount))
		if len(data.Architecture.Violations) == 0 {
			sb.WriteString("- **Boundary Violations**: None detected\n\n")
		} else {
			sb.WriteString(fmt.Sprintf("- **Boundary Violations**: `%d` detected\n\n", len(data.Architecture.Violations)))
			sb.WriteString("| Rule | Severity | From | To | Reason |\n")
			sb.WriteString("| :--- | :--- | :--- | :--- | :--- |\n")
			for _, v := range data.Architecture.Violations {
				sb.WriteString(fmt.Sprintf("| %s | %s | `%s:%d` | `%s` | %s |\n", v.RuleName, v.Severity, v.FromFile, v.Line, v.ToFile, v.Reason))
			}
			sb.WriteString("\n")
		}
	}

	// Change Impact
	if data.Impact != nil && len(data.Impact.ChangedFiles) > 0 {
		sb.WriteString("## 6. Change Impact Analysis\n\n")
		sb.WriteString(fmt.Sprintf("- **Base Ref**: `%s`\n", data.Impact.BaseRef))
		sb.WriteString(fmt.Sprintf("- **Risk Score**: `%.1f / 100` (%s)\n", data.Impact.RiskScore, data.Impact.RiskLevel))
		sb.WriteString(fmt.Sprintf("- **Summary**: %s\n", data.Impact.Summary))
		if len(data.Impact.AffectedEndpoints) > 0 {
			sb.WriteString(fmt.Sprintf("- **Affected APIs**: %s\n", strings.Join(data.Impact.AffectedEndpoints, ", ")))
		}
		if len(data.Impact.AffectedTests) > 0 {
			sb.WriteString(fmt.Sprintf("- **Relevant Tests**: %s\n", strings.Join(data.Impact.AffectedTests, ", ")))
		}
		sb.WriteString("\n")
	}

	// Runtime Behavior
	if data.Runtime != nil && data.Runtime.TotalEvents > 0 {
		sb.WriteString("## 7. Observed Runtime Behavior\n\n")
		sb.WriteString(fmt.Sprintf("- **Captured Execution Events**: `%d`\n", data.Runtime.TotalEvents))
		sb.WriteString(fmt.Sprintf("- **Unique Functions Hit**: `%d`\n", data.Runtime.UniqueFunctions))
		sb.WriteString(fmt.Sprintf("- **Average Function Duration**: `%.2f µs`\n", data.Runtime.AvgDurationUs))
		sb.WriteString(fmt.Sprintf("- **Static vs Runtime Graph Coverage**: `%.1f%%` of static edges verified at runtime\n\n", data.Runtime.Comparison.CoveragePercentage))
	}

	return sb.String()
}

// GenerateJSON creates structured JSON output.
func GenerateJSON(data CompleteReportData) ([]byte, error) {
	return json.MarshalIndent(data, "", "  ")
}
