package filter

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"net/http"
	"strings"
	"sync"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	libp2pProtocol "github.com/libp2p/go-libp2p/core/protocol"
	"github.com/libp2p/go-msgio/pbio"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/waku-org/go-waku/logging"
	"github.com/waku-org/go-waku/waku/v2/peermanager"
	"github.com/waku-org/go-waku/waku/v2/peerstore"
	"github.com/waku-org/go-waku/waku/v2/protocol"
	"github.com/waku-org/go-waku/waku/v2/protocol/filter/pb"
	wpb "github.com/waku-org/go-waku/waku/v2/protocol/pb"
	"github.com/waku-org/go-waku/waku/v2/protocol/relay"
	"github.com/waku-org/go-waku/waku/v2/protocol/subscription"
	"github.com/waku-org/go-waku/waku/v2/service"
	"github.com/waku-org/go-waku/waku/v2/timesource"
	"go.uber.org/zap"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
)

// FilterPushID_v20beta1 is the current Waku Filter protocol identifier used to allow
// filter service nodes to push messages matching registered subscriptions to this client.
const FilterPushID_v20beta1 = libp2pProtocol.ID("/vac/waku/filter-push/2.0.0-beta1")

var (
	ErrNoPeersAvailable     = errors.New("no suitable remote peers")
	ErrSubscriptionNotFound = errors.New("subscription not found")
)

type WakuFilterLightNode struct {
	*service.CommonService
	h             host.Host
	broadcaster   relay.Broadcaster //TODO: Move the broadcast functionality outside of relay client to a higher SDK layer.s
	timesource    timesource.Timesource
	metrics       Metrics
	log           *zap.Logger
	subscriptions *subscription.SubscriptionsMap
	pm            *peermanager.PeerManager
}

type WakuFilterPushError struct {
	Err    error
	PeerID peer.ID
}

type WakuFilterPushResult struct {
	errs []WakuFilterPushError
	sync.RWMutex
}

func (arr *WakuFilterPushResult) Add(err WakuFilterPushError) {
	arr.Lock()
	defer arr.Unlock()
	arr.errs = append(arr.errs, err)
}
func (arr *WakuFilterPushResult) Errors() []WakuFilterPushError {
	arr.RLock()
	defer arr.RUnlock()
	return arr.errs
}

// NewWakuFilterLightnode returns a new instance of Waku Filter struct setup according to the chosen parameter and options
// Note that broadcaster is optional.
// Takes an optional peermanager if WakuFilterLightnode is being created along with WakuNode.
// If using libp2p host, then pass peermanager as nil
func NewWakuFilterLightNode(broadcaster relay.Broadcaster, pm *peermanager.PeerManager,
	timesource timesource.Timesource, reg prometheus.Registerer, log *zap.Logger) *WakuFilterLightNode {
	wf := new(WakuFilterLightNode)
	wf.log = log.Named("filterv2-lightnode")
	wf.broadcaster = broadcaster
	wf.timesource = timesource
	wf.pm = pm
	wf.CommonService = service.NewCommonService()
	wf.metrics = newMetrics(reg)

	return wf
}

// Sets the host to be able to mount or consume a protocol
func (wf *WakuFilterLightNode) SetHost(h host.Host) {
	wf.h = h
}

func (wf *WakuFilterLightNode) Start(ctx context.Context) error {
	return wf.CommonService.Start(ctx, wf.start)

}

func (wf *WakuFilterLightNode) start() error {
	wf.subscriptions = subscription.NewSubscriptionMap(wf.log)
	wf.h.SetStreamHandlerMatch(FilterPushID_v20beta1, protocol.PrefixTextMatch(string(FilterPushID_v20beta1)), wf.onRequest(wf.Context()))

	wf.log.Info("filter-push protocol started")
	return nil
}

