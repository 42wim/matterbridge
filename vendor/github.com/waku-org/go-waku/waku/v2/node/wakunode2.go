package node

import (
	"context"
	"math/rand"
	"net"
	"sync"
	"time"

	backoffv4 "github.com/cenkalti/backoff/v4"
	golog "github.com/ipfs/go-log/v2"
	"github.com/libp2p/go-libp2p"
	"go.uber.org/zap"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/p2p/enode"

	"github.com/libp2p/go-libp2p/core/event"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/peerstore"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/libp2p/go-libp2p/p2p/host/autorelay"
	"github.com/libp2p/go-libp2p/p2p/host/peerstore/pstoremem"
	"github.com/libp2p/go-libp2p/p2p/protocol/circuitv2/proto"
	ws "github.com/libp2p/go-libp2p/p2p/transport/websocket"
	ma "github.com/multiformats/go-multiaddr"

	"github.com/waku-org/go-waku/logging"
	"github.com/waku-org/go-waku/waku/v2/discv5"
	"github.com/waku-org/go-waku/waku/v2/dnsdisc"
	"github.com/waku-org/go-waku/waku/v2/peermanager"
	wps "github.com/waku-org/go-waku/waku/v2/peerstore"
	wakuprotocol "github.com/waku-org/go-waku/waku/v2/protocol"
	"github.com/waku-org/go-waku/waku/v2/protocol/enr"
	"github.com/waku-org/go-waku/waku/v2/protocol/filter"
	"github.com/waku-org/go-waku/waku/v2/protocol/legacy_filter"
	"github.com/waku-org/go-waku/waku/v2/protocol/lightpush"
	"github.com/waku-org/go-waku/waku/v2/protocol/metadata"
	"github.com/waku-org/go-waku/waku/v2/protocol/pb"
	"github.com/waku-org/go-waku/waku/v2/protocol/peer_exchange"
	"github.com/waku-org/go-waku/waku/v2/protocol/relay"
	"github.com/waku-org/go-waku/waku/v2/protocol/store"
	"github.com/waku-org/go-waku/waku/v2/rendezvous"
	"github.com/waku-org/go-waku/waku/v2/service"
	"github.com/waku-org/go-waku/waku/v2/timesource"

	"github.com/waku-org/go-waku/waku/v2/utils"
)

const discoveryConnectTimeout = 20 * time.Second

type Peer struct {
	ID           peer.ID        `json:"peerID"`
	Protocols    []protocol.ID  `json:"protocols"`
	Addrs        []ma.Multiaddr `json:"addrs"`
	Connected    bool           `json:"connected"`
	PubsubTopics []string       `json:"pubsubTopics"`
}

type storeFactory func(w *WakuNode) store.Store

type byte32 = [32]byte

type IdentityCredential = struct {
	IDTrapdoor   byte32 `json:"idTrapdoor"`
	IDNullifier  byte32 `json:"idNullifier"`
	IDSecretHash byte32 `json:"idSecretHash"`
	IDCommitment byte32 `json:"idCommitment"`
}

type SpamHandler = func(message *pb.WakuMessage, topic string) error

type RLNRelay interface {
	IdentityCredential() (IdentityCredential, error)
	MembershipIndex() uint
	AppendRLNProof(msg *pb.WakuMessage, senderEpochTime time.Time) error
	Validator(spamHandler SpamHandler) func(ctx context.Context, message *pb.WakuMessage, topic string) bool
	Start(ctx context.Context) error
	Stop() error
	IsReady(ctx context.Context) (bool, error)
}

