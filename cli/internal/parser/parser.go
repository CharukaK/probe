package parser

//go:generate go run ./gen/main.go -in ../../../tree-sitter-probe/src/node-types.json -out ast/ast.go -pkg ast
import (
	ts_probe "github.com/charukak/probe/tree-sitter-probe/bindings/go"
	ts "github.com/tree-sitter/go-tree-sitter"
)

func RootNodeFromSource(s []byte) *ts.Node {
	lang := ts.NewLanguage(ts_probe.Language())
	parser := ts.NewParser()
	parser.SetLanguage(lang)
	tree := parser.Parse(s, nil)

	return tree.RootNode()
}