// Stop unmounts the filter protocol
func (wf *WakuFilterLightNode) Stop() {
	wf.CommonService.Stop(func() {
		wf.h.RemoveStreamHandler(FilterPushID_v20beta1)
		if wf.subscriptions.Count() > 0 {
			go func() {
				defer func() {
					_ = recover()
				}()
				res, err := wf.unsubscribeAll(wf.Context())
				if err != nil {
					wf.log.Warn("unsubscribing from full nodes", zap.Error(err))
				}

				for _, r := range res.Errors() {
					if r.Err != nil {
						wf.log.Warn("unsubscribing from full nodes", zap.Error(r.Err), logging.HostID("peerID", r.PeerID))
					}

				}
				wf.subscriptions.Clear()
			}()
		}
	})
}

func (wf *WakuFilterLightNode) onRequest(ctx context.Context) func(network.Stream) {
	return func(stream network.Stream) {
		peerID := stream.Conn().RemotePeer()

		logger := wf.log.With(logging.HostID("peerID", peerID))

		if !wf.subscriptions.IsSubscribedTo(peerID) {
			logger.Warn("received message push from unknown peer", logging.HostID("peerID", peerID))
			wf.metrics.RecordError(unknownPeerMessagePush)
			if err := stream.Reset(); err != nil {
				wf.log.Error("resetting connection", zap.Error(err))
			}
			return
		}

		reader := pbio.NewDelimitedReader(stream, math.MaxInt32)

		messagePush := &pb.MessagePush{}
		err := reader.ReadMsg(messagePush)
		if err != nil {
			logger.Error("reading message push", zap.Error(err))
			wf.metrics.RecordError(decodeRPCFailure)
			if err := stream.Reset(); err != nil {
				wf.log.Error("resetting connection", zap.Error(err))
			}
			return
		}

		stream.Close()

		if err = messagePush.Validate(); err != nil {
			logger.Warn("received invalid messagepush")
			return
		}

		pubSubTopic := ""
		//For now returning failure, this will get addressed with autosharding changes for filter.
		if messagePush.PubsubTopic == nil {
			pubSubTopic, err = protocol.GetPubSubTopicFromContentTopic(messagePush.WakuMessage.ContentTopic)
			if err != nil {
				logger.Error("could not derive pubSubTopic from contentTopic", zap.Error(err))
				wf.metrics.RecordError(decodeRPCFailure)
				if err := stream.Reset(); err != nil {
					wf.log.Error("resetting connection", zap.Error(err))
				}
				return
			}
		} else {
			pubSubTopic = *messagePush.PubsubTopic
		}

		logger = messagePush.WakuMessage.Logger(logger, pubSubTopic)

		if !wf.subscriptions.Has(peerID, protocol.NewContentFilter(pubSubTopic, messagePush.WakuMessage.ContentTopic)) {
			logger.Warn("received messagepush with invalid subscription parameters")
			wf.metrics.RecordError(invalidSubscriptionMessage)
			return
		}

		wf.metrics.RecordMessage()

		wf.notify(peerID, pubSubTopic, messagePush.WakuMessage)

		logger.Info("received message push")
	}
}

func (wf *WakuFilterLightNode) notify(remotePeerID peer.ID, pubsubTopic string, msg *wpb.WakuMessage) {
	envelope := protocol.NewEnvelope(msg, wf.timesource.Now().UnixNano(), pubsubTopic)

	if wf.broadcaster != nil {
		// Broadcasting message so it's stored
		wf.broadcaster.Submit(envelope)
	}
	// Notify filter subscribers
	wf.subscriptions.Notify(remotePeerID, envelope)
}

