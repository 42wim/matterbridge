package containers

import (
	"github.com/anacrolix/dht/v2/int160"
	"github.com/anacrolix/dht/v2/types"
	"github.com/anacrolix/missinggo/v2/iter"
	"github.com/anacrolix/stm/stmutil"
)

type addrMaybeId = types.AddrMaybeId

type AddrMaybeIdsByDistance interface {
	Add(addrMaybeId) AddrMaybeIdsByDistance
	Next() addrMaybeId
	Delete(addrMaybeId) AddrMaybeIdsByDistance
	Len() int
}

type stmSettishWrapper struct {
	set stmutil.Settish
}

func (me stmSettishWrapper) Next() addrMaybeId {
	first, _ := iter.First(me.set.Iter)
	return first.(addrMaybeId)
}

func (me stmSettishWrapper) Delete(x addrMaybeId) AddrMaybeIdsByDistance {
	return stmSettishWrapper{me.set.Delete(x)}
}

func (me stmSettishWrapper) Len() int {
	return me.set.Len()
}

func (me stmSettishWrapper) Add(x addrMaybeId) AddrMaybeIdsByDistance {
	return stmSettishWrapper{me.set.Add(x)}
}

func NewImmutableAddrMaybeIdsByDistance(target int160.T) AddrMaybeIdsByDistance {
	return stmSettishWrapper{stmutil.NewSortedSet(func(l, r interface{}) bool {
		return l.(addrMaybeId).CloserThan(r.(addrMaybeId), target)
	})}
}
