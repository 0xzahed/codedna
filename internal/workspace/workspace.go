// Package workspace provides multi-repository system mapping, cross-service
// API matching, composite dependency graphs, and cross-repo change blast radius.
package workspace

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/codedna/codedna/internal/model"
)

// Regex patterns for backend route producers
var (
	// Django patterns: path('api/courses/', views.CourseListView.as_view())
	djangoPathRe     = regexp.MustCompile(`(?:path|re_path|url)\s*\(\s*["']([^"']+)["']\s*,\s*([a-zA-Z0-9_\.]+)`)
	djangoRouterRe   = regexp.MustCompile(`router\.register\s*\(\s*["']([^"']+)["']\s*,\s*([a-zA-Z0-9_\.]+)`)
	djangoViewSetRe  = regexp.MustCompile(`class\s+([a-zA-Z0-9_]+)\s*\((?:[^)]*ViewSet|[^)]*APIView)`)

	// FastAPI / Flask patterns: @router.get("/users/{id}")
	fastapiRouteRe   = regexp.MustCompile(`@(?:app|router|api)\.(get|post|put|delete|patch)\s*\(\s*["']([^"']+)["']`)
	flaskRouteRe     = regexp.MustCompile(`@app\.route\s*\(\s*["']([^"']+)["']`)

	// Express patterns: router.get('/users/:id', handler)
	expressRouteRe   = regexp.MustCompile(`(?:app|router)\.(get|post|put|delete|patch)\s*\(\s*["']([^"']+)["']`)

	// Go HTTP patterns: r.GET("/api/v1/users", handler)
	goGinRouteRe     = regexp.MustCompile(`(?:r|router|api|v\d|app)\.(GET|POST|PUT|DELETE|PATCH)\s*\(\s*["']([^"']+)["']`)
	goHttpHandleRe   = regexp.MustCompile(`http\.HandleFunc\s*\(\s*["']([^"']+)["']`)
)

// Regex patterns for consumer API client calls
var (
	fetchRe        = regexp.MustCompile(`fetch\s*\(\s*["'\x60]([^"'\x60]+)["'\x60]`)
	axiosRe        = regexp.MustCompile(`axios\.(get|post|put|delete|patch)\s*\(\s*["'\x60]([^"'\x60]+)["'\x60]`)
	apiClientRe    = regexp.MustCompile(`(?:api|client|http|request)\.(get|post|put|delete|patch)\s*\(\s*["'\x60]([^"'\x60]+)["'\x60]`)
	pythonReqRe    = regexp.MustCompile(`requests\.(get|post|put|delete|patch)\s*\(\s*["']([^"']+)["']`)
)

