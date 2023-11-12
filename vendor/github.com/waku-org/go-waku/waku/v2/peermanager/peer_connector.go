package peermanager

// Adapted from github.com/libp2p/go-libp2p@v0.23.2/p2p/discovery/backoff/backoffconnector.go

import (
	"context"
	"errors"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"

	"github.com/libp2p/go-libp2p/p2p/discovery/backoff"
	"github.com/waku-org/go-waku/logging"
	wps "github.com/waku-org/go-waku/waku/v2/peerstore"
	waku_proto "github.com/waku-org/go-waku/waku/v2/protocol"
	"github.com/waku-org/go-waku/waku/v2/service"

	"go.uber.org/zap"

	lru "github.com/hashicorp/golang-lru"
)

// PeerConnectionStrategy is a utility to connect to peers,
// but only if we have not recently tried connecting to them already
type PeerConnectionStrategy struct {
	mux   sync.Mutex
	cache *lru.TwoQueueCache
	host  host.Host
	pm    *PeerManager

	paused      atomic.Bool
	dialTimeout time.Duration
	*service.CommonDiscoveryService
	subscriptions []subscription

	backoff backoff.BackoffFactory
	logger  *zap.Logger
}

type subscription struct {
	ctx context.Context
	ch  <-chan service.PeerData
}

// backoff describes the strategy used to decide how long to backoff after previously attempting to connect to a peer
func getBackOff() backoff.BackoffFactory {
	rngSrc := rand.NewSource(rand.Int63())
	minBackoff, maxBackoff := time.Minute, time.Hour
	bkf := backoff.NewExponentialBackoff(minBackoff, maxBackoff, backoff.FullJitter, time.Second, 5.0, 0, rand.New(rngSrc))
	return bkf
}

// NewPeerConnectionStrategy creates a utility to connect to peers,
// but only if we have not recently tried connecting to them already.
//
// dialTimeout is how long we attempt to connect to a peer before giving up
// minPeers is the minimum number of peers that the node should have
func NewPeerConnectionStrategy(pm *PeerManager,
	dialTimeout time.Duration, logger *zap.Logger) (*PeerConnectionStrategy, error) {
	// cacheSize is the size of a TwoQueueCache
	cacheSize := 600
	cache, err := lru.New2Q(cacheSize)
	if err != nil {
		return nil, err
	}
	//
	pc := &PeerConnectionStrategy{
		cache:                  cache,
		dialTimeout:            dialTimeout,
		CommonDiscoveryService: service.NewCommonDiscoveryService(),
		pm:                     pm,
		backoff:                getBackOff(),
		logger:                 logger.Named("discovery-connector"),
	}
	pm.SetPeerConnector(pc)
	return pc, nil
}

type connCacheData struct {
	nextTry time.Time
	strat   backoff.BackoffStrategy
}

// Subscribe receives channels on which discovered peers should be pushed
func (c *PeerConnectionStrategy) Subscribe(ctx context.Context, ch <-chan service.PeerData) {
	// if not running yet, store the subscription and return
	if err := c.ErrOnNotRunning(); err != nil {
		c.mux.Lock()
		c.subscriptions = append(c.subscriptions, subscription{ctx, ch})
		c.mux.Unlock()
		return
	}
	// if running start a goroutine to consume the subscription
	c.WaitGroup().Add(1)
	go func() {
		defer c.WaitGroup().Done()
		c.consumeSubscription(subscription{ctx, ch})
	}()
}

func (c *PeerConnectionStrategy) consumeSubscription(s subscription) {
	for {
		// for returning from the loop when peerConnector is paused.
		select {
		case <-c.Context().Done():
			return
		case <-s.ctx.Done():
			return
		default:
		}
		//
		if !c.isPaused() {
			select {
			case <-c.Context().Done():
				return
			case <-s.ctx.Done():
				return
			case p, ok := <-s.ch:
				if !ok {
					return
				}
				triggerImmediateConnection := false
				//Not connecting to peer as soon as it is discovered,
				// rather expecting this to be pushed from PeerManager based on the need.
				if len(c.host.Network().Peers()) < waku_proto.GossipSubOptimalFullMeshSize {
					triggerImmediateConnection = true
				}
				c.logger.Debug("adding discovered peer", logging.HostID("peer", p.AddrInfo.ID))
				c.pm.AddDiscoveredPeer(p, triggerImmediateConnection)

			case <-time.After(1 * time.Second):
				// This timeout is to not lock the goroutine
				break
			}
		} else {
			time.Sleep(1 * time.Second) // sleep while the peerConnector is paused.
		}
	}
}

