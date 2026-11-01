package impact

import (
	"testing"

	"github.com/codedna/codedna/internal/analysis"
	"github.com/codedna/codedna/internal/model"
)

func TestParseDiffOutput(t *testing.T) {
	diffText := `diff --git a/src/service.go b/src/service.go
index 1234567..89abcdef 100644
--- a/src/service.go
+++ b/src/service.go
@@ -10,3 +10,4 @@ func Process() {
+	newStep()
diff --git a/src/utils.py b/src/utils.py
new file mode 100644
index 0000000..abcdef1
--- /dev/null
+++ b/src/utils.py
@@ -0,0 +1,5 @@
+def helper():
+    pass
`
	diffs := ParseDiffOutput(diffText)
	if len(diffs) != 2 {
		t.Fatalf("Expected 2 diffs, got %d", len(diffs))
	}
	if diffs[0].Path != "src/service.go" || diffs[0].Status != "modified" {
		t.Errorf("Unexpected diff 0: %+v", diffs[0])
	}
	if len(diffs[0].ChangedLines) != 4 || diffs[0].ChangedLines[0] != 10 {
		t.Errorf("Unexpected changed lines: %+v", diffs[0].ChangedLines)
	}
	if diffs[1].Path != "src/utils.py" || diffs[1].Status != "added" {
		t.Errorf("Unexpected diff 1: %+v", diffs[1])
	}
}

func TestAnalyzeImpact(t *testing.T) {
	syms := []model.Symbol{
		{
			ID:            "payment/process.go:ProcessPayment",
			Name:          "ProcessPayment",
			QualifiedName: "ProcessPayment",
			Kind:          model.SymbolFunction,
			File:          "payment/process.go",
			Line:          10,
			EndLine:       20,
			Complexity:    8,
		},
		{
			ID:            "api/handler.go:PaymentHandler",
			Name:          "PaymentHandler",
			QualifiedName: "PaymentHandler",
			Kind:          model.SymbolFunction,
			File:          "api/handler.go",
			Line:          15,
			EndLine:       30,
			Complexity:    2,
		},
	}

	rels := []model.Relationship{
		{
			Kind:       model.RelCalls,
			FromID:     "api/handler.go:PaymentHandler",
			ToID:       "payment/process.go:ProcessPayment",
			Confidence: 0.95,
		},
	}

	files := []model.FileNode{
		{Path: "payment/process.go", Language: "go"},
		{Path: "api/handler.go", Language: "go"},
		{Path: "payment/process_test.go", Language: "go", Tags: []string{"test"}},
	}

	idx := &analysis.Index{
		Symbols: syms,
		Rel:     rels,
	}

	diffs := []FileDiff{
		{
			Path:         "payment/process.go",
			Status:       "modified",
			ChangedLines: []int{12, 13},
		},
	}

	report := AnalyzeImpact(".", "main", diffs, idx, files)

	if len(report.ChangedSymbols) != 1 {
		t.Fatalf("Expected 1 changed symbol, got %d", len(report.ChangedSymbols))
	}
	if report.ChangedSymbols[0].Symbol.Name != "ProcessPayment" {
		t.Errorf("Expected ProcessPayment changed, got %s", report.ChangedSymbols[0].Symbol.Name)
	}

	// Verify upstream blast radius (PaymentHandler)
	hasHandler := false
	for _, u := range report.UpstreamAffected {
		if u == "api/handler.go:PaymentHandler" || u == "api/handler.go" {
			hasHandler = true
			break
		}
	}
	if !hasHandler {
		t.Errorf("Expected api/handler.go to be in upstream affected, got %v", report.UpstreamAffected)
	}

	// Verify affected API detection
	if len(report.AffectedEndpoints) == 0 {
		t.Errorf("Expected affected endpoints to include PaymentHandler, got %v", report.AffectedEndpoints)
	}

	// Verify affected test detection
	if len(report.AffectedTests) == 0 {
		t.Errorf("Expected payment/process_test.go to be detected, got %v", report.AffectedTests)
	}

	// Risk score should be calculated
	if report.RiskScore <= 0 {
		t.Errorf("Expected positive risk score, got %f", report.RiskScore)
	}
}
