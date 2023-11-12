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

package wakuv2

import (
	"context"
	"crypto/ecdsa"
	"crypto/sha256"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"net"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/protocol/identify"
	"github.com/multiformats/go-multiaddr"
	"google.golang.org/protobuf/proto"

	"go.uber.org/zap"

	mapset "github.com/deckarep/golang-set"
	"golang.org/x/crypto/pbkdf2"

	gethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/event"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/ethereum/go-ethereum/rpc"

	"github.com/libp2p/go-libp2p"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/metrics"

	"github.com/waku-org/go-waku/waku/v2/dnsdisc"
	wps "github.com/waku-org/go-waku/waku/v2/peerstore"
	"github.com/waku-org/go-waku/waku/v2/protocol"
	"github.com/waku-org/go-waku/waku/v2/protocol/filter"
	"github.com/waku-org/go-waku/waku/v2/protocol/lightpush"
	"github.com/waku-org/go-waku/waku/v2/protocol/peer_exchange"
	"github.com/waku-org/go-waku/waku/v2/protocol/relay"

	ethdisc "github.com/ethereum/go-ethereum/p2p/dnsdisc"

	"github.com/status-im/status-go/connection"
	"github.com/status-im/status-go/eth-node/types"
	"github.com/status-im/status-go/timesource"
	"github.com/status-im/status-go/wakuv2/common"
	"github.com/status-im/status-go/wakuv2/persistence"

	node "github.com/waku-org/go-waku/waku/v2/node"
	"github.com/waku-org/go-waku/waku/v2/protocol/pb"
	"github.com/waku-org/go-waku/waku/v2/protocol/store"
	storepb "github.com/waku-org/go-waku/waku/v2/protocol/store/pb"
)

const messageQueueLimit = 1024
const requestTimeout = 30 * time.Second
const bootnodesQueryBackoffMs = 200
const bootnodesMaxRetries = 7

type settings struct {
	LightClient         bool             // Indicates if the node is a light client
	MinPeersForRelay    int              // Indicates the minimum number of peers required for using Relay Protocol
	MinPeersForFilter   int              // Indicates the minimum number of peers required for using Filter Protocol
	MaxMsgSize          uint32           // Maximal message length allowed by the waku node
	EnableConfirmations bool             // Enable sending message confirmations
	PeerExchange        bool             // Enable peer exchange
	DiscoveryLimit      int              // Indicates the number of nodes to discover
	Nameserver          string           // Optional nameserver to use for dns discovery
	Resolver            ethdisc.Resolver // Optional resolver to use for dns discovery
	EnableDiscV5        bool             // Indicates whether discv5 is enabled or not
	DefaultPubsubTopic  string           // Pubsub topic to be used by default for messages that do not have a topic assigned (depending whether sharding is used or not)
	Options             []node.WakuNodeOption
	SkipPublishToTopic  bool // used in testing
}

type ITelemetryClient interface {
	PushReceivedEnvelope(*protocol.Envelope)
}

// Waku represents a dark communication interface through the Ethereum
// network, using its very own P2P communication layer.
type Waku struct {
	node            *node.WakuNode // reference to a libp2p waku node
	identifyService identify.IDService
	appDB           *sql.DB

	dnsAddressCache     map[string][]dnsdisc.DiscoveredNode // Map to store the multiaddresses returned by dns discovery
	dnsAddressCacheLock *sync.RWMutex                       // lock to handle access to the map

	// Filter-related
	filters       *common.Filters // Message filters installed with Subscribe function
	filterManager *FilterManager

	privateKeys map[string]*ecdsa.PrivateKey // Private key storage
	symKeys     map[string][]byte            // Symmetric key storage
	keyMu       sync.RWMutex                 // Mutex associated with key stores

	envelopes   map[gethcommon.Hash]*common.ReceivedMessage // Pool of envelopes currently tracked by this node
	expirations map[uint32]mapset.Set                       // Message expiration pool
	poolMu      sync.RWMutex                                // Mutex to sync the message and expiration pools

	bandwidthCounter *metrics.BandwidthCounter

	protectedTopicStore *persistence.ProtectedTopicsStore
	sendQueue           chan *protocol.Envelope
	msgQueue            chan *common.ReceivedMessage // Message queue for waku messages that havent been decoded

	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	cfg        *Config
	settings   settings     // Holds configuration settings that can be dynamically changed
	settingsMu sync.RWMutex // Mutex to sync the settings access

	envelopeFeed event.Feed

	storeMsgIDs   map[gethcommon.Hash]bool // Map of the currently processing ids
	storeMsgIDsMu sync.RWMutex

	connStatusChan          chan node.ConnStatus
	connStatusSubscriptions map[string]*types.ConnStatusSubscription
	connStatusMu            sync.Mutex

	logger *zap.Logger

	// NTP Synced timesource
	timesource *timesource.NTPTimeSource

	// seededBootnodesForDiscV5 indicates whether we manage to retrieve discovery
	// bootnodes successfully
	seededBootnodesForDiscV5 bool

	// offline indicates whether we have detected connectivity
	offline bool

	// connectionChanged is channel that notifies when connectivity has changed
	connectionChanged chan struct{}

	// discV5BootstrapNodes is the ENR to be used to fetch bootstrap nodes for discovery
	discV5BootstrapNodes []string

	onHistoricMessagesRequestFailed func([]byte, peer.ID, error)
	onPeerStats                     func(types.ConnStatus)

	statusTelemetryClient ITelemetryClient
}

func (w *Waku) SetStatusTelemetryClient(client ITelemetryClient) {
	w.statusTelemetryClient = client
}