type WakuNode struct {
	host       host.Host
	opts       *WakuNodeParameters
	log        *zap.Logger
	timesource timesource.Timesource
	metrics    Metrics

	peerstore     peerstore.Peerstore
	peerConnector *peermanager.PeerConnectionStrategy

	relay           Service
	lightPush       Service
	discoveryV5     Service
	peerExchange    Service
	rendezvous      Service
	metadata        Service
	legacyFilter    ReceptorService
	filterFullNode  ReceptorService
	filterLightNode Service
	store           ReceptorService
	rlnRelay        RLNRelay

	wakuFlag          enr.WakuEnrBitfield
	circuitRelayNodes chan peer.AddrInfo

	localNode *enode.LocalNode

	bcaster relay.Broadcaster

	connectionNotif        ConnectionNotifier
	protocolEventSub       event.Subscription
	identificationEventSub event.Subscription
	addressChangesSub      event.Subscription
	enrChangeCh            chan struct{}

	keepAliveMutex sync.Mutex
	keepAliveFails map[peer.ID]int

	cancel context.CancelFunc
	wg     *sync.WaitGroup

	// Channel passed to WakuNode constructor
	// receiving connection status notifications
	connStatusChan chan<- ConnStatus

	storeFactory storeFactory

	peermanager *peermanager.PeerManager
}

func defaultStoreFactory(w *WakuNode) store.Store {
	return store.NewWakuStore(w.opts.messageProvider, w.peermanager, w.timesource, w.opts.prometheusReg, w.log)
}

