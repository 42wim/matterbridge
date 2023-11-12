package dht

import (
	"time"

	"github.com/anacrolix/chansync"

	"github.com/anacrolix/dht/v2/int160"
)

type bucket struct {
	// Per the "Routing Table" section of BEP 5.
	changed     chansync.BroadcastCond
	lastChanged time.Time
	nodes       map[*node]struct{}
}

func (b *bucket) Len() int {
	return len(b.nodes)
}

func (b *bucket) EachNode(f func(*node) bool) bool {
	for n := range b.nodes {
		if !f(n) {
			return false
		}
	}
	return true
}

func (b *bucket) AddNode(n *node, k int) {
	if _, ok := b.nodes[n]; ok {
		return
	}
	if b.nodes == nil {
		b.nodes = make(map[*node]struct{}, k)
	}
	b.nodes[n] = struct{}{}
	b.lastChanged = time.Now()
	b.changed.Broadcast()
}

func (b *bucket) GetNode(addr Addr, id int160.T) *node {
	for n := range b.nodes {
		if n.hasAddrAndID(addr, id) {
			return n
		}
	}
	return nil
}
