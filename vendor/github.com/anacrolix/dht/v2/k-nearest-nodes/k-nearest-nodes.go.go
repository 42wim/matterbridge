package k_nearest_nodes

import (
	"hash/maphash"

	"github.com/anacrolix/multiless"
	"github.com/benbjohnson/immutable"

	"github.com/anacrolix/dht/v2/int160"
	"github.com/anacrolix/dht/v2/krpc"
)

type Key = krpc.NodeInfo

type Elem struct {
	Key
	Data interface{}
}

type Type struct {
	inner *immutable.SortedMap
	k     int
}

func New(target int160.T, k int) Type {
	seed := maphash.MakeSeed()
	return Type{
		k: k,
		inner: immutable.NewSortedMap(lessComparer{less: func(_l, _r interface{}) bool {
			l := _l.(Key)
			r := _r.(Key)
			return multiless.New().Cmp(
				l.ID.Int160().Distance(target).Cmp(r.ID.Int160().Distance(target)),
			).Lazy(func() multiless.Computation {
				var lh, rh maphash.Hash
				lh.SetSeed(seed)
				rh.SetSeed(seed)
				lh.WriteString(l.Addr.String())
				rh.WriteString(r.Addr.String())
				return multiless.New().Int64(int64(lh.Sum64()), int64(rh.Sum64()))
			}).Less()
		}}),
	}
}

func (me *Type) Range(f func(Elem)) {
	iter := me.inner.Iterator()
	for !iter.Done() {
		key, value := iter.Next()
		f(Elem{
			Key:  key.(Key),
			Data: value,
		})
	}
}

func (me Type) Len() int {
	return me.inner.Len()
}

func (me Type) Push(elem Elem) Type {
	me.inner = me.inner.Set(elem.Key, elem.Data)
	for me.inner.Len() > me.k {
		iter := me.inner.Iterator()
		iter.Last()
		key, _ := iter.Next()
		me.inner = me.inner.Delete(key)
	}
	return me
}

func (me Type) Farthest() (elem Elem) {
	iter := me.inner.Iterator()
	iter.Last()
	if iter.Done() {
		panic(me.k)
	}
	key, value := iter.Next()
	return Elem{
		Key:  key.(Key),
		Data: value,
	}
}

func (me Type) Full() bool {
	return me.Len() >= me.k
}

type lessFunc func(l, r interface{}) bool

type lessComparer struct {
	less lessFunc
}

func (me lessComparer) Compare(i, j interface{}) int {
	if me.less(i, j) {
		return -1
	} else if me.less(j, i) {
		return 1
	} else {
		return 0
	}
}
