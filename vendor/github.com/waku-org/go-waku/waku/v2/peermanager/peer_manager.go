package peermanager

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/p2p/enr"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/event"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/peerstore"
	"github.com/libp2p/go-libp2p/core/protocol"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/waku-org/go-waku/logging"
	"github.com/waku-org/go-waku/waku/v2/discv5"
	wps "github.com/waku-org/go-waku/waku/v2/peerstore"
	waku_proto "github.com/waku-org/go-waku/waku/v2/protocol"
	wenr "github.com/waku-org/go-waku/waku/v2/protocol/enr"
	"github.com/waku-org/go-waku/waku/v2/protocol/relay"
	"github.com/waku-org/go-waku/waku/v2/service"

	"go.uber.org/zap"
)

// NodeTopicDetails stores pubSubTopic related data like topicHandle for the node.
type NodeTopicDetails struct {
	topic *pubsub.Topic
}

// WakuProtoInfo holds protocol specific info
// To be used at a later stage to set various config such as criteria for peer management specific to each Waku protocols
// This should make peer-manager agnostic to protocol
type WakuProtoInfo struct {
	waku2ENRBitField uint8
}

// PeerManager applies various controls and manage connections towards peers.
type PeerManager struct {
	peerConnector          *PeerConnectionStrategy
	maxPeers               int
	maxRelayPeers          int
	logger                 *zap.Logger
	InRelayPeersTarget     int
	OutRelayPeersTarget    int
	host                   host.Host
	serviceSlots           *ServiceSlots
	ctx                    context.Context
	sub                    event.Subscription
	topicMutex             sync.RWMutex
	subRelayTopics         map[string]*NodeTopicDetails
	discoveryService       *discv5.DiscoveryV5
	wakuprotoToENRFieldMap map[protocol.ID]WakuProtoInfo
}

// PeerSelection provides various options based on which Peer is selected from a list of peers.
type PeerSelection int

const (
	Automatic PeerSelection = iota
	LowestRTT
)

// ErrNoPeersAvailable is emitted when no suitable peers are found for
// some protocol
var ErrNoPeersAvailable = errors.New("no suitable peers found")

const peerConnectivityLoopSecs = 15
const maxConnsToPeerRatio = 5

// 80% relay peers 20% service peers
func relayAndServicePeers(maxConnections int) (int, int) {
	return maxConnections - maxConnections/5, maxConnections / 5
}

// 66% inRelayPeers 33% outRelayPeers
func inAndOutRelayPeers(relayPeers int) (int, int) {
	outRelayPeers := relayPeers / 3
	//
	const minOutRelayConns = 10
	if outRelayPeers < minOutRelayConns {
		outRelayPeers = minOutRelayConns
	}
	return relayPeers - outRelayPeers, outRelayPeers
}

// NewPeerManager creates a new peerManager instance.
func NewPeerManager(maxConnections int, maxPeers int, logger *zap.Logger) *PeerManager {

	maxRelayPeers, _ := relayAndServicePeers(maxConnections)
	inRelayPeersTarget, outRelayPeersTarget := inAndOutRelayPeers(maxRelayPeers)

	if maxPeers == 0 || maxConnections > maxPeers {
		maxPeers = maxConnsToPeerRatio * maxConnections
	}

	pm := &PeerManager{
		logger:                 logger.Named("peer-manager"),
		maxRelayPeers:          maxRelayPeers,
		InRelayPeersTarget:     inRelayPeersTarget,
		OutRelayPeersTarget:    outRelayPeersTarget,
		serviceSlots:           NewServiceSlot(),
		subRelayTopics:         make(map[string]*NodeTopicDetails),
		maxPeers:               maxPeers,
		wakuprotoToENRFieldMap: map[protocol.ID]WakuProtoInfo{},
	}
	logger.Info("PeerManager init values", zap.Int("maxConnections", maxConnections),
		zap.Int("maxRelayPeers", maxRelayPeers),
		zap.Int("outRelayPeersTarget", outRelayPeersTarget),
		zap.Int("inRelayPeersTarget", pm.InRelayPeersTarget),
		zap.Int("maxPeers", maxPeers))

	return pm
}

// SetDiscv5 sets the discoveryv5 service to be used for peer discovery.
func (pm *PeerManager) SetDiscv5(discv5 *discv5.DiscoveryV5) {
	pm.discoveryService = discv5
}

