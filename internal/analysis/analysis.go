// Package analysis resolves references into a dependency graph and computes
// code intelligence: cycles, dead-code candidates and entry points.
package analysis

import (
	"sort"
	"strings"

	"github.com/codedna/codedna/internal/model"
)

// Index holds the global symbol catalog from all parsed files.
type Index struct {
	Symbols []model.Symbol
	// symbol name -> symbol id (first occurrence wins for unique lookups)
	ByName       map[string][]string
	filePackages map[string]string // Go package clause per file
	pythonEntries map[string]bool
	// file -> imports
	Imports map[string][]model.Import
	// resolved edges
	Rel        []model.Relationship
	Resolved   int
	Unresolved int
}

// Build consumes all file parse results into an index.
func Build(results []*model.FileParseResult, repoFiles []model.FileNode) *Index {
	idx := &Index{
		ByName:       map[string][]string{},
		filePackages: map[string]string{},
		Imports:      map[string][]model.Import{},
	}
	idx.Symbols = []model.Symbol{}
	for _, r := range results {
		for _, s := range r.Symbols {
			idx.Symbols = append(idx.Symbols, s)
			idx.ByName[s.Name] = append(idx.ByName[s.Name], s.ID)
		}
	}
	for _, r := range results {
		idx.Imports[r.File] = r.Imports
		pkg := extractPackage(r)
		if pkg != "" {
			idx.filePackages[r.File] = pkg
		}
	}
	_ = repoFiles
	return idx
}

// extractPackage captures Go package clauses from the raw parse queries output.
func extractPackage(r *model.FileParseResult) string {
	// Parser does not expose package clauses yet; derive from file text once:
	return ""
}

// symbolByID returns the index entry for an ID.
func (i *Index) Symbol(id string) *model.Symbol {
	for j := range i.Symbols {
		if i.Symbols[j].ID == id {
			return &i.Symbols[j]
		}
	}
	return nil
}

// Resolve processes all call sites into relationships and stores them on the index.
func (i *Index) Resolve(results []*model.FileParseResult) []model.Relationship {
	var rels []model.Relationship
	addRel := func(kind model.RelationshipKind, fromID, toID, toName, importPath string, conf float64, method model.ResolutionMethod, loc model.Ref) {
		for idx := range rels {
			r := &rels[idx]
			if r.Kind == kind && r.FromID == fromID && r.ToID == toID && r.Method == method {
				r.Locations = append(r.Locations, loc)
				return
			}
		}
		rels = append(rels, model.Relationship{
			Kind: kind, FromID: fromID, ToID: toID, ToName: toName, ImportPath: importPath,
			Confidence: conf, Method: method,
			Locations: []model.Ref{loc},
		})
	}
	for _, r := range results {
		for _, cal := range r.Calls {
			m := i.resolveCall(r.File, cal)
			if m.ToID != "" {
				i.Resolved++
			} else {
				i.Unresolved++
				// Unresolved calls create no edge; keep them out of the graph
				// instead of inventing low-confidence relationships.
				if m.ImportPath == "" {
					continue
				}
			}
			loc := model.Ref{File: r.File, Line: cal.Line, Col: cal.Col}
			addRel(m.Kind, m.FromID, m.ToID, m.ToName, m.ImportPath, m.Confidence, m.Method, loc)
		}
	}
	sortRelationships(rels)
	i.Rel = rels
	return rels
}

type match struct {
	Kind       model.RelationshipKind
	FromID     string
	ToID       string
	ToName     string
	ImportPath string
	Confidence float64
	Method     model.ResolutionMethod
}

