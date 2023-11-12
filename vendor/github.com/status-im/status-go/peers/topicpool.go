package peers

import (
	"container/heap"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/discv5"
	"github.com/ethereum/go-ethereum/p2p/enode"

	"github.com/status-im/status-go/discovery"
	"github.com/status-im/status-go/params"
)

const (
	// notQueuedIndex used to define that item is not queued in the heap queue.
	notQueuedIndex = -1
)

// maxCachedPeersMultiplier peers max limit will be multiplied by this number
// to get the maximum number of cached peers allowed.
var maxCachedPeersMultiplier = 1

// maxPendingPeersMultiplier peers max limit will be multiplied by this number
// to get the maximum number of pending peers allowed.
var maxPendingPeersMultiplier = 2

// TopicPoolInterface the TopicPool interface.
type TopicPoolInterface interface {
	StopSearch(server *p2p.Server)
	BelowMin() bool
	SearchRunning() bool
	StartSearch(server *p2p.Server) error
	ConfirmDropped(server *p2p.Server, nodeID enode.ID) bool
	AddPeerFromTable(server *p2p.Server) *discv5.Node
	MaxReached() bool
	ConfirmAdded(server *p2p.Server, nodeID enode.ID)
	isStopped() bool
	Topic() discv5.Topic
	SetLimits(limits params.Limits)
	setStopSearchTimeout(delay time.Duration)
	readyToStopSearch() bool
}

type Clock interface {
	Now() time.Time
}

type realClock struct{}

func (realClock) Now() time.Time { return time.Now() }

// newTopicPool returns instance of TopicPool.
func newTopicPool(discovery discovery.Discovery, topic discv5.Topic, limits params.Limits, slowMode, fastMode time.Duration, cache *Cache) *TopicPool {
	pool := TopicPool{
		discovery:            discovery,
		topic:                topic,
		limits:               limits,
		fastMode:             fastMode,
		slowMode:             slowMode,
		fastModeTimeout:      DefaultTopicFastModeTimeout,
		pendingPeers:         make(map[enode.ID]*peerInfoItem),
		discoveredPeersQueue: make(peerPriorityQueue, 0),
		discoveredPeers:      make(map[enode.ID]bool),
		connectedPeers:       make(map[enode.ID]*peerInfo),
		cache:                cache,
		maxCachedPeers:       limits.Max * maxCachedPeersMultiplier,
		maxPendingPeers:      limits.Max * maxPendingPeersMultiplier,
		clock:                realClock{},
	}
	heap.Init(&pool.discoveredPeersQueue)

	return &pool
}

// TopicPool manages peers for topic.
type TopicPool struct {
	discovery discovery.Discovery

	// configuration
	topic           discv5.Topic
	limits          params.Limits
	fastMode        time.Duration
	slowMode        time.Duration
	fastModeTimeout time.Duration

	mu     sync.RWMutex
	discWG sync.WaitGroup
	poolWG sync.WaitGroup
	quit   chan struct{}

	running int32

	currentMode           time.Duration
	period                chan time.Duration
	fastModeTimeoutCancel chan struct{}

	pendingPeers         map[enode.ID]*peerInfoItem // contains found and requested to be connected peers but not confirmed
	discoveredPeersQueue peerPriorityQueue          // priority queue to find the most recently discovered peers; does not containt peers requested to connect
	discoveredPeers      map[enode.ID]bool          // remembers which peers have already been discovered and are enqueued
	connectedPeers       map[enode.ID]*peerInfo     // currently connected peers

	stopSearchTimeout *time.Time

	maxPendingPeers int
	maxCachedPeers  int
	cache           *Cache

	clock Clock
}

func (t *TopicPool) addToPendingPeers(peer *peerInfo) {
	if _, ok := t.pendingPeers[peer.NodeID()]; ok {
		return
	}
	t.pendingPeers[peer.NodeID()] = &peerInfoItem{
		peerInfo: peer,
		index:    notQueuedIndex,
	}

	// maxPendingPeers = 0 means no limits.
	if t.maxPendingPeers == 0 || t.maxPendingPeers >= len(t.pendingPeers) {
		return
	}

	var oldestPeer *peerInfo
	for _, i := range t.pendingPeers {
		if oldestPeer != nil && oldestPeer.discoveredTime.Before(i.peerInfo.discoveredTime) {
			continue
		}

		oldestPeer = i.peerInfo
	}

	t.removeFromPendingPeers(oldestPeer.NodeID())
}

// addToQueue adds the passed peer to the queue if it is already pending.
func (t *TopicPool) addToQueue(peer *peerInfo) {
	if p, ok := t.pendingPeers[peer.NodeID()]; ok {
		if _, ok := t.discoveredPeers[peer.NodeID()]; ok {
			return
		}

		heap.Push(&t.discoveredPeersQueue, p)
		t.discoveredPeers[peer.NodeID()] = true
	}
}

