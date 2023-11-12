// Copyright 2015 Jean Niklas L'orange.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package edn

import (
	"strconv"
	u "unicode"
)

type lexState int

const (
	lexCont    = lexState(iota) // continue reading
	lexIgnore                   // values you can ignore, just whitespace and comments atm
	lexEnd                      // value ended with input given in
	lexEndPrev                  // value ended with previous input
	lexError                    // erroneous input
)

type tokenType int

const ( // value types from lexer
	tokenSymbol = tokenType(iota)
	tokenKeyword
	tokenString
	tokenInt
	tokenFloat
	tokenTag
	tokenChar
	tokenListStart
	tokenListEnd
	tokenVectorStart
	tokenVectorEnd
	tokenMapStart
	tokenMapEnd
	tokenSetStart
	tokenDiscard

	tokenError
)

func (t tokenType) String() string {
	switch t {
	case tokenSymbol:
		return "symbol"
	case tokenKeyword:
		return "keyword"
	case tokenString:
		return "string"
	case tokenInt:
		return "integer"
	case tokenFloat:
		return "float"
	case tokenTag:
		return "tag"
	case tokenChar:
		return "character"
	case tokenListStart:
		return "list start"
	case tokenListEnd:
		return "list end"
	case tokenVectorStart:
		return "vector start"
	case tokenVectorEnd:
		return "vector end"
	case tokenMapStart:
		return "map start"
	case tokenMapEnd:
		return "map/set end"
	case tokenSetStart:
		return "set start"
	case tokenDiscard:
		return "discard token"
	case tokenError:
		return "error"
	default:
		return "[unknown]"
	}
}

const tokenSetEnd = tokenMapEnd // sets ends the same way as maps do

// A SyntaxError is a description of an EDN syntax error.
type SyntaxError struct {
	msg    string // description of error
	Offset int64  // error occurred after reading Offset bytes
}

func (e *SyntaxError) Error() string {
	return e.msg
}

func okSymbolFirst(r rune) bool {
	switch r {
	case '.', '*', '+', '!', '-', '_', '?', '$', '%', '&', '=', '<', '>':
		return true
	}
	return false
}

func okSymbol(r rune) bool {
	switch r {
	case '.', '*', '+', '!', '-', '_', '?', '$', '%', '&', '=', '<', '>', ':', '#', '\'':
		return true
	}
	return false
}

func isWhitespace(r rune) bool {
	return u.IsSpace(r) || r == ','
}

type lexer struct {
	state    func(rune) lexState
	err      error
	position int64
	token    tokenType

	count     int    // counter is used in some functions within the lexer
	expecting []rune // expecting is used to avoid duplication when we expect e.g. \newline
}

func (l *lexer) reset() {
	l.state = l.stateBegin
	l.token = tokenType(-1)
	l.err = nil
}

func (l *lexer) eof() lexState {
	if l.err != nil {
		return lexError
	}
	lt := l.state(' ')
	if lt == lexCont {
		l.err = &SyntaxError{"unexpected end of EDN input", l.position}
		lt = lexError
	}
	if l.err != nil {
		return lexError
	}
	if lt == lexEndPrev {
		return lexEnd
	}
	return lt
}

