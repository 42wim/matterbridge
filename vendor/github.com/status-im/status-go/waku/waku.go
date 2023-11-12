// Copyright 2019 The Waku Library Authors.
//
// The Waku library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The Waku library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty off
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the Waku library. If not, see <http://www.gnu.org/licenses/>.
//
// This software uses the go-ethereum library, which is licensed
// under the GNU Lesser General Public Library, version 3 or any later.

package waku

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/sha256"
	"errors"
	"fmt"
	"runtime"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common/hexutil"

	"go.uber.org/zap"

	mapset "github.com/deckarep/golang-set"
	"golang.org/x/crypto/pbkdf2"

	gethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/event"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethereum/go-ethereum/rpc"

	"github.com/status-im/status-go/eth-node/types"
	"github.com/status-im/status-go/logutils"
	"github.com/status-im/status-go/waku/common"
	v0 "github.com/status-im/status-go/waku/v0"
	v1 "github.com/status-im/status-go/waku/v1"
)

const messageQueueLimit = 1024

type Bridge interface {
	Pipe() (<-chan *common.Envelope, chan<- *common.Envelope)
}

type settings struct {
	MaxMsgSize               uint32                    // Maximal message length allowed by the waku node
	EnableConfirmations      bool                      // Enable sending message confirmations
	MinPow                   float64                   // Minimal PoW required by the waku node
	MinPowTolerance          float64                   // Minimal PoW tolerated by the waku node for a limited time
	BloomFilter              []byte                    // Bloom filter for topics of interest for this node
	BloomFilterTolerance     []byte                    // Bloom filter tolerated by the waku node for a limited time
	TopicInterest            map[common.TopicType]bool // Topic interest for this node
	TopicInterestTolerance   map[common.TopicType]bool // Topic interest tolerated by the waku node for a limited time
	SoftBlacklistedPeerIDs   map[string]bool           // SoftBlacklistedPeerIDs is a list of peer ids that we want to keep connected but silently drop any envelope from
	BloomFilterMode          bool                      // Whether we should match against bloom-filter only
	LightClient              bool                      // Light client mode enabled does not forward messages
	RestrictLightClientsConn bool                      // Restrict connection between two light clients
	SyncAllowance            int                       // Maximum time in seconds allowed to process the waku-related messages
	FullNode                 bool                      // Whether this is to be run in FullNode settings
}

// Waku represents a dark communication interface through the Ethereum
// network, using its very own P2P communication layer.
type Waku struct {
	protocols []p2p.Protocol  // Peer description and parameters
	filters   *common.Filters // Message filters installed with Subscribe function

	privateKeys map[string]*ecdsa.PrivateKey // Private key storage
	symKeys     map[string][]byte            // Symmetric key storage
	keyMu       sync.RWMutex                 // Mutex associated with key stores

	envelopes   map[gethcommon.Hash]*common.Envelope // Pool of envelopes currently tracked by this node
	expirations map[uint32]mapset.Set                // Message expiration pool
	poolMu      sync.RWMutex                         // Mutex to sync the message and expiration pools

	stats  *common.StatsTracker
	peers  map[common.Peer]struct{} // Set of currently active peers
	peerMu sync.RWMutex             // Mutex to sync the active peer set

	msgQueue    chan *common.Envelope    // Message queue for normal waku messages
	p2pMsgQueue chan interface{}         // Message queue for peer-to-peer messages (not to be forwarded any further) and history delivery confirmations.
	p2pMsgIDs   map[gethcommon.Hash]bool // Map of the currently processing ids
	p2pMsgIDsMu sync.RWMutex

	quit chan struct{} // Channel used for graceful exit

	settings   settings     // Holds configuration settings that can be dynamically changed
	settingsMu sync.RWMutex // Mutex to sync the settings access

	mailServer MailServer

	rateLimiter *common.PeerRateLimiter

	envelopeFeed event.Feed

	timeSource func() time.Time // source of time for waku

	bridge       Bridge
	bridgeWg     sync.WaitGroup
	cancelBridge chan struct{}

	logger *zap.Logger
}

// New creates a Waku client ready to communicate through the Ethereum P2P network.
func New(cfg *Config, logger *zap.Logger) *Waku {
	if logger == nil {
		logger = logutils.ZapLogger()
	}

	logger.Debug("starting waku with config", zap.Any("config", cfg))
	if cfg == nil {
		c := DefaultConfig
		cfg = &c
	}

	waku := &Waku{
		privateKeys: make(map[string]*ecdsa.PrivateKey),
		symKeys:     make(map[string][]byte),
		envelopes:   make(map[gethcommon.Hash]*common.Envelope),
		expirations: make(map[uint32]mapset.Set),
		peers:       make(map[common.Peer]struct{}),
		msgQueue:    make(chan *common.Envelope, messageQueueLimit),
		p2pMsgQueue: make(chan interface{}, messageQueueLimit),
		p2pMsgIDs:   make(map[gethcommon.Hash]bool),
		quit:        make(chan struct{}),
		timeSource:  time.Now,
		logger:      logger,
	}

	waku.settings = settings{
		MaxMsgSize:               cfg.MaxMessageSize,
		MinPow:                   cfg.MinimumAcceptedPoW,
		MinPowTolerance:          cfg.MinimumAcceptedPoW,
		EnableConfirmations:      cfg.EnableConfirmations,
		LightClient:              cfg.LightClient,
		FullNode:                 cfg.FullNode,
		BloomFilterMode:          cfg.BloomFilterMode,
		SoftBlacklistedPeerIDs:   make(map[string]bool),
		RestrictLightClientsConn: cfg.RestrictLightClientsConn,
		SyncAllowance:            common.DefaultSyncAllowance,
	}

	for _, peerID := range cfg.SoftBlacklistedPeerIDs {
		waku.settings.SoftBlacklistedPeerIDs[peerID] = true
	}

	if cfg.FullNode {
		waku.settings.BloomFilter = common.MakeFullNodeBloom()
		waku.settings.BloomFilterTolerance = common.MakeFullNodeBloom()
	}

	waku.filters = common.NewFilters()
	waku.stats = &common.StatsTracker{}

	// p2p waku sub-protocol handler
	waku.protocols = []p2p.Protocol{{
		Name:    v0.Name,
		Version: uint(v0.Version),
		Length:  v0.NumberOfMessageCodes,
		Run:     waku.handlePeerV0,
		NodeInfo: func() interface{} {
			return map[string]interface{}{
				"version":        v0.VersionStr,
				"maxMessageSize": waku.MaxMessageSize(),
				"minimumPoW":     waku.MinPow(),
			}
		},
	},
		{
			Name:    v1.Name,
			Version: uint(v1.Version),
			Length:  v1.NumberOfMessageCodes,
			Run:     waku.handlePeerV1,
			NodeInfo: func() interface{} {
				return map[string]interface{}{
					"version":        v1.VersionStr,
					"maxMessageSize": waku.MaxMessageSize(),
					"minimumPoW":     waku.MinPow(),
				}
			},
		},
	}

	return waku
}

