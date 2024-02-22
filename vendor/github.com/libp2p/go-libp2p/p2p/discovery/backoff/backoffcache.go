package backoff

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p/core/discovery"
	"github.com/libp2p/go-libp2p/core/peer"

	ma "github.com/multiformats/go-multiaddr"
)

// BackoffDiscovery is an implementation of discovery that caches peer data and attenuates repeated queries
type BackoffDiscovery struct {
	disc         discovery.Discovery
	stratFactory BackoffFactory
	peerCache    map[string]*backoffCache
	peerCacheMux sync.RWMutex

	parallelBufSz int
	returnedBufSz int

	clock clock
}

type BackoffDiscoveryOption func(*BackoffDiscovery) error

func NewBackoffDiscovery(disc discovery.Discovery, stratFactory BackoffFactory, opts ...BackoffDiscoveryOption) (discovery.Discovery, error) {
	b := &BackoffDiscovery{
		disc:         disc,
		stratFactory: stratFactory,
		peerCache:    make(map[string]*backoffCache),

		parallelBufSz: 32,
		returnedBufSz: 32,

		clock: realClock{},
	}

	for _, opt := range opts {
		if err := opt(b); err != nil {
			return nil, err
		}
	}

	return b, nil
}

// WithBackoffDiscoverySimultaneousQueryBufferSize sets the buffer size for the channels between the main FindPeers query
// for a given namespace and all simultaneous FindPeers queries for the namespace
func WithBackoffDiscoverySimultaneousQueryBufferSize(size int) BackoffDiscoveryOption {
	return func(b *BackoffDiscovery) error {
		if size < 0 {
			return fmt.Errorf("cannot set size to be smaller than 0")
		}
		b.parallelBufSz = size
		return nil
	}
}

// WithBackoffDiscoveryReturnedChannelSize sets the size of the buffer to be used during a FindPeer query.
// Note: This does not apply if the query occurs during the backoff time
func WithBackoffDiscoveryReturnedChannelSize(size int) BackoffDiscoveryOption {
	return func(b *BackoffDiscovery) error {
		if size < 0 {
			return fmt.Errorf("cannot set size to be smaller than 0")
		}
		b.returnedBufSz = size
		return nil
	}
}

type clock interface {
	Now() time.Time
}

type realClock struct{}

func (c realClock) Now() time.Time {
	return time.Now()
}

// withClock lets you override the default time.Now() call. Useful for tests.
func withClock(c clock) BackoffDiscoveryOption {
	return func(b *BackoffDiscovery) error {
		b.clock = c
		return nil
	}
}

type backoffCache struct {
	// strat is assigned on creation and not written to
	strat BackoffStrategy

	mux          sync.Mutex // guards writes to all following fields
	nextDiscover time.Time
	prevPeers    map[peer.ID]peer.AddrInfo
	peers        map[peer.ID]peer.AddrInfo
	sendingChs   map[chan peer.AddrInfo]int
	ongoing      bool

	clock clock
}

func (d *BackoffDiscovery) Advertise(ctx context.Context, ns string, opts ...discovery.Option) (time.Duration, error) {
	return d.disc.Advertise(ctx, ns, opts...)
}

func (d *BackoffDiscovery) FindPeers(ctx context.Context, ns string, opts ...discovery.Option) (<-chan peer.AddrInfo, error) {
	// Get options
	var options discovery.Options
	err := options.Apply(opts...)
	if err != nil {
		return nil, err
	}

	// Get cached peers
	d.peerCacheMux.RLock()
	c, ok := d.peerCache[ns]
	d.peerCacheMux.RUnlock()

	/*
		Overall plan:
		If it's time to look for peers, look for peers, then return them
		If it's not time then return cache
		If it's time to look for peers, but we have already started looking. Get up to speed with ongoing request
	*/

	// Setup cache if we don't have one yet
	if !ok {
		pc := &backoffCache{
			nextDiscover: time.Time{},
			prevPeers:    make(map[peer.ID]peer.AddrInfo),
			peers:        make(map[peer.ID]peer.AddrInfo),
			sendingChs:   make(map[chan peer.AddrInfo]int),
			strat:        d.stratFactory(),
			clock:        d.clock,
		}

		d.peerCacheMux.Lock()
		c, ok = d.peerCache[ns]

		if !ok {
			d.peerCache[ns] = pc
			c = pc
		}

		d.peerCacheMux.Unlock()
	}

	c.mux.Lock()
	defer c.mux.Unlock()

	timeExpired := d.clock.Now().After(c.nextDiscover)

	// If it's not yet time to search again and no searches are in progress then return cached peers
	if !(timeExpired || c.ongoing) {
		chLen := options.Limit

		if chLen == 0 {
			chLen = len(c.prevPeers)
		} else if chLen > len(c.prevPeers) {
			chLen = len(c.prevPeers)
		}
		pch := make(chan peer.AddrInfo, chLen)
		for _, ai := range c.prevPeers {
			select {
			case pch <- ai:
			default:
				// skip if we have asked for a lower limit than the number of peers known
			}
		}
		close(pch)
		return pch, nil
	}

	// If a request is not already in progress setup a dispatcher channel for dispatching incoming peers
	if !c.ongoing {
		pch, err := d.disc.FindPeers(ctx, ns, opts...)
		if err != nil {
			return nil, err
		}

		c.ongoing = true
		go findPeerDispatcher(ctx, c, pch)
	}

	// Setup receiver channel for receiving peers from ongoing requests
	evtCh := make(chan peer.AddrInfo, d.parallelBufSz)
	pch := make(chan peer.AddrInfo, d.returnedBufSz)
	rcvPeers := make([]peer.AddrInfo, 0, 32)
	for _, ai := range c.peers {
		rcvPeers = append(rcvPeers, ai)
	}
	c.sendingChs[evtCh] = options.Limit

	go findPeerReceiver(ctx, pch, evtCh, rcvPeers)

	return pch, nil
}