func (t *TopicPool) popFromQueue() *peerInfo {
	if t.discoveredPeersQueue.Len() == 0 {
		return nil
	}
	item := heap.Pop(&t.discoveredPeersQueue).(*peerInfoItem)
	item.index = notQueuedIndex
	delete(t.discoveredPeers, item.peerInfo.NodeID())
	return item.peerInfo
}

func (t *TopicPool) removeFromPendingPeers(nodeID enode.ID) {
	peer, ok := t.pendingPeers[nodeID]
	if !ok {
		return
	}
	delete(t.pendingPeers, nodeID)
	if peer.index != notQueuedIndex {
		heap.Remove(&t.discoveredPeersQueue, peer.index)
		delete(t.discoveredPeers, nodeID)
	}
}

func (t *TopicPool) updatePendingPeer(nodeID enode.ID) {
	peer, ok := t.pendingPeers[nodeID]
	if !ok {
		return
	}
	peer.discoveredTime = t.clock.Now()
	if peer.index != notQueuedIndex {
		heap.Fix(&t.discoveredPeersQueue, peer.index)
	}
}

func (t *TopicPool) movePeerFromPoolToConnected(nodeID enode.ID) {
	peer, ok := t.pendingPeers[nodeID]
	if !ok {
		return
	}
	t.removeFromPendingPeers(nodeID)
	t.connectedPeers[nodeID] = peer.peerInfo
}

// SearchRunning returns true if search is running
func (t *TopicPool) SearchRunning() bool {
	return atomic.LoadInt32(&t.running) == 1
}

// MaxReached returns true if we connected with max number of peers.
func (t *TopicPool) MaxReached() bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return len(t.connectedPeers) == t.limits.Max
}

// BelowMin returns true if current number of peers is below min limit.
func (t *TopicPool) BelowMin() bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return len(t.connectedPeers) < t.limits.Min
}

// maxCachedPeersReached returns true if max number of cached peers is reached.
func (t *TopicPool) maxCachedPeersReached() bool {
	if t.maxCachedPeers == 0 {
		return true
	}
	peers := t.cache.GetPeersRange(t.topic, t.maxCachedPeers)

	return len(peers) >= t.maxCachedPeers
}

// setStopSearchTimeout sets the timeout to stop current topic search if it's not
// been stopped before.
func (t *TopicPool) setStopSearchTimeout(delay time.Duration) {
	if t.stopSearchTimeout != nil {
		return
	}
	now := t.clock.Now().Add(delay)
	t.stopSearchTimeout = &now
}

// isStopSearchDelayExpired returns true if the timeout to stop current topic
// search has been accomplished.
func (t *TopicPool) isStopSearchDelayExpired() bool {
	if t.stopSearchTimeout == nil {
		return false
	}
	return t.stopSearchTimeout.Before(t.clock.Now())
}

// readyToStopSearch return true if all conditions to stop search are ok.
func (t *TopicPool) readyToStopSearch() bool {
	return t.isStopSearchDelayExpired() || t.maxCachedPeersReached()
}

// updateSyncMode changes the sync mode depending on the current number
// of connected peers and limits.
func (t *TopicPool) updateSyncMode() {
	newMode := t.slowMode
	if len(t.connectedPeers) < t.limits.Min {
		newMode = t.fastMode
	}
	t.setSyncMode(newMode)
}

func (t *TopicPool) setSyncMode(mode time.Duration) {
	if mode == t.currentMode {
		return
	}

	t.period <- mode
	t.currentMode = mode

	// if selected mode is fast mode and fast mode timeout was not set yet,
	// do it now
	if mode == t.fastMode && t.fastModeTimeoutCancel == nil {
		t.fastModeTimeoutCancel = t.limitFastMode(t.fastModeTimeout)
	}
	// remove fast mode timeout as slow mode is selected now
	if mode == t.slowMode && t.fastModeTimeoutCancel != nil {
		close(t.fastModeTimeoutCancel)
		t.fastModeTimeoutCancel = nil
	}
}

func (t *TopicPool) limitFastMode(timeout time.Duration) chan struct{} {
	if timeout == 0 {
		return nil
	}

	cancel := make(chan struct{})

	t.poolWG.Add(1)
	go func() {
		defer t.poolWG.Done()

		select {
		case <-time.After(timeout):
			t.mu.Lock()
			t.setSyncMode(t.slowMode)
			t.mu.Unlock()
		case <-cancel:
			return
		}
	}()

	return cancel
}

