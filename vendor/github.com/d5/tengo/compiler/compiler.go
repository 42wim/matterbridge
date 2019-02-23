package compiler

import (
	"fmt"
	"io"
	"reflect"

	"github.com/d5/tengo/compiler/ast"
	"github.com/d5/tengo/compiler/source"
	"github.com/d5/tengo/compiler/token"
	"github.com/d5/tengo/objects"
	"github.com/d5/tengo/stdlib"
)

// Compiler compiles the AST into a bytecode.
type Compiler struct {
	file            *source.File
	parent          *Compiler
	moduleName      string
	constants       []objects.Object
	symbolTable     *SymbolTable
	scopes          []CompilationScope
	scopeIndex      int
	moduleLoader    ModuleLoader
	builtinModules  map[string]bool
	compiledModules map[string]*objects.CompiledFunction
	loops           []*Loop
	loopIndex       int
	trace           io.Writer
	indent          int
}

// NewCompiler creates a Compiler.
// User can optionally provide the symbol table if one wants to add or remove
// some global- or builtin- scope symbols. If not (nil), Compile will create
// a new symbol table and use the default builtin functions. Likewise, standard
// modules can be explicitly provided if user wants to add or remove some modules.
// By default, Compile will use all the standard modules otherwise.
func NewCompiler(file *source.File, symbolTable *SymbolTable, constants []objects.Object, builtinModules map[string]bool, trace io.Writer) *Compiler {
	mainScope := CompilationScope{
		symbolInit: make(map[string]bool),
		sourceMap:  make(map[int]source.Pos),
	}

	// symbol table
	if symbolTable == nil {
		symbolTable = NewSymbolTable()

		for idx, fn := range objects.Builtins {
			symbolTable.DefineBuiltin(idx, fn.Name)
		}
	}

	// builtin modules
	if builtinModules == nil {
		builtinModules = make(map[string]bool)
		for name := range stdlib.Modules {
			builtinModules[name] = true
		}
	}

	return &Compiler{
		file:            file,
		symbolTable:     symbolTable,
		constants:       constants,
		scopes:          []CompilationScope{mainScope},
		scopeIndex:      0,
		loopIndex:       -1,
		trace:           trace,
		builtinModules:  builtinModules,
		compiledModules: make(map[string]*objects.CompiledFunction),
	}
}