// New creates a WakuV2 client ready to communicate through the LibP2P network.
func New(nodeKey string, fleet string, cfg *Config, logger *zap.Logger, appDB *sql.DB, ts *timesource.NTPTimeSource, onHistoricMessagesRequestFailed func([]byte, peer.ID, error), onPeerStats func(types.ConnStatus)) (*Waku, error) {
	var err error
	if logger == nil {
		logger, err = zap.NewDevelopment()
		if err != nil {
			return nil, err
		}
	}

	if ts == nil {
		ts = timesource.Default()
	}

	cfg = setDefaults(cfg)

	logger.Info("starting wakuv2 with config", zap.Any("config", cfg))

	ctx, cancel := context.WithCancel(context.Background())

	waku := &Waku{
		appDB:                           appDB,
		cfg:                             cfg,
		privateKeys:                     make(map[string]*ecdsa.PrivateKey),
		symKeys:                         make(map[string][]byte),
		envelopes:                       make(map[gethcommon.Hash]*common.ReceivedMessage),
		expirations:                     make(map[uint32]mapset.Set),
		msgQueue:                        make(chan *common.ReceivedMessage, messageQueueLimit),
		sendQueue:                       make(chan *protocol.Envelope, 1000),
		connStatusChan:                  make(chan node.ConnStatus, 100),
		connStatusSubscriptions:         make(map[string]*types.ConnStatusSubscription),
		ctx:                             ctx,
		cancel:                          cancel,
		wg:                              sync.WaitGroup{},
		dnsAddressCache:                 make(map[string][]dnsdisc.DiscoveredNode),
		dnsAddressCacheLock:             &sync.RWMutex{},
		storeMsgIDs:                     make(map[gethcommon.Hash]bool),
		timesource:                      ts,
		storeMsgIDsMu:                   sync.RWMutex{},
		logger:                          logger,
		discV5BootstrapNodes:            cfg.DiscV5BootstrapNodes,
		onHistoricMessagesRequestFailed: onHistoricMessagesRequestFailed,
		onPeerStats:                     onPeerStats,
	}

	waku.settings = settings{
		MaxMsgSize:        cfg.MaxMessageSize,
		LightClient:       cfg.LightClient,
		MinPeersForRelay:  cfg.MinPeersForRelay,
		MinPeersForFilter: cfg.MinPeersForFilter,
		PeerExchange:      cfg.PeerExchange,
		DiscoveryLimit:    cfg.DiscoveryLimit,
		Nameserver:        cfg.Nameserver,
		Resolver:          cfg.Resolver,
		EnableDiscV5:      cfg.EnableDiscV5,
	}

	waku.settings.DefaultPubsubTopic = cfg.DefaultShardPubsubTopic
	waku.filters = common.NewFilters(waku.settings.DefaultPubsubTopic, waku.logger)
	waku.bandwidthCounter = metrics.NewBandwidthCounter()

	var privateKey *ecdsa.PrivateKey
	if nodeKey != "" {
		privateKey, err = crypto.HexToECDSA(nodeKey)
	} else {
		// If no nodekey is provided, create an ephemeral key
		privateKey, err = crypto.GenerateKey()
	}

	if err != nil {
		return nil, fmt.Errorf("failed to setup the go-waku private key: %v", err)
	}

	hostAddr, err := net.ResolveTCPAddr("tcp", fmt.Sprint(cfg.Host, ":", cfg.Port))
	if err != nil {
		return nil, fmt.Errorf("failed to setup the network interface: %v", err)
	}

	if cfg.KeepAliveInterval == 0 {
		cfg.KeepAliveInterval = DefaultConfig.KeepAliveInterval
	}

	libp2pOpts := node.DefaultLibP2POptions
	libp2pOpts = append(libp2pOpts, libp2p.BandwidthReporter(waku.bandwidthCounter))
	libp2pOpts = append(libp2pOpts, libp2p.NATPortMap())

	opts := []node.WakuNodeOption{
		node.WithLibP2POptions(libp2pOpts...),
		node.WithPrivateKey(privateKey),
		node.WithHostAddress(hostAddr),
		node.WithConnectionStatusChannel(waku.connStatusChan),
		node.WithKeepAlive(time.Duration(cfg.KeepAliveInterval) * time.Second),
		node.WithMaxPeerConnections(cfg.DiscoveryLimit),
		node.WithLogger(logger),
		node.WithLogLevel(logger.Level()),
		node.WithClusterID(cfg.ClusterID),
		node.WithMaxMsgSize(1024 * 1024),
	}

	if cfg.EnableDiscV5 {
		bootnodes, err := waku.getDiscV5BootstrapNodes(waku.ctx, cfg.DiscV5BootstrapNodes)
		if err != nil {
			logger.Error("failed to get bootstrap nodes", zap.Error(err))
			return nil, err
		}

		opts = append(opts, node.WithDiscoveryV5(uint(cfg.UDPPort), bootnodes, cfg.AutoUpdate))
	}

	if cfg.LightClient {
		opts = append(opts, node.WithWakuFilterLightNode())
	} else {
		relayOpts := []pubsub.Option{
			pubsub.WithMaxMessageSize(int(waku.settings.MaxMsgSize)),
		}

		if waku.logger.Level() == zap.DebugLevel {
			relayOpts = append(relayOpts, pubsub.WithEventTracer(waku))
		}

		opts = append(opts, node.WithWakuRelayAndMinPeers(waku.settings.MinPeersForRelay, relayOpts...))
	}

	if cfg.EnableStore {
		if appDB == nil {
			return nil, errors.New("appDB is required for store")
		}
		opts = append(opts, node.WithWakuStore())
		dbStore, err := persistence.NewDBStore(logger, persistence.WithDB(appDB), persistence.WithRetentionPolicy(cfg.StoreCapacity, time.Duration(cfg.StoreSeconds)*time.Second))
		if err != nil {
			return nil, err
		}
		opts = append(opts, node.WithMessageProvider(dbStore))
	}

	if !cfg.LightClient {
		opts = append(opts, node.WithWakuFilterFullNode(filter.WithMaxSubscribers(20)))
		opts = append(opts, node.WithLightPush())
	}

	if appDB != nil {
		waku.protectedTopicStore, err = persistence.NewProtectedTopicsStore(logger, appDB)
		if err != nil {
			return nil, err
		}
	}

	waku.settings.Options = opts
	waku.logger.Info("setup the go-waku node successfully")

	return waku, nil
}

func (w *Waku) SubscribeToConnStatusChanges() *types.ConnStatusSubscription {
	w.connStatusMu.Lock()
	defer w.connStatusMu.Unlock()
	subscription := types.NewConnStatusSubscription()
	w.connStatusSubscriptions[subscription.ID] = subscription
	return subscription
}

func (w *Waku) getDiscV5BootstrapNodes(ctx context.Context, addresses []string) ([]*enode.Node, error) {
	wg := sync.WaitGroup{}
	mu := sync.Mutex{}
	var result []*enode.Node

	retrieveENR := func(d dnsdisc.DiscoveredNode, wg *sync.WaitGroup) {
		mu.Lock()
		defer mu.Unlock()
		defer wg.Done()
		if d.ENR != nil {
			result = append(result, d.ENR)
		}
	}

	for _, addrString := range addresses {
		if addrString == "" {
			continue
		}

		if strings.HasPrefix(addrString, "enrtree://") {
			// Use DNS Discovery
			wg.Add(1)
			go func(addr string) {
				defer wg.Done()
				w.dnsDiscover(ctx, addr, retrieveENR)
			}(addrString)
		} else {
			// It's a normal enr
			bootnode, err := enode.Parse(enode.ValidSchemes, addrString)
			if err != nil {
				return nil, err
			}
			result = append(result, bootnode)
		}
	}
	wg.Wait()

	w.seededBootnodesForDiscV5 = len(result) > 0

	return result, nil
}

