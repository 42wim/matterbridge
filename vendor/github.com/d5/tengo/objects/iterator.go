package objects

// Iterator represents an iterator for underlying data type.
type Iterator interface {
	Object

	// Next returns true if there are more elements to iterate.
	Next() bool

	// Key returns the key or index value of the current element.
	Key() Object

	// Value returns the value of the current element.
	Value() Object
}
