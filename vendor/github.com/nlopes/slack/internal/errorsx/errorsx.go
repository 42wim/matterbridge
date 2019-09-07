package errorsx

// String representing an error, useful for declaring string constants as errors.
type String string

func (t String) Error() string {
	return string(t)
}
