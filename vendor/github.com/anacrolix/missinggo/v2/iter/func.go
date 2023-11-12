package iter

// Callback receives a value and returns true if another value should be
// received or false to stop iteration.
type Callback func(value interface{}) (more bool)

// Func iterates by calling Callback for each of its values.
type Func func(Callback)

func All(cb Callback, fs ...Func) bool {
	for _, f := range fs {
		all := true
		f(func(v interface{}) bool {
			all = all && cb(v)
			return all
		})
		if !all {
			return false
		}
	}
	return true
}

// Calls `cb` with the first value yielded by `f` and then stops iteration. `ok` if `cb` was called
// with a value. Returning the value interface{} would require the caller to keep a
func First(f Func) (value interface{}, ok bool) {
	f(func(x interface{}) bool {
		value = x
		ok = true
		return false
	})
	return
}
