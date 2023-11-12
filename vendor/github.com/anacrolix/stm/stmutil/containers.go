package stmutil

import (
	"unsafe"

	"github.com/benbjohnson/immutable"

	"github.com/anacrolix/missinggo/v2/iter"
)

type Settish interface {
	Add(interface{}) Settish
	Delete(interface{}) Settish
	Contains(interface{}) bool
	Range(func(interface{}) bool)
	iter.Iterable
	Len() int
}

type mapToSet struct {
	m Mappish
}

type interhash struct{}

func (interhash) Hash(x interface{}) uint32 {
	return uint32(nilinterhash(unsafe.Pointer(&x), 0))
}

func (interhash) Equal(i, j interface{}) bool {
	return i == j
}

func NewSet() Settish {
	return mapToSet{NewMap()}
}

func NewSortedSet(lesser lessFunc) Settish {
	return mapToSet{NewSortedMap(lesser)}
}

func (s mapToSet) Add(x interface{}) Settish {
	s.m = s.m.Set(x, nil)
	return s
}

func (s mapToSet) Delete(x interface{}) Settish {
	s.m = s.m.Delete(x)
	return s
}

func (s mapToSet) Len() int {
	return s.m.Len()
}

func (s mapToSet) Contains(x interface{}) bool {
	_, ok := s.m.Get(x)
	return ok
}

func (s mapToSet) Range(f func(interface{}) bool) {
	s.m.Range(func(k, _ interface{}) bool {
		return f(k)
	})
}

func (s mapToSet) Iter(cb iter.Callback) {
	s.Range(cb)
}

type Map struct {
	*immutable.Map
}

func NewMap() Mappish {
	return Map{immutable.NewMap(interhash{})}
}

var _ Mappish = Map{}

func (m Map) Delete(x interface{}) Mappish {
	m.Map = m.Map.Delete(x)
	return m
}

func (m Map) Set(key, value interface{}) Mappish {
	m.Map = m.Map.Set(key, value)
	return m
}

func (sm Map) Range(f func(key, value interface{}) bool) {
	iter := sm.Map.Iterator()
	for !iter.Done() {
		if !f(iter.Next()) {
			return
		}
	}
}

func (sm Map) Iter(cb iter.Callback) {
	sm.Range(func(key, _ interface{}) bool {
		return cb(key)
	})
}

type SortedMap struct {
	*immutable.SortedMap
}

func (sm SortedMap) Set(key, value interface{}) Mappish {
	sm.SortedMap = sm.SortedMap.Set(key, value)
	return sm
}

func (sm SortedMap) Delete(key interface{}) Mappish {
	sm.SortedMap = sm.SortedMap.Delete(key)
	return sm
}

func (sm SortedMap) Range(f func(key, value interface{}) bool) {
	iter := sm.SortedMap.Iterator()
	for !iter.Done() {
		if !f(iter.Next()) {
			return
		}
	}
}

func (sm SortedMap) Iter(cb iter.Callback) {
	sm.Range(func(key, _ interface{}) bool {
		return cb(key)
	})
}

type lessFunc func(l, r interface{}) bool

type comparer struct {
	less lessFunc
}

func (me comparer) Compare(i, j interface{}) int {
	if me.less(i, j) {
		return -1
	} else if me.less(j, i) {
		return 1
	} else {
		return 0
	}
}

func NewSortedMap(less lessFunc) Mappish {
	return SortedMap{
		SortedMap: immutable.NewSortedMap(comparer{less}),
	}
}

type Mappish interface {
	Set(key, value interface{}) Mappish
	Delete(key interface{}) Mappish
	Get(key interface{}) (interface{}, bool)
	Range(func(_, _ interface{}) bool)
	Len() int
	iter.Iterable
}

func GetLeft(l, _ interface{}) interface{} {
	return l
}

//go:noescape
//go:linkname nilinterhash runtime.nilinterhash
func nilinterhash(p unsafe.Pointer, h uintptr) uintptr

func interfaceHash(x interface{}) uint32 {
	return uint32(nilinterhash(unsafe.Pointer(&x), 0))
}

type Lenner interface {
	Len() int
}

type List = *immutable.List
