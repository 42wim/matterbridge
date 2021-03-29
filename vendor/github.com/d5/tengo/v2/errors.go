package tengo

import (
	"errors"
	"fmt"
)

var (
	// ErrStackOverflow is a stack overflow error.
	ErrStackOverflow = errors.New("stack overflow")

	// ErrObjectAllocLimit is an objects allocation limit error.
	ErrObjectAllocLimit = errors.New("object allocation limit exceeded")

	// ErrIndexOutOfBounds is an error where a given index is out of the
	// bounds.
	ErrIndexOutOfBounds = errors.New("index out of bounds")

	// ErrInvalidIndexType represents an invalid index type.
	ErrInvalidIndexType = errors.New("invalid index type")

	// ErrInvalidIndexValueType represents an invalid index value type.
	ErrInvalidIndexValueType = errors.New("invalid index value type")

	// ErrInvalidIndexOnError represents an invalid index on error.
	ErrInvalidIndexOnError = errors.New("invalid index on error")

	// ErrInvalidOperator represents an error for invalid operator usage.
	ErrInvalidOperator = errors.New("invalid operator")

	// ErrWrongNumArguments represents a wrong number of arguments error.
	ErrWrongNumArguments = errors.New("wrong number of arguments")

	// ErrBytesLimit represents an error where the size of bytes value exceeds
	// the limit.
	ErrBytesLimit = errors.New("exceeding bytes size limit")

	// ErrStringLimit represents an error where the size of string value
	// exceeds the limit.
	ErrStringLimit = errors.New("exceeding string size limit")

	// ErrNotIndexable is an error where an Object is not indexable.
	ErrNotIndexable = errors.New("not indexable")

	// ErrNotIndexAssignable is an error where an Object is not index
	// assignable.
	ErrNotIndexAssignable = errors.New("not index-assignable")

	// ErrNotImplemented is an error where an Object has not implemented a
	// required method.
	ErrNotImplemented = errors.New("not implemented")

	// ErrInvalidRangeStep is an error where the step parameter is less than or equal to 0 when using builtin range function.
	ErrInvalidRangeStep = errors.New("range step must be greater than 0")
)

// ErrInvalidArgumentType represents an invalid argument value type error.
type ErrInvalidArgumentType struct {
	Name     string
	Expected string
	Found    string
}

func (e ErrInvalidArgumentType) Error() string {
	return fmt.Sprintf("invalid type for argument '%s': expected %s, found %s",
		e.Name, e.Expected, e.Found)
}
