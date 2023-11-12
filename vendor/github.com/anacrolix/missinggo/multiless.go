package missinggo

type (
	SameLessFunc func() (same, less bool)
	MultiLess    struct {
		ok   bool
		less bool
	}
)

func (me *MultiLess) Less() bool {
	return me.ok && me.less
}

func (me *MultiLess) Final() bool {
	if !me.ok {
		panic("undetermined")
	}
	return me.less
}

func (me *MultiLess) FinalOk() (left, ok bool) {
	return me.less, me.ok
}

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

func (me *MultiLess) StrictNext(same, less bool) {
	if me.ok {
		return
	}
	me.Next(func() (bool, bool) { return same, less })
}

func (me *MultiLess) NextBool(l, r bool) {
	me.StrictNext(l == r, l)
}
