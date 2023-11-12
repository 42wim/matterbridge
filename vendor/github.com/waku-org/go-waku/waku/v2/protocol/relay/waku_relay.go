package relay

import (
	"context"
	"errors"
	"sync"

	"github.com/libp2p/go-libp2p/core/event"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/libp2p/go-libp2p/p2p/host/eventbus"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
	proto "google.golang.org/protobuf/proto"

	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/waku-org/go-waku/logging"
	waku_proto "github.com/waku-org/go-waku/waku/v2/protocol"
	"github.com/waku-org/go-waku/waku/v2/protocol/pb"
	"github.com/waku-org/go-waku/waku/v2/service"
	"github.com/waku-org/go-waku/waku/v2/timesource"
	"github.com/waku-org/go-waku/waku/v2/utils"
)

// WakuRelayID_v200 is the current protocol ID used for WakuRelay
const WakuRelayID_v200 = protocol.ID("/vac/waku/relay/2.0.0")
const WakuRelayENRField = uint8(1 << 0)

const defaultMaxMsgSizeBytes = 150 * 1024

// DefaultWakuTopic is the default pubsub topic used across all Waku protocols
var DefaultWakuTopic string = waku_proto.DefaultPubsubTopic{}.String()

// WakuRelay is the implementation of the Waku Relay protocol
type WakuRelay struct {
	host                host.Host
	relayParams         *relayParameters
	pubsub              *pubsub.PubSub
	params              pubsub.GossipSubParams
	peerScoreParams     *pubsub.PeerScoreParams
	peerScoreThresholds *pubsub.PeerScoreThresholds
	topicParams         *pubsub.TopicScoreParams
	timesource          timesource.Timesource
	metrics             Metrics
	log                 *zap.Logger
	logMessages         *zap.Logger

	bcaster Broadcaster

	minPeersToPublish int

	topicValidatorMutex    sync.RWMutex
	topicValidators        map[string][]validatorFn
	defaultTopicValidators []validatorFn

	topicsMutex sync.RWMutex
	topics      map[string]*pubsubTopicSubscriptionDetails

	events   event.Bus
	emitters struct {
		EvtRelaySubscribed   event.Emitter
		EvtRelayUnsubscribed event.Emitter
		EvtPeerTopic         event.Emitter
	}

	*service.CommonService
}

type pubsubTopicSubscriptionDetails struct {
	topic             *pubsub.Topic
	subscription      *pubsub.Subscription
	topicEventHandler *pubsub.TopicEventHandler
	contentSubs       map[int]*Subscription
}

// NewWakuRelay returns a new instance of a WakuRelay struct
func NewWakuRelay(bcaster Broadcaster, minPeersToPublish int, timesource timesource.Timesource,
	reg prometheus.Registerer, log *zap.Logger, opts ...RelayOption) *WakuRelay {
	w := new(WakuRelay)
	w.timesource = timesource
	w.topics = make(map[string]*pubsubTopicSubscriptionDetails)
	w.topicValidators = make(map[string][]validatorFn)
	w.bcaster = bcaster
	w.minPeersToPublish = minPeersToPublish
	w.CommonService = service.NewCommonService()
	w.log = log.Named("relay")
	w.logMessages = utils.MessagesLogger("relay")
	w.events = eventbus.NewBus()
	w.metrics = newMetrics(reg, w.logMessages)
	w.relayParams = new(relayParameters)
	w.relayParams.pubsubOpts = w.defaultPubsubOptions()

	options := defaultOptions()
	options = append(options, opts...)
	for _, opt := range options {
		opt(w.relayParams)
	}
	w.log.Info("relay config", zap.Int("max-msg-size-bytes", w.relayParams.maxMsgSizeBytes),
		zap.Int("min-peers-to-publish", w.minPeersToPublish))
	return w
}

