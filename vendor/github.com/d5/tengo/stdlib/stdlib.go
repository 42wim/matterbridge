package stdlib

import "github.com/d5/tengo/objects"

// Modules contain the standard modules.
var Modules = map[string]*objects.Object{
	"math":  objectPtr(&objects.ImmutableMap{Value: mathModule}),
	"os":    objectPtr(&objects.ImmutableMap{Value: osModule}),
	"text":  objectPtr(&objects.ImmutableMap{Value: textModule}),
	"times": objectPtr(&objects.ImmutableMap{Value: timesModule}),
	"rand":  objectPtr(&objects.ImmutableMap{Value: randModule}),
}

func objectPtr(o objects.Object) *objects.Object {
	return &o
}