func (w *Waku) GetStats() types.StatsSummary {
	return w.stats.GetStats()
}

// MinPow returns the PoW value required by this node.
func (w *Waku) MinPow() float64 {
	w.settingsMu.RLock()
	defer w.settingsMu.RUnlock()
	return w.settings.MinPow
}

// SetMinimumPoW sets the minimal PoW required by this node
func (w *Waku) SetMinimumPoW(val float64, tolerate bool) error {
	if val < 0.0 {
		return fmt.Errorf("invalid PoW: %f", val)
	}

	w.settingsMu.Lock()
	w.settings.MinPow = val
	w.settingsMu.Unlock()

	w.notifyPeersAboutPowRequirementChange(val)

	if tolerate {
		go func() {
			// allow some time before all the peers have processed the notification
			select {
			case <-w.quit:
				return
			case <-time.After(time.Duration(w.settings.SyncAllowance) * time.Second):
				w.settingsMu.Lock()
				w.settings.MinPowTolerance = val
				w.settingsMu.Unlock()
			}
		}()
	}

	return nil
}

// MinPowTolerance returns the value of minimum PoW which is tolerated for a limited
// time after PoW was changed. If sufficient time have elapsed or no change of PoW
// have ever occurred, the return value will be the same as return value of MinPow().
func (w *Waku) MinPowTolerance() float64 {
	w.settingsMu.RLock()
	defer w.settingsMu.RUnlock()
	return w.settings.MinPowTolerance
}

// BloomFilter returns the aggregated bloom filter for all the topics of interest.
// The nodes are required to send only messages that match the advertised bloom filter.
// If a message does not match the bloom, it will tantamount to spam, and the peer will
// be disconnected.
func (w *Waku) BloomFilter() []byte {
	if w.FullNode() {
		return common.MakeFullNodeBloom()
	}

	w.settingsMu.RLock()
	defer w.settingsMu.RUnlock()
	return w.settings.BloomFilter
}

// BloomFilterTolerance returns the bloom filter which is tolerated for a limited
// time after new bloom was advertised to the peers. If sufficient time have elapsed
// or no change of bloom filter have ever occurred, the return value will be the same
// as return value of BloomFilter().
func (w *Waku) BloomFilterTolerance() []byte {
	if w.FullNode() {
		return common.MakeFullNodeBloom()
	}

	w.settingsMu.RLock()
	defer w.settingsMu.RUnlock()
	return w.settings.BloomFilterTolerance
}

// BloomFilterMode returns whether the node is running in bloom filter mode
func (w *Waku) BloomFilterMode() bool {
	if w.FullNode() {
		return true
	}

	w.settingsMu.RLock()
	defer w.settingsMu.RUnlock()
	return w.settings.BloomFilterMode
}

// SetBloomFilter sets the new bloom filter
func (w *Waku) SetBloomFilter(bloom []byte) error {
	if len(bloom) != common.BloomFilterSize {
		return fmt.Errorf("invalid bloom filter size: %d", len(bloom))
	}

	b := make([]byte, common.BloomFilterSize)
	copy(b, bloom)

	w.settingsMu.Lock()
	w.settings.BloomFilter = b
	// Setting bloom filter reset topic interest
	w.settings.TopicInterest = nil
	w.settingsMu.Unlock()
	w.notifyPeersAboutBloomFilterChange(b)

	go func() {
		// allow some time before all the peers have processed the notification
		select {
		case <-w.quit:
			return
		case <-time.After(time.Duration(w.settings.SyncAllowance) * time.Second):
			w.settingsMu.Lock()
			w.settings.BloomFilterTolerance = b
			w.settingsMu.Unlock()
		}

	}()

	return nil
}

// TopicInterest returns the all the topics of interest.
// The nodes are required to send only messages that match the advertised topics.
// If a message does not match the topic-interest, it will tantamount to spam, and the peer will
// be disconnected.
func (w *Waku) TopicInterest() []common.TopicType {
	w.settingsMu.RLock()
	defer w.settingsMu.RUnlock()
	// Return nil if FullNode as otherwise topic interest will have precedence
	if w.settings.FullNode || w.settings.TopicInterest == nil {
		return nil
	}
	topicInterest := make([]common.TopicType, len(w.settings.TopicInterest))

	i := 0
	for topic := range w.settings.TopicInterest {
		topicInterest[i] = topic
		i++
	}
	return topicInterest
}

// updateTopicInterest adds a new topic interest
// and informs the peers
func (w *Waku) updateTopicInterest(f *common.Filter) error {
	newTopicInterest := w.TopicInterest()
	for _, t := range f.Topics {
		top := common.BytesToTopic(t)
		newTopicInterest = append(newTopicInterest, top)
	}

	return w.SetTopicInterest(newTopicInterest)
}

// SetTopicInterest sets the new topicInterest
func (w *Waku) SetTopicInterest(topicInterest []common.TopicType) error {
	var topicInterestMap map[common.TopicType]bool
	if len(topicInterest) > common.MaxTopicInterest {
		return fmt.Errorf("invalid topic interest: %d", len(topicInterest))
	}

	if topicInterest != nil {
		topicInterestMap = make(map[common.TopicType]bool, len(topicInterest))
		for _, topic := range topicInterest {
			topicInterestMap[topic] = true
		}
	}

	w.settingsMu.Lock()
	w.settings.TopicInterest = topicInterestMap
	// Setting topic interest resets bloom filter
	w.settings.BloomFilter = nil
	w.settingsMu.Unlock()
	w.notifyPeersAboutTopicInterestChange(topicInterest)

	go func() {
		// allow some time before all the peers have processed the notification
		select {
		case <-w.quit:
			return
		case <-time.After(time.Duration(w.settings.SyncAllowance) * time.Second):
			w.settingsMu.Lock()
			w.settings.TopicInterestTolerance = topicInterestMap
			w.settingsMu.Unlock()
		}
	}()

	return nil
}

// MaxMessageSize returns the maximum accepted message size.
func (w *Waku) MaxMessageSize() uint32 {
	w.settingsMu.RLock()
	defer w.settingsMu.RUnlock()
	return w.settings.MaxMsgSize
}

// SetMaxMessageSize sets the maximal message size allowed by this node
func (w *Waku) SetMaxMessageSize(size uint32) error {
	if size > common.MaxMessageSize {
		return fmt.Errorf("message size too large [%d>%d]", size, common.MaxMessageSize)
	}
	w.settingsMu.Lock()
	w.settings.MaxMsgSize = size
	w.settingsMu.Unlock()
	return nil
}

// LightClientMode indicates is this node is light client (does not forward any messages)
func (w *Waku) LightClientMode() bool {
	w.settingsMu.RLock()
	defer w.settingsMu.RUnlock()
	return w.settings.LightClient
}