// ExtractEndpoints extracts API routes produced by a backend codebase.
func ExtractEndpoints(repoID int64, repoName, repoRoot string, files []model.FileNode) []model.APIEndpoint {
	var endpoints []model.APIEndpoint
	seen := make(map[string]bool)

	for _, f := range files {
		// Only check backend-relevant files
		ext := filepath.Ext(f.Path)
		if ext != ".py" && ext != ".go" && ext != ".js" && ext != ".ts" && ext != ".java" && ext != ".php" {
			continue
		}

		fullPath := filepath.Join(repoRoot, f.Path)
		file, err := os.Open(fullPath)
		if err != nil {
			continue
		}

		scanner := bufio.NewScanner(file)
		lineNum := 0

		// Check Next.js App Router route handlers e.g. app/api/auth/route.ts
		if strings.Contains(f.Path, "app/api/") && (strings.HasSuffix(f.Path, "route.ts") || strings.HasSuffix(f.Path, "route.js")) {
			routePath := "/" + strings.TrimSuffix(strings.TrimPrefix(f.Path, "src/"), "/route.ts")
			routePath = strings.TrimSuffix(routePath, "/route.js")
			key := "ANY:" + routePath
			if !seen[key] {
				seen[key] = true
				endpoints = append(endpoints, model.APIEndpoint{
					RepoID:   repoID,
					RepoName: repoName,
					Method:   "ANY",
					Path:     cleanRoute(routePath),
					Handler:  "RouteHandler",
					File:     f.Path,
					Line:     1,
				})
			}
		}

		for scanner.Scan() {
			lineNum++
			line := scanner.Text()

			// Django URL patterns
			if m := djangoPathRe.FindStringSubmatch(line); len(m) >= 3 {
				p := cleanRoute(m[1])
				key := "ANY:" + p
				if !seen[key] {
					seen[key] = true
					endpoints = append(endpoints, model.APIEndpoint{
						RepoID:   repoID,
						RepoName: repoName,
						Method:   "ANY",
						Path:     p,
						Handler:  m[2],
						File:     f.Path,
						Line:     lineNum,
					})
				}
			}

			// Django router.register
			if m := djangoRouterRe.FindStringSubmatch(line); len(m) >= 3 {
				p := cleanRoute(m[1])
				key := "ANY:" + p
				if !seen[key] {
					seen[key] = true
					endpoints = append(endpoints, model.APIEndpoint{
						RepoID:   repoID,
						RepoName: repoName,
						Method:   "ANY",
						Path:     p,
						Handler:  m[2],
						File:     f.Path,
						Line:     lineNum,
					})
				}
			}

			// FastAPI / Flask
			if m := fastapiRouteRe.FindStringSubmatch(line); len(m) >= 3 {
				p := cleanRoute(m[2])
				method := strings.ToUpper(m[1])
				key := method + ":" + p
				if !seen[key] {
					seen[key] = true
					endpoints = append(endpoints, model.APIEndpoint{
						RepoID:   repoID,
						RepoName: repoName,
						Method:   method,
						Path:     p,
						Handler:  "FastAPIHandler",
						File:     f.Path,
						Line:     lineNum,
					})
				}
			}

			// Express
			if m := expressRouteRe.FindStringSubmatch(line); len(m) >= 3 {
				p := cleanRoute(m[2])
				method := strings.ToUpper(m[1])
				key := method + ":" + p
				if !seen[key] {
					seen[key] = true
					endpoints = append(endpoints, model.APIEndpoint{
						RepoID:   repoID,
						RepoName: repoName,
						Method:   method,
						Path:     p,
						Handler:  "ExpressHandler",
						File:     f.Path,
						Line:     lineNum,
					})
				}
			}

			// Go Gin / Echo / Fiber
			if m := goGinRouteRe.FindStringSubmatch(line); len(m) >= 3 {
				p := cleanRoute(m[2])
				method := strings.ToUpper(m[1])
				key := method + ":" + p
				if !seen[key] {
					seen[key] = true
					endpoints = append(endpoints, model.APIEndpoint{
						RepoID:   repoID,
						RepoName: repoName,
						Method:   method,
						Path:     p,
						Handler:  "GoHTTPHandler",
						File:     f.Path,
						Line:     lineNum,
					})
				}
			}
		}
		file.Close()
	}

	return endpoints
}

// ExtractAPICalls extracts outbound API calls made by consumer code (e.g. Frontend).
func ExtractAPICalls(repoID int64, repoName, repoRoot string, files []model.FileNode) []model.APICallSite {
	var calls []model.APICallSite
	seen := make(map[string]bool)

	for _, f := range files {
		ext := filepath.Ext(f.Path)
		if ext != ".js" && ext != ".ts" && ext != ".tsx" && ext != ".jsx" && ext != ".py" && ext != ".go" {
			continue
		}

		fullPath := filepath.Join(repoRoot, f.Path)
		file, err := os.Open(fullPath)
		if err != nil {
			continue
		}

		scanner := bufio.NewScanner(file)
		lineNum := 0

		for scanner.Scan() {
			lineNum++
			line := scanner.Text()

			// Check fetch()
			if m := fetchRe.FindStringSubmatch(line); len(m) >= 2 {
				url := extractRelativePath(m[1])
				if url != "" {
					key := f.Path + ":" + url
					if !seen[key] {
						seen[key] = true
						calls = append(calls, model.APICallSite{
							RepoID:   repoID,
							RepoName: repoName,
							Method:   "GET",
							Path:     url,
							Caller:   filepath.Base(f.Path),
							File:     f.Path,
							Line:     lineNum,
						})
					}
				}
			}

			// Check axios
			if m := axiosRe.FindStringSubmatch(line); len(m) >= 3 {
				url := extractRelativePath(m[2])
				if url != "" {
					method := strings.ToUpper(m[1])
					key := f.Path + ":" + method + ":" + url
					if !seen[key] {
						seen[key] = true
						calls = append(calls, model.APICallSite{
							RepoID:   repoID,
							RepoName: repoName,
							Method:   method,
							Path:     url,
							Caller:   filepath.Base(f.Path),
							File:     f.Path,
							Line:     lineNum,
						})
					}
				}
			}

			// Check apiClient / http helper
			if m := apiClientRe.FindStringSubmatch(line); len(m) >= 3 {
				url := extractRelativePath(m[2])
				if url != "" {
					method := strings.ToUpper(m[1])
					key := f.Path + ":" + method + ":" + url
					if !seen[key] {
						seen[key] = true
						calls = append(calls, model.APICallSite{
							RepoID:   repoID,
							RepoName: repoName,
							Method:   method,
							Path:     url,
							Caller:   filepath.Base(f.Path),
							File:     f.Path,
							Line:     lineNum,
						})
					}
				}
			}
		}
		file.Close()
	}

	return calls
}

