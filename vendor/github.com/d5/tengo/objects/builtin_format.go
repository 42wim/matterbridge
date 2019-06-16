package objects

func builtinFormat(args ...Object) (Object, error) {
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

	s, err := Format(format.Value, args[1:]...)
	if err != nil {
		return nil, err
	}

	return &String{Value: s}, nil
}
