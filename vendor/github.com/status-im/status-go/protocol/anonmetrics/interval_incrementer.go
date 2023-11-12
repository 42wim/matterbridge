package anonmetrics

type FibonacciIntervalIncrementer struct {
	Last    int64
	Current int64
}

func (f *FibonacciIntervalIncrementer) Next() int64 {
	out := f.Last + f.Current

	f.Last = f.Current
	f.Current = out

	return out
}