func (l *lexer) stateBegin(r rune) lexState {
	switch {
	case isWhitespace(r):
		return lexIgnore
	case r == '{':
		l.token = tokenMapStart
		return lexEnd
	case r == '}':
		l.token = tokenMapEnd
		return lexEnd
	case r == '[':
		l.token = tokenVectorStart
		return lexEnd
	case r == ']':
		l.token = tokenVectorEnd
		return lexEnd
	case r == '(':
		l.token = tokenListStart
		return lexEnd
	case r == ')':
		l.token = tokenListEnd
		return lexEnd
	case r == '#':
		l.state = l.statePound
		return lexCont
	case r == ':':
		l.state = l.stateKeyword
		return lexCont
	case r == '/': // ohh, the lovely slash edge case
		l.token = tokenSymbol
		l.state = l.stateEndLit
		return lexCont
	case r == '+':
		l.state = l.statePos
		return lexCont
	case r == '-':
		l.state = l.stateNeg
		return lexCont
	case r == '.':
		l.token = tokenSymbol
		l.state = l.stateDotPre
		return lexCont
	case r == '"':
		l.state = l.stateInString
		return lexCont
	case r == '\\':
		l.state = l.stateChar
		return lexCont
	case okSymbolFirst(r) || u.IsLetter(r):
		l.token = tokenSymbol
		l.state = l.stateSym
		return lexCont
	case '0' < r && r <= '9':
		l.state = l.state1
		return lexCont
	case r == '0':
		l.state = l.state0
		return lexCont
	case r == ';':
		l.state = l.stateComment
		return lexIgnore
	}
	return l.error(r, "- unexpected rune")
}

func (l *lexer) stateComment(r rune) lexState {
	if r == '\n' {
		l.state = l.stateBegin
	}
	return lexIgnore
}

func (l *lexer) stateEndLit(r rune) lexState {
	if isWhitespace(r) || r == '"' || r == '{' || r == '[' || r == '(' || r == ')' || r == ']' || r == '}' || r == '\\' || r == ';' {
		return lexEndPrev
	}
	return l.error(r, "- unexpected rune after legal "+l.token.String())
}

func (l *lexer) stateKeyword(r rune) lexState {
	switch {
	case r == ':':
		l.state = l.stateError
		l.err = &SyntaxError{"EDN does not support namespace-qualified keywords", l.position}
		return lexError
	case r == '/':
		l.state = l.stateError
		l.err = &SyntaxError{"keywords cannot begin with /", l.position}
		return lexError
	case okSymbol(r) || u.IsLetter(r) || ('0' <= r && r <= '9'):
		l.token = tokenKeyword
		l.state = l.stateSym
		return lexCont
	}
	return l.error(r, "after keyword start")
}

// examples: 'foo' 'bar'
// we reuse this from the keyword states, so we don't set token at the end,
// but before we call this
func (l *lexer) stateSym(r rune) lexState {
	switch {
	case okSymbol(r) || u.IsLetter(r) || ('0' <= r && r <= '9'):
		l.state = l.stateSym
		return lexCont
	case r == '/':
		l.state = l.stateSlash
		return lexCont
	}
	return l.stateEndLit(r)
}

// example: 'foo/'
func (l *lexer) stateSlash(r rune) lexState {
	switch {
	case okSymbol(r) || u.IsLetter(r) || ('0' <= r && r <= '9'):
		l.state = l.statePostSlash
		return lexCont
	}
	return l.error(r, "directly after '/' in namespaced symbol")
}

// example : 'foo/bar'
func (l *lexer) statePostSlash(r rune) lexState {
	switch {
	case okSymbol(r) || u.IsLetter(r) || ('0' <= r && r <= '9'):
		l.state = l.statePostSlash
		return lexCont
	}
	return l.stateEndLit(r)
}

// example: '-'
func (l *lexer) stateNeg(r rune) lexState {
	switch {
	case r == '0':
		l.state = l.state0
		return lexCont
	case '1' <= r && r <= '9':
		l.state = l.state1
		return lexCont
	case okSymbol(r) || u.IsLetter(r):
		l.token = tokenSymbol
		l.state = l.stateSym
		return lexCont
	case r == '/':
		l.token = tokenSymbol
		l.state = l.stateSlash
		return lexCont
	}
	l.token = tokenSymbol
	return l.stateEndLit(r)
}

// example: '+'
func (l *lexer) statePos(r rune) lexState {
	switch {
	case r == '0':
		l.state = l.state0
		return lexCont
	case '1' <= r && r <= '9':
		l.state = l.state1
		return lexCont
	case okSymbol(r) || u.IsLetter(r):
		l.token = tokenSymbol
		l.state = l.stateSym
		return lexCont
	case r == '/':
		l.token = tokenSymbol
		l.state = l.stateSlash
		return lexCont
	}
	l.token = tokenSymbol
	return l.stateEndLit(r)
}

