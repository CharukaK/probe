package plan

import (
	"fmt"

	"github.com/charukak/probe/cli/internal/parser/ast"
)

type ValueKind int

type InstructionType int

type TargetKind int

const (
	LiteralValue ValueKind = iota
	VariableRef
	UtilCall
)

const (
	LetDirectiveKind InstructionType = iota
	RequestDirectiveKind
	SaveDirectiveKind
	AssertDirectiveKind
)

const (
	HeaderTarget TargetKind = iota
	BodyTarget
)

type Target struct {
	Kind       TargetKind
	HeaderName string
	Accessors  []string
}

type Value struct {
	Kind      ValueKind
	Literal   string
	Root      string
	Accessors []string
	UtilName  string
	Args      []Value
}

type Plan struct {
	Steps []InstructionStep
}

type InstructionStep interface {
	Type() InstructionType
}

func FromAst(sourceFile ast.SourceFileNode, source []byte) (*Plan, error) {
	p := &Plan{}

	for _, let := range sourceFile.Let {
		ld, err := newLetDirective(*let, source)

		if err != nil {
			return nil, err
		}

		p.Steps = append(p.Steps, ld)
	}

	for _, b:= range sourceFile.Block {
		// TODO: request blocks
		rd, err := newRequestDirective(b, source)
		if err != nil {
			return nil, err 
		}

		for _, t := range rd.Request.Target {
			fmt.Printf("%v\n", t)
		}
	}

	return p, nil
}