// SetLightClientMode makes node light client (does not forward any messages)
func (w *Waku) SetLightClientMode(v bool) {
	w.settingsMu.Lock()
	w.settings.LightClient = v
	w.settingsMu.Unlock()
}

// LightClientModeConnectionRestricted indicates that connection to light client in light client mode not allowed
func (w *Waku) LightClientModeConnectionRestricted() bool {
	w.settingsMu.RLock()
	defer w.settingsMu.RUnlock()
	return w.settings.RestrictLightClientsConn
}

// PacketRateLimiting returns RateLimits information for packets
func (w *Waku) PacketRateLimits() common.RateLimits {
	if w.rateLimiter == nil {
		return common.RateLimits{}
	}
	return common.RateLimits{
		IPLimits:     uint64(w.rateLimiter.PacketLimitPerSecIP),
		PeerIDLimits: uint64(w.rateLimiter.PacketLimitPerSecPeerID),
	}
}

// BytesRateLimiting returns RateLimits information for bytes
func (w *Waku) BytesRateLimits() common.RateLimits {
	if w.rateLimiter == nil {
		return common.RateLimits{}
	}
	return common.RateLimits{
		IPLimits:     uint64(w.rateLimiter.BytesLimitPerSecIP),
		PeerIDLimits: uint64(w.rateLimiter.BytesLimitPerSecPeerID),
	}
}

// ConfirmationsEnabled returns true if message confirmations are enabled.
func (w *Waku) ConfirmationsEnabled() bool {
	w.settingsMu.RLock()
	defer w.settingsMu.RUnlock()
	return w.settings.EnableConfirmations
}

// CurrentTime returns current time.
func (w *Waku) CurrentTime() time.Time {
	return w.timeSource()
}

// SetTimeSource assigns a particular source of time to a waku object.
func (w *Waku) SetTimeSource(timesource func() time.Time) {
	w.timeSource = timesource
}

// APIs returns the RPC descriptors the Waku implementation offers
func (w *Waku) APIs() []rpc.API {
	return []rpc.API{
		{
			Namespace: v0.Name,
			Version:   v0.VersionStr,
			Service:   NewPublicWakuAPI(w),
			Public:    false,
		},
	}
}

// Protocols returns the waku sub-protocols ran by this particular client.
func (w *Waku) Protocols() []p2p.Protocol {
	return w.protocols
}

// RegisterMailServer registers MailServer interface.
// MailServer will process all the incoming messages with p2pRequestCode.
func (w *Waku) RegisterMailServer(server MailServer) {
	w.mailServer = server
}

// SetRateLimiter registers a rate limiter.
func (w *Waku) RegisterRateLimiter(r *common.PeerRateLimiter) {
	w.rateLimiter = r
}

// RegisterBridge registers a new Bridge that moves envelopes
// between different subprotocols.
// It's important that a bridge is registered before the service
// is started, otherwise, it won't read and propagate envelopes.
func (w *Waku) RegisterBridge(b Bridge) {
	if w.cancelBridge != nil {
		close(w.cancelBridge)
	}
	w.bridge = b
	w.cancelBridge = make(chan struct{})
	w.bridgeWg.Add(1)
	go w.readBridgeLoop()
}

func (w *Waku) readBridgeLoop() {
	defer w.bridgeWg.Done()
	out, _ := w.bridge.Pipe()
	for {
		select {
		case <-w.cancelBridge:
			return
		case env := <-out:
			_, err := w.addAndBridge(env, false, true)
			if err != nil {
				common.BridgeReceivedFailed.Inc()
				w.logger.Warn(
					"failed to add a bridged envelope",
					zap.Binary("ID", env.Hash().Bytes()),
					zap.Error(err),
				)
			} else {
				common.BridgeReceivedSucceed.Inc()
				w.logger.Debug("bridged envelope successfully", zap.Binary("ID", env.Hash().Bytes()))
				w.envelopeFeed.Send(common.EnvelopeEvent{
					Event: common.EventEnvelopeReceived,
					Topic: env.Topic,
					Hash:  env.Hash(),
				})
			}
		}
	}
}

func (w *Waku) SendEnvelopeEvent(event common.EnvelopeEvent) int {
	return w.envelopeFeed.Send(event)
}

// SubscribeEnvelopeEvents subscribes to envelopes feed.
// In order to prevent blocking waku producers events must be amply buffered.
func (w *Waku) SubscribeEnvelopeEvents(events chan<- common.EnvelopeEvent) event.Subscription {
	return w.envelopeFeed.Subscribe(events)
}

func (w *Waku) notifyPeersAboutPowRequirementChange(pow float64) {
	arr := w.getPeers()
	for _, p := range arr {
		err := p.NotifyAboutPowRequirementChange(pow)
		if err != nil {
			// allow one retry
			err = p.NotifyAboutPowRequirementChange(pow)
		}
		if err != nil {
			w.logger.Warn("failed to notify peer about new pow requirement", zap.Binary("peer", p.ID()), zap.Error(err))
		}
	}
}

func (w *Waku) FullNode() bool {
	w.settingsMu.RLock()
	// If full node, nothing to do
	fullNode := w.settings.FullNode
	w.settingsMu.RUnlock()
	return fullNode
}

func (w *Waku) notifyPeersAboutBloomFilterChange(bloom []byte) {

	if w.FullNode() {
		return
	}
	arr := w.getPeers()
	for _, p := range arr {
		err := p.NotifyAboutBloomFilterChange(bloom)
		if err != nil {
			// allow one retry
			err = p.NotifyAboutBloomFilterChange(bloom)
		}
		if err != nil {
			w.logger.Warn("failed to notify peer about new bloom filter change", zap.Binary("peer", p.ID()), zap.Error(err))
		}
	}
}

func (w *Waku) notifyPeersAboutTopicInterestChange(topicInterest []common.TopicType) {
	if w.FullNode() {
		return
	}
	arr := w.getPeers()
	for _, p := range arr {
		err := p.NotifyAboutTopicInterestChange(topicInterest)
		if err != nil {
			// allow one retry
			err = p.NotifyAboutTopicInterestChange(topicInterest)
		}
		if err != nil {
			w.logger.Warn("failed to notify peer about new topic interest", zap.Binary("peer", p.ID()), zap.Error(err))
		}
	}
}

func (w *Waku) getPeers() []common.Peer {
	w.peerMu.Lock()
	arr := make([]common.Peer, len(w.peers))
	i := 0
	for p := range w.peers {
		arr[i] = p
		i++
	}
	w.peerMu.Unlock()
	return arr
}

// getPeer retrieves peer by ID
func (w *Waku) getPeer(peerID []byte) (common.Peer, error) {
	w.peerMu.Lock()
	defer w.peerMu.Unlock()
	for p := range w.peers {
		if bytes.Equal(peerID, p.ID()) {
			return p, nil
		}
	}
	return nil, fmt.Errorf("could not find peer with ID: %x", peerID)
}

