package compiler

import (
	"fmt"
)

// MakeInstruction returns a bytecode for an opcode and the operands.
func MakeInstruction(opcode Opcode, operands ...int) []byte {
	numOperands := OpcodeOperands[opcode]

	totalLen := 1
	for _, w := range numOperands {
		totalLen += w
	}

	instruction := make([]byte, totalLen)
	instruction[0] = byte(opcode)

	offset := 1
	for i, o := range operands {
		width := numOperands[i]
		switch width {
		case 1:
			instruction[offset] = byte(o)
		case 2:
			n := uint16(o)
			instruction[offset] = byte(n >> 8)
			instruction[offset+1] = byte(n)
		}
		offset += width
	}

	return instruction
}

// FormatInstructions returns string representation of
// bytecode instructions.
func FormatInstructions(b []byte, posOffset int) []string {
	var out []string

	i := 0
	for i < len(b) {
		numOperands := OpcodeOperands[Opcode(b[i])]
		operands, read := ReadOperands(numOperands, b[i+1:])

		switch len(numOperands) {
		case 0:
			out = append(out, fmt.Sprintf("%04d %-7s", posOffset+i, OpcodeNames[Opcode(b[i])]))
		case 1:
			out = append(out, fmt.Sprintf("%04d %-7s %-5d", posOffset+i, OpcodeNames[Opcode(b[i])], operands[0]))
		case 2:
			out = append(out, fmt.Sprintf("%04d %-7s %-5d %-5d", posOffset+i, OpcodeNames[Opcode(b[i])], operands[0], operands[1]))
		}

		i += 1 + read
	}

	return out
}

func iterateInstructions(b []byte, fn func(pos int, opcode Opcode, operands []int) bool) {
	for i := 0; i < len(b); i++ {
		numOperands := OpcodeOperands[Opcode(b[i])]
		operands, read := ReadOperands(numOperands, b[i+1:])

		if !fn(i, b[i], operands) {
			break
		}

		i += read
	}
}
