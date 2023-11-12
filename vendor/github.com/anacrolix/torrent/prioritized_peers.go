package torrent

import (
	"hash/maphash"

	"github.com/anacrolix/multiless"
	"github.com/google/btree"
)

// Peers are stored with their priority at insertion. Their priority may
// change if our apparent IP changes, we don't currently handle that.
type prioritizedPeersItem struct {
	prio peerPriority
	p    PeerInfo
}

var hashSeed = maphash.MakeSeed()

func (me prioritizedPeersItem) addrHash() int64 {
	var h maphash.Hash
	h.SetSeed(hashSeed)
	h.WriteString(me.p.Addr.String())
	return int64(h.Sum64())
}

func (me prioritizedPeersItem) Less(than btree.Item) bool {
	other := than.(prioritizedPeersItem)
	return multiless.New().Bool(
		me.p.Trusted, other.p.Trusted).Uint32(
		me.prio, other.prio).Int64(
		me.addrHash(), other.addrHash(),
	).Less()
}

type prioritizedPeers struct {
	om      *btree.BTree
	getPrio func(PeerInfo) peerPriority
}

func (me *prioritizedPeers) Each(f func(PeerInfo)) {
	me.om.Ascend(func(i btree.Item) bool {
		f(i.(prioritizedPeersItem).p)
		return true
	})
}

func (me *prioritizedPeers) Len() int {
	return me.om.Len()
}

// Returns true if a peer is replaced.
func (me *prioritizedPeers) Add(p PeerInfo) bool {
	return me.om.ReplaceOrInsert(prioritizedPeersItem{me.getPrio(p), p}) != nil
}

// Returns true if a peer is replaced.
func (me *prioritizedPeers) AddReturningReplacedPeer(p PeerInfo) (ret PeerInfo, ok bool) {
	item := me.om.ReplaceOrInsert(prioritizedPeersItem{me.getPrio(p), p})
	if item == nil {
		return
	}
	ret = item.(prioritizedPeersItem).p
	ok = true
	return
}

func (me *prioritizedPeers) DeleteMin() (ret prioritizedPeersItem, ok bool) {
	i := me.om.DeleteMin()
	if i == nil {
		return
	}
	ret = i.(prioritizedPeersItem)
	ok = true
	return
}

func (me *prioritizedPeers) PopMax() PeerInfo {
	return me.om.DeleteMax().(prioritizedPeersItem).p
}
