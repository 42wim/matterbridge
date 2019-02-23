package objects

var (
	// TrueValue represents a true value.
	TrueValue Object = &Bool{value: true}

	// FalseValue represents a false value.
	FalseValue Object = &Bool{value: false}

	// UndefinedValue represents an undefined value.
	UndefinedValue Object = &Undefined{}
)
