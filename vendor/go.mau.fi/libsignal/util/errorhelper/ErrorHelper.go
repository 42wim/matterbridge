package errorhelper

// NewMultiError returns a new MultiError object.
func NewMultiError() *MultiError {
	return &MultiError{
		errors: []error{},
	}
}

// MultiError is a structure for holding multiple errors so they
// can be checked at a later point.
type MultiError struct {
	errors []error
}

// Add will add the given error if it is not nil.
func (m *MultiError) Add(err error) {
	if err != nil {
		m.errors = append(m.errors, err)
	}
}

// HasErrors will return true if any non-nil errors have been
// added.
func (m *MultiError) HasErrors() bool {
	if len(m.errors) > 0 {
		return true
	}

	return false
}

// Error will print the first error is encountered.
func (m *MultiError) Error() string {
	if !m.HasErrors() {
		return ""
	}

	return m.errors[0].Error()
}