func (wf *WakuFilterLightNode) request(ctx context.Context, params *FilterSubscribeParameters,
	reqType pb.FilterSubscribeRequest_FilterSubscribeType, contentFilter protocol.ContentFilter) error {
	request := &pb.FilterSubscribeRequest{
		RequestId:           hex.EncodeToString(params.requestID),
		FilterSubscribeType: reqType,
		PubsubTopic:         &contentFilter.PubsubTopic,
		ContentTopics:       contentFilter.ContentTopicsList(),
	}

	err := request.Validate()
	if err != nil {
		return err
	}

	logger := wf.log.With(logging.HostID("peerID", params.selectedPeer))

	stream, err := wf.h.NewStream(ctx, params.selectedPeer, FilterSubscribeID_v20beta1)
	if err != nil {
		wf.metrics.RecordError(dialFailure)
		return err
	}

	writer := pbio.NewDelimitedWriter(stream)
	reader := pbio.NewDelimitedReader(stream, math.MaxInt32)

	logger.Debug("sending FilterSubscribeRequest", zap.Stringer("request", request))
	err = writer.WriteMsg(request)
	if err != nil {
		wf.metrics.RecordError(writeRequestFailure)
		logger.Error("sending FilterSubscribeRequest", zap.Error(err))
		if err := stream.Reset(); err != nil {
			logger.Error("resetting connection", zap.Error(err))
		}
		return err
	}

	filterSubscribeResponse := &pb.FilterSubscribeResponse{}
	err = reader.ReadMsg(filterSubscribeResponse)
	if err != nil {
		logger.Error("receiving FilterSubscribeResponse", zap.Error(err))
		wf.metrics.RecordError(decodeRPCFailure)
		if err := stream.Reset(); err != nil {
			logger.Error("resetting connection", zap.Error(err))
		}
		return err
	}

	stream.Close()

	if err = filterSubscribeResponse.Validate(); err != nil {
		wf.metrics.RecordError(decodeRPCFailure)
		logger.Error("validating response", zap.Error(err))
		return err

	}

	if filterSubscribeResponse.RequestId != request.RequestId {
		wf.log.Error("requestID mismatch", zap.String("expected", request.RequestId), zap.String("received", filterSubscribeResponse.RequestId))
		wf.metrics.RecordError(requestIDMismatch)
		err := NewFilterError(300, "request_id_mismatch")
		return &err
	}

	if filterSubscribeResponse.StatusCode != http.StatusOK {
		wf.metrics.RecordError(errorResponse)
		errMessage := ""
		if filterSubscribeResponse.StatusDesc != nil {
			errMessage = *filterSubscribeResponse.StatusDesc
		}
		err := NewFilterError(int(filterSubscribeResponse.StatusCode), errMessage)
		return &err
	}

	return nil
}

func (wf *WakuFilterLightNode) handleFilterSubscribeOptions(ctx context.Context, contentFilter protocol.ContentFilter, opts []FilterSubscribeOption) (*FilterSubscribeParameters, map[string][]string, error) {
	params := new(FilterSubscribeParameters)
	params.log = wf.log
	params.host = wf.h
	params.pm = wf.pm

	optList := DefaultSubscriptionOptions()
	optList = append(optList, opts...)
	for _, opt := range optList {
		err := opt(params)
		if err != nil {
			return nil, nil, err
		}
	}

	pubSubTopicMap, err := protocol.ContentFilterToPubSubTopicMap(contentFilter)
	if err != nil {
		return nil, nil, err
	}

	//Add Peer to peerstore.
	if params.pm != nil && params.peerAddr != nil {
		pData, err := wf.pm.AddPeer(params.peerAddr, peerstore.Static, maps.Keys(pubSubTopicMap), FilterSubscribeID_v20beta1)
		if err != nil {
			return nil, nil, err
		}
		wf.pm.Connect(pData)
		params.selectedPeer = pData.AddrInfo.ID
	}
	if params.pm != nil && params.selectedPeer == "" {
		params.selectedPeer, err = wf.pm.SelectPeer(
			peermanager.PeerSelectionCriteria{
				SelectionType: params.peerSelectionType,
				Proto:         FilterSubscribeID_v20beta1,
				PubsubTopics:  maps.Keys(pubSubTopicMap),
				SpecificPeers: params.preferredPeers,
				Ctx:           ctx,
			},
		)
		if err != nil {
			return nil, nil, err
		}
	}
	return params, pubSubTopicMap, nil
}