// New is used to instantiate a WakuNode using a set of WakuNodeOptions
func New(opts ...WakuNodeOption) (*WakuNode, error) {
	var err error
	params := new(WakuNodeParameters)
	params.libP2POpts = DefaultLibP2POptions

	opts = append(DefaultWakuNodeOptions, opts...)
	for _, opt := range opts {
		err := opt(params)
		if err != nil {
			return nil, err
		}
	}

	if params.logger == nil {
		params.logger = utils.Logger()
		//golog.SetPrimaryCore(params.logger.Core())
		golog.SetAllLoggers(params.logLevel)
	}

	if params.privKey == nil {
		prvKey, err := crypto.GenerateKey()
		if err != nil {
			return nil, err
		}
		params.privKey = prvKey
	}

	if params.enableWSS {
		params.libP2POpts = append(params.libP2POpts, libp2p.Transport(ws.New, ws.WithTLSConfig(params.tlsConfig)))
	} else {
		// Enable WS transport by default
		params.libP2POpts = append(params.libP2POpts, libp2p.Transport(ws.New))
	}

	// Setting default host address if none was provided
	if params.hostAddr == nil {
		params.hostAddr, err = net.ResolveTCPAddr("tcp", "0.0.0.0:0")
		if err != nil {
			return nil, err
		}
		err = WithHostAddress(params.hostAddr)(params)
		if err != nil {
			return nil, err
		}
	}

	if len(params.multiAddr) > 0 {
		params.libP2POpts = append(params.libP2POpts, libp2p.ListenAddrs(params.multiAddr...))
	}

	params.libP2POpts = append(params.libP2POpts, params.Identity())

	if params.addressFactory != nil {
		params.libP2POpts = append(params.libP2POpts, libp2p.AddrsFactory(params.addressFactory))
	}

	w := new(WakuNode)
	w.bcaster = relay.NewBroadcaster(1024)
	w.opts = params
	w.log = params.logger.Named("node2")
	w.wg = &sync.WaitGroup{}
	w.keepAliveFails = make(map[peer.ID]int)
	w.wakuFlag = enr.NewWakuEnrBitfield(w.opts.enableLightPush, w.opts.enableLegacyFilter, w.opts.enableStore, w.opts.enableRelay)
	w.circuitRelayNodes = make(chan peer.AddrInfo)
	w.metrics = newMetrics(params.prometheusReg)

	w.metrics.RecordVersion(Version, GitCommit)

	// Setup peerstore wrapper
	if params.peerstore != nil {
		w.peerstore = wps.NewWakuPeerstore(params.peerstore)
		params.libP2POpts = append(params.libP2POpts, libp2p.Peerstore(w.peerstore))
	} else {
		ps, err := pstoremem.NewPeerstore()
		if err != nil {
			return nil, err
		}
		w.peerstore = wps.NewWakuPeerstore(ps)
		params.libP2POpts = append(params.libP2POpts, libp2p.Peerstore(w.peerstore))
	}

	// Use circuit relay with nodes received on circuitRelayNodes channel
	params.libP2POpts = append(params.libP2POpts, libp2p.EnableAutoRelayWithPeerSource(
		func(ctx context.Context, numPeers int) <-chan peer.AddrInfo {
			r := make(chan peer.AddrInfo)
			go func() {
				defer close(r)
				for ; numPeers != 0; numPeers-- {
					select {
					case v, ok := <-w.circuitRelayNodes:
						if !ok {
							return
						}
						select {
						case r <- v:
						case <-ctx.Done():
							return
						}
					case <-ctx.Done():
						return
					}
				}
			}()
			return r
		},
		autorelay.WithMinInterval(params.circuitRelayMinInterval),
		autorelay.WithBootDelay(params.circuitRelayBootDelay),
	))

	if params.enableNTP {
		w.timesource = timesource.NewNTPTimesource(w.opts.ntpURLs, w.log)
	} else {
		w.timesource = timesource.NewDefaultClock()
	}

	w.localNode, err = enr.NewLocalnode(w.opts.privKey)
	if err != nil {
		w.log.Error("creating localnode", zap.Error(err))
	}

	w.metadata = metadata.NewWakuMetadata(w.opts.clusterID, w.localNode, w.log)

	//Initialize peer manager.
	w.peermanager = peermanager.NewPeerManager(w.opts.maxPeerConnections, w.opts.peerStoreCapacity, w.log)

	w.peerConnector, err = peermanager.NewPeerConnectionStrategy(w.peermanager, discoveryConnectTimeout, w.log)
	if err != nil {
		w.log.Error("creating peer connection strategy", zap.Error(err))
	}

	if w.opts.enableDiscV5 {
		err := w.mountDiscV5()
		if err != nil {
			return nil, err
		}
	}

	w.peerExchange, err = peer_exchange.NewWakuPeerExchange(w.DiscV5(), w.peerConnector, w.peermanager, w.opts.prometheusReg, w.log)
	if err != nil {
		return nil, err
	}

	w.rendezvous = rendezvous.NewRendezvous(w.opts.rendezvousDB, w.peerConnector, w.log)
	w.relay = relay.NewWakuRelay(w.bcaster, w.opts.minRelayPeersToPublish, w.timesource, w.opts.prometheusReg, w.log,
		relay.WithPubSubOptions(w.opts.pubsubOpts),
		relay.WithMaxMsgSize(w.opts.maxMsgSizeBytes))

	if w.opts.enableRelay {
		err = w.setupRLNRelay()
		if err != nil {
			return nil, err
		}
	}

	w.opts.legacyFilterOpts = append(w.opts.legacyFilterOpts, legacy_filter.WithPeerManager(w.peermanager))
	w.opts.filterOpts = append(w.opts.filterOpts, filter.WithPeerManager(w.peermanager))

	w.legacyFilter = legacy_filter.NewWakuFilter(w.bcaster, w.opts.isLegacyFilterFullNode, w.timesource, w.opts.prometheusReg, w.log, w.opts.legacyFilterOpts...)
	w.filterFullNode = filter.NewWakuFilterFullNode(w.timesource, w.opts.prometheusReg, w.log, w.opts.filterOpts...)
	w.filterLightNode = filter.NewWakuFilterLightNode(w.bcaster, w.peermanager, w.timesource, w.opts.prometheusReg, w.log)
	w.lightPush = lightpush.NewWakuLightPush(w.Relay(), w.peermanager, w.opts.prometheusReg, w.log)

	if params.storeFactory != nil {
		w.storeFactory = params.storeFactory
	} else {
		w.storeFactory = defaultStoreFactory
	}

	if params.connStatusC != nil {
		w.connStatusChan = params.connStatusC
	}

	return w, nil
}

func (w *WakuNode) watchMultiaddressChanges(ctx context.Context) {
	defer w.wg.Done()

	addrsSet := utils.MultiAddrSet(w.ListenAddresses()...)

	first := make(chan struct{}, 1)
	first <- struct{}{}
	for {
		select {
		case <-ctx.Done():
			return
		case <-first:
			addr := utils.MultiAddrFromSet(addrsSet)
			w.log.Info("listening", logging.MultiAddrs("multiaddr", addr...))
		case <-w.addressChangesSub.Out():
			newAddrs := utils.MultiAddrSet(w.ListenAddresses()...)
			if !utils.MultiAddrSetEquals(addrsSet, newAddrs) {
				addrsSet = newAddrs
				addrs := utils.MultiAddrFromSet(addrsSet)
				w.log.Info("listening addresses update received", logging.MultiAddrs("multiaddr", addrs...))
				err := w.setupENR(ctx, addrs)
				if err != nil {
					w.log.Warn("could not update ENR", zap.Error(err))
				}
			}
		}
	}
}

