package script

import (
	"github.com/d5/tengo/objects"
)

func objectToInterface(o objects.Object) interface{} {
	switch val := o.(type) {
	case *objects.Array:
		return val.Value
	case *objects.Map:
		return val.Value
	case *objects.Int:
		return val.Value
	case *objects.Float:
		return val.Value
	case *objects.Bool:
		if val == objects.TrueValue {
			return true
		}
		return false
	case *objects.Char:
		return val.Value
	case *objects.String:
		return val.Value
	case *objects.Bytes:
		return val.Value
	case *objects.Undefined:
		return nil
	}

	return o
}
