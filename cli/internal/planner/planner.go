package planner

import "github.com/CharukaK/probe/cli/internal/parser/ast"

type InstructionStepType int

const (
	RequestInstructionType InstructionStepType = iota
)

type InstructionStep interface {
	Type() InstructionStepType
}

type Plan struct {
	Steps []InstructionStep
}

func GeneratePlan(rootNode ast.SourceFileNode) (*Plan, error) {
	plan := &Plan{}

	return plan, nil
}