// ResolveCrossServiceLinks matches outbound API calls against backend endpoints.
func ResolveCrossServiceLinks(
	consumerRepo model.RepoInfo,
	consumerCalls []model.APICallSite,
	providerRepo model.RepoInfo,
	providerEndpoints []model.APIEndpoint,
) []model.CrossServiceLink {
	var links []model.CrossServiceLink
	seen := make(map[string]bool)

	for _, call := range consumerCalls {
		normCall := normalizePath(call.Path)
		if normCall == "" {
			continue
		}

		for _, ep := range providerEndpoints {
			normEp := normalizePath(ep.Path)
			if normEp == "" {
				continue
			}

			// Compare methods if specified
			if call.Method != "ANY" && ep.Method != "ANY" && call.Method != ep.Method {
				continue
			}

			// Route matching score
			isMatch, confidence := matchRoutePaths(normCall, normEp)
			if !isMatch {
				continue
			}

			linkID := fmt.Sprintf("cross:%d:%s->%d:%s", consumerRepo.ID, call.File, providerRepo.ID, ep.File)
			if seen[linkID] {
				continue
			}
			seen[linkID] = true

			links = append(links, model.CrossServiceLink{
				ID:             linkID,
				SourceRepoID:   consumerRepo.ID,
				SourceRepoName: consumerRepo.Name,
				SourceFile:     call.File,
				SourceSymbol:   call.Caller,
				TargetRepoID:   providerRepo.ID,
				TargetRepoName: providerRepo.Name,
				TargetFile:     ep.File,
				TargetSymbol:   ep.Handler,
				HTTPMethod:     call.Method,
				RoutePath:      ep.Path,
				Confidence:     confidence,
				Description:    fmt.Sprintf("%s (%s) calls %s [%s %s]", consumerRepo.Name, call.Caller, providerRepo.Name, call.Method, ep.Path),
			})
		}
	}

	return links
}

