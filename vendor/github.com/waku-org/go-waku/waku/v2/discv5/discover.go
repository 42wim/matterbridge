package discv5

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/waku-org/go-discover/discover"
	"github.com/waku-org/go-waku/logging"
	"github.com/waku-org/go-waku/waku/v2/peerstore"
	wenr "github.com/waku-org/go-waku/waku/v2/protocol/enr"
	"github.com/waku-org/go-waku/waku/v2/service"
	"github.com/waku-org/go-waku/waku/v2/utils"
	"go.uber.org/zap"

	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/ethereum/go-ethereum/p2p/enr"
	"github.com/ethereum/go-ethereum/p2p/nat"
)

var ErrNoDiscV5Listener = errors.New("no discv5 listener")

// PeerConnector will subscribe to a channel containing the information for all peers found by this discovery protocol
type PeerConnector interface {
	Subscribe(context.Context, <-chan service.PeerData)
}

type DiscoveryV5 struct {
	params    *discV5Parameters
	host      host.Host
	config    discover.Config
	udpAddr   *net.UDPAddr
	listener  *discover.UDPv5
	localnode *enode.LocalNode
	metrics   Metrics

	peerConnector PeerConnector
	NAT           nat.Interface

	log *zap.Logger

	*service.CommonDiscoveryService
}

type discV5Parameters struct {
	autoUpdate    bool
	autoFindPeers bool
	bootnodes     map[enode.ID]*enode.Node
	udpPort       uint
	advertiseAddr []multiaddr.Multiaddr
	loopPredicate func(*enode.Node) bool
}

type DiscoveryV5Option func(*discV5Parameters)

var protocolID = [6]byte{'d', '5', 'w', 'a', 'k', 'u'}

const peerDelay = 100 * time.Millisecond
const bucketSize = 16
const delayBetweenDiscoveredPeerCnt = 5 * time.Second

func WithAutoUpdate(autoUpdate bool) DiscoveryV5Option {
	return func(params *discV5Parameters) {
		params.autoUpdate = autoUpdate
	}
}

// WithBootnodes is an option used to specify the bootstrap nodes to use with DiscV5
func WithBootnodes(bootnodes []*enode.Node) DiscoveryV5Option {
	return func(params *discV5Parameters) {
		params.bootnodes = make(map[enode.ID]*enode.Node)
		for _, b := range bootnodes {
			params.bootnodes[b.ID()] = b
		}
	}
}

func WithAdvertiseAddr(addr []multiaddr.Multiaddr) DiscoveryV5Option {
	return func(params *discV5Parameters) {
		params.advertiseAddr = addr
	}
}

func WithUDPPort(port uint) DiscoveryV5Option {
	return func(params *discV5Parameters) {
		params.udpPort = port
	}
}

func WithPredicate(predicate func(*enode.Node) bool) DiscoveryV5Option {
	return func(params *discV5Parameters) {
		params.loopPredicate = predicate
	}
}

func WithAutoFindPeers(find bool) DiscoveryV5Option {
	return func(params *discV5Parameters) {
		params.autoFindPeers = find
	}
}

// DefaultOptions contains the default list of options used when setting up DiscoveryV5
func DefaultOptions() []DiscoveryV5Option {
	return []DiscoveryV5Option{
		WithUDPPort(9000),
		WithAutoFindPeers(true),
	}
}

// NewDiscoveryV5 returns a new instance of a DiscoveryV5 struct
func NewDiscoveryV5(priv *ecdsa.PrivateKey, localnode *enode.LocalNode, peerConnector PeerConnector, reg prometheus.Registerer, log *zap.Logger, opts ...DiscoveryV5Option) (*DiscoveryV5, error) {
	params := new(discV5Parameters)
	optList := DefaultOptions()
	optList = append(optList, opts...)
	for _, opt := range optList {
		opt(params)
	}

	logger := log.Named("discv5")

	var NAT nat.Interface
	if params.advertiseAddr == nil {
		NAT = nat.Any()
	}

	var bootnodes []*enode.Node
	for _, bootnode := range params.bootnodes {
		bootnodes = append(bootnodes, bootnode)
	}

	return &DiscoveryV5{
		params:                 params,
		peerConnector:          peerConnector,
		NAT:                    NAT,
		CommonDiscoveryService: service.NewCommonDiscoveryService(),
		localnode:              localnode,
		metrics:                newMetrics(reg),
		config: discover.Config{
			PrivateKey: priv,
			Bootnodes:  bootnodes,
			V5Config: discover.V5Config{
				ProtocolID: &protocolID,
			},
		},
		udpAddr: &net.UDPAddr{
			IP:   net.IPv4zero,
			Port: int(params.udpPort),
		},
		log: logger,
	}, nil
}

func (d *DiscoveryV5) Node() *enode.Node {
	return d.localnode.Node()
}

