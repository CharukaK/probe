package parser

import (
	"strings"

	ts_probe "github.com/charukak/probe/tree-sitter-probe/bindings/go"
	ts "github.com/tree-sitter/go-tree-sitter"
)

type NodeKind string

const (
	RequestBlockNodeKind NodeKind = "request_block"
	RequestLineKind      NodeKind = "request_line"
	RequestMethodKind    NodeKind = "method"
	RequestTargetKind    NodeKind = "request_target"
	RequestVersionKind   NodeKind = "http_version"
	FieldLineKind        NodeKind = "field_line"
	FieldNameKind        NodeKind = "field_name"
	FieldValueKind       NodeKind = "field_value"
	MessageBodyKind      NodeKind = "message_body"
	OctetBodyKind        NodeKind = "octet_body"
	MultipartBodyKind    NodeKind = "multipart_body"
	MultipartPartKind    NodeKind = "multipart_part"
)

type SourceFile struct {
	Children []RequestBlock
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
		Children: make([]RequestBlock, 0),
	}

	for _, child := range children {
		switch child.Kind() {
		case string(RequestBlockNodeKind):
			requestBlock := walkRequestBlock(&child, source)
			sf.Children = append(sf.Children, *requestBlock)
		}
	}

	return sf
}

func walkRequestBlock(node *ts.Node, source []byte) *RequestBlock {
	reqBlock := &RequestBlock{}

	cursor := node.Walk()
	children := node.Children(cursor)

	for _, child := range children {
		switch child.Kind() {
		case string(RequestLineKind):
			reqLine := walkRequestLine(&child, source)
			reqBlock.RequestLine.Method = reqLine.Method
			reqBlock.RequestLine.Target = reqLine.Target
			reqBlock.RequestLine.Version = reqLine.Version
		case string(FieldLineKind):
			headerLine := walkFieldLine(&child, source)
			reqBlock.HeaderLines = append(reqBlock.HeaderLines, *headerLine)
		case string(MessageBodyKind):
			body := walkMessageBody(&child, source)
			reqBlock.Body = *body
		}
	}

	return reqBlock
}

func walkFieldLine(node *ts.Node, source []byte) *HeaderLine {
	headerLine := &HeaderLine{}

	cursor := node.Walk()
	children := node.Children(cursor)

	for _, child := range children {
		switch child.Kind() {
		case string(FieldNameKind):
			headerLine.Key = child.Utf8Text(source)
		case string(FieldValueKind):
			headerLine.Value = child.Utf8Text(source)
		}
	}

	return headerLine
}

func walkMessageBody(node *ts.Node, source []byte) *MessageBody {
	body := &MessageBody{}

	cursor := node.Walk()
	children := node.Children(cursor)

	for _, child := range children {
		switch child.Kind() {
		case string(OctetBodyKind):
			body.Octets = []byte(child.Utf8Text(source))
		case string(MultipartBodyKind):
			body.MultipartEntries = walkMultipartBody(&child, source)
		}
	}

	return body
}

func walkMultipartBody(node *ts.Node, source []byte) []MultipartEntry {
	entries := make([]MultipartEntry, 0)

	cursor := node.Walk()
	children := node.Children(cursor)

	for _, child := range children {
		if child.Kind() != string(MultipartPartKind) {
			continue
		}
		entries = append(entries, walkMultipartPart(&child, source))
	}

	return entries
}

// walkMultipartPart parses a raw multipart_part node, e.g. "@field username alice\r\n"
// or "@file avatar ./avatar.png\n". The grammar doesn't split the directive keyword,
// key, and value into separate fields, so that's done here from the raw text.
func walkMultipartPart(node *ts.Node, source []byte) MultipartEntry {
	entry := MultipartEntry{}

	text := strings.TrimRight(node.Utf8Text(source), "\r\n")

	switch {
	case strings.HasPrefix(text, "@field"):
		entry.EntryType = FieldType
		text = strings.TrimPrefix(text, "@field")
	case strings.HasPrefix(text, "@file"):
		entry.EntryType = FileType
		text = strings.TrimPrefix(text, "@file")
	}

	fields := strings.Fields(text)
	if len(fields) > 0 {
		entry.Key = fields[0]
	}
	if len(fields) > 1 {
		entry.Value = strings.Join(fields[1:], " ")
	}

	return entry
}

func walkRequestLine(node *ts.Node, source []byte) *RequestLine {
	reqLine := &RequestLine{}

	cursor := node.Walk()
	children := node.Children(cursor)

	for _, child := range children {
		switch child.Kind() {
		case string(RequestMethodKind):
			reqLine.Method = child.Utf8Text(source)
		case string(RequestTargetKind):
			reqLine.Target = child.Utf8Text(source)
		case string(RequestVersionKind):
			reqLine.Version = child.Utf8Text(source)
		}
	}

	return reqLine
}
