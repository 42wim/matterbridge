/*
	Parser parses the Tengo source files.

	Parser is a modified version of Go's parser implementation.

	Copyright 2009 The Go Authors. All rights reserved.
	Use of this source code is governed by a BSD-style
	license that can be found in the LICENSE file.
*/

package parser

import (
	"fmt"
	"io"
	"strconv"

	"github.com/d5/tengo/compiler/ast"
	"github.com/d5/tengo/compiler/scanner"
	"github.com/d5/tengo/compiler/source"
	"github.com/d5/tengo/compiler/token"
)

type bailout struct{}

// Parser parses the Tengo source files.
type Parser struct {
	file      *source.File
	errors    ErrorList
	scanner   *scanner.Scanner
	pos       source.Pos
	token     token.Token
	tokenLit  string
	exprLevel int        // < 0: in control clause, >= 0: in expression
	syncPos   source.Pos // last sync position
	syncCount int        // number of advance calls without progress
	trace     bool
	indent    int
	traceOut  io.Writer
}

// NewParser creates a Parser.
func NewParser(file *source.File, src []byte, trace io.Writer) *Parser {
	p := &Parser{
		file:     file,
		trace:    trace != nil,
		traceOut: trace,
	}

	p.scanner = scanner.NewScanner(p.file, src, func(pos source.FilePos, msg string) {
		p.errors.Add(pos, msg)
	}, 0)

	p.next()

	return p
}

// ParseFile parses the source and returns an AST file unit.
func (p *Parser) ParseFile() (file *ast.File, err error) {
	defer func() {
		if e := recover(); e != nil {
			if _, ok := e.(bailout); !ok {
				panic(e)
			}
		}

		p.errors.Sort()
		err = p.errors.Err()
	}()

	if p.trace {
		defer un(trace(p, "File"))
	}

	if p.errors.Len() > 0 {
		return nil, p.errors.Err()
	}

	stmts := p.parseStmtList()
	if p.errors.Len() > 0 {
		return nil, p.errors.Err()
	}

	file = &ast.File{
		InputFile: p.file,
		Stmts:     stmts,
	}

	return
}

func (p *Parser) parseExpr() ast.Expr {
	if p.trace {
		defer un(trace(p, "Expression"))
	}

	expr := p.parseBinaryExpr(token.LowestPrec + 1)

	// ternary conditional expression
	if p.token == token.Question {
		return p.parseCondExpr(expr)
	}

	return expr
}

func (p *Parser) parseBinaryExpr(prec1 int) ast.Expr {
	if p.trace {
		defer un(trace(p, "BinaryExpression"))
	}

	x := p.parseUnaryExpr()

	for {
		op, prec := p.token, p.token.Precedence()
		if prec < prec1 {
			return x
		}

		pos := p.expect(op)

		y := p.parseBinaryExpr(prec + 1)

		x = &ast.BinaryExpr{
			LHS:      x,
			RHS:      y,
			Token:    op,
			TokenPos: pos,
		}
	}
}

func (p *Parser) parseCondExpr(cond ast.Expr) ast.Expr {
	questionPos := p.expect(token.Question)

	trueExpr := p.parseExpr()

	colonPos := p.expect(token.Colon)

	falseExpr := p.parseExpr()

	return &ast.CondExpr{
		Cond:        cond,
		True:        trueExpr,
		False:       falseExpr,
		QuestionPos: questionPos,
		ColonPos:    colonPos,
	}
}

func (p *Parser) parseUnaryExpr() ast.Expr {
	if p.trace {
		defer un(trace(p, "UnaryExpression"))
	}

	switch p.token {
	case token.Add, token.Sub, token.Not, token.Xor:
		pos, op := p.pos, p.token
		p.next()
		x := p.parseUnaryExpr()
		return &ast.UnaryExpr{
			Token:    op,
			TokenPos: pos,
			Expr:     x,
		}
	}

	return p.parsePrimaryExpr()
}