// Compile compiles the AST node.
func (c *Compiler) Compile(node ast.Node) error {
	if c.trace != nil {
		if node != nil {
			defer un(trace(c, fmt.Sprintf("%s (%s)", node.String(), reflect.TypeOf(node).Elem().Name())))
		} else {
			defer un(trace(c, "<nil>"))
		}
	}

	switch node := node.(type) {
	case *ast.File:
		for _, stmt := range node.Stmts {
			if err := c.Compile(stmt); err != nil {
				return err
			}
		}

	case *ast.ExprStmt:
		if err := c.Compile(node.Expr); err != nil {
			return err
		}
		c.emit(node, OpPop)

	case *ast.IncDecStmt:
		op := token.AddAssign
		if node.Token == token.Dec {
			op = token.SubAssign
		}

		return c.compileAssign(node, []ast.Expr{node.Expr}, []ast.Expr{&ast.IntLit{Value: 1}}, op)

	case *ast.ParenExpr:
		if err := c.Compile(node.Expr); err != nil {
			return err
		}

	case *ast.BinaryExpr:
		if node.Token == token.LAnd || node.Token == token.LOr {
			return c.compileLogical(node)
		}

		if node.Token == token.Less {
			if err := c.Compile(node.RHS); err != nil {
				return err
			}

			if err := c.Compile(node.LHS); err != nil {
				return err
			}

			c.emit(node, OpGreaterThan)

			return nil
		} else if node.Token == token.LessEq {
			if err := c.Compile(node.RHS); err != nil {
				return err
			}
			if err := c.Compile(node.LHS); err != nil {
				return err
			}

			c.emit(node, OpGreaterThanEqual)

			return nil
		}

		if err := c.Compile(node.LHS); err != nil {
			return err
		}
		if err := c.Compile(node.RHS); err != nil {
			return err
		}

		switch node.Token {
		case token.Add:
			c.emit(node, OpAdd)
		case token.Sub:
			c.emit(node, OpSub)
		case token.Mul:
			c.emit(node, OpMul)
		case token.Quo:
			c.emit(node, OpDiv)
		case token.Rem:
			c.emit(node, OpRem)
		case token.Greater:
			c.emit(node, OpGreaterThan)
		case token.GreaterEq:
			c.emit(node, OpGreaterThanEqual)
		case token.Equal:
			c.emit(node, OpEqual)
		case token.NotEqual:
			c.emit(node, OpNotEqual)
		case token.And:
			c.emit(node, OpBAnd)
		case token.Or:
			c.emit(node, OpBOr)
		case token.Xor:
			c.emit(node, OpBXor)
		case token.AndNot:
			c.emit(node, OpBAndNot)
		case token.Shl:
			c.emit(node, OpBShiftLeft)
		case token.Shr:
			c.emit(node, OpBShiftRight)
		default:
			return c.errorf(node, "invalid binary operator: %s", node.Token.String())
		}

	case *ast.IntLit:
		c.emit(node, OpConstant, c.addConstant(&objects.Int{Value: node.Value}))

	case *ast.FloatLit:
		c.emit(node, OpConstant, c.addConstant(&objects.Float{Value: node.Value}))

	case *ast.BoolLit:
		if node.Value {
			c.emit(node, OpTrue)
		} else {
			c.emit(node, OpFalse)
		}

	case *ast.StringLit:
		c.emit(node, OpConstant, c.addConstant(&objects.String{Value: node.Value}))

	case *ast.CharLit:
		c.emit(node, OpConstant, c.addConstant(&objects.Char{Value: node.Value}))

	case *ast.UndefinedLit:
		c.emit(node, OpNull)

	case *ast.UnaryExpr:
		if err := c.Compile(node.Expr); err != nil {
			return err
		}

		switch node.Token {
		case token.Not:
			c.emit(node, OpLNot)
		case token.Sub:
			c.emit(node, OpMinus)
		case token.Xor:
			c.emit(node, OpBComplement)
		case token.Add:
			// do nothing?
		default:
			return c.errorf(node, "invalid unary operator: %s", node.Token.String())
		}

	case *ast.IfStmt:
		// open new symbol table for the statement
		c.symbolTable = c.symbolTable.Fork(true)
		defer func() {
			c.symbolTable = c.symbolTable.Parent(false)
		}()

		if node.Init != nil {
			if err := c.Compile(node.Init); err != nil {
				return err
			}
		}

		if err := c.Compile(node.Cond); err != nil {
			return err
		}

		// first jump placeholder
		jumpPos1 := c.emit(node, OpJumpFalsy, 0)

		if err := c.Compile(node.Body); err != nil {
			return err
		}

		if node.Else != nil {
			// second jump placeholder
			jumpPos2 := c.emit(node, OpJump, 0)

			// update first jump offset
			curPos := len(c.currentInstructions())
			c.changeOperand(jumpPos1, curPos)

			if err := c.Compile(node.Else); err != nil {
				return err
			}

			// update second jump offset
			curPos = len(c.currentInstructions())
			c.changeOperand(jumpPos2, curPos)
		} else {
			// update first jump offset
			curPos := len(c.currentInstructions())
			c.changeOperand(jumpPos1, curPos)
		}

	case *ast.ForStmt:
		return c.compileForStmt(node)

	case *ast.ForInStmt:
		return c.compileForInStmt(node)

	case *ast.BranchStmt:
		if node.Token == token.Break {
			curLoop := c.currentLoop()
			if curLoop == nil {
				return c.errorf(node, "break not allowed outside loop")
			}
			pos := c.emit(node, OpJump, 0)
			curLoop.Breaks = append(curLoop.Breaks, pos)
		} else if node.Token == token.Continue {
			curLoop := c.currentLoop()
			if curLoop == nil {
				return c.errorf(node, "continue not allowed outside loop")
			}
			pos := c.emit(node, OpJump, 0)
			curLoop.Continues = append(curLoop.Continues, pos)
		} else {
			panic(fmt.Errorf("invalid branch statement: %s", node.Token.String()))
		}

	case *ast.BlockStmt:
		for _, stmt := range node.Stmts {
			if err := c.Compile(stmt); err != nil {
				return err
			}
		}

	case *ast.AssignStmt:
		if err := c.compileAssign(node, node.LHS, node.RHS, node.Token); err != nil {
			return err
		}

	case *ast.Ident:
		symbol, _, ok := c.symbolTable.Resolve(node.Name)
		if !ok {
			return c.errorf(node, "unresolved reference '%s'", node.Name)
		}

		switch symbol.Scope {
		case ScopeGlobal:
			c.emit(node, OpGetGlobal, symbol.Index)
		case ScopeLocal:
			c.emit(node, OpGetLocal, symbol.Index)
		case ScopeBuiltin:
			c.emit(node, OpGetBuiltin, symbol.Index)
		case ScopeFree:
			c.emit(node, OpGetFree, symbol.Index)
		}

	case *ast.ArrayLit:
		for _, elem := range node.Elements {
			if err := c.Compile(elem); err != nil {
				return err
			}
		}

		c.emit(node, OpArray, len(node.Elements))

	case *ast.MapLit:
		for _, elt := range node.Elements {
			// key
			c.emit(node, OpConstant, c.addConstant(&objects.String{Value: elt.Key}))

			// value
			if err := c.Compile(elt.Value); err != nil {
				return err
			}
		}

		c.emit(node, OpMap, len(node.Elements)*2)

	case *ast.SelectorExpr: // selector on RHS side
		if err := c.Compile(node.Expr); err != nil {
			return err
		}

		if err := c.Compile(node.Sel); err != nil {
			return err
		}

		c.emit(node, OpIndex)

	case *ast.IndexExpr:
		if err := c.Compile(node.Expr); err != nil {
			return err
		}

		if err := c.Compile(node.Index); err != nil {
			return err
		}

		c.emit(node, OpIndex)

	case *ast.SliceExpr:
		if err := c.Compile(node.Expr); err != nil {
			return err
		}

		if node.Low != nil {
			if err := c.Compile(node.Low); err != nil {
				return err
			}
		} else {
			c.emit(node, OpNull)
		}

		if node.High != nil {
			if err := c.Compile(node.High); err != nil {
				return err
			}
		} else {
			c.emit(node, OpNull)
		}

		c.emit(node, OpSliceIndex)

	case *ast.FuncLit:
		c.enterScope()

		for _, p := range node.Type.Params.List {
			s := c.symbolTable.Define(p.Name)

			// function arguments is not assigned directly.
			s.LocalAssigned = true
		}

		if err := c.Compile(node.Body); err != nil {
			return err
		}

		// add OpReturn if function returns nothing
		if !c.lastInstructionIs(OpReturnValue) && !c.lastInstructionIs(OpReturn) {
			c.emit(node, OpReturn)
		}

		freeSymbols := c.symbolTable.FreeSymbols()
		numLocals := c.symbolTable.MaxSymbols()
		instructions, sourceMap := c.leaveScope()

		for _, s := range freeSymbols {
			switch s.Scope {
			case ScopeLocal:
				if !s.LocalAssigned {
					// Here, the closure is capturing a local variable that's not yet assigned its value.
					// One example is a local recursive function:
					//
					//   func() {
					//     foo := func(x) {
					//       // ..
					//       return foo(x-1)
					//     }
					//   }
					//
					// which translate into
					//
					//   0000 GETL    0
					//   0002 CLOSURE ?     1
					//   0006 DEFL    0
					//
					// . So the local variable (0) is being captured before it's assigned the value.
					//
					// Solution is to transform the code into something like this:
					//
					//   func() {
					//     foo := undefined
					//     foo = func(x) {
					//       // ..
					//       return foo(x-1)
					//     }
					//   }
					//
					// that is equivalent to
					//
					//   0000 NULL
					//   0001 DEFL    0
					//   0003 GETL    0
					//   0005 CLOSURE ?     1
					//   0009 SETL    0
					//

					c.emit(node, OpNull)
					c.emit(node, OpDefineLocal, s.Index)

					s.LocalAssigned = true
				}

				c.emit(node, OpGetLocal, s.Index)
			case ScopeFree:
				c.emit(node, OpGetFree, s.Index)
			}
		}

		compiledFunction := &objects.CompiledFunction{
			Instructions:  instructions,
			NumLocals:     numLocals,
			NumParameters: len(node.Type.Params.List),
			SourceMap:     sourceMap,
		}

		if len(freeSymbols) > 0 {
			c.emit(node, OpClosure, c.addConstant(compiledFunction), len(freeSymbols))
		} else {
			c.emit(node, OpConstant, c.addConstant(compiledFunction))
		}

	case *ast.ReturnStmt:
		if c.symbolTable.Parent(true) == nil {
			// outside the function
			return c.errorf(node, "return not allowed outside function")
		}

		if node.Result == nil {
			c.emit(node, OpReturn)
		} else {
			if err := c.Compile(node.Result); err != nil {
				return err
			}

			c.emit(node, OpReturnValue)
		}

	case *ast.CallExpr:
		if err := c.Compile(node.Func); err != nil {
			return err
		}

		for _, arg := range node.Args {
			if err := c.Compile(arg); err != nil {
				return err
			}
		}

		c.emit(node, OpCall, len(node.Args))

	case *ast.ImportExpr:
		if c.builtinModules[node.ModuleName] {
			c.emit(node, OpConstant, c.addConstant(&objects.String{Value: node.ModuleName}))
			c.emit(node, OpGetBuiltinModule)
		} else {
			userMod, err := c.compileModule(node)
			if err != nil {
				return err
			}

			c.emit(node, OpConstant, c.addConstant(userMod))
			c.emit(node, OpCall, 0)
		}

	case *ast.ExportStmt:
		// export statement must be in top-level scope
		if c.scopeIndex != 0 {
			return c.errorf(node, "export not allowed inside function")
		}

		// export statement is simply ignore when compiling non-module code
		if c.parent == nil {
			break
		}

		if err := c.Compile(node.Result); err != nil {
			return err
		}

		c.emit(node, OpImmutable)
		c.emit(node, OpReturnValue)

	case *ast.ErrorExpr:
		if err := c.Compile(node.Expr); err != nil {
			return err
		}

		c.emit(node, OpError)

	case *ast.ImmutableExpr:
		if err := c.Compile(node.Expr); err != nil {
			return err
		}

		c.emit(node, OpImmutable)

	case *ast.CondExpr:
		if err := c.Compile(node.Cond); err != nil {
			return err
		}

		// first jump placeholder
		jumpPos1 := c.emit(node, OpJumpFalsy, 0)

		if err := c.Compile(node.True); err != nil {
			return err
		}

		// second jump placeholder
		jumpPos2 := c.emit(node, OpJump, 0)

		// update first jump offset
		curPos := len(c.currentInstructions())
		c.changeOperand(jumpPos1, curPos)

		if err := c.Compile(node.False); err != nil {
			return err
		}

		// update second jump offset
		curPos = len(c.currentInstructions())
		c.changeOperand(jumpPos2, curPos)
	}

	return nil
}

