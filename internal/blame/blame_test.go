package blame

import (
	"testing"

	"github.com/codedna/codedna/internal/model"
)

func TestParseGitLogOutputMultiAuthor(t *testing.T) {
	rawLog := `AUTHOR|Alice Smith|alice@example.com
100	10	src/auth/jwt.go
50	5	src/auth/login.go

AUTHOR|Bob Jones|bob@example.com
30	2	src/auth/jwt.go
120	15	src/payment/stripe.go

AUTHOR|Charlie Brown|charlie@example.com
10	1	src/payment/stripe.go
80	0	src/utils/helpers.go
`

	report, err := parseGitLogOutput([]byte(rawLog))
	if err != nil {
		t.Fatalf("parseGitLogOutput failed: %v", err)
	}

	if !report.IsGitRepo {
		t.Error("Expected IsGitRepo to be true")
	}

	if report.TotalAuthors != 3 {
		t.Errorf("Expected 3 authors, got %d", report.TotalAuthors)
	}

	if report.TotalCommits != 3 {
		t.Errorf("Expected 3 commits, got %d", report.TotalCommits)
	}

	// Verify Top Contributors ordering by lines added
	if len(report.TopContributors) != 3 {
		t.Fatalf("Expected 3 top contributors, got %d", len(report.TopContributors))
	}

	topAuthor := report.TopContributors[0]
	if topAuthor.Name != "Alice Smith" && topAuthor.Name != "Bob Jones" {
		t.Errorf("Expected Alice or Bob as top contributor, got %s", topAuthor.Name)
	}

	// Verify File Ownership
	var jwtFile *model.FileOwnership
	var helpersFile *model.FileOwnership
	for i := range report.Files {
		if report.Files[i].FilePath == "src/auth/jwt.go" {
			jwtFile = &report.Files[i]
		}
		if report.Files[i].FilePath == "src/utils/helpers.go" {
			helpersFile = &report.Files[i]
		}
	}

	if jwtFile == nil {
		t.Fatal("Expected src/auth/jwt.go in files")
	}
	if jwtFile.PrimaryAuthor != "Alice Smith" {
		t.Errorf("Expected Alice as primary author of jwt.go, got %s", jwtFile.PrimaryAuthor)
	}
	if jwtFile.ContributorsCount != 2 {
		t.Errorf("Expected 2 contributors for jwt.go, got %d", jwtFile.ContributorsCount)
	}

	if helpersFile == nil {
		t.Fatal("Expected src/utils/helpers.go in files")
	}
	if helpersFile.BusFactorRisk != model.SevHigh {
		t.Errorf("Expected single-contributor file to have SevHigh bus factor risk, got %s", helpersFile.BusFactorRisk)
	}
}

func TestAnalyzeKnowledgeNonGitFallback(t *testing.T) {
	tempDir := t.TempDir()
	rep, err := AnalyzeKnowledge(tempDir)
	if err != nil {
		t.Fatalf("AnalyzeKnowledge error on empty dir: %v", err)
	}
	if rep.IsGitRepo {
		t.Error("Expected IsGitRepo false on empty non-git dir")
	}
	if rep.TotalAuthors != 0 {
		t.Errorf("Expected 0 authors, got %d", rep.TotalAuthors)
	}
}
