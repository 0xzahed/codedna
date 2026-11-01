package parser

import (
	"fmt"
	"sync"

	sitter "github.com/tree-sitter/go-tree-sitter"
	golang "github.com/tree-sitter/tree-sitter-go/bindings/go"
	java "github.com/tree-sitter/tree-sitter-java/bindings/go"
	tsjs "github.com/tree-sitter/tree-sitter-javascript/bindings/go"
	php "github.com/tree-sitter/tree-sitter-php/bindings/go"
	python "github.com/tree-sitter/tree-sitter-python/bindings/go"
	ts "github.com/tree-sitter/tree-sitter-typescript/bindings/go"
)

// LanguageSpec bundles a tree-sitter language with the queries CodeDNA uses.
type LanguageSpec struct {
	Name        string // canonical language name (or lang/tsx variant)
	BaseLang    string // canonical language name (typescript-tsx -> typescript)
	Extensions  []string
	Lang        *sitter.Language
	SymbolQuery string
	ImportQuery string
	CallQuery   string
	BranchQuery string
	NewQuery    bool // grammar understands new_expression
}

// languages returns canonical language names supported by the parser.
func languages() []string {
	return []string{"go", "python", "javascript", "typescript", "java", "php"}
}

// LanguageForExt maps a file extension to a canonical language name.
func LanguageForExt(ext string) string {
	switch ext {
	case ".go":
		return "go"
	case ".py", ".pyi":
		return "python"
	case ".js", ".mjs", ".cjs", ".jsx":
		return "javascript"
	case ".ts", ".mts", ".cts":
		return "typescript"
	case ".tsx":
		return "typescript"
	case ".java":
		return "java"
	case ".php":
		return "php"
	default:
		return ""
	}
}

// grammarKey distinguishes tsx from ts.
func grammarKey(lang, ext string) string {
	if lang == "typescript" && ext == ".tsx" {
		return "typescript/tsx"
	}
	return lang
}

var (
	specs     map[string]*LanguageSpec // by grammarKey
	specOnce  sync.Once
	specError error
)

type compiled struct {
	symQuery, importQuery, callQuery, branchQuery *sitter.Query
}

var (
	compiledQ     map[string]*compiled
	compiledQOnce sync.Once
	compiledQErr  error
)

func getSpecs() (map[string]*LanguageSpec, error) {
	specOnce.Do(buildSpecs)
	return specs, specError
}

func getCompiled() (map[string]*compiled, error) {
	compiledQOnce.Do(buildCompiled)
	return compiledQ, compiledQErr
}