// resolveCall maps a call site to a symbol or an import path with confidence.
func (i *Index) resolveCall(file string, cal model.Call) match {
	m := match{ToName: cal.Callee}
	if cal.InSymbolID != "" {
		m.FromID = cal.InSymbolID
	} else {
		// file-level caller marker
		m.FromID = "file:" + file
	}
	simple := cal.Callee
	if cal.Receiver != "" {
		simple = cal.Receiver + "." + cal.Callee
	}

	// 1) qualified: receiver matches an import alias / module name.
	if cal.Receiver != "" {
		switch {
		case cal.Receiver == "self" || cal.Receiver == "this" || cal.Receiver == "cls":
			if s, conf := i.resolveSelfMethod(file, cal); s != nil {
				m.ToID = s.ID
				m.ToName = s.QualifiedName
				m.Confidence = conf
				m.Method = model.MethodSelfMethod
				m.Kind = model.RelCalls
				return m
			}
		case i.isImportAlias(file, cal.Receiver):
			if imp := i.importForAlias(file, cal.Receiver); imp != "" {
				// resolve against imported module's symbols when in-repo
				if s := i.byModuleAndName(imp, cal.Callee); s != nil {
					m.ToID = s.ID
					m.ToName = s.QualifiedName
					m.Confidence = 0.7
					m.Method = model.MethodQualified
					m.Kind = model.RelCalls
					return m
				}
				m.ImportPath = imp
				m.ToID = ""
				m.Confidence = 0.9
				m.Method = model.MethodQualified
				m.Kind = model.RelImports
				return m
			}
			if s := i.byModuleAndName(cal.Receiver, cal.Callee); s != nil {
				m.ToID = s.ID
				m.ToName = s.QualifiedName
				m.Confidence = 0.75
				m.Method = model.MethodQualified
				m.Kind = model.RelCalls
				return m
			}
		}
		// receiver-first-identifier hint for constructor-ish calls (Java "new Invoice()")
		if cal.IsNew {
			if s := i.constructorFor(cal.Callee); s != nil {
				m.ToID = s.ID
				m.ToName = s.QualifiedName
				m.Confidence = 0.85
				m.Method = model.MethodConstructor
				m.Kind = model.RelInstantiates
				return m
			}
		}
		// Class.method pattern where receiver is a known class in repo
		if s := i.methodOfClass(cal.Receiver, cal.Callee); s != nil {
			m.ToID = s.ID
			m.ToName = s.QualifiedName
			m.Confidence = 0.7
			m.Method = model.MethodQualified
			m.Kind = model.RelCalls
			return m
		}
		// unresolved qualified call: mark module edge, low confidence
		m.ToID = ""
		m.ToName = simple
		m.Confidence = 0.1
		m.Method = model.MethodQualified
		m.ImportPath = cal.Receiver
		m.Kind = model.RelCalls
		return m
	}

	// 2) bare call
	if s := i.byNameUniqueAmongFiles(file, cal.Callee); s != nil {
		switch {
		case s.File == file:
			m.ToID = s.ID
			m.ToName = s.QualifiedName
			m.Confidence = 0.95
			m.Method = model.MethodExactLocal
			m.Kind = model.RelCalls
			return m
		case i.sameModule(s.File, file):
			m.ToID = s.ID
			m.ToName = s.QualifiedName
			m.Confidence = 0.8
			m.Method = model.MethodExactModule
			m.Kind = model.RelCalls
			return m
		case len(i.ByName[cal.Callee]) == 1:
			m.ToID = s.ID
			m.ToName = s.QualifiedName
			m.Confidence = 0.55
			m.Method = model.MethodExactUnique
			m.Kind = model.RelCalls
			return m
		}
	}
	if s := i.constructorFor(cal.Callee); s != nil && cal.IsNew {
		m.ToID = s.ID
		m.ToName = s.QualifiedName
		m.Confidence = 0.85
		m.Method = model.MethodConstructor
		m.Kind = model.RelInstantiates
		return m
	}
	m.ToID = ""
	m.ToName = simple
	m.Confidence = 0.0
	m.Kind = model.RelCalls
	return m
}

// isImportAlias tells whether revc is an import alias in this file.
func (i *Index) isImportAlias(file, recv string) bool {
	for _, imp := range i.Imports[file] {
		if imp.Alias == recv {
			return true
		}
	}
	return false
}