type fnApplyToEachPeer func(d dnsdisc.DiscoveredNode, wg *sync.WaitGroup)

func (w *Waku) dnsDiscover(ctx context.Context, enrtreeAddress string, apply fnApplyToEachPeer) {
	w.logger.Info("retrieving nodes", zap.String("enr", enrtreeAddress))
	ctx, cancel := context.WithTimeout(ctx, requestTimeout)
	defer cancel()

	w.dnsAddressCacheLock.Lock()
	defer w.dnsAddressCacheLock.Unlock()

	discNodes, ok := w.dnsAddressCache[enrtreeAddress]
	if !ok {
		w.settingsMu.RLock()
		nameserver := w.settings.Nameserver
		resolver := w.settings.Resolver
		w.settingsMu.RUnlock()

		var opts []dnsdisc.DNSDiscoveryOption
		if nameserver != "" {
			opts = append(opts, dnsdisc.WithNameserver(nameserver))
		}
		if resolver != nil {
			opts = append(opts, dnsdisc.WithResolver(resolver))
		}

		discoveredNodes, err := dnsdisc.RetrieveNodes(ctx, enrtreeAddress, opts...)
		if err != nil {
			w.logger.Warn("dns discovery error ", zap.Error(err))
			return
		}

		if len(discoveredNodes) != 0 {
			w.dnsAddressCache[enrtreeAddress] = append(w.dnsAddressCache[enrtreeAddress], discoveredNodes...)
			discNodes = w.dnsAddressCache[enrtreeAddress]
		}
	}

	wg := &sync.WaitGroup{}
	wg.Add(len(discNodes))
	for _, d := range discNodes {
		apply(d, wg)
	}
	wg.Wait()
}

func (w *Waku) addWakuV2Peers(ctx context.Context, cfg *Config) error {
	fnApply := func(d dnsdisc.DiscoveredNode, wg *sync.WaitGroup) {
		defer wg.Done()
		if len(d.PeerInfo.Addrs) != 0 {
			go w.identifyAndConnect(ctx, w.settings.LightClient, d.PeerInfo)
		}
	}

	for _, addrString := range cfg.WakuNodes {
		addrString := addrString
		if strings.HasPrefix(addrString, "enrtree://") {
			// Use DNS Discovery
			go w.dnsDiscover(ctx, addrString, fnApply)
		} else {
			// It is a normal multiaddress
			addr, err := multiaddr.NewMultiaddr(addrString)
			if err != nil {
				w.logger.Warn("invalid peer multiaddress", zap.String("ma", addrString), zap.Error(err))
				continue
			}

			peerInfo, err := peer.AddrInfoFromP2pAddr(addr)
			if err != nil {
				w.logger.Warn("invalid peer multiaddress", zap.Stringer("addr", addr), zap.Error(err))
				continue
			}

			go w.identifyAndConnect(ctx, cfg.LightClient, *peerInfo)
		}
	}

	return nil
}

func (w *Waku) identifyAndConnect(ctx context.Context, isLightClient bool, peerInfo peer.AddrInfo) {
	ctx, cancel := context.WithTimeout(ctx, 7*time.Second)
	defer cancel()

	err := w.node.Host().Connect(ctx, peerInfo)
	if err != nil {
		w.logger.Error("could not connect to peer", zap.Any("peer", peerInfo), zap.Error(err))
		return
	}

	conns := w.node.Host().Network().ConnsToPeer(peerInfo.ID)
	if len(conns) == 0 {
		return // No connection
	}

	select {
	case <-w.ctx.Done():
		return
	case <-w.identifyService.IdentifyWait(conns[0]):
		if isLightClient {
			err = w.node.Host().Network().ClosePeer(peerInfo.ID)
			if err != nil {
				w.logger.Error("could not close connections to peer", zap.Stringer("peer", peerInfo.ID), zap.Error(err))
			}
			return
		}

		supportedProtocols, err := w.node.Host().Peerstore().SupportsProtocols(peerInfo.ID, relay.WakuRelayID_v200)
		if err != nil {
			w.logger.Error("could not obtain protocols", zap.Stringer("peer", peerInfo.ID), zap.Error(err))
			return
		}

		if len(supportedProtocols) == 0 {
			err = w.node.Host().Network().ClosePeer(peerInfo.ID)
			if err != nil {
				w.logger.Error("could not close connections to peer", zap.Stringer("peer", peerInfo.ID), zap.Error(err))
			}
		}
	}
}

func (w *Waku) telemetryBandwidthStats(telemetryServerURL string) {
	if telemetryServerURL == "" {
		return
	}

	telemetry := NewBandwidthTelemetryClient(w.logger, telemetryServerURL)

	ticker := time.NewTicker(time.Second * 20)
	defer ticker.Stop()

	today := time.Now()

	for {
		select {
		case <-w.ctx.Done():
			return
		case now := <-ticker.C:
			// Reset totals when day changes
			if now.Day() != today.Day() {
				today = now
				w.bandwidthCounter.Reset()
			}

			storeStats := w.bandwidthCounter.GetBandwidthForProtocol(store.StoreID_v20beta4)
			relayStats := w.bandwidthCounter.GetBandwidthForProtocol(relay.WakuRelayID_v200)
			go telemetry.PushProtocolStats(relayStats, storeStats)
		}
	}
}

func (w *Waku) GetStats() types.StatsSummary {
	stats := w.bandwidthCounter.GetBandwidthTotals()
	return types.StatsSummary{
		UploadRate:   uint64(stats.RateOut),
		DownloadRate: uint64(stats.RateIn),
	}
}

