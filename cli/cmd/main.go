package main

import (
	"fmt"
	"os"

	ts_probe "github.com/charukak/probe/tree-sitter-probe/bindings/go"
	ts "github.com/tree-sitter/go-tree-sitter"
)

func main() {
	parser := ts.NewParser()
	defer parser.Close()

	parser.SetLanguage(ts.NewLanguage(ts_probe.Language()))

	source, err := os.ReadFile("./hello.probe")

	if err != nil {
		return
	}


	// Parse the string into an Abstract Syntax Tree (AST)
	tree := parser.Parse(source, nil)
	if tree == nil {
		panic("Failed to parse code.")
	}

	// CRITICAL: Manually close the tree when finished
	defer tree.Close()

	// Retrieve the root node of the syntax tree
	rootNode := tree.RootNode()

	// Output the S-expression format of the AST
	fmt.Println("AST Structure:")
	fmt.Println(rootNode.ToSexp())
}
