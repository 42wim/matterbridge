package filter

import (
	"context"
	"errors"
	"math"
	"net/http"
	"time"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	libp2pProtocol "github.com/libp2p/go-libp2p/core/protocol"
	"github.com/libp2p/go-msgio/pbio"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/waku-org/go-waku/logging"
	"github.com/waku-org/go-waku/waku/v2/protocol"
	"github.com/waku-org/go-waku/waku/v2/protocol/filter/pb"
	"github.com/waku-org/go-waku/waku/v2/protocol/relay"
	"github.com/waku-org/go-waku/waku/v2/service"
	"github.com/waku-org/go-waku/waku/v2/timesource"
	"github.com/waku-org/go-waku/waku/v2/utils"
	"go.uber.org/zap"
)

// FilterSubscribeID_v20beta1 is the current Waku Filter protocol identifier for servers to
// allow filter clients to subscribe, modify, refresh and unsubscribe a desired set of filter criteria
const FilterSubscribeID_v20beta1 = libp2pProtocol.ID("/vac/waku/filter-subscribe/2.0.0-beta1")
const FilterSubscribeENRField = uint8(1 << 2)
const peerHasNoSubscription = "peer has no subscriptions"

type (
	WakuFilterFullNode struct {
		h       host.Host
		msgSub  *relay.Subscription
		metrics Metrics
		log     *zap.Logger
		*service.CommonService
		subscriptions *SubscribersMap

		maxSubscriptions int
	}
)

// NewWakuFilterFullNode returns a new instance of Waku Filter struct setup according to the chosen parameter and options
func NewWakuFilterFullNode(timesource timesource.Timesource, reg prometheus.Registerer, log *zap.Logger, opts ...Option) *WakuFilterFullNode {
	wf := new(WakuFilterFullNode)
	wf.log = log.Named("filterv2-fullnode")

	params := new(FilterParameters)
	optList := DefaultOptions()
	optList = append(optList, opts...)
	for _, opt := range optList {
		opt(params)
	}

	wf.CommonService = service.NewCommonService()
	wf.metrics = newMetrics(reg)
	wf.subscriptions = NewSubscribersMap(params.Timeout)
	wf.maxSubscriptions = params.MaxSubscribers
	if params.pm != nil {
		params.pm.RegisterWakuProtocol(FilterSubscribeID_v20beta1, FilterSubscribeENRField)
	}
	return wf
}

// Sets the host to be able to mount or consume a protocol
func (wf *WakuFilterFullNode) SetHost(h host.Host) {
	wf.h = h
}

func (wf *WakuFilterFullNode) Start(ctx context.Context, sub *relay.Subscription) error {
	return wf.CommonService.Start(ctx, func() error {
		return wf.start(sub)
	})
}

func (wf *WakuFilterFullNode) start(sub *relay.Subscription) error {
	wf.h.SetStreamHandlerMatch(FilterSubscribeID_v20beta1, protocol.PrefixTextMatch(string(FilterSubscribeID_v20beta1)), wf.onRequest(wf.Context()))

	wf.msgSub = sub
	wf.WaitGroup().Add(1)
	go wf.filterListener(wf.Context())

	wf.log.Info("filter-subscriber protocol started")
	return nil
}