func (d *DiscoveryV5) listen(ctx context.Context) error {
	conn, err := net.ListenUDP("udp", d.udpAddr)
	if err != nil {
		return err
	}

	d.udpAddr = conn.LocalAddr().(*net.UDPAddr)

	if d.NAT != nil && !d.udpAddr.IP.IsLoopback() {
		d.WaitGroup().Add(1)
		go func() {
			defer d.WaitGroup().Done()
			nat.Map(d.NAT, ctx.Done(), "udp", d.udpAddr.Port, d.udpAddr.Port, "go-waku discv5 discovery")
		}()

	}

	d.params.udpPort = uint(d.udpAddr.Port)
	d.localnode.SetFallbackUDP(d.udpAddr.Port)

	listener, err := discover.ListenV5(ctx, conn, d.localnode, d.config)
	if err != nil {
		return err
	}

	d.listener = listener

	d.log.Info("started Discovery V5",
		zap.Stringer("listening", d.udpAddr),
		logging.TCPAddr("advertising", d.localnode.Node().IP(), d.localnode.Node().TCP()))
	d.log.Info("Discovery V5: discoverable ENR ", logging.ENode("enr", d.localnode.Node()))

	return nil
}

// Sets the host to be able to mount or consume a protocol
func (d *DiscoveryV5) SetHost(h host.Host) {
	d.host = h
}

// only works if the discovery v5 hasn't been started yet.
func (d *DiscoveryV5) Start(ctx context.Context) error {
	return d.CommonDiscoveryService.Start(ctx, d.start)
}

func (d *DiscoveryV5) start() error {
	d.peerConnector.Subscribe(d.Context(), d.GetListeningChan())

	err := d.listen(d.Context())
	if err != nil {
		return err
	}

	if d.params.autoFindPeers {
		d.WaitGroup().Add(1)
		go func() {
			defer d.WaitGroup().Done()
			d.runDiscoveryV5Loop(d.Context())
		}()
	}

	return nil
}

// SetBootnodes is used to setup the bootstrap nodes to use for discovering new peers
func (d *DiscoveryV5) SetBootnodes(nodes []*enode.Node) error {
	if d.listener == nil {
		return ErrNoDiscV5Listener
	}

	return d.listener.SetFallbackNodes(nodes)
}

// Stop is a function that stops the execution of DiscV5.
// only works if the discovery v5 is in running state
// so we can assume that cancel method is set
func (d *DiscoveryV5) Stop() {
	defer func() {
		if r := recover(); r != nil {
			d.log.Info("recovering from panic and quitting")
		}
	}()
	d.CommonDiscoveryService.Stop(func() {
		if d.listener != nil {
			d.listener.Close()
			d.listener = nil
			d.log.Info("stopped Discovery V5")
		}
	})
}

func isWakuNode(node *enode.Node) bool {
	enrField := new(wenr.WakuEnrBitfield)
	if err := node.Record().Load(enr.WithEntry(wenr.WakuENRField, &enrField)); err != nil {
		if !enr.IsNotFound(err) {
			utils.Logger().Named("discv5").Error("could not retrieve waku2 ENR field for enr ", zap.Any("node", node))
		}
		return false
	}

	if enrField != nil {
		return *enrField != uint8(0) // #RFC 31 requirement
	}

	return false
}

func (d *DiscoveryV5) evaluateNode() func(node *enode.Node) bool {
	return func(node *enode.Node) bool {
		if node == nil {
			return false
		}

		//  node filtering based on ENR; we do not filter based on ENR in the first waku discv5 beta stage
		if !isWakuNode(node) {
			d.log.Debug("peer is not waku node", logging.ENode("enr", node))
			return false
		}

		_, err := wenr.EnodeToPeerInfo(node)
		if err != nil {
			d.metrics.RecordError(peerInfoFailure)
			d.log.Error("obtaining peer info from enode", logging.ENode("enr", node), zap.Error(err))
			return false
		}

		return true
	}
}

// Predicate is a function that is applied to an iterator to filter the nodes to be retrieved according to some logic
type Predicate func(enode.Iterator) enode.Iterator

// PeerIterator gets random nodes from DHT via discv5 listener.
// Used for caching enr address in peerExchange
// Used for connecting to peers in discovery_connector
func (d *DiscoveryV5) PeerIterator(predicate ...Predicate) (enode.Iterator, error) {
	if d.listener == nil {
		return nil, ErrNoDiscV5Listener
	}

	iterator := enode.Filter(d.listener.RandomNodes(), d.evaluateNode())
	if d.params.loopPredicate != nil {
		iterator = enode.Filter(iterator, d.params.loopPredicate)
	}

	for _, p := range predicate {
		iterator = p(iterator)
	}

	return iterator, nil
}