// SetHost sets the host to be used in order to access the peerStore.
func (pm *PeerManager) SetHost(host host.Host) {
	pm.host = host
}

// SetPeerConnector sets the peer connector to be used for establishing relay connections.
func (pm *PeerManager) SetPeerConnector(pc *PeerConnectionStrategy) {
	pm.peerConnector = pc
}

// Start starts the processing to be done by peer manager.
func (pm *PeerManager) Start(ctx context.Context) {

	pm.RegisterWakuProtocol(relay.WakuRelayID_v200, relay.WakuRelayENRField)

	pm.ctx = ctx
	if pm.sub != nil {
		go pm.peerEventLoop(ctx)
	}
	go pm.connectivityLoop(ctx)
}

// This is a connectivity loop, which currently checks and prunes inbound connections.
func (pm *PeerManager) connectivityLoop(ctx context.Context) {
	pm.connectToRelayPeers()
	t := time.NewTicker(peerConnectivityLoopSecs * time.Second)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			pm.connectToRelayPeers()
		}
	}
}

// GroupPeersByDirection returns all the connected peers in peer store grouped by Inbound or outBound direction
func (pm *PeerManager) GroupPeersByDirection(specificPeers ...peer.ID) (inPeers peer.IDSlice, outPeers peer.IDSlice, err error) {
	if len(specificPeers) == 0 {
		specificPeers = pm.host.Network().Peers()
	}

	for _, p := range specificPeers {
		direction, err := pm.host.Peerstore().(wps.WakuPeerstore).Direction(p)
		if err == nil {
			if direction == network.DirInbound {
				inPeers = append(inPeers, p)
			} else if direction == network.DirOutbound {
				outPeers = append(outPeers, p)
			}
		} else {
			pm.logger.Error("failed to retrieve peer direction",
				logging.HostID("peerID", p), zap.Error(err))
		}
	}
	return inPeers, outPeers, nil
}

// getRelayPeers - Returns list of in and out peers supporting WakuRelayProtocol within specifiedPeers.
// If specifiedPeers is empty, it checks within all peers in peerStore.
func (pm *PeerManager) getRelayPeers(specificPeers ...peer.ID) (inRelayPeers peer.IDSlice, outRelayPeers peer.IDSlice) {
	//Group peers by their connected direction inbound or outbound.
	inPeers, outPeers, err := pm.GroupPeersByDirection(specificPeers...)
	if err != nil {
		return
	}
	pm.logger.Debug("number of peers connected", zap.Int("inPeers", inPeers.Len()),
		zap.Int("outPeers", outPeers.Len()))

	//Need to filter peers to check if they support relay
	if inPeers.Len() != 0 {
		inRelayPeers, _ = pm.FilterPeersByProto(inPeers, relay.WakuRelayID_v200)
	}
	if outPeers.Len() != 0 {
		outRelayPeers, _ = pm.FilterPeersByProto(outPeers, relay.WakuRelayID_v200)
	}
	return
}

// ensureMinRelayConnsPerTopic makes sure there are min of D conns per pubsubTopic.
// If not it will look into peerStore to initiate more connections.
// If peerStore doesn't have enough peers, will wait for discv5 to find more and try in next cycle
func (pm *PeerManager) ensureMinRelayConnsPerTopic() {
	pm.topicMutex.RLock()
	defer pm.topicMutex.RUnlock()
	for topicStr, topicInst := range pm.subRelayTopics {

		// @cammellos reported that ListPeers returned an invalid number of
		// peers. This will ensure that the peers returned by this function
		// match those peers that are currently connected
		curPeerLen := 0
		for _, p := range topicInst.topic.ListPeers() {
			if pm.host.Network().Connectedness(p) == network.Connected {
				curPeerLen++
			}
		}
		if curPeerLen < waku_proto.GossipSubOptimalFullMeshSize {
			pm.logger.Debug("subscribed topic is unhealthy, initiating more connections to maintain health",
				zap.String("pubSubTopic", topicStr), zap.Int("connectedPeerCount", curPeerLen),
				zap.Int("optimumPeers", waku_proto.GossipSubOptimalFullMeshSize))
			//Find not connected peers.
			notConnectedPeers := pm.getNotConnectedPers(topicStr)
			if notConnectedPeers.Len() == 0 {
				pm.logger.Debug("could not find any peers in peerstore to connect to, discovering more", zap.String("pubSubTopic", topicStr))
				pm.discoverPeersByPubsubTopics([]string{topicStr}, relay.WakuRelayID_v200, pm.ctx, 2)
				continue
			}
			pm.logger.Debug("connecting to eligible peers in peerstore", zap.String("pubSubTopic", topicStr))
			//Connect to eligible peers.
			numPeersToConnect := waku_proto.GossipSubOptimalFullMeshSize - curPeerLen

			if numPeersToConnect > notConnectedPeers.Len() {
				numPeersToConnect = notConnectedPeers.Len()
			}
			pm.connectToPeers(notConnectedPeers[0:numPeersToConnect])
		}
	}
}

