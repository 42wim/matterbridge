package peermanager

import (
	"context"

	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/event"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/waku-org/go-waku/logging"
	wps "github.com/waku-org/go-waku/waku/v2/peerstore"
	waku_proto "github.com/waku-org/go-waku/waku/v2/protocol"
	"github.com/waku-org/go-waku/waku/v2/protocol/relay"
	"go.uber.org/zap"
)

func (pm *PeerManager) SubscribeToRelayEvtBus(bus event.Bus) error {
	var err error
	pm.sub, err = bus.Subscribe([]interface{}{new(relay.EvtPeerTopic), new(relay.EvtRelaySubscribed), new(relay.EvtRelayUnsubscribed)})
	return err
}

func (pm *PeerManager) handleNewRelayTopicSubscription(pubsubTopic string, topicInst *pubsub.Topic) {
	pm.logger.Info("handleNewRelayTopicSubscription", zap.String("pubSubTopic", pubsubTopic))
	pm.topicMutex.Lock()
	defer pm.topicMutex.Unlock()

	_, ok := pm.subRelayTopics[pubsubTopic]
	if ok {
		//Nothing to be done, as we are already subscribed to this topic.
		return
	}
	pm.subRelayTopics[pubsubTopic] = &NodeTopicDetails{topicInst}
	//Check how many relay peers we are connected to that subscribe to this topic, if less than D find peers in peerstore and connect.
	//If no peers in peerStore, trigger discovery for this topic?
	relevantPeersForPubSubTopic := pm.host.Peerstore().(*wps.WakuPeerstoreImpl).PeersByPubSubTopic(pubsubTopic)
	var notConnectedPeers peer.IDSlice
	connectedPeers := 0
	for _, peer := range relevantPeersForPubSubTopic {
		if pm.host.Network().Connectedness(peer) == network.Connected {
			connectedPeers++
		} else {
			notConnectedPeers = append(notConnectedPeers, peer)
		}
	}

	if connectedPeers >= waku_proto.GossipSubOptimalFullMeshSize { //TODO: Use a config rather than hard-coding.
		// Should we use optimal number or define some sort of a config for the node to choose from?
		// A desktop node may choose this to be 4-6, whereas a service node may choose this to be 8-12 based on resources it has
		// or bandwidth it can support.
		// Should we link this to bandwidth management somehow or just depend on some sort of config profile?
		pm.logger.Info("Optimal required relay peers for new pubSubTopic are already connected ", zap.String("pubSubTopic", pubsubTopic),
			zap.Int("connectedPeerCount", connectedPeers))
		return
	}
	triggerDiscovery := false
	if notConnectedPeers.Len() > 0 {
		numPeersToConnect := notConnectedPeers.Len() - connectedPeers
		if numPeersToConnect < 0 {
			numPeersToConnect = notConnectedPeers.Len()
		} else if numPeersToConnect-connectedPeers > waku_proto.GossipSubOptimalFullMeshSize {
			numPeersToConnect = waku_proto.GossipSubOptimalFullMeshSize - connectedPeers
		}
		if numPeersToConnect+connectedPeers < waku_proto.GossipSubOptimalFullMeshSize {
			triggerDiscovery = true
		}
		//For now all peers are being given same priority,
		// Later we may want to choose peers that have more shards in common over others.
		pm.connectToPeers(notConnectedPeers[0:numPeersToConnect])
	} else {
		triggerDiscovery = true
	}

	if triggerDiscovery {
		//TODO: Initiate on-demand discovery for this pubSubTopic.
		// Use peer-exchange and rendevouz?
		//Should we query discoverycache to find out if there are any more peers before triggering discovery?
		return
	}
}

func (pm *PeerManager) handleNewRelayTopicUnSubscription(pubsubTopic string) {
	pm.logger.Info("handleNewRelayTopicUnSubscription", zap.String("pubSubTopic", pubsubTopic))
	pm.topicMutex.Lock()
	defer pm.topicMutex.Unlock()
	_, ok := pm.subRelayTopics[pubsubTopic]
	if !ok {
		//Nothing to be done, as we are already unsubscribed from this topic.
		return
	}
	delete(pm.subRelayTopics, pubsubTopic)

	//If there are peers only subscribed to this topic, disconnect them.
	relevantPeersForPubSubTopic := pm.host.Peerstore().(*wps.WakuPeerstoreImpl).PeersByPubSubTopic(pubsubTopic)
	for _, peer := range relevantPeersForPubSubTopic {
		if pm.host.Network().Connectedness(peer) == network.Connected {
			peerTopics, err := pm.host.Peerstore().(*wps.WakuPeerstoreImpl).PubSubTopics(peer)
			if err != nil {
				pm.logger.Error("Could not retrieve pubsub topics for peer", zap.Error(err),
					logging.HostID("peerID", peer))
				continue
			}
			if len(peerTopics) == 1 && peerTopics[0] == pubsubTopic {
				err := pm.host.Network().ClosePeer(peer)
				if err != nil {
					pm.logger.Warn("Failed to disconnect connection towards peer",
						logging.HostID("peerID", peer))
					continue
				}
				pm.logger.Debug("Successfully disconnected connection towards peer",
					logging.HostID("peerID", peer))
			}
		}
	}
}

func (pm *PeerManager) handlerPeerTopicEvent(peerEvt relay.EvtPeerTopic) {
	wps := pm.host.Peerstore().(*wps.WakuPeerstoreImpl)
	peerID := peerEvt.PeerID
	if peerEvt.State == relay.PEER_JOINED {
		err := wps.AddPubSubTopic(peerID, peerEvt.PubsubTopic)
		if err != nil {
			pm.logger.Error("failed to add pubSubTopic for peer",
				logging.HostID("peerID", peerID), zap.String("topic", peerEvt.PubsubTopic), zap.Error(err))
		}
	} else if peerEvt.State == relay.PEER_LEFT {
		err := wps.RemovePubSubTopic(peerID, peerEvt.PubsubTopic)
		if err != nil {
			pm.logger.Error("failed to remove pubSubTopic for peer",
				logging.HostID("peerID", peerID), zap.Error(err))
		}
	} else {
		pm.logger.Error("unknown peer event received", zap.Int("eventState", int(peerEvt.State)))
	}
}

func (pm *PeerManager) peerEventLoop(ctx context.Context) {
	defer pm.sub.Close()
	for {
		select {
		case e := <-pm.sub.Out():
			switch e := e.(type) {
			case relay.EvtPeerTopic:
				{
					peerEvt := (relay.EvtPeerTopic)(e)
					pm.handlerPeerTopicEvent(peerEvt)
				}
			case relay.EvtRelaySubscribed:
				{
					eventDetails := (relay.EvtRelaySubscribed)(e)
					pm.handleNewRelayTopicSubscription(eventDetails.Topic, eventDetails.TopicInst)
				}
			case relay.EvtRelayUnsubscribed:
				{
					eventDetails := (relay.EvtRelayUnsubscribed)(e)
					pm.handleNewRelayTopicUnSubscription(eventDetails.Topic)
				}
			default:
				pm.logger.Error("unsupported event type", zap.Any("eventType", e))
			}

		case <-ctx.Done():
			return
		}
	}
}