func (p *Parser) parsePrimaryExpr() ast.Expr {
	if p.trace {
		defer un(trace(p, "PrimaryExpression"))
	}

	x := p.parseOperand()

L:
	for {
		switch p.token {
		case token.Period:
			p.next()

			switch p.token {
			case token.Ident:
				x = p.parseSelector(x)
			default:
				pos := p.pos
				p.errorExpected(pos, "selector")
				p.advance(stmtStart)
				return &ast.BadExpr{From: pos, To: p.pos}
			}
		case token.LBrack:
			x = p.parseIndexOrSlice(x)
		case token.LParen:
			x = p.parseCall(x)
		default:
			break L
		}
	}

	return x
}

func (p *Parser) parseCall(x ast.Expr) *ast.CallExpr {
	if p.trace {
		defer un(trace(p, "Call"))
	}

	lparen := p.expect(token.LParen)
	p.exprLevel++

	var list []ast.Expr
	for p.token != token.RParen && p.token != token.EOF {
		list = append(list, p.parseExpr())

		if !p.expectComma(token.RParen, "call argument") {
			break
		}
	}

	p.exprLevel--
	rparen := p.expect(token.RParen)

	return &ast.CallExpr{
		Func:   x,
		LParen: lparen,
		RParen: rparen,
		Args:   list,
	}
}

func (p *Parser) expectComma(closing token.Token, want string) bool {
	if p.token == token.Comma {
		p.next()

		if p.token == closing {
			p.errorExpected(p.pos, want)
			return false
		}

		return true
	}

	if p.token == token.Semicolon && p.tokenLit == "\n" {
		p.next()
	}

	return false
}

func (p *Parser) parseIndexOrSlice(x ast.Expr) ast.Expr {
	if p.trace {
		defer un(trace(p, "IndexOrSlice"))
	}

	lbrack := p.expect(token.LBrack)
	p.exprLevel++

	var index [2]ast.Expr
	if p.token != token.Colon {
		index[0] = p.parseExpr()
	}
	numColons := 0
	if p.token == token.Colon {
		numColons++
		p.next()

		if p.token != token.RBrack && p.token != token.EOF {
			index[1] = p.parseExpr()
		}
	}

	p.exprLevel--
	rbrack := p.expect(token.RBrack)

	if numColons > 0 {
		// slice expression
		return &ast.SliceExpr{
			Expr:   x,
			LBrack: lbrack,
			RBrack: rbrack,
			Low:    index[0],
			High:   index[1],
		}
	}

	return &ast.IndexExpr{
		Expr:   x,
		LBrack: lbrack,
		RBrack: rbrack,
		Index:  index[0],
	}
}

func (p *Parser) parseSelector(x ast.Expr) ast.Expr {
	if p.trace {
		defer un(trace(p, "Selector"))
	}

	sel := p.parseIdent()

	return &ast.SelectorExpr{Expr: x, Sel: &ast.StringLit{
		Value:    sel.Name,
		ValuePos: sel.NamePos,
		Literal:  sel.Name,
	}}
}

func (p *Parser) parseOperand() ast.Expr {
	if p.trace {
		defer un(trace(p, "Operand"))
	}

	switch p.token {
	case token.Ident:
		return p.parseIdent()

	case token.Int:
		v, _ := strconv.ParseInt(p.tokenLit, 10, 64)
		x := &ast.IntLit{
			Value:    v,
			ValuePos: p.pos,
			Literal:  p.tokenLit,
		}
		p.next()
		return x

	case token.Float:
		v, _ := strconv.ParseFloat(p.tokenLit, 64)
		x := &ast.FloatLit{
			Value:    v,
			ValuePos: p.pos,
			Literal:  p.tokenLit,
		}
		p.next()
		return x

	case token.Char:
		return p.parseCharLit()

	case token.String:
		v, _ := strconv.Unquote(p.tokenLit)
		x := &ast.StringLit{
			Value:    v,
			ValuePos: p.pos,
			Literal:  p.tokenLit,
		}
		p.next()
		return x

	case token.True:
		x := &ast.BoolLit{
			Value:    true,
			ValuePos: p.pos,
			Literal:  p.tokenLit,
		}
		p.next()
		return x

	case token.False:
		x := &ast.BoolLit{
			Value:    false,
			ValuePos: p.pos,
			Literal:  p.tokenLit,
		}
		p.next()
		return x

	case token.Undefined:
		x := &ast.UndefinedLit{TokenPos: p.pos}
		p.next()
		return x

	case token.Import:
		return p.parseImportExpr()

	case token.LParen:
		lparen := p.pos
		p.next()
		p.exprLevel++
		x := p.parseExpr()
		p.exprLevel--
		rparen := p.expect(token.RParen)
		return &ast.ParenExpr{
			LParen: lparen,
			Expr:   x,
			RParen: rparen,
		}

	case token.LBrack: // array literal
		return p.parseArrayLit()

	case token.LBrace: // map literal
		return p.parseMapLit()

	case token.Func: // function literal
		return p.parseFuncLit()

	case token.Error: // error expression
		return p.parseErrorExpr()

	case token.Immutable: // immutable expression
		return p.parseImmutableExpr()
	}

	pos := p.pos
	p.errorExpected(pos, "operand")
	p.advance(stmtStart)
	return &ast.BadExpr{From: pos, To: p.pos}
}

