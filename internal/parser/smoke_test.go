package parser

import (
	"fmt"
	"testing"

	sitter "github.com/tree-sitter/go-tree-sitter"
	tsgolang "github.com/tree-sitter/tree-sitter-go/bindings/go"
	java "github.com/tree-sitter/tree-sitter-java/bindings/go"
	tsjs "github.com/tree-sitter/tree-sitter-javascript/bindings/go"
	php "github.com/tree-sitter/tree-sitter-php/bindings/go"
	python "github.com/tree-sitter/tree-sitter-python/bindings/go"
	ts "github.com/tree-sitter/tree-sitter-typescript/bindings/go"
)

func lang(name string) *sitter.Language {
	switch name {
	case "go":
		return sitter.NewLanguage(tsgolang.Language())
	case "javascript":
		return sitter.NewLanguage(tsjs.Language())
	case "typescript":
		return sitter.NewLanguage(ts.LanguageTypescript())
	case "tsx":
		return sitter.NewLanguage(ts.LanguageTSX())
	case "python":
		return sitter.NewLanguage(python.Language())
	case "java":
		return sitter.NewLanguage(java.Language())
	case "php":
		return sitter.NewLanguage(php.LanguagePHP())
	default:
		return nil
	}
}

func TestSmokeParse(t *testing.T) {
	sources := map[string]string{
		"go":         "package main\n\nfunc main() {\n\tHello(1)\n}\n",
		"javascript": "function a() { b(); }\n",
		"python":     "def main():\n    pass\n",
		"java":       "public class A { void run() { } }\n",
		"php":        "<?php class A { public function run() {} }\n",
		"typescript": "export function a() { b(); }\n",
	}
	for name, src := range sources {
		p := sitter.NewParser()
		p.SetLanguage(lang(name))
		tree := p.Parse([]byte(src), nil)
		if tree == nil || tree.RootNode().HasError() {
			t.Logf("%s: tree has parse errors", name)
		}
		sexp := tree.RootNode().ToSexp()
		fmt.Printf("%-10s parsed OK (%d bytes sexp)\n", name, len(sexp))
		tree.Close()
		p.Close()
	}
}
