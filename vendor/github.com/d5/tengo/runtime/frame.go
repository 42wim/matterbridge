package runtime

import (
	"github.com/d5/tengo/objects"
)

// Frame represents a function call frame.
type Frame struct {
	fn          *objects.CompiledFunction
	freeVars    []*objects.ObjectPtr
	ip          int
	basePointer int
}