func (w *WakuRelay) peerScoreInspector(peerScoresSnapshots map[peer.ID]*pubsub.PeerScoreSnapshot) {
	if w.host == nil {
		return
	}

	for pid, snap := range peerScoresSnapshots {
		if snap.Score < w.peerScoreThresholds.GraylistThreshold {
			// Disconnect bad peers
			err := w.host.Network().ClosePeer(pid)
			if err != nil {
				w.log.Error("could not disconnect peer", logging.HostID("peer", pid), zap.Error(err))
			}
		}
	}
}

// SetHost sets the host to be able to mount or consume a protocol
func (w *WakuRelay) SetHost(h host.Host) {
	w.host = h
}

// Start initiates the WakuRelay protocol
func (w *WakuRelay) Start(ctx context.Context) error {
	return w.CommonService.Start(ctx, w.start)
}

func (w *WakuRelay) start() error {
	if w.bcaster == nil {
		return errors.New("broadcaster not specified for relay")
	}
	ps, err := pubsub.NewGossipSub(w.Context(), w.host, w.relayParams.pubsubOpts...)
	if err != nil {
		return err
	}
	w.pubsub = ps

	err = w.CreateEventEmitters()
	if err != nil {
		return err
	}

	w.log.Info("Relay protocol started")
	return nil
}

// PubSub returns the implementation of the pubsub system
func (w *WakuRelay) PubSub() *pubsub.PubSub {
	return w.pubsub
}

// Topics returns a list of all the pubsub topics currently subscribed to
func (w *WakuRelay) Topics() []string {
	defer w.topicsMutex.RUnlock()
	w.topicsMutex.RLock()

	var result []string
	for topic := range w.topics {
		result = append(result, topic)
	}
	return result
}

// IsSubscribed indicates whether the node is subscribed to a pubsub topic or not
func (w *WakuRelay) IsSubscribed(topic string) bool {
	w.topicsMutex.RLock()
	defer w.topicsMutex.RUnlock()
	_, ok := w.topics[topic]
	return ok
}

// SetPubSub is used to set an implementation of the pubsub system
func (w *WakuRelay) SetPubSub(pubSub *pubsub.PubSub) {
	w.pubsub = pubSub
}

func (w *WakuRelay) upsertTopic(topic string) (*pubsub.Topic, error) {
	topicData, ok := w.topics[topic]
	if !ok { // Joins topic if node hasn't joined yet
		err := w.pubsub.RegisterTopicValidator(topic, w.topicValidator(topic))
		if err != nil {
			w.log.Error("failed to register topic validator", zap.String("pubsubTopic", topic), zap.Error(err))
			return nil, err
		}

		newTopic, err := w.pubsub.Join(string(topic))
		if err != nil {
			w.log.Error("failed to join pubsubTopic", zap.String("pubsubTopic", topic), zap.Error(err))
			return nil, err
		}

		err = newTopic.SetScoreParams(w.topicParams)
		if err != nil {
			w.log.Error("failed to set score params", zap.String("pubsubTopic", topic), zap.Error(err))
			return nil, err
		}

		w.topics[topic] = &pubsubTopicSubscriptionDetails{
			topic: newTopic,
		}

		return newTopic, nil
	}

	return topicData.topic, nil
}

func (w *WakuRelay) subscribeToPubsubTopic(topic string) (*pubsubTopicSubscriptionDetails, error) {
	w.topicsMutex.Lock()
	defer w.topicsMutex.Unlock()
	w.log.Info("subscribing to underlying pubsubTopic", zap.String("pubsubTopic", topic))

	result, ok := w.topics[topic]
	if !ok {
		pubSubTopic, err := w.upsertTopic(topic)
		if err != nil {
			w.log.Error("failed to upsert topic", zap.String("pubsubTopic", topic), zap.Error(err))
			return nil, err
		}

		subscription, err := pubSubTopic.Subscribe(pubsub.WithBufferSize(1024))
		if err != nil {
			return nil, err
		}

		w.WaitGroup().Add(1)
		go w.pubsubTopicMsgHandler(subscription)

		evtHandler, err := w.addPeerTopicEventListener(pubSubTopic)
		if err != nil {
			return nil, err
		}

		w.topics[topic].contentSubs = make(map[int]*Subscription)
		w.topics[topic].subscription = subscription
		w.topics[topic].topicEventHandler = evtHandler

		err = w.emitters.EvtRelaySubscribed.Emit(EvtRelaySubscribed{topic, pubSubTopic})
		if err != nil {
			return nil, err
		}

		w.log.Info("gossipsub subscription", zap.String("pubsubTopic", subscription.Topic()))
		w.metrics.SetPubSubTopics(len(w.topics))
		result = w.topics[topic]
	}

	return result, nil
}

