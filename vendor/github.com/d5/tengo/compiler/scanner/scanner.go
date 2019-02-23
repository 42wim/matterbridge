/*
	Scanner reads the Tengo source text and tokenize them.

	Scanner is a modified version of Go's scanner implementation.

	Copyright 2009 The Go Authors. All rights reserved.
	Use of this source code is governed by a BSD-style
	license that can be found in the LICENSE file.
*/

package scanner

import (
	"fmt"
	"unicode"
	"unicode/utf8"

	"github.com/d5/tengo/compiler/source"
	"github.com/d5/tengo/compiler/token"
)

// byte order mark
const bom = 0xFEFF

// Scanner reads the Tengo source text.
type Scanner struct {
	file         *source.File // source file handle
	src          []byte       // source
	ch           rune         // current character
	offset       int          // character offset
	readOffset   int          // reading offset (position after current character)
	lineOffset   int          // current line offset
	insertSemi   bool         // insert a semicolon before next newline
	errorHandler ErrorHandler // error reporting; or nil
	errorCount   int          // number of errors encountered
	mode         Mode
}

// NewScanner creates a Scanner.
func NewScanner(file *source.File, src []byte, errorHandler ErrorHandler, mode Mode) *Scanner {
	if file.Size != len(src) {
		panic(fmt.Sprintf("file size (%d) does not match src len (%d)", file.Size, len(src)))
	}

	s := &Scanner{
		file:         file,
		src:          src,
		errorHandler: errorHandler,
		ch:           ' ',
		mode:         mode,
	}

	s.next()
	if s.ch == bom {
		s.next() // ignore BOM at file beginning
	}

	return s
}

// ErrorCount returns the number of errors.
func (s *Scanner) ErrorCount() int {
	return s.errorCount
}

// Scan returns a token, token literal and its position.
func (s *Scanner) Scan() (tok token.Token, literal string, pos source.Pos) {
	s.skipWhitespace()

	pos = s.file.FileSetPos(s.offset)

	insertSemi := false

	// determine token value
	switch ch := s.ch; {
	case isLetter(ch):
		literal = s.scanIdentifier()
		tok = token.Lookup(literal)
		switch tok {
		case token.Ident, token.Break, token.Continue, token.Return, token.Export, token.True, token.False, token.Undefined:
			insertSemi = true
		}
	case '0' <= ch && ch <= '9':
		insertSemi = true
		tok, literal = s.scanNumber(false)
	default:
		s.next() // always make progress

		switch ch {
		case -1: // EOF
			if s.insertSemi {
				s.insertSemi = false // EOF consumed
				return token.Semicolon, "\n", pos
			}
			tok = token.EOF
		case '\n':
			// we only reach here if s.insertSemi was set in the first place
			s.insertSemi = false // newline consumed
			return token.Semicolon, "\n", pos
		case '"':
			insertSemi = true
			tok = token.String
			literal = s.scanString()
		case '\'':
			insertSemi = true
			tok = token.Char
			literal = s.scanRune()
		case '`':
			insertSemi = true
			tok = token.String
			literal = s.scanRawString()
		case ':':
			tok = s.switch2(token.Colon, token.Define)
		case '.':
			if '0' <= s.ch && s.ch <= '9' {
				insertSemi = true
				tok, literal = s.scanNumber(true)
			} else {
				tok = token.Period
				if s.ch == '.' && s.peek() == '.' {
					s.next()
					s.next() // consume last '.'
					tok = token.Ellipsis
				}
			}
		case ',':
			tok = token.Comma
		case '?':
			tok = token.Question
		case ';':
			tok = token.Semicolon
			literal = ";"
		case '(':
			tok = token.LParen
		case ')':
			insertSemi = true
			tok = token.RParen
		case '[':
			tok = token.LBrack
		case ']':
			insertSemi = true
			tok = token.RBrack
		case '{':
			tok = token.LBrace
		case '}':
			insertSemi = true
			tok = token.RBrace
		case '+':
			tok = s.switch3(token.Add, token.AddAssign, '+', token.Inc)
			if tok == token.Inc {
				insertSemi = true
			}
		case '-':
			tok = s.switch3(token.Sub, token.SubAssign, '-', token.Dec)
			if tok == token.Dec {
				insertSemi = true
			}
		case '*':
			tok = s.switch2(token.Mul, token.MulAssign)
		case '/':
			if s.ch == '/' || s.ch == '*' {
				// comment
				if s.insertSemi && s.findLineEnd() {
					// reset position to the beginning of the comment
					s.ch = '/'
					s.offset = s.file.Offset(pos)
					s.readOffset = s.offset + 1
					s.insertSemi = false // newline consumed
					return token.Semicolon, "\n", pos
				}
				comment := s.scanComment()
				if s.mode&ScanComments == 0 {
					// skip comment
					s.insertSemi = false // newline consumed
					return s.Scan()
				}
				tok = token.Comment
				literal = comment
			} else {
				tok = s.switch2(token.Quo, token.QuoAssign)
			}
		case '%':
			tok = s.switch2(token.Rem, token.RemAssign)
		case '^':
			tok = s.switch2(token.Xor, token.XorAssign)
		case '<':
			tok = s.switch4(token.Less, token.LessEq, '<', token.Shl, token.ShlAssign)
		case '>':
			tok = s.switch4(token.Greater, token.GreaterEq, '>', token.Shr, token.ShrAssign)
		case '=':
			tok = s.switch2(token.Assign, token.Equal)
		case '!':
			tok = s.switch2(token.Not, token.NotEqual)
		case '&':
			if s.ch == '^' {
				s.next()
				tok = s.switch2(token.AndNot, token.AndNotAssign)
			} else {
				tok = s.switch3(token.And, token.AndAssign, '&', token.LAnd)
			}
		case '|':
			tok = s.switch3(token.Or, token.OrAssign, '|', token.LOr)
		default:
			// next reports unexpected BOMs - don't repeat
			if ch != bom {
				s.error(s.file.Offset(pos), fmt.Sprintf("illegal character %#U", ch))
			}
			insertSemi = s.insertSemi // preserve insertSemi info
			tok = token.Illegal
			literal = string(ch)
		}
	}

	if s.mode&DontInsertSemis == 0 {
		s.insertSemi = insertSemi
	}

	return
}

