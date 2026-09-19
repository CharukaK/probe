package plan

import (
	"github.com/charukak/probe/cli/internal/parser/ast"
)

type TargetSegmentKind int

const (
	TargetLiteral TargetSegmentKind = iota
	TargetInterpollation
)

type TargetSegment struct {
	Kind  TargetSegmentKind
	Text  string
	Value Value
}

type RequestLine struct {
	Method  string
	Target  []TargetSegment
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

type RequestDirective struct {
	Name    string // optional, from a preceding seprator/request_name
	Request *RequestLine
	Headers []HeaderLine
	Body    MessageBody
}

func (rd *RequestDirective) Type() InstructionType {
	return RequestDirectiveKind
}

func newRequestDirective(node *ast.RequestBlockNode, source []byte) (*RequestDirective, error) {
	rd := &RequestDirective{}

	line, err := requestLine(*node.RequestLine, source)
	if err != nil {
		return nil, err
	}

	rd.Request = line

	return rd, nil
}

func requestLine(node ast.RequestLineNode, source []byte) (*RequestLine, error) {
	rl := &RequestLine{}

	rl.Method = node.Method.Utf8Text(source)

	target := node.Target
	for i := uint(0); i < target.NamedChildCount(); i++ {
		child := target.NamedChild(i)

		switch child.Kind() {
		case string(ast.TargetTextKind):
			rl.Target = append(rl.Target, TargetSegment{
				Kind: TargetLiteral,
				Text: child.Utf8Text(source),
			})
		case string(ast.InterpolationKind):
			value, err := letInterpolationValue(ast.NewInterpolationNode(child), source)
			if err != nil {
				return nil, err
			}
			rl.Target = append(rl.Target, TargetSegment{
				Kind:  TargetLiteral,
				Value: value,
			})
		case string(ast.StartParenthesisKind):
			rl.Target = append(rl.Target, TargetSegment{
				Kind: TargetLiteral,
				Text: child.Utf8Text(source),
			})
		}

	}

	return rl, nil
}