// Start initializes all the protocols that were setup in the WakuNode
func (w *WakuNode) Start(ctx context.Context) error {
	connGater := peermanager.NewConnectionGater(w.log)

	ctx, cancel := context.WithCancel(ctx)
	w.cancel = cancel

	libP2POpts := append(w.opts.libP2POpts, libp2p.ConnectionGater(connGater))

	host, err := libp2p.New(libP2POpts...)
	if err != nil {
		return err
	}

	host.Network().Notify(&network.NotifyBundle{
		DisconnectedF: func(net network.Network, conn network.Conn) {
			go connGater.NotifyDisconnect(conn.RemoteMultiaddr())
		},
	})

	w.host = host

	if w.protocolEventSub, err = host.EventBus().Subscribe(new(event.EvtPeerProtocolsUpdated)); err != nil {
		return err
	}

	if w.identificationEventSub, err = host.EventBus().Subscribe(new(event.EvtPeerIdentificationCompleted)); err != nil {
		return err
	}

	if w.addressChangesSub, err = host.EventBus().Subscribe(new(event.EvtLocalAddressesUpdated)); err != nil {
		return err
	}

	w.connectionNotif = NewConnectionNotifier(ctx, w.host, w.opts.connNotifCh, w.metrics, w.log)
	w.host.Network().Notify(w.connectionNotif)

	w.enrChangeCh = make(chan struct{}, 10)

	w.wg.Add(4)
	go w.connectednessListener(ctx)
	go w.watchMultiaddressChanges(ctx)
	go w.watchENRChanges(ctx)
	go w.findRelayNodes(ctx)

	err = w.bcaster.Start(ctx)
	if err != nil {
		return err
	}

	if w.opts.keepAliveInterval > time.Duration(0) {
		w.wg.Add(1)
		go w.startKeepAlive(ctx, w.opts.keepAliveInterval)
	}

	w.metadata.SetHost(host)
	err = w.metadata.Start(ctx)
	if err != nil {
		return err
	}

	w.peerConnector.SetHost(host)
	w.peermanager.SetHost(host)
	err = w.peerConnector.Start(ctx)
	if err != nil {
		return err
	}

	if w.opts.enableNTP {
		err := w.timesource.Start(ctx)
		if err != nil {
			return err
		}
	}

	if w.opts.enableRLN {
		err = w.startRlnRelay(ctx)
		if err != nil {
			return err
		}
	}

	w.relay.SetHost(host)

	if w.opts.enableRelay {
		err := w.relay.Start(ctx)
		if err != nil {
			return err
		}
		err = w.peermanager.SubscribeToRelayEvtBus(w.relay.(*relay.WakuRelay).Events())
		if err != nil {
			return err
		}
		w.peermanager.Start(ctx)
		w.registerAndMonitorReachability(ctx)
	}

	w.store = w.storeFactory(w)
	w.store.SetHost(host)
	if w.opts.enableStore {
		sub := w.bcaster.RegisterForAll()
		err := w.startStore(ctx, sub)
		if err != nil {
			return err
		}
		w.log.Info("Subscribing store to broadcaster")
	}

	w.lightPush.SetHost(host)
	if w.opts.enableLightPush {
		if err := w.lightPush.Start(ctx); err != nil {
			return err
		}
	}

	w.legacyFilter.SetHost(host)
	if w.opts.enableLegacyFilter {
		sub := w.bcaster.RegisterForAll()
		err := w.legacyFilter.Start(ctx, sub)
		if err != nil {
			return err
		}
		w.log.Info("Subscribing filter to broadcaster")
	}

	w.filterFullNode.SetHost(host)
	if w.opts.enableFilterFullNode {
		sub := w.bcaster.RegisterForAll()
		err := w.filterFullNode.Start(ctx, sub)
		if err != nil {
			return err
		}
		w.log.Info("Subscribing filterV2 to broadcaster")

	}

	w.filterLightNode.SetHost(host)
	if w.opts.enableFilterLightNode {
		err := w.filterLightNode.Start(ctx)
		if err != nil {
			return err
		}
	}

	err = w.setupENR(ctx, w.ListenAddresses())
	if err != nil {
		return err
	}

	w.peerExchange.SetHost(host)
	if w.opts.enablePeerExchange {
		err := w.peerExchange.Start(ctx)
		if err != nil {
			return err
		}
	}

	w.rendezvous.SetHost(host)
	if w.opts.enableRendezvousPoint {
		err := w.rendezvous.Start(ctx)
		if err != nil {
			return err
		}
	}

	return nil
}