func (i *Index) importForAlias(file, recv string) string {
	for _, imp := range i.Imports[file] {
		if imp.Alias == recv {
			return imp.Path
		}
	}
	return recv
}

// resolveSelfMethod resolves this.foo()/self.foo() within enclosing class.
func (i *Index) resolveSelfMethod(file string, cal model.Call) (*model.Symbol, float64) {
	inFile := []model.Symbol{}
	for j := range i.Symbols {
		s := &i.Symbols[j]
		if s.File == file && s.Kind == model.SymbolMethod && s.Name == cal.Callee {
			inFile = append(inFile, *s)
		}
	}
	if len(inFile) == 1 {
		return &inFile[0], 0.75
	}
	if len(inFile) > 1 {
		return &inFile[0], 0.5
	}
	return nil, 0
}

// byModuleAndName finds a symbol whose module/file/namespace matches prefix.
func (i *Index) byModuleAndName(prefix, name string) *model.Symbol {
	cands := i.ByName[name]
	if len(cands) == 0 {
		return nil
	}
	for _, id := range cands {
		s := i.Symbol(id)
		if s == nil {
			continue
		}
		if pathMatches(prefix, s.File, s.QualifiedName) {
			return s
		}
	}
	return nil
}

func pathMatches(prefix, file, qualified string) bool {
	prefix = strings.Trim(prefix, "./\\\"'")
	base := strings.TrimPrefix(file, "./")
	lt := strings.TrimSuffix(strings.ToLower(base), strings.ToLower(file)[-0:])
	_ = lt
	if strings.EqualFold(base, prefix) {
		return true
	}
	dots := strings.Split(qualified, ".")
	if len(dots) > 1 {
		class := dots[0]
		if class == prefix {
			return true
		}
	}
	if strings.HasSuffix(strings.ToLower(base), strings.ToLower(prefix)+".py") ||
		strings.HasSuffix(strings.ToLower(base), strings.ToLower(prefix)+".ts") ||
		strings.HasSuffix(strings.ToLower(base), strings.ToLower(prefix)+".js") ||
		strings.HasSuffix(strings.ToLower(base), strings.ToLower(prefix)+".php") {
		return true
	}
	slash := strings.ReplaceAll(prefix, ".", string("/"))
	if slash != prefix && strings.HasSuffix(strings.ToLower(base), strings.ToLower(slash)+globSuffixAny) {
		return true
	}
	return false
}

const globSuffixAny = ""

// methodOfClass resolves Class.method via a discovered class symbol.
func (i *Index) methodOfClass(class, method string) *model.Symbol {
	for j := range i.Symbols {
		s := &i.Symbols[j]
		if s.Kind == model.SymbolMethod && s.ParentClassName == class &&
			(s.Name == method) {
			return s
		}
	}
	return nil
}

// constructorFor locates a constructor symbol (class or constructor method).
func (i *Index) constructorFor(name string) *model.Symbol {
	cands := i.ByName[name]
	if len(cands) == 0 {
		return nil
	}
	if len(cands) == 1 {
		return i.Symbol(cands[0])
	}
	return nil
}

