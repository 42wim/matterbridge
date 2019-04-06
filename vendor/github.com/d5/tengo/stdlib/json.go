package stdlib

import (
	"bytes"
	gojson "encoding/json"

	"github.com/d5/tengo/objects"
	"github.com/d5/tengo/stdlib/json"
)

var jsonModule = map[string]objects.Object{
	"decode":      &objects.UserFunction{Name: "decode", Value: jsonDecode},
	"encode":      &objects.UserFunction{Name: "encode", Value: jsonEncode},
	"indent":      &objects.UserFunction{Name: "encode", Value: jsonIndent},
	"html_escape": &objects.UserFunction{Name: "html_escape", Value: jsonHTMLEscape},
}

func jsonDecode(args ...objects.Object) (ret objects.Object, err error) {
	if len(args) != 1 {
		return nil, objects.ErrWrongNumArguments
	}

	switch o := args[0].(type) {
	case *objects.Bytes:
		v, err := json.Decode(o.Value)
		if err != nil {
			return &objects.Error{Value: &objects.String{Value: err.Error()}}, nil
		}
		return v, nil
	case *objects.String:
		v, err := json.Decode([]byte(o.Value))
		if err != nil {
			return &objects.Error{Value: &objects.String{Value: err.Error()}}, nil
		}
		return v, nil
	default:
		return nil, objects.ErrInvalidArgumentType{
			Name:     "first",
			Expected: "bytes/string",
			Found:    args[0].TypeName(),
		}
	}
}

func jsonEncode(args ...objects.Object) (ret objects.Object, err error) {
	if len(args) != 1 {
		return nil, objects.ErrWrongNumArguments
	}

	b, err := json.Encode(args[0])
	if err != nil {
		return &objects.Error{Value: &objects.String{Value: err.Error()}}, nil
	}

	return &objects.Bytes{Value: b}, nil
}

func jsonIndent(args ...objects.Object) (ret objects.Object, err error) {
	if len(args) != 3 {
		return nil, objects.ErrWrongNumArguments
	}

	prefix, ok := objects.ToString(args[1])
	if !ok {
		return nil, objects.ErrInvalidArgumentType{
			Name:     "prefix",
			Expected: "string(compatible)",
			Found:    args[1].TypeName(),
		}
	}

	indent, ok := objects.ToString(args[2])
	if !ok {
		return nil, objects.ErrInvalidArgumentType{
			Name:     "indent",
			Expected: "string(compatible)",
			Found:    args[2].TypeName(),
		}
	}

	switch o := args[0].(type) {
	case *objects.Bytes:
		var dst bytes.Buffer
		err := gojson.Indent(&dst, o.Value, prefix, indent)
		if err != nil {
			return &objects.Error{Value: &objects.String{Value: err.Error()}}, nil
		}
		return &objects.Bytes{Value: dst.Bytes()}, nil
	case *objects.String:
		var dst bytes.Buffer
		err := gojson.Indent(&dst, []byte(o.Value), prefix, indent)
		if err != nil {
			return &objects.Error{Value: &objects.String{Value: err.Error()}}, nil
		}
		return &objects.Bytes{Value: dst.Bytes()}, nil
	default:
		return nil, objects.ErrInvalidArgumentType{
			Name:     "first",
			Expected: "bytes/string",
			Found:    args[0].TypeName(),
		}
	}
}

func jsonHTMLEscape(args ...objects.Object) (ret objects.Object, err error) {
	if len(args) != 1 {
		return nil, objects.ErrWrongNumArguments
	}

	switch o := args[0].(type) {
	case *objects.Bytes:
		var dst bytes.Buffer
		gojson.HTMLEscape(&dst, o.Value)
		return &objects.Bytes{Value: dst.Bytes()}, nil
	case *objects.String:
		var dst bytes.Buffer
		gojson.HTMLEscape(&dst, []byte(o.Value))
		return &objects.Bytes{Value: dst.Bytes()}, nil
	default:
		return nil, objects.ErrInvalidArgumentType{
			Name:     "first",
			Expected: "bytes/string",
			Found:    args[0].TypeName(),
		}
	}
}
