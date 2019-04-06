package stdlib

import (
	"fmt"

	"github.com/d5/tengo"
	"github.com/d5/tengo/objects"
)

var fmtModule = map[string]objects.Object{
	"print":   &objects.UserFunction{Name: "print", Value: fmtPrint},
	"printf":  &objects.UserFunction{Name: "printf", Value: fmtPrintf},
	"println": &objects.UserFunction{Name: "println", Value: fmtPrintln},
	"sprintf": &objects.UserFunction{Name: "sprintf", Value: fmtSprintf},
}

func fmtPrint(args ...objects.Object) (ret objects.Object, err error) {
	printArgs, err := getPrintArgs(args...)
	if err != nil {
		return nil, err
	}

	_, _ = fmt.Print(printArgs...)

	return nil, nil
}

func fmtPrintf(args ...objects.Object) (ret objects.Object, err error) {
	numArgs := len(args)
	if numArgs == 0 {
		return nil, objects.ErrWrongNumArguments
	}

	format, ok := args[0].(*objects.String)
	if !ok {
		return nil, objects.ErrInvalidArgumentType{
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
		formatArgs[idx] = objects.ToInterface(arg)
	}

	fmt.Printf(format.Value, formatArgs...)

	return nil, nil
}

func fmtPrintln(args ...objects.Object) (ret objects.Object, err error) {
	printArgs, err := getPrintArgs(args...)
	if err != nil {
		return nil, err
	}

	printArgs = append(printArgs, "\n")
	_, _ = fmt.Print(printArgs...)

	return nil, nil
}

func fmtSprintf(args ...objects.Object) (ret objects.Object, err error) {
	numArgs := len(args)
	if numArgs == 0 {
		return nil, objects.ErrWrongNumArguments
	}

	format, ok := args[0].(*objects.String)
	if !ok {
		return nil, objects.ErrInvalidArgumentType{
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
		formatArgs[idx] = objects.ToInterface(arg)
	}

	s := fmt.Sprintf(format.Value, formatArgs...)

	if len(s) > tengo.MaxStringLen {
		return nil, objects.ErrStringLimit
	}

	return &objects.String{Value: s}, nil
}

func getPrintArgs(args ...objects.Object) ([]interface{}, error) {
	var printArgs []interface{}
	l := 0
	for _, arg := range args {
		s, _ := objects.ToString(arg)
		slen := len(s)
		if l+slen > tengo.MaxStringLen { // make sure length does not exceed the limit
			return nil, objects.ErrStringLimit
		}
		l += slen

		printArgs = append(printArgs, s)
	}

	return printArgs, nil
}
