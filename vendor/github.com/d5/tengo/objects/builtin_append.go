package objects

// append(arr, items...)
func builtinAppend(args ...Object) (Object, error) {
	if len(args) < 2 {
		return nil, ErrWrongNumArguments
	}

	switch arg := args[0].(type) {
	case *Array:
		return &Array{Value: append(arg.Value, args[1:]...)}, nil
	case *ImmutableArray:
		return &Array{Value: append(arg.Value, args[1:]...)}, nil
	default:
		return nil, ErrInvalidArgumentType{
			Name:     "first",
			Expected: "array",
			Found:    arg.TypeName(),
		}
	}
}
