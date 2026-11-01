package analysis

import (
	"testing"

	"github.com/codedna/codedna/internal/model"
)

func TestAnalysisTarjanCycles(t *testing.T) {
	// A -> B -> C -> A (cycle of 3)
	// D -> E (no cycle)
	edges := map[string][]string{
		"a.go": {"b.go"},
		"b.go": {"c.go"},
		"c.go": {"a.go"},
		"d.go": {"e.go"},
		"e.go": {},
	}
	nodes := []string{"a.go", "b.go", "c.go", "d.go", "e.go"}

	sccs := tarjanSCC(nodes, edges)
	var cycles [][]string
	for _, scc := range sccs {
		if len(scc) > 1 {
			cycles = append(cycles, scc)
		}
	}

	if len(cycles) != 1 {
		t.Fatalf("Expected 1 cycle, got %d", len(cycles))
	}
	if len(cycles[0]) != 3 {
		t.Errorf("Expected cycle of 3 nodes, got %d", len(cycles[0]))
	}
}

func TestAnalysisReachability(t *testing.T) {
	// A -> B -> C
	// D -> B
	g := &FileGraph{
		Nodes: []string{"a.go", "b.go", "c.go", "d.go"},
		Edges: map[string][]string{
			"a.go": {"b.go"},
			"b.go": {"c.go"},
			"c.go": {},
			"d.go": {"b.go"},
		},
	}

	down := g.Downstream("a.go")
	if len(down) != 2 {
		t.Errorf("Expected 2 downstream nodes from a.go, got %v", down)
	}

	up := g.Upstream("c.go")
	if len(up) != 3 { // b.go, a.go, d.go
		t.Errorf("Expected 3 upstream nodes leading to c.go, got %v", up)
	}
}

func TestIndexResolutionAndDeadCode(t *testing.T) {
	results := []*model.FileParseResult{
		{
			File:     "service.go",
			Language: "go",
			Symbols: []model.Symbol{
				{
					ID:            "service.go:Process",
					Name:          "Process",
					QualifiedName: "Process",
					Kind:          model.SymbolFunction,
					File:          "service.go",
					Exported:      true,
					Complexity:    18,
					Line:          10,
				},
				{
					ID:            "service.go:UnusedHelper",
					Name:          "UnusedHelper",
					QualifiedName: "UnusedHelper",
					Kind:          model.SymbolFunction,
					File:          "service.go",
					Exported:      false,
					Complexity:    1,
					Line:          30,
				},
			},
			Calls: []model.Call{
				{
					Callee:     "Helper",
					Line:       15,
					InSymbolID: "service.go:Process",
				},
			},
		},
		{
			File:     "helper.go",
			Language: "go",
			Symbols: []model.Symbol{
				{
					ID:            "helper.go:Helper",
					Name:          "Helper",
					QualifiedName: "Helper",
					Kind:          model.SymbolFunction,
					File:          "helper.go",
					Exported:      true,
					Complexity:    1,
					Line:          5,
				},
			},
		},
	}

	files := []model.FileNode{
		{Path: "service.go", Language: "go"},
		{Path: "helper.go", Language: "go"},
	}

	idx := Build(results, files)
	rels := idx.Resolve(results)

	if len(rels) != 1 {
		t.Fatalf("Expected 1 resolved relationship, got %d", len(rels))
	}
	if rels[0].ToID != "helper.go:Helper" {
		t.Errorf("Expected call to helper.go:Helper, got %s", rels[0].ToID)
	}
	if rels[0].Confidence < 0.8 {
		t.Errorf("Expected high confidence for exact unique resolution, got %f", rels[0].Confidence)
	}

	// Dead code detection: UnusedHelper should be detected as dead code candidate
	dead := idx.DeadCodeCandidates()
	foundDead := false
	for _, issue := range dead {
		if issue.SymbolID == "service.go:UnusedHelper" {
			foundDead = true
			break
		}
	}
	if !foundDead {
		t.Errorf("Expected UnusedHelper to be detected as dead code candidate")
	}

	// Hotspots
	hotspots := idx.Hotspots(5)
	if len(hotspots) == 0 {
		t.Errorf("Expected hotspots to return symbols")
	}
}