// Publish is used to broadcast a WakuMessage to a pubsub topic. The pubsubTopic is derived from contentTopic
// specified in the message via autosharding. To publish to a specific pubsubTopic, the `WithPubSubTopic` option should
// be provided
func (w *WakuRelay) Publish(ctx context.Context, message *pb.WakuMessage, opts ...PublishOption) ([]byte, error) {
	// Publish a `WakuMessage` to a PubSub topic.
	if w.pubsub == nil {
		return nil, errors.New("PubSub hasn't been set")
	}

	if message == nil {
		return nil, errors.New("message can't be null")
	}

	err := message.Validate()
	if err != nil {
		return nil, err
	}

	params := new(publishParameters)
	for _, opt := range opts {
		opt(params)
	}

	if params.pubsubTopic == "" {
		params.pubsubTopic, err = waku_proto.GetPubSubTopicFromContentTopic(message.ContentTopic)
		if err != nil {
			return nil, err
		}
	}

	if !w.EnoughPeersToPublishToTopic(params.pubsubTopic) {
		return nil, errors.New("not enough peers to publish")
	}

	w.topicsMutex.RLock()
	defer w.topicsMutex.RUnlock()

	pubSubTopic, err := w.upsertTopic(params.pubsubTopic)
	if err != nil {
		return nil, err
	}

	out, err := proto.Marshal(message)
	if err != nil {
		return nil, err
	}

	if len(out) > w.relayParams.maxMsgSizeBytes {
		return nil, errors.New("message size exceeds gossipsub max message size")
	}

	err = pubSubTopic.Publish(ctx, out)
	if err != nil {
		return nil, err
	}

	hash := message.Hash(params.pubsubTopic)

	w.logMessages.Debug("waku.relay published", zap.String("pubsubTopic", params.pubsubTopic), logging.HexBytes("hash", hash), zap.Int64("publishTime", w.timesource.Now().UnixNano()), zap.Int("payloadSizeBytes", len(message.Payload)))

	return hash, nil
}

func (w *WakuRelay) getSubscription(contentFilter waku_proto.ContentFilter) (*Subscription, error) {
	w.topicsMutex.RLock()
	defer w.topicsMutex.RUnlock()
	topicData, ok := w.topics[contentFilter.PubsubTopic]
	if ok {
		for _, sub := range topicData.contentSubs {
			if sub.contentFilter.Equals(contentFilter) {
				if sub.noConsume { //This check is to ensure that default no-consumer subscription is not returned
					continue
				}
				return sub, nil
			}
		}
	}

	return nil, errors.New("no subscription found for content topic")
}

// GetSubscriptionWithPubsubTopic fetches subscription matching pubsub and contentTopic
func (w *WakuRelay) GetSubscriptionWithPubsubTopic(pubsubTopic string, contentTopic string) (*Subscription, error) {
	var contentFilter waku_proto.ContentFilter
	if contentTopic != "" {
		contentFilter = waku_proto.NewContentFilter(pubsubTopic, contentTopic)
	} else {
		contentFilter = waku_proto.NewContentFilter(pubsubTopic)
	}
	sub, err := w.getSubscription(contentFilter)
	if err != nil {
		err = errors.New("no subscription found for pubsubTopic")
	}
	return sub, err
}

