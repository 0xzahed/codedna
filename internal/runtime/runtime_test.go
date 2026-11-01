package runtime

import (
	"testing"

	"github.com/codedna/codedna/internal/model"
)

func TestRuntimeIngestAndComparison(t *testing.T) {
	events := []model.RuntimeTraceEvent{
		{
			RunID:          "run-1",
			FunctionName:   "handlePayment",
			File:           "payment/handler.go",
			Line:           20,
			DurationUs:     1500,
			TimestampNs:    1000,
			CallerFunction: "router",
			CallerFile:     "server.go",
		},
		{
			RunID:          "run-1",
			FunctionName:   "validateCard",
			File:           "payment/validator.go",
			Line:           10,
			DurationUs:     400,
			TimestampNs:    2000,
			CallerFunction: "handlePayment",
			CallerFile:     "payment/handler.go",
		},
		{
			RunID:          "run-1",
			FunctionName:   "handlePayment",
			File:           "payment/handler.go",
			Line:           20,
			DurationUs:     1200,
			TimestampNs:    3000,
			CallerFunction: "router",
			CallerFile:     "server.go",
		},
	}

	staticRels := []model.Relationship{
		{
			Kind:   model.RelCalls,
			FromID: "payment/handler.go:handlePayment",
			ToID:   "payment/validator.go:validateCard",
			ToName: "validateCard",
		},
		{
			Kind:   model.RelCalls,
			FromID: "payment/handler.go:handlePayment",
			ToID:   "payment/db.go:saveTransaction",
			ToName: "saveTransaction",
		},
	}

	summary := IngestTraces(events, staticRels)

	if summary.TotalEvents != 3 {
		t.Errorf("Expected 3 events, got %d", summary.TotalEvents)
	}
	if summary.UniqueFunctions != 2 {
		t.Errorf("Expected 2 unique functions, got %d", summary.UniqueFunctions)
	}

	// Frequent functions check
	if len(summary.FrequentFunctions) == 0 {
		t.Fatal("Expected frequent functions")
	}
	if summary.FrequentFunctions[0].Function != "handlePayment" || summary.FrequentFunctions[0].CallCount != 2 {
		t.Errorf("Expected handlePayment with 2 calls, got %+v", summary.FrequentFunctions[0])
	}

	// Execution paths check
	if len(summary.ExecutionPaths) != 1 {
		t.Fatalf("Expected 1 execution path for run-1, got %d", len(summary.ExecutionPaths))
	}
	if len(summary.ExecutionPaths[0].Steps) != 3 {
		t.Errorf("Expected 3 steps in path, got %d", len(summary.ExecutionPaths[0].Steps))
	}

	// Static vs runtime comparison check
	comp := summary.Comparison
	if comp.StaticEdgeCount != 2 {
		t.Errorf("Expected 2 static edges, got %d", comp.StaticEdgeCount)
	}
	if comp.ObservedStaticEdges != 1 { // handlePayment -> validateCard was observed
		t.Errorf("Expected 1 observed static edge, got %d", comp.ObservedStaticEdges)
	}
	if comp.UnexecutedStaticEdges != 1 { // saveTransaction was not observed
		t.Errorf("Expected 1 unexecuted static edge, got %d", comp.UnexecutedStaticEdges)
	}
	if comp.CoveragePercentage != 50.0 {
		t.Errorf("Expected 50%% coverage, got %f", comp.CoveragePercentage)
	}
}