// ConfirmAdded called when peer was added by p2p Server.
//  1. Skip a peer if it not in our peer table
//  2. Add a peer to a cache.
//  3. Disconnect a peer if it was connected after we reached max limit of peers.
//     (we can't know in advance if peer will be connected, thats why we allow
//     to overflow for short duration)
//  4. Switch search to slow mode if it is running.
func (t *TopicPool) ConfirmAdded(server *p2p.Server, nodeID enode.ID) {
	t.mu.Lock()
	defer t.mu.Unlock()

	peerInfoItem, ok := t.pendingPeers[nodeID]
	inbound := !ok || !peerInfoItem.added

	log.Debug("peer added event", "peer", nodeID.String(), "inbound", inbound)

	if inbound {
		return
	}

	peer := peerInfoItem.peerInfo // get explicit reference

	// established connection means that the node
	// is a viable candidate for a connection and can be cached
	if err := t.cache.AddPeer(peer.node, t.topic); err != nil {
		log.Error("failed to persist a peer", "error", err)
	}

	t.movePeerFromPoolToConnected(nodeID)
	// if the upper limit is already reached, drop this peer
	if len(t.connectedPeers) > t.limits.Max {
		log.Debug("max limit is reached drop the peer", "ID", nodeID, "topic", t.topic)
		peer.dismissed = true
		t.removeServerPeer(server, peer)
		return
	}

	// make sure `dismissed` is reset
	peer.dismissed = false

	// A peer was added so check if we can switch to slow mode.
	if t.SearchRunning() {
		t.updateSyncMode()
	}
}

// ConfirmDropped called when server receives drop event.
// 1. Skip peer if it is not in our peer table.
// 2. If disconnect request - we could drop that peer ourselves.
// 3. If connected number will drop below min limit - switch to fast mode.
// 4. Delete a peer from cache and peer table.
// Returns false if peer is not in our table or we requested removal of this peer.
// Otherwise peer is removed and true is returned.
func (t *TopicPool) ConfirmDropped(server *p2p.Server, nodeID enode.ID) bool {
	t.mu.Lock()
	defer t.mu.Unlock()

	// either inbound or connected from another topic
	peer, exist := t.connectedPeers[nodeID]
	if !exist {
		return false
	}

	log.Debug("disconnect", "ID", nodeID, "dismissed", peer.dismissed)

	delete(t.connectedPeers, nodeID)
	// Peer was removed by us because exceeded the limit.
	// Add it back to the pool as it can be useful in the future.
	if peer.dismissed {
		t.addToPendingPeers(peer)
		// use queue for peers that weren't added to p2p server
		t.addToQueue(peer)
		return false
	}

	// If there was a network error, this event will be received
	// but the peer won't be removed from the static nodes set.
	// That's why we need to call `removeServerPeer` manually.
	t.removeServerPeer(server, peer)

	if err := t.cache.RemovePeer(nodeID, t.topic); err != nil {
		log.Error("failed to remove peer from cache", "error", err)
	}

	// As we removed a peer, update a sync strategy if needed.
	if t.SearchRunning() {
		t.updateSyncMode()
	}

	return true
}

// AddPeerFromTable checks if there is a valid peer in local table and adds it to a server.
func (t *TopicPool) AddPeerFromTable(server *p2p.Server) *discv5.Node {
	t.mu.RLock()
	defer t.mu.RUnlock()

	// The most recently added peer is removed from the queue.
	// If it did not expire yet, it will be added to the server.
	// TODO(adam): investigate if it's worth to keep the peer in the queue
	// until the server confirms it is added and in the meanwhile only adjust its priority.
	peer := t.popFromQueue()
	if peer != nil && t.clock.Now().Before(peer.discoveredTime.Add(expirationPeriod)) {
		t.addServerPeer(server, peer)
		return peer.node
	}

	return nil
}

// StartSearch creates discv5 queries and runs a loop to consume found peers.
func (t *TopicPool) StartSearch(server *p2p.Server) error {
	if atomic.LoadInt32(&t.running) == 1 {
		return nil
	}
	if !t.discovery.Running() {
		return ErrDiscv5NotRunning
	}
	atomic.StoreInt32(&t.running, 1)

	t.mu.Lock()
	defer t.mu.Unlock()

	t.quit = make(chan struct{})
	t.stopSearchTimeout = nil

	// `period` is used to notify about the current sync mode.
	t.period = make(chan time.Duration, 2)
	// use fast sync mode at the beginning
	t.setSyncMode(t.fastMode)

	// peers management
	found := make(chan *discv5.Node, 5) // 5 reasonable number for concurrently found nodes
	lookup := make(chan bool, 10)       // sufficiently buffered channel, just prevents blocking because of lookup

	for _, peer := range t.cache.GetPeersRange(t.topic, 5) {
		log.Debug("adding a peer from cache", "peer", peer)
		found <- peer
	}

	t.discWG.Add(1)
	go func() {
		if err := t.discovery.Discover(string(t.topic), t.period, found, lookup); err != nil {
			log.Error("error searching foro", "topic", t.topic, "err", err)
		}
		t.discWG.Done()
	}()
	t.poolWG.Add(1)
	go func() {
		t.handleFoundPeers(server, found, lookup)
		t.poolWG.Done()
	}()

	return nil
}