func (w *Waku) runPeerExchangeLoop() {
	defer w.wg.Done()
	if !w.settings.PeerExchange || !w.settings.LightClient {
		// Currently peer exchange is only used for full nodes
		// TODO: should it be used for lightpush? or lightpush nodes
		// are only going to be selected from a specific set of peers?
		return
	}

	ticker := time.NewTicker(time.Second * 5)
	defer ticker.Stop()

	for {
		select {
		case <-w.ctx.Done():
			w.logger.Debug("Peer exchange loop stopped")
			return
		case <-ticker.C:
			w.logger.Info("Running peer exchange loop")

			connectedPeers := w.node.Host().Network().Peers()
			peersWithRelay := 0
			for _, p := range connectedPeers {
				supportedProtocols, err := w.node.Host().Peerstore().SupportsProtocols(p, relay.WakuRelayID_v200)
				if err != nil {
					continue
				}
				if len(supportedProtocols) != 0 {
					peersWithRelay++
				}
			}

			peersToDiscover := w.settings.DiscoveryLimit - peersWithRelay
			if peersToDiscover <= 0 {
				continue
			}

			// We select only the nodes discovered via DNS Discovery that support peer exchange
			w.dnsAddressCacheLock.RLock()
			var withThesePeers []peer.ID
			for _, record := range w.dnsAddressCache {
				for _, discoveredNode := range record {
					if len(discoveredNode.PeerInfo.Addrs) == 0 {
						continue
					}

					// Obtaining peer ID
					peerIDString := discoveredNode.PeerID.String()

					peerID, err := peer.Decode(peerIDString)
					if err != nil {
						w.logger.Warn("couldnt decode peerID", zap.String("peerIDString", peerIDString))
						continue // Couldnt decode the peerID for some reason?
					}

					supportsProtocol, _ := w.node.Host().Peerstore().SupportsProtocols(peerID, peer_exchange.PeerExchangeID_v20alpha1)
					if len(supportsProtocol) != 0 {
						withThesePeers = append(withThesePeers, peerID)
					}
				}
			}
			w.dnsAddressCacheLock.RUnlock()

			if len(withThesePeers) == 0 {
				continue // No peers with peer exchange have been discovered via DNS Discovery so far, skip this iteration
			}

			err := w.node.PeerExchange().Request(w.ctx, peersToDiscover, peer_exchange.WithAutomaticPeerSelection(withThesePeers...))
			if err != nil {
				w.logger.Error("couldnt request peers via peer exchange", zap.Error(err))
			}
		}
	}
}

func (w *Waku) getPubsubTopic(topic string) string {
	if topic == "" || !w.cfg.UseShardAsDefaultTopic {
		topic = w.settings.DefaultPubsubTopic
	}
	return topic
}

func (w *Waku) unsubscribeFromPubsubTopicWithWakuRelay(topic string) error {
	topic = w.getPubsubTopic(topic)

	if !w.node.Relay().IsSubscribed(topic) {
		return nil
	}

	contentFilter := protocol.NewContentFilter(topic)

	return w.node.Relay().Unsubscribe(w.ctx, contentFilter)
}