// AllowP2PMessagesFromPeer marks specific peer trusted,
// which will allow it to send historic (expired) messages.
func (w *Waku) AllowP2PMessagesFromPeer(peerID []byte) error {
	p, err := w.getPeer(peerID)
	if err != nil {
		return err
	}
	p.SetPeerTrusted(true)
	return nil
}

// RequestHistoricMessages sends a message with p2pRequestCode to a specific peer,
// which is known to implement MailServer interface, and is supposed to process this
// request and respond with a number of peer-to-peer messages (possibly expired),
// which are not supposed to be forwarded any further.
// The waku protocol is agnostic of the format and contents of envelope.
func (w *Waku) RequestHistoricMessages(peerID []byte, envelope *common.Envelope) error {
	return w.RequestHistoricMessagesWithTimeout(peerID, envelope, 0)
}

// RequestHistoricMessagesWithTimeout acts as RequestHistoricMessages but requires to pass a timeout.
// It sends an event EventMailServerRequestExpired after the timeout.
func (w *Waku) RequestHistoricMessagesWithTimeout(peerID []byte, envelope *common.Envelope, timeout time.Duration) error {
	p, err := w.getPeer(peerID)
	if err != nil {
		return err
	}
	p.SetPeerTrusted(true)

	w.envelopeFeed.Send(common.EnvelopeEvent{
		Peer:  p.EnodeID(),
		Topic: envelope.Topic,
		Hash:  envelope.Hash(),
		Event: common.EventMailServerRequestSent,
	})

	err = p.RequestHistoricMessages(envelope)
	if timeout != 0 {
		go w.expireRequestHistoricMessages(p.EnodeID(), envelope.Hash(), timeout)
	}
	return err
}

func (w *Waku) SendMessagesRequest(peerID []byte, request common.MessagesRequest) error {
	if err := request.Validate(); err != nil {
		return err
	}
	p, err := w.getPeer(peerID)
	if err != nil {
		return err
	}
	p.SetPeerTrusted(true)
	if err := p.SendMessagesRequest(request); err != nil {
		return err
	}
	w.envelopeFeed.Send(common.EnvelopeEvent{
		Peer:  p.EnodeID(),
		Hash:  gethcommon.BytesToHash(request.ID),
		Event: common.EventMailServerRequestSent,
	})
	return nil
}

func (w *Waku) expireRequestHistoricMessages(peer enode.ID, hash gethcommon.Hash, timeout time.Duration) {
	timer := time.NewTimer(timeout)
	defer timer.Stop()
	select {
	case <-w.quit:
		return
	case <-timer.C:
		w.envelopeFeed.Send(common.EnvelopeEvent{
			Peer:  peer,
			Hash:  hash,
			Event: common.EventMailServerRequestExpired,
		})
	}
}

func (w *Waku) SendHistoricMessageResponse(peerID []byte, payload []byte) error {
	peer, err := w.getPeer(peerID)
	if err != nil {
		return err
	}
	return peer.SendHistoricMessageResponse(payload)
}

// SendP2PMessage sends a peer-to-peer message to a specific peer.
// It sends one or more envelopes in a single batch.
func (w *Waku) SendP2PMessages(peerID []byte, envelopes ...*common.Envelope) error {
	p, err := w.getPeer(peerID)
	if err != nil {
		return err
	}
	return p.SendP2PMessages(envelopes)
}

// SendRawP2PDirect sends a peer-to-peer message to a specific peer.
// It sends one or more envelopes in a single batch.
func (w *Waku) SendRawP2PDirect(peerID []byte, envelopes ...rlp.RawValue) error {
	p, err := w.getPeer(peerID)
	if err != nil {
		return err
	}
	return p.SendRawP2PDirect(envelopes)
}

// NewKeyPair generates a new cryptographic identity for the client, and injects
// it into the known identities for message decryption. Returns ID of the new key pair.
func (w *Waku) NewKeyPair() (string, error) {
	key, err := crypto.GenerateKey()
	if err != nil || !validatePrivateKey(key) {
		key, err = crypto.GenerateKey() // retry once
	}
	if err != nil {
		return "", err
	}
	if !validatePrivateKey(key) {
		return "", fmt.Errorf("failed to generate valid key")
	}

	id, err := toDeterministicID(hexutil.Encode(crypto.FromECDSAPub(&key.PublicKey)), common.KeyIDSize)
	if err != nil {
		return "", err
	}

	w.keyMu.Lock()
	defer w.keyMu.Unlock()

	if w.privateKeys[id] != nil {
		return "", fmt.Errorf("failed to generate unique ID")
	}
	w.privateKeys[id] = key
	return id, nil
}

// DeleteKeyPair deletes the specified key if it exists.
func (w *Waku) DeleteKeyPair(key string) bool {
	deterministicID, err := toDeterministicID(key, common.KeyIDSize)
	if err != nil {
		return false
	}

	w.keyMu.Lock()
	defer w.keyMu.Unlock()

	if w.privateKeys[deterministicID] != nil {
		delete(w.privateKeys, deterministicID)
		return true
	}
	return false
}

// AddKeyPair imports a asymmetric private key and returns it identifier.
func (w *Waku) AddKeyPair(key *ecdsa.PrivateKey) (string, error) {
	id, err := makeDeterministicID(hexutil.Encode(crypto.FromECDSAPub(&key.PublicKey)), common.KeyIDSize)
	if err != nil {
		return "", err
	}
	if w.HasKeyPair(id) {
		return id, nil // no need to re-inject
	}

	w.keyMu.Lock()
	w.privateKeys[id] = key
	w.keyMu.Unlock()

	return id, nil
}

// SelectKeyPair adds cryptographic identity, and makes sure
// that it is the only private key known to the node.
func (w *Waku) SelectKeyPair(key *ecdsa.PrivateKey) error {
	id, err := makeDeterministicID(hexutil.Encode(crypto.FromECDSAPub(&key.PublicKey)), common.KeyIDSize)
	if err != nil {
		return err
	}

	w.keyMu.Lock()
	defer w.keyMu.Unlock()

	w.privateKeys = make(map[string]*ecdsa.PrivateKey) // reset key store
	w.privateKeys[id] = key

	return nil
}

// DeleteKeyPairs removes all cryptographic identities known to the node
func (w *Waku) DeleteKeyPairs() error {
	w.keyMu.Lock()
	defer w.keyMu.Unlock()

	w.privateKeys = make(map[string]*ecdsa.PrivateKey)

	return nil
}

// HasKeyPair checks if the waku node is configured with the private key
// of the specified public pair.
func (w *Waku) HasKeyPair(id string) bool {
	deterministicID, err := toDeterministicID(id, common.KeyIDSize)
	if err != nil {
		return false
	}

	w.keyMu.RLock()
	defer w.keyMu.RUnlock()
	return w.privateKeys[deterministicID] != nil
}