func (p *Parser) parseImportExpr() ast.Expr {
	pos := p.pos

	p.next()

	p.expect(token.LParen)

	if p.token != token.String {
		p.errorExpected(p.pos, "module name")
		p.advance(stmtStart)
		return &ast.BadExpr{From: pos, To: p.pos}
	}

	// module name
	moduleName, _ := strconv.Unquote(p.tokenLit)

	expr := &ast.ImportExpr{
		ModuleName: moduleName,
		Token:      token.Import,
		TokenPos:   pos,
	}

	p.next()

	p.expect(token.RParen)

	return expr
}

func (p *Parser) parseCharLit() ast.Expr {
	if n := len(p.tokenLit); n >= 3 {
		if code, _, _, err := strconv.UnquoteChar(p.tokenLit[1:n-1], '\''); err == nil {
			x := &ast.CharLit{
				Value:    rune(code),
				ValuePos: p.pos,
				Literal:  p.tokenLit,
			}
			p.next()
			return x
		}
	}

	pos := p.pos
	p.error(pos, "illegal char literal")
	p.next()
	return &ast.BadExpr{
		From: pos,
		To:   p.pos,
	}
}

func (p *Parser) parseFuncLit() ast.Expr {
	if p.trace {
		defer un(trace(p, "FuncLit"))
	}

	typ := p.parseFuncType()

	p.exprLevel++
	body := p.parseBody()
	p.exprLevel--

	return &ast.FuncLit{
		Type: typ,
		Body: body,
	}
}

func (p *Parser) parseArrayLit() ast.Expr {
	if p.trace {
		defer un(trace(p, "ArrayLit"))
	}

	lbrack := p.expect(token.LBrack)
	p.exprLevel++

	var elements []ast.Expr
	for p.token != token.RBrack && p.token != token.EOF {
		elements = append(elements, p.parseExpr())

		if !p.expectComma(token.RBrack, "array element") {
			break
		}
	}

	p.exprLevel--
	rbrack := p.expect(token.RBrack)

	return &ast.ArrayLit{
		Elements: elements,
		LBrack:   lbrack,
		RBrack:   rbrack,
	}
}

func (p *Parser) parseErrorExpr() ast.Expr {
	pos := p.pos

	p.next()

	lparen := p.expect(token.LParen)
	value := p.parseExpr()
	rparen := p.expect(token.RParen)

	expr := &ast.ErrorExpr{
		ErrorPos: pos,
		Expr:     value,
		LParen:   lparen,
		RParen:   rparen,
	}

	return expr
}

func (p *Parser) parseImmutableExpr() ast.Expr {
	pos := p.pos

	p.next()

	lparen := p.expect(token.LParen)
	value := p.parseExpr()
	rparen := p.expect(token.RParen)

	expr := &ast.ImmutableExpr{
		ErrorPos: pos,
		Expr:     value,
		LParen:   lparen,
		RParen:   rparen,
	}

	return expr
}

func (p *Parser) parseFuncType() *ast.FuncType {
	if p.trace {
		defer un(trace(p, "FuncType"))
	}

	pos := p.expect(token.Func)
	params := p.parseIdentList()

	return &ast.FuncType{
		FuncPos: pos,
		Params:  params,
	}
}