func (w *Waku) subscribeToPubsubTopicWithWakuRelay(topic string, pubkey *ecdsa.PublicKey) error {
	if w.settings.LightClient {
		return errors.New("only available for full nodes")
	}

	topic = w.getPubsubTopic(topic)

	if w.node.Relay().IsSubscribed(topic) {
		return nil
	}

	if pubkey != nil {
		err := w.node.Relay().AddSignedTopicValidator(topic, pubkey)
		if err != nil {
			return err
		}
	}

	contentFilter := protocol.NewContentFilter(topic)

	sub, err := w.node.Relay().Subscribe(w.ctx, contentFilter)
	if err != nil {
		return err
	}

	w.wg.Add(1)
	go func() {
		defer w.wg.Done()
		for {
			select {
			case <-w.ctx.Done():
				err := w.node.Relay().Unsubscribe(w.ctx, contentFilter)
				if err != nil && !errors.Is(err, context.Canceled) {
					w.logger.Error("could not unsubscribe", zap.Error(err))
				}
				return
			case env := <-sub[0].Ch:
				err := w.OnNewEnvelopes(env, common.RelayedMessageType, false)
				if err != nil {
					w.logger.Error("OnNewEnvelopes error", zap.Error(err))
				}
			}
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

// ConfirmationsEnabled returns true if message confirmations are enabled.
func (w *Waku) ConfirmationsEnabled() bool {
	w.settingsMu.RLock()
	defer w.settingsMu.RUnlock()
	return w.settings.EnableConfirmations
}

// CurrentTime returns current time.
func (w *Waku) CurrentTime() time.Time {
	return w.timesource.Now()
}

// APIs returns the RPC descriptors the Waku implementation offers
func (w *Waku) APIs() []rpc.API {
	return []rpc.API{
		{
			Namespace: Name,
			Version:   VersionStr,
			Service:   NewPublicWakuAPI(w),
			Public:    false,
		},
	}
}

// Protocols returns the waku sub-protocols ran by this particular client.
func (w *Waku) Protocols() []p2p.Protocol {
	return []p2p.Protocol{}
}

func (w *Waku) SendEnvelopeEvent(event common.EnvelopeEvent) int {
	return w.envelopeFeed.Send(event)
}

// SubscribeEnvelopeEvents subscribes to envelopes feed.
// In order to prevent blocking waku producers events must be amply buffered.
func (w *Waku) SubscribeEnvelopeEvents(events chan<- common.EnvelopeEvent) event.Subscription {
	return w.envelopeFeed.Subscribe(events)
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
	f.PubsubTopic = w.getPubsubTopic(f.PubsubTopic)
	id, err := w.filters.Install(f)
	if err != nil {
		return id, err
	}

	if w.settings.LightClient {
		w.filterManager.eventChan <- FilterEvent{eventType: FilterEventAdded, filterID: id}
	}

	return id, nil
}

// Unsubscribe removes an installed message handler.
func (w *Waku) Unsubscribe(ctx context.Context, id string) error {
	ok := w.filters.Uninstall(id)
	if !ok {
		return fmt.Errorf("failed to unsubscribe: invalid ID '%s'", id)
	}

	if w.settings.LightClient {
		w.filterManager.eventChan <- FilterEvent{eventType: FilterEventRemoved, filterID: id}
	}

	return nil
}

// Used for testing
func (w *Waku) getFilterStats() FilterSubs {
	ch := make(chan FilterSubs)
	w.filterManager.eventChan <- FilterEvent{eventType: FilterEventGetStats, ch: ch}
	stats := <-ch

	return stats
}

// GetFilter returns the filter by id.
func (w *Waku) GetFilter(id string) *common.Filter {
	return w.filters.Get(id)
}

// Unsubscribe removes an installed message handler.
func (w *Waku) UnsubscribeMany(ids []string) error {
	for _, id := range ids {
		w.logger.Info("cleaning up filter", zap.String("id", id))
		ok := w.filters.Uninstall(id)
		if !ok {
			w.logger.Warn("could not remove filter with id", zap.String("id", id))
		}
	}
	return nil
}

func (w *Waku) SkipPublishToTopic(value bool) {
	w.settings.SkipPublishToTopic = value
}

func (w *Waku) broadcast() {
	for {
		select {
		case envelope := <-w.sendQueue:
			logger := w.logger.With(zap.String("envelopeHash", hexutil.Encode(envelope.Hash())), zap.String("pubsubTopic", envelope.PubsubTopic()), zap.String("contentTopic", envelope.Message().ContentTopic), zap.Int64("timestamp", envelope.Message().GetTimestamp()))
			var fn publishFn
			if w.settings.SkipPublishToTopic {
				// For now only used in testing to simulate going offline
				fn = func(env *protocol.Envelope, logger *zap.Logger) error {
					return errors.New("test send failure")
				}
			} else if w.settings.LightClient {
				fn = func(env *protocol.Envelope, logger *zap.Logger) error {
					logger.Info("publishing message via lightpush")
					_, err := w.node.Lightpush().Publish(w.ctx, env.Message(), lightpush.WithPubSubTopic(env.PubsubTopic()))
					return err
				}
			} else {
				fn = func(env *protocol.Envelope, logger *zap.Logger) error {
					peerCnt := len(w.node.Relay().PubSub().ListPeers(env.PubsubTopic()))
					logger.Info("publishing message via relay", zap.Int("peerCnt", peerCnt))
					_, err := w.node.Relay().Publish(w.ctx, env.Message(), relay.WithPubSubTopic(env.PubsubTopic()))
					return err
				}
			}

			w.wg.Add(1)
			go w.publishEnvelope(envelope, fn, logger)
		case <-w.ctx.Done():
			return
		}
	}
}

type publishFn = func(envelope *protocol.Envelope, logger *zap.Logger) error

func (w *Waku) publishEnvelope(envelope *protocol.Envelope, publishFn publishFn, logger *zap.Logger) {
	defer w.wg.Done()

	var event common.EventType
	if err := publishFn(envelope, logger); err != nil {
		logger.Error("could not send message", zap.Error(err))
		event = common.EventEnvelopeExpired
	} else {
		event = common.EventEnvelopeSent
	}

	w.SendEnvelopeEvent(common.EnvelopeEvent{
		Hash:  gethcommon.BytesToHash(envelope.Hash()),
		Event: event,
	})
}

// Send injects a message into the waku send queue, to be distributed in the
// network in the coming cycles.
func (w *Waku) Send(pubsubTopic string, msg *pb.WakuMessage) ([]byte, error) {
	pubsubTopic = w.getPubsubTopic(pubsubTopic)
	if w.protectedTopicStore != nil {
		privKey, err := w.protectedTopicStore.FetchPrivateKey(pubsubTopic)
		if err != nil {
			return nil, err
		}

		if privKey != nil {
			err = relay.SignMessage(privKey, msg, pubsubTopic)
			if err != nil {
				return nil, err
			}
		}
	}

	envelope := protocol.NewEnvelope(msg, msg.GetTimestamp(), pubsubTopic)

	w.sendQueue <- envelope

	w.poolMu.Lock()
	_, alreadyCached := w.envelopes[gethcommon.BytesToHash(envelope.Hash())]
	w.poolMu.Unlock()
	if !alreadyCached {
		recvMessage := common.NewReceivedMessage(envelope, common.RelayedMessageType)
		w.postEvent(recvMessage) // notify the local node about the new message
		w.addEnvelope(recvMessage)
	}

	return envelope.Hash(), nil
}

func (w *Waku) query(ctx context.Context, peerID peer.ID, pubsubTopic string, topics []common.TopicType, from uint64, to uint64, requestID []byte, opts []store.HistoryRequestOption) (*store.Result, error) {

	opts = append(opts, store.WithRequestID(requestID))

	strTopics := make([]string, len(topics))
	for i, t := range topics {
		strTopics[i] = t.ContentTopic()
	}

	opts = append(opts, store.WithPeer(peerID))

	query := store.Query{
		StartTime:     proto.Int64(int64(from) * int64(time.Second)),
		EndTime:       proto.Int64(int64(to) * int64(time.Second)),
		ContentTopics: strTopics,
		PubsubTopic:   pubsubTopic,
	}

	w.logger.Debug("store.query",
		zap.String("requestID", hexutil.Encode(requestID)),
		zap.Int64p("startTime", query.StartTime),
		zap.Int64p("endTime", query.EndTime),
		zap.Strings("contentTopics", query.ContentTopics),
		zap.String("pubsubTopic", query.PubsubTopic),
		zap.Stringer("peerID", peerID))

	return w.node.Store().Query(ctx, query, opts...)
}

func (w *Waku) Query(ctx context.Context, peerID peer.ID, pubsubTopic string, topics []common.TopicType, from uint64, to uint64, opts []store.HistoryRequestOption, processEnvelopes bool) (cursor *storepb.Index, envelopesCount int, err error) {
	requestID := protocol.GenerateRequestID()
	pubsubTopic = w.getPubsubTopic(pubsubTopic)

	result, err := w.query(ctx, peerID, pubsubTopic, topics, from, to, requestID, opts)
	if err != nil {
		w.logger.Error("error querying storenode",
			zap.String("requestID", hexutil.Encode(requestID)),
			zap.String("peerID", peerID.String()),
			zap.Error(err))

		if w.onHistoricMessagesRequestFailed != nil {
			w.onHistoricMessagesRequestFailed(requestID, peerID, err)
		}
		return nil, 0, err
	}

	envelopesCount = len(result.Messages)

	for _, msg := range result.Messages {
		// Temporarily setting RateLimitProof to nil so it matches the WakuMessage protobuffer we are sending
		// See https://github.com/vacp2p/rfc/issues/563
		msg.RateLimitProof = nil

		envelope := protocol.NewEnvelope(msg, msg.GetTimestamp(), pubsubTopic)
		w.logger.Info("received waku2 store message",
			zap.Any("envelopeHash", hexutil.Encode(envelope.Hash())),
			zap.String("pubsubTopic", pubsubTopic),
			zap.Int64p("timestamp", envelope.Message().Timestamp),
		)

		err = w.OnNewEnvelopes(envelope, common.StoreMessageType, processEnvelopes)
		if err != nil {
			return nil, 0, err
		}
	}

	if !result.IsComplete() {
		cursor = result.Cursor()
	}

	return
}

// Start implements node.Service, starting the background data propagation thread
// of the Waku protocol.
func (w *Waku) Start() error {
	if w.ctx == nil {
		w.ctx, w.cancel = context.WithCancel(context.Background())
	}

	var err error
	if w.node, err = node.New(w.settings.Options...); err != nil {
		return fmt.Errorf("failed to create a go-waku node: %v", err)
	}

	w.connectionChanged = make(chan struct{})

	if err = w.node.Start(w.ctx); err != nil {
		return fmt.Errorf("failed to start go-waku node: %v", err)
	}

	w.logger.Info("WakuV2 PeerID", zap.Stringer("id", w.node.Host().ID()))

	idService, err := identify.NewIDService(w.node.Host())
	if err != nil {
		return err
	}
	idService.Start()

	w.identifyService = idService

	if err = w.addWakuV2Peers(w.ctx, w.cfg); err != nil {
		return fmt.Errorf("failed to add wakuv2 peers: %v", err)
	}

	if w.cfg.EnableDiscV5 {
		err := w.node.DiscV5().Start(w.ctx)
		if err != nil {
			return err
		}
	}

	if w.cfg.PeerExchange {
		err := w.node.PeerExchange().Start(w.ctx)
		if err != nil {
			return err
		}
	}

	w.wg.Add(2)

	go func() {
		defer w.wg.Done()

		isConnected := false
		for {
			select {
			case <-w.ctx.Done():
				return
			case c := <-w.connStatusChan:
				w.connStatusMu.Lock()
				latestConnStatus := formatConnStatus(w.node, c)
				w.logger.Debug("peer stats",
					zap.Int("peersCount", len(latestConnStatus.Peers)),
					zap.Any("stats", latestConnStatus))
				for k, subs := range w.connStatusSubscriptions {
					if subs.Active() {
						subs.C <- latestConnStatus
					} else {
						delete(w.connStatusSubscriptions, k)
					}
				}
				w.connStatusMu.Unlock()
				if w.onPeerStats != nil {
					w.onPeerStats(latestConnStatus)
				}

				if w.cfg.EnableDiscV5 {
					// Restarting DiscV5
					if !latestConnStatus.IsOnline && isConnected {
						w.logger.Info("Restarting DiscV5: offline and is connected")
						isConnected = false
						w.node.DiscV5().Stop()
					} else if latestConnStatus.IsOnline && !isConnected {
						w.logger.Info("Restarting DiscV5: online and is not connected")
						isConnected = true
						if w.node.DiscV5().ErrOnNotRunning() != nil {
							err := w.node.DiscV5().Start(w.ctx)
							if err != nil {
								w.logger.Error("Could not start DiscV5", zap.Error(err))
							}
						}
					}
				}
			}
		}
	}()

	go w.telemetryBandwidthStats(w.cfg.TelemetryServerURL)
	go w.runPeerExchangeLoop()

	if w.settings.LightClient {
		// Create FilterManager that will main peer connectivity
		// for installed filters
		w.filterManager = newFilterManager(w.ctx, w.logger,
			func(id string) *common.Filter { return w.GetFilter(id) },
			w.settings,
			func(env *protocol.Envelope) error { return w.OnNewEnvelopes(env, common.RelayedMessageType, false) },
			w.node)

		w.wg.Add(1)
		go w.filterManager.runFilterLoop(&w.wg)
	}

	err = w.setupRelaySubscriptions()
	if err != nil {
		return err
	}

	numCPU := runtime.NumCPU()
	for i := 0; i < numCPU; i++ {
		go w.processQueueLoop()
	}

	go w.broadcast()

	// we should wait `seedBootnodesForDiscV5` shutdown smoothly before set w.ctx to nil within `w.Stop()`
	w.wg.Add(1)
	go w.seedBootnodesForDiscV5()

	return nil
}

func (w *Waku) setupRelaySubscriptions() error {
	if w.settings.LightClient {
		return nil
	}

	if w.protectedTopicStore != nil {
		protectedTopics, err := w.protectedTopicStore.ProtectedTopics()
		if err != nil {
			return err
		}

		for _, pt := range protectedTopics {
			// Adding subscription to protected topics
			err = w.subscribeToPubsubTopicWithWakuRelay(pt.Topic, pt.PubKey)
			if err != nil {
				return err
			}
		}
	}

	err := w.subscribeToPubsubTopicWithWakuRelay(w.settings.DefaultPubsubTopic, nil)
	if err != nil {
		return err
	}

	return nil
}

// Stop implements node.Service, stopping the background data propagation thread
// of the Waku protocol.
func (w *Waku) Stop() error {
	w.cancel()

	err := w.identifyService.Close()
	if err != nil {
		return err
	}

	w.node.Stop()

	if w.protectedTopicStore != nil {
		err = w.protectedTopicStore.Close()
		if err != nil {
			return err
		}
	}

	close(w.connectionChanged)
	w.wg.Wait()

	w.ctx = nil
	w.cancel = nil

	return nil
}

func (w *Waku) OnNewEnvelopes(envelope *protocol.Envelope, msgType common.MessageType, processImmediately bool) error {
	if envelope == nil {
		return nil
	}

	recvMessage := common.NewReceivedMessage(envelope, msgType)
	if recvMessage == nil {
		return nil
	}

	if w.statusTelemetryClient != nil {
		w.statusTelemetryClient.PushReceivedEnvelope(envelope)
	}

	logger := w.logger.With(
		zap.Any("messageType", msgType),
		zap.String("envelopeHash", hexutil.Encode(envelope.Hash())),
		zap.String("contentTopic", envelope.Message().ContentTopic),
		zap.Int64("timestamp", envelope.Message().GetTimestamp()),
	)

	logger.Debug("received new envelope")
	trouble := false

	_, err := w.add(recvMessage, processImmediately)
	if err != nil {
		logger.Info("invalid envelope received", zap.Error(err))
		trouble = true
	}

	common.EnvelopesValidatedCounter.Inc()

	if trouble {
		return errors.New("received invalid envelope")
	}

	return nil
}

// addEnvelope adds an envelope to the envelope map, used for sending
func (w *Waku) addEnvelope(envelope *common.ReceivedMessage) {
	hash := envelope.Hash()

	w.poolMu.Lock()
	w.envelopes[hash] = envelope
	w.poolMu.Unlock()
}

func (w *Waku) add(recvMessage *common.ReceivedMessage, processImmediately bool) (bool, error) {
	common.EnvelopesReceivedCounter.Inc()

	hash := recvMessage.Hash()

	w.poolMu.Lock()
	envelope, alreadyCached := w.envelopes[hash]
	w.poolMu.Unlock()
	if !alreadyCached {
		recvMessage.Processed.Store(false)
		w.addEnvelope(recvMessage)
	}

	logger := w.logger.With(zap.String("envelopeHash", recvMessage.Hash().Hex()))

	if alreadyCached {
		logger.Debug("w envelope already cached")
		common.EnvelopesCachedCounter.WithLabelValues("hit").Inc()
	} else {
		logger.Debug("cached w envelope")
		common.EnvelopesCachedCounter.WithLabelValues("miss").Inc()
		common.EnvelopesSizeMeter.Observe(float64(len(recvMessage.Envelope.Message().Payload)))
	}

	if !alreadyCached || !envelope.Processed.Load() {
		if processImmediately {
			logger.Debug("immediately processing envelope")
			w.processReceivedMessage(recvMessage)
		} else {
			logger.Debug("posting event")
			w.postEvent(recvMessage) // notify the local node about the new message
		}
	}

	return true, nil
}

// postEvent queues the message for further processing.
func (w *Waku) postEvent(envelope *common.ReceivedMessage) {
	w.msgQueue <- envelope
}

// processQueueLoop delivers the messages to the watchers during the lifetime of the waku node.
func (w *Waku) processQueueLoop() {
	if w.ctx == nil {
		return
	}
	for {
		select {
		case <-w.ctx.Done():
			return
		case e := <-w.msgQueue:
			w.processReceivedMessage(e)
		}
	}
}

func (w *Waku) processReceivedMessage(e *common.ReceivedMessage) {
	logger := w.logger.With(
		zap.String("envelopeHash", hexutil.Encode(e.Envelope.Hash())),
		zap.String("pubsubTopic", e.PubsubTopic),
		zap.String("contentTopic", e.ContentTopic.ContentTopic()),
		zap.Int64("timestamp", e.Envelope.Message().GetTimestamp()),
	)

	if e.MsgType == common.StoreMessageType {
		// We need to insert it first, and then remove it if not matched,
		// as messages are processed asynchronously
		w.storeMsgIDsMu.Lock()
		w.storeMsgIDs[e.Hash()] = true
		w.storeMsgIDsMu.Unlock()
	}

	matched := w.filters.NotifyWatchers(e)

	// If not matched we remove it
	if !matched {
		logger.Debug("filters did not match")
		w.storeMsgIDsMu.Lock()
		delete(w.storeMsgIDs, e.Hash())
		w.storeMsgIDsMu.Unlock()
	} else {
		logger.Debug("filters did match")
		e.Processed.Store(true)
	}

	w.envelopeFeed.Send(common.EnvelopeEvent{
		Topic: e.ContentTopic,
		Hash:  e.Hash(),
		Event: common.EventEnvelopeAvailable,
	})
}

// Envelopes retrieves all the messages currently pooled by the node.
func (w *Waku) Envelopes() []*common.ReceivedMessage {
	w.poolMu.RLock()
	defer w.poolMu.RUnlock()

	all := make([]*common.ReceivedMessage, 0, len(w.envelopes))
	for _, envelope := range w.envelopes {
		all = append(all, envelope)
	}
	return all
}

// GetEnvelope retrieves an envelope from the message queue by its hash.
// It returns nil if the envelope can not be found.
func (w *Waku) GetEnvelope(hash gethcommon.Hash) *common.ReceivedMessage {
	w.poolMu.RLock()
	defer w.poolMu.RUnlock()
	return w.envelopes[hash]
}

// isEnvelopeCached checks if envelope with specific hash has already been received and cached.
func (w *Waku) IsEnvelopeCached(hash gethcommon.Hash) bool {
	w.poolMu.Lock()
	defer w.poolMu.Unlock()

	_, exist := w.envelopes[hash]
	return exist
}

func (w *Waku) ClearEnvelopesCache() {
	w.poolMu.Lock()
	defer w.poolMu.Unlock()
	w.envelopes = make(map[gethcommon.Hash]*common.ReceivedMessage)
}

func (w *Waku) PeerCount() int {
	return w.node.PeerCount()
}

func (w *Waku) Peers() map[string]types.WakuV2Peer {
	return FormatPeerStats(w.node, w.node.PeerStats())
}

func (w *Waku) ListenAddresses() []string {
	addrs := w.node.ListenAddresses()
	var result []string
	for _, addr := range addrs {
		result = append(result, addr.String())
	}
	return result
}

func (w *Waku) SubscribeToPubsubTopic(topic string, pubkey *ecdsa.PublicKey) error {
	topic = w.getPubsubTopic(topic)

	if !w.settings.LightClient {
		err := w.subscribeToPubsubTopicWithWakuRelay(topic, pubkey)
		if err != nil {
			return err
		}
	}
	return nil
}

func (w *Waku) UnsubscribeFromPubsubTopic(topic string) error {
	topic = w.getPubsubTopic(topic)

	if !w.settings.LightClient {
		err := w.unsubscribeFromPubsubTopicWithWakuRelay(topic)
		if err != nil {
			return err
		}
	}
	return nil
}

func (w *Waku) RetrievePubsubTopicKey(topic string) (*ecdsa.PrivateKey, error) {
	topic = w.getPubsubTopic(topic)
	if w.protectedTopicStore == nil {
		return nil, nil
	}

	return w.protectedTopicStore.FetchPrivateKey(topic)
}

func (w *Waku) StorePubsubTopicKey(topic string, privKey *ecdsa.PrivateKey) error {
	topic = w.getPubsubTopic(topic)
	if w.protectedTopicStore == nil {
		return nil
	}

	return w.protectedTopicStore.Insert(topic, privKey, &privKey.PublicKey)
}

func (w *Waku) RemovePubsubTopicKey(topic string) error {
	topic = w.getPubsubTopic(topic)
	if w.protectedTopicStore == nil {
		return nil
	}

	return w.protectedTopicStore.Delete(topic)
}

func (w *Waku) StartDiscV5() error {
	if w.node.DiscV5() == nil {
		return errors.New("discv5 is not setup")
	}

	return w.node.DiscV5().Start(w.ctx)
}

func (w *Waku) StopDiscV5() error {
	if w.node.DiscV5() == nil {
		return errors.New("discv5 is not setup")
	}

	w.node.DiscV5().Stop()
	return nil
}

func (w *Waku) ConnectionChanged(state connection.State) {
	if !state.Offline && w.offline {
		select {
		case w.connectionChanged <- struct{}{}:
		default:
			w.logger.Warn("could not write on connection changed channel")
		}
	}

	w.offline = !state.Offline
}

// seedBootnodesForDiscV5 tries to fetch bootnodes
// from an ENR periodically.
// It backs off exponentially until maxRetries, at which point it restarts from 0
// It also restarts if there's a connection change signalled from the client
func (w *Waku) seedBootnodesForDiscV5() {
	if !w.settings.EnableDiscV5 || w.node.DiscV5() == nil {
		w.wg.Done()
		return
	}

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()
	var retries = 0

	now := func() int64 {
		return time.Now().UnixNano() / int64(time.Millisecond)

	}

	var lastTry = now()

	canQuery := func() bool {
		backoff := bootnodesQueryBackoffMs * int64(math.Exp2(float64(retries)))

		return lastTry+backoff < now()
	}

	for {
		select {
		case <-ticker.C:
			if w.seededBootnodesForDiscV5 && len(w.node.Host().Network().Peers()) > 3 {
				w.logger.Debug("not querying bootnodes", zap.Bool("seeded", w.seededBootnodesForDiscV5), zap.Int("peer-count", len(w.node.Host().Network().Peers())))
				continue
			}
			if canQuery() {
				w.logger.Info("querying bootnodes to restore connectivity", zap.Int("peer-count", len(w.node.Host().Network().Peers())))
				err := w.restartDiscV5()
				if err != nil {
					w.logger.Warn("failed to restart discv5", zap.Error(err))
				}

				lastTry = now()
				retries++
				// We reset the retries after a while and restart
				if retries > bootnodesMaxRetries {
					retries = 0
				}

			} else {
				w.logger.Info("can't query bootnodes", zap.Int("peer-count", len(w.node.Host().Network().Peers())), zap.Int64("lastTry", lastTry), zap.Int64("now", now()), zap.Int64("backoff", bootnodesQueryBackoffMs*int64(math.Exp2(float64(retries)))), zap.Int("retries", retries))

			}
		// If we go online, trigger immediately
		case <-w.connectionChanged:
			if canQuery() {
				err := w.restartDiscV5()
				if err != nil {
					w.logger.Warn("failed to restart discv5", zap.Error(err))
				}

			}
			retries = 0
			lastTry = now()

		case <-w.ctx.Done():
			w.wg.Done()
			w.logger.Debug("bootnode seeding stopped")
			return
		}
	}
}

// Restart discv5, re-retrieving bootstrap nodes
func (w *Waku) restartDiscV5() error {
	ctx, cancel := context.WithTimeout(w.ctx, 30*time.Second)
	defer cancel()
	bootnodes, err := w.getDiscV5BootstrapNodes(ctx, w.discV5BootstrapNodes)
	if err != nil {
		return err
	}
	if len(bootnodes) == 0 {
		return errors.New("failed to fetch bootnodes")
	}

	if w.node.DiscV5().ErrOnNotRunning() != nil {
		w.logger.Info("is not started restarting")
		err := w.node.DiscV5().Start(ctx)
		if err != nil {
			w.logger.Error("Could not start DiscV5", zap.Error(err))
		}
	} else {
		w.node.DiscV5().Stop()
		w.logger.Info("is started restarting")

		select {
		case <-w.ctx.Done(): // Don't start discv5 if we are stopping waku
			return nil
		default:
		}

		err := w.node.DiscV5().Start(ctx)
		if err != nil {
			w.logger.Error("Could not start DiscV5", zap.Error(err))
		}
	}

	w.logger.Info("restarting discv5 with nodes", zap.Any("nodes", bootnodes))
	return w.node.SetDiscV5Bootnodes(bootnodes)
}

func (w *Waku) AddStorePeer(address string) (peer.ID, error) {
	addr, err := multiaddr.NewMultiaddr(address)
	if err != nil {
		return "", err
	}

	peerID, err := w.node.AddPeer(addr, wps.Static, []string{}, store.StoreID_v20beta4)
	if err != nil {
		return "", err
	}
	return peerID, nil
}

func (w *Waku) timestamp() int64 {
	return w.timesource.Now().UnixNano()
}

func (w *Waku) AddRelayPeer(address string) (peer.ID, error) {
	addr, err := multiaddr.NewMultiaddr(address)
	if err != nil {
		return "", err
	}

	peerID, err := w.node.AddPeer(addr, wps.Static, []string{}, relay.WakuRelayID_v200)
	if err != nil {
		return "", err
	}
	return peerID, nil
}

func (w *Waku) DialPeer(address string) error {
	ctx, cancel := context.WithTimeout(w.ctx, requestTimeout)
	defer cancel()
	return w.node.DialPeer(ctx, address)
}

func (w *Waku) DialPeerByID(peerID string) error {
	ctx, cancel := context.WithTimeout(w.ctx, requestTimeout)
	defer cancel()
	pid, err := peer.Decode(peerID)
	if err != nil {
		return err
	}
	return w.node.DialPeerByID(ctx, pid)
}

func (w *Waku) DropPeer(peerID string) error {
	pid, err := peer.Decode(peerID)
	if err != nil {
		return err
	}
	return w.node.ClosePeerById(pid)
}

func (w *Waku) ProcessingP2PMessages() bool {
	w.storeMsgIDsMu.Lock()
	defer w.storeMsgIDsMu.Unlock()
	return len(w.storeMsgIDs) != 0
}

func (w *Waku) MarkP2PMessageAsProcessed(hash gethcommon.Hash) {
	w.storeMsgIDsMu.Lock()
	defer w.storeMsgIDsMu.Unlock()
	delete(w.storeMsgIDs, hash)
}

func (w *Waku) Clean() error {
	w.msgQueue = make(chan *common.ReceivedMessage, messageQueueLimit)

	for _, f := range w.filters.All() {
		f.Messages = common.NewMemoryMessageStore()
	}

	return nil
}

func (w *Waku) PeerID() peer.ID {
	return w.node.Host().ID()
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

func FormatPeerStats(wakuNode *node.WakuNode, peers node.PeerStats) map[string]types.WakuV2Peer {
	p := make(map[string]types.WakuV2Peer)
	for k, v := range peers {
		peerInfo := wakuNode.Host().Peerstore().PeerInfo(k)
		wakuV2Peer := types.WakuV2Peer{}
		wakuV2Peer.Protocols = v
		hostInfo, _ := multiaddr.NewMultiaddr(fmt.Sprintf("/p2p/%s", k.Pretty()))
		for _, addr := range peerInfo.Addrs {
			wakuV2Peer.Addresses = append(wakuV2Peer.Addresses, addr.Encapsulate(hostInfo).String())
		}
		p[k.Pretty()] = wakuV2Peer
	}
	return p
}

func formatConnStatus(wakuNode *node.WakuNode, c node.ConnStatus) types.ConnStatus {
	return types.ConnStatus{
		IsOnline:   c.IsOnline,
		HasHistory: c.HasHistory,
		Peers:      FormatPeerStats(wakuNode, c.Peers),
	}
}
