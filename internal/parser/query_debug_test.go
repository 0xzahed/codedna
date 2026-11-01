package parser

import (
	"testing"

	sitter "github.com/tree-sitter/go-tree-sitter"
	java "github.com/tree-sitter/tree-sitter-java/bindings/go"
	python "github.com/tree-sitter/tree-sitter-python/bindings/go"
)

func tryQuery(t *testing.T, name, q string, l *sitter.Language) {
	_, err := sitter.NewQuery(l, q)
	if err != nil {
		t.Logf("%-28s FAIL %v", name, err)
	} else {
		t.Logf("%-28s OK", name)
	}
}

func TestQueryDebug(t *testing.T) {
	jv := sitter.NewLanguage(java.Language())
	py := sitter.NewLanguage(python.Language())

	tryQuery(t, "java re-order", `(method_invocation object: (_) @r name: (identifier) @c)`, jv)
	tryQuery(t, "java obj only", `(method_invocation object: (_) @r)`, jv)
	tryQuery(t, "java 2 patterns", `
(method_invocation name: (identifier) @c)
(method_invocation object: (_) @r name: (identifier) @c)`, jv)
	tryQuery(t, "py reorder", `(call function: (attribute object: (_) @r attribute: (identifier) @c))`, py)
	tryQuery(t, "py reorder 2pat", `
(call function: (identifier) @c)
(call function: (attribute object: (_) @r attribute: (identifier) @c))`, py)
	tryQuery(t, "py 3 field obj", `(call function: (attribute object: (identifier) @r attribute: (identifier) @c))`, py)
}
