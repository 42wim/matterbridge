package common

func NewAndSet[T any](v T) *T {
	res := new(T)
	*res = v
	return res
}
