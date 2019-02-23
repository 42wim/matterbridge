package objects

func builtinTypeName(args ...Object) (Object, error) {
	if len(args) != 1 {
		return nil, ErrWrongNumArguments
	}

	return &String{Value: args[0].TypeName()}, nil
}