// GetPrivateKey retrieves the private key of the specified identity.
func (w *Waku) GetPrivateKey(id string) (*ecdsa.PrivateKey, error) {
	deterministicID, err := toDeterministicID(id, common.KeyIDSize)
	if err != nil {
		return nil, err
	}

	w.keyMu.RLock()
	defer w.keyMu.RUnlock()
	key := w.privateKeys[deterministicID]
	if key == nil {
		return nil, fmt.Errorf("invalid id")
	}
	return key, nil
}

// GenerateSymKey generates a random symmetric key and stores it under id,
// which is then returned. Will be used in the future for session key exchange.
func (w *Waku) GenerateSymKey() (string, error) {
	key, err := common.GenerateSecureRandomData(common.AESKeyLength)
	if err != nil {
		return "", err
	} else if !common.ValidateDataIntegrity(key, common.AESKeyLength) {
		return "", fmt.Errorf("error in GenerateSymKey: crypto/rand failed to generate random data")
	}

	id, err := common.GenerateRandomID()
	if err != nil {
		return "", fmt.Errorf("failed to generate ID: %s", err)
	}

	w.keyMu.Lock()
	defer w.keyMu.Unlock()

	if w.symKeys[id] != nil {
		return "", fmt.Errorf("failed to generate unique ID")
	}
	w.symKeys[id] = key
	return id, nil
}

// AddSymKey stores the key with a given id.
func (w *Waku) AddSymKey(id string, key []byte) (string, error) {
	deterministicID, err := toDeterministicID(id, common.KeyIDSize)
	if err != nil {
		return "", err
	}

	w.keyMu.Lock()
	defer w.keyMu.Unlock()

	if w.symKeys[deterministicID] != nil {
		return "", fmt.Errorf("key already exists: %v", id)
	}
	w.symKeys[deterministicID] = key
	return deterministicID, nil
}

// AddSymKeyDirect stores the key, and returns its id.
func (w *Waku) AddSymKeyDirect(key []byte) (string, error) {
	if len(key) != common.AESKeyLength {
		return "", fmt.Errorf("wrong key size: %d", len(key))
	}

	id, err := common.GenerateRandomID()
	if err != nil {
		return "", fmt.Errorf("failed to generate ID: %s", err)
	}

	w.keyMu.Lock()
	defer w.keyMu.Unlock()

	if w.symKeys[id] != nil {
		return "", fmt.Errorf("failed to generate unique ID")
	}
	w.symKeys[id] = key
	return id, nil
}

// AddSymKeyFromPassword generates the key from password, stores it, and returns its id.
func (w *Waku) AddSymKeyFromPassword(password string) (string, error) {
	id, err := common.GenerateRandomID()
	if err != nil {
		return "", fmt.Errorf("failed to generate ID: %s", err)
	}
	if w.HasSymKey(id) {
		return "", fmt.Errorf("failed to generate unique ID")
	}

	// kdf should run no less than 0.1 seconds on an average computer,
	// because it's an once in a session experience
	derived := pbkdf2.Key([]byte(password), nil, 65356, common.AESKeyLength, sha256.New)

	w.keyMu.Lock()
	defer w.keyMu.Unlock()

	// double check is necessary, because deriveKeyMaterial() is very slow
	if w.symKeys[id] != nil {
		return "", fmt.Errorf("critical error: failed to generate unique ID")
	}
	w.symKeys[id] = derived
	return id, nil
}

// HasSymKey returns true if there is a key associated with the given id.
// Otherwise returns false.
func (w *Waku) HasSymKey(id string) bool {
	w.keyMu.RLock()
	defer w.keyMu.RUnlock()
	return w.symKeys[id] != nil
}

// DeleteSymKey deletes the key associated with the name string if it exists.
func (w *Waku) DeleteSymKey(id string) bool {
	w.keyMu.Lock()
	defer w.keyMu.Unlock()
	if w.symKeys[id] != nil {
		delete(w.symKeys, id)
		return true
	}
	return false
}

// GetSymKey returns the symmetric key associated with the given id.
func (w *Waku) GetSymKey(id string) ([]byte, error) {
	w.keyMu.RLock()
	defer w.keyMu.RUnlock()
	if w.symKeys[id] != nil {
		return w.symKeys[id], nil
	}
	return nil, fmt.Errorf("non-existent key ID")
}

// Subscribe installs a new message handler used for filtering, decrypting
// and subsequent storing of incoming messages.
func (w *Waku) Subscribe(f *common.Filter) (string, error) {
	s, err := w.filters.Install(f)
	if err != nil {
		return s, err
	}

	err = w.updateSettingsForFilter(f)
	if err != nil {
		w.filters.Uninstall(s)
		return s, err
	}
	return s, nil
}

func (w *Waku) updateSettingsForFilter(f *common.Filter) error {
	w.settingsMu.RLock()
	topicInterestMode := !w.settings.BloomFilterMode
	w.settingsMu.RUnlock()

	if topicInterestMode {
		err := w.updateTopicInterest(f)
		if err != nil {
			return err
		}
	} else {
		err := w.updateBloomFilter(f)
		if err != nil {
			return err
		}
	}
	return nil
}

// updateBloomFilter recalculates the new value of bloom filter,
// and informs the peers if necessary.
func (w *Waku) updateBloomFilter(f *common.Filter) error {
	aggregate := make([]byte, common.BloomFilterSize)
	for _, t := range f.Topics {
		top := common.BytesToTopic(t)
		b := top.ToBloom()
		aggregate = addBloom(aggregate, b)
	}

	if !common.BloomFilterMatch(w.BloomFilter(), aggregate) {
		// existing bloom filter must be updated
		aggregate = addBloom(w.BloomFilter(), aggregate)
		return w.SetBloomFilter(aggregate)
	}
	return nil
}

// GetFilter returns the filter by id.
func (w *Waku) GetFilter(id string) *common.Filter {
	return w.filters.Get(id)
}

// Unsubscribe removes an installed message handler.
// TODO: This does not update the bloom filter, but does update
// the topic interest map
func (w *Waku) Unsubscribe(id string) error {
	ok := w.filters.Uninstall(id)
	if !ok {
		return fmt.Errorf("failed to unsubscribe: invalid ID '%s'", id)
	}
	if !w.settings.BloomFilterMode {
		return w.SetTopicInterest(w.filters.AllTopics())
	}
	return nil
}

// Unsubscribe removes an installed message handler.
// TODO: This does not update the bloom filter, but does update
// the topic interest map
func (w *Waku) UnsubscribeMany(ids []string) error {
	for _, id := range ids {
		w.logger.Debug("cleaning up filter", zap.String("id", id))
		ok := w.filters.Uninstall(id)
		if !ok {
			w.logger.Warn("could not remove filter with id", zap.String("id", id))
		}
	}
	if !w.settings.BloomFilterMode {
		return w.SetTopicInterest(w.filters.AllTopics())
	}
	return nil
}