// Bytecode returns a compiled bytecode.
func (c *Compiler) Bytecode() *Bytecode {
	return &Bytecode{
		FileSet: c.file.Set(),
		MainFunction: &objects.CompiledFunction{
			Instructions: c.currentInstructions(),
			SourceMap:    c.currentSourceMap(),
		},
		Constants: c.constants,
	}
}

// SetModuleLoader sets or replaces the current module loader.
// Note that the module loader is used for user modules,
// not for the standard modules.
func (c *Compiler) SetModuleLoader(moduleLoader ModuleLoader) {
	c.moduleLoader = moduleLoader
}

func (c *Compiler) fork(file *source.File, moduleName string, symbolTable *SymbolTable) *Compiler {
	child := NewCompiler(file, symbolTable, nil, c.builtinModules, c.trace)
	child.moduleName = moduleName       // name of the module to compile
	child.parent = c                    // parent to set to current compiler
	child.moduleLoader = c.moduleLoader // share module loader

	return child
}

func (c *Compiler) errorf(node ast.Node, format string, args ...interface{}) error {
	return &Error{
		fileSet: c.file.Set(),
		node:    node,
		error:   fmt.Errorf(format, args...),
	}
}

func (c *Compiler) addConstant(o objects.Object) int {
	if c.parent != nil {
		// module compilers will use their parent's constants array
		return c.parent.addConstant(o)
	}

	c.constants = append(c.constants, o)

	if c.trace != nil {
		c.printTrace(fmt.Sprintf("CONST %04d %s", len(c.constants)-1, o))
	}

	return len(c.constants) - 1
}