// Stop stops the WakuNode and closess all connections to the host
func (w *WakuNode) Stop() {
	if w.cancel == nil {
		return
	}

	w.bcaster.Stop()

	defer w.connectionNotif.Close()
	defer w.protocolEventSub.Close()
	defer w.identificationEventSub.Close()
	defer w.addressChangesSub.Close()

	w.host.Network().StopNotify(w.connectionNotif)

	w.relay.Stop()
	w.lightPush.Stop()
	w.store.Stop()
	w.legacyFilter.Stop()
	w.filterFullNode.Stop()
	w.filterLightNode.Stop()

	if w.opts.enableDiscV5 {
		w.discoveryV5.Stop()
	}
	w.peerExchange.Stop()
	w.rendezvous.Stop()

	w.peerConnector.Stop()

	_ = w.stopRlnRelay()

	w.timesource.Stop()

	w.host.Close()

	w.cancel()

	w.wg.Wait()

	close(w.enrChangeCh)

	w.cancel = nil
}

// Host returns the libp2p Host used by the WakuNode
func (w *WakuNode) Host() host.Host {
	return w.host
}

// ID returns the base58 encoded ID from the host
func (w *WakuNode) ID() string {
	return w.host.ID().Pretty()
}

func (w *WakuNode) watchENRChanges(ctx context.Context) {
	defer w.wg.Done()

	var prevNodeVal string
	for {
		select {
		case <-ctx.Done():
			return
		case <-w.enrChangeCh:
			if w.localNode != nil {
				currNodeVal := w.localNode.Node().String()
				if prevNodeVal != currNodeVal {
					if prevNodeVal == "" {
						w.log.Info("enr record", logging.ENode("enr", w.localNode.Node()))
					} else {
						w.log.Info("new enr record", logging.ENode("enr", w.localNode.Node()))
					}
					prevNodeVal = currNodeVal
				}
			}
		}
	}
}

// ListenAddresses returns all the multiaddresses used by the host
func (w *WakuNode) ListenAddresses() []ma.Multiaddr {
	return utils.EncapsulatePeerID(w.host.ID(), w.host.Addrs()...)
}

// ENR returns the ENR address of the node
func (w *WakuNode) ENR() *enode.Node {
	return w.localNode.Node()
}

// Timesource returns the timesource used by this node to obtain the current wall time
// Depending on the configuration it will be the local time or a ntp syncd time
func (w *WakuNode) Timesource() timesource.Timesource {
	return w.timesource
}

// Relay is used to access any operation related to Waku Relay protocol
func (w *WakuNode) Relay() *relay.WakuRelay {
	if result, ok := w.relay.(*relay.WakuRelay); ok {
		return result
	}
	return nil
}

// Store is used to access any operation related to Waku Store protocol
func (w *WakuNode) Store() store.Store {
	return w.store.(store.Store)
}

// LegacyFilter is used to access any operation related to Waku LegacyFilter protocol
func (w *WakuNode) LegacyFilter() *legacy_filter.WakuFilter {
	if result, ok := w.legacyFilter.(*legacy_filter.WakuFilter); ok {
		return result
	}
	return nil
}

// FilterLightnode is used to access any operation related to Waku Filter protocol Full node feature
func (w *WakuNode) FilterFullNode() *filter.WakuFilterFullNode {
	if result, ok := w.filterFullNode.(*filter.WakuFilterFullNode); ok {
		return result
	}
	return nil
}

