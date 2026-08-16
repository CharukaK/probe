package parser

import (
	"fmt"

	ts_probe "github.com/charukak/probe/tree-sitter-probe/bindings/go"
	ts "github.com/tree-sitter/go-tree-sitter"
)

type NodeKind string

const (
	RequestBlockNodeKind NodeKind = "request_block"
)

type SourceFile struct {
	children []RequestBlock
}

type RequestBlock struct {
	RequestLine RequestLine
	HeaderLines []HeaderLine
	Body        MessageBody
}

type RequestLine struct {
	Method  string
	Target  string
	Version string // optional; empty when the request line omits "HTTP/major.minor"
}

type HeaderLine struct {
	Key   string
	Value string
}

type MultipartEntryType int

const (
	FieldType MultipartEntryType = iota
	FileType
)

type MultipartEntry struct {
	EntryType MultipartEntryType
	Key       string
	Value     string
}

type MessageBody struct {
	Octets           []byte
	MultipartEntries []MultipartEntry
}

func Parse(source []byte) (*SourceFile, error) {
	parser := ts.NewParser()
	defer parser.Close()

	parser.SetLanguage(ts.NewLanguage(ts_probe.Language()))

	// Parse the string into an Abstract Syntax Tree (AST)
	tree := parser.Parse(source, nil)
	if tree == nil {
		panic("Failed to parse code.")
	}

	// CRITICAL: Manually close the tree when finished
	defer tree.Close()

	// Retrieve the root node of the syntax tree
	rootNode := tree.RootNode()

	return walkSourceFile(rootNode, source), nil
}

func walkSourceFile(node *ts.Node, source []byte) *SourceFile {
	cursor := node.Walk()
	children := node.Children(cursor)
	sf := &SourceFile{
		children: make([]RequestBlock, 0),
	}

	for _, child := range children {
		switch child.Kind() {
		case string(RequestBlockNodeKind):
			requestBlock := walkRequestBlock(&child, source)
			sf.children = append(sf.children, *requestBlock)
		}
	}

	return sf
}

func walkRequestBlock(node *ts.Node, source []byte) *RequestBlock {
	reqBlock := &RequestBlock{}

	cursor := node.Walk()
	children := node.Children(cursor)

	for _, child := range children {
		fmt.Println(child.Kind())
	}


	return reqBlock
}

