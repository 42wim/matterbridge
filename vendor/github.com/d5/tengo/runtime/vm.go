package runtime

import (
	"fmt"
	"sync/atomic"

	"github.com/d5/tengo/compiler"
	"github.com/d5/tengo/compiler/source"
	"github.com/d5/tengo/compiler/token"
	"github.com/d5/tengo/objects"
)

const (
	// StackSize is the maximum stack size.
	StackSize = 2048

	// GlobalsSize is the maximum number of global variables.
	GlobalsSize = 1024

	// MaxFrames is the maximum number of function frames.
	MaxFrames = 1024
)

var (
	truePtr      = &objects.TrueValue
	falsePtr     = &objects.FalseValue
	undefinedPtr = &objects.UndefinedValue
)

// VM is a virtual machine that executes the bytecode compiled by Compiler.
type VM struct {
	constants      []objects.Object
	stack          []*objects.Object
	sp             int
	globals        []*objects.Object
	fileSet        *source.FileSet
	frames         []Frame
	framesIndex    int
	curFrame       *Frame
	curInsts       []byte
	curIPLimit     int
	ip             int
	aborting       int64
	builtinFuncs   []objects.Object
	builtinModules map[string]*objects.Object
	err            error
	errOffset      int
}

// NewVM creates a VM.
func NewVM(bytecode *compiler.Bytecode, globals []*objects.Object, builtinFuncs []objects.Object, builtinModules map[string]*objects.Object) *VM {
	if globals == nil {
		globals = make([]*objects.Object, GlobalsSize)
	}

	if builtinModules == nil {
		builtinModules = make(map[string]*objects.Object)
	}

	if builtinFuncs == nil {
		builtinFuncs = make([]objects.Object, len(objects.Builtins))
		for idx, fn := range objects.Builtins {
			builtinFuncs[idx] = &objects.BuiltinFunction{
				Name:  fn.Name,
				Value: fn.Value,
			}
		}
	}

	frames := make([]Frame, MaxFrames)
	frames[0].fn = bytecode.MainFunction
	frames[0].freeVars = nil
	frames[0].ip = -1
	frames[0].basePointer = 0

	return &VM{
		constants:      bytecode.Constants,
		stack:          make([]*objects.Object, StackSize),
		sp:             0,
		globals:        globals,
		fileSet:        bytecode.FileSet,
		frames:         frames,
		framesIndex:    1,
		curFrame:       &(frames[0]),
		curInsts:       frames[0].fn.Instructions,
		curIPLimit:     len(frames[0].fn.Instructions) - 1,
		ip:             -1,
		builtinFuncs:   builtinFuncs,
		builtinModules: builtinModules,
	}
}

// Abort aborts the execution.
func (v *VM) Abort() {
	atomic.StoreInt64(&v.aborting, 1)
}

// Run starts the execution.
func (v *VM) Run() (err error) {
	// reset VM states
	v.sp = 0
	v.curFrame = &(v.frames[0])
	v.curInsts = v.curFrame.fn.Instructions
	v.curIPLimit = len(v.curInsts) - 1
	v.framesIndex = 1
	v.ip = -1
	atomic.StoreInt64(&v.aborting, 0)

	v.run()

	err = v.err
	if err != nil {
		filePos := v.fileSet.Position(v.curFrame.fn.SourceMap[v.ip-v.errOffset])
		err = fmt.Errorf("Runtime Error: %s\n\tat %s", err.Error(), filePos)
		for v.framesIndex > 1 {
			v.framesIndex--
			v.curFrame = &v.frames[v.framesIndex-1]

			filePos = v.fileSet.Position(v.curFrame.fn.SourceMap[v.curFrame.ip-1])
			err = fmt.Errorf("%s\n\tat %s", err.Error(), filePos)
		}
		return err
	}

	// check if stack still has some objects left
	if v.sp > 0 && atomic.LoadInt64(&v.aborting) == 0 {
		panic(fmt.Errorf("non empty stack after execution: %d", v.sp))
	}

	return nil
}

