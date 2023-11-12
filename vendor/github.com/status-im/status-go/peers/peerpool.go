package peers

import (
	"crypto/ecdsa"
	"errors"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/event"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/discv5"
	"github.com/ethereum/go-ethereum/p2p/enode"

	"github.com/status-im/status-go/discovery"
	"github.com/status-im/status-go/params"
	"github.com/status-im/status-go/peers/verifier"
	"github.com/status-im/status-go/signal"
)

var (
	// ErrDiscv5NotRunning returned when pool is started but discover v5 is not running or not enabled.
	ErrDiscv5NotRunning = errors.New("Discovery v5 is not running")
)

// PoolEvent is a type used to for peer pool events.
type PoolEvent string

const (
	immediately = 0 * time.Minute
	// expirationPeriod is an amount of time while peer is considered as a connectable
	expirationPeriod = 60 * time.Minute
	// discoveryRestartTimeout defines how often loop will try to start discovery server
	discoveryRestartTimeout = 2 * time.Second
	// DefaultFastSync is a recommended value for aggressive peers search.
	DefaultFastSync = 3 * time.Second
	// DefaultSlowSync is a recommended value for slow (background) peers search.
	DefaultSlowSync = 30 * time.Second
	// DefaultDiscV5Timeout is a timeout after which Discv5 is stopped.
	DefaultDiscV5Timeout = 3 * time.Minute
	// DefaultTopicFastModeTimeout is a timeout after which sync mode is switched to slow mode.
	DefaultTopicFastModeTimeout = 30 * time.Second
	// DefaultTopicStopSearchDelay is the default delay when stopping a topic search.
	DefaultTopicStopSearchDelay = 10 * time.Second
)

// Options is a struct with PeerPool configuration.
type Options struct {
	FastSync time.Duration
	SlowSync time.Duration
	// After this time, Discovery is stopped even if max peers is not reached.
	DiscServerTimeout time.Duration
	// AllowStop allows stopping Discovery when reaching max peers or after timeout.
	AllowStop bool
	// TopicStopSearchDelay time stopSearch will be waiting for max cached peers to be
	// filled before really stopping the search.
	TopicStopSearchDelay time.Duration
	// TrustedMailServers is a list of trusted nodes.
	TrustedMailServers []enode.ID
}

// NewDefaultOptions returns a struct with default Options.
func NewDefaultOptions() *Options {
	return &Options{
		FastSync:             DefaultFastSync,
		SlowSync:             DefaultSlowSync,
		DiscServerTimeout:    DefaultDiscV5Timeout,
		AllowStop:            false,
		TopicStopSearchDelay: DefaultTopicStopSearchDelay,
	}
}

type peerInfo struct {
	// discoveredTime last time when node was found by v5
	discoveredTime time.Time
	// dismissed is true when our node requested a disconnect
	dismissed bool
	// added is true when the node tries to add this peer to a server
	added bool

	node *discv5.Node
	// store public key separately to make peerInfo more independent from discv5
	publicKey *ecdsa.PublicKey
}

func (p *peerInfo) NodeID() enode.ID {
	return enode.PubkeyToIDV4(p.publicKey)
}

// PeerPool manages discovered peers and connects them to p2p server
type PeerPool struct {
	opts *Options

	discovery discovery.Discovery

	// config can be set only once per pool life cycle
	config map[discv5.Topic]params.Limits
	cache  *Cache

	mu                 sync.RWMutex
	timeoutMu          sync.RWMutex
	topics             []TopicPoolInterface
	serverSubscription event.Subscription
	events             chan *p2p.PeerEvent
	quit               chan struct{}
	wg                 sync.WaitGroup
	timeout            <-chan time.Time
	updateTopic        chan *updateTopicRequest
}

// NewPeerPool creates instance of PeerPool
func NewPeerPool(discovery discovery.Discovery, config map[discv5.Topic]params.Limits, cache *Cache, options *Options) *PeerPool {
	return &PeerPool{
		opts:      options,
		discovery: discovery,
		config:    config,
		cache:     cache,
	}
}

func (p *PeerPool) setDiscoveryTimeout() {
	p.timeoutMu.Lock()
	defer p.timeoutMu.Unlock()
	if p.opts.AllowStop && p.opts.DiscServerTimeout > 0 {
		p.timeout = time.After(p.opts.DiscServerTimeout)
	}
}