func (s *Scanner) next() {
	if s.readOffset < len(s.src) {
		s.offset = s.readOffset
		if s.ch == '\n' {
			s.lineOffset = s.offset
			s.file.AddLine(s.offset)
		}
		r, w := rune(s.src[s.readOffset]), 1
		switch {
		case r == 0:
			s.error(s.offset, "illegal character NUL")
		case r >= utf8.RuneSelf:
			// not ASCII
			r, w = utf8.DecodeRune(s.src[s.readOffset:])
			if r == utf8.RuneError && w == 1 {
				s.error(s.offset, "illegal UTF-8 encoding")
			} else if r == bom && s.offset > 0 {
				s.error(s.offset, "illegal byte order mark")
			}
		}
		s.readOffset += w
		s.ch = r
	} else {
		s.offset = len(s.src)
		if s.ch == '\n' {
			s.lineOffset = s.offset
			s.file.AddLine(s.offset)
		}
		s.ch = -1 // eof
	}
}

func (s *Scanner) peek() byte {
	if s.readOffset < len(s.src) {
		return s.src[s.readOffset]
	}

	return 0
}

func (s *Scanner) error(offset int, msg string) {
	if s.errorHandler != nil {
		s.errorHandler(s.file.Position(s.file.FileSetPos(offset)), msg)
	}

	s.errorCount++
}

func (s *Scanner) scanComment() string {
	// initial '/' already consumed; s.ch == '/' || s.ch == '*'
	offs := s.offset - 1 // position of initial '/'
	var numCR int

	if s.ch == '/' {
		//-style comment
		// (the final '\n' is not considered part of the comment)
		s.next()
		for s.ch != '\n' && s.ch >= 0 {
			if s.ch == '\r' {
				numCR++
			}
			s.next()
		}
		goto exit
	}

	/*-style comment */
	s.next()
	for s.ch >= 0 {
		ch := s.ch
		if ch == '\r' {
			numCR++
		}
		s.next()
		if ch == '*' && s.ch == '/' {
			s.next()
			goto exit
		}
	}

	s.error(offs, "comment not terminated")

exit:
	lit := s.src[offs:s.offset]

	// On Windows, a (//-comment) line may end in "\r\n".
	// Remove the final '\r' before analyzing the text for line directives (matching the compiler).
	// Remove any other '\r' afterwards (matching the pre-existing behavior of the scanner).
	if numCR > 0 && len(lit) >= 2 && lit[1] == '/' && lit[len(lit)-1] == '\r' {
		lit = lit[:len(lit)-1]
		numCR--
	}

	if numCR > 0 {
		lit = StripCR(lit, lit[1] == '*')
	}

	return string(lit)
}