func (v *VM) run() {
mainloop:
	for v.ip < v.curIPLimit && (atomic.LoadInt64(&v.aborting) == 0) {
		v.ip++

		switch v.curInsts[v.ip] {
		case compiler.OpConstant:
			cidx := int(v.curInsts[v.ip+2]) | int(v.curInsts[v.ip+1])<<8
			v.ip += 2

			if v.sp >= StackSize {
				v.err = ErrStackOverflow
				return
			}

			v.stack[v.sp] = &v.constants[cidx]
			v.sp++

		case compiler.OpNull:
			if v.sp >= StackSize {
				v.err = ErrStackOverflow
				return
			}

			v.stack[v.sp] = undefinedPtr
			v.sp++

		case compiler.OpAdd:
			v.binaryOp(token.Add)

		case compiler.OpSub:
			v.binaryOp(token.Sub)

		case compiler.OpMul:
			v.binaryOp(token.Mul)

		case compiler.OpDiv:
			v.binaryOp(token.Quo)

		case compiler.OpRem:
			v.binaryOp(token.Rem)

		case compiler.OpBAnd:
			v.binaryOp(token.And)

		case compiler.OpBOr:
			v.binaryOp(token.Or)

		case compiler.OpBXor:
			v.binaryOp(token.Xor)

		case compiler.OpBAndNot:
			v.binaryOp(token.AndNot)

		case compiler.OpBShiftLeft:
			v.binaryOp(token.Shl)

		case compiler.OpBShiftRight:
			v.binaryOp(token.Shr)

		case compiler.OpEqual:
			right := v.stack[v.sp-1]
			left := v.stack[v.sp-2]
			v.sp -= 2

			if v.sp >= StackSize {
				v.err = ErrStackOverflow
				return
			}

			if (*left).Equals(*right) {
				v.stack[v.sp] = truePtr
			} else {
				v.stack[v.sp] = falsePtr
			}
			v.sp++

		case compiler.OpNotEqual:
			right := v.stack[v.sp-1]
			left := v.stack[v.sp-2]
			v.sp -= 2

			if v.sp >= StackSize {
				v.err = ErrStackOverflow
				return
			}

			if (*left).Equals(*right) {
				v.stack[v.sp] = falsePtr
			} else {
				v.stack[v.sp] = truePtr
			}
			v.sp++

		case compiler.OpGreaterThan:
			v.binaryOp(token.Greater)

		case compiler.OpGreaterThanEqual:
			v.binaryOp(token.GreaterEq)

		case compiler.OpPop:
			v.sp--

		case compiler.OpTrue:
			if v.sp >= StackSize {
				v.err = ErrStackOverflow
				return
			}

			v.stack[v.sp] = truePtr
			v.sp++

		case compiler.OpFalse:
			if v.sp >= StackSize {
				v.err = ErrStackOverflow
				return
			}

			v.stack[v.sp] = falsePtr
			v.sp++

		case compiler.OpLNot:
			operand := v.stack[v.sp-1]
			v.sp--

			if v.sp >= StackSize {
				v.err = ErrStackOverflow
				return
			}

			if (*operand).IsFalsy() {
				v.stack[v.sp] = truePtr
			} else {
				v.stack[v.sp] = falsePtr
			}
			v.sp++

		case compiler.OpBComplement:
			operand := v.stack[v.sp-1]
			v.sp--

			switch x := (*operand).(type) {
			case *objects.Int:
				if v.sp >= StackSize {
					v.err = ErrStackOverflow
					return
				}

				var res objects.Object = &objects.Int{Value: ^x.Value}

				v.stack[v.sp] = &res
				v.sp++
			default:
				v.err = fmt.Errorf("invalid operation: ^%s", (*operand).TypeName())
				return
			}

		case compiler.OpMinus:
			operand := v.stack[v.sp-1]
			v.sp--

			switch x := (*operand).(type) {
			case *objects.Int:
				if v.sp >= StackSize {
					v.err = ErrStackOverflow
					return
				}

				var res objects.Object = &objects.Int{Value: -x.Value}

				v.stack[v.sp] = &res
				v.sp++
			case *objects.Float:
				if v.sp >= StackSize {
					v.err = ErrStackOverflow
					return
				}

				var res objects.Object = &objects.Float{Value: -x.Value}

				v.stack[v.sp] = &res
				v.sp++
			default:
				v.err = fmt.Errorf("invalid operation: -%s", (*operand).TypeName())
				return
			}

		case compiler.OpJumpFalsy:
			pos := int(v.curInsts[v.ip+2]) | int(v.curInsts[v.ip+1])<<8
			v.ip += 2

			condition := v.stack[v.sp-1]
			v.sp--

			if (*condition).IsFalsy() {
				v.ip = pos - 1
			}

		case compiler.OpAndJump:
			pos := int(v.curInsts[v.ip+2]) | int(v.curInsts[v.ip+1])<<8
			v.ip += 2

			condition := *v.stack[v.sp-1]
			if condition.IsFalsy() {
				v.ip = pos - 1
			} else {
				v.sp--
			}

		case compiler.OpOrJump:
			pos := int(v.curInsts[v.ip+2]) | int(v.curInsts[v.ip+1])<<8
			v.ip += 2

			condition := *v.stack[v.sp-1]
			if !condition.IsFalsy() {
				v.ip = pos - 1
			} else {
				v.sp--
			}

		case compiler.OpJump:
			pos := int(v.curInsts[v.ip+2]) | int(v.curInsts[v.ip+1])<<8
			v.ip = pos - 1

		case compiler.OpSetGlobal:
			globalIndex := int(v.curInsts[v.ip+2]) | int(v.curInsts[v.ip+1])<<8
			v.ip += 2

			v.sp--

			v.globals[globalIndex] = v.stack[v.sp]

		case compiler.OpSetSelGlobal:
			globalIndex := int(v.curInsts[v.ip+2]) | int(v.curInsts[v.ip+1])<<8
			numSelectors := int(v.curInsts[v.ip+3])
			v.ip += 3

			// selectors and RHS value
			selectors := v.stack[v.sp-numSelectors : v.sp]
			val := v.stack[v.sp-numSelectors-1]
			v.sp -= numSelectors + 1

			if e := indexAssign(v.globals[globalIndex], val, selectors); e != nil {
				v.errOffset = 3
				v.err = e
				return
			}

		case compiler.OpGetGlobal:
			globalIndex := int(v.curInsts[v.ip+2]) | int(v.curInsts[v.ip+1])<<8
			v.ip += 2

			val := v.globals[globalIndex]

			if v.sp >= StackSize {
				v.err = ErrStackOverflow
				return
			}

			v.stack[v.sp] = val
			v.sp++

		case compiler.OpArray:
			numElements := int(v.curInsts[v.ip+2]) | int(v.curInsts[v.ip+1])<<8
			v.ip += 2

			var elements []objects.Object
			for i := v.sp - numElements; i < v.sp; i++ {
				elements = append(elements, *v.stack[i])
			}
			v.sp -= numElements

			var arr objects.Object = &objects.Array{Value: elements}

			if v.sp >= StackSize {
				v.err = ErrStackOverflow
				return
			}

			v.stack[v.sp] = &arr
			v.sp++

		case compiler.OpMap:
			numElements := int(v.curInsts[v.ip+2]) | int(v.curInsts[v.ip+1])<<8
			v.ip += 2

			kv := make(map[string]objects.Object)
			for i := v.sp - numElements; i < v.sp; i += 2 {
				key := *v.stack[i]
				value := *v.stack[i+1]
				kv[key.(*objects.String).Value] = value
			}
			v.sp -= numElements

			var m objects.Object = &objects.Map{Value: kv}

			if v.sp >= StackSize {
				v.err = ErrStackOverflow
				return
			}

			v.stack[v.sp] = &m
			v.sp++

		case compiler.OpError:
			value := v.stack[v.sp-1]

			var e objects.Object = &objects.Error{
				Value: *value,
			}

			v.stack[v.sp-1] = &e

		case compiler.OpImmutable:
			value := v.stack[v.sp-1]

			switch value := (*value).(type) {
			case *objects.Array:
				var immutableArray objects.Object = &objects.ImmutableArray{
					Value: value.Value,
				}
				v.stack[v.sp-1] = &immutableArray
			case *objects.Map:
				var immutableMap objects.Object = &objects.ImmutableMap{
					Value: value.Value,
				}
				v.stack[v.sp-1] = &immutableMap
			}

		case compiler.OpIndex:
			index := v.stack[v.sp-1]
			left := v.stack[v.sp-2]
			v.sp -= 2

			switch left := (*left).(type) {
			case objects.Indexable:
				val, e := left.IndexGet(*index)
				if e != nil {

					if e == objects.ErrInvalidIndexType {
						v.err = fmt.Errorf("invalid index type: %s", (*index).TypeName())
						return
					}

					v.err = e
					return
				}
				if val == nil {
					val = objects.UndefinedValue
				}

				if v.sp >= StackSize {
					v.err = ErrStackOverflow
					return
				}

				v.stack[v.sp] = &val
				v.sp++

			case *objects.Error: // e.value
				key, ok := (*index).(*objects.String)
				if !ok || key.Value != "value" {
					v.err = fmt.Errorf("invalid index on error")
					return
				}

				if v.sp >= StackSize {
					v.err = ErrStackOverflow
					return
				}

				v.stack[v.sp] = &left.Value
				v.sp++

			default:
				v.err = fmt.Errorf("not indexable: %s", left.TypeName())
				return
			}

		case compiler.OpSliceIndex:
			high := v.stack[v.sp-1]
			low := v.stack[v.sp-2]
			left := v.stack[v.sp-3]
			v.sp -= 3

			var lowIdx int64
			if *low != objects.UndefinedValue {
				if low, ok := (*low).(*objects.Int); ok {
					lowIdx = low.Value
				} else {
					v.err = fmt.Errorf("invalid slice index type: %s", low.TypeName())
					return
				}
			}

			switch left := (*left).(type) {
			case *objects.Array:
				numElements := int64(len(left.Value))
				var highIdx int64
				if *high == objects.UndefinedValue {
					highIdx = numElements
				} else if high, ok := (*high).(*objects.Int); ok {
					highIdx = high.Value
				} else {
					v.err = fmt.Errorf("invalid slice index type: %s", high.TypeName())
					return
				}

				if lowIdx > highIdx {
					v.err = fmt.Errorf("invalid slice index: %d > %d", lowIdx, highIdx)
					return
				}

				if lowIdx < 0 {
					lowIdx = 0
				} else if lowIdx > numElements {
					lowIdx = numElements
				}

				if highIdx < 0 {
					highIdx = 0
				} else if highIdx > numElements {
					highIdx = numElements
				}

				if v.sp >= StackSize {
					v.err = ErrStackOverflow
					return
				}

				var val objects.Object = &objects.Array{Value: left.Value[lowIdx:highIdx]}
				v.stack[v.sp] = &val
				v.sp++

			case *objects.ImmutableArray:
				numElements := int64(len(left.Value))
				var highIdx int64
				if *high == objects.UndefinedValue {
					highIdx = numElements
				} else if high, ok := (*high).(*objects.Int); ok {
					highIdx = high.Value
				} else {
					v.err = fmt.Errorf("invalid slice index type: %s", high.TypeName())
					return
				}

				if lowIdx > highIdx {
					v.err = fmt.Errorf("invalid slice index: %d > %d", lowIdx, highIdx)
					return
				}

				if lowIdx < 0 {
					lowIdx = 0
				} else if lowIdx > numElements {
					lowIdx = numElements
				}

				if highIdx < 0 {
					highIdx = 0
				} else if highIdx > numElements {
					highIdx = numElements
				}

				if v.sp >= StackSize {
					v.err = ErrStackOverflow
					return
				}

				var val objects.Object = &objects.Array{Value: left.Value[lowIdx:highIdx]}

				v.stack[v.sp] = &val
				v.sp++

			case *objects.String:
				numElements := int64(len(left.Value))
				var highIdx int64
				if *high == objects.UndefinedValue {
					highIdx = numElements
				} else if high, ok := (*high).(*objects.Int); ok {
					highIdx = high.Value
				} else {
					v.err = fmt.Errorf("invalid slice index type: %s", high.TypeName())
					return
				}

				if lowIdx > highIdx {
					v.err = fmt.Errorf("invalid slice index: %d > %d", lowIdx, highIdx)
					return
				}

				if lowIdx < 0 {
					lowIdx = 0
				} else if lowIdx > numElements {
					lowIdx = numElements
				}

				if highIdx < 0 {
					highIdx = 0
				} else if highIdx > numElements {
					highIdx = numElements
				}

				if v.sp >= StackSize {
					v.err = ErrStackOverflow
					return
				}

				var val objects.Object = &objects.String{Value: left.Value[lowIdx:highIdx]}

				v.stack[v.sp] = &val
				v.sp++

			case *objects.Bytes:
				numElements := int64(len(left.Value))
				var highIdx int64
				if *high == objects.UndefinedValue {
					highIdx = numElements
				} else if high, ok := (*high).(*objects.Int); ok {
					highIdx = high.Value
				} else {
					v.err = fmt.Errorf("invalid slice index type: %s", high.TypeName())
					return
				}

				if lowIdx > highIdx {
					v.err = fmt.Errorf("invalid slice index: %d > %d", lowIdx, highIdx)
					return
				}

				if lowIdx < 0 {
					lowIdx = 0
				} else if lowIdx > numElements {
					lowIdx = numElements
				}

				if highIdx < 0 {
					highIdx = 0
				} else if highIdx > numElements {
					highIdx = numElements
				}

				if v.sp >= StackSize {
					v.err = ErrStackOverflow
					return
				}

				var val objects.Object = &objects.Bytes{Value: left.Value[lowIdx:highIdx]}

				v.stack[v.sp] = &val
				v.sp++
			}

		case compiler.OpCall:
			numArgs := int(v.curInsts[v.ip+1])
			v.ip++

			value := *v.stack[v.sp-1-numArgs]

			switch callee := value.(type) {
			case *objects.Closure:
				if numArgs != callee.Fn.NumParameters {
					v.errOffset = 1
					v.err = fmt.Errorf("wrong number of arguments: want=%d, got=%d",
						callee.Fn.NumParameters, numArgs)
					return
				}

				// test if it's tail-call
				if callee.Fn == v.curFrame.fn { // recursion
					nextOp := v.curInsts[v.ip+1]
					if nextOp == compiler.OpReturnValue ||
						(nextOp == compiler.OpPop && compiler.OpReturn == v.curInsts[v.ip+2]) {
						for p := 0; p < numArgs; p++ {
							v.stack[v.curFrame.basePointer+p] = v.stack[v.sp-numArgs+p]
						}
						v.sp -= numArgs + 1
						v.ip = -1 // reset IP to beginning of the frame
						continue mainloop
					}
				}

				// update call frame
				v.curFrame.ip = v.ip // store current ip before call
				v.curFrame = &(v.frames[v.framesIndex])
				v.curFrame.fn = callee.Fn
				v.curFrame.freeVars = callee.Free
				v.curFrame.basePointer = v.sp - numArgs
				v.curInsts = callee.Fn.Instructions
				v.ip = -1
				v.curIPLimit = len(v.curInsts) - 1
				v.framesIndex++
				v.sp = v.sp - numArgs + callee.Fn.NumLocals

			case *objects.CompiledFunction:
				if numArgs != callee.NumParameters {
					v.errOffset = 1
					v.err = fmt.Errorf("wrong number of arguments: want=%d, got=%d",
						callee.NumParameters, numArgs)
					return
				}

				// test if it's tail-call
				if callee == v.curFrame.fn { // recursion
					nextOp := v.curInsts[v.ip+1]
					if nextOp == compiler.OpReturnValue ||
						(nextOp == compiler.OpPop && compiler.OpReturn == v.curInsts[v.ip+2]) {
						for p := 0; p < numArgs; p++ {
							v.stack[v.curFrame.basePointer+p] = v.stack[v.sp-numArgs+p]
						}
						v.sp -= numArgs + 1
						v.ip = -1 // reset IP to beginning of the frame
						continue mainloop
					}
				}

				// update call frame
				v.curFrame.ip = v.ip // store current ip before call
				v.curFrame = &(v.frames[v.framesIndex])
				v.curFrame.fn = callee
				v.curFrame.freeVars = nil
				v.curFrame.basePointer = v.sp - numArgs
				v.curInsts = callee.Instructions
				v.ip = -1
				v.curIPLimit = len(v.curInsts) - 1
				v.framesIndex++
				v.sp = v.sp - numArgs + callee.NumLocals

			case objects.Callable:
				var args []objects.Object
				for _, arg := range v.stack[v.sp-numArgs : v.sp] {
					args = append(args, *arg)
				}

				ret, e := callee.Call(args...)
				v.sp -= numArgs + 1

				// runtime error
				if e != nil {
					v.errOffset = 1

					if e == objects.ErrWrongNumArguments {
						v.err = fmt.Errorf("wrong number of arguments in call to '%s'",
							value.TypeName())
						return
					}

					if e, ok := e.(objects.ErrInvalidArgumentType); ok {
						v.err = fmt.Errorf("invalid type for argument '%s' in call to '%s': expected %s, found %s",
							e.Name, value.TypeName(), e.Expected, e.Found)
						return
					}

					v.err = e
					return
				}

				// nil return -> undefined
				if ret == nil {
					ret = objects.UndefinedValue
				}

				if v.sp >= StackSize {
					v.err = ErrStackOverflow
					return
				}

				v.stack[v.sp] = &ret
				v.sp++

			default:
				v.errOffset = 1
				v.err = fmt.Errorf("not callable: %s", callee.TypeName())
				return
			}

		case compiler.OpReturnValue:
			retVal := v.stack[v.sp-1]
			//v.sp--

			v.framesIndex--
			lastFrame := v.frames[v.framesIndex]
			v.curFrame = &v.frames[v.framesIndex-1]
			v.curInsts = v.curFrame.fn.Instructions
			v.curIPLimit = len(v.curInsts) - 1
			v.ip = v.curFrame.ip

			//v.sp = lastFrame.basePointer - 1
			v.sp = lastFrame.basePointer

			// skip stack overflow check because (newSP) <= (oldSP)
			v.stack[v.sp-1] = retVal
			//v.sp++

		case compiler.OpReturn:
			v.framesIndex--
			lastFrame := v.frames[v.framesIndex]
			v.curFrame = &v.frames[v.framesIndex-1]
			v.curInsts = v.curFrame.fn.Instructions
			v.curIPLimit = len(v.curInsts) - 1
			v.ip = v.curFrame.ip

			//v.sp = lastFrame.basePointer - 1
			v.sp = lastFrame.basePointer

			// skip stack overflow check because (newSP) <= (oldSP)
			v.stack[v.sp-1] = undefinedPtr
			//v.sp++

		case compiler.OpDefineLocal:
			localIndex := int(v.curInsts[v.ip+1])
			v.ip++

			sp := v.curFrame.basePointer + localIndex

			// local variables can be mutated by other actions
			// so always store the copy of popped value
			val := *v.stack[v.sp-1]
			v.sp--

			v.stack[sp] = &val

		case compiler.OpSetLocal:
			localIndex := int(v.curInsts[v.ip+1])
			v.ip++

			sp := v.curFrame.basePointer + localIndex

			// update pointee of v.stack[sp] instead of replacing the pointer itself.
			// this is needed because there can be free variables referencing the same local variables.
			val := v.stack[v.sp-1]
			v.sp--

			*v.stack[sp] = *val // also use a copy of popped value

		case compiler.OpSetSelLocal:
			localIndex := int(v.curInsts[v.ip+1])
			numSelectors := int(v.curInsts[v.ip+2])
			v.ip += 2

			// selectors and RHS value
			selectors := v.stack[v.sp-numSelectors : v.sp]
			val := v.stack[v.sp-numSelectors-1]
			v.sp -= numSelectors + 1

			sp := v.curFrame.basePointer + localIndex

			if e := indexAssign(v.stack[sp], val, selectors); e != nil {
				v.errOffset = 2
				v.err = e
				return
			}

		case compiler.OpGetLocal:
			localIndex := int(v.curInsts[v.ip+1])
			v.ip++

			val := v.stack[v.curFrame.basePointer+localIndex]

			if v.sp >= StackSize {
				v.err = ErrStackOverflow
				return
			}

			v.stack[v.sp] = val
			v.sp++

		case compiler.OpGetBuiltin:
			builtinIndex := int(v.curInsts[v.ip+1])
			v.ip++

			if v.sp >= StackSize {
				v.err = ErrStackOverflow
				return
			}

			v.stack[v.sp] = &v.builtinFuncs[builtinIndex]
			v.sp++

		case compiler.OpGetBuiltinModule:
			val := v.stack[v.sp-1]
			v.sp--

			moduleName := (*val).(*objects.String).Value

			module, ok := v.builtinModules[moduleName]
			if !ok {
				v.errOffset = 3
				v.err = fmt.Errorf("module '%s' not found", moduleName)
				return
			}

			if v.sp >= StackSize {
				v.err = ErrStackOverflow
				return
			}

			v.stack[v.sp] = module
			v.sp++

		case compiler.OpClosure:
			constIndex := int(v.curInsts[v.ip+2]) | int(v.curInsts[v.ip+1])<<8
			numFree := int(v.curInsts[v.ip+3])
			v.ip += 3

			fn, ok := v.constants[constIndex].(*objects.CompiledFunction)
			if !ok {
				v.errOffset = 3
				v.err = fmt.Errorf("not function: %s", fn.TypeName())
				return
			}

			free := make([]*objects.Object, numFree)
			for i := 0; i < numFree; i++ {
				free[i] = v.stack[v.sp-numFree+i]
			}
			v.sp -= numFree

			if v.sp >= StackSize {
				v.err = ErrStackOverflow
				return
			}

			var cl objects.Object = &objects.Closure{
				Fn:   fn,
				Free: free,
			}

			v.stack[v.sp] = &cl
			v.sp++

		case compiler.OpGetFree:
			freeIndex := int(v.curInsts[v.ip+1])
			v.ip++

			val := v.curFrame.freeVars[freeIndex]

			if v.sp >= StackSize {
				v.err = ErrStackOverflow
				return
			}

			v.stack[v.sp] = val
			v.sp++

		case compiler.OpSetSelFree:
			freeIndex := int(v.curInsts[v.ip+1])
			numSelectors := int(v.curInsts[v.ip+2])
			v.ip += 2

			// selectors and RHS value
			selectors := v.stack[v.sp-numSelectors : v.sp]
			val := v.stack[v.sp-numSelectors-1]
			v.sp -= numSelectors + 1

			if e := indexAssign(v.curFrame.freeVars[freeIndex], val, selectors); e != nil {
				v.errOffset = 2
				v.err = e
				return
			}

		case compiler.OpSetFree:
			freeIndex := int(v.curInsts[v.ip+1])
			v.ip++

			val := v.stack[v.sp-1]
			v.sp--

			*v.curFrame.freeVars[freeIndex] = *val

		case compiler.OpIteratorInit:
			var iterator objects.Object

			dst := v.stack[v.sp-1]
			v.sp--

			iterable, ok := (*dst).(objects.Iterable)
			if !ok {
				v.err = fmt.Errorf("not iterable: %s", (*dst).TypeName())
				return
			}

			iterator = iterable.Iterate()

			if v.sp >= StackSize {
				v.err = ErrStackOverflow
				return
			}

			v.stack[v.sp] = &iterator
			v.sp++

		case compiler.OpIteratorNext:
			iterator := v.stack[v.sp-1]
			v.sp--

			hasMore := (*iterator).(objects.Iterator).Next()

			if v.sp >= StackSize {
				v.err = ErrStackOverflow
				return
			}

			if hasMore {
				v.stack[v.sp] = truePtr
			} else {
				v.stack[v.sp] = falsePtr
			}
			v.sp++

		case compiler.OpIteratorKey:
			iterator := v.stack[v.sp-1]
			v.sp--

			val := (*iterator).(objects.Iterator).Key()

			if v.sp >= StackSize {
				v.err = ErrStackOverflow
				return
			}

			v.stack[v.sp] = &val
			v.sp++

		case compiler.OpIteratorValue:
			iterator := v.stack[v.sp-1]
			v.sp--

			val := (*iterator).(objects.Iterator).Value()

			if v.sp >= StackSize {
				v.err = ErrStackOverflow
				return
			}

			v.stack[v.sp] = &val
			v.sp++

		default:
			panic(fmt.Errorf("unknown opcode: %d", v.curInsts[v.ip]))
		}
	}
}