// BuildCompositeGraph merges individual repository graphs and adds cross-service links.
func BuildCompositeGraph(
	ws model.Workspace,
	repoGraphs map[int64]model.GraphData,
	links []model.CrossServiceLink,
) model.GraphData {
	var allNodes []model.GraphNode
	var allEdges []model.GraphEdge

	nodeIDMap := make(map[string]bool)

	// Combine repo nodes and edges with namespacing
	for repoID, g := range repoGraphs {
		repoName := fmt.Sprintf("Repo-%d", repoID)
		for _, r := range ws.Repos {
			if r.ID == repoID {
				repoName = r.Name
				break
			}
		}

		for _, n := range g.Nodes {
			namespacedID := fmt.Sprintf("r%d:%s", repoID, n.ID)
			if n.Data == nil {
				n.Data = make(map[string]interface{})
			}
			n.Data["repoId"] = repoID
			n.Data["repoName"] = repoName
			n.Data["service"] = repoName

			n.ID = namespacedID
			allNodes = append(allNodes, n)
			nodeIDMap[namespacedID] = true
		}

		for _, e := range g.Edges {
			src := fmt.Sprintf("r%d:%s", repoID, e.Source)
			tgt := fmt.Sprintf("r%d:%s", repoID, e.Target)
			if nodeIDMap[src] && nodeIDMap[tgt] {
				e.ID = fmt.Sprintf("r%d:%s", repoID, e.ID)
				e.Source = src
				e.Target = tgt
				allEdges = append(allEdges, e)
			}
		}
	}

	// Add Cross-Service API edges
	for i, link := range links {
		srcID := fmt.Sprintf("r%d:file:%s", link.SourceRepoID, link.SourceFile)
		tgtID := fmt.Sprintf("r%d:file:%s", link.TargetRepoID, link.TargetFile)

		// If exact file nodes not in graph, look for symbols or fallback to first available node in that repo
		if !nodeIDMap[srcID] {
			for _, n := range allNodes {
				if n.Data["repoId"] == link.SourceRepoID && strings.Contains(n.File, link.SourceFile) {
					srcID = n.ID
					break
				}
			}
		}

		if !nodeIDMap[tgtID] {
			for _, n := range allNodes {
				if n.Data["repoId"] == link.TargetRepoID && strings.Contains(n.File, link.TargetFile) {
					tgtID = n.ID
					break
				}
			}
		}

		if nodeIDMap[srcID] && nodeIDMap[tgtID] {
			allEdges = append(allEdges, model.GraphEdge{
				ID:                fmt.Sprintf("cross-edge-%d", i),
				Source:            srcID,
				Target:            tgtID,
				Kind:              "cross_service_api",
				Confidence:        link.Confidence,
				Method:            fmt.Sprintf("%s %s", link.HTTPMethod, link.RoutePath),
				IsRuntimeObserved: false,
			})
		}
	}

	return model.GraphData{
		Nodes: allNodes,
		Edges: allEdges,
	}
}

// Helpers
func cleanRoute(r string) string {
	r = strings.TrimSpace(r)
	r = strings.TrimPrefix(r, "^")
	r = strings.TrimSuffix(r, "$")
	if !strings.HasPrefix(r, "/") && !strings.HasPrefix(r, "api/") {
		r = "/" + r
	}
	return r
}

func extractRelativePath(raw string) string {
	raw = strings.TrimSpace(raw)
	// Remove variable prefixes e.g. `${API_URL}/api/courses` -> `/api/courses`
	if idx := strings.Index(raw, "/api/"); idx != -1 {
		return raw[idx:]
	}
	if strings.HasPrefix(raw, "http://") || strings.HasPrefix(raw, "https://") {
		parts := strings.SplitN(raw, "/", 4)
		if len(parts) >= 4 {
			return "/" + parts[3]
		}
	}
	if strings.HasPrefix(raw, "/") {
		return raw
	}
	return ""
}

func normalizePath(p string) string {
	p = strings.TrimSpace(p)
	p = strings.TrimSuffix(p, "/")
	// Convert parameters e.g. <int:id>, {id}, :id, <slug:slug> to :param
	paramRe := regexp.MustCompile(`(?:<[^>]+>|\{[^}]+\}|:[a-zA-Z0-9_]+)`)
	p = paramRe.ReplaceAllString(p, ":param")
	return p
}

func matchRoutePaths(callPath, epPath string) (bool, float64) {
	if callPath == epPath {
		return true, 0.95
	}
	// Prefix match: /api/v1/courses vs courses
	if strings.HasSuffix(callPath, epPath) || strings.HasSuffix(epPath, callPath) {
		return true, 0.85
	}
	// Segments match
	callParts := strings.Split(strings.Trim(callPath, "/"), "/")
	epParts := strings.Split(strings.Trim(epPath, "/"), "/")

	if len(callParts) > 0 && len(epParts) > 0 {
		// If last 2 segments match (e.g. auth/login or courses/:param)
		if len(callParts) >= 2 && len(epParts) >= 2 {
			if callParts[len(callParts)-1] == epParts[len(epParts)-1] &&
				callParts[len(callParts)-2] == epParts[len(epParts)-2] {
				return true, 0.88
			}
		}
		if callParts[len(callParts)-1] == epParts[len(epParts)-1] &&
			(callParts[len(callParts)-1] != "" && callParts[len(callParts)-1] != ":param") {
			return true, 0.75
		}
	}

	return false, 0
}