// SetHost sets the host to be able to mount or consume a protocol
func (c *PeerConnectionStrategy) SetHost(h host.Host) {
	c.host = h
}

// Start attempts to connect to the peers passed in by peerCh.
// Will not connect to peers if they are within the backoff period.
func (c *PeerConnectionStrategy) Start(ctx context.Context) error {
	return c.CommonDiscoveryService.Start(ctx, c.start)

}
func (c *PeerConnectionStrategy) start() error {
	c.WaitGroup().Add(1)

	go c.dialPeers()

	c.consumeSubscriptions()

	return nil
}

// Stop terminates the peer-connector
func (c *PeerConnectionStrategy) Stop() {
	c.CommonDiscoveryService.Stop(func() {})
}

func (c *PeerConnectionStrategy) isPaused() bool {
	return c.paused.Load()
}

// it might happen Subscribe is called before peerConnector has started so store these subscriptions in subscriptions array and custom after c.cancel is set.
func (c *PeerConnectionStrategy) consumeSubscriptions() {
	for _, subs := range c.subscriptions {
		c.WaitGroup().Add(1)
		go func(s subscription) {
			defer c.WaitGroup().Done()
			c.consumeSubscription(s)
		}(subs)
	}
	c.subscriptions = nil
}

const maxActiveDials = 5

// c.cache is thread safe
// only reason why mutex is used: if canDialPeer is queried twice for the same peer.
func (c *PeerConnectionStrategy) canDialPeer(pi peer.AddrInfo) bool {
	c.mux.Lock()
	defer c.mux.Unlock()
	val, ok := c.cache.Get(pi.ID)
	var cachedPeer *connCacheData
	if ok {
		tv := val.(*connCacheData)
		now := time.Now()
		if now.Before(tv.nextTry) {
			return false
		}

		tv.nextTry = now.Add(tv.strat.Delay())
	} else {
		cachedPeer = &connCacheData{strat: c.backoff()}
		cachedPeer.nextTry = time.Now().Add(cachedPeer.strat.Delay())
		c.cache.Add(pi.ID, cachedPeer)
	}
	return true
}

func (c *PeerConnectionStrategy) dialPeers() {
	defer c.WaitGroup().Done()

	maxGoRoutines := c.pm.OutRelayPeersTarget
	if maxGoRoutines > maxActiveDials {
		maxGoRoutines = maxActiveDials
	}

	sem := make(chan struct{}, maxGoRoutines)

	for {
		select {
		case pd, ok := <-c.GetListeningChan():
			if !ok {
				return
			}
			addrInfo := pd.AddrInfo

			if addrInfo.ID == c.host.ID() || addrInfo.ID == "" ||
				c.host.Network().Connectedness(addrInfo.ID) == network.Connected {
				continue
			}

			if c.canDialPeer(addrInfo) {
				sem <- struct{}{}
				c.WaitGroup().Add(1)
				go c.dialPeer(addrInfo, sem)
			}
		case <-c.Context().Done():
			return
		}
	}
}

func (c *PeerConnectionStrategy) dialPeer(pi peer.AddrInfo, sem chan struct{}) {
	defer c.WaitGroup().Done()
	ctx, cancel := context.WithTimeout(c.Context(), c.dialTimeout)
	defer cancel()
	err := c.host.Connect(ctx, pi)
	if err != nil && !errors.Is(err, context.Canceled) {
		c.host.Peerstore().(wps.WakuPeerstore).AddConnFailure(pi)
		c.logger.Warn("connecting to peer", logging.HostID("peerID", pi.ID), zap.Error(err))
	}
	<-sem
}
