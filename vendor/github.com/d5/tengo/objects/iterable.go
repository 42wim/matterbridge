package objects

// Iterable represents an object that has iterator.
type Iterable interface {
	// Iterate should return an Iterator for the type.
	Iterate() Iterator
}
