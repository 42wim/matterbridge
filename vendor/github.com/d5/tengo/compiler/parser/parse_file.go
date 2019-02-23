package parser

import (
	"io"

	"github.com/d5/tengo/compiler/ast"
	"github.com/d5/tengo/compiler/source"
)

// ParseFile parses a file with a given src.
func ParseFile(file *source.File, src []byte, trace io.Writer) (res *ast.File, err error) {
	p := NewParser(file, src, trace)

	defer func() {
		if e := recover(); e != nil {
			if _, ok := e.(bailout); !ok {
				panic(e)
			}
		}

		p.errors.Sort()
		err = p.errors.Err()
	}()

	res, err = p.ParseFile()

	return
}
