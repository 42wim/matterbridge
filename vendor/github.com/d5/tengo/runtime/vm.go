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

// VM is a virtual machine that executes the bytecode compiled by Compiler.
type VM struct {
	constants   []objects.Object
	stack       [StackSize]objects.Object
	sp          int
	globals     []objects.Object
	fileSet     *source.FileSet
	frames      [MaxFrames]Frame
	framesIndex int
	curFrame    *Frame
	curInsts    []byte
	ip          int
	aborting    int64
	maxAllocs   int64
	allocs      int64
	err         error
}

// NewVM creates a VM.
func NewVM(bytecode *compiler.Bytecode, globals []objects.Object, maxAllocs int64) *VM {
	if globals == nil {
		globals = make([]objects.Object, GlobalsSize)
	}

	v := &VM{
		constants:   bytecode.Constants,
		sp:          0,
		globals:     globals,
		fileSet:     bytecode.FileSet,
		framesIndex: 1,
		ip:          -1,
		maxAllocs:   maxAllocs,
	}

	v.frames[0].fn = bytecode.MainFunction
	v.frames[0].ip = -1
	v.curFrame = &v.frames[0]
	v.curInsts = v.curFrame.fn.Instructions

	return v
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
	v.framesIndex = 1
	v.ip = -1
	v.allocs = v.maxAllocs + 1

	v.run()

	atomic.StoreInt64(&v.aborting, 0)

	err = v.err
	if err != nil {
		filePos := v.fileSet.Position(v.curFrame.fn.SourcePos(v.ip - 1))
		err = fmt.Errorf("Runtime Error: %s\n\tat %s", err.Error(), filePos)
		for v.framesIndex > 1 {
			v.framesIndex--
			v.curFrame = &v.frames[v.framesIndex-1]

			filePos = v.fileSet.Position(v.curFrame.fn.SourcePos(v.curFrame.ip - 1))
			err = fmt.Errorf("%s\n\tat %s", err.Error(), filePos)
		}
		return err
	}

	return nil
}

