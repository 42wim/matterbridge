package objects

import "github.com/d5/tengo"

func builtinString(args ...Object) (Object, error) {
	argsLen := len(args)
	if !(argsLen == 1 || argsLen == 2) {
		return nil, ErrWrongNumArguments
	}

	if _, ok := args[0].(*String); ok {
		return args[0], nil
	}

	v, ok := ToString(args[0])
	if ok {
		if len(v) > tengo.MaxStringLen {
			return nil, ErrStringLimit
		}

		return &String{Value: v}, nil
	}

	if argsLen == 2 {
		return args[1], nil
	}

	return UndefinedValue, nil
}

func builtinInt(args ...Object) (Object, error) {
	argsLen := len(args)
	if !(argsLen == 1 || argsLen == 2) {
		return nil, ErrWrongNumArguments
	}

	if _, ok := args[0].(*Int); ok {
		return args[0], nil
	}

	v, ok := ToInt64(args[0])
	if ok {
		return &Int{Value: v}, nil
	}

	if argsLen == 2 {
		return args[1], nil
	}

	return UndefinedValue, nil
}

func builtinFloat(args ...Object) (Object, error) {
	argsLen := len(args)
	if !(argsLen == 1 || argsLen == 2) {
		return nil, ErrWrongNumArguments
	}

	if _, ok := args[0].(*Float); ok {
		return args[0], nil
	}

	v, ok := ToFloat64(args[0])
	if ok {
		return &Float{Value: v}, nil
	}

	if argsLen == 2 {
		return args[1], nil
	}

	return UndefinedValue, nil
}

func builtinBool(args ...Object) (Object, error) {
	if len(args) != 1 {
		return nil, ErrWrongNumArguments
	}

	if _, ok := args[0].(*Bool); ok {
		return args[0], nil
	}

	v, ok := ToBool(args[0])
	if ok {
		if v {
			return TrueValue, nil
		}

		return FalseValue, nil
	}

	return UndefinedValue, nil
}

func builtinChar(args ...Object) (Object, error) {
	argsLen := len(args)
	if !(argsLen == 1 || argsLen == 2) {
		return nil, ErrWrongNumArguments
	}

	if _, ok := args[0].(*Char); ok {
		return args[0], nil
	}

	v, ok := ToRune(args[0])
	if ok {
		return &Char{Value: v}, nil
	}

	if argsLen == 2 {
		return args[1], nil
	}

	return UndefinedValue, nil
}

func builtinBytes(args ...Object) (Object, error) {
	argsLen := len(args)
	if !(argsLen == 1 || argsLen == 2) {
		return nil, ErrWrongNumArguments
	}

	// bytes(N) => create a new bytes with given size N
	if n, ok := args[0].(*Int); ok {
		if n.Value > int64(tengo.MaxBytesLen) {
			return nil, ErrBytesLimit
		}

		return &Bytes{Value: make([]byte, int(n.Value))}, nil
	}

	v, ok := ToByteSlice(args[0])
	if ok {
		if len(v) > tengo.MaxBytesLen {
			return nil, ErrBytesLimit
		}

		return &Bytes{Value: v}, nil
	}

	if argsLen == 2 {
		return args[1], nil
	}

	return UndefinedValue, nil
}

func builtinTime(args ...Object) (Object, error) {
	argsLen := len(args)
	if !(argsLen == 1 || argsLen == 2) {
		return nil, ErrWrongNumArguments
	}

	if _, ok := args[0].(*Time); ok {
		return args[0], nil
	}

	v, ok := ToTime(args[0])
	if ok {
		return &Time{Value: v}, nil
	}

	if argsLen == 2 {
		return args[1], nil
	}

	return UndefinedValue, nil
}
