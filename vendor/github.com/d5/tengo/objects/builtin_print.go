package objects

import (
	"fmt"

	"github.com/d5/tengo"
)

// print(args...)
func builtinPrint(args ...Object) (Object, error) {
	for _, arg := range args {
		if str, ok := arg.(*String); ok {
			fmt.Println(str.Value)
		} else {
			fmt.Println(arg.String())
		}
	}

	return nil, nil
}

// printf("format", args...)
func builtinPrintf(args ...Object) (Object, error) {
	numArgs := len(args)
	if numArgs == 0 {
		return nil, ErrWrongNumArguments
	}

	format, ok := args[0].(*String)
	if !ok {
		return nil, ErrInvalidArgumentType{
			Name:     "format",
			Expected: "string",
			Found:    args[0].TypeName(),
		}
	}
	if numArgs == 1 {
		fmt.Print(format)
		return nil, nil
	}

	formatArgs := make([]interface{}, numArgs-1, numArgs-1)
	for idx, arg := range args[1:] {
		formatArgs[idx] = objectToInterface(arg)
	}

	fmt.Printf(format.Value, formatArgs...)

	return nil, nil
}

// sprintf("format", args...)
func builtinSprintf(args ...Object) (Object, error) {
	numArgs := len(args)
	if numArgs == 0 {
		return nil, ErrWrongNumArguments
	}

	format, ok := args[0].(*String)
	if !ok {
		return nil, ErrInvalidArgumentType{
			Name:     "format",
			Expected: "string",
			Found:    args[0].TypeName(),
		}
	}
	if numArgs == 1 {
		return format, nil // okay to return 'format' directly as String is immutable
	}

	formatArgs := make([]interface{}, numArgs-1, numArgs-1)
	for idx, arg := range args[1:] {
		formatArgs[idx] = objectToInterface(arg)
	}

	s := fmt.Sprintf(format.Value, formatArgs...)

	if len(s) > tengo.MaxStringLen {
		return nil, ErrStringLimit
	}

	return &String{Value: s}, nil
}