func (v *VM) run() {
	defer func() {
		if r := recover(); r != nil {
			if v.sp >= StackSize || v.framesIndex >= MaxFrames {
				v.err = ErrStackOverflow
				return
			}

			if v.ip < len(v.curInsts)-1 {
				if err, ok := r.(error); ok {
					v.err = err
				} else {
					v.err = fmt.Errorf("panic: %v", r)
				}
			}
		}
	}()

	for atomic.LoadInt64(&v.aborting) == 0 {
		v.ip++

		switch v.curInsts[v.ip] {
		case compiler.OpConstant:
			v.ip += 2
			cidx := int(v.curInsts[v.ip]) | int(v.curInsts[v.ip-1])<<8

			v.stack[v.sp] = v.constants[cidx]
			v.sp++

		case compiler.OpNull:
			v.stack[v.sp] = objects.UndefinedValue
			v.sp++

		case compiler.OpBinaryOp:
			v.ip++
			right := v.stack[v.sp-1]
			left := v.stack[v.sp-2]

			tok := token.Token(v.curInsts[v.ip])
			res, e := left.BinaryOp(tok, right)
			if e != nil {
				v.sp -= 2

				if e == objects.ErrInvalidOperator {
					v.err = fmt.Errorf("invalid operation: %s %s %s",
						left.TypeName(), tok.String(), right.TypeName())
					return
				}

				v.err = e
				return
			}

			v.allocs--
			if v.allocs == 0 {
				v.err = ErrObjectAllocLimit
				return
			}

			v.stack[v.sp-2] = res
			v.sp--

		case compiler.OpEqual:
			right := v.stack[v.sp-1]
			left := v.stack[v.sp-2]
			v.sp -= 2

			if left.Equals(right) {
				v.stack[v.sp] = objects.TrueValue
			} else {
				v.stack[v.sp] = objects.FalseValue
			}
			v.sp++

		case compiler.OpNotEqual:
			right := v.stack[v.sp-1]
			left := v.stack[v.sp-2]
			v.sp -= 2

			if left.Equals(right) {
				v.stack[v.sp] = objects.FalseValue
			} else {
				v.stack[v.sp] = objects.TrueValue
			}
			v.sp++

		case compiler.OpPop:
			v.sp--

		case compiler.OpTrue:
			v.stack[v.sp] = objects.TrueValue
			v.sp++

		case compiler.OpFalse:
			v.stack[v.sp] = objects.FalseValue
			v.sp++

		case compiler.OpLNot:
			operand := v.stack[v.sp-1]
			v.sp--

			if operand.IsFalsy() {
				v.stack[v.sp] = objects.TrueValue
			} else {
				v.stack[v.sp] = objects.FalseValue
			}
			v.sp++

		case compiler.OpBComplement:
			operand := v.stack[v.sp-1]
			v.sp--

			switch x := operand.(type) {
			case *objects.Int:
				var res objects.Object = &objects.Int{Value: ^x.Value}

				v.allocs--
				if v.allocs == 0 {
					v.err = ErrObjectAllocLimit
					return
				}

				v.stack[v.sp] = res
				v.sp++
			default:
				v.err = fmt.Errorf("invalid operation: ^%s", operand.TypeName())
				return
			}

		case compiler.OpMinus:
			operand := v.stack[v.sp-1]
			v.sp--

			switch x := operand.(type) {
			case *objects.Int:
				var res objects.Object = &objects.Int{Value: -x.Value}

				v.allocs--
				if v.allocs == 0 {
					v.err = ErrObjectAllocLimit
					return
				}

				v.stack[v.sp] = res
				v.sp++
			case *objects.Float:
				var res objects.Object = &objects.Float{Value: -x.Value}

				v.allocs--
				if v.allocs == 0 {
					v.err = ErrObjectAllocLimit
					return
				}

				v.stack[v.sp] = res
				v.sp++
			default:
				v.err = fmt.Errorf("invalid operation: -%s", operand.TypeName())
				return
			}

		case compiler.OpJumpFalsy:
			v.ip += 2
			v.sp--
			if v.stack[v.sp].IsFalsy() {
				pos := int(v.curInsts[v.ip]) | int(v.curInsts[v.ip-1])<<8
				v.ip = pos - 1
			}

		case compiler.OpAndJump:
			v.ip += 2

			if v.stack[v.sp-1].IsFalsy() {
				pos := int(v.curInsts[v.ip]) | int(v.curInsts[v.ip-1])<<8
				v.ip = pos - 1
			} else {
				v.sp--
			}

		case compiler.OpOrJump:
			v.ip += 2

			if v.stack[v.sp-1].IsFalsy() {
				v.sp--
			} else {
				pos := int(v.curInsts[v.ip]) | int(v.curInsts[v.ip-1])<<8
				v.ip = pos - 1
			}

		case compiler.OpJump:
			pos := int(v.curInsts[v.ip+2]) | int(v.curInsts[v.ip+1])<<8
			v.ip = pos - 1

		case compiler.OpSetGlobal:
			v.ip += 2
			v.sp--

			globalIndex := int(v.curInsts[v.ip]) | int(v.curInsts[v.ip-1])<<8
			v.globals[globalIndex] = v.stack[v.sp]

		case compiler.OpSetSelGlobal:
			v.ip += 3
			globalIndex := int(v.curInsts[v.ip-1]) | int(v.curInsts[v.ip-2])<<8
			numSelectors := int(v.curInsts[v.ip])

			// selectors and RHS value
			selectors := make([]objects.Object, numSelectors)
			for i := 0; i < numSelectors; i++ {
				selectors[i] = v.stack[v.sp-numSelectors+i]
			}

			val := v.stack[v.sp-numSelectors-1]
			v.sp -= numSelectors + 1

			if e := indexAssign(v.globals[globalIndex], val, selectors); e != nil {
				v.err = e
				return
			}

		case compiler.OpGetGlobal:
			v.ip += 2
			globalIndex := int(v.curInsts[v.ip]) | int(v.curInsts[v.ip-1])<<8

			val := v.globals[globalIndex]

			v.stack[v.sp] = val
			v.sp++

		case compiler.OpArray:
			v.ip += 2
			numElements := int(v.curInsts[v.ip]) | int(v.curInsts[v.ip-1])<<8

			var elements []objects.Object
			for i := v.sp - numElements; i < v.sp; i++ {
				elements = append(elements, v.stack[i])
			}
			v.sp -= numElements

			var arr objects.Object = &objects.Array{Value: elements}

			v.allocs--
			if v.allocs == 0 {
				v.err = ErrObjectAllocLimit
				return
			}

			v.stack[v.sp] = arr
			v.sp++

		case compiler.OpMap:
			v.ip += 2
			numElements := int(v.curInsts[v.ip]) | int(v.curInsts[v.ip-1])<<8

			kv := make(map[string]objects.Object)
			for i := v.sp - numElements; i < v.sp; i += 2 {
				key := v.stack[i]
				value := v.stack[i+1]
				kv[key.(*objects.String).Value] = value
			}
			v.sp -= numElements

			var m objects.Object = &objects.Map{Value: kv}

			v.allocs--
			if v.allocs == 0 {
				v.err = ErrObjectAllocLimit
				return
			}

			v.stack[v.sp] = m
			v.sp++

		case compiler.OpError:
			value := v.stack[v.sp-1]

			var e objects.Object = &objects.Error{
				Value: value,
			}

			v.allocs--
			if v.allocs == 0 {
				v.err = ErrObjectAllocLimit
				return
			}

			v.stack[v.sp-1] = e

		case compiler.OpImmutable:
			value := v.stack[v.sp-1]

			switch value := value.(type) {
			case *objects.Array:
				var immutableArray objects.Object = &objects.ImmutableArray{
					Value: value.Value,
				}

				v.allocs--
				if v.allocs == 0 {
					v.err = ErrObjectAllocLimit
					return
				}

				v.stack[v.sp-1] = immutableArray
			case *objects.Map:
				var immutableMap objects.Object = &objects.ImmutableMap{
					Value: value.Value,
				}

				v.allocs--
				if v.allocs == 0 {
					v.err = ErrObjectAllocLimit
					return
				}

				v.stack[v.sp-1] = immutableMap
			}

		case compiler.OpIndex:
			index := v.stack[v.sp-1]
			left := v.stack[v.sp-2]
			v.sp -= 2

			switch left := left.(type) {
			case objects.Indexable:
				val, e := left.IndexGet(index)
				if e != nil {

					if e == objects.ErrInvalidIndexType {
						v.err = fmt.Errorf("invalid index type: %s", index.TypeName())
						return
					}

					v.err = e
					return
				}
				if val == nil {
					val = objects.UndefinedValue
				}

				v.stack[v.sp] = val
				v.sp++

			case *objects.Error: // e.value
				key, ok := index.(*objects.String)
				if !ok || key.Value != "value" {
					v.err = fmt.Errorf("invalid index on error")
					return
				}

				v.stack[v.sp] = left.Value
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
			if low != objects.UndefinedValue {
				if low, ok := low.(*objects.Int); ok {
					lowIdx = low.Value
				} else {
					v.err = fmt.Errorf("invalid slice index type: %s", low.TypeName())
					return
				}
			}

			switch left := left.(type) {
			case *objects.Array:
				numElements := int64(len(left.Value))
				var highIdx int64
				if high == objects.UndefinedValue {
					highIdx = numElements
				} else if high, ok := high.(*objects.Int); ok {
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

				var val objects.Object = &objects.Array{Value: left.Value[lowIdx:highIdx]}

				v.allocs--
				if v.allocs == 0 {
					v.err = ErrObjectAllocLimit
					return
				}

				v.stack[v.sp] = val
				v.sp++

			case *objects.ImmutableArray:
				numElements := int64(len(left.Value))
				var highIdx int64
				if high == objects.UndefinedValue {
					highIdx = numElements
				} else if high, ok := high.(*objects.Int); ok {
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

				var val objects.Object = &objects.Array{Value: left.Value[lowIdx:highIdx]}

				v.allocs--
				if v.allocs == 0 {
					v.err = ErrObjectAllocLimit
					return
				}

				v.stack[v.sp] = val
				v.sp++

			case *objects.String:
				numElements := int64(len(left.Value))
				var highIdx int64
				if high == objects.UndefinedValue {
					highIdx = numElements
				} else if high, ok := high.(*objects.Int); ok {
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

				var val objects.Object = &objects.String{Value: left.Value[lowIdx:highIdx]}

				v.allocs--
				if v.allocs == 0 {
					v.err = ErrObjectAllocLimit
					return
				}

				v.stack[v.sp] = val
				v.sp++

			case *objects.Bytes:
				numElements := int64(len(left.Value))
				var highIdx int64
				if high == objects.UndefinedValue {
					highIdx = numElements
				} else if high, ok := high.(*objects.Int); ok {
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

				var val objects.Object = &objects.Bytes{Value: left.Value[lowIdx:highIdx]}

				v.allocs--
				if v.allocs == 0 {
					v.err = ErrObjectAllocLimit
					return
				}

				v.stack[v.sp] = val
				v.sp++
			}

		case compiler.OpCall:
			numArgs := int(v.curInsts[v.ip+1])
			v.ip++

			value := v.stack[v.sp-1-numArgs]

			switch callee := value.(type) {
			case *objects.Closure:
				if callee.Fn.VarArgs {
					// if the closure is variadic,
					// roll up all variadic parameters into an array
					realArgs := callee.Fn.NumParameters - 1
					varArgs := numArgs - realArgs
					if varArgs >= 0 {
						numArgs = realArgs + 1
						args := make([]objects.Object, varArgs)
						spStart := v.sp - varArgs
						for i := spStart; i < v.sp; i++ {
							args[i-spStart] = v.stack[i]
						}
						v.stack[spStart] = &objects.Array{Value: args}
						v.sp = spStart + 1
					}
				}

				if numArgs != callee.Fn.NumParameters {
					if callee.Fn.VarArgs {
						v.err = fmt.Errorf("wrong number of arguments: want>=%d, got=%d",
							callee.Fn.NumParameters-1, numArgs)
					} else {
						v.err = fmt.Errorf("wrong number of arguments: want=%d, got=%d",
							callee.Fn.NumParameters, numArgs)
					}
					return
				}

				// test if it's tail-call
				if callee.Fn == v.curFrame.fn { // recursion
					nextOp := v.curInsts[v.ip+1]
					if nextOp == compiler.OpReturn ||
						(nextOp == compiler.OpPop && compiler.OpReturn == v.curInsts[v.ip+2]) {
						for p := 0; p < numArgs; p++ {
							v.stack[v.curFrame.basePointer+p] = v.stack[v.sp-numArgs+p]
						}
						v.sp -= numArgs + 1
						v.ip = -1 // reset IP to beginning of the frame
						continue
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
				v.framesIndex++
				v.sp = v.sp - numArgs + callee.Fn.NumLocals

			case *objects.CompiledFunction:
				if callee.VarArgs {
					// if the closure is variadic,
					// roll up all variadic parameters into an array
					realArgs := callee.NumParameters - 1
					varArgs := numArgs - realArgs
					if varArgs >= 0 {
						numArgs = realArgs + 1
						args := make([]objects.Object, varArgs)
						spStart := v.sp - varArgs
						for i := spStart; i < v.sp; i++ {
							args[i-spStart] = v.stack[i]
						}
						v.stack[spStart] = &objects.Array{Value: args}
						v.sp = spStart + 1
					}
				}

				if numArgs != callee.NumParameters {
					if callee.VarArgs {
						v.err = fmt.Errorf("wrong number of arguments: want>=%d, got=%d",
							callee.NumParameters-1, numArgs)
					} else {
						v.err = fmt.Errorf("wrong number of arguments: want=%d, got=%d",
							callee.NumParameters, numArgs)
					}
					return
				}

				// test if it's tail-call
				if callee == v.curFrame.fn { // recursion
					nextOp := v.curInsts[v.ip+1]
					if nextOp == compiler.OpReturn ||
						(nextOp == compiler.OpPop && compiler.OpReturn == v.curInsts[v.ip+2]) {
						for p := 0; p < numArgs; p++ {
							v.stack[v.curFrame.basePointer+p] = v.stack[v.sp-numArgs+p]
						}
						v.sp -= numArgs + 1
						v.ip = -1 // reset IP to beginning of the frame
						continue
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
				v.framesIndex++
				v.sp = v.sp - numArgs + callee.NumLocals

			case objects.Callable:
				var args []objects.Object
				args = append(args, v.stack[v.sp-numArgs:v.sp]...)

				ret, e := callee.Call(args...)
				v.sp -= numArgs + 1

				// runtime error
				if e != nil {
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

				v.allocs--
				if v.allocs == 0 {
					v.err = ErrObjectAllocLimit
					return
				}

				v.stack[v.sp] = ret
				v.sp++

			default:
				v.err = fmt.Errorf("not callable: %s", callee.TypeName())
				return
			}

		case compiler.OpReturn:
			v.ip++
			var retVal objects.Object
			if int(v.curInsts[v.ip]) == 1 {
				retVal = v.stack[v.sp-1]
			} else {
				retVal = objects.UndefinedValue
			}
			//v.sp--

			v.framesIndex--
			v.curFrame = &v.frames[v.framesIndex-1]
			v.curInsts = v.curFrame.fn.Instructions
			v.ip = v.curFrame.ip

			//v.sp = lastFrame.basePointer - 1
			v.sp = v.frames[v.framesIndex].basePointer

			// skip stack overflow check because (newSP) <= (oldSP)
			v.stack[v.sp-1] = retVal
			//v.sp++

		case compiler.OpDefineLocal:
			v.ip++
			localIndex := int(v.curInsts[v.ip])

			sp := v.curFrame.basePointer + localIndex

			// local variables can be mutated by other actions
			// so always store the copy of popped value
			val := v.stack[v.sp-1]
			v.sp--

			v.stack[sp] = val

		case compiler.OpSetLocal:
			localIndex := int(v.curInsts[v.ip+1])
			v.ip++

			sp := v.curFrame.basePointer + localIndex

			// update pointee of v.stack[sp] instead of replacing the pointer itself.
			// this is needed because there can be free variables referencing the same local variables.
			val := v.stack[v.sp-1]
			v.sp--

			if obj, ok := v.stack[sp].(*objects.ObjectPtr); ok {
				*obj.Value = val
				val = obj
			}
			v.stack[sp] = val // also use a copy of popped value

		case compiler.OpSetSelLocal:
			localIndex := int(v.curInsts[v.ip+1])
			numSelectors := int(v.curInsts[v.ip+2])
			v.ip += 2

			// selectors and RHS value
			selectors := make([]objects.Object, numSelectors)
			for i := 0; i < numSelectors; i++ {
				selectors[i] = v.stack[v.sp-numSelectors+i]
			}

			val := v.stack[v.sp-numSelectors-1]
			v.sp -= numSelectors + 1

			dst := v.stack[v.curFrame.basePointer+localIndex]
			if obj, ok := dst.(*objects.ObjectPtr); ok {
				dst = *obj.Value
			}

			if e := indexAssign(dst, val, selectors); e != nil {
				v.err = e
				return
			}

		case compiler.OpGetLocal:
			v.ip++
			localIndex := int(v.curInsts[v.ip])

			val := v.stack[v.curFrame.basePointer+localIndex]

			if obj, ok := val.(*objects.ObjectPtr); ok {
				val = *obj.Value
			}

			v.stack[v.sp] = val
			v.sp++

		case compiler.OpGetBuiltin:
			v.ip++
			builtinIndex := int(v.curInsts[v.ip])

			v.stack[v.sp] = objects.Builtins[builtinIndex]
			v.sp++

		case compiler.OpClosure:
			v.ip += 3
			constIndex := int(v.curInsts[v.ip-1]) | int(v.curInsts[v.ip-2])<<8
			numFree := int(v.curInsts[v.ip])

			fn, ok := v.constants[constIndex].(*objects.CompiledFunction)
			if !ok {
				v.err = fmt.Errorf("not function: %s", fn.TypeName())
				return
			}

			free := make([]*objects.ObjectPtr, numFree)
			for i := 0; i < numFree; i++ {
				switch freeVar := (v.stack[v.sp-numFree+i]).(type) {
				case *objects.ObjectPtr:
					free[i] = freeVar
				default:
					free[i] = &objects.ObjectPtr{Value: &v.stack[v.sp-numFree+i]}
				}
			}

			v.sp -= numFree

			var cl = &objects.Closure{
				Fn:   fn,
				Free: free,
			}

			v.allocs--
			if v.allocs == 0 {
				v.err = ErrObjectAllocLimit
				return
			}

			v.stack[v.sp] = cl
			v.sp++

		case compiler.OpGetFreePtr:
			v.ip++
			freeIndex := int(v.curInsts[v.ip])

			val := v.curFrame.freeVars[freeIndex]

			v.stack[v.sp] = val
			v.sp++

		case compiler.OpGetFree:
			v.ip++
			freeIndex := int(v.curInsts[v.ip])

			val := *v.curFrame.freeVars[freeIndex].Value

			v.stack[v.sp] = val
			v.sp++

		case compiler.OpSetFree:
			v.ip++
			freeIndex := int(v.curInsts[v.ip])

			*v.curFrame.freeVars[freeIndex].Value = v.stack[v.sp-1]

			v.sp--

		case compiler.OpGetLocalPtr:
			v.ip++
			localIndex := int(v.curInsts[v.ip])

			sp := v.curFrame.basePointer + localIndex
			val := v.stack[sp]

			var freeVar *objects.ObjectPtr
			if obj, ok := val.(*objects.ObjectPtr); ok {
				freeVar = obj
			} else {
				freeVar = &objects.ObjectPtr{Value: &val}
				v.stack[sp] = freeVar
			}

			v.stack[v.sp] = freeVar
			v.sp++

		case compiler.OpSetSelFree:
			v.ip += 2
			freeIndex := int(v.curInsts[v.ip-1])
			numSelectors := int(v.curInsts[v.ip])

			// selectors and RHS value
			selectors := make([]objects.Object, numSelectors)
			for i := 0; i < numSelectors; i++ {
				selectors[i] = v.stack[v.sp-numSelectors+i]
			}
			val := v.stack[v.sp-numSelectors-1]
			v.sp -= numSelectors + 1

			if e := indexAssign(*v.curFrame.freeVars[freeIndex].Value, val, selectors); e != nil {
				v.err = e
				return
			}

		case compiler.OpIteratorInit:
			var iterator objects.Object

			dst := v.stack[v.sp-1]
			v.sp--

			iterable, ok := dst.(objects.Iterable)
			if !ok {
				v.err = fmt.Errorf("not iterable: %s", dst.TypeName())
				return
			}

			iterator = iterable.Iterate()

			v.allocs--
			if v.allocs == 0 {
				v.err = ErrObjectAllocLimit
				return
			}

			v.stack[v.sp] = iterator
			v.sp++

		case compiler.OpIteratorNext:
			iterator := v.stack[v.sp-1]
			v.sp--

			hasMore := iterator.(objects.Iterator).Next()

			if hasMore {
				v.stack[v.sp] = objects.TrueValue
			} else {
				v.stack[v.sp] = objects.FalseValue
			}
			v.sp++

		case compiler.OpIteratorKey:
			iterator := v.stack[v.sp-1]
			v.sp--

			val := iterator.(objects.Iterator).Key()

			v.stack[v.sp] = val
			v.sp++

		case compiler.OpIteratorValue:
			iterator := v.stack[v.sp-1]
			v.sp--

			val := iterator.(objects.Iterator).Value()

			v.stack[v.sp] = val
			v.sp++

		default:
			v.err = fmt.Errorf("unknown opcode: %d", v.curInsts[v.ip])
			return
		}
	}
}

// IsStackEmpty tests if the stack is empty or not.
func (v *VM) IsStackEmpty() bool {
	return v.sp == 0
}

func indexAssign(dst, src objects.Object, selectors []objects.Object) error {
	numSel := len(selectors)

	for sidx := numSel - 1; sidx > 0; sidx-- {
		indexable, ok := dst.(objects.Indexable)
		if !ok {
			return fmt.Errorf("not indexable: %s", dst.TypeName())
		}

		next, err := indexable.IndexGet(selectors[sidx])
		if err != nil {
			if err == objects.ErrInvalidIndexType {
				return fmt.Errorf("invalid index type: %s", selectors[sidx].TypeName())
			}

			return err
		}

		dst = next
	}

	indexAssignable, ok := dst.(objects.IndexAssignable)
	if !ok {
		return fmt.Errorf("not index-assignable: %s", dst.TypeName())
	}

	if err := indexAssignable.IndexSet(selectors[0], src); err != nil {
		if err == objects.ErrInvalidIndexValueType {
			return fmt.Errorf("invaid index value type: %s", src.TypeName())
		}

		return err
	}

	return nil
}
