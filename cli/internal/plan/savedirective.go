package plan

type SaveDirective struct {
	Name   string
	Target Target
}

func (sd *SaveDirective) Type() InstructionType {
	return SaveDirectiveKind
}
