package report

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/codedna/codedna/internal/model"
)

func TestReportGeneration(t *testing.T) {
	data := CompleteReportData{
		Repo: model.RepoInfo{
			Name: "test-service",
			Root: "/home/test/service",
			Git: model.GitMeta{
				IsGitRepo:  true,
				Branch:     "main",
				Commit:     "1234567890ab",
				AuthorName: "Developer",
			},
		},
		Scan: model.Scan{
			ID:     1,
			Status: "complete",
		},
		Stats: &model.RepoScanStats{
			FilesParsed:   12,
			LOC:           2400,
			SymbolsFound:  45,
			Relationships: 30,
			ResolvedCalls: 26,
			Languages:     map[string]int{"go": 10, "python": 2},
			LOCByLanguage: map[string]int{"go": 2000, "python": 400},
			Frameworks:    []string{"Gin", "Gorm"},
		},
		Issues: []model.Issue{
			{
				Kind:     model.IssueComplexity,
				Severity: model.SevHigh,
				Title:    "High complexity: ProcessAll",
				File:     "process.go",
				Detail:   map[string]interface{}{"complexity": 22},
			},
		},
		Architecture: &model.ArchHealthReport{
			LayerHealthScore: 92.5,
			RuleCount:        3,
		},
		GeneratedAt: time.Now(),
	}

	data.HealthScore = CalculateHealthScore(data.Stats, data.Issues, data.Architecture)

	// Markdown test
	md := GenerateMarkdown(data)
	if !strings.Contains(md, "test-service") {
		t.Errorf("Expected repo name in markdown report")
	}
	if !strings.Contains(md, "Executive Summary") {
		t.Errorf("Expected Executive Summary section")
	}
	if !strings.Contains(md, "Language Breakdown") {
		t.Errorf("Expected Language Breakdown section")
	}
	if !strings.Contains(md, "ProcessAll") {
		t.Errorf("Expected hotspot ProcessAll in report")
	}

	// JSON test
	js, err := GenerateJSON(data)
	if err != nil {
		t.Fatalf("GenerateJSON failed: %v", err)
	}

	var parsed CompleteReportData
	if err := json.Unmarshal(js, &parsed); err != nil {
		t.Fatalf("Failed to unmarshal JSON report: %v", err)
	}
	if parsed.Repo.Name != "test-service" {
		t.Errorf("Expected test-service, got %s", parsed.Repo.Name)
	}
}
