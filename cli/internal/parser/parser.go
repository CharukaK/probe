package parser

//go:generate go run ./gen/main.go -in ../../../tree-sitter-probe/src/node-types.json -out ast/ast.go -pkg ast
import (
	"github.com/charukak/probe/cli/internal/parser/ast"
	ts_probe "github.com/charukak/probe/tree-sitter-probe/bindings/go"
	ts "github.com/tree-sitter/go-tree-sitter"
)

// ParseResult carries both the raw tree-sitter tree — needed for
// incremental reparsing on edits and for error/diagnostic queries (an LSP
// use case) — and the populated, typed root that consumers like the
// planner and executor work against.
type ParseResult struct {
	Tree *ts.Tree
	Root ast.SourceFileNode
}

// Parse parses source and returns both the raw tree and its populated
// ast.SourceFileNode. Callers still need to keep source around themselves:
// every ast node is a byte-offset view into it, not an owner of its text.
func Parse(source []byte) ParseResult {
	lang := ts.NewLanguage(ts_probe.Language())
	p := ts.NewParser()
	p.SetLanguage(lang)
	tree := p.Parse(source, nil)

	return ParseResult{
		Tree: tree,
		Root: ast.NewSourceFileNode(tree.RootNode()),
	}
}