// value is '0'
func (l *lexer) state0(r rune) lexState {
	switch {
	case r == '.':
		l.state = l.stateDot
		return lexCont
	case r == 'e' || r == 'E':
		l.state = l.stateE
		return lexCont
	case r == 'M': // bigdecimal
		l.token = tokenFloat
		l.state = l.stateEndLit
		return lexCont // must be ws or delimiter afterwards
	case r == 'N': // bigint
		l.token = tokenInt
		l.state = l.stateEndLit
		return lexCont // must be ws or delimiter afterwards
	}
	l.token = tokenInt
	return l.stateEndLit(r)
}

// anything but a result starting with 0. example '10', '34'
func (l *lexer) state1(r rune) lexState {
	if '0' <= r && r <= '9' {
		return lexCont
	}
	return l.state0(r)
}

// example: '.', can only receive non-numerics here
func (l *lexer) stateDotPre(r rune) lexState {
	switch {
	case okSymbol(r) || u.IsLetter(r):
		l.token = tokenSymbol
		l.state = l.stateSym
		return lexCont
	case r == '/':
		l.token = tokenSymbol
		l.state = l.stateSlash
		return lexCont
	}
	return l.stateEndLit(r)
}

// after reading numeric values plus '.', example: '12.'
func (l *lexer) stateDot(r rune) lexState {
	if '0' <= r && r <= '9' {
		l.state = l.stateDot0
		return lexCont
	}
	// TODO (?): The spec says that there must be numbers after the dot, yet
	// (clojure.edn/read-string "1.e1") returns 10.0
	return l.error(r, "after decimal point in numeric literal")
}

// after reading numeric values plus '.', example: '12.34'
func (l *lexer) stateDot0(r rune) lexState {
	switch {
	case '0' <= r && r <= '9':
		return lexCont
	case r == 'e' || r == 'E':
		l.state = l.stateE
		return lexCont
	case r == 'M':
		l.token = tokenFloat
		l.state = l.stateEndLit
		return lexCont
	}
	l.token = tokenFloat
	return l.stateEndLit(r)
}

// stateE is the state after reading the mantissa and e in a number,
// such as after reading `314e` or `0.314e`.
func (l *lexer) stateE(r rune) lexState {
	if r == '+' || r == '-' {
		l.state = l.stateESign
		return lexCont
	}
	return l.stateESign(r)
}

// stateESign is the state after reading the mantissa, e, and sign in a number,
// such as after reading `314e-` or `0.314e+`.
func (l *lexer) stateESign(r rune) lexState {
	if '0' <= r && r <= '9' {
		l.state = l.stateE0
		return lexCont
	}
	return l.error(r, "in exponent of numeric literal")
}

// stateE0 is the state after reading the mantissa, e, optional sign,
// and at least one digit of the exponent in a number,
// such as after reading `314e-2` or `0.314e+1` or `3.14e0`.
func (l *lexer) stateE0(r rune) lexState {
	if '0' <= r && r <= '9' {
		return lexCont
	}
	if r == 'M' {
		l.token = tokenFloat
		l.state = l.stateEndLit
		return lexCont
	}
	l.token = tokenFloat
	return l.stateEndLit(r)
}

var (
	newlineRunes  = []rune("newline")
	returnRunes   = []rune("return")
	spaceRunes    = []rune("space")
	tabRunes      = []rune("tab")
	formfeedRunes = []rune("formfeed")
)