// Send injects a message into the waku send queue, to be distributed in the
// network in the coming cycles.
func (w *Waku) Send(envelope *common.Envelope) error {
	w.logger.Debug("send: sending envelope", zap.String("hash", envelope.Hash().String()))
	ok, err := w.add(envelope, false)
	if err == nil && !ok {
		return fmt.Errorf("failed to add envelope")
	}
	return err
}

// Start implements node.Service, starting the background data propagation thread
// of the Waku protocol.
func (w *Waku) Start() error {
	go w.update()

	numCPU := runtime.NumCPU()
	for i := 0; i < numCPU; i++ {
		go w.processQueue()
	}
	go w.processP2P()

	return nil
}

// Stop implements node.Service, stopping the background data propagation thread
// of the Waku protocol.
func (w *Waku) Stop() error {
	if w.cancelBridge != nil {
		close(w.cancelBridge)
		w.cancelBridge = nil
		w.bridgeWg.Wait()
	}
	close(w.quit)
	return nil
}

func (w *Waku) handlePeerV0(p2pPeer *p2p.Peer, rw p2p.MsgReadWriter) error {
	return w.HandlePeer(v0.NewPeer(w, p2pPeer, rw, w.logger.Named("waku/peerv0"), w.stats), rw)
}

func (w *Waku) handlePeerV1(p2pPeer *p2p.Peer, rw p2p.MsgReadWriter) error {
	return w.HandlePeer(v1.NewPeer(w, p2pPeer, rw, w.logger.Named("waku/peerv1"), w.stats), rw)
}

// HandlePeer is called by the underlying P2P layer when the waku sub-protocol
// connection is negotiated.
func (w *Waku) HandlePeer(peer common.Peer, rw p2p.MsgReadWriter) error {
	w.peerMu.Lock()
	w.peers[peer] = struct{}{}
	w.peerMu.Unlock()

	w.logger.Info("handling peer", zap.String("peerID", types.EncodeHex(peer.ID())))

	defer func() {
		w.peerMu.Lock()
		delete(w.peers, peer)
		w.peerMu.Unlock()
	}()

	if err := peer.Start(); err != nil {
		return err
	}
	defer peer.Stop()

	if w.rateLimiter != nil {
		runLoop := func(out p2p.MsgReadWriter) error {
			peer.SetRWWriter(out)
			err := peer.Run()
			w.logger.Info("handled peer", zap.String("peerID", types.EncodeHex(peer.ID())), zap.Error(err))
			return err
		}
		return w.rateLimiter.Decorate(peer, rw, runLoop)
	}

	err := peer.Run()
	w.logger.Info("handled peer", zap.String("peerID", types.EncodeHex(peer.ID())), zap.Error(err))
	return err
}

func (w *Waku) softBlacklisted(peerID string) bool {
	w.settingsMu.RLock()
	defer w.settingsMu.RUnlock()
	return w.settings.SoftBlacklistedPeerIDs[peerID]
}

func (w *Waku) OnNewEnvelopes(envelopes []*common.Envelope, peer common.Peer) ([]common.EnvelopeError, error) {
	envelopeErrors := make([]common.EnvelopeError, 0)
	peerID := types.EncodeHex(peer.ID())
	w.logger.Debug("received new envelopes", zap.Int("count", len(envelopes)), zap.String("peer", peerID))
	trouble := false

	if w.softBlacklisted(peerID) {
		w.logger.Debug("peer is soft blacklisted", zap.String("peer", peerID))
		return nil, nil
	}

	for _, env := range envelopes {
		w.logger.Debug("received new envelope", zap.String("peer", peerID), zap.String("hash", env.Hash().Hex()))
		cached, err := w.add(env, w.LightClientMode())
		if err != nil {
			_, isTimeSyncError := err.(common.TimeSyncError)
			if !isTimeSyncError {
				trouble = true
				w.logger.Info("invalid envelope received", zap.String("peer", types.EncodeHex(peer.ID())), zap.Error(err))
			}
			envelopeErrors = append(envelopeErrors, common.ErrorToEnvelopeError(env.Hash(), err))
		} else if cached {
			peer.Mark(env)
		}

		w.envelopeFeed.Send(common.EnvelopeEvent{
			Event: common.EventEnvelopeReceived,
			Topic: env.Topic,
			Hash:  env.Hash(),
			Peer:  peer.EnodeID(),
		})
		common.EnvelopesValidatedCounter.Inc()
	}

	if trouble {
		return envelopeErrors, errors.New("received invalid envelope")
	}
	return envelopeErrors, nil
}

func (w *Waku) OnNewP2PEnvelopes(envelopes []*common.Envelope) error {
	for _, envelope := range envelopes {
		w.postP2P(envelope)
	}
	return nil
}

func (w *Waku) Mailserver() bool {
	return w.mailServer != nil
}

func (w *Waku) OnMessagesRequest(request common.MessagesRequest, p common.Peer) error {
	w.mailServer.Deliver(p.ID(), request)
	return nil
}

func (w *Waku) OnDeprecatedMessagesRequest(request *common.Envelope, p common.Peer) error {
	w.mailServer.DeliverMail(p.ID(), request)
	return nil
}

func (w *Waku) OnP2PRequestCompleted(payload []byte, p common.Peer) error {
	msEvent, err := CreateMailServerEvent(p.EnodeID(), payload)
	if err != nil {
		return fmt.Errorf("invalid p2p request complete payload: %v", err)
	}

	w.postP2P(*msEvent)
	return nil
}

func (w *Waku) OnMessagesResponse(response common.MessagesResponse, p common.Peer) error {
	w.envelopeFeed.Send(common.EnvelopeEvent{
		Batch: response.Hash,
		Event: common.EventBatchAcknowledged,
		Peer:  p.EnodeID(),
		Data:  response.Errors,
	})

	return nil
}

func (w *Waku) OnBatchAcknowledged(batchHash gethcommon.Hash, p common.Peer) error {
	w.envelopeFeed.Send(common.EnvelopeEvent{
		Batch: batchHash,
		Event: common.EventBatchAcknowledged,
		Peer:  p.EnodeID(),
	})
	return nil
}

func (w *Waku) add(envelope *common.Envelope, isP2P bool) (bool, error) {
	return w.addAndBridge(envelope, isP2P, false)
}

