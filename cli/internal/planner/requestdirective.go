package planner

import "github.com/CharukaK/probe/cli/internal/parser/ast"

type RequestDirective struct {
}

func (rd *RequestDirective) Type() InstructionStepType {
	return RequestInstructionType
}

func NewRequestDirective(node ast.RequestNode) *RequestDirective {
	return &RequestDirective{
	}
}
