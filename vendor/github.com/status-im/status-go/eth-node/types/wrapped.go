package types

// Wrapped tells that a given object has an underlying representation
// and this representation can be accessed using `Unwrap` method.
type Wrapped interface {
	Unwrap() interface{}
}