func (p *Parser) parseBody() *ast.BlockStmt {
	if p.trace {
		defer un(trace(p, "Body"))
	}

	lbrace := p.expect(token.LBrace)
	list := p.parseStmtList()
	rbrace := p.expect(token.RBrace)

	return &ast.BlockStmt{
		LBrace: lbrace,
		RBrace: rbrace,
		Stmts:  list,
	}
}

func (p *Parser) parseStmtList() (list []ast.Stmt) {
	if p.trace {
		defer un(trace(p, "StatementList"))
	}

	for p.token != token.RBrace && p.token != token.EOF {
		list = append(list, p.parseStmt())
	}

	return
}

func (p *Parser) parseIdent() *ast.Ident {
	pos := p.pos
	name := "_"

	if p.token == token.Ident {
		name = p.tokenLit
		p.next()
	} else {
		p.expect(token.Ident)
	}

	return &ast.Ident{
		NamePos: pos,
		Name:    name,
	}
}

func (p *Parser) parseIdentList() *ast.IdentList {
	if p.trace {
		defer un(trace(p, "IdentList"))
	}

	var params []*ast.Ident
	lparen := p.expect(token.LParen)
	isVarArgs := false
	if p.token != token.RParen {
		if p.token == token.Ellipsis {
			isVarArgs = true
			p.next()
		}

		params = append(params, p.parseIdent())
		for !isVarArgs && p.token == token.Comma {
			p.next()
			if p.token == token.Ellipsis {
				isVarArgs = true
				p.next()
			}
			params = append(params, p.parseIdent())
		}
	}

	rparen := p.expect(token.RParen)

	return &ast.IdentList{
		LParen:  lparen,
		RParen:  rparen,
		VarArgs: isVarArgs,
		List:    params,
	}
}

func (p *Parser) parseStmt() (stmt ast.Stmt) {
	if p.trace {
		defer un(trace(p, "Statement"))
	}

	switch p.token {
	case // simple statements
		token.Func, token.Error, token.Immutable, token.Ident, token.Int, token.Float, token.Char, token.String, token.True, token.False,
		token.Undefined, token.Import, token.LParen, token.LBrace, token.LBrack,
		token.Add, token.Sub, token.Mul, token.And, token.Xor, token.Not:
		s := p.parseSimpleStmt(false)
		p.expectSemi()
		return s
	case token.Return:
		return p.parseReturnStmt()
	case token.Export:
		return p.parseExportStmt()
	case token.If:
		return p.parseIfStmt()
	case token.For:
		return p.parseForStmt()
	case token.Break, token.Continue:
		return p.parseBranchStmt(p.token)
	case token.Semicolon:
		s := &ast.EmptyStmt{Semicolon: p.pos, Implicit: p.tokenLit == "\n"}
		p.next()
		return s
	case token.RBrace:
		// semicolon may be omitted before a closing "}"
		return &ast.EmptyStmt{Semicolon: p.pos, Implicit: true}
	default:
		pos := p.pos
		p.errorExpected(pos, "statement")
		p.advance(stmtStart)
		return &ast.BadStmt{From: pos, To: p.pos}
	}
}

func (p *Parser) parseForStmt() ast.Stmt {
	if p.trace {
		defer un(trace(p, "ForStmt"))
	}

	pos := p.expect(token.For)

	// for {}
	if p.token == token.LBrace {
		body := p.parseBlockStmt()
		p.expectSemi()

		return &ast.ForStmt{
			ForPos: pos,
			Body:   body,
		}
	}

	prevLevel := p.exprLevel
	p.exprLevel = -1

	var s1 ast.Stmt
	if p.token != token.Semicolon { // skipping init
		s1 = p.parseSimpleStmt(true)
	}

	// for _ in seq {}            or
	// for value in seq {}        or
	// for key, value in seq {}
	if forInStmt, isForIn := s1.(*ast.ForInStmt); isForIn {
		forInStmt.ForPos = pos
		p.exprLevel = prevLevel
		forInStmt.Body = p.parseBlockStmt()
		p.expectSemi()
		return forInStmt
	}

	// for init; cond; post {}
	var s2, s3 ast.Stmt
	if p.token == token.Semicolon {
		p.next()
		if p.token != token.Semicolon {
			s2 = p.parseSimpleStmt(false) // cond
		}
		p.expect(token.Semicolon)
		if p.token != token.LBrace {
			s3 = p.parseSimpleStmt(false) // post
		}
	} else {
		// for cond {}
		s2 = s1
		s1 = nil
	}

	// body
	p.exprLevel = prevLevel
	body := p.parseBlockStmt()
	p.expectSemi()

	cond := p.makeExpr(s2, "condition expression")

	return &ast.ForStmt{
		ForPos: pos,
		Init:   s1,
		Cond:   cond,
		Post:   s3,
		Body:   body,
	}

}