// Subscribe setups a subscription to receive messages that match a specific content filter
// If contentTopics passed result in different pubSub topics (due to Auto/Static sharding), then multiple subscription requests are sent to the peer.
// This may change if Filterv2 protocol is updated to handle such a scenario in a single request.
// Note: In case of partial failure, results are returned for successful subscriptions along with error indicating failed contentTopics.
func (wf *WakuFilterLightNode) Subscribe(ctx context.Context, contentFilter protocol.ContentFilter, opts ...FilterSubscribeOption) ([]*subscription.SubscriptionDetails, error) {
	wf.RLock()
	defer wf.RUnlock()
	if err := wf.ErrOnNotRunning(); err != nil {
		return nil, err
	}

	params, pubSubTopicMap, err := wf.handleFilterSubscribeOptions(ctx, contentFilter, opts)
	if err != nil {
		return nil, err
	}

	failedContentTopics := []string{}
	subscriptions := make([]*subscription.SubscriptionDetails, 0)
	for pubSubTopic, cTopics := range pubSubTopicMap {
		var selectedPeer peer.ID
		if params.pm != nil && params.selectedPeer == "" {
			selectedPeer, err = wf.pm.SelectPeer(
				peermanager.PeerSelectionCriteria{
					SelectionType: params.peerSelectionType,
					Proto:         FilterSubscribeID_v20beta1,
					PubsubTopics:  []string{pubSubTopic},
					SpecificPeers: params.preferredPeers,
					Ctx:           ctx,
				},
			)
		} else {
			selectedPeer = params.selectedPeer
		}
		if selectedPeer == "" {
			wf.metrics.RecordError(peerNotFoundFailure)
			wf.log.Error("selecting peer", zap.String("pubSubTopic", pubSubTopic), zap.Strings("contentTopics", cTopics),
				zap.Error(err))
			failedContentTopics = append(failedContentTopics, cTopics...)
			continue
		}

		var cFilter protocol.ContentFilter
		cFilter.PubsubTopic = pubSubTopic
		cFilter.ContentTopics = protocol.NewContentTopicSet(cTopics...)

		paramsCopy := params.Copy()
		paramsCopy.selectedPeer = selectedPeer
		err := wf.request(
			ctx,
			paramsCopy,
			pb.FilterSubscribeRequest_SUBSCRIBE,
			cFilter,
		)
		if err != nil {
			wf.log.Error("Failed to subscribe", zap.String("pubSubTopic", pubSubTopic), zap.Strings("contentTopics", cTopics),
				zap.Error(err))
			failedContentTopics = append(failedContentTopics, cTopics...)
			continue
		}
		subscriptions = append(subscriptions, wf.subscriptions.NewSubscription(selectedPeer, cFilter))
	}

	if len(failedContentTopics) > 0 {
		return subscriptions, fmt.Errorf("subscriptions failed for contentTopics: %s", strings.Join(failedContentTopics, ","))
	} else {
		return subscriptions, nil
	}
}

// FilterSubscription is used to obtain an object from which you could receive messages received via filter protocol
func (wf *WakuFilterLightNode) FilterSubscription(peerID peer.ID, contentFilter protocol.ContentFilter) (*subscription.SubscriptionDetails, error) {
	wf.RLock()
	defer wf.RUnlock()
	if err := wf.ErrOnNotRunning(); err != nil {
		return nil, err
	}

	if !wf.subscriptions.Has(peerID, contentFilter) {
		return nil, errors.New("subscription does not exist")
	}

	return wf.subscriptions.NewSubscription(peerID, contentFilter), nil
}

func (wf *WakuFilterLightNode) getUnsubscribeParameters(opts ...FilterSubscribeOption) (*FilterSubscribeParameters, error) {
	params := new(FilterSubscribeParameters)
	params.log = wf.log
	opts = append(DefaultUnsubscribeOptions(), opts...)
	for _, opt := range opts {
		err := opt(params)
		if err != nil {
			return nil, err
		}
	}

	return params, nil
}

