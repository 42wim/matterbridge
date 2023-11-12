package discv5

import (
	"context"
	"sync"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/waku-org/go-waku/waku/v2/service"
)

// TestPeerDiscoverer is mock peer discoverer for testing
type TestPeerDiscoverer struct {
	sync.RWMutex
	peerMap map[peer.ID]struct{}
}

// NewTestPeerDiscoverer is a constructor for TestPeerDiscoverer
func NewTestPeerDiscoverer() *TestPeerDiscoverer {
	result := &TestPeerDiscoverer{
		peerMap: make(map[peer.ID]struct{}),
	}

	return result
}

// Subscribe is for subscribing to peer discoverer
func (t *TestPeerDiscoverer) Subscribe(ctx context.Context, ch <-chan service.PeerData) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case p := <-ch:
				t.Lock()
				t.peerMap[p.AddrInfo.ID] = struct{}{}
				t.Unlock()
			}
		}
	}()
}

// HasPeer is for checking if a peer is present in peer discoverer
func (t *TestPeerDiscoverer) HasPeer(p peer.ID) bool {
	t.RLock()
	defer t.RUnlock()
	_, ok := t.peerMap[p]
	return ok
}

// PeerCount is for getting the number of peers in peer discoverer
func (t *TestPeerDiscoverer) PeerCount() int {
	t.RLock()
	defer t.RUnlock()
	return len(t.peerMap)
}

// Clear is for clearing the peer discoverer
func (t *TestPeerDiscoverer) Clear() {
	t.Lock()
	defer t.Unlock()
	t.peerMap = make(map[peer.ID]struct{})
}
