package stdlib

import (
	"encoding/hex"
	"github.com/d5/tengo/objects"
)

var hexModule = map[string]objects.Object{
	"encode": &objects.UserFunction{Value: FuncAYRS(hex.EncodeToString)},
	"decode": &objects.UserFunction{Value: FuncASRYE(hex.DecodeString)},
}