// FilterFullNode is used to access any operation related to Waku Filter protocol Light node feature
func (w *WakuNode) FilterLightnode() *filter.WakuFilterLightNode {
	if result, ok := w.filterLightNode.(*filter.WakuFilterLightNode); ok {
		return result
	}
	return nil
}

// PeerManager for getting peer filterv2 protocol
func (w *WakuNode) PeerManager() *peermanager.PeerManager {
	return w.peermanager
}

// Lightpush is used to access any operation related to Waku Lightpush protocol
func (w *WakuNode) Lightpush() *lightpush.WakuLightPush {
	if result, ok := w.lightPush.(*lightpush.WakuLightPush); ok {
		return result
	}
	return nil
}

// DiscV5 is used to access any operation related to DiscoveryV5
func (w *WakuNode) DiscV5() *discv5.DiscoveryV5 {
	if result, ok := w.discoveryV5.(*discv5.DiscoveryV5); ok {
		return result
	}
	return nil
}

// PeerExchange is used to access any operation related to Peer Exchange
func (w *WakuNode) PeerExchange() *peer_exchange.WakuPeerExchange {
	if result, ok := w.peerExchange.(*peer_exchange.WakuPeerExchange); ok {
		return result
	}
	return nil
}

// Rendezvous is used to access any operation related to Rendezvous
func (w *WakuNode) Rendezvous() *rendezvous.Rendezvous {
	if result, ok := w.rendezvous.(*rendezvous.Rendezvous); ok {
		return result
	}
	return nil
}

// Broadcaster is used to access the message broadcaster that is used to push
// messages to different protocols
func (w *WakuNode) Broadcaster() relay.Broadcaster {
	return w.bcaster
}

func (w *WakuNode) mountDiscV5() error {
	discV5Options := []discv5.DiscoveryV5Option{
		discv5.WithBootnodes(w.opts.discV5bootnodes),
		discv5.WithUDPPort(w.opts.udpPort),
		discv5.WithAutoUpdate(w.opts.discV5autoUpdate),
	}

	if w.opts.advertiseAddrs != nil {
		discV5Options = append(discV5Options, discv5.WithAdvertiseAddr(w.opts.advertiseAddrs))
	}

	var err error
	discv5Inst, err := discv5.NewDiscoveryV5(w.opts.privKey, w.localNode, w.peerConnector, w.opts.prometheusReg, w.log, discV5Options...)
	w.discoveryV5 = discv5Inst
	w.peermanager.SetDiscv5(discv5Inst)

	return err
}

func (w *WakuNode) startStore(ctx context.Context, sub *relay.Subscription) error {
	err := w.store.Start(ctx, sub)
	if err != nil {
		w.log.Error("starting store", zap.Error(err))
		return err
	}

	return nil
}

// AddPeer is used to add a peer and the protocols it support to the node peerstore
// TODO: Need to update this for autosharding, to only take contentTopics and optional pubSubTopics or provide an alternate API only for contentTopics.
func (w *WakuNode) AddPeer(address ma.Multiaddr, origin wps.Origin, pubSubTopics []string, protocols ...protocol.ID) (peer.ID, error) {
	pData, err := w.peermanager.AddPeer(address, origin, pubSubTopics, protocols...)
	if err != nil {
		return "", err
	}
	return pData.AddrInfo.ID, nil
}

// AddDiscoveredPeer to add a discovered peer to the node peerStore
func (w *WakuNode) AddDiscoveredPeer(ID peer.ID, addrs []ma.Multiaddr, origin wps.Origin, pubsubTopics []string, connectNow bool) {
	p := service.PeerData{
		Origin: origin,
		AddrInfo: peer.AddrInfo{
			ID:    ID,
			Addrs: addrs,
		},
		PubsubTopics: pubsubTopics,
	}
	w.peermanager.AddDiscoveredPeer(p, connectNow)
}

// DialPeerWithMultiAddress is used to connect to a peer using a multiaddress
func (w *WakuNode) DialPeerWithMultiAddress(ctx context.Context, address ma.Multiaddr) error {
	info, err := peer.AddrInfoFromP2pAddr(address)
	if err != nil {
		return err
	}

	return w.connect(ctx, *info)
}