func (wf *WakuFilterLightNode) Ping(ctx context.Context, peerID peer.ID, opts ...FilterPingOption) error {
	wf.RLock()
	defer wf.RUnlock()
	if err := wf.ErrOnNotRunning(); err != nil {
		return err
	}

	params := &FilterPingParameters{}
	for _, opt := range opts {
		opt(params)
	}
	if len(params.requestID) == 0 {
		params.requestID = protocol.GenerateRequestID()
	}

	return wf.request(
		ctx,
		&FilterSubscribeParameters{selectedPeer: peerID, requestID: params.requestID},
		pb.FilterSubscribeRequest_SUBSCRIBER_PING,
		protocol.ContentFilter{})
}

func (wf *WakuFilterLightNode) IsSubscriptionAlive(ctx context.Context, subscription *subscription.SubscriptionDetails) error {
	wf.RLock()
	defer wf.RUnlock()
	if err := wf.ErrOnNotRunning(); err != nil {
		return err
	}

	return wf.Ping(ctx, subscription.PeerID)
}

// Unsubscribe is used to stop receiving messages from a peer that match a content filter
func (wf *WakuFilterLightNode) Unsubscribe(ctx context.Context, contentFilter protocol.ContentFilter, opts ...FilterSubscribeOption) (*WakuFilterPushResult, error) {
	wf.RLock()
	defer wf.RUnlock()
	if err := wf.ErrOnNotRunning(); err != nil {
		return nil, err
	}

	if len(contentFilter.ContentTopics) == 0 {
		return nil, errors.New("at least one content topic is required")
	}

	if slices.Contains[string](contentFilter.ContentTopicsList(), "") {
		return nil, errors.New("one or more content topics specified is empty")
	}

	if len(contentFilter.ContentTopics) > MaxContentTopicsPerRequest {
		return nil, fmt.Errorf("exceeds maximum content topics: %d", MaxContentTopicsPerRequest)
	}

	params, err := wf.getUnsubscribeParameters(opts...)
	if err != nil {
		return nil, err
	}

	pubSubTopicMap, err := protocol.ContentFilterToPubSubTopicMap(contentFilter)
	if err != nil {
		return nil, err
	}
	result := &WakuFilterPushResult{}
	for pTopic, cTopics := range pubSubTopicMap {
		cFilter := protocol.NewContentFilter(pTopic, cTopics...)

		peers := make(map[peer.ID]struct{})
		subs := wf.subscriptions.GetSubscription(params.selectedPeer, cFilter)
		if len(subs) == 0 {
			result.Add(WakuFilterPushError{
				Err:    ErrSubscriptionNotFound,
				PeerID: params.selectedPeer,
			})
			continue
		}
		for _, sub := range subs {
			sub.Remove(cTopics...)
			peers[sub.PeerID] = struct{}{}
		}
		if params.wg != nil {
			params.wg.Add(len(peers))
		}
		// send unsubscribe request to all the peers
		for peerID := range peers {
			go func(peerID peer.ID) {
				defer func() {
					if params.wg != nil {
						params.wg.Done()
					}
				}()
				err := wf.unsubscribeFromServer(ctx, &FilterSubscribeParameters{selectedPeer: peerID, requestID: params.requestID}, cFilter)

				if params.wg != nil {
					result.Add(WakuFilterPushError{
						Err:    err,
						PeerID: peerID,
					})
				}
			}(peerID)
		}
	}
	if params.wg != nil {
		params.wg.Wait()
	}

	return result, nil
}

func (wf *WakuFilterLightNode) Subscriptions() []*subscription.SubscriptionDetails {
	subs := wf.subscriptions.GetSubscription("", protocol.ContentFilter{})
	return subs
}

func (wf *WakuFilterLightNode) IsListening(pubsubTopic, contentTopic string) bool {
	return wf.subscriptions.IsListening(pubsubTopic, contentTopic)

}

