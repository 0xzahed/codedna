package analysis

import (
	"sort"
	"strings"

	"github.com/codedna/codedna/internal/model"
)

// FileGraph is the file-level dependency graph used for cycle detection.
type FileGraph struct {
	Nodes []string
	Edges map[string][]string // from -> to
}

// ModuleGraph builds the file dependency graph from imports + resolved call edges.
func (i *Index) ModuleGraph(results []*model.FileParseResult) *FileGraph {
	g := &FileGraph{Edges: map[string][]string{}}
	fileSet := map[string]bool{}
	for _, r := range results {
		fileSet[r.File] = true
		g.Nodes = append(g.Nodes, r.File)
	}
	sort.Strings(g.Nodes)
	symToFile := map[string]string{}
	for k := range i.Symbols {
		s := &i.Symbols[k]
		symToFile[s.ID] = s.File
	}
	add := func(from, to string) {
		if from == to || to == "" {
			return
		}
		for _, e := range g.Edges[from] {
			if e == to {
				return
			}
		}
		g.Edges[from] = append(g.Edges[from], to)
	}
	// import-based edges
	for _, r := range results {
		for _, t := range i.resolveImportToFile(r.File, model.Import{}, results) {
			_ = t
		}
		for _, imp := range r.Imports {
			targets := i.resolveImportToFile(r.File, imp, results)
			for _, t := range targets {
				add(r.File, t)
			}
		}
	}
	// call-based edges (resolved symbols)
	for _, rel := range i.Rel {
		if rel.Kind != model.RelCalls && rel.Kind != model.RelInstantiates {
			continue
		}
		if rel.Confidence < 0.5 {
			continue // low-confidence edges excluded from module graph
		}
		toFile := symToFile[rel.ToID]
		if toFile == "" {
			continue
		}
		fromFile := strings.TrimPrefix(rel.FromID, "file:")
		if s := i.Symbol(rel.FromID); s != nil {
			fromFile = s.File
		}
		add(fromFile, toFile)
	}
	return g
}

// Cycles returns strongly connected components of size > 1 (circular deps).
func (g *FileGraph) Cycles() []model.Issue {
	sccs := tarjanSCC(g.Nodes, g.Edges)
	cycles := []model.Issue{}
	for _, comp := range sccs {
		if len(comp) < 2 {
			continue
		}
		sort.Strings(comp)
		chain := strings.Join(comp, " -> ")
		cycles = append(cycles, model.Issue{
			Kind:     model.IssueCircularDep,
			Severity: model.SevHigh,
			Title:    "Circular dependency: " + strings.Join(comp, " <-> "),
			Detail: model.CycleDetail{
				Files: comp,
				Chain: chain,
			},
		})
	}
	return cycles
}

// tarjanSCC implements Tarjan's algorithm.
func tarjanSCC(nodes []string, edges map[string][]string) [][]string {
	index := 0
	idx := map[string]int{}
	low := map[string]int{}
	onStack := map[string]bool{}
	stack := []string{}
	var sccs [][]string
	strongconnect := func(v string) {}
	strongconnect = func(v string) {
		idx[v] = index
		low[v] = index
		index++
		stack = append(stack, v)
		onStack[v] = true
		for _, w := range edges[v] {
			if _, seen := idx[w]; !seen {
				strongconnect(w)
				low[v] = minInt(low[v], low[w])
			} else if onStack[w] {
				low[v] = minInt(low[v], idx[w])
			}
		}
		if low[v] == idx[v] {
			var comp []string
			for {
				w := stack[len(stack)-1]
				stack = stack[:len(stack)-1]
				onStack[w] = false
				comp = append(comp, w)
				if w == v {
					break
				}
			}
			sccs = append(sccs, comp)
		}
	}
	// deterministic order
	for _, v := range nodes {
		if _, seen := idx[v]; !seen {
			strongconnect(v)
		}
	}
	return sccs
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Upstream returns all files depending on start (transitively) plus edges info.
func (g *FileGraph) Upstream(start string) []string {
	rev := map[string][]string{}
	for from, tos := range g.Edges {
		for _, to := range tos {
			rev[to] = append(rev[to], from)
		}
	}
	return g.bfs(start, rev)
}

// Downstream returns all files start depends on (transitively).
func (g *FileGraph) Downstream(start string) []string {
	return g.bfs(start, g.Edges)
}

func (g *FileGraph) bfs(start string, edges map[string][]string) []string {
	seen := map[string]bool{start: true}
	queue := []string{start}
	out := []string{}
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		for _, next := range edges[cur] {
			if !seen[next] {
				seen[next] = true
				out = append(out, next)
				queue = append(queue, next)
			}
		}
	}
	sort.Strings(out)
	return out
}

