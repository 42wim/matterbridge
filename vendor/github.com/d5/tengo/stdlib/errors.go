package stdlib

import "github.com/d5/tengo/objects"

func wrapError(err error) objects.Object {
	if err == nil {
		return objects.TrueValue
	}

	return &objects.Error{Value: &objects.String{Value: err.Error()}}
}
