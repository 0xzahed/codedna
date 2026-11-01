package report

import (
	"strings"
	"testing"

	"github.com/codedna/codedna/internal/model"
)

func TestGenerateMermaid(t *testing.T) {
	graph := model.GraphData{
		Nodes: []model.GraphNode{
			{ID: "file:api/views.py", Label: "views.py", File: "api/views.py", Type: "file"},
			{ID: "file:api/models.py", Label: "models.py", File: "api/models.py", Type: "file"},
		},
		Edges: []model.GraphEdge{
			{ID: "e1", Source: "file:api/views.py", Target: "file:api/models.py", Kind: "imports"},
		},
	}

	mmd := GenerateMermaid("test-repo", graph)
	if !strings.Contains(mmd, "flowchart TD") {
		t.Fatalf("expected flowchart TD, got: %s", mmd)
	}
	if !strings.Contains(mmd, "views.py") {
		t.Fatalf("expected views.py in output, got: %s", mmd)
	}
	if !strings.Contains(mmd, "imports") {
		t.Fatalf("expected imports edge label in output, got: %s", mmd)
	}
}

func TestGenerateC4(t *testing.T) {
	graph := model.GraphData{
		Nodes: []model.GraphNode{
			{ID: "file:controllers/order.go", Label: "order.go", File: "controllers/order.go", Type: "file", Language: "go"},
			{ID: "file:models/order.go", Label: "order_model.go", File: "models/order.go", Type: "file", Language: "go"},
		},
		Edges: []model.GraphEdge{
			{ID: "e1", Source: "file:controllers/order.go", Target: "file:models/order.go", Kind: "calls"},
		},
	}

	c4 := GenerateC4("ecommerce-service", graph)
	if !strings.Contains(c4, "C4Component") {
		t.Fatalf("expected C4Component, got: %s", c4)
	}
	if !strings.Contains(c4, "Container_Boundary") {
		t.Fatalf("expected Container_Boundary, got: %s", c4)
	}
	if !strings.Contains(c4, "Rel(") {
		t.Fatalf("expected Rel statement, got: %s", c4)
	}
}
