package objects

import "github.com/d5/tengo/compiler/token"

// Object represents an object in the VM.
type Object interface {
	// TypeName should return the name of the type.
	TypeName() string

	// String should return a string representation of the type's value.
	String() string

	// BinaryOp should return another object that is the result of
	// a given binary operator and a right-hand side object.
	// If BinaryOp returns an error, the VM will treat it as a run-time error.
	BinaryOp(op token.Token, rhs Object) (Object, error)

	// IsFalsy should return true if the value of the type
	// should be considered as falsy.
	IsFalsy() bool

	// Equals should return true if the value of the type
	// should be considered as equal to the value of another object.
	Equals(another Object) bool

	// Copy should return a copy of the type (and its value).
	// Copy function will be used for copy() builtin function
	// which is expected to deep-copy the values generally.
	Copy() Object
}