func (wf *WakuFilterFullNode) onRequest(ctx context.Context) func(network.Stream) {
	return func(stream network.Stream) {
		logger := wf.log.With(logging.HostID("peer", stream.Conn().RemotePeer()))

		reader := pbio.NewDelimitedReader(stream, math.MaxInt32)

		subscribeRequest := &pb.FilterSubscribeRequest{}
		err := reader.ReadMsg(subscribeRequest)
		if err != nil {
			wf.metrics.RecordError(decodeRPCFailure)
			logger.Error("reading request", zap.Error(err))
			if err := stream.Reset(); err != nil {
				wf.log.Error("resetting connection", zap.Error(err))
			}
			return
		}

		logger = logger.With(zap.String("requestID", subscribeRequest.RequestId))

		start := time.Now()

		if err := subscribeRequest.Validate(); err != nil {
			wf.reply(ctx, stream, subscribeRequest, http.StatusBadRequest, err.Error())
		} else {
			switch subscribeRequest.FilterSubscribeType {
			case pb.FilterSubscribeRequest_SUBSCRIBE:
				wf.subscribe(ctx, stream, subscribeRequest)
			case pb.FilterSubscribeRequest_SUBSCRIBER_PING:
				wf.ping(ctx, stream, subscribeRequest)
			case pb.FilterSubscribeRequest_UNSUBSCRIBE:
				wf.unsubscribe(ctx, stream, subscribeRequest)
			case pb.FilterSubscribeRequest_UNSUBSCRIBE_ALL:
				wf.unsubscribeAll(ctx, stream, subscribeRequest)
			}
		}

		stream.Close()

		wf.metrics.RecordRequest(subscribeRequest.FilterSubscribeType.String(), time.Since(start))

		logger.Info("received request", zap.String("requestType", subscribeRequest.FilterSubscribeType.String()))
	}
}

func (wf *WakuFilterFullNode) reply(ctx context.Context, stream network.Stream, request *pb.FilterSubscribeRequest, statusCode int, description ...string) {
	response := &pb.FilterSubscribeResponse{
		RequestId:  request.RequestId,
		StatusCode: uint32(statusCode),
	}

	if len(description) != 0 {
		response.StatusDesc = &description[0]
	} else {
		desc := http.StatusText(statusCode)
		response.StatusDesc = &desc
	}

	writer := pbio.NewDelimitedWriter(stream)
	err := writer.WriteMsg(response)
	if err != nil {
		wf.metrics.RecordError(writeResponseFailure)
		wf.log.Error("sending response", zap.Error(err))
		if err := stream.Reset(); err != nil {
			wf.log.Error("resetting connection", zap.Error(err))
		}
	}
}

func (wf *WakuFilterFullNode) ping(ctx context.Context, stream network.Stream, request *pb.FilterSubscribeRequest) {
	exists := wf.subscriptions.Has(stream.Conn().RemotePeer())

	if exists {
		wf.reply(ctx, stream, request, http.StatusOK)
	} else {
		wf.reply(ctx, stream, request, http.StatusNotFound, peerHasNoSubscription)
	}
}

func (wf *WakuFilterFullNode) subscribe(ctx context.Context, stream network.Stream, request *pb.FilterSubscribeRequest) {
	if wf.subscriptions.Count() >= wf.maxSubscriptions {
		wf.reply(ctx, stream, request, http.StatusServiceUnavailable, "node has reached maximum number of subscriptions")
		return
	}

	peerID := stream.Conn().RemotePeer()

	if totalSubs, exists := wf.subscriptions.Get(peerID); exists {
		ctTotal := 0
		for _, contentTopicSet := range totalSubs {
			ctTotal += len(contentTopicSet)
		}

		if ctTotal+len(request.ContentTopics) > MaxCriteriaPerSubscription {
			wf.reply(ctx, stream, request, http.StatusServiceUnavailable, "peer has reached maximum number of filter criteria")
			return
		}
	}

	wf.subscriptions.Set(peerID, *request.PubsubTopic, request.ContentTopics)

	wf.metrics.RecordSubscriptions(wf.subscriptions.Count())
	wf.reply(ctx, stream, request, http.StatusOK)
}

func (wf *WakuFilterFullNode) unsubscribe(ctx context.Context, stream network.Stream, request *pb.FilterSubscribeRequest) {
	err := wf.subscriptions.Delete(stream.Conn().RemotePeer(), *request.PubsubTopic, request.ContentTopics)
	if err != nil {
		wf.reply(ctx, stream, request, http.StatusNotFound, peerHasNoSubscription)
	} else {
		wf.metrics.RecordSubscriptions(wf.subscriptions.Count())
		wf.reply(ctx, stream, request, http.StatusOK)
	}
}