func (p *Parser) parseBranchStmt(tok token.Token) ast.Stmt {
	if p.trace {
		defer un(trace(p, "BranchStmt"))
	}

	pos := p.expect(tok)

	var label *ast.Ident
	if p.token == token.Ident {
		label = p.parseIdent()
	}
	p.expectSemi()

	return &ast.BranchStmt{
		Token:    tok,
		TokenPos: pos,
		Label:    label,
	}
}

func (p *Parser) parseIfStmt() ast.Stmt {
	if p.trace {
		defer un(trace(p, "IfStmt"))
	}

	pos := p.expect(token.If)

	init, cond := p.parseIfHeader()
	body := p.parseBlockStmt()

	var elseStmt ast.Stmt
	if p.token == token.Else {
		p.next()

		switch p.token {
		case token.If:
			elseStmt = p.parseIfStmt()
		case token.LBrace:
			elseStmt = p.parseBlockStmt()
			p.expectSemi()
		default:
			p.errorExpected(p.pos, "if or {")
			elseStmt = &ast.BadStmt{From: p.pos, To: p.pos}
		}
	} else {
		p.expectSemi()
	}

	return &ast.IfStmt{
		IfPos: pos,
		Init:  init,
		Cond:  cond,
		Body:  body,
		Else:  elseStmt,
	}
}

func (p *Parser) parseBlockStmt() *ast.BlockStmt {
	if p.trace {
		defer un(trace(p, "BlockStmt"))
	}

	lbrace := p.expect(token.LBrace)
	list := p.parseStmtList()
	rbrace := p.expect(token.RBrace)

	return &ast.BlockStmt{
		LBrace: lbrace,
		RBrace: rbrace,
		Stmts:  list,
	}
}

func (p *Parser) parseIfHeader() (init ast.Stmt, cond ast.Expr) {
	if p.token == token.LBrace {
		p.error(p.pos, "missing condition in if statement")
		cond = &ast.BadExpr{From: p.pos, To: p.pos}
		return
	}

	outer := p.exprLevel
	p.exprLevel = -1

	if p.token == token.Semicolon {
		p.error(p.pos, "missing init in if statement")
		return
	}

	init = p.parseSimpleStmt(false)

	var condStmt ast.Stmt
	if p.token == token.LBrace {
		condStmt = init
		init = nil
	} else if p.token == token.Semicolon {
		p.next()

		condStmt = p.parseSimpleStmt(false)
	} else {
		p.error(p.pos, "missing condition in if statement")
	}

	if condStmt != nil {
		cond = p.makeExpr(condStmt, "boolean expression")
	}

	if cond == nil {
		cond = &ast.BadExpr{From: p.pos, To: p.pos}
	}

	p.exprLevel = outer

	return
}

func (p *Parser) makeExpr(s ast.Stmt, want string) ast.Expr {
	if s == nil {
		return nil
	}

	if es, isExpr := s.(*ast.ExprStmt); isExpr {
		return es.Expr
	}

	found := "simple statement"
	if _, isAss := s.(*ast.AssignStmt); isAss {
		found = "assignment"
	}

	p.error(s.Pos(), fmt.Sprintf("expected %s, found %s", want, found))

	return &ast.BadExpr{From: s.Pos(), To: p.safePos(s.End())}
}