func findPeerDispatcher(ctx context.Context, c *backoffCache, pch <-chan peer.AddrInfo) {
	defer func() {
		c.mux.Lock()

		// If the peer addresses have changed reset the backoff
		if checkUpdates(c.prevPeers, c.peers) {
			c.strat.Reset()
			c.prevPeers = c.peers
		}
		c.nextDiscover = c.clock.Now().Add(c.strat.Delay())

		c.ongoing = false
		c.peers = make(map[peer.ID]peer.AddrInfo)

		for ch := range c.sendingChs {
			close(ch)
		}
		c.sendingChs = make(map[chan peer.AddrInfo]int)
		c.mux.Unlock()
	}()

	for {
		select {
		case ai, ok := <-pch:
			if !ok {
				return
			}
			c.mux.Lock()

			// If we receive the same peer multiple times return the address union
			var sendAi peer.AddrInfo
			if prevAi, ok := c.peers[ai.ID]; ok {
				if combinedAi := mergeAddrInfos(prevAi, ai); combinedAi != nil {
					sendAi = *combinedAi
				} else {
					c.mux.Unlock()
					continue
				}
			} else {
				sendAi = ai
			}

			c.peers[ai.ID] = sendAi

			for ch, rem := range c.sendingChs {
				if rem > 0 {
					ch <- sendAi
					c.sendingChs[ch] = rem - 1
				}
			}

			c.mux.Unlock()
		case <-ctx.Done():
			return
		}
	}
}

func findPeerReceiver(ctx context.Context, pch, evtCh chan peer.AddrInfo, rcvPeers []peer.AddrInfo) {
	defer close(pch)

	for {
		select {
		case ai, ok := <-evtCh:
			if ok {
				rcvPeers = append(rcvPeers, ai)

				sentAll := true
			sendPeers:
				for i, p := range rcvPeers {
					select {
					case pch <- p:
					default:
						rcvPeers = rcvPeers[i:]
						sentAll = false
						break sendPeers
					}
				}
				if sentAll {
					rcvPeers = []peer.AddrInfo{}
				}
			} else {
				for _, p := range rcvPeers {
					select {
					case pch <- p:
					case <-ctx.Done():
						return
					}
				}
				return
			}
		case <-ctx.Done():
			return
		}
	}
}

func mergeAddrInfos(prevAi, newAi peer.AddrInfo) *peer.AddrInfo {
	seen := make(map[string]struct{}, len(prevAi.Addrs))
	combinedAddrs := make([]ma.Multiaddr, 0, len(prevAi.Addrs))
	addAddrs := func(addrs []ma.Multiaddr) {
		for _, addr := range addrs {
			if _, ok := seen[addr.String()]; ok {
				continue
			}
			seen[addr.String()] = struct{}{}
			combinedAddrs = append(combinedAddrs, addr)
		}
	}
	addAddrs(prevAi.Addrs)
	addAddrs(newAi.Addrs)

	if len(combinedAddrs) > len(prevAi.Addrs) {
		combinedAi := &peer.AddrInfo{ID: prevAi.ID, Addrs: combinedAddrs}
		return combinedAi
	}
	return nil
}

func checkUpdates(orig, update map[peer.ID]peer.AddrInfo) bool {
	if len(orig) != len(update) {
		return true
	}
	for p, ai := range update {
		if prevAi, ok := orig[p]; ok {
			if combinedAi := mergeAddrInfos(prevAi, ai); combinedAi != nil {
				return true
			}
		} else {
			return true
		}
	}
	return false
}
