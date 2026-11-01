package architecture

import (
	"testing"

	"github.com/codedna/codedna/internal/model"
)

func TestArchitectureRulesViolation(t *testing.T) {
	rules := []model.ArchRule{
		{
			ID:          "test-rule-domain-web",
			Name:        "Domain Web Isolation",
			FromPattern: "**/domain/**",
			ToPattern:   "**/controllers/**",
			Forbidden:   true,
			Description: "Domain must not call controllers",
			Severity:    model.SevHigh,
		},
	}

	syms := []model.Symbol{
		{
			ID:   "src/domain/user.go:Validate",
			File: "src/domain/user.go",
			Name: "Validate",
		},
		{
			ID:   "src/controllers/auth.go:Login",
			File: "src/controllers/auth.go",
			Name: "Login",
		},
	}

	// Violation: domain user calls controller Login
	rels := []model.Relationship{
		{
			Kind:   model.RelCalls,
			FromID: "src/domain/user.go:Validate",
			ToID:   "src/controllers/auth.go:Login",
			Locations: []model.Ref{
				{File: "src/domain/user.go", Line: 42, Col: 5},
			},
		},
	}

	report := CheckArchitecture(rules, rels, syms, nil)

	if len(report.Violations) != 1 {
		t.Fatalf("Expected 1 violation, got %d", len(report.Violations))
	}
	if report.Violations[0].RuleID != "test-rule-domain-web" {
		t.Errorf("Expected violation of test-rule-domain-web, got %s", report.Violations[0].RuleID)
	}
	if report.Violations[0].Line != 42 {
		t.Errorf("Expected line 42, got %d", report.Violations[0].Line)
	}
	if report.LayerHealthScore >= 100.0 {
		t.Errorf("Expected layer health penalty, got %f", report.LayerHealthScore)
	}
}

func TestPatternMatching(t *testing.T) {
	if !matchPattern("**/domain/**", "src/domain/entity/user.go") {
		t.Error("Expected match for **/domain/**")
	}
	if matchPattern("**/domain/**", "src/controllers/user.go") {
		t.Error("Expected no match for **/domain/** on controllers")
	}
	if !matchPattern("controllers/*", "controllers/auth.go") {
		t.Error("Expected match for controllers/*")
	}
}
