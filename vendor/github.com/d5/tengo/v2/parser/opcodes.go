package parser

// Opcode represents a single byte operation code.
type Opcode = byte

// List of opcodes
const (
	OpConstant      Opcode = iota // Load constant
	OpBComplement                 // bitwise complement
	OpPop                         // Pop
	OpTrue                        // Push true
	OpFalse                       // Push false
	OpEqual                       // Equal ==
	OpNotEqual                    // Not equal !=
	OpMinus                       // Minus -
	OpLNot                        // Logical not !
	OpJumpFalsy                   // Jump if falsy
	OpAndJump                     // Logical AND jump
	OpOrJump                      // Logical OR jump
	OpJump                        // Jump
	OpNull                        // Push null
	OpArray                       // Array object
	OpMap                         // Map object
	OpError                       // Error object
	OpImmutable                   // Immutable object
	OpIndex                       // Index operation
	OpSliceIndex                  // Slice operation
	OpCall                        // Call function
	OpReturn                      // Return
	OpGetGlobal                   // Get global variable
	OpSetGlobal                   // Set global variable
	OpSetSelGlobal                // Set global variable using selectors
	OpGetLocal                    // Get local variable
	OpSetLocal                    // Set local variable
	OpDefineLocal                 // Define local variable
	OpSetSelLocal                 // Set local variable using selectors
	OpGetFreePtr                  // Get free variable pointer object
	OpGetFree                     // Get free variables
	OpSetFree                     // Set free variables
	OpGetLocalPtr                 // Get local variable as a pointer
	OpSetSelFree                  // Set free variables using selectors
	OpGetBuiltin                  // Get builtin function
	OpClosure                     // Push closure
	OpIteratorInit                // Iterator init
	OpIteratorNext                // Iterator next
	OpIteratorKey                 // Iterator key
	OpIteratorValue               // Iterator value
	OpBinaryOp                    // Binary operation
	OpSuspend                     // Suspend VM
)

// OpcodeNames are string representation of opcodes.
var OpcodeNames = [...]string{
	OpConstant:      "CONST",
	OpPop:           "POP",
	OpTrue:          "TRUE",
	OpFalse:         "FALSE",
	OpBComplement:   "NEG",
	OpEqual:         "EQL",
	OpNotEqual:      "NEQ",
	OpMinus:         "NEG",
	OpLNot:          "NOT",
	OpJumpFalsy:     "JMPF",
	OpAndJump:       "ANDJMP",
	OpOrJump:        "ORJMP",
	OpJump:          "JMP",
	OpNull:          "NULL",
	OpGetGlobal:     "GETG",
	OpSetGlobal:     "SETG",
	OpSetSelGlobal:  "SETSG",
	OpArray:         "ARR",
	OpMap:           "MAP",
	OpError:         "ERROR",
	OpImmutable:     "IMMUT",
	OpIndex:         "INDEX",
	OpSliceIndex:    "SLICE",
	OpCall:          "CALL",
	OpReturn:        "RET",
	OpGetLocal:      "GETL",
	OpSetLocal:      "SETL",
	OpDefineLocal:   "DEFL",
	OpSetSelLocal:   "SETSL",
	OpGetBuiltin:    "BUILTIN",
	OpClosure:       "CLOSURE",
	OpGetFreePtr:    "GETFP",
	OpGetFree:       "GETF",
	OpSetFree:       "SETF",
	OpGetLocalPtr:   "GETLP",
	OpSetSelFree:    "SETSF",
	OpIteratorInit:  "ITER",
	OpIteratorNext:  "ITNXT",
	OpIteratorKey:   "ITKEY",
	OpIteratorValue: "ITVAL",
	OpBinaryOp:      "BINARYOP",
	OpSuspend:       "SUSPEND",
}

// OpcodeOperands is the number of operands.
var OpcodeOperands = [...][]int{
	OpConstant:      {2},
	OpPop:           {},
	OpTrue:          {},
	OpFalse:         {},
	OpBComplement:   {},
	OpEqual:         {},
	OpNotEqual:      {},
	OpMinus:         {},
	OpLNot:          {},
	OpJumpFalsy:     {2},
	OpAndJump:       {2},
	OpOrJump:        {2},
	OpJump:          {2},
	OpNull:          {},
	OpGetGlobal:     {2},
	OpSetGlobal:     {2},
	OpSetSelGlobal:  {2, 1},
	OpArray:         {2},
	OpMap:           {2},
	OpError:         {},
	OpImmutable:     {},
	OpIndex:         {},
	OpSliceIndex:    {},
	OpCall:          {1, 1},
	OpReturn:        {1},
	OpGetLocal:      {1},
	OpSetLocal:      {1},
	OpDefineLocal:   {1},
	OpSetSelLocal:   {1, 1},
	OpGetBuiltin:    {1},
	OpClosure:       {2, 1},
	OpGetFreePtr:    {1},
	OpGetFree:       {1},
	OpSetFree:       {1},
	OpGetLocalPtr:   {1},
	OpSetSelFree:    {1, 1},
	OpIteratorInit:  {},
	OpIteratorNext:  {},
	OpIteratorKey:   {},
	OpIteratorValue: {},
	OpBinaryOp:      {1},
	OpSuspend:       {},
}

// ReadOperands reads operands from the bytecode.
func ReadOperands(numOperands []int, ins []byte) (operands []int, offset int) {
	for _, width := range numOperands {
		switch width {
		case 1:
			operands = append(operands, int(ins[offset]))
		case 2:
			operands = append(operands, int(ins[offset+1])|int(ins[offset])<<8)
		}
		offset += width
	}
	return
}