func (wf *WakuFilterFullNode) unsubscribeAll(ctx context.Context, stream network.Stream, request *pb.FilterSubscribeRequest) {
	err := wf.subscriptions.DeleteAll(stream.Conn().RemotePeer())
	if err != nil {
		wf.reply(ctx, stream, request, http.StatusNotFound, peerHasNoSubscription)
	} else {
		wf.metrics.RecordSubscriptions(wf.subscriptions.Count())
		wf.reply(ctx, stream, request, http.StatusOK)
	}
}

func (wf *WakuFilterFullNode) filterListener(ctx context.Context) {
	defer wf.WaitGroup().Done()

	// This function is invoked for each message received
	// on the full node in context of Waku2-Filter
	handle := func(envelope *protocol.Envelope) error {
		msg := envelope.Message()
		pubsubTopic := envelope.PubsubTopic()
		logger := utils.MessagesLogger("filter").With(logging.HexBytes("hash", envelope.Hash()),
			zap.String("pubsubTopic", envelope.PubsubTopic()),
			zap.String("contentTopic", envelope.Message().ContentTopic),
		)
		logger.Debug("push message to filter subscribers")

		// Each subscriber is a light node that earlier on invoked
		// a FilterRequest on this node
		for subscriber := range wf.subscriptions.Items(pubsubTopic, msg.ContentTopic) {
			logger := logger.With(logging.HostID("peer", subscriber))
			// Do a message push to light node
			logger.Debug("pushing message to light node")
			wf.WaitGroup().Add(1)
			go func(subscriber peer.ID) {
				defer wf.WaitGroup().Done()
				start := time.Now()
				err := wf.pushMessage(ctx, logger, subscriber, envelope)
				if err != nil {
					logger.Error("pushing message", zap.Error(err))
					return
				}
				wf.metrics.RecordPushDuration(time.Since(start))
			}(subscriber)
		}

		return nil
	}

	for m := range wf.msgSub.Ch {
		if err := handle(m); err != nil {
			wf.log.Error("handling message", zap.Error(err))
		}
	}
}

func (wf *WakuFilterFullNode) pushMessage(ctx context.Context, logger *zap.Logger, peerID peer.ID, env *protocol.Envelope) error {
	pubSubTopic := env.PubsubTopic()
	messagePush := &pb.MessagePush{
		PubsubTopic: &pubSubTopic,
		WakuMessage: env.Message(),
	}

	ctx, cancel := context.WithTimeout(ctx, MessagePushTimeout)
	defer cancel()

	stream, err := wf.h.NewStream(ctx, peerID, FilterPushID_v20beta1)
	if err != nil {
		wf.subscriptions.FlagAsFailure(peerID)
		if errors.Is(context.DeadlineExceeded, err) {
			wf.metrics.RecordError(pushTimeoutFailure)
		} else {
			wf.metrics.RecordError(dialFailure)
		}
		logger.Error("opening peer stream", zap.Error(err))
		return err
	}

	writer := pbio.NewDelimitedWriter(stream)
	err = writer.WriteMsg(messagePush)
	if err != nil {
		if errors.Is(context.DeadlineExceeded, err) {
			wf.metrics.RecordError(pushTimeoutFailure)
		} else {
			wf.metrics.RecordError(writeResponseFailure)
		}
		logger.Error("pushing messages to peer", zap.Error(err))
		wf.subscriptions.FlagAsFailure(peerID)
		if err := stream.Reset(); err != nil {
			wf.log.Error("resetting connection", zap.Error(err))
		}
		return nil
	}

	stream.Close()

	wf.subscriptions.FlagAsSuccess(peerID)

	logger.Debug("message pushed succesfully")

	return nil
}

// Stop unmounts the filter protocol
func (wf *WakuFilterFullNode) Stop() {
	wf.CommonService.Stop(func() {
		wf.h.RemoveStreamHandler(FilterSubscribeID_v20beta1)
		wf.msgSub.Unsubscribe()
	})
}
