package tree_sitter_probe_test

import (
	"testing"

	tree_sitter "github.com/tree-sitter/go-tree-sitter"
	tree_sitter_probe "github.com/charukak/probe/bindings/go"
)

func TestCanLoadGrammar(t *testing.T) {
	language := tree_sitter.NewLanguage(tree_sitter_probe.Language())
	if language == nil {
		t.Errorf("Error loading Probe grammar")
	}
}
