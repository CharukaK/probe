package tree_sitter_http_test

import (
	"testing"

	tree_sitter "github.com/tree-sitter/go-tree-sitter"
	tree_sitter_http "github.com/charukak/probe/bindings/go"
)

func TestCanLoadGrammar(t *testing.T) {
	language := tree_sitter.NewLanguage(tree_sitter_http.Language())
	if language == nil {
		t.Errorf("Error loading HTTP grammar")
	}
}
