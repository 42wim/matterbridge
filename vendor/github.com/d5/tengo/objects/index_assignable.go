package objects

// IndexAssignable is an object that can take an index and a value
// on the left-hand side of the assignment statement.
type IndexAssignable interface {
	// IndexSet should take an index Object and a value Object.
	// If an error is returned, it will be treated as a run-time error.
	IndexSet(index, value Object) error
}