// stateChar after a backslash ('\')
func (l *lexer) stateChar(r rune) lexState {
	switch {
	// oh my, I'm so happy that none of these share the same prefix.
	case r == 'n':
		l.count = 1
		l.expecting = newlineRunes
		l.state = l.stateSpecialChar
		return lexCont
	case r == 'r':
		l.count = 1
		l.expecting = returnRunes
		l.state = l.stateSpecialChar
		return lexCont
	case r == 's':
		l.count = 1
		l.expecting = spaceRunes
		l.state = l.stateSpecialChar
		return lexCont
	case r == 't':
		l.count = 1
		l.expecting = tabRunes
		l.state = l.stateSpecialChar
		return lexCont
	case r == 'f':
		l.count = 1
		l.expecting = formfeedRunes
		l.state = l.stateSpecialChar
		return lexCont
	case r == 'u':
		l.count = 0
		l.state = l.stateUnicodeChar
		return lexCont
	case isWhitespace(r):
		l.state = l.stateError
		l.err = &SyntaxError{"backslash cannot be followed by whitespace", l.position}
		return lexError
	}
	// default is single name character
	l.token = tokenChar
	l.state = l.stateEndLit
	return lexCont
}

func (l *lexer) stateSpecialChar(r rune) lexState {
	if r == l.expecting[l.count] {
		l.count++
		if l.count == len(l.expecting) {
			l.token = tokenChar
			l.state = l.stateEndLit
			return lexCont
		}
		return lexCont
	}
	if l.count != 1 {
		return l.error(r, "after start of special character")
	}
	// it is likely just a normal character, like 'n' or 't'
	l.token = tokenChar
	return l.stateEndLit(r)
}

func (l *lexer) stateUnicodeChar(r rune) lexState {
	if '0' <= r && r <= '9' || 'a' <= r && r <= 'f' || 'A' <= r && r <= 'F' {
		l.count++
		if l.count == 4 {
			l.token = tokenChar
			l.state = l.stateEndLit
		}
		return lexCont
	}
	if l.count != 0 {
		return l.error(r, "after start of unicode character")
	}
	// likely just '\u'
	l.token = tokenChar
	return l.stateEndLit(r)
}

// stateInString is the state after reading `"`.
func (l *lexer) stateInString(r rune) lexState {
	if r == '"' {
		l.token = tokenString
		return lexEnd
	}
	if r == '\\' {
		l.state = l.stateInStringEsc
		return lexCont
	}
	return lexCont
}

// stateInStringEsc is the state after reading `"\` during a quoted string.
func (l *lexer) stateInStringEsc(r rune) lexState {
	switch r {
	case 'b', 'f', 'n', 'r', 't', '\\', '/', '"':
		l.state = l.stateInString
		return lexCont
	case 'u':
		l.state = l.stateInStringEscU
		l.count = 0
		return lexCont
	}
	return l.error(r, "in string escape code")
}

// stateInStringEscU is the state after reading `"\u` and l.count elements in a
// quoted string.
func (l *lexer) stateInStringEscU(r rune) lexState {
	if '0' <= r && r <= '9' || 'a' <= r && r <= 'f' || 'A' <= r && r <= 'F' {
		l.count++
		if l.count == 4 {
			l.state = l.stateInString
		}
		return lexCont
	}
	// numbers
	return l.error(r, "in \\u hexadecimal character escape")
}

// after reading the character '#'
func (l *lexer) statePound(r rune) lexState {
	switch {
	case r == '_':
		l.token = tokenDiscard
		return lexEnd
	case r == '{':
		l.token = tokenSetStart
		return lexEnd
	case u.IsLetter(r):
		l.token = tokenTag
		l.state = l.stateSym
		return lexCont
	}
	return l.error(r, `after token starting with "#"`)
}

func (l *lexer) stateError(r rune) lexState {
	return lexError
}

// error records an error and switches to the error state.
func (l *lexer) error(r rune, context string) lexState {
	l.state = l.stateError
	l.err = &SyntaxError{"invalid character " + quoteRune(r) + " " + context, l.position}
	return lexError
}

// quoteRune formats r as a quoted rune literal
func quoteRune(r rune) string {
	// special cases - different from quoted strings
	if r == '\'' {
		return `'\''`
	}
	if r == '"' {
		return `'"'`
	}

	// use quoted string with different quotation marks
	s := strconv.Quote(string(r))
	return "'" + s[1:len(s)-1] + "'"
}
