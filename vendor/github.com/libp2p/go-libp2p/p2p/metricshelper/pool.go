package metricshelper

import (
	"fmt"
	"sync"
)

const capacity = 8

var stringPool = sync.Pool{New: func() any {
	s := make([]string, 0, capacity)
	return &s
}}

func GetStringSlice() *[]string {
	s := stringPool.Get().(*[]string)
	*s = (*s)[:0]
	return s
}

func PutStringSlice(s *[]string) {
	if c := cap(*s); c < capacity {
		panic(fmt.Sprintf("expected a string slice with capacity 8 or greater, got %d", c))
	}
	stringPool.Put(s)
}