func (c *Compiler) addInstruction(b []byte) int {
	posNewIns := len(c.currentInstructions())

	c.scopes[c.scopeIndex].instructions = append(c.currentInstructions(), b...)

	return posNewIns
}

func (c *Compiler) setLastInstruction(op Opcode, pos int) {
	c.scopes[c.scopeIndex].lastInstructions[1] = c.scopes[c.scopeIndex].lastInstructions[0]

	c.scopes[c.scopeIndex].lastInstructions[0].Opcode = op
	c.scopes[c.scopeIndex].lastInstructions[0].Position = pos
}

func (c *Compiler) lastInstructionIs(op Opcode) bool {
	if len(c.currentInstructions()) == 0 {
		return false
	}

	return c.scopes[c.scopeIndex].lastInstructions[0].Opcode == op
}

func (c *Compiler) removeLastInstruction() {
	lastPos := c.scopes[c.scopeIndex].lastInstructions[0].Position

	if c.trace != nil {
		c.printTrace(fmt.Sprintf("DELET %s",
			FormatInstructions(c.scopes[c.scopeIndex].instructions[lastPos:], lastPos)[0]))
	}

	c.scopes[c.scopeIndex].instructions = c.currentInstructions()[:lastPos]
	c.scopes[c.scopeIndex].lastInstructions[0] = c.scopes[c.scopeIndex].lastInstructions[1]
}