// DialPeer is used to connect to a peer using a string containing a multiaddress
func (w *WakuNode) DialPeer(ctx context.Context, address string) error {
	p, err := ma.NewMultiaddr(address)
	if err != nil {
		return err
	}

	info, err := peer.AddrInfoFromP2pAddr(p)
	if err != nil {
		return err
	}

	return w.connect(ctx, *info)
}

// DialPeerWithInfo is used to connect to a peer using its address information
func (w *WakuNode) DialPeerWithInfo(ctx context.Context, peerInfo peer.AddrInfo) error {
	return w.connect(ctx, peerInfo)
}

func (w *WakuNode) connect(ctx context.Context, info peer.AddrInfo) error {
	err := w.host.Connect(ctx, info)
	if err != nil {
		w.host.Peerstore().(wps.WakuPeerstore).AddConnFailure(info)
		return err
	}

	for _, addr := range info.Addrs {
		// TODO: this is a temporary fix
		// host.Connect adds the addresses with a TempAddressTTL
		// however, identify will filter out all non IP addresses
		// and expire all temporary addrs. So in the meantime, let's
		// store dns4 addresses with a RecentlyConnectedAddrTTL, otherwise
		// it will have trouble with the status fleet circuit relay addresses
		// See https://github.com/libp2p/go-libp2p/issues/2550
		_, err := addr.ValueForProtocol(ma.P_DNS4)
		if err == nil {
			w.host.Peerstore().AddAddrs(info.ID, info.Addrs, peerstore.RecentlyConnectedAddrTTL)
		}
	}

	w.host.Peerstore().(wps.WakuPeerstore).ResetConnFailures(info)

	w.metrics.RecordDial()

	return nil
}

// DialPeerByID is used to connect to an already known peer
func (w *WakuNode) DialPeerByID(ctx context.Context, peerID peer.ID) error {
	info := w.host.Peerstore().PeerInfo(peerID)
	return w.connect(ctx, info)
}

// ClosePeerByAddress is used to disconnect from a peer using its multiaddress
func (w *WakuNode) ClosePeerByAddress(address string) error {
	p, err := ma.NewMultiaddr(address)
	if err != nil {
		return err
	}

	// Extract the peer ID from the multiaddr.
	info, err := peer.AddrInfoFromP2pAddr(p)
	if err != nil {
		return err
	}

	return w.ClosePeerById(info.ID)
}

// ClosePeerById is used to close a connection to a peer
func (w *WakuNode) ClosePeerById(id peer.ID) error {
	err := w.host.Network().ClosePeer(id)
	if err != nil {
		return err
	}
	return nil
}

// PeerCount return the number of connected peers
func (w *WakuNode) PeerCount() int {
	return len(w.host.Network().Peers())
}

// PeerStats returns a list of peers and the protocols supported by them
func (w *WakuNode) PeerStats() PeerStats {
	p := make(PeerStats)
	for _, peerID := range w.host.Network().Peers() {
		protocols, err := w.host.Peerstore().GetProtocols(peerID)
		if err != nil {
			continue
		}
		p[peerID] = protocols
	}
	return p
}

// Set the bootnodes on discv5
func (w *WakuNode) SetDiscV5Bootnodes(nodes []*enode.Node) error {
	w.opts.discV5bootnodes = nodes
	return w.DiscV5().SetBootnodes(nodes)
}

// Peers return the list of peers, addresses, protocols supported and connection status
func (w *WakuNode) Peers() ([]*Peer, error) {
	var peers []*Peer
	for _, peerId := range w.host.Peerstore().Peers() {
		connected := w.host.Network().Connectedness(peerId) == network.Connected
		protocols, err := w.host.Peerstore().GetProtocols(peerId)
		if err != nil {
			return nil, err
		}

		addrs := utils.EncapsulatePeerID(peerId, w.host.Peerstore().Addrs(peerId)...)
		topics, err := w.host.Peerstore().(*wps.WakuPeerstoreImpl).PubSubTopics(peerId)
		if err != nil {
			return nil, err
		}
		peers = append(peers, &Peer{
			ID:           peerId,
			Protocols:    protocols,
			Connected:    connected,
			Addrs:        addrs,
			PubsubTopics: topics,
		})
	}
	return peers, nil
}