func (p *Parser) parseReturnStmt() ast.Stmt {
	if p.trace {
		defer un(trace(p, "ReturnStmt"))
	}

	pos := p.pos
	p.expect(token.Return)

	var x ast.Expr
	if p.token != token.Semicolon && p.token != token.RBrace {
		x = p.parseExpr()
	}
	p.expectSemi()

	return &ast.ReturnStmt{
		ReturnPos: pos,
		Result:    x,
	}
}

func (p *Parser) parseExportStmt() ast.Stmt {
	if p.trace {
		defer un(trace(p, "ExportStmt"))
	}

	pos := p.pos
	p.expect(token.Export)

	x := p.parseExpr()
	p.expectSemi()

	return &ast.ExportStmt{
		ExportPos: pos,
		Result:    x,
	}
}

func (p *Parser) parseSimpleStmt(forIn bool) ast.Stmt {
	if p.trace {
		defer un(trace(p, "SimpleStmt"))
	}

	x := p.parseExprList()

	switch p.token {
	case token.Assign, token.Define: // assignment statement
		pos, tok := p.pos, p.token
		p.next()

		y := p.parseExprList()

		return &ast.AssignStmt{
			LHS:      x,
			RHS:      y,
			Token:    tok,
			TokenPos: pos,
		}
	case token.In:
		if forIn {
			p.next()

			y := p.parseExpr()

			var key, value *ast.Ident
			var ok bool

			switch len(x) {
			case 1:
				key = &ast.Ident{Name: "_", NamePos: x[0].Pos()}

				value, ok = x[0].(*ast.Ident)
				if !ok {
					p.errorExpected(x[0].Pos(), "identifier")
					value = &ast.Ident{Name: "_", NamePos: x[0].Pos()}
				}
			case 2:
				key, ok = x[0].(*ast.Ident)
				if !ok {
					p.errorExpected(x[0].Pos(), "identifier")
					key = &ast.Ident{Name: "_", NamePos: x[0].Pos()}
				}
				value, ok = x[1].(*ast.Ident)
				if !ok {
					p.errorExpected(x[1].Pos(), "identifier")
					value = &ast.Ident{Name: "_", NamePos: x[1].Pos()}
				}
			}

			return &ast.ForInStmt{
				Key:      key,
				Value:    value,
				Iterable: y,
			}
		}
	}

	if len(x) > 1 {
		p.errorExpected(x[0].Pos(), "1 expression")
		// continue with first expression
	}

	switch p.token {
	case token.Define,
		token.AddAssign, token.SubAssign, token.MulAssign, token.QuoAssign, token.RemAssign,
		token.AndAssign, token.OrAssign, token.XorAssign, token.ShlAssign, token.ShrAssign, token.AndNotAssign:
		pos, tok := p.pos, p.token
		p.next()

		y := p.parseExpr()

		return &ast.AssignStmt{
			LHS:      []ast.Expr{x[0]},
			RHS:      []ast.Expr{y},
			Token:    tok,
			TokenPos: pos,
		}
	case token.Inc, token.Dec:
		// increment or decrement statement
		s := &ast.IncDecStmt{Expr: x[0], Token: p.token, TokenPos: p.pos}
		p.next()
		return s
	}

	// expression statement
	return &ast.ExprStmt{Expr: x[0]}
}

func (p *Parser) parseExprList() (list []ast.Expr) {
	if p.trace {
		defer un(trace(p, "ExpressionList"))
	}

	list = append(list, p.parseExpr())
	for p.token == token.Comma {
		p.next()
		list = append(list, p.parseExpr())
	}

	return
}

func (p *Parser) parseMapElementLit() *ast.MapElementLit {
	if p.trace {
		defer un(trace(p, "MapElementLit"))
	}

	pos := p.pos
	name := "_"

	if p.token == token.Ident {
		name = p.tokenLit
	} else if p.token == token.String {
		v, _ := strconv.Unquote(p.tokenLit)
		name = v
	} else {
		p.errorExpected(pos, "map key")
	}

	p.next()

	colonPos := p.expect(token.Colon)
	valueExpr := p.parseExpr()

	return &ast.MapElementLit{
		Key:      name,
		KeyPos:   pos,
		ColonPos: colonPos,
		Value:    valueExpr,
	}
}

