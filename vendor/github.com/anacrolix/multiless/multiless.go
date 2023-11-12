package multiless

type (
	// A helper for long chains of "less-than" comparisons, where later comparisons are only
	// required if earlier ones haven't resolved the comparison.
	Computation struct {
		ok   bool
		less bool
	}
)

func New() Computation {
	return Computation{}
}

func (me Computation) EagerSameLess(same, less bool) Computation {
	if me.ok || same {
		return me
	}
	return Computation{
		ok:   true,
		less: less,
	}
}

func (me Computation) LazySameLess(lazy func() (same, less bool)) Computation {
	if me.ok {
		return me
	}
	same, less := lazy()
	if !same {
		me.less = less
	}
	return me
}

// Sorts so that false comes before true.
func (me Computation) Bool(l, r bool) Computation {
	return me.EagerSameLess(l == r, r)
}

func (me Computation) Uint32(l, r uint32) Computation {
	return me.EagerSameLess(l == r, l < r)
}

func (me Computation) Int64(l, r int64) Computation {
	return me.EagerSameLess(l == r, l < r)
}

func (me Computation) Uint64(l, r uint64) Computation {
	return me.EagerSameLess(l == r, l < r)
}

func (me Computation) Int(l, r int) Computation {
	return me.EagerSameLess(l == r, l < r)
}

func (me Computation) CmpInt64(i int64) Computation {
	return me.EagerSameLess(i == 0, i < 0)
}

func (me Computation) Cmp(i int) Computation {
	return me.EagerSameLess(i == 0, i < 0)
}

func (me Computation) Uintptr(l, r uintptr) Computation {
	return me.EagerSameLess(l == r, l < r)
}

func (me Computation) Less() bool {
	return me.less
}

func (me Computation) Ok() bool {
	return me.ok
}

func (me Computation) LessOk() (less, ok bool) {
	return me.less, me.ok
}

func (me Computation) MustLess() bool {
	less, ok := me.LessOk()
	if !ok {
		panic("computation has not differentiated yet")
	}
	return less
}

func (me Computation) Float64(l, r float64) Computation {
	return me.EagerSameLess(l == r, l < r)
}

func (me Computation) Lazy(f func() Computation) Computation {
	if me.ok {
		return me
	}
	return f()
}

func (me Computation) AndThen(then Computation) Computation {
	if me.ok {
		return me
	} else {
		return then
	}
}