// GetSubscription fetches subscription matching a contentTopic(via autosharding)
func (w *WakuRelay) GetSubscription(contentTopic string) (*Subscription, error) {
	pubsubTopic, err := waku_proto.GetPubSubTopicFromContentTopic(contentTopic)
	if err != nil {
		w.log.Error("failed to derive pubsubTopic", zap.Error(err), zap.String("contentTopic", contentTopic))
		return nil, err
	}
	contentFilter := waku_proto.NewContentFilter(pubsubTopic, contentTopic)

	return w.getSubscription(contentFilter)
}

// Stop unmounts the relay protocol and stops all subscriptions
func (w *WakuRelay) Stop() {
	w.CommonService.Stop(func() {
		w.host.RemoveStreamHandler(WakuRelayID_v200)
		w.emitters.EvtRelaySubscribed.Close()
		w.emitters.EvtRelayUnsubscribed.Close()
	})
}

// EnoughPeersToPublish returns whether there are enough peers connected in the default waku pubsub topic
func (w *WakuRelay) EnoughPeersToPublish() bool {
	return w.EnoughPeersToPublishToTopic(DefaultWakuTopic)
}

// EnoughPeersToPublish returns whether there are enough peers connected in a pubsub topic
func (w *WakuRelay) EnoughPeersToPublishToTopic(topic string) bool {
	return len(w.PubSub().ListPeers(topic)) >= w.minPeersToPublish
}

// subscribe returns list of Subscription to receive messages based on content filter
func (w *WakuRelay) subscribe(ctx context.Context, contentFilter waku_proto.ContentFilter, opts ...RelaySubscribeOption) ([]*Subscription, error) {

	var subscriptions []*Subscription
	pubSubTopicMap, err := waku_proto.ContentFilterToPubSubTopicMap(contentFilter)
	if err != nil {
		return nil, err
	}
	params := new(RelaySubscribeParameters)

	var optList []RelaySubscribeOption
	optList = append(optList, opts...)
	for _, opt := range optList {
		err := opt(params)
		if err != nil {
			w.log.Error("failed to apply option", zap.Error(err))
			return nil, err
		}
	}
	if params.cacheSize <= 0 {
		params.cacheSize = uint(DefaultRelaySubscriptionBufferSize)
	}

	for pubSubTopic, cTopics := range pubSubTopicMap {
		w.log.Info("subscribing to", zap.String("pubsubTopic", pubSubTopic), zap.Strings("contentTopics", cTopics))
		var cFilter waku_proto.ContentFilter
		cFilter.PubsubTopic = pubSubTopic
		cFilter.ContentTopics = waku_proto.NewContentTopicSet(cTopics...)

		//Check if gossipsub subscription already exists for pubSubTopic
		if !w.IsSubscribed(pubSubTopic) {
			_, err := w.subscribeToPubsubTopic(cFilter.PubsubTopic)
			if err != nil {
				//TODO: Handle partial errors.
				w.log.Error("failed to subscribe to pubsubTopic", zap.Error(err), zap.String("pubsubTopic", cFilter.PubsubTopic))
				return nil, err
			}
		}

		subscription := w.bcaster.Register(cFilter, WithBufferSize(int(params.cacheSize)),
			WithConsumerOption(params.dontConsume))

		// Create Content subscription
		w.topicsMutex.Lock()
		topicData, ok := w.topics[pubSubTopic]
		if ok {
			topicData.contentSubs[subscription.ID] = subscription
		}
		w.topicsMutex.Unlock()

		subscriptions = append(subscriptions, subscription)
		go func() {
			<-ctx.Done()
			subscription.Unsubscribe()
		}()
	}

	return subscriptions, nil
}

// Subscribe returns a Subscription to receive messages as per contentFilter
// contentFilter can contain pubSubTopic and contentTopics or only contentTopics(in case of autosharding)
func (w *WakuRelay) Subscribe(ctx context.Context, contentFilter waku_proto.ContentFilter, opts ...RelaySubscribeOption) ([]*Subscription, error) {
	return w.subscribe(ctx, contentFilter, opts...)
}