// PeersByShard filters peers based on shard information following static sharding
func (w *WakuNode) PeersByStaticShard(cluster uint16, shard uint16) peer.IDSlice {
	pTopic := wakuprotocol.NewStaticShardingPubsubTopic(cluster, shard).String()
	return w.peerstore.(wps.WakuPeerstore).PeersByPubSubTopic(pTopic)
}

// PeersByContentTopics filters peers based on contentTopic
func (w *WakuNode) PeersByContentTopic(contentTopic string) peer.IDSlice {
	pTopic, err := wakuprotocol.GetPubSubTopicFromContentTopic(contentTopic)
	if err != nil {
		return nil
	}
	return w.peerstore.(wps.WakuPeerstore).PeersByPubSubTopic(pTopic)
}

func (w *WakuNode) findRelayNodes(ctx context.Context) {
	defer w.wg.Done()

	// Feed peers more often right after the bootstrap, then backoff
	bo := backoffv4.NewExponentialBackOff()
	bo.InitialInterval = 15 * time.Second
	bo.Multiplier = 3
	bo.MaxInterval = 1 * time.Hour
	bo.MaxElapsedTime = 0 // never stop
	t := backoffv4.NewTicker(bo)
	defer t.Stop()
	for {
		select {
		case <-t.C:
		case <-ctx.Done():
			return
		}

		peers, err := w.Peers()
		if err != nil {
			w.log.Error("failed to fetch peers", zap.Error(err))
			continue
		}

		// Shuffle peers
		rand.Shuffle(len(peers), func(i, j int) { peers[i], peers[j] = peers[j], peers[i] })

		for _, p := range peers {
			info := w.Host().Peerstore().PeerInfo(p.ID)
			supportedProtocols, err := w.Host().Peerstore().SupportsProtocols(p.ID, proto.ProtoIDv2Hop)
			if err != nil {
				w.log.Error("could not check supported protocols", zap.Error(err))
				continue
			}

			if len(supportedProtocols) == 0 {
				continue
			}

			select {
			case <-ctx.Done():
				w.log.Debug("context done, auto-relay has enough peers")
				return

			case w.circuitRelayNodes <- info:
				w.log.Debug("published auto-relay peer info", zap.Any("peer-id", p.ID))
			}
		}
	}
}

func GetNodesFromDNSDiscovery(logger *zap.Logger, ctx context.Context, nameServer string, discoveryURLs []string) []dnsdisc.DiscoveredNode {
	var discoveredNodes []dnsdisc.DiscoveredNode
	for _, url := range discoveryURLs {
		logger.Info("attempting DNS discovery with ", zap.String("URL", url))
		nodes, err := dnsdisc.RetrieveNodes(ctx, url, dnsdisc.WithNameserver(nameServer))
		if err != nil {
			logger.Warn("dns discovery error ", zap.Error(err))
		} else {
			var discPeerInfo []peer.AddrInfo
			for _, n := range nodes {
				discPeerInfo = append(discPeerInfo, n.PeerInfo)
			}
			logger.Info("found dns entries ", zap.Any("nodes", discPeerInfo))
			discoveredNodes = append(discoveredNodes, nodes...)
		}
	}
	return discoveredNodes
}

func GetDiscv5Option(dnsDiscoveredNodes []dnsdisc.DiscoveredNode, discv5Nodes []string, port uint, autoUpdate bool) (WakuNodeOption, error) {
	var bootnodes []*enode.Node
	for _, addr := range discv5Nodes {
		bootnode, err := enode.Parse(enode.ValidSchemes, addr)
		if err != nil {
			return nil, err
		}
		bootnodes = append(bootnodes, bootnode)
	}

	for _, n := range dnsDiscoveredNodes {
		if n.ENR != nil {
			bootnodes = append(bootnodes, n.ENR)
		}
	}

	return WithDiscoveryV5(port, bootnodes, autoUpdate), nil
}

func (w *WakuNode) ClusterID() uint16 {
	return w.opts.clusterID
}