func (w *Waku) bloomMatch(envelope *common.Envelope) (bool, error) {
	if !common.BloomFilterMatch(w.BloomFilter(), envelope.Bloom()) {
		// maybe the value was recently changed, and the peers did not adjust yet.
		// in this case the previous value is retrieved by BloomFilterTolerance()
		// for a short period of peer synchronization.
		if !common.BloomFilterMatch(w.BloomFilterTolerance(), envelope.Bloom()) {
			common.EnvelopesCacheFailedCounter.WithLabelValues("no_bloom_match").Inc()
			return false, fmt.Errorf("envelope does not match bloom filter, hash=[%v], bloom: \n%x \n%x \n%x",
				envelope.Hash().Hex(), w.BloomFilter(), envelope.Bloom(), envelope.Topic)
		}
	}
	return true, nil
}

func (w *Waku) topicInterestMatch(envelope *common.Envelope) (bool, error) {
	w.settingsMu.RLock()
	defer w.settingsMu.RUnlock()
	if w.settings.TopicInterest == nil {
		return false, nil
	}
	if !w.settings.TopicInterest[envelope.Topic] {
		if !w.settings.TopicInterestTolerance[envelope.Topic] {
			common.EnvelopesCacheFailedCounter.WithLabelValues("no_topic_interest_match").Inc()
			return false, fmt.Errorf("envelope does not match topic interest, hash=[%v], bloom: \n%x \n%x",
				envelope.Hash().Hex(), envelope.Bloom(), envelope.Topic)

		}
	}

	return true, nil
}

func (w *Waku) topicInterestOrBloomMatch(envelope *common.Envelope) (bool, error) {
	if w.FullNode() {
		return true, nil
	}
	w.settingsMu.RLock()
	topicInterestMode := !w.settings.BloomFilterMode
	w.settingsMu.RUnlock()

	if topicInterestMode {
		match, err := w.topicInterestMatch(envelope)
		if err != nil {
			return false, err
		}
		if match {
			return true, nil
		}
	}
	return w.bloomMatch(envelope)
}

func (w *Waku) SetBloomFilterMode(mode bool) {
	w.settingsMu.Lock()
	w.settings.BloomFilterMode = mode
	w.settingsMu.Unlock()
	// Recalculate and notify topic interest or bloom, currently not implemented
}

func (w *Waku) SetFullNode(set bool) {
	w.settingsMu.Lock()
	w.settings.FullNode = set
	w.settingsMu.Unlock()

	// We advertise the topic interest if full node has been disabled
	// or bloom filter if enabled, as that's how we indicate to a peer we are a full node or not
	if set {
		w.notifyPeersAboutBloomFilterChange(w.BloomFilter())
	} else {
		w.notifyPeersAboutTopicInterestChange(w.TopicInterest())
	}
}

// addEnvelope adds an envelope to the envelope map, used for sending
func (w *Waku) addEnvelope(envelope *common.Envelope) {

	hash := envelope.Hash()

	w.poolMu.Lock()
	w.envelopes[hash] = envelope
	if w.expirations[envelope.Expiry] == nil {
		w.expirations[envelope.Expiry] = mapset.NewThreadUnsafeSet()
	}
	if !w.expirations[envelope.Expiry].Contains(hash) {
		w.expirations[envelope.Expiry].Add(hash)
	}
	w.poolMu.Unlock()
}

// addAndBridge inserts a new envelope into the message pool to be distributed within the
// waku network. It also inserts the envelope into the expiration pool at the
// appropriate time-stamp. In case of error, connection should be dropped.
// param isP2P indicates whether the message is peer-to-peer (should not be forwarded).
func (w *Waku) addAndBridge(envelope *common.Envelope, isP2P bool, bridged bool) (bool, error) {
	now := uint32(w.timeSource().Unix())
	sent := envelope.Expiry - envelope.TTL
	logger := w.logger.With(zap.String("hash", envelope.Hash().String()), zap.String("site", "addAndBridge"), zap.String("topic", envelope.Topic.String()), zap.Bool("isP2P", isP2P))

	logger.Debug("addAndBridge: processing envelope")

	common.EnvelopesReceivedCounter.Inc()
	if sent > now {
		if sent-common.DefaultSyncAllowance > now {
			common.EnvelopesCacheFailedCounter.WithLabelValues("in_future").Inc()
			logger.Warn("envelope created in the future")
			return false, common.TimeSyncError(errors.New("envelope from future"))
		}
		// recalculate PoW, adjusted for the time difference, plus one second for latency
		envelope.CalculatePoW(sent - now + 1)
	}

	if envelope.Expiry < now {
		if envelope.Expiry+common.DefaultSyncAllowance*2 < now {
			common.EnvelopesCacheFailedCounter.WithLabelValues("very_old").Inc()
			logger.Warn("very old envelope")
			return false, common.TimeSyncError(errors.New("very old envelope"))
		}
		logger.Debug("expired envelope dropped")
		common.EnvelopesCacheFailedCounter.WithLabelValues("expired").Inc()
		return false, nil // drop envelope without error
	}

	if uint32(envelope.Size()) > w.MaxMessageSize() {
		common.EnvelopesCacheFailedCounter.WithLabelValues("oversized").Inc()
		return false, fmt.Errorf("huge messages are not allowed [%s][%d][%d]", envelope.Hash().String(), envelope.Size(), w.MaxMessageSize())
	}

	if envelope.PoW() < w.MinPow() {
		// maybe the value was recently changed, and the peers did not adjust yet.
		// in this case the previous value is retrieved by MinPowTolerance()
		// for a short period of peer synchronization.
		if envelope.PoW() < w.MinPowTolerance() {
			common.EnvelopesCacheFailedCounter.WithLabelValues("low_pow").Inc()
			return false, fmt.Errorf("envelope with low PoW received: PoW=%f, hash=[%s]", envelope.PoW(), envelope.Hash().String())
		}
	}

	match, err := w.topicInterestOrBloomMatch(envelope)
	if err != nil {
		return false, err
	}

	if !match {
		logger.Debug("addAndBridge: no matches for envelope")
		return false, nil
	}

	hash := envelope.Hash()

	w.poolMu.Lock()
	_, alreadyCached := w.envelopes[hash]
	w.poolMu.Unlock()
	if !alreadyCached {
		logger.Debug("addAndBridge: adding envelope")
		w.addEnvelope(envelope)
	}

	if alreadyCached {
		logger.Debug("addAndBridge: already cached")
		common.EnvelopesCachedCounter.WithLabelValues("hit").Inc()
	} else {
		common.EnvelopesCachedCounter.WithLabelValues("miss").Inc()
		common.EnvelopesSizeMeter.Observe(float64(envelope.Size()))
		w.postEvent(envelope, isP2P) // notify the local node about the new message
		if w.mailServer != nil {
			w.mailServer.Archive(envelope)
			w.envelopeFeed.Send(common.EnvelopeEvent{
				Topic: envelope.Topic,
				Hash:  envelope.Hash(),
				Event: common.EventMailServerEnvelopeArchived,
			})
		}
		// Bridge only envelopes that are not p2p messages.
		// In particular, if a node is a lightweight node,
		// it should not bridge any envelopes.
		if !isP2P && !bridged && w.bridge != nil {
			logger.Debug("bridging envelope from Waku")
			_, in := w.bridge.Pipe()
			in <- envelope
			common.BridgeSent.Inc()
		}
	}
	return true, nil
}