func (t *TopicPool) handleFoundPeers(server *p2p.Server, found <-chan *discv5.Node, lookup <-chan bool) {
	selfID := discv5.PubkeyID(server.Self().Pubkey())
	for {
		select {
		case <-t.quit:
			return
		case <-lookup:
		case node := <-found:
			if node.ID == selfID {
				continue
			}
			if err := t.processFoundNode(server, node); err != nil {
				log.Error("failed to process found node", "node", node, "error", err)
			}
		}
	}
}

// processFoundNode called when node is discovered by kademlia search query
// 2 important conditions
//  1. every time when node is processed we need to update discoveredTime.
//     peer will be considered as valid later only if it was discovered < 60m ago
//  2. if peer is connected or if max limit is reached we are not a adding peer to p2p server
func (t *TopicPool) processFoundNode(server *p2p.Server, node *discv5.Node) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	pk, err := node.ID.Pubkey()
	if err != nil {
		return err
	}

	nodeID := enode.PubkeyToIDV4(pk)

	log.Debug("peer found", "ID", nodeID, "topic", t.topic)

	// peer is already connected so update only discoveredTime
	if peer, ok := t.connectedPeers[nodeID]; ok {
		peer.discoveredTime = t.clock.Now()
		return nil
	}

	if _, ok := t.pendingPeers[nodeID]; ok {
		t.updatePendingPeer(nodeID)
	} else {
		t.addToPendingPeers(&peerInfo{
			discoveredTime: t.clock.Now(),
			node:           node,
			publicKey:      pk,
		})
	}
	log.Debug(
		"adding peer to a server", "peer", node.ID.String(),
		"connected", len(t.connectedPeers), "max", t.maxCachedPeers)

	// This can happen when the monotonic clock is not precise enough and
	// multiple peers gets added at the same clock time, resulting in all
	// of them having the same discoveredTime.
	// At which point a random peer will be removed, sometimes being the
	// peer we just added.
	// We could make sure that the latest added peer is not removed,
	// but this is simpler, and peers will be fresh enough as resolution
	// should be quite high (ms at least).
	// This has been reported on windows builds
	// only https://github.com/status-im/nim-status-client/issues/522
	if t.pendingPeers[nodeID] == nil {
		log.Debug("peer added has just been removed", "peer", nodeID)
		return nil
	}

	// the upper limit is not reached, so let's add this peer
	if len(t.connectedPeers) < t.maxCachedPeers {
		t.addServerPeer(server, t.pendingPeers[nodeID].peerInfo)
	} else {
		t.addToQueue(t.pendingPeers[nodeID].peerInfo)
	}

	return nil
}

func (t *TopicPool) addServerPeer(server *p2p.Server, info *peerInfo) {
	info.added = true
	n := enode.NewV4(info.publicKey, info.node.IP, int(info.node.TCP), int(info.node.UDP))
	server.AddPeer(n)
}

func (t *TopicPool) removeServerPeer(server *p2p.Server, info *peerInfo) {
	info.added = false
	n := enode.NewV4(info.publicKey, info.node.IP, int(info.node.TCP), int(info.node.UDP))
	server.RemovePeer(n)
}

func (t *TopicPool) isStopped() bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.currentMode == 0
}

// StopSearch stops the closes stop
func (t *TopicPool) StopSearch(server *p2p.Server) {
	if !atomic.CompareAndSwapInt32(&t.running, 1, 0) {
		return
	}
	if t.quit == nil {
		return
	}
	select {
	case <-t.quit:
		return
	default:
	}
	log.Debug("stoping search", "topic", t.topic)
	close(t.quit)
	t.mu.Lock()
	if t.fastModeTimeoutCancel != nil {
		close(t.fastModeTimeoutCancel)
		t.fastModeTimeoutCancel = nil
	}
	t.currentMode = 0
	t.mu.Unlock()
	// wait for poolWG to exit because it writes to period channel
	t.poolWG.Wait()
	close(t.period)
	t.discWG.Wait()
}

// Topic exposes the internal discovery topic.
func (t *TopicPool) Topic() discv5.Topic {
	return t.topic
}

// SetLimits set the limits for the current TopicPool.
func (t *TopicPool) SetLimits(limits params.Limits) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.limits = limits
}