func (d *DiscoveryV5) Iterate(ctx context.Context, iterator enode.Iterator, onNode func(*enode.Node, peer.AddrInfo) error) {
	defer iterator.Close()

	peerCnt := 0
	for DelayedHasNext(ctx, iterator, &peerCnt) {
		_, addresses, err := wenr.Multiaddress(iterator.Node())
		if err != nil {
			d.metrics.RecordError(peerInfoFailure)
			d.log.Error("extracting multiaddrs from enr", zap.Error(err))
			continue
		}

		peerAddrs, err := peer.AddrInfosFromP2pAddrs(addresses...)
		if err != nil {
			d.metrics.RecordError(peerInfoFailure)
			d.log.Error("converting multiaddrs to addrinfos", zap.Error(err))
			continue
		}

		if len(peerAddrs) != 0 {
			err := onNode(iterator.Node(), peerAddrs[0])
			if err != nil {
				d.log.Error("processing node", zap.Error(err))
			}
		}

		select {
		case <-ctx.Done():
			return
		default:
		}
	}
}

func DelayedHasNext(ctx context.Context, iterator enode.Iterator, peerCnt *int) bool {
	// Delay if .Next() is too fast
	start := time.Now()
	hasNext := iterator.Next()
	if !hasNext {
		return false
	}

	elapsed := time.Since(start)
	if elapsed < peerDelay {
		t := time.NewTimer(peerDelay - elapsed)
		select {
		case <-ctx.Done():
			return false
		case <-t.C:
			t.Stop()
		}
	}

	*peerCnt++
	if *peerCnt == bucketSize { // Delay every bucketSize peers discovered
		*peerCnt = 0
		t := time.NewTimer(delayBetweenDiscoveredPeerCnt)
		select {
		case <-ctx.Done():
			return false
		case <-t.C:
			t.Stop()
		}
	}

	return true
}

// DefaultPredicate contains the conditions to be applied when filtering peers discovered via discv5
func (d *DiscoveryV5) DefaultPredicate() Predicate {
	return FilterPredicate(func(n *enode.Node) bool {
		localRS, err := wenr.RelaySharding(d.localnode.Node().Record())
		if err != nil {
			return false
		}

		if localRS == nil { // No shard registered, so no need to check for shards
			return true
		}

		if _, ok := d.params.bootnodes[n.ID()]; ok {
			return true // The record is a bootnode. Assume it's valid and dont filter it out
		}

		nodeRS, err := wenr.RelaySharding(n.Record())
		if err != nil {
			d.log.Debug("failed to get relay shards from node record", logging.ENode("node", n), zap.Error(err))
			return false
		}

		if nodeRS == nil {
			// Node has no shards registered.
			return false
		}

		if nodeRS.ClusterID != localRS.ClusterID {
			return false
		}

		// Contains any
		for _, idx := range localRS.ShardIDs {
			if nodeRS.Contains(localRS.ClusterID, idx) {
				return true
			}
		}

		return false
	})
}

// Iterates over the nodes found via discv5 belonging to the node's current shard, and sends them to peerConnector
func (d *DiscoveryV5) peerLoop(ctx context.Context) error {
	iterator, err := d.PeerIterator(d.DefaultPredicate())
	if err != nil {
		d.metrics.RecordError(iteratorFailure)
		return fmt.Errorf("obtaining iterator: %w", err)
	}

	defer iterator.Close()

	d.Iterate(ctx, iterator, func(n *enode.Node, p peer.AddrInfo) error {
		peer := service.PeerData{
			Origin:   peerstore.Discv5,
			AddrInfo: p,
			ENR:      n,
		}

		if d.PushToChan(peer) {
			d.log.Debug("published peer into peer channel", logging.HostID("peerID", peer.AddrInfo.ID))
		} else {
			d.log.Debug("could not publish peer into peer channel", logging.HostID("peerID", peer.AddrInfo.ID))
		}

		return nil
	})

	return nil
}

func (d *DiscoveryV5) runDiscoveryV5Loop(ctx context.Context) {
	if len(d.config.Bootnodes) > 0 {
		localRS, err := wenr.RelaySharding(d.localnode.Node().Record())
		if err == nil && localRS != nil {
			iterator := d.DefaultPredicate()(enode.IterNodes(d.config.Bootnodes))
			validBootCount := 0
			for iterator.Next() {
				validBootCount++
			}

			if validBootCount == 0 {
				d.log.Warn("no discv5 bootstrap nodes share this node configured shards")
			}
		}
	}

restartLoop:
	for {
		err := d.peerLoop(ctx)
		if err != nil {
			d.log.Debug("iterating discv5", zap.Error(err))
		}

		t := time.NewTimer(5 * time.Second)
		select {
		case <-t.C:
			t.Stop()
		case <-ctx.Done():
			t.Stop()
			break restartLoop
		}
	}
	d.log.Warn("Discv5 loop stopped")
}
