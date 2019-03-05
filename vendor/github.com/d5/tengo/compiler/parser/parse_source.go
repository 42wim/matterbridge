package parser

import (
	"io"

	"github.com/d5/tengo/compiler/ast"
	"github.com/d5/tengo/compiler/source"
)

// ParseSource parses source code 'src' and builds an AST.
func ParseSource(filename string, src []byte, trace io.Writer) (res *ast.File, err error) {
	fileSet := source.NewFileSet()
	file := fileSet.AddFile(filename, -1, len(src))

	p := NewParser(file, src, trace)
	return p.ParseFile()
}
