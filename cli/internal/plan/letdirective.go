package plan

import (
	"fmt"

	"github.com/charukak/probe/cli/internal/parser/ast"
)

type LetDirective struct {
	Name  string
	Value Value
}

func (ls *LetDirective) Type() InstructionType {
	return LetDirectiveKind
}

func newLetDirective(node ast.LetDirectiveNode, source []byte) (*LetDirective, error) {
	val, err := letValue(node.Target, source)
	if err != nil {
		return nil, err
	}

	return &LetDirective{
		Name:  node.Name.Utf8Text(source),
		Value: val,
	}, nil
}

// letValue resolves a let_value node — a union of json_literal and
// interpolation — into a plan Value.
func letValue(node *ast.LetValueNode, source []byte) (Value, error) {
	switch node.Value.Kind() {
	case string(ast.JsonLiteralKind):
		return jsonLiteralValue(ast.NewJsonLiteralNode(node.Value), source), nil
	case string(ast.InterpolationKind):
		return letInterpolationValue(ast.NewInterpolationNode(node.Value), source)
	default:
		return Value{}, fmt.Errorf("let: unexpected value kind %q", node.Value.Kind())
	}
}

func jsonLiteralValue(node ast.JsonLiteralNode, source []byte) Value {
	// json_literal: 'null' | 'true' | 'false' | value: number | value: string
	if node.Value != nil {
		return Value{Kind: LiteralValue, Literal: node.Value.Utf8Text(source)}
	}
	return Value{Kind: LiteralValue, Literal: node.Utf8Text(source)} // bare null/true/false
}

func letInterpolationValue(node ast.InterpolationNode, source []byte) (Value, error) {
	body := node.Body
	switch body.Ref.Kind() {
	case string(ast.ValueReferenceKind):
		ref := ast.NewValueReferenceNode(body.Ref)
		accessors := make([]string, 0, len(ref.Accessor))
		for _, acc := range ref.Accessor {
			accessors = append(accessors, acc.Utf8Text(source))
		}
		return Value{
			Kind:      VariableRef,
			Root:      ref.Root.Utf8Text(source),
			Accessors: accessors,
		}, nil
	case string(ast.UtilReferenceKind):
		ref := ast.NewUtilReferenceNode(body.Ref)
		return Value{
			Kind:     UtilCall,
			UtilName: ref.UtilName.Utf8Text(source),
			// Args: walk ref's `arguments` child similarly, once needed
		}, nil
	default:
		return Value{}, fmt.Errorf("let: unexpected interpolation body kind %q", body.Ref.Kind())
	}
}
