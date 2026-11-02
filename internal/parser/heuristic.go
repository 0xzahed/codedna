package parser

import (
	"bufio"
	"bytes"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/codedna/codedna/internal/model"
)

var (
	// Functions & Methods
	reRustFn    = regexp.MustCompile(`^\s*(?:pub(?:\s*\([^)]*\))?\s+)?(?:async\s+)?fn\s+([a-zA-Z0-9_]+)`)
	reRubyDef   = regexp.MustCompile(`^\s*def\s+([a-zA-Z0-9_!?]+)`)
	reKotlinFun = regexp.MustCompile(`^\s*(?:(?:public|private|protected|internal|override|inline|suspend)\s+)*fun\s+([a-zA-Z0-9_]+)`)
	reSwiftFunc = regexp.MustCompile(`^\s*(?:(?:public|private|internal|fileprivate|open|override|static)\s+)*func\s+([a-zA-Z0-9_]+)`)
	reCMethod   = regexp.MustCompile(`^\s*(?:(?:public|private|protected|static|virtual|override|async)\s+)*(?:[a-zA-Z0-9_<>[\]?]+\s+)+([a-zA-Z0-9_]+)\s*\([^)]*\)\s*\{?`)
	reSQLProc   = regexp.MustCompile(`(?i)^\s*CREATE\s+(?:OR\s+REPLACE\s+)?(?:PROCEDURE|FUNCTION)\s+([a-zA-Z0-9_.]+)`)

	// Classes, Structs, Interfaces
	reClass     = regexp.MustCompile(`^\s*(?:(?:public|private|protected|internal|abstract|sealed|final|pub(?:\s*\([^)]*\))?)\s+)*(?:class|struct|interface|trait|enum|record)\s+([a-zA-Z0-9_]+)`)
	reRubyClass = regexp.MustCompile(`^\s*(?:class|module)\s+([a-zA-Z0-9_:]+)`)
	reSQLTable  = regexp.MustCompile(`(?i)^\s*CREATE\s+TABLE\s+(?:IF\s+NOT\s+EXISTS\s+)?([a-zA-Z0-9_.]+)`)
	reGQLType   = regexp.MustCompile(`^\s*(?:type|interface|input|enum)\s+([a-zA-Z0-9_]+)`)

	// Imports & Includes
	reCInclude  = regexp.MustCompile(`^\s*#include\s+["<]([^">]+)[">]`)
	reRustUse   = regexp.MustCompile(`^\s*use\s+([^;]+);`)
	reRubyReq   = regexp.MustCompile(`^\s*(?:require|require_relative)\s+['"]([^'"]+)['"]`)
	reUsing     = regexp.MustCompile(`^\s*using\s+([a-zA-Z0-9_.]+);`)
	reImport    = regexp.MustCompile(`^\s*import\s+([^;]+);?`)
)