func (p *Parser) parseMapLit() *ast.MapLit {
	if p.trace {
		defer un(trace(p, "MapLit"))
	}

	lbrace := p.expect(token.LBrace)
	p.exprLevel++

	var elements []*ast.MapElementLit
	for p.token != token.RBrace && p.token != token.EOF {
		elements = append(elements, p.parseMapElementLit())

		if !p.expectComma(token.RBrace, "map element") {
			break
		}
	}

	p.exprLevel--
	rbrace := p.expect(token.RBrace)

	return &ast.MapLit{
		LBrace:   lbrace,
		RBrace:   rbrace,
		Elements: elements,
	}
}

func (p *Parser) expect(token token.Token) source.Pos {
	pos := p.pos

	if p.token != token {
		p.errorExpected(pos, "'"+token.String()+"'")
	}
	p.next()

	return pos
}

func (p *Parser) expectSemi() {
	switch p.token {
	case token.RParen, token.RBrace:
		// semicolon is optional before a closing ')' or '}'
	case token.Comma:
		// permit a ',' instead of a ';' but complain
		p.errorExpected(p.pos, "';'")
		fallthrough
	case token.Semicolon:
		p.next()
	default:
		p.errorExpected(p.pos, "';'")
		p.advance(stmtStart)
	}

}

func (p *Parser) advance(to map[token.Token]bool) {
	for ; p.token != token.EOF; p.next() {
		if to[p.token] {
			if p.pos == p.syncPos && p.syncCount < 10 {
				p.syncCount++
				return
			}

			if p.pos > p.syncPos {
				p.syncPos = p.pos
				p.syncCount = 0
				return
			}
		}
	}
}

func (p *Parser) error(pos source.Pos, msg string) {
	filePos := p.file.Position(pos)

	n := len(p.errors)
	if n > 0 && p.errors[n-1].Pos.Line == filePos.Line {
		// discard errors reported on the same line
		return
	}

	if n > 10 {
		// too many errors; terminate early
		panic(bailout{})
	}

	p.errors.Add(filePos, msg)
}

func (p *Parser) errorExpected(pos source.Pos, msg string) {
	msg = "expected " + msg
	if pos == p.pos {
		// error happened at the current position: provide more specific
		switch {
		case p.token == token.Semicolon && p.tokenLit == "\n":
			msg += ", found newline"
		case p.token.IsLiteral():
			msg += ", found " + p.tokenLit
		default:
			msg += ", found '" + p.token.String() + "'"
		}
	}

	p.error(pos, msg)
}

func (p *Parser) next() {
	if p.trace && p.pos.IsValid() {
		s := p.token.String()
		switch {
		case p.token.IsLiteral():
			p.printTrace(s, p.tokenLit)
		case p.token.IsOperator(), p.token.IsKeyword():
			p.printTrace(`"` + s + `"`)
		default:
			p.printTrace(s)
		}
	}

	p.token, p.tokenLit, p.pos = p.scanner.Scan()
}

func (p *Parser) printTrace(a ...interface{}) {
	const (
		dots = ". . . . . . . . . . . . . . . . . . . . . . . . . . . . . . . . "
		n    = len(dots)
	)

	filePos := p.file.Position(p.pos)
	_, _ = fmt.Fprintf(p.traceOut, "%5d: %5d:%3d: ", p.pos, filePos.Line, filePos.Column)

	i := 2 * p.indent
	for i > n {
		_, _ = fmt.Fprint(p.traceOut, dots)
		i -= n
	}
	_, _ = fmt.Fprint(p.traceOut, dots[0:i])
	_, _ = fmt.Fprintln(p.traceOut, a...)
}

func (p *Parser) safePos(pos source.Pos) source.Pos {
	fileBase := p.file.Base
	fileSize := p.file.Size

	if int(pos) < fileBase || int(pos) > fileBase+fileSize {
		return source.Pos(fileBase + fileSize)
	}

	return pos
}

func trace(p *Parser, msg string) *Parser {
	p.printTrace(msg, "(")
	p.indent++

	return p
}

func un(p *Parser) {
	p.indent--
	p.printTrace(")")
}
