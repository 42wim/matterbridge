package stdlib

import "github.com/d5/tengo/objects"

// BuiltinModules are builtin type standard library modules.
var BuiltinModules = map[string]map[string]objects.Object{
	"math":  mathModule,
	"os":    osModule,
	"text":  textModule,
	"times": timesModule,
	"rand":  randModule,
	"fmt":   fmtModule,
	"json":  jsonModule,
	"base64": base64Module,
	"hex":  hexModule,
}