func (s *Scanner) findLineEnd() bool {
	// initial '/' already consumed

	defer func(offs int) {
		// reset scanner state to where it was upon calling findLineEnd
		s.ch = '/'
		s.offset = offs
		s.readOffset = offs + 1
		s.next() // consume initial '/' again
	}(s.offset - 1)

	// read ahead until a newline, EOF, or non-comment tok is found
	for s.ch == '/' || s.ch == '*' {
		if s.ch == '/' {
			//-style comment always contains a newline
			return true
		}
		/*-style comment: look for newline */
		s.next()
		for s.ch >= 0 {
			ch := s.ch
			if ch == '\n' {
				return true
			}
			s.next()
			if ch == '*' && s.ch == '/' {
				s.next()
				break
			}
		}
		s.skipWhitespace() // s.insertSemi is set
		if s.ch < 0 || s.ch == '\n' {
			return true
		}
		if s.ch != '/' {
			// non-comment tok
			return false
		}
		s.next() // consume '/'
	}

	return false
}

func (s *Scanner) scanIdentifier() string {
	offs := s.offset
	for isLetter(s.ch) || isDigit(s.ch) {
		s.next()
	}

	return string(s.src[offs:s.offset])
}

func (s *Scanner) scanMantissa(base int) {
	for digitVal(s.ch) < base {
		s.next()
	}
}

func (s *Scanner) scanNumber(seenDecimalPoint bool) (tok token.Token, lit string) {
	// digitVal(s.ch) < 10
	offs := s.offset
	tok = token.Int

	defer func() {
		lit = string(s.src[offs:s.offset])
	}()

	if seenDecimalPoint {
		offs--
		tok = token.Float
		s.scanMantissa(10)
		goto exponent
	}

	if s.ch == '0' {
		// int or float
		offs := s.offset
		s.next()
		if s.ch == 'x' || s.ch == 'X' {
			// hexadecimal int
			s.next()
			s.scanMantissa(16)
			if s.offset-offs <= 2 {
				// only scanned "0x" or "0X"
				s.error(offs, "illegal hexadecimal number")
			}
		} else {
			// octal int or float
			seenDecimalDigit := false
			s.scanMantissa(8)
			if s.ch == '8' || s.ch == '9' {
				// illegal octal int or float
				seenDecimalDigit = true
				s.scanMantissa(10)
			}
			if s.ch == '.' || s.ch == 'e' || s.ch == 'E' || s.ch == 'i' {
				goto fraction
			}
			// octal int
			if seenDecimalDigit {
				s.error(offs, "illegal octal number")
			}
		}

		return
	}

	// decimal int or float
	s.scanMantissa(10)

fraction:
	if s.ch == '.' {
		tok = token.Float
		s.next()
		s.scanMantissa(10)
	}

exponent:
	if s.ch == 'e' || s.ch == 'E' {
		tok = token.Float
		s.next()
		if s.ch == '-' || s.ch == '+' {
			s.next()
		}
		if digitVal(s.ch) < 10 {
			s.scanMantissa(10)
		} else {
			s.error(offs, "illegal floating-point exponent")
		}
	}

	return
}

func (s *Scanner) scanEscape(quote rune) bool {
	offs := s.offset

	var n int
	var base, max uint32
	switch s.ch {
	case 'a', 'b', 'f', 'n', 'r', 't', 'v', '\\', quote:
		s.next()
		return true
	case '0', '1', '2', '3', '4', '5', '6', '7':
		n, base, max = 3, 8, 255
	case 'x':
		s.next()
		n, base, max = 2, 16, 255
	case 'u':
		s.next()
		n, base, max = 4, 16, unicode.MaxRune
	case 'U':
		s.next()
		n, base, max = 8, 16, unicode.MaxRune
	default:
		msg := "unknown escape sequence"
		if s.ch < 0 {
			msg = "escape sequence not terminated"
		}
		s.error(offs, msg)
		return false
	}

	var x uint32
	for n > 0 {
		d := uint32(digitVal(s.ch))
		if d >= base {
			msg := fmt.Sprintf("illegal character %#U in escape sequence", s.ch)
			if s.ch < 0 {
				msg = "escape sequence not terminated"
			}
			s.error(s.offset, msg)
			return false
		}
		x = x*base + d
		s.next()
		n--
	}

	if x > max || 0xD800 <= x && x < 0xE000 {
		s.error(offs, "escape sequence is invalid Unicode code point")
		return false
	}

	return true
}

