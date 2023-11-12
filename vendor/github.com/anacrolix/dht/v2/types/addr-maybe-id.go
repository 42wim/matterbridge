package types

import (
	"fmt"
	"hash/fnv"

	"github.com/anacrolix/multiless"

	"github.com/anacrolix/dht/v2/int160"
	"github.com/anacrolix/dht/v2/krpc"
)

func AddrMaybeIdSliceFromNodeInfoSlice(nis []krpc.NodeInfo) (ret []AddrMaybeId) {
	ret = make([]AddrMaybeId, 0, len(nis))
	for _, ni := range nis {
		id := int160.FromByteArray(ni.ID)
		ret = append(ret, AddrMaybeId{
			Addr: ni.Addr,
			Id:   &id,
		})
	}
	return
}

type AddrMaybeId struct {
	Addr krpc.NodeAddr
	Id   *int160.T
}

func (me AddrMaybeId) TryIntoNodeInfo() *krpc.NodeInfo {
	if me.Id == nil {
		return nil
	}
	return &krpc.NodeInfo{
		ID:   me.Id.AsByteArray(),
		Addr: me.Addr,
	}
}

func (me *AddrMaybeId) FromNodeInfo(ni krpc.NodeInfo) {
	id := int160.FromByteArray(ni.ID)
	*me = AddrMaybeId{
		Addr: ni.Addr,
		Id:   &id,
	}
}

func (me AddrMaybeId) String() string {
	if me.Id == nil {
		return fmt.Sprintf("unknown id at %s", me.Addr)
	} else {
		return fmt.Sprintf("%v at %v", *me.Id, me.Addr)
	}
}

func (l AddrMaybeId) CloserThan(r AddrMaybeId, target int160.T) bool {
	ml := multiless.New().Bool(l.Id == nil, r.Id == nil)
	if l.Id != nil && r.Id != nil {
		ml = ml.Cmp(l.Id.Distance(target).Cmp(r.Id.Distance(target)))
	}
	if !ml.Ok() {
		// We could use maphash, but it wasn't much faster, and requires a seed. A seed would allow
		// us to prevent deterministic handling of addrs for different uses.
		hashString := func(s string) uint64 {
			h := fnv.New64a()
			h.Write([]byte(s))
			return h.Sum64()
		}
		ml = ml.Uint64(hashString(l.Addr.String()), hashString(r.Addr.String()))
	}
	return ml.Less()
}