// UnsubscribeWithSubscription is used to close a particular subscription
// If there are no more subscriptions matching the passed [peer, contentFilter] pair,
// server unsubscribe is also performed
func (wf *WakuFilterLightNode) UnsubscribeWithSubscription(ctx context.Context, sub *subscription.SubscriptionDetails,
	opts ...FilterSubscribeOption) (*WakuFilterPushResult, error) {
	wf.RLock()
	defer wf.RUnlock()
	if err := wf.ErrOnNotRunning(); err != nil {
		return nil, err
	}

	params, err := wf.getUnsubscribeParameters(opts...)
	if err != nil {
		return nil, err
	}

	// Close this sub
	sub.Close()

	result := &WakuFilterPushResult{}

	if !wf.subscriptions.Has(sub.PeerID, sub.ContentFilter) {
		// Last sub for this [peer, contentFilter] pair
		params.selectedPeer = sub.PeerID
		err = wf.unsubscribeFromServer(ctx, params, sub.ContentFilter)
		result.Add(WakuFilterPushError{
			Err:    err,
			PeerID: sub.PeerID,
		})
	}
	return result, err

}

func (wf *WakuFilterLightNode) unsubscribeFromServer(ctx context.Context, params *FilterSubscribeParameters, cFilter protocol.ContentFilter) error {
	err := wf.request(ctx, params, pb.FilterSubscribeRequest_UNSUBSCRIBE, cFilter)
	if err != nil {
		ferr, ok := err.(*FilterError)
		if ok && ferr.Code == http.StatusNotFound {
			wf.log.Warn("peer does not have a subscription", logging.HostID("peerID", params.selectedPeer), zap.Error(err))
		} else {
			wf.log.Error("could not unsubscribe from peer", logging.HostID("peerID", params.selectedPeer), zap.Error(err))
		}
	}

	return err
}

// close all subscribe for selectedPeer or if selectedPeer == "", then all peers
// send the unsubscribeAll request to the peers
func (wf *WakuFilterLightNode) unsubscribeAll(ctx context.Context, opts ...FilterSubscribeOption) (*WakuFilterPushResult, error) {
	params, err := wf.getUnsubscribeParameters(opts...)
	if err != nil {
		return nil, err
	}
	result := &WakuFilterPushResult{}

	peers := make(map[peer.ID]struct{})
	subs := wf.subscriptions.GetSubscription(params.selectedPeer, protocol.ContentFilter{})
	if len(subs) == 0 && params.selectedPeer != "" {
		result.Add(WakuFilterPushError{
			Err:    err,
			PeerID: params.selectedPeer,
		})
		return result, ErrSubscriptionNotFound
	}
	for _, sub := range subs {
		sub.Close()
		peers[sub.PeerID] = struct{}{}
	}
	if params.wg != nil {
		params.wg.Add(len(peers))
	}
	for peerId := range peers {
		go func(peerID peer.ID) {
			defer func() {
				if params.wg != nil {
					params.wg.Done()
				}
				_ = recover()
			}()

			paramsCopy := params.Copy()
			paramsCopy.selectedPeer = peerID
			err := wf.request(
				ctx,
				params,
				pb.FilterSubscribeRequest_UNSUBSCRIBE_ALL,
				protocol.ContentFilter{})
			if err != nil {
				wf.log.Error("could not unsubscribe from peer", logging.HostID("peerID", peerID), zap.Error(err))
			}
			if params.wg != nil {
				result.Add(WakuFilterPushError{
					Err:    err,
					PeerID: peerID,
				})
			}
		}(peerId)
	}

	if params.wg != nil {
		params.wg.Wait()
	}

	return result, nil
}

// UnsubscribeAll is used to stop receiving messages from peer(s). It does not close subscriptions
func (wf *WakuFilterLightNode) UnsubscribeAll(ctx context.Context, opts ...FilterSubscribeOption) (*WakuFilterPushResult, error) {
	wf.RLock()
	defer wf.RUnlock()
	if err := wf.ErrOnNotRunning(); err != nil {
		return nil, err
	}

	return wf.unsubscribeAll(ctx, opts...)
}