func (s *Scanner) scanRune() string {
	offs := s.offset - 1 // '\'' opening already consumed

	valid := true
	n := 0
	for {
		ch := s.ch
		if ch == '\n' || ch < 0 {
			// only report error if we don't have one already
			if valid {
				s.error(offs, "rune literal not terminated")
				valid = false
			}
			break
		}
		s.next()
		if ch == '\'' {
			break
		}
		n++
		if ch == '\\' {
			if !s.scanEscape('\'') {
				valid = false
			}
			// continue to read to closing quote
		}
	}

	if valid && n != 1 {
		s.error(offs, "illegal rune literal")
	}

	return string(s.src[offs:s.offset])
}

func (s *Scanner) scanString() string {
	offs := s.offset - 1 // '"' opening already consumed

	for {
		ch := s.ch
		if ch == '\n' || ch < 0 {
			s.error(offs, "string literal not terminated")
			break
		}
		s.next()
		if ch == '"' {
			break
		}
		if ch == '\\' {
			s.scanEscape('"')
		}
	}

	return string(s.src[offs:s.offset])
}

func (s *Scanner) scanRawString() string {
	offs := s.offset - 1 // '`' opening already consumed

	hasCR := false
	for {
		ch := s.ch
		if ch < 0 {
			s.error(offs, "raw string literal not terminated")
			break
		}

		s.next()

		if ch == '`' {
			break
		}

		if ch == '\r' {
			hasCR = true
		}
	}

	lit := s.src[offs:s.offset]
	if hasCR {
		lit = StripCR(lit, false)
	}

	return string(lit)
}

// StripCR removes carriage return characters.
func StripCR(b []byte, comment bool) []byte {
	c := make([]byte, len(b))

	i := 0
	for j, ch := range b {
		// In a /*-style comment, don't strip \r from *\r/ (incl. sequences of \r from *\r\r...\r/)
		// since the resulting  */ would terminate the comment too early unless the \r is immediately
		// following the opening /* in which case it's ok because /*/ is not closed yet.
		if ch != '\r' || comment && i > len("/*") && c[i-1] == '*' && j+1 < len(b) && b[j+1] == '/' {
			c[i] = ch
			i++
		}
	}

	return c[:i]
}

func (s *Scanner) skipWhitespace() {
	for s.ch == ' ' || s.ch == '\t' || s.ch == '\n' && !s.insertSemi || s.ch == '\r' {
		s.next()
	}
}

func (s *Scanner) switch2(tok0, tok1 token.Token) token.Token {
	if s.ch == '=' {
		s.next()
		return tok1
	}

	return tok0
}

func (s *Scanner) switch3(tok0, tok1 token.Token, ch2 rune, tok2 token.Token) token.Token {
	if s.ch == '=' {
		s.next()
		return tok1
	}

	if s.ch == ch2 {
		s.next()
		return tok2
	}

	return tok0
}

func (s *Scanner) switch4(tok0, tok1 token.Token, ch2 rune, tok2, tok3 token.Token) token.Token {
	if s.ch == '=' {
		s.next()
		return tok1
	}

	if s.ch == ch2 {
		s.next()
		if s.ch == '=' {
			s.next()
			return tok3
		}

		return tok2
	}

	return tok0
}

func isLetter(ch rune) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z' || ch == '_' || ch >= utf8.RuneSelf && unicode.IsLetter(ch)
}

func isDigit(ch rune) bool {
	return '0' <= ch && ch <= '9' || ch >= utf8.RuneSelf && unicode.IsDigit(ch)
}

func digitVal(ch rune) int {
	switch {
	case '0' <= ch && ch <= '9':
		return int(ch - '0')
	case 'a' <= ch && ch <= 'f':
		return int(ch - 'a' + 10)
	case 'A' <= ch && ch <= 'F':
		return int(ch - 'A' + 10)
	}

	return 16 // larger than any legal digit val
}