// connectToRelayPeers ensures minimum D connections are there for each pubSubTopic.
// If not, initiates connections to additional peers.
// It also checks for incoming relay connections and prunes once they cross inRelayTarget
func (pm *PeerManager) connectToRelayPeers() {
	//Check for out peer connections and connect to more peers.
	pm.ensureMinRelayConnsPerTopic()

	inRelayPeers, outRelayPeers := pm.getRelayPeers()
	pm.logger.Debug("number of relay peers connected",
		zap.Int("in", inRelayPeers.Len()),
		zap.Int("out", outRelayPeers.Len()))
	if inRelayPeers.Len() > 0 &&
		inRelayPeers.Len() > pm.InRelayPeersTarget {
		pm.pruneInRelayConns(inRelayPeers)
	}
}

// connectToPeers connects to peers provided in the list if the addresses have not expired.
func (pm *PeerManager) connectToPeers(peers peer.IDSlice) {
	for _, peerID := range peers {
		peerData := AddrInfoToPeerData(wps.PeerManager, peerID, pm.host)
		if peerData == nil {
			continue
		}
		pm.peerConnector.PushToChan(*peerData)
	}
}

// getNotConnectedPers returns peers for a pubSubTopic that are not connected.
func (pm *PeerManager) getNotConnectedPers(pubsubTopic string) (notConnectedPeers peer.IDSlice) {
	var peerList peer.IDSlice
	if pubsubTopic == "" {
		peerList = pm.host.Peerstore().Peers()
	} else {
		peerList = pm.host.Peerstore().(*wps.WakuPeerstoreImpl).PeersByPubSubTopic(pubsubTopic)
	}
	for _, peerID := range peerList {
		if pm.host.Network().Connectedness(peerID) != network.Connected {
			notConnectedPeers = append(notConnectedPeers, peerID)
		}
	}
	return
}

// pruneInRelayConns prune any incoming relay connections crossing derived inrelayPeerTarget
func (pm *PeerManager) pruneInRelayConns(inRelayPeers peer.IDSlice) {

	//Start disconnecting peers, based on what?
	//For now no preference is used
	//TODO: Need to have more intelligent way of doing this, maybe peer scores.
	//TODO: Keep optimalPeersRequired for a pubSubTopic in mind while pruning connections to peers.
	pm.logger.Info("peer connections exceed target relay peers, hence pruning",
		zap.Int("cnt", inRelayPeers.Len()), zap.Int("target", pm.InRelayPeersTarget))
	for pruningStartIndex := pm.InRelayPeersTarget; pruningStartIndex < inRelayPeers.Len(); pruningStartIndex++ {
		p := inRelayPeers[pruningStartIndex]
		err := pm.host.Network().ClosePeer(p)
		if err != nil {
			pm.logger.Warn("failed to disconnect connection towards peer",
				logging.HostID("peerID", p))
		}
		pm.logger.Debug("successfully disconnected connection towards peer",
			logging.HostID("peerID", p))
	}
}