func (c *Compiler) replaceInstruction(pos int, inst []byte) {
	copy(c.currentInstructions()[pos:], inst)

	if c.trace != nil {
		c.printTrace(fmt.Sprintf("REPLC %s",
			FormatInstructions(c.scopes[c.scopeIndex].instructions[pos:], pos)[0]))
	}
}

func (c *Compiler) changeOperand(opPos int, operand ...int) {
	op := Opcode(c.currentInstructions()[opPos])
	inst := MakeInstruction(op, operand...)

	c.replaceInstruction(opPos, inst)
}

func (c *Compiler) emit(node ast.Node, opcode Opcode, operands ...int) int {
	filePos := source.NoPos
	if node != nil {
		filePos = node.Pos()
	}

	inst := MakeInstruction(opcode, operands...)
	pos := c.addInstruction(inst)
	c.scopes[c.scopeIndex].sourceMap[pos] = filePos
	c.setLastInstruction(opcode, pos)

	if c.trace != nil {
		c.printTrace(fmt.Sprintf("EMIT  %s",
			FormatInstructions(c.scopes[c.scopeIndex].instructions[pos:], pos)[0]))
	}

	return pos
}

func (c *Compiler) printTrace(a ...interface{}) {
	const (
		dots = ". . . . . . . . . . . . . . . . . . . . . . . . . . . . . . . . "
		n    = len(dots)
	)

	i := 2 * c.indent
	for i > n {
		_, _ = fmt.Fprint(c.trace, dots)
		i -= n
	}
	_, _ = fmt.Fprint(c.trace, dots[0:i])
	_, _ = fmt.Fprintln(c.trace, a...)
}

func trace(c *Compiler, msg string) *Compiler {
	c.printTrace(msg, "{")
	c.indent++

	return c
}

func un(c *Compiler) {
	c.indent--
	c.printTrace("}")
}
