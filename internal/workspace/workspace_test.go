package workspace

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/codedna/codedna/internal/model"
)

func TestExtractEndpointsAndCalls(t *testing.T) {
	tmpDir := t.TempDir()

	// 1. Create mock backend Django urls.py
	backendDir := filepath.Join(tmpDir, "backend")
	os.MkdirAll(backendDir, 0o755)
	urlsPy := `
urlpatterns = [
    path('api/v1/auth/login/', LoginView.as_view()),
    path('api/v1/courses/', CourseListView.as_view()),
    re_path(r'^api/v1/courses/(?P<pk>\d+)/$', CourseDetailView.as_view()),
]
`
	os.WriteFile(filepath.Join(backendDir, "urls.py"), []byte(urlsPy), 0o644)

	backendFiles := []model.FileNode{
		{Path: "urls.py", Language: "python"},
	}

	endpoints := ExtractEndpoints(1, "backend-service", backendDir, backendFiles)
	if len(endpoints) < 2 {
		t.Fatalf("Expected at least 2 endpoints, got %d", len(endpoints))
	}

	// 2. Create mock frontend api.ts
	frontendDir := filepath.Join(tmpDir, "frontend")
	os.MkdirAll(frontendDir, 0o755)
	apiTs := `
export async function loginUser(creds) {
    return axios.post('/api/v1/auth/login', creds);
}
export async function getCourses() {
    return fetch('/api/v1/courses');
}
`
	os.WriteFile(filepath.Join(frontendDir, "api.ts"), []byte(apiTs), 0o644)

	frontendFiles := []model.FileNode{
		{Path: "api.ts", Language: "typescript"},
	}

	calls := ExtractAPICalls(2, "frontend-app", frontendDir, frontendFiles)
	if len(calls) < 2 {
		t.Fatalf("Expected at least 2 calls, got %d", len(calls))
	}

	// 3. Match cross-service links
	links := ResolveCrossServiceLinks(
		model.RepoInfo{ID: 2, Name: "frontend-app"},
		calls,
		model.RepoInfo{ID: 1, Name: "backend-service"},
		endpoints,
	)

	if len(links) == 0 {
		t.Fatalf("Expected cross-service links to be found, got 0")
	}

	t.Logf("Successfully discovered %d cross-service link(s):", len(links))
	for _, l := range links {
		t.Logf("  • %s -> %s (%s, confidence: %.2f)", l.SourceRepoName, l.TargetRepoName, l.RoutePath, l.Confidence)
	}
}

func TestBuildCompositeGraph(t *testing.T) {
	ws := model.Workspace{
		ID:   1,
		Name: "test-workspace",
		Repos: []model.RepoInfo{
			{ID: 1, Name: "backend"},
			{ID: 2, Name: "frontend"},
		},
	}

	repoGraphs := map[int64]model.GraphData{
		1: {
			Nodes: []model.GraphNode{
				{ID: "file:urls.py", Type: "file", Label: "urls.py", File: "urls.py"},
			},
			Edges: []model.GraphEdge{},
		},
		2: {
			Nodes: []model.GraphNode{
				{ID: "file:api.ts", Type: "file", Label: "api.ts", File: "api.ts"},
			},
			Edges: []model.GraphEdge{},
		},
	}

	links := []model.CrossServiceLink{
		{
			ID:             "cross:1",
			SourceRepoID:   2,
			SourceFile:     "api.ts",
			TargetRepoID:   1,
			TargetFile:     "urls.py",
			HTTPMethod:     "GET",
			RoutePath:      "/api/v1/courses",
			Confidence:     0.95,
		},
	}

	comp := BuildCompositeGraph(ws, repoGraphs, links)
	if len(comp.Nodes) != 2 {
		t.Errorf("Expected 2 nodes, got %d", len(comp.Nodes))
	}
	if len(comp.Edges) != 1 {
		t.Errorf("Expected 1 cross-service edge, got %d", len(comp.Edges))
	}
	if comp.Edges[0].Kind != "cross_service_api" {
		t.Errorf("Expected edge kind cross_service_api, got %s", comp.Edges[0].Kind)
	}
}
