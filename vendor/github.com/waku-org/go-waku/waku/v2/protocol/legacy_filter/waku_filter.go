package legacy_filter

import (
	"context"
	"encoding/hex"
	"errors"
	"math"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	libp2pProtocol "github.com/libp2p/go-libp2p/core/protocol"
	"github.com/libp2p/go-msgio/pbio"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/waku-org/go-waku/logging"
	"github.com/waku-org/go-waku/waku/v2/peermanager"
	"github.com/waku-org/go-waku/waku/v2/protocol"
	"github.com/waku-org/go-waku/waku/v2/protocol/legacy_filter/pb"
	wpb "github.com/waku-org/go-waku/waku/v2/protocol/pb"
	"github.com/waku-org/go-waku/waku/v2/protocol/relay"
	"github.com/waku-org/go-waku/waku/v2/service"
	"github.com/waku-org/go-waku/waku/v2/timesource"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

var (
	ErrNoPeersAvailable = errors.New("no suitable remote peers")
)

type (
	Filter struct {
		filterID       string
		PeerID         peer.ID
		Topic          string
		ContentFilters []string
		Chan           chan *protocol.Envelope
	}

	ContentFilter struct {
		Topic         string
		ContentTopics []string
	}

	FilterSubscription struct {
		RequestID string
		Peer      peer.ID
	}

	WakuFilter struct {
		*service.CommonService
		h          host.Host
		pm         *peermanager.PeerManager
		isFullNode bool
		msgSub     *relay.Subscription
		metrics    Metrics
		log        *zap.Logger

		filters     *FilterMap
		subscribers *Subscribers
	}
)

// FilterID_v20beta1 is the current Waku Filter protocol identifier
const FilterID_v20beta1 = libp2pProtocol.ID("/vac/waku/filter/2.0.0-beta1")

// NewWakuRelay returns a new instance of Waku Filter struct setup according to the chosen parameter and options
func NewWakuFilter(broadcaster relay.Broadcaster, isFullNode bool, timesource timesource.Timesource, reg prometheus.Registerer, log *zap.Logger, opts ...Option) *WakuFilter {
	wf := new(WakuFilter)
	wf.log = log.Named("filter").With(zap.Bool("fullNode", isFullNode))

	params := new(FilterParameters)
	optList := DefaultOptions()
	optList = append(optList, opts...)
	for _, opt := range optList {
		opt(params)
	}

	wf.isFullNode = isFullNode
	wf.CommonService = service.NewCommonService()
	wf.filters = NewFilterMap(broadcaster, timesource)
	wf.subscribers = NewSubscribers(params.Timeout)
	wf.metrics = newMetrics(reg)

	return wf
}

// Sets the host to be able to mount or consume a protocol
func (wf *WakuFilter) SetHost(h host.Host) {
	wf.h = h
}

func (wf *WakuFilter) Start(ctx context.Context, sub *relay.Subscription) error {
	return wf.CommonService.Start(ctx, func() error {
		return wf.start(sub)
	})
}

func (wf *WakuFilter) start(sub *relay.Subscription) error {
	wf.h.SetStreamHandlerMatch(FilterID_v20beta1, protocol.PrefixTextMatch(string(FilterID_v20beta1)), wf.onRequest(wf.Context()))
	wf.msgSub = sub
	wf.WaitGroup().Add(1)
	go wf.filterListener(wf.Context())
	wf.log.Info("filter protocol started")
	return nil
}
func (wf *WakuFilter) onRequest(ctx context.Context) func(network.Stream) {
	return func(stream network.Stream) {
		peerID := stream.Conn().RemotePeer()
		logger := wf.log.With(logging.HostID("peer", peerID))

		filterRPCRequest := &pb.FilterRpc{}

		reader := pbio.NewDelimitedReader(stream, math.MaxInt32)

		err := reader.ReadMsg(filterRPCRequest)
		if err != nil {
			wf.metrics.RecordError(decodeRPCFailure)
			logger.Error("reading request", zap.Error(err))
			if err := stream.Reset(); err != nil {
				wf.log.Error("resetting connection", zap.Error(err))
			}
			return
		}

		logger.Info("received request")

		if filterRPCRequest.Push != nil && len(filterRPCRequest.Push.Messages) > 0 {
			// We're on a light node.
			// This is a message push coming from a full node.
			for _, message := range filterRPCRequest.Push.Messages {
				wf.filters.Notify(message, filterRPCRequest.RequestId) // Trigger filter handlers on a light node
			}

			logger.Info("received a message push", zap.Int("messages", len(filterRPCRequest.Push.Messages)))
			wf.metrics.RecordMessages(len(filterRPCRequest.Push.Messages))
		} else if filterRPCRequest.Request != nil && wf.isFullNode {
			// We're on a full node.
			// This is a filter request coming from a light node.
			if filterRPCRequest.Request.Subscribe {
				subscriber := Subscriber{peer: stream.Conn().RemotePeer(), requestID: filterRPCRequest.RequestId, filter: filterRPCRequest.Request}
				if subscriber.filter.Topic == "" { // @TODO: review if empty topic is possible
					subscriber.filter.Topic = relay.DefaultWakuTopic
				}

				subscribersLen := wf.subscribers.Append(subscriber)

				logger.Info("adding subscriber")
				wf.metrics.RecordSubscribers(subscribersLen)
			} else {
				wf.subscribers.RemoveContentFilters(peerID, filterRPCRequest.RequestId, filterRPCRequest.Request.ContentFilters)

				logger.Info("removing subscriber")
				wf.metrics.RecordSubscribers(wf.subscribers.Length())
			}
		} else {
			logger.Error("can't serve request")
			if err := stream.Reset(); err != nil {
				wf.log.Error("resetting connection", zap.Error(err))
			}
			return
		}

		stream.Close()
	}
}

func (wf *WakuFilter) pushMessage(ctx context.Context, subscriber Subscriber, msg *wpb.WakuMessage) error {
	pushRPC := &pb.FilterRpc{RequestId: subscriber.requestID, Push: &pb.MessagePush{Messages: []*wpb.WakuMessage{msg}}}
	logger := wf.log.With(logging.HostID("peer", subscriber.peer))

	stream, err := wf.h.NewStream(ctx, subscriber.peer, FilterID_v20beta1)
	if err != nil {
		wf.subscribers.FlagAsFailure(subscriber.peer)
		logger.Error("opening peer stream", zap.Error(err))
		wf.metrics.RecordError(dialFailure)
		return err
	}

	writer := pbio.NewDelimitedWriter(stream)
	err = writer.WriteMsg(pushRPC)
	if err != nil {
		logger.Error("pushing messages to peer", zap.Error(err))
		wf.subscribers.FlagAsFailure(subscriber.peer)
		wf.metrics.RecordError(pushWriteError)
		if err := stream.Reset(); err != nil {
			wf.log.Error("resetting connection", zap.Error(err))
		}
		return nil
	}

	stream.Close()

	wf.subscribers.FlagAsSuccess(subscriber.peer)
	return nil
}

func (wf *WakuFilter) filterListener(ctx context.Context) {
	defer wf.WaitGroup().Done()

	// This function is invoked for each message received
	// on the full node in context of Waku2-Filter
	handle := func(envelope *protocol.Envelope) error { // async
		msg := envelope.Message()
		pubsubTopic := envelope.PubsubTopic()
		logger := wf.log.With(zap.Stringer("message", msg))
		g := new(errgroup.Group)
		// Each subscriber is a light node that earlier on invoked
		// a FilterRequest on this node
		for subscriber := range wf.subscribers.Items(&(msg.ContentTopic)) {
			logger := logger.With(logging.HostID("subscriber", subscriber.peer))
			subscriber := subscriber // https://golang.org/doc/faq#closures_and_goroutines
			if subscriber.filter.Topic != pubsubTopic {
				logger.Info("pubsub topic mismatch",
					zap.String("subscriberTopic", subscriber.filter.Topic),
					zap.String("messageTopic", pubsubTopic))
				continue
			}

			// Do a message push to light node
			logger.Info("pushing message to light node", zap.String("contentTopic", msg.ContentTopic))
			g.Go(func() (err error) {
				err = wf.pushMessage(ctx, subscriber, msg)
				if err != nil {
					logger.Error("pushing message", zap.Error(err))
				}
				return err
			})
		}

		return g.Wait()
	}

	for m := range wf.msgSub.Ch {
		if err := handle(m); err != nil {
			wf.log.Error("handling message", zap.Error(err))
		}
	}
}

// Having a FilterRequest struct,
// select a peer with filter support, dial it,
// and submit FilterRequest wrapped in FilterRPC
func (wf *WakuFilter) requestSubscription(ctx context.Context, filter ContentFilter, opts ...FilterSubscribeOption) (subscription *FilterSubscription, err error) {
	params := new(FilterSubscribeParameters)
	params.log = wf.log
	params.host = wf.h

	optList := DefaultSubscribtionOptions()
	optList = append(optList, opts...)
	for _, opt := range optList {
		opt(params)
	}
	if wf.pm != nil && params.selectedPeer == "" {
		params.selectedPeer, _ = wf.pm.SelectPeer(
			peermanager.PeerSelectionCriteria{
				SelectionType: params.peerSelectionType,
				Proto:         FilterID_v20beta1,
				PubsubTopics:  []string{filter.Topic},
				SpecificPeers: params.preferredPeers,
				Ctx:           ctx,
			},
		)
	}
	if params.selectedPeer == "" {
		wf.metrics.RecordError(peerNotFoundFailure)
		return nil, ErrNoPeersAvailable
	}

	var contentFilters []*pb.FilterRequest_ContentFilter
	for _, ct := range filter.ContentTopics {
		contentFilters = append(contentFilters, &pb.FilterRequest_ContentFilter{ContentTopic: ct})
	}

	request := &pb.FilterRequest{
		Subscribe:      true,
		Topic:          filter.Topic,
		ContentFilters: contentFilters,
	}

	stream, err := wf.h.NewStream(ctx, params.selectedPeer, FilterID_v20beta1)
	if err != nil {
		wf.metrics.RecordError(dialFailure)
		return
	}

	// This is the only successful path to subscription
	requestID := hex.EncodeToString(protocol.GenerateRequestID())

	writer := pbio.NewDelimitedWriter(stream)
	filterRPC := &pb.FilterRpc{RequestId: requestID, Request: request}
	wf.log.Debug("sending filterRPC", zap.Stringer("rpc", filterRPC))
	err = writer.WriteMsg(filterRPC)
	if err != nil {
		wf.metrics.RecordError(writeRequestFailure)
		wf.log.Error("sending filterRPC", zap.Error(err))
		if err := stream.Reset(); err != nil {
			wf.log.Error("resetting connection", zap.Error(err))
		}
		return
	}

	stream.Close()

	subscription = new(FilterSubscription)
	subscription.Peer = params.selectedPeer
	subscription.RequestID = requestID

	return
}

// Unsubscribe is used to stop receiving messages from a peer that match a content filter
func (wf *WakuFilter) Unsubscribe(ctx context.Context, contentFilter ContentFilter, peer peer.ID) error {
	stream, err := wf.h.NewStream(ctx, peer, FilterID_v20beta1)
	if err != nil {
		wf.metrics.RecordError(dialFailure)
		return err
	}

	// This is the only successful path to subscription
	id := protocol.GenerateRequestID()

	var contentFilters []*pb.FilterRequest_ContentFilter
	for _, ct := range contentFilter.ContentTopics {
		contentFilters = append(contentFilters, &pb.FilterRequest_ContentFilter{ContentTopic: ct})
	}

	request := &pb.FilterRequest{
		Subscribe:      false,
		Topic:          contentFilter.Topic,
		ContentFilters: contentFilters,
	}

	writer := pbio.NewDelimitedWriter(stream)
	filterRPC := &pb.FilterRpc{RequestId: hex.EncodeToString(id), Request: request}
	err = writer.WriteMsg(filterRPC)
	if err != nil {
		wf.metrics.RecordError(writeRequestFailure)
		if err := stream.Reset(); err != nil {
			wf.log.Error("resetting connection", zap.Error(err))
		}
		return err
	}

	stream.Close()

	return nil
}

// Stop unmounts the filter protocol
func (wf *WakuFilter) Stop() {
	wf.CommonService.Stop(func() {
		wf.msgSub.Unsubscribe()

		wf.h.RemoveStreamHandler(FilterID_v20beta1)
		wf.filters.RemoveAll()
		wf.subscribers.Clear()
	})
}

// Subscribe setups a subscription to receive messages that match a specific content filter
func (wf *WakuFilter) Subscribe(ctx context.Context, f ContentFilter, opts ...FilterSubscribeOption) (filterID string, theFilter Filter, err error) {
	// TODO: check if there's an existing pubsub topic that uses the same peer. If so, reuse filter, and return same channel and filterID

	// Registers for messages that match a specific filter. Triggers the handler whenever a message is received.
	// ContentFilterChan takes MessagePush structs
	remoteSubs, err := wf.requestSubscription(ctx, f, opts...)
	if err != nil || remoteSubs.RequestID == "" {
		// Failed to subscribe
		wf.log.Error("requesting subscription", zap.Error(err))
		return
	}

	// Register handler for filter, whether remote subscription succeeded or not

	filterID = remoteSubs.RequestID
	theFilter = Filter{
		filterID:       filterID,
		PeerID:         remoteSubs.Peer,
		Topic:          f.Topic,
		ContentFilters: f.ContentTopics,
		Chan:           make(chan *protocol.Envelope, 1024), // To avoid blocking
	}
	wf.filters.Set(filterID, theFilter)

	return
}

// UnsubscribeFilterByID removes a subscription to a filter node completely
// using using a filter. It also closes the filter channel
func (wf *WakuFilter) UnsubscribeByFilter(ctx context.Context, filter Filter) error {
	err := wf.UnsubscribeFilterByID(ctx, filter.filterID)
	if err != nil {
		close(filter.Chan)
	}
	return err
}

// UnsubscribeFilterByID removes a subscription to a filter node completely
// using the filterID returned when the subscription was created
func (wf *WakuFilter) UnsubscribeFilterByID(ctx context.Context, filterID string) error {

	var f Filter
	var ok bool

	if f, ok = wf.filters.Get(filterID); !ok {
		return errors.New("filter not found")
	}

	cf := ContentFilter{
		Topic:         f.Topic,
		ContentTopics: f.ContentFilters,
	}

	err := wf.Unsubscribe(ctx, cf, f.PeerID)
	if err != nil {
		return err
	}

	wf.filters.Delete(filterID)

	return nil
}

// Unsubscribe filter removes content topics from a filter subscription. If all
// the contentTopics are removed the subscription is dropped completely
func (wf *WakuFilter) UnsubscribeFilter(ctx context.Context, cf ContentFilter) error {
	// Remove local filter
	idsToRemove := make(map[string]struct{})
	for filterMapItem := range wf.filters.Items() {
		f := filterMapItem.Value
		id := filterMapItem.Key

		if f.Topic != cf.Topic {
			continue
		}

		// Send message to full node in order to unsubscribe
		err := wf.Unsubscribe(ctx, cf, f.PeerID)
		if err != nil {
			return err
		}

		// Iterate filter entries to remove matching content topics
		// make sure we delete the content filter
		// if no more topics are left
		for _, cfToDelete := range cf.ContentTopics {
			for i, cf := range f.ContentFilters {
				if cf == cfToDelete {
					l := len(f.ContentFilters) - 1
					f.ContentFilters[l], f.ContentFilters[i] = f.ContentFilters[i], f.ContentFilters[l]
					f.ContentFilters = f.ContentFilters[:l]
					break
				}

			}
			if len(f.ContentFilters) == 0 {
				idsToRemove[id] = struct{}{}
			}
		}
	}

	for rID := range idsToRemove {
		wf.filters.Delete(rID)
	}

	return nil
}