func buildSpecs() {
	specs = map[string]*LanguageSpec{}
	set := func(spec *LanguageSpec) { specs[spec.Name] = spec }

	// --- Go ---
	set(&LanguageSpec{
		Name:       "go",
		BaseLang:   "go",
		Extensions: []string{".go"},
		Lang:       sitter.NewLanguage(golang.Language()),
		SymbolQuery: `
			(function_declaration name: (identifier) @name) @decl
			(method_declaration name: (field_identifier) @name) @decl
			(type_declaration (type_spec name: (type_identifier) @name type: (struct_type)) @decl)
			(type_declaration (type_spec name: (type_identifier) @name type: (interface_type)) @decl)
			(type_declaration (type_spec name: (type_identifier) @name type: [(slice_type) (map_type) (array_type) (channel_type) (pointer_type)]) @decl)
			(source_file (package_clause (package_identifier) @package_name))
		`,
		ImportQuery: `
			(import_spec path: (interpreted_string_literal) @import_path)
			(import_spec name: (package_identifier) @import_alias path: (interpreted_string_literal) @import_path)
		`,
		CallQuery: `
			(call_expression function: (identifier) @call_callee)
			(call_expression function: (selector_expression operand: (_) @call_recv field: (field_identifier) @call_callee))
		`,
		BranchQuery: `
			(if_statement) @branch
			(for_statement) @branch
			(expression_case) @branch
			(type_case) @branch
			(communication_case) @branch
			(binary_expression operator: "&&") @branch
			(binary_expression operator: "||") @branch
			(select_statement) @branch
		`,
	})

	// --- Python ---
	set(&LanguageSpec{
		Name:       "python",
		BaseLang:   "python",
		Extensions: []string{".py", ".pyi"},
		Lang:       sitter.NewLanguage(python.Language()),
		SymbolQuery: `
			(function_definition name: (identifier) @name) @decl
			(class_definition name: (identifier) @name) @decl
		`,
		ImportQuery: `
			(import_statement (dotted_name) @import_path)
			(import_from_statement module_name: (dotted_name) @import_path (dotted_name) @import_name)
			(import_from_statement module_name: (relative_import) @import_path)
		`,
		CallQuery: `
			(call function: (identifier) @call_callee)
			(call function: (attribute object: (_) @call_recv attribute: (identifier) @call_callee))
		`,
		BranchQuery: `
			(if_statement) @branch
			(for_statement) @branch
			(while_statement) @branch
			(except_clause) @branch
			(boolean_operator) @branch
			(for_in_clause) @branch
		`,
	})

	// --- JavaScript ---
	set(&LanguageSpec{
		Name:       "javascript",
		BaseLang:   "javascript",
		Extensions: []string{".js", ".mjs", ".cjs", ".jsx"},
		Lang:       sitter.NewLanguage(tsjs.Language()),
		SymbolQuery: `
			(function_declaration name: (identifier) @name) @decl
			(class_declaration name: (identifier) @name) @decl
			(method_definition name: [(property_identifier) (private_property_identifier)] @name) @decl
			(lexical_declaration (variable_declarator name: (identifier) @name value: [(arrow_function) (function_expression)])) @decl
		`,
		ImportQuery: `
			(import_statement source: (string) @import_path)
			(export_statement source: (string) @import_path)
			(call_expression function: (identifier) @req arguments: (arguments (string) @import_path) (#match? @req "require"))
		`,
		CallQuery: `
			(call_expression function: (identifier) @call_callee)
			(call_expression function: (member_expression object: (_) @call_recv property: (property_identifier) @call_callee))
			(new_expression constructor: (identifier) @call_callee)
			(new_expression constructor: (member_expression object: (_) @call_recv property: (property_identifier) @call_callee))
		`,
		NewQuery: true,
		BranchQuery: `
			(if_statement) @branch
			(for_statement) @branch
			(while_statement) @branch
			(do_statement) @branch
			(switch_case) @branch
			(catch_clause) @branch
			(ternary_expression) @branch
			(binary_expression operator: "&&") @branch
			(binary_expression operator: "||") @branch
		`,
	})

	// --- TypeScript (ts and tsx variants share query text) ---
	tsSymbol := `
			(function_declaration name: (identifier) @name) @decl
			(class_declaration name: (type_identifier) @name) @decl
			(interface_declaration name: (type_identifier) @name) @decl
			(type_alias_declaration name: (type_identifier) @name) @decl
			(enum_declaration name: (identifier) @name) @decl
			(method_definition name: [(property_identifier) (private_property_identifier)] @name) @decl
			(lexical_declaration (variable_declarator name: (identifier) @name value: [(arrow_function) (function_expression)])) @decl
	`
	tsImport := `
			(import_statement source: (string) @import_path)
			(export_statement source: (string) @import_path)
			(call_expression function: (identifier) @req arguments: (arguments (string) @import_path) (#match? @req "require"))
	`
	tsCall := `
			(call_expression function: (identifier) @call_callee)
			(call_expression function: (member_expression object: (_) @call_recv property: (property_identifier) @call_callee))
			(new_expression constructor: (identifier) @call_callee)
			(new_expression constructor: (member_expression object: (_) @call_recv property: (property_identifier) @call_callee))
	`
	tsBranch := `
			(if_statement) @branch
			(for_statement) @branch
			(while_statement) @branch
			(do_statement) @branch
			(switch_case) @branch
			(catch_clause) @branch
			(ternary_expression) @branch
			(binary_expression operator: "&&") @branch
			(binary_expression operator: "||") @branch
	`
	set(&LanguageSpec{
		Name:     "typescript",
		BaseLang: "typescript",
		Extensions: []string{".ts", ".mts", ".cts"},
		Lang: sitter.NewLanguage(ts.LanguageTypescript()),
		SymbolQuery: tsSymbol, ImportQuery: tsImport, CallQuery: tsCall, BranchQuery: tsBranch,
		NewQuery: true,
	})
	set(&LanguageSpec{
		Name:     "typescript/tsx",
		BaseLang: "typescript",
		Extensions: []string{".tsx"},
		Lang: sitter.NewLanguage(ts.LanguageTSX()),
		SymbolQuery: tsSymbol, ImportQuery: tsImport, CallQuery: tsCall, BranchQuery: tsBranch,
		NewQuery: true,
	})

	// --- Java ---
	set(&LanguageSpec{
		Name:       "java",
		BaseLang:   "java",
		Extensions: []string{".java"},
		Lang:       sitter.NewLanguage(java.Language()),
		SymbolQuery: `
			(class_declaration name: (identifier) @name) @decl
			(interface_declaration name: (identifier) @name) @decl
			(method_declaration name: (identifier) @name) @decl
			(constructor_declaration name: (identifier) @name) @decl
			(record_declaration name: (identifier) @name) @decl
		`,
		ImportQuery: `
			(import_declaration) @import_stmt
		`,
		CallQuery: `
			(method_invocation object: (_) @call_recv name: (identifier) @call_callee)
			(method_invocation name: (identifier) @call_callee)
			(object_creation_expression [(type_identifier) (scoped_type_identifier) (generic_type)] @call_callee)
		`,
		NewQuery: true,
		BranchQuery: `
			(if_statement) @branch
			(for_statement) @branch
			(enhanced_for_statement) @branch
			(while_statement) @branch
			(do_statement) @branch
			(switch_block_statement_group) @branch
			(catch_clause) @branch
			(binary_expression operator: "&&") @branch
			(binary_expression operator: "||") @branch
		`,
	})

	// --- PHP ---
	set(&LanguageSpec{
		Name:       "php",
		BaseLang:   "php",
		Extensions: []string{".php"},
		Lang:       sitter.NewLanguage(php.LanguagePHP()),
		SymbolQuery: `
			(function_definition name: (name) @name) @decl
			(class_declaration name: (name) @name) @decl
			(interface_declaration name: (name) @name) @decl
			(method_declaration name: (name) @name) @decl
			(trait_declaration name: (name) @name) @decl
		`,
		ImportQuery: `
			(namespace_use_declaration) @import_stmt
		`,
		CallQuery: `
			(function_call_expression function: [(name) (qualified_name)] @call_callee)
			(member_call_expression object: (_) @call_recv name: (name) @call_callee)
			(scoped_call_expression scope: (_) @call_recv name: (name) @call_callee)
			(object_creation_expression [(name) (class_constant_access_expression)] @call_callee)
		`,
		NewQuery: true,
		BranchQuery: `
			(if_statement) @branch
			(else_clause) @branch
			(else_if_clause) @branch
			(while_statement) @branch
			(for_statement) @branch
			(foreach_statement) @branch
			(case_statement) @branch
			(catch_clause) @branch
		`,
	})
}

func buildCompiled() {
	compiledQ = map[string]*compiled{}
	c := &compiled{}
	compile := func(key, sym, imp, call, branch string) error {
		sq, err := sitter.NewQuery(specs[key].Lang, sym)
		if err != nil {
			return fmt.Errorf("symbol query %s: %v", key, err)
		}
		iq, err := sitter.NewQuery(specs[key].Lang, imp)
		if err != nil {
			return fmt.Errorf("import query %s: %v", key, err)
		}
		cq, err := sitter.NewQuery(specs[key].Lang, call)
		if err != nil {
			return fmt.Errorf("call query %s: %v", key, err)
		}
		bq, err := sitter.NewQuery(specs[key].Lang, branch)
		if err != nil {
			return fmt.Errorf("branch query %s: %v", key, err)
		}
		compiledQ[key] = &compiled{sq, iq, cq, bq}
		return nil
	}
	_ = c
	for _, spec := range specs {
		if err := compile(spec.Name, spec.SymbolQuery, spec.ImportQuery, spec.CallQuery, spec.BranchQuery); err != nil {
			compiledQErr = err
			return
		}
	}
}