// ParseHeuristic provides reliable, non-fabricating symbol extraction for
// languages where a tree-sitter grammar is not compiled into CodeDNA.
func ParseHeuristic(relPath, lang string, src []byte) (*model.FileParseResult, error) {
	normPath := filepath.ToSlash(relPath)
	res := &model.FileParseResult{
		File:        normPath,
		Language:    lang,
		Symbols:     []model.Symbol{},
		Imports:     []model.Import{},
		Calls:       []model.Call{},
		BranchCount: 0,
		ParseErrors: []model.ParseError{
			{
				File:      normPath,
				Language:  lang,
				Message:   fmt.Sprintf("Heuristic static analysis active for %s (Tree-sitter grammar not compiled)", lang),
				Recovered: true,
			},
		},
	}

	sc := bufio.NewScanner(bytes.NewReader(src))
	lineNum := 0

	for sc.Scan() {
		lineNum++
		line := sc.Text()
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "//") || strings.HasPrefix(trimmed, "/*") || strings.HasPrefix(trimmed, "*") {
			continue
		}

		// Imports / Includes
		var importPath string
		if m := reCInclude.FindStringSubmatch(line); len(m) > 1 {
			importPath = m[1]
		} else if m := reRustUse.FindStringSubmatch(line); len(m) > 1 {
			importPath = strings.TrimSpace(m[1])
		} else if m := reRubyReq.FindStringSubmatch(line); len(m) > 1 {
			importPath = m[1]
		} else if m := reUsing.FindStringSubmatch(line); len(m) > 1 {
			importPath = m[1]
		} else if m := reImport.FindStringSubmatch(line); len(m) > 1 {
			importPath = strings.TrimSpace(m[1])
		}

		if importPath != "" {
			res.Imports = append(res.Imports, model.Import{
				Path: importPath,
				Line: lineNum,
				Col:  0,
			})
			continue
		}

		// Classes / Types / Structs
		var typeName string
		var kind model.SymbolKind = model.SymbolClass

		if m := reClass.FindStringSubmatch(line); len(m) > 1 {
			typeName = m[1]
			if strings.Contains(line, "interface") || strings.Contains(line, "trait") {
				kind = model.SymbolInterface
			} else if strings.Contains(line, "struct") {
				kind = model.SymbolStruct
			}
		} else if m := reRubyClass.FindStringSubmatch(line); len(m) > 1 {
			typeName = m[1]
		} else if m := reSQLTable.FindStringSubmatch(line); len(m) > 1 {
			typeName = m[1]
			kind = model.SymbolStruct
		} else if m := reGQLType.FindStringSubmatch(line); len(m) > 1 {
			typeName = m[1]
			kind = model.SymbolType
		}

		if typeName != "" && isValidIdent(typeName) {
			symID := fmt.Sprintf("%s:%s:%d", normPath, typeName, lineNum)
			res.Symbols = append(res.Symbols, model.Symbol{
				ID:            symID,
				Name:          typeName,
				QualifiedName: typeName,
				Kind:          kind,
				File:          normPath,
				Language:      lang,
				Line:          lineNum,
				Column:        0,
				EndLine:       lineNum,
				Exported:      true,
				LineCount:     1,
			})
			continue
		}

		// Functions / Methods / Procedures
		var funcName string
		if m := reRustFn.FindStringSubmatch(line); len(m) > 1 {
			funcName = m[1]
		} else if m := reRubyDef.FindStringSubmatch(line); len(m) > 1 {
			funcName = m[1]
		} else if m := reKotlinFun.FindStringSubmatch(line); len(m) > 1 {
			funcName = m[1]
		} else if m := reSwiftFunc.FindStringSubmatch(line); len(m) > 1 {
			funcName = m[1]
		} else if m := reSQLProc.FindStringSubmatch(line); len(m) > 1 {
			funcName = m[1]
		} else if m := reCMethod.FindStringSubmatch(line); len(m) > 1 {
			cand := m[1]
			if cand != "if" && cand != "for" && cand != "while" && cand != "switch" && cand != "return" {
				funcName = cand
			}
		}

		if funcName != "" && isValidIdent(funcName) {
			symID := fmt.Sprintf("%s:%s:%d", normPath, funcName, lineNum)
			res.Symbols = append(res.Symbols, model.Symbol{
				ID:            symID,
				Name:          funcName,
				QualifiedName: funcName,
				Kind:          model.SymbolFunction,
				File:          normPath,
				Language:      lang,
				Line:          lineNum,
				Column:        0,
				EndLine:       lineNum,
				Exported:      true,
				LineCount:     1,
			})
		}
	}

	return res, nil
}

func isValidIdent(s string) bool {
	if s == "" || len(s) > 100 {
		return false
	}
	for i, r := range s {
		if i == 0 && (r >= '0' && r <= '9') {
			return false
		}
		if (r < 'a' || r > 'z') && (r < 'A' || r > 'Z') && (r < '0' || r > '9') && r != '_' && r != '.' && r != ':' && r != '!' && r != '?' {
			return false
		}
	}
	return true
}
