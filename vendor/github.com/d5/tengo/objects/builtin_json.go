package objects

import (
	"encoding/json"
)

// to_json(v object) => bytes
func builtinToJSON(args ...Object) (Object, error) {
	if len(args) != 1 {
		return nil, ErrWrongNumArguments
	}

	res, err := json.Marshal(objectToInterface(args[0]))
	if err != nil {
		return &Error{Value: &String{Value: err.Error()}}, nil
	}

	return &Bytes{Value: res}, nil
}

// from_json(data string/bytes) => object
func builtinFromJSON(args ...Object) (Object, error) {
	if len(args) != 1 {
		return nil, ErrWrongNumArguments
	}

	var target interface{}

	switch o := args[0].(type) {
	case *Bytes:
		err := json.Unmarshal(o.Value, &target)
		if err != nil {
			return &Error{Value: &String{Value: err.Error()}}, nil
		}
	case *String:
		err := json.Unmarshal([]byte(o.Value), &target)
		if err != nil {
			return &Error{Value: &String{Value: err.Error()}}, nil
		}
	default:
		return nil, ErrInvalidArgumentType{
			Name:     "first",
			Expected: "bytes/string",
			Found:    args[0].TypeName(),
		}
	}

	res, err := FromInterface(target)
	if err != nil {
		return nil, err
	}

	return res, nil
}
