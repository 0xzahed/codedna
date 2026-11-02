package sysdesign

import (
	"testing"

	"github.com/codedna/codedna/internal/model"
)

func TestClassifyTiers(t *testing.T) {
	files := []model.FileNode{
		{Path: "web/src/pages/index.tsx"},
		{Path: "internal/api/handler.go"},
		{Path: "internal/service/user.go"},
		{Path: "internal/store/repository.go"},
		{Path: "internal/worker/queue.go"},
		{Path: "internal/client/stripe.go"},
	}

	syms := []model.Symbol{
		{ID: "sym1", Name: "UserHandler", QualifiedName: "api.UserHandler", File: "internal/api/handler.go", Kind: model.SymbolFunction},
		{ID: "sym2", Name: "UserService", QualifiedName: "service.UserService", File: "internal/service/user.go", Kind: model.SymbolStruct},
		{ID: "sym3", Name: "UserRepository", QualifiedName: "store.UserRepository", File: "internal/store/repository.go", Kind: model.SymbolStruct},
		{ID: "sym4", Name: "QueueWorker", QualifiedName: "worker.QueueWorker", File: "internal/worker/queue.go", Kind: model.SymbolFunction},
	}

	tiers := ClassifyTiers(files, syms)
	if len(tiers) != 6 {
		t.Fatalf("expected 6 tiers, got %d", len(tiers))
	}

	tierMap := make(map[model.TierKind]model.SystemTier)
	for _, tier := range tiers {
		tierMap[tier.ID] = tier
	}

	if tierMap[model.TierClient].FileCount != 1 {
		t.Errorf("expected 1 client file, got %d", tierMap[model.TierClient].FileCount)
	}
	if tierMap[model.TierAPI].FileCount != 1 {
		t.Errorf("expected 1 api file, got %d", tierMap[model.TierAPI].FileCount)
	}
	if tierMap[model.TierService].FileCount != 1 {
		t.Errorf("expected 1 service file, got %d", tierMap[model.TierService].FileCount)
	}
	if tierMap[model.TierStorage].FileCount != 1 {
		t.Errorf("expected 1 storage file, got %d", tierMap[model.TierStorage].FileCount)
	}
	if tierMap[model.TierMessaging].FileCount != 1 {
		t.Errorf("expected 1 messaging file, got %d", tierMap[model.TierMessaging].FileCount)
	}
	if tierMap[model.TierIntegration].FileCount != 1 {
		t.Errorf("expected 1 integration file, got %d", tierMap[model.TierIntegration].FileCount)
	}
}

func TestComputeDataFlow(t *testing.T) {
	syms := []model.Symbol{
		{ID: "sym_api", Name: "Handler", File: "internal/api/handler.go"},
		{ID: "sym_svc", Name: "Service", File: "internal/service/service.go"},
		{ID: "sym_store", Name: "Store", File: "internal/store/db.go"},
	}

	rels := []model.Relationship{
		{FromID: "sym_api", ToID: "sym_svc", Method: "ProcessRequest"},
		{FromID: "sym_api", ToID: "sym_svc", Method: "Validate"},
		{FromID: "sym_svc", ToID: "sym_store", Method: "QueryUser"},
	}

	hops := ComputeDataFlow(rels, syms)
	if len(hops) != 2 {
		t.Fatalf("expected 2 hops, got %d", len(hops))
	}

	// First hop should have higher call count (2 calls)
	if hops[0].FromTier != model.TierAPI || hops[0].ToTier != model.TierService {
		t.Errorf("expected API -> Service hop first, got %s -> %s", hops[0].FromTier, hops[0].ToTier)
	}
	if hops[0].CallCount != 2 {
		t.Errorf("expected 2 calls, got %d", hops[0].CallCount)
	}
	if hops[0].Protocol != "In-Memory Orchestration" {
		t.Errorf("expected In-Memory Orchestration protocol, got %s", hops[0].Protocol)
	}

	if hops[1].FromTier != model.TierService || hops[1].ToTier != model.TierStorage {
		t.Errorf("expected Service -> Storage hop second, got %s -> %s", hops[1].FromTier, hops[1].ToTier)
	}
	if hops[1].Protocol != "SQL / Storage Interface" {
		t.Errorf("expected SQL / Storage Interface protocol, got %s", hops[1].Protocol)
	}
}

