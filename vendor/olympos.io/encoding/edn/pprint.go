// Copyright 2015 Jean Niklas L'orange.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package edn

import (
	"bytes"
	"io"
	"unicode/utf8"
)

var (
	// we can't call it spaceBytes not to conflict with decode.go's spaceBytes.
	spaceOutputBytes = []byte(" ")
	commaOutputBytes = []byte(",")
)

func newline(dst io.Writer, prefix, indent string, depth int) {
	dst.Write([]byte{'\n'})
	dst.Write([]byte(prefix))
	for i := 0; i < depth; i++ {
		dst.Write([]byte(indent))
	}
}

// Indent writes to dst an indented form of the EDN-encoded src. Each EDN
// collection begins on a new, indented line beginning with prefix followed by
// one or more copies of indent according to the indentation nesting. The data
// written to dst does not begin with the prefix nor any indentation, and has
// no trailing newline, to make it easier to embed inside other formatted EDN
// data.
//
// Indent filters away whitespace, including comments and discards.
func Indent(dst *bytes.Buffer, src []byte, prefix, indent string) error {
	origLen := dst.Len()
	err := IndentStream(dst, bytes.NewBuffer(src), prefix, indent)
	if err != nil {
		dst.Truncate(origLen)
	}
	return err
}

// IndentStream is an implementation of PPrint for generic readers and writers
func IndentStream(dst io.Writer, src io.Reader, prefix, indent string) error {
	var lex lexer
	lex.reset()
	tokStack := newTokenStack()
	curType := tokenError
	curSize := 0
	d := NewDecoder(src)
	depth := 0
	for {
		bs, tt, err := d.nextToken()
		if err != nil {
			return err
		}
		err = tokStack.push(tt)
		if err != nil {
			return err
		}
		prevType := curType
		prevSize := curSize
		if len(tokStack.toks) > 0 {
			curType = tokStack.peek()
			curSize = tokStack.peekCount()
		}
		switch tt {
		case tokenMapStart, tokenVectorStart, tokenListStart, tokenSetStart:
			if prevType == tokenMapStart {
				dst.Write([]byte{' '})
			} else if depth > 0 {
				newline(dst, prefix, indent, depth)
			}
			dst.Write(bs)
			depth++
		case tokenVectorEnd, tokenListEnd, tokenMapEnd: // tokenSetEnd == tokenMapEnd
			depth--
			if prevSize > 0 { // suppress indent for empty collections
				newline(dst, prefix, indent, depth)
			}
			// all of these are of length 1 in bytes, so utilise this for perf
			dst.Write(bs)
		case tokenTag:
			// need to know what the previous type was.
			switch prevType {
			case tokenMapStart:
				if prevSize%2 == 0 { // If previous size modulo 2 is equal to 0, we're a key
					if prevSize > 0 {
						dst.Write(commaOutputBytes)
					}
					newline(dst, prefix, indent, depth)
				} else { // We're a value, add a space after the key
					dst.Write(spaceOutputBytes)
				}
				dst.Write(bs)
				dst.Write(spaceOutputBytes)
			case tokenSetStart, tokenVectorStart, tokenListStart:
				newline(dst, prefix, indent, depth)
				dst.Write(bs)
				dst.Write(spaceOutputBytes)
			default: // tokenError or nested tag
				dst.Write(bs)
				dst.Write(spaceOutputBytes)
			}
		default:
			switch prevType {
			case tokenMapStart:
				if prevSize%2 == 0 { // If previous size modulo 2 is equal to 0, we're a key
					if prevSize > 0 {
						dst.Write(commaOutputBytes)
					}
					newline(dst, prefix, indent, depth)
				} else { // We're a value, add a space after the key
					dst.Write(spaceOutputBytes)
				}
				dst.Write(bs)
			case tokenSetStart, tokenVectorStart, tokenListStart:
				newline(dst, prefix, indent, depth)
				dst.Write(bs)
			default: // toplevel or nested tag. This should collapse the whole tag tower
				dst.Write(bs)
			}
		}
		if tokStack.done() {
			break
		}
	}
	return nil
}

