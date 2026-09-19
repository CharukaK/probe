package plan

type AssertDirective struct {
	Comparator string
	Target     Target
	Expected   Value
}

func (ad *AssertDirective) Type() InstructionType {
	return AssertDirectiveKind
}