// Unsubscribe closes a subscription to a pubsub topic
func (w *WakuRelay) Unsubscribe(ctx context.Context, contentFilter waku_proto.ContentFilter) error {

	pubSubTopicMap, err := waku_proto.ContentFilterToPubSubTopicMap(contentFilter)
	if err != nil {
		w.log.Error("failed to derive pubsubTopic from contentFilter", zap.String("pubsubTopic", contentFilter.PubsubTopic),
			zap.Strings("contentTopics", contentFilter.ContentTopicsList()))
		return err
	}

	w.topicsMutex.Lock()
	defer w.topicsMutex.Unlock()

	for pubSubTopic, cTopics := range pubSubTopicMap {
		cfTemp := waku_proto.NewContentFilter(pubSubTopic, cTopics...)
		pubsubUnsubscribe := false
		sub, ok := w.topics[pubSubTopic]
		if !ok {
			w.log.Error("not subscribed to topic", zap.String("topic", pubSubTopic))
			return errors.New("not subscribed to topic")
		}

		topicData, ok := w.topics[pubSubTopic]
		if ok {
			//Remove relevant subscription
			for subID, sub := range topicData.contentSubs {
				if sub.contentFilter.Equals(cfTemp) {
					sub.Unsubscribe()
					delete(topicData.contentSubs, subID)
				}
			}

			if len(topicData.contentSubs) == 0 {
				pubsubUnsubscribe = true
			}
		} else {
			//Should not land here ideally
			w.log.Error("pubsub subscriptions exists, but contentSubscription doesn't for contentFilter",
				zap.String("pubsubTopic", pubSubTopic), zap.Strings("contentTopics", cTopics))

			return errors.New("unexpected error in unsubscribe")
		}

		if pubsubUnsubscribe {
			err = w.unsubscribeFromPubsubTopic(sub)
			if err != nil {
				return err
			}
			w.metrics.SetPubSubTopics(len(w.topics))
		}
	}
	return nil
}

// unsubscribeFromPubsubTopic unsubscribes subscription from underlying pubsub.
// Note: caller has to acquire topicsMutex in order to avoid race conditions
func (w *WakuRelay) unsubscribeFromPubsubTopic(topicData *pubsubTopicSubscriptionDetails) error {

	pubSubTopic := topicData.subscription.Topic()
	w.log.Info("unsubscribing from pubsubTopic", zap.String("topic", pubSubTopic))

	topicData.subscription.Cancel()
	topicData.topicEventHandler.Cancel()

	w.bcaster.UnRegister(pubSubTopic)

	err := topicData.topic.Close()
	if err != nil {
		w.log.Error("failed to close the pubsubTopic", zap.String("topic", pubSubTopic))
		return err
	}

	w.RemoveTopicValidator(pubSubTopic)

	err = w.pubsub.UnregisterTopicValidator(pubSubTopic)
	if err != nil {
		w.log.Error("failed to unregister topic validator", zap.String("topic", pubSubTopic))
		return err
	}

	delete(w.topics, pubSubTopic)

	return w.emitters.EvtRelayUnsubscribed.Emit(EvtRelayUnsubscribed{pubSubTopic})
}

func (w *WakuRelay) pubsubTopicMsgHandler(sub *pubsub.Subscription) {
	defer w.WaitGroup().Done()

	for {
		msg, err := sub.Next(w.Context())
		if err != nil {
			if !errors.Is(err, context.Canceled) {
				w.log.Error("getting message from subscription", zap.Error(err))
			}
			sub.Cancel()
			return
		}

		wakuMessage, err := pb.Unmarshal(msg.Data)
		if err != nil {
			w.log.Error("decoding message", zap.Error(err))
			return
		}

		envelope := waku_proto.NewEnvelope(wakuMessage, w.timesource.Now().UnixNano(), sub.Topic())
		w.metrics.RecordMessage(envelope)

		w.bcaster.Submit(envelope)
	}

}

// Params returns the gossipsub configuration parameters used by WakuRelay
func (w *WakuRelay) Params() pubsub.GossipSubParams {
	return w.params
}
