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

func TestRulePersistenceAndLifecycle(t *testing.T) {
	tempDir := t.TempDir()

	// 1. Initial load should return default rules
	rules, err := LoadRules(tempDir)
	if err != nil {
		t.Fatalf("LoadRules failed: %v", err)
	}
	initialCount := len(rules)
	if initialCount == 0 {
		t.Fatal("Expected default rules, got 0")
	}

	// 2. Add custom rule
	newRule := model.ArchRule{
		Name:        "Services ↛ Controllers",
		FromPattern: "**/services/**",
		ToPattern:   "**/controllers/**",
		Forbidden:   true,
		Description: "Services layer must never invoke controllers",
		Severity:    model.SevHigh,
	}

	updated, err := AddRule(tempDir, newRule)
	if err != nil {
		t.Fatalf("AddRule failed: %v", err)
	}
	if len(updated) != initialCount+1 {
		t.Fatalf("Expected %d rules, got %d", initialCount+1, len(updated))
	}

	addedID := updated[len(updated)-1].ID
	if addedID == "" {
		t.Fatal("Expected auto-generated rule ID")
	}

	// 3. Delete the newly added rule
	afterDel, err := DeleteRule(tempDir, addedID)
	if err != nil {
		t.Fatalf("DeleteRule failed: %v", err)
	}
	if len(afterDel) != initialCount {
		t.Fatalf("Expected %d rules after deletion, got %d", initialCount, len(afterDel))
	}

	// 4. Reset rules
	resetted, err := ResetRules(tempDir)
	if err != nil {
		t.Fatalf("ResetRules failed: %v", err)
	}
	if len(resetted) != len(DefaultRules()) {
		t.Fatalf("Expected %d default rules after reset, got %d", len(DefaultRules()), len(resetted))
	}
}