func (pm *PeerManager) processPeerENR(p *service.PeerData) []protocol.ID {
	shards, err := wenr.RelaySharding(p.ENR.Record())
	if err != nil {
		pm.logger.Error("could not derive relayShards from ENR", zap.Error(err),
			logging.HostID("peer", p.AddrInfo.ID), zap.String("enr", p.ENR.String()))
	} else {
		if shards != nil {
			p.PubsubTopics = make([]string, 0)
			topics := shards.Topics()
			for _, topic := range topics {
				topicStr := topic.String()
				p.PubsubTopics = append(p.PubsubTopics, topicStr)
			}
		} else {
			pm.logger.Debug("ENR doesn't have relay shards", logging.HostID("peer", p.AddrInfo.ID))
		}
	}
	supportedProtos := []protocol.ID{}
	//Identify and specify protocols supported by the peer based on the discovered peer's ENR
	var enrField wenr.WakuEnrBitfield
	if err := p.ENR.Record().Load(enr.WithEntry(wenr.WakuENRField, &enrField)); err == nil {
		for proto, protoENR := range pm.wakuprotoToENRFieldMap {
			protoENRField := protoENR.waku2ENRBitField
			if protoENRField&enrField != 0 {
				supportedProtos = append(supportedProtos, proto)
				//Add Service peers to serviceSlots.
				pm.addPeerToServiceSlot(proto, p.AddrInfo.ID)
			}
		}
	}
	return supportedProtos
}

// AddDiscoveredPeer to add dynamically discovered peers.
// Note that these peers will not be set in service-slots.
func (pm *PeerManager) AddDiscoveredPeer(p service.PeerData, connectNow bool) {
	//Doing this check again inside addPeer, in order to avoid additional complexity of rollingBack other changes.
	if pm.maxPeers <= pm.host.Peerstore().Peers().Len() {
		return
	}
	//Check if the peer is already present, if so skip adding
	_, err := pm.host.Peerstore().(wps.WakuPeerstore).Origin(p.AddrInfo.ID)
	if err == nil {
		enr, err := pm.host.Peerstore().(wps.WakuPeerstore).ENR(p.AddrInfo.ID)
		// Verifying if the enr record is more recent (DiscV5 and peer exchange can return peers already seen)
		if err == nil && enr.Record().Seq() >= p.ENR.Seq() {
			return
		}
		if err != nil {
			//Peer is already in peer-store but it doesn't have an enr, but discovered peer has ENR
			pm.logger.Info("peer already found in peerstore, but doesn't have an ENR record, re-adding",
				logging.HostID("peer", p.AddrInfo.ID), zap.Uint64("newENRSeq", p.ENR.Seq()))
		} else {
			//Peer is already in peer-store but stored ENR is older than discovered one.
			pm.logger.Info("peer already found in peerstore, but re-adding it as ENR sequence is higher than locally stored",
				logging.HostID("peer", p.AddrInfo.ID), zap.Uint64("newENRSeq", p.ENR.Seq()), zap.Uint64("storedENRSeq", enr.Record().Seq()))
		}
	}

	supportedProtos := []protocol.ID{}
	if len(p.PubsubTopics) == 0 && p.ENR != nil {
		// Try to fetch shard info and supported protocols from ENR to arrive at pubSub topics.
		supportedProtos = pm.processPeerENR(&p)
	}

	_ = pm.addPeer(p.AddrInfo.ID, p.AddrInfo.Addrs, p.Origin, p.PubsubTopics, supportedProtos...)

	if p.ENR != nil {
		err := pm.host.Peerstore().(wps.WakuPeerstore).SetENR(p.AddrInfo.ID, p.ENR)
		if err != nil {
			pm.logger.Error("could not store enr", zap.Error(err),
				logging.HostID("peer", p.AddrInfo.ID), zap.String("enr", p.ENR.String()))
		}
	}
	if connectNow {
		pm.logger.Debug("connecting now to discovered peer", logging.HostID("peer", p.AddrInfo.ID))
		go pm.peerConnector.PushToChan(p)
	}
}

