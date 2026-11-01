package parser

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"

	sitter "github.com/tree-sitter/go-tree-sitter"
	"github.com/codedna/codedna/internal/model"
)

// Parse parses a file into symbols, imports and call sites.
func Parse(root, relPath string) (*model.FileParseResult, error) {
	ext := strings.ToLower(filepath.Ext(relPath))
	lang := LanguageForExt(ext)
	if lang == "" {
		return nil, fmt.Errorf("unsupported language for %s", relPath)
	}
	sp, err := getSpecs()
	if err != nil {
		return nil, err
	}
	spec, ok := sp[grammarKey(lang, ext)]
	if !ok {
		return nil, fmt.Errorf("no grammar for %s", relPath)
	}
	src, err := os.ReadFile(filepath.Join(root, relPath))
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", relPath, err)
	}
	return ParseSource(relPath, spec, src)
}

func stripQuotes(s string) string {
	for len(s) > 1 && (s[0] == '"' || s[0] == '\'' || s[0] == '`') {
		open := s[0]
		closing := s[len(s)-1]
		if closing == open {
			s = s[1 : len(s)-1]
			continue
		}
		break
	}
	return s
}

// ParseSource runs all queries over one file and assembles the result.
func ParseSource(relPath string, spec *LanguageSpec, src []byte) (*model.FileParseResult, error) {
	qs, err := getCompiled()
	if err != nil {
		return nil, err
	}
	comp, ok := qs[spec.Name]
	if !ok {
		return nil, fmt.Errorf("queries not compiled for %s", spec.Name)
	}

	res := &model.FileParseResult{
		File:     filepath.ToSlash(relPath),
		Language: spec.BaseLang,
	}

	parser := sitter.NewParser()
	defer parser.Close()
	if err := parser.SetLanguage(spec.Lang); err != nil {
		return nil, fmt.Errorf("set language %s: %v", spec.Name, err)
	}
	tree := parser.Parse(src, nil)
	if tree == nil {
		return nil, fmt.Errorf("no parse tree for %s", relPath)
	}
	defer tree.Close()
	root := tree.RootNode()
	if root.HasError() {
		res.ParseErrors = append(res.ParseErrors, model.ParseError{
			File: res.File, Language: spec.BaseLang,
			Message: "syntax errors present; analysis may be partial",
			Recovered: true,
		})
	}

	text := func(n sitter.Node) string { return strings.TrimSpace(n.Utf8Text(src)) }

	capture := func(m *sitter.QueryMatch, q *sitter.Query, name string) *sitter.Node {
		idx, ok := q.CaptureIndexForName(name)
		if !ok {
			return nil
		}
		for i := range m.Captures {
			if m.Captures[i].Index == uint32(idx) {
				n := m.Captures[i].Node
				return &n
			}
		}
		return nil
	}

	run := func(q *sitter.Query, cb func(*sitter.QueryMatch)) error {
		cursor := sitter.NewQueryCursor()
		matches := cursor.Matches(q, root, src)
		for {
			m := matches.Next()
			if m == nil {
				break
			}
			cb(m)
		}
		return nil
	}

	// ---------- symbols ----------
type raw struct {
	decl sitter.Node
	name string
}
	var raws []raw
	if err := run(comp.symQuery, func(m *sitter.QueryMatch) {
		nameNode := capture(m, comp.symQuery, "name")
		declNode := capture(m, comp.symQuery, "decl")
		if nameNode == nil || declNode == nil {
			return
		}
		raws = append(raws, raw{decl: *declNode, name: text(*nameNode)})
	}); err != nil {
		return nil, fmt.Errorf("symbol query: %v", err)
	}
	sort.SliceStable(raws, func(i, j int) bool { return raws[i].decl.StartByte() < raws[j].decl.StartByte() })

	kindFor := func(decl *sitter.Node) model.SymbolKind {
		switch spec.BaseLang {
		case "go":
			if decl.Kind() == "function_declaration" {
				return model.SymbolFunction
			}
			if decl.Kind() == "method_declaration" {
				return model.SymbolMethod
			}
			for i := 0; i < int(decl.ChildCount()); i++ {
				n := decl.Child(uint(i))
				if n == nil {
					continue
				}
				switch n.Kind() {
				case "struct_type":
					return model.SymbolStruct
				case "interface_type":
					return model.SymbolInterface
				}
			}
			return model.SymbolType
		case "python":
			if decl.Kind() == "class_definition" {
				return model.SymbolClass
			}
			return model.SymbolFunction
		case "javascript", "typescript":
			switch decl.Kind() {
			case "class_declaration":
				return model.SymbolClass
			case "interface_declaration":
				return model.SymbolInterface
			case "type_alias_declaration", "enum_declaration":
				return model.SymbolType
			case "method_definition":
				return model.SymbolMethod
			default:
				return model.SymbolFunction
			}
		case "java":
			switch decl.Kind() {
			case "class_declaration", "record_declaration":
				return model.SymbolClass
			case "interface_declaration":
				return model.SymbolInterface
			default:
				return model.SymbolMethod
			}
		case "php":
			switch decl.Kind() {
			case "class_declaration", "trait_declaration":
				return model.SymbolClass
			case "interface_declaration":
				return model.SymbolInterface
			case "method_declaration":
				return model.SymbolMethod
			default:
				return model.SymbolFunction
			}
		}
		return model.SymbolFunction
	}

	goReceiverBase := func(decl *sitter.Node) string {
		if recv := decl.ChildByFieldName("receiver"); recv != nil {
			recvText := strings.Trim(recv.Utf8Text(src), "() \t\n*")
			fields := strings.Fields(recvText)
			if len(fields) == 0 {
				return ""
			}
			return fields[len(fields)-1]
		}
		return ""
	}

	// Build symbols from raw declarations; compute parents via containment.
	for _, r := range raws {
		sp0, ep0 := r.decl.StartPosition(), r.decl.EndPosition()
		kind := kindFor(&r.decl)
		parent := ""
		if spec.BaseLang == "go" {
			if kind == model.SymbolMethod {
				parent = goReceiverBase(&r.decl)
			}
		} else {
			// tightest container class/interface/struct by byte range
			bestEnd := 1 << 62
			for _, cRaw := range raws {
				cKind := kindFor(&cRaw.decl)
				if !(cKind == model.SymbolClass || cKind == model.SymbolStruct || cKind == model.SymbolInterface) {
					continue
				}
				if cRaw.decl.StartByte() <= r.decl.StartByte() && cRaw.decl.EndByte() >= r.decl.EndByte() &&
					int(cRaw.decl.EndByte()) < bestEnd {
					parent = cRaw.name
					bestEnd = int(cRaw.decl.EndByte())
				}
			}
		}
		sym := model.Symbol{
			Name:     r.name,
			Kind:     kind,
			File:     res.File,
			Language: spec.BaseLang,
			Line:     int(sp0.Row) + 1,
			Column:   int(sp0.Column),
			EndLine:  int(ep0.Row) + 1,
			EndColumn: int(ep0.Column),
			ParentClassName: parent,
		}
		if sym.ParentClassName != "" {
			sym.QualifiedName = sym.ParentClassName + "." + sym.Name
		} else {
			sym.QualifiedName = sym.Name
		}
		sym.ID = SymbolID(sym)
		sym.ParentStartByte = int(r.decl.StartByte())
		sym.ParentEndByte = int(r.decl.EndByte())
		res.Symbols = append(res.Symbols, sym)
	}

	symByTextRange := func(n sitter.Node) model.Symbol {
		best := 1 << 62
		var found model.Symbol
		for _, s := range res.Symbols {
			if uint(s.ParentStartByte) <= n.StartByte() && uint(s.ParentEndByte) >= n.EndByte() && s.ParentEndByte < best {
				best = s.ParentEndByte
				found = s
			}
		}
		return found
	}

	// exported detection
	for i := range res.Symbols {
		res.Symbols[i].Exported = exportedFlag(&raws[i].decl, res.Symbols[i].Name, src, spec)
	}

	// ---------- imports ----------
	impSeen := map[string]bool{}
	if err := run(comp.importQuery, func(m *sitter.QueryMatch) {
		imp := model.Import{}
		if p := capture(m, comp.importQuery, "import_path"); p != nil {
			imp.Path = stripQuotes(text(*p))
			imp.Line = int(p.StartPosition().Row) + 1
			imp.Col = int(p.StartPosition().Column)
		}
		if stmt := capture(m, comp.importQuery, "import_stmt"); stmt != nil {
			t := stmt.Utf8Text(src)
			if spec.BaseLang == "java" {
				t = strings.TrimSpace(t)
				t = strings.TrimPrefix(t, "import ")
				t = strings.TrimSpace(strings.TrimSuffix(t, ";"))
				t = strings.TrimPrefix(t, "static ")
				imp.Path = t
			} else if spec.BaseLang == "php" {
				t = strings.TrimSpace(t)
				t = strings.TrimPrefix(t, "use ")
				t = strings.TrimSpace(strings.TrimSuffix(t, ";"))
				// "use Foo\Bar as B;" -> path Foo\Bar, alias B
				parts := strings.Fields(t)
				if len(parts) >= 3 && parts[len(parts)-2] == "as" {
					imp.Alias = parts[len(parts)-1]
					imp.Path = parts[0]
					t = parts[0]
				} else {
					imp.Path = t
				}
			} else {
				imp.Path = t
			}
			imp.Line = int(stmt.StartPosition().Row) + 1
		}
		if a := capture(m, comp.importQuery, "import_alias"); a != nil {
			imp.Alias = text(*a)
		}
		if idx, ok := comp.importQuery.CaptureIndexForName("import_name"); ok {
			for i := range m.Captures {
				if m.Captures[i].Index == uint32(idx) {
					imp.Names = append(imp.Names, text(m.Captures[i].Node))
				}
			}
		}
		if imp.Path == "" {
			return
		}
		key := fmt.Sprintf("%s|%d|%v", imp.Path, imp.Line, imp.Names)
		if impSeen[key] {
			return
		}
		impSeen[key] = true
		res.Imports = append(res.Imports, imp)
	}); err != nil {
		return nil, fmt.Errorf("import query: %v", err)
	}

	// ---------- calls ----------
	if err := run(comp.callQuery, func(m *sitter.QueryMatch) {
		callee := capture(m, comp.callQuery, "call_callee")
		if callee == nil {
			return
		}
		cal := model.Call{
			Callee: text(*callee),
			Line:   int(callee.StartPosition().Row) + 1,
			Col:    int(callee.StartPosition().Column),
		}
		if spec.NewQuery {
			// new_expression patterns follow the two call patterns in every spec
			declParent := callee.Parent()
			for d := 0; d < 4 && declParent != nil; d++ {
				if declParent.Kind() == "new_expression" || declParent.Kind() == "object_creation_expression" {
					cal.IsNew = true
					break
				}
				declParent = declParent.Parent()
			}
		}
		if r := capture(m, comp.callQuery, "call_recv"); r != nil {
			cal.Receiver = text(*r)
		}
		if s := symByTextRange(*callee); s.ID != "" {
			cal.InSymbolID = s.ID
		}
		if cal.Callee == "" || cal.Callee == "func" && false {
			return
		}
		res.Calls = append(res.Calls, cal)
	}); err != nil {
		return nil, fmt.Errorf("call query: %v", err)
	}

	// ---------- branches / complexity ----------
	if err := run(comp.branchQuery, func(m *sitter.QueryMatch) {
		b := capture(m, comp.branchQuery, "branch")
		if b == nil {
			return
		}
		res.BranchCount++
		if s := symByTextRange(*b); s.ID != "" {
			for i := range res.Symbols {
				if res.Symbols[i].ID == s.ID {
					res.Symbols[i].Complexity++
				}
			}
		}
	}); err != nil {
		return nil, fmt.Errorf("branch query: %v", err)
	}
	for i := range res.Symbols {
		res.Symbols[i].Complexity++
		res.Symbols[i].LineCount = res.Symbols[i].EndLine - res.Symbols[i].Line + 1
		res.Symbols[i].ParentStartByte = -1
		res.Symbols[i].ParentEndByte = -1
	}

	return res, nil
}

// exportedFlag decides whether a declared symbol is visible/exported.
func exportedFlag(decl *sitter.Node, name string, src []byte, spec *LanguageSpec) bool {
	switch spec.BaseLang {
	case "go":
		r, _ := utf8.DecodeRuneInString(name)
		return unicode.IsUpper(r)
	case "python":
		return !strings.HasPrefix(name, "_")
	case "javascript", "typescript":
		for p, d := 0, decl.Parent(); p < 3 && d != nil; p, d = p+1, d.Parent() {
			if d.Kind() == "export_statement" {
				return true
			}
		}
		return false
	case "java":
		t := decl.Utf8Text(src)
		return strings.HasPrefix(t, "public")
	case "php":
		if decl.Kind() == "method_declaration" {
			return strings.Contains(decl.Utf8Text(src), "public")
		}
		return true
	}
	return true
}

// SymbolID builds a stable, human-inspectable identifier.
func SymbolID(s model.Symbol) string {
	return fmt.Sprintf("sym:%s:%s:%s:%s:L%d", s.Language, s.Kind, s.QualifiedName, s.File, s.Line)
}
