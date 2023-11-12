package missinggo

type (
	// A function that returns equality, and less than. Used for lazily evaluating arguments.
	SameLessFunc func() (same, less bool)
	// A helper for long chains of "less-than" comparisons, where later comparisons are only
	// required if earlier ones haven't resolved the comparison.
	MultiLess struct {
		ok   bool
		less bool
	}
)

// True iff the left is less than the right. Will return false if they're equal, or unresolved.
// (Which is okay in certain circumstances.)
func (me *MultiLess) Less() bool {
	return me.ok && me.less
}

// Returns the result of the less-than comparison chains. Panics if the case was not resolved.
func (me *MultiLess) Final() bool {
	if !me.ok {
		panic("undetermined")
	}
	return me.less
}

// Returns less-than, and whether the comparison was definitely resolved.
func (me *MultiLess) FinalOk() (left, ok bool) {
	return me.less, me.ok
}

// `f` is only evaluated if the result is not yet determined.
func (me *MultiLess) Next(f SameLessFunc) {
	if me.ok {
		return
	}
	same, less := f()
	if same {
		return
	}
	me.ok = true
	me.less = less
}

// Like Next, but the arguments are already evaluated.
func (me *MultiLess) StrictNext(same, less bool) {
	me.Next(func() (bool, bool) { return same, less })
}

// Compare booleans, where the lesser is the true one, if the other is false.
func (me *MultiLess) NextBool(l, r bool) {
	me.StrictNext(l == r, l)
}

// Next use a common comparison result, where < 0 is less and 0 is equal.
func (me *MultiLess) Compare(i int) {
	me.StrictNext(i == 0, i < 0)
}
