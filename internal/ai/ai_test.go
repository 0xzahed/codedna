package ai

import (
	"context"
	"strings"
	"testing"

	"github.com/codedna/codedna/internal/model"
)

func TestAIPrivacyProtection(t *testing.T) {
	client := NewClient(Config{
		Provider:      "openai",
		AllowRemoteAI: false, // Default: no remote AI without consent
	})

	req := model.AIQueryRequest{
		Query:       "Explain the architecture",
		Provider:    "openai",
		UserConsent: false, // No explicit consent given
	}

	_, err := client.QueryCodebase(context.Background(), req, nil, nil, nil)
	if err != ErrConsentRequired {
		t.Fatalf("Expected ErrConsentRequired, got %v", err)
	}
}

func TestAIDeterministicSynthesis(t *testing.T) {
	client := NewClient(Config{
		Provider: "offline",
	})

	syms := []model.Symbol{
		{
			ID:            "payment/process.go:ProcessPayment",
			Name:          "ProcessPayment",
			QualifiedName: "ProcessPayment",
			Kind:          model.SymbolFunction,
			File:          "payment/process.go",
			Line:          10,
			Complexity:    6,
		},
	}

	rels := []model.Relationship{
		{
			FromID:     "api/handler.go:CheckoutHandler",
			ToID:       "payment/process.go:ProcessPayment",
			ToName:     "ProcessPayment",
			Confidence: 0.95,
			Method:     model.MethodExactUnique,
		},
	}

	req := model.AIQueryRequest{
		Query: "ProcessPayment",
	}

	resp, err := client.QueryCodebase(context.Background(), req, syms, rels, nil)
	if err != nil {
		t.Fatalf("QueryCodebase failed: %v", err)
	}

	if !strings.Contains(resp.Answer, "ProcessPayment") {
		t.Errorf("Expected answer to contain symbol name, got: %s", resp.Answer)
	}
	if len(resp.SourceRefs) == 0 {
		t.Errorf("Expected source references in response")
	}
	if resp.Provider != "deterministic-engine" {
		t.Errorf("Expected deterministic-engine provider, got %s", resp.Provider)
	}
}

func TestContextGeneration(t *testing.T) {
	syms := []model.Symbol{
		{
			ID:            "server.go:Start",
			Name:          "Start",
			QualifiedName: "Start",
			Kind:          model.SymbolFunction,
			File:          "server.go",
			Line:          15,
			Complexity:    4,
		},
	}

	stats := &model.RepoScanStats{
		FilesParsed:   10,
		LOC:           1500,
		SymbolsFound:  25,
		Relationships: 18,
		Frameworks:    []string{"Next.js", "Express"},
		Languages:     map[string]int{"typescript": 8, "go": 2},
	}

	// Test Cursor export
	cursorCtx := GenerateContext("cursor", "MyProject", nil, syms, nil, stats)
	if !strings.Contains(cursorCtx, "Cursor") || !strings.Contains(cursorCtx, "MyProject") {
		t.Errorf("Unexpected cursor export: %s", cursorCtx)
	}

	// Test Claude Code export
	claudeCtx := GenerateContext("claude", "MyProject", nil, syms, nil, stats)
	if !strings.Contains(claudeCtx, "CLAUDE.md") || !strings.Contains(claudeCtx, "Next.js") {
		t.Errorf("Unexpected claude export: %s", claudeCtx)
	}
}

func TestMapBusinessLogic(t *testing.T) {
	files := []model.FileNode{
		{Path: "src/auth/jwt.go", Language: "go"},
		{Path: "src/auth/login.go", Language: "go"},
		{Path: "src/payment/stripe.go", Language: "go"},
	}

	syms := []model.Symbol{
		{Name: "GenerateToken", QualifiedName: "GenerateToken", File: "src/auth/jwt.go"},
		{Name: "ChargeCard", QualifiedName: "ChargeCard", File: "src/payment/stripe.go"},
	}

	modules := MapBusinessLogic(files, syms)
	if len(modules) < 2 {
		t.Fatalf("Expected at least 2 business logic modules, got %d", len(modules))
	}

	hasAuth := false
	hasPayment := false
	for _, m := range modules {
		if strings.Contains(m.Domain, "Authentication") {
			hasAuth = true
		}
		if strings.Contains(m.Domain, "Payment") {
			hasPayment = true
		}
	}

	if !hasAuth || !hasPayment {
		t.Errorf("Expected Auth and Payment modules, got: %+v", modules)
	}
}
