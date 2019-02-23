package compiler

// Loop represents a loop construct that
// the compiler uses to track the current loop.
type Loop struct {
	Continues []int
	Breaks    []int
}
