package compiler

// EmittedInstruction represents an opcode
// with its emitted position.
type EmittedInstruction struct {
	Opcode   Opcode
	Position int
}