// Start creates topic pool for each topic in config and subscribes to server events.
func (p *PeerPool) Start(server *p2p.Server) error {
	if !p.discovery.Running() {
		return ErrDiscv5NotRunning
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	// init channels
	p.quit = make(chan struct{})
	p.updateTopic = make(chan *updateTopicRequest)
	p.setDiscoveryTimeout()

	// subscribe to peer events
	p.events = make(chan *p2p.PeerEvent, 20)
	p.serverSubscription = server.SubscribeEvents(p.events)
	p.wg.Add(1)
	go func() {
		p.handleServerPeers(server, p.events)
		p.wg.Done()
	}()

	// collect topics and start searching for nodes
	p.topics = make([]TopicPoolInterface, 0, len(p.config))
	for topic, limits := range p.config {
		var topicPool TopicPoolInterface
		t := newTopicPool(p.discovery, topic, limits, p.opts.SlowSync, p.opts.FastSync, p.cache)
		if topic == MailServerDiscoveryTopic {
			v, err := p.initVerifier()
			if err != nil {
				return err
			}
			topicPool = newCacheOnlyTopicPool(t, v)
		} else {
			topicPool = t
		}
		if err := topicPool.StartSearch(server); err != nil {
			return err
		}
		p.topics = append(p.topics, topicPool)
	}

	// discovery must be already started when pool is started
	signal.SendDiscoveryStarted()

	return nil
}

func (p *PeerPool) initVerifier() (v Verifier, err error) {
	return verifier.NewLocalVerifier(p.opts.TrustedMailServers), nil
}

func (p *PeerPool) startDiscovery() error {
	if p.discovery.Running() {
		return nil
	}

	if err := p.discovery.Start(); err != nil {
		return err
	}

	p.mu.Lock()
	p.setDiscoveryTimeout()
	p.mu.Unlock()

	signal.SendDiscoveryStarted()

	return nil
}

func (p *PeerPool) stopDiscovery(server *p2p.Server) {
	if !p.discovery.Running() {
		return
	}

	if err := p.discovery.Stop(); err != nil {
		log.Error("discovery errored when stopping", "err", err)
	}
	for _, t := range p.topics {
		t.StopSearch(server)
	}

	p.timeoutMu.Lock()
	p.timeout = nil
	p.timeoutMu.Unlock()

	signal.SendDiscoveryStopped()
}

// restartDiscovery and search for topics that have peer count below min
func (p *PeerPool) restartDiscovery(server *p2p.Server) error {
	if !p.discovery.Running() {
		if err := p.startDiscovery(); err != nil {
			return err
		}
		log.Debug("restarted discovery from peer pool")
	}
	for _, t := range p.topics {
		if !t.BelowMin() || t.SearchRunning() {
			continue
		}
		err := t.StartSearch(server)
		if err != nil {
			log.Error("search failed to start", "error", err)
		}
	}
	return nil
}

// handleServerPeers watches server peer events, notifies topic pools about changes
// in the peer set and stops the discv5 if all topic pools collected enough peers.
//
// @TODO(adam): split it into peers and discovery management loops. This should
// simplify the whole logic and allow to remove `timeout` field from `PeerPool`.
func (p *PeerPool) handleServerPeers(server *p2p.Server, events <-chan *p2p.PeerEvent) {
	retryDiscv5 := make(chan struct{}, 1)
	stopDiscv5 := make(chan struct{}, 1)

	queueRetry := func(d time.Duration) {
		go func() {
			time.Sleep(d)
			select {
			case retryDiscv5 <- struct{}{}:
			default:
			}
		}()

	}

	queueStop := func() {
		go func() {
			select {
			case stopDiscv5 <- struct{}{}:
			default:
			}
		}()

	}

	for {
		// We use a separate lock for timeout, as this loop should
		// always be running, otherwise the p2p.Server will hang.
		// Because the handler of events might potentially hang on the
		// server, deadlocking if this loop is waiting for the global lock.
		// NOTE: this code probably needs to be refactored and simplified
		// as it's difficult to follow the asynchronous nature of it.
		p.timeoutMu.RLock()
		timeout := p.timeout
		p.timeoutMu.RUnlock()

		select {
		case <-p.quit:
			log.Debug("stopping DiscV5 because of quit")
			p.stopDiscovery(server)
			return
		case <-timeout:
			log.Info("DiscV5 timed out")
			p.stopDiscovery(server)
		case <-retryDiscv5:
			if err := p.restartDiscovery(server); err != nil {
				log.Error("starting discv5 failed", "error", err, "retry", discoveryRestartTimeout)
				queueRetry(discoveryRestartTimeout)
			}
		case <-stopDiscv5:
			p.handleStopTopics(server)
		case req := <-p.updateTopic:
			if p.updateTopicLimits(server, req) == nil {
				if !p.discovery.Running() {
					queueRetry(immediately)
				}
			}
		case event := <-events:
			// NOTE: handlePeerEventType needs to be called asynchronously
			// as it publishes on the <-events channel, leading to a deadlock
			// if events channel is full.
			go p.handlePeerEventType(server, event, queueRetry, queueStop)
		}
	}
}

func (p *PeerPool) handlePeerEventType(server *p2p.Server, event *p2p.PeerEvent, queueRetry func(time.Duration), queueStop func()) {
	p.mu.Lock()
	defer p.mu.Unlock()

	var shouldRetry bool
	var shouldStop bool
	switch event.Type {
	case p2p.PeerEventTypeDrop:
		log.Debug("confirm peer dropped", "ID", event.Peer)
		if p.handleDroppedPeer(server, event.Peer) {
			shouldRetry = true
		}
	case p2p.PeerEventTypeAdd: // skip other events
		log.Debug("confirm peer added", "ID", event.Peer)
		p.handleAddedPeer(server, event.Peer)
		shouldStop = true
	default:
		return
	}

	// First we send the discovery summary
	SendDiscoverySummary(server.PeersInfo())

	// then we send the stop event
	if shouldRetry {
		queueRetry(immediately)
	} else if shouldStop {
		queueStop()
	}
}

// handleAddedPeer notifies all topics about added peer.
func (p *PeerPool) handleAddedPeer(server *p2p.Server, nodeID enode.ID) {
	for _, t := range p.topics {
		t.ConfirmAdded(server, nodeID)
		if p.opts.AllowStop && t.MaxReached() {
			t.setStopSearchTimeout(p.opts.TopicStopSearchDelay)
		}
	}
}

// handleStopTopics stops the search on any topics having reached its max cached
// limit or its delay stop is expired, additionally will stop discovery if all
// peers are stopped.
func (p *PeerPool) handleStopTopics(server *p2p.Server) {
	if !p.opts.AllowStop {
		return
	}
	for _, t := range p.topics {
		if t.readyToStopSearch() {
			t.StopSearch(server)
		}
	}
	if p.allTopicsStopped() {
		log.Debug("closing discv5 connection because all topics reached max limit")
		p.stopDiscovery(server)
	}
}

// allTopicsStopped returns true if all topics are stopped.
func (p *PeerPool) allTopicsStopped() (all bool) {
	if !p.opts.AllowStop {
		return false
	}
	all = true
	for _, t := range p.topics {
		if !t.isStopped() {
			all = false
		}
	}
	return all
}

// handleDroppedPeer notifies every topic about dropped peer and returns true if any peer have connections
// below min limit
func (p *PeerPool) handleDroppedPeer(server *p2p.Server, nodeID enode.ID) (any bool) {
	for _, t := range p.topics {
		confirmed := t.ConfirmDropped(server, nodeID)
		if confirmed {
			newPeer := t.AddPeerFromTable(server)
			if newPeer != nil {
				log.Debug("added peer from local table", "ID", newPeer.ID)
			}
		}
		log.Debug("search", "topic", t.Topic(), "below min", t.BelowMin())
		if t.BelowMin() && !t.SearchRunning() {
			any = true
		}
	}
	return any
}

// Stop closes pool quit channel and all channels that are watched by search queries
// and waits till all goroutines will exit.
func (p *PeerPool) Stop() {
	// pool wasn't started
	if p.quit == nil {
		return
	}
	select {
	case <-p.quit:
		return
	default:
		log.Debug("started closing peer pool")
		close(p.quit)
	}
	p.serverSubscription.Unsubscribe()
	p.wg.Wait()
}

type updateTopicRequest struct {
	Topic  string
	Limits params.Limits
}

// UpdateTopic updates the pre-existing TopicPool limits.
func (p *PeerPool) UpdateTopic(topic string, limits params.Limits) error {
	if _, err := p.getTopic(topic); err != nil {
		return err
	}

	p.updateTopic <- &updateTopicRequest{
		Topic:  topic,
		Limits: limits,
	}

	return nil
}

func (p *PeerPool) updateTopicLimits(server *p2p.Server, req *updateTopicRequest) error {
	t, err := p.getTopic(req.Topic)
	if err != nil {
		return err
	}
	t.SetLimits(req.Limits)
	return nil
}

func (p *PeerPool) getTopic(topic string) (TopicPoolInterface, error) {
	for _, t := range p.topics {
		if t.Topic() == discv5.Topic(topic) {
			return t, nil
		}
	}
	return nil, errors.New("topic not found")
}