func TestDetectPatterns(t *testing.T) {
	files := []model.FileNode{
		{Path: "internal/store/user_repo.go"},
		{Path: "internal/events/queue_watcher.go"},
		{Path: "internal/api/middleware.go"},
		{Path: "internal/service/auth.go"},
	}

	syms := []model.Symbol{
		{ID: "s1", Name: "NewUserRepository", QualifiedName: "store.NewUserRepository", Kind: model.SymbolFunction, File: "internal/store/user_repo.go"},
		{ID: "s2", Name: "NewAuthService", QualifiedName: "service.NewAuthService", Kind: model.SymbolFunction, File: "internal/service/auth.go"},
		{ID: "s3", Name: "UserRepo", QualifiedName: "store.UserRepo", Kind: model.SymbolStruct, File: "internal/store/user_repo.go"},
		{ID: "s4", Name: "PublishEvent", QualifiedName: "events.PublishEvent", Kind: model.SymbolFunction, File: "internal/events/queue_watcher.go"},
		{ID: "s5", Name: "AuthMiddleware", QualifiedName: "api.AuthMiddleware", Kind: model.SymbolFunction, File: "internal/api/middleware.go"},
		{ID: "s6", Name: "defaultInstance", QualifiedName: "service.defaultInstance", Kind: model.SymbolVariable, File: "internal/service/auth.go"},
		{ID: "s7", Name: "ExecuteCommand", QualifiedName: "service.ExecuteCommand", Kind: model.SymbolFunction, File: "internal/service/auth.go"},
		{ID: "s8", Name: "QueryUser", QualifiedName: "service.QueryUser", Kind: model.SymbolFunction, File: "internal/service/auth.go"},
	}

	rels := []model.Relationship{
		{FromID: "s5", ToID: "s2", Method: "Call"},
	}

	patterns := DetectPatterns(files, syms, rels)
	if len(patterns) == 0 {
		t.Fatal("expected at least one pattern detected")
	}

	patternIDs := make(map[string]bool)
	for _, p := range patterns {
		patternIDs[p.ID] = true
	}

	expectedPatterns := []string{
		"pattern-repository",
		"pattern-factory",
		"pattern-singleton",
		"pattern-event-driven",
		"pattern-middleware",
		"pattern-clean-architecture",
		"pattern-cqrs",
	}

	for _, ep := range expectedPatterns {
		if !patternIDs[ep] {
			t.Errorf("expected pattern %s to be detected", ep)
		}
	}
}

func TestAnalyzeSystemDesign(t *testing.T) {
	files := []model.FileNode{
		{Path: "cmd/codedna/main.go"},
		{Path: "internal/api/server.go"},
		{Path: "internal/service/analyzer.go"},
		{Path: "internal/store/sqlite.go"},
	}

	syms := []model.Symbol{
		{ID: "s_main", Name: "main", QualifiedName: "main.main", File: "cmd/codedna/main.go", Kind: model.SymbolFunction},
		{ID: "s_api", Name: "StartServer", QualifiedName: "api.StartServer", File: "internal/api/server.go", Kind: model.SymbolFunction},
		{ID: "s_svc", Name: "NewAnalyzer", QualifiedName: "service.NewAnalyzer", File: "internal/service/analyzer.go", Kind: model.SymbolFunction},
		{ID: "s_db", Name: "NewStore", QualifiedName: "store.NewStore", File: "internal/store/sqlite.go", Kind: model.SymbolFunction},
	}

	rels := []model.Relationship{
		{FromID: "s_api", ToID: "s_svc", Method: "Analyze"},
		{FromID: "s_svc", ToID: "s_db", Method: "Save"},
	}

	report := AnalyzeSystemDesign(files, syms, rels)

	if report.ArchitectureStyle == "" {
		t.Error("expected non-empty architecture style")
	}
	if len(report.Tiers) != 6 {
		t.Errorf("expected 6 tiers, got %d", len(report.Tiers))
	}
	if len(report.DataFlow) != 2 {
		t.Errorf("expected 2 data flow hops, got %d", len(report.DataFlow))
	}
	if len(report.KeyEntrypoints) == 0 {
		t.Error("expected key entrypoints to be identified")
	}
	if report.ReliabilityScore <= 0 || report.ReliabilityScore > 100 {
		t.Errorf("expected reliability score between 0 and 100, got %f", report.ReliabilityScore)
	}
	if len(report.Recommendations) == 0 {
		t.Error("expected architectural recommendations to be generated")
	}
}