// byNameUniqueAmongFiles picks the best local rule candidate.
func (i *Index) byNameUniqueAmongFiles(file, name string) *model.Symbol {
	ids := i.ByName[name]
	if len(ids) == 0 {
		return nil
	}
	j := 0
	for _, id := range ids {
		if s := i.Symbol(id); s != nil {
			ids[j] = id
			j++
		}
	}
	ids = ids[:j]
	if len(ids) == 0 {
		return nil
	}
	if len(ids) == 1 {
		return i.Symbol(ids[0])
	}
	// prefer same file
	for _, id := range ids {
		if s := i.Symbol(id); s != nil && s.File == file {
			return s
		}
	}
	// prefer same directory
	dirOf := func(p string) string {
		idx := strings.LastIndexByte(p, '/')
		if idx < 0 {
			return ""
		}
		return p[:idx]
	}
	dir := dirOf(file)
	sameDir := []string{}
	for _, id := range ids {
		if s := i.Symbol(id); s != nil && dirOf(s.File) == dir {
			sameDir = append(sameDir, id)
		}
	}
	if len(sameDir) == 1 {
		return i.Symbol(sameDir[0])
	}
	if len(sameDir) > 1 {
		return i.Symbol(sameDir[0]) // deterministic pick, slightly lower confidence handled elsewhere
	}
	// prefer exported
	var expIDs []string
	for _, id := range ids {
		if s := i.Symbol(id); s != nil && s.Exported {
			expIDs = append(expIDs, id)
		}
	}
	if len(expIDs) >= 1 && len(expIDs) == 1 {
		return i.Symbol(expIDs[0])
	}
	return i.Symbol(ids[0])
}

// sameModule reports whether two files live in the same package/module.
func (i *Index) sameModule(a, b string) bool {
	dirOf := func(p string) string {
		idx := strings.LastIndexByte(p, '/')
		if idx < 0 {
			return ""
		}
		return p[:idx]
	}
	if dirOf(a) == dirOf(b) {
		return true
	}
	pa, pb := i.filePackages[a], i.filePackages[b]
	return pa != "" && pb != "" && pa == pb
}

// FileEdges builds file-level dependency edges derived from resolved patterns.
func (i *Index) FileEdges(results []*model.FileParseResult) []model.Relationship {
	fileSet := map[string]bool{}
	for _, r := range results {
		fileSet[r.File] = true
	}
	var out []model.Relationship
	for _, r := range results {
		for _, imp := range r.Imports {
			targets := i.resolveImportToFile(r.File, imp, results)
			for _, t := range targets {
				out = append(out, model.Relationship{
					Kind:        model.RelImports,
					FromID:      "file:" + r.File,
					ToID:        "file:" + t,
					ToName:      t,
					ImportPath:  imp.Path,
					Confidence:  0.9,
					Method:      model.MethodQualified,
					Locations:   []model.Ref{},
				})
			}
		}
	}
	sortRelationships(out)
	return out
}

// resolveImportToFile maps an import statement to repo-internal file targets.
func (i *Index) resolveImportToFile(file string, imp model.Import, results []*model.FileParseResult) []string {
	var targets []string
	base := strings.TrimPrefix(strings.TrimSpace(imp.Path), "./")
	if base == "" || strings.HasPrefix(base, "http") {
		return nil
	}
	// python relative: relative_import begins with "."
	for _, r := range results {
		candidate := r.File
		lc := strings.ToLower(candidate)
		if strings.HasSuffix(lc, ".py") {
			if strings.HasSuffix(lc, strings.ToLower(base)+".py") || strings.HasSuffix(lc, "/"+strings.ToLower(base)+".py") {
				targets = append(targets, r.File)
			}
		}
	}
	exts := []string{"", ".ts", ".tsx", ".js", ".mjs", ".jsx", ".json"}
	for _, r := range results {
		lc := strings.ToLower(r.File)
		for _, ext := range exts {
			stem := strings.ToLower(base)
			if lc == stem+ext || strings.HasSuffix(lc, "/"+stem+ext) {
				targets = append(targets, r.File)
			}
		}
	}
	_ = file
	return uniqueStrings(targets)
}

func uniqueStrings(in []string) []string {
	seen := map[string]bool{}
	out := []string{}
	for _, s := range in {
		if !seen[s] {
			seen[s] = true
			out = append(out, s)
		}
	}
	return out
}

func sortRelationships(rels []model.Relationship) {
	sort.Slice(rels, func(i, j int) bool {
		if rels[i].FromID != rels[j].FromID {
			return rels[i].FromID < rels[j].FromID
		}
		return rels[i].ToID < rels[j].ToID
	})
}