// addPeer adds peer to only the peerStore.
// It also sets additional metadata such as origin, ENR and supported protocols
func (pm *PeerManager) addPeer(ID peer.ID, addrs []ma.Multiaddr, origin wps.Origin, pubSubTopics []string, protocols ...protocol.ID) error {
	if pm.maxPeers <= pm.host.Peerstore().Peers().Len() {
		pm.logger.Error("could not add peer as peer store capacity is reached", logging.HostID("peer", ID), zap.Int("capacity", pm.maxPeers))
		return errors.New("peer store capacity reached")
	}
	pm.logger.Info("adding peer to peerstore", logging.HostID("peer", ID))
	if origin == wps.Static {
		pm.host.Peerstore().AddAddrs(ID, addrs, peerstore.PermanentAddrTTL)
	} else {
		//Need to re-evaluate the address expiry
		// For now expiring them with default addressTTL which is an hour.
		pm.host.Peerstore().AddAddrs(ID, addrs, peerstore.AddressTTL)
	}
	err := pm.host.Peerstore().(wps.WakuPeerstore).SetOrigin(ID, origin)
	if err != nil {
		pm.logger.Error("could not set origin", zap.Error(err), logging.HostID("peer", ID))
		return err
	}

	if len(protocols) > 0 {
		err = pm.host.Peerstore().AddProtocols(ID, protocols...)
		if err != nil {
			pm.logger.Error("could not set protocols", zap.Error(err), logging.HostID("peer", ID))
			return err
		}
	}
	if len(pubSubTopics) == 0 {
		// Probably the peer is discovered via DNSDiscovery (for which we don't have pubSubTopic info)
		//If pubSubTopic and enr is empty or no shard info in ENR,then set to defaultPubSubTopic
		pubSubTopics = []string{relay.DefaultWakuTopic}
	}
	err = pm.host.Peerstore().(wps.WakuPeerstore).SetPubSubTopics(ID, pubSubTopics)
	if err != nil {
		pm.logger.Error("could not store pubSubTopic", zap.Error(err),
			logging.HostID("peer", ID), zap.Strings("topics", pubSubTopics))
	}
	return nil
}

func AddrInfoToPeerData(origin wps.Origin, peerID peer.ID, host host.Host, pubsubTopics ...string) *service.PeerData {
	addrs := host.Peerstore().Addrs(peerID)
	if len(addrs) == 0 {
		//Addresses expired, remove peer from peerStore
		host.Peerstore().RemovePeer(peerID)
		return nil
	}
	return &service.PeerData{
		Origin: origin,
		AddrInfo: peer.AddrInfo{
			ID:    peerID,
			Addrs: addrs,
		},
		PubsubTopics: pubsubTopics,
	}
}

// AddPeer adds peer to the peerStore and also to service slots
func (pm *PeerManager) AddPeer(address ma.Multiaddr, origin wps.Origin, pubsubTopics []string, protocols ...protocol.ID) (*service.PeerData, error) {
	//Assuming all addresses have peerId
	info, err := peer.AddrInfoFromP2pAddr(address)
	if err != nil {
		return nil, err
	}

	//Add Service peers to serviceSlots.
	for _, proto := range protocols {
		pm.addPeerToServiceSlot(proto, info.ID)
	}

	//Add to the peer-store
	err = pm.addPeer(info.ID, info.Addrs, origin, pubsubTopics, protocols...)
	if err != nil {
		return nil, err
	}

	pData := &service.PeerData{
		Origin: origin,
		AddrInfo: peer.AddrInfo{
			ID:    info.ID,
			Addrs: info.Addrs,
		},
		PubsubTopics: pubsubTopics,
	}

	return pData, nil
}

// Connect establishes a connection to a
func (pm *PeerManager) Connect(pData *service.PeerData) {
	go pm.peerConnector.PushToChan(*pData)
}

// RemovePeer deletes peer from the peerStore after disconnecting it.
// It also removes the peer from serviceSlot.
func (pm *PeerManager) RemovePeer(peerID peer.ID) {
	pm.host.Peerstore().RemovePeer(peerID)
	//Search if this peer is in serviceSlot and if so, remove it from there
	// TODO:Add another peer which is statically configured to the serviceSlot.
	pm.serviceSlots.removePeer(peerID)
}

// addPeerToServiceSlot adds a peerID to serviceSlot.
// Adding to peerStore is expected to be already done by caller.
// If relay proto is passed, it is not added to serviceSlot.
func (pm *PeerManager) addPeerToServiceSlot(proto protocol.ID, peerID peer.ID) {
	if proto == relay.WakuRelayID_v200 {
		pm.logger.Debug("cannot add Relay peer to service peer slots")
		return
	}

	//For now adding the peer to serviceSlot which means the latest added peer would be given priority.
	//TODO: Ideally we should sort the peers per service and return best peer based on peer score or RTT etc.
	pm.logger.Info("adding peer to service slots", logging.HostID("peer", peerID),
		zap.String("service", string(proto)))
	// getPeers returns nil for WakuRelayIDv200 protocol, but we don't run this ServiceSlot code for WakuRelayIDv200 protocol
	pm.serviceSlots.getPeers(proto).add(peerID)
}