// Globals returns the global variables.
func (v *VM) Globals() []*objects.Object {
	return v.globals
}

func indexAssign(dst, src *objects.Object, selectors []*objects.Object) error {
	numSel := len(selectors)

	for sidx := numSel - 1; sidx > 0; sidx-- {
		indexable, ok := (*dst).(objects.Indexable)
		if !ok {
			return fmt.Errorf("not indexable: %s", (*dst).TypeName())
		}

		next, err := indexable.IndexGet(*selectors[sidx])
		if err != nil {
			if err == objects.ErrInvalidIndexType {
				return fmt.Errorf("invalid index type: %s", (*selectors[sidx]).TypeName())
			}

			return err
		}

		dst = &next
	}

	indexAssignable, ok := (*dst).(objects.IndexAssignable)
	if !ok {
		return fmt.Errorf("not index-assignable: %s", (*dst).TypeName())
	}

	if err := indexAssignable.IndexSet(*selectors[0], *src); err != nil {
		if err == objects.ErrInvalidIndexValueType {
			return fmt.Errorf("invaid index value type: %s", (*src).TypeName())
		}

		return err
	}

	return nil
}

func (v *VM) binaryOp(tok token.Token) {
	right := v.stack[v.sp-1]
	left := v.stack[v.sp-2]

	res, e := (*left).BinaryOp(tok, *right)
	if e != nil {
		v.sp -= 2
		atomic.StoreInt64(&v.aborting, 1)

		if e == objects.ErrInvalidOperator {
			v.err = fmt.Errorf("invalid operation: %s + %s",
				(*left).TypeName(), (*right).TypeName())
			return
		}

		v.err = e
		return
	}

	v.stack[v.sp-2] = &res
	v.sp--
}