func (w *Waku) postP2P(event interface{}) {
	w.p2pMsgQueue <- event
}

// postEvent queues the message for further processing.
func (w *Waku) postEvent(envelope *common.Envelope, isP2P bool) {
	if isP2P {
		w.postP2P(envelope)
	} else {
		w.msgQueue <- envelope
	}
}

// processQueue delivers the messages to the watchers during the lifetime of the waku node.
func (w *Waku) processQueue() {
	for {
		select {
		case <-w.quit:
			return
		case e := <-w.msgQueue:
			w.filters.NotifyWatchers(e, false)
			w.envelopeFeed.Send(common.EnvelopeEvent{
				Topic: e.Topic,
				Hash:  e.Hash(),
				Event: common.EventEnvelopeAvailable,
			})
		}
	}
}

func (w *Waku) processP2P() {
	for {
		select {
		case <-w.quit:
			return
		case e := <-w.p2pMsgQueue:
			switch evn := e.(type) {
			case *common.Envelope:
				// We need to insert it first, and then remove it if not matched,
				// as messages are processed asynchronously
				w.p2pMsgIDsMu.Lock()
				w.p2pMsgIDs[evn.Hash()] = true
				w.p2pMsgIDsMu.Unlock()

				matched := w.filters.NotifyWatchers(evn, true)

				// If not matched we remove it
				if !matched {
					w.p2pMsgIDsMu.Lock()
					delete(w.p2pMsgIDs, evn.Hash())
					w.p2pMsgIDsMu.Unlock()
				}

				w.envelopeFeed.Send(common.EnvelopeEvent{
					Topic: evn.Topic,
					Hash:  evn.Hash(),
					Event: common.EventEnvelopeAvailable,
				})
			case common.EnvelopeEvent:
				w.envelopeFeed.Send(evn)
			}
		}
	}
}

// update loops until the lifetime of the waku node, updating its internal
// state by expiring stale messages from the pool.
func (w *Waku) update() {
	// Start a ticker to check for expirations
	expire := time.NewTicker(common.ExpirationCycle)

	// Repeat updates until termination is requested
	for {
		select {
		case <-expire.C:
			w.expire()

		case <-w.quit:
			return
		}
	}
}

// expire iterates over all the expiration timestamps, removing all stale
// messages from the pools.
func (w *Waku) expire() {
	w.poolMu.Lock()
	defer w.poolMu.Unlock()
	logger := w.logger.With(zap.String("site", "expire"))

	now := uint32(w.timeSource().Unix())
	for expiry, hashSet := range w.expirations {
		if expiry < now {
			// Dump all expired messages and remove timestamp
			hashSet.Each(func(v interface{}) bool {
				logger.Debug("expiring envelope", zap.String("hash", v.(gethcommon.Hash).String()))
				delete(w.envelopes, v.(gethcommon.Hash))
				common.EnvelopesCachedCounter.WithLabelValues("clear").Inc()
				w.envelopeFeed.Send(common.EnvelopeEvent{
					Hash:  v.(gethcommon.Hash),
					Event: common.EventEnvelopeExpired,
				})
				return false
			})
			w.expirations[expiry].Clear()
			delete(w.expirations, expiry)
		}
	}
}

// Envelopes retrieves all the messages currently pooled by the node.
func (w *Waku) Envelopes() []*common.Envelope {
	w.poolMu.RLock()
	defer w.poolMu.RUnlock()

	all := make([]*common.Envelope, 0, len(w.envelopes))
	for _, envelope := range w.envelopes {
		all = append(all, envelope)
	}
	return all
}

// GetEnvelope retrieves an envelope from the message queue by its hash.
// It returns nil if the envelope can not be found.
func (w *Waku) GetEnvelope(hash gethcommon.Hash) *common.Envelope {
	w.poolMu.RLock()
	defer w.poolMu.RUnlock()
	return w.envelopes[hash]
}

func (w *Waku) Version() uint {
	return 1
}

// isEnvelopeCached checks if envelope with specific hash has already been received and cached.
func (w *Waku) IsEnvelopeCached(hash gethcommon.Hash) bool {
	w.poolMu.Lock()
	defer w.poolMu.Unlock()

	_, exist := w.envelopes[hash]
	return exist
}

func (w *Waku) ProcessingP2PMessages() bool {
	w.p2pMsgIDsMu.Lock()
	defer w.p2pMsgIDsMu.Unlock()
	return len(w.p2pMsgIDs) != 0
}

func (w *Waku) MarkP2PMessageAsProcessed(hash gethcommon.Hash) {
	w.p2pMsgIDsMu.Lock()
	defer w.p2pMsgIDsMu.Unlock()
	delete(w.p2pMsgIDs, hash)
}

func (w *Waku) ClearEnvelopesCache() {
	w.poolMu.Lock()
	defer w.poolMu.Unlock()
	w.envelopes = make(map[gethcommon.Hash]*common.Envelope)
}

func (w *Waku) Clean() error {
	w.poolMu.Lock()
	defer w.poolMu.Unlock()
	w.envelopes = make(map[gethcommon.Hash]*common.Envelope)
	for _, f := range w.filters.All() {
		f.Messages = common.NewMemoryMessageStore()
	}

	return nil
}

// validatePrivateKey checks the format of the given private key.
func validatePrivateKey(k *ecdsa.PrivateKey) bool {
	if k == nil || k.D == nil || k.D.Sign() == 0 {
		return false
	}
	return common.ValidatePublicKey(&k.PublicKey)
}

// makeDeterministicID generates a deterministic ID, based on a given input
func makeDeterministicID(input string, keyLen int) (id string, err error) {
	buf := pbkdf2.Key([]byte(input), nil, 4096, keyLen, sha256.New)
	if !common.ValidateDataIntegrity(buf, common.KeyIDSize) {
		return "", fmt.Errorf("error in GenerateDeterministicID: failed to generate key")
	}
	id = gethcommon.Bytes2Hex(buf)
	return id, err
}

// toDeterministicID reviews incoming id, and transforms it to format
// expected internally be private key store. Originally, public keys
// were used as keys, now random keys are being used. And in order to
// make it easier to consume, we now allow both random IDs and public
// keys to be passed.
func toDeterministicID(id string, expectedLen int) (string, error) {
	if len(id) != (expectedLen * 2) { // we received hex key, so number of chars in id is doubled
		var err error
		id, err = makeDeterministicID(id, expectedLen)
		if err != nil {
			return "", err
		}
	}

	return id, nil
}

func addBloom(a, b []byte) []byte {
	c := make([]byte, common.BloomFilterSize)
	for i := 0; i < common.BloomFilterSize; i++ {
		c[i] = a[i] | b[i]
	}
	return c
}