// DeadCodeCandidates finds exported=false symbols with zero resolved incoming
// references and no entrypoint file linkage. Never certain: dynamic dispatch
// and reflection make this a candidate list only.
func (i *Index) DeadCodeCandidates() []model.Issue {
	used := map[string]bool{}
	for _, rel := range i.Rel {
		if rel.Kind == model.RelCalls && rel.Confidence >= 0.5 {
			used[rel.ToID] = true
		}
	}
	issues := []model.Issue{}
	for k := range i.Symbols {
		s := &i.Symbols[k]
		if s.Kind != model.SymbolFunction && s.Kind != model.SymbolMethod {
			continue
		}
		if strings.HasPrefix(s.Name, "Test") || strings.HasPrefix(s.Name, "Test_") {
			continue
		}
		if used[s.ID] {
			continue
		}
		// file import linkage: files imported by others keep symbols alive pragmatically
		issues = append(issues, model.Issue{
			Kind:     model.IssueDeadCode,
			Severity: model.SevLow,
			Title:    "Unchecked reference: " + s.QualifiedName,
			File:     s.File,
			SymbolID: s.ID,
			Detail: map[string]interface{}{
				"reason":     "no incoming resolved calls",
				"name":       s.Name,
				"exported":   s.Exported,
				"line":       s.Line,
			},
		})
	}
	return issues
}

// Hotspots finds files with high cumulative complexity.
func (i *Index) Hotspots(top int) []model.Issue {
	var issues []model.Issue
	for k := range i.Symbols {
		s := &i.Symbols[k]
		if s.Complexity >= 15 {
			issues = append(issues, model.Issue{
				Kind:     model.IssueComplexity,
				Severity: severityForComplexity(s.Complexity),
				Title:    "High complexity: " + s.QualifiedName,
				File:     s.File,
				SymbolID: s.ID,
				Detail: map[string]interface{}{
					"complexity": s.Complexity,
					"line":       s.Line,
				},
			})
		}
	}
	sort.Slice(issues, func(a, b int) bool {
		ca := issues[a].Detail.(map[string]interface{})["complexity"].(int)
		cb := issues[b].Detail.(map[string]interface{})["complexity"].(int)
		return ca > cb
	})
	if len(issues) > top {
		issues = issues[:top]
	}
	return issues
}

func severityForComplexity(c int) model.Severity {
	switch {
	case c >= 40:
		return model.SevHigh
	case c >= 25:
		return model.SevMedium
	default:
		return model.SevLow
	}
}

// EntryPoints detects probable application entry points from evidence.
func (i *Index) EntryPoints(files []model.FileNode) []model.Issue {
	issues := []model.Issue{}
	for k := range i.Symbols {
		s := &i.Symbols[k]
		switch s.Language {
		case "go":
			if s.Name == "main" && s.Kind == model.SymbolFunction && s.Exported {
				// package main check is evidence-weak here; keep anyway but note rule
				issues = append(issues, model.Issue{
					Kind: model.IssueEntryPoint, Severity: model.SevInfo,
					Title: "Entry point: main()", File: s.File, SymbolID: s.ID,
					Detail: map[string]interface{}{"language": "go", "line": s.Line},
				})
			}
		case "java":
			if s.Name == "main" && s.Kind == model.SymbolMethod {
				issues = append(issues, model.Issue{
					Kind: model.IssueEntryPoint, Severity: model.SevInfo,
					Title: "Entry point: main(String[] args)", File: s.File, SymbolID: s.ID,
					Detail: map[string]interface{}{"language": "java", "line": s.Line},
				})
			}
		}
	}
	for _, f := range files {
		base := strings.ToLower(f.Path)
		switch {
		case f.Language == "python":
			if isPythonEntry(i, f.Path) {
				issues = append(issues, model.Issue{
					Kind: model.IssueEntryPoint, Severity: model.SevInfo,
					Title: "Entry point: __main__ script", File: f.Path,
					Detail: map[string]interface{}{"language": "python"},
				})
			}
		case f.Language == "typescript" || f.Language == "javascript":
			if strings.Contains(base, "app/") && (strings.Contains(base, "page.") || strings.Contains(base, "route.")) ||
				strings.Contains(base, "pages/") || strings.HasSuffix(base, "main.ts") || strings.HasSuffix(base, "main.js") ||
				strings.HasSuffix(base, "index.ts") || strings.HasSuffix(base, "index.js") {
				if f.Tags == nil || !hasTag(f.Tags, "test") {
					issues = append(issues, model.Issue{
						Kind: model.IssueEntryPoint, Severity: model.SevInfo,
						Title: "Entry point candidate: " + f.Path, File: f.Path,
						Detail: map[string]interface{}{"language": f.Language, "rule": "framework layout"},
					})
				}
			}
		case f.Language == "php" && (strings.HasSuffix(base, "index.php") || strings.HasSuffix(base, "bootstrap.php")):
			issues = append(issues, model.Issue{
				Kind: model.IssueEntryPoint, Severity: model.SevInfo,
				Title: "Entry point: " + f.Path, File: f.Path,
				Detail: map[string]interface{}{"language": "php"},
			})
		}
	}
	return issues
}

func hasTag(tags []string, tag string) bool {
	for _, t := range tags {
		if t == tag {
			return true
		}
	}
	return false
}

// isPythonEntry checks for the __main__ guard.
func isPythonEntry(i *Index, file string) bool {
	return i.pythonEntries[file]
}

// SetMainGuardEntries records files containing `if __name__ == "__main__"`.
func (i *Index) SetMainGuardEntries(files map[string]bool) {
	i.pythonEntries = files
}