// PPrintOpts is a configuration map for PPrint. The values in this struct has
// no effect as of now.
type PPrintOpts struct {
	RightMargin int
	MiserWidth  int
}

func pprintIndent(dst io.Writer, shift int) {
	spaces := make([]byte, shift+1)

	spaces[0] = '\n'

	// TODO: This may be slower than caching the size as a byte slice
	for i := 1; i <= shift; i++ {
		spaces[i] = ' '
	}

	dst.Write(spaces)
}

// PPrint writes to dst an indented form of the EDN-encoded src. This
// implementation attempts to write idiomatic/readable EDN values, in a fashion
// close to (but not quite equal to) clojure.pprint/pprint.
//
// PPrint filters away whitespace, including comments and discards.
func PPrint(dst *bytes.Buffer, src []byte, opt *PPrintOpts) error {
	origLen := dst.Len()
	err := PPrintStream(dst, bytes.NewBuffer(src), opt)
	if err != nil {
		dst.Truncate(origLen)
	}
	return err
}

// PPrintStream is an implementation of PPrint for generic readers and writers
func PPrintStream(dst io.Writer, src io.Reader, opt *PPrintOpts) error {
	var lex lexer
	var col, prevCollStart, curSize int
	var prevColl bool

	lex.reset()
	tokStack := newTokenStack()

	shift := make([]int, 1, 8) // pre-allocate some space
	curType := tokenError
	d := NewDecoder(src)

	for {
		bs, tt, err := d.nextToken()
		if err != nil {
			return err
		}
		err = tokStack.push(tt)
		if err != nil {
			return err
		}
		prevType := curType
		prevSize := curSize
		if len(tokStack.toks) > 0 {
			curType = tokStack.peek()
			curSize = tokStack.peekCount()
		}
		// Indentation
		switch tt {
		case tokenVectorEnd, tokenListEnd, tokenMapEnd:
		default:
			switch prevType {
			case tokenMapStart:
				if prevSize%2 == 0 && prevSize > 0 {
					dst.Write(commaOutputBytes)
					pprintIndent(dst, shift[len(shift)-1])
					col = shift[len(shift)-1]
				} else if prevSize%2 == 1 { // We're a value, add a space after the key
					dst.Write(spaceOutputBytes)
					col++
				}
			case tokenSetStart, tokenVectorStart, tokenListStart:
				if prevColl {
					// begin on new line where prevColl started
					// This will look so strange for heterogenous maps.
					pprintIndent(dst, prevCollStart)
					col = prevCollStart
				} else if prevSize > 0 {
					dst.Write(spaceOutputBytes)
					col++
				}
			}
		}
		switch tt {
		case tokenMapStart, tokenVectorStart, tokenListStart, tokenSetStart:
			dst.Write(bs)
			col += len(bs)             // either 2 or 1
			shift = append(shift, col) // we only use maps for now, but we'll utilise this more thoroughly later on
		case tokenVectorEnd, tokenListEnd, tokenMapEnd: // tokenSetEnd == tokenMapEnd
			dst.Write(bs) // all of these are of length 1 in bytes, so this is ok
			prevCollStart = shift[len(shift)-1] - 1
			shift = shift[:len(shift)-1]
		case tokenTag:
			bslen := utf8.RuneCount(bs)
			dst.Write(bs)
			dst.Write(spaceOutputBytes)
			col += bslen + 1
		default:
			bslen := utf8.RuneCount(bs)
			dst.Write(bs)
			col += bslen
		}
		prevColl = (tt == tokenMapEnd || tt == tokenVectorEnd || tt == tokenListEnd)
		if tokStack.done() {
			break
		}
	}
	return nil
}
