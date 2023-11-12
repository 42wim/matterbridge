package store

import (
	"context"
	"encoding/hex"
	"errors"
	"math"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-msgio/pbio"
	"github.com/multiformats/go-multiaddr"
	"go.uber.org/zap"

	"github.com/waku-org/go-waku/logging"
	"github.com/waku-org/go-waku/waku/v2/peermanager"
	"github.com/waku-org/go-waku/waku/v2/peerstore"
	"github.com/waku-org/go-waku/waku/v2/protocol"
	wpb "github.com/waku-org/go-waku/waku/v2/protocol/pb"
	"github.com/waku-org/go-waku/waku/v2/protocol/store/pb"
)

type Query struct {
	PubsubTopic   string
	ContentTopics []string
	StartTime     *int64
	EndTime       *int64
}

// Result represents a valid response from a store node
type Result struct {
	started  bool
	Messages []*wpb.WakuMessage
	store    Store
	query    *pb.HistoryQuery
	cursor   *pb.Index
	peerID   peer.ID
}

func (r *Result) Cursor() *pb.Index {
	return r.cursor
}

func (r *Result) IsComplete() bool {
	return r.cursor == nil
}

func (r *Result) PeerID() peer.ID {
	return r.peerID
}

func (r *Result) Query() *pb.HistoryQuery {
	return r.query
}

func (r *Result) Next(ctx context.Context) (bool, error) {
	if !r.started {
		r.started = true
		return len(r.Messages) != 0, nil
	}

	if r.IsComplete() {
		return false, nil
	}

	newResult, err := r.store.Next(ctx, r)
	if err != nil {
		return false, err
	}

	r.cursor = newResult.cursor
	r.Messages = newResult.Messages

	return true, nil
}

func (r *Result) GetMessages() []*wpb.WakuMessage {
	if !r.started {
		return nil
	}
	return r.Messages
}

type CriteriaFN = func(msg *wpb.WakuMessage) (bool, error)

type HistoryRequestParameters struct {
	selectedPeer      peer.ID
	peerAddr          multiaddr.Multiaddr
	peerSelectionType peermanager.PeerSelection
	preferredPeers    peer.IDSlice
	localQuery        bool
	requestID         []byte
	cursor            *pb.Index
	pageSize          uint64
	asc               bool

	s *WakuStore
}

type HistoryRequestOption func(*HistoryRequestParameters) error

// WithPeer is an option used to specify the peerID to request the message history.
// Note that this option is mutually exclusive to WithPeerAddr, only one of them can be used.
func WithPeer(p peer.ID) HistoryRequestOption {
	return func(params *HistoryRequestParameters) error {
		params.selectedPeer = p
		if params.peerAddr != nil {
			return errors.New("peerId and peerAddr options are mutually exclusive")
		}
		return nil
	}
}

//WithPeerAddr is an option used to specify a peerAddress to request the message history.
// This new peer will be added to peerStore.
// Note that this option is mutually exclusive to WithPeerAddr, only one of them can be used.

func WithPeerAddr(pAddr multiaddr.Multiaddr) HistoryRequestOption {
	return func(params *HistoryRequestParameters) error {
		params.peerAddr = pAddr
		if params.selectedPeer != "" {
			return errors.New("peerAddr and peerId options are mutually exclusive")
		}
		return nil
	}
}

// WithAutomaticPeerSelection is an option used to randomly select a peer from the peer store
// to request the message history. If a list of specific peers is passed, the peer will be chosen
// from that list assuming it supports the chosen protocol, otherwise it will chose a peer
// from the node peerstore
// Note: This option is avaiable only with peerManager
func WithAutomaticPeerSelection(fromThesePeers ...peer.ID) HistoryRequestOption {
	return func(params *HistoryRequestParameters) error {
		params.peerSelectionType = peermanager.Automatic
		params.preferredPeers = fromThesePeers
		return nil
	}
}

// WithFastestPeerSelection is an option used to select a peer from the peer store
// with the lowest ping. If a list of specific peers is passed, the peer will be chosen
// from that list assuming it supports the chosen protocol, otherwise it will chose a peer
// from the node peerstore
// Note: This option is avaiable only with peerManager
func WithFastestPeerSelection(fromThesePeers ...peer.ID) HistoryRequestOption {
	return func(params *HistoryRequestParameters) error {
		params.peerSelectionType = peermanager.LowestRTT
		return nil
	}
}

// WithRequestID is an option to set a specific request ID to be used when
// creating a store request
func WithRequestID(requestID []byte) HistoryRequestOption {
	return func(params *HistoryRequestParameters) error {
		params.requestID = requestID
		return nil
	}
}

// WithAutomaticRequestID is an option to automatically generate a request ID
// when creating a store request
func WithAutomaticRequestID() HistoryRequestOption {
	return func(params *HistoryRequestParameters) error {
		params.requestID = protocol.GenerateRequestID()
		return nil
	}
}

func WithCursor(c *pb.Index) HistoryRequestOption {
	return func(params *HistoryRequestParameters) error {
		params.cursor = c
		return nil
	}
}

// WithPaging is an option used to specify the order and maximum number of records to return
func WithPaging(asc bool, pageSize uint64) HistoryRequestOption {
	return func(params *HistoryRequestParameters) error {
		params.asc = asc
		params.pageSize = pageSize
		return nil
	}
}

func WithLocalQuery() HistoryRequestOption {
	return func(params *HistoryRequestParameters) error {
		params.localQuery = true
		return nil
	}
}

// Default options to be used when querying a store node for results
func DefaultOptions() []HistoryRequestOption {
	return []HistoryRequestOption{
		WithAutomaticRequestID(),
		WithAutomaticPeerSelection(),
		WithPaging(true, DefaultPageSize),
	}
}

func (store *WakuStore) queryFrom(ctx context.Context, historyRequest *pb.HistoryRPC, selectedPeer peer.ID) (*pb.HistoryResponse, error) {
	logger := store.log.With(logging.HostID("peer", selectedPeer))
	logger.Info("querying message history")

	stream, err := store.h.NewStream(ctx, selectedPeer, StoreID_v20beta4)
	if err != nil {
		logger.Error("creating stream to peer", zap.Error(err))
		store.metrics.RecordError(dialFailure)
		return nil, err
	}

	writer := pbio.NewDelimitedWriter(stream)
	reader := pbio.NewDelimitedReader(stream, math.MaxInt32)

	err = writer.WriteMsg(historyRequest)
	if err != nil {
		logger.Error("writing request", zap.Error(err))
		store.metrics.RecordError(writeRequestFailure)
		if err := stream.Reset(); err != nil {
			store.log.Error("resetting connection", zap.Error(err))
		}
		return nil, err
	}

	historyResponseRPC := &pb.HistoryRPC{RequestId: historyRequest.RequestId}
	err = reader.ReadMsg(historyResponseRPC)
	if err != nil {
		logger.Error("reading response", zap.Error(err))
		store.metrics.RecordError(decodeRPCFailure)
		if err := stream.Reset(); err != nil {
			store.log.Error("resetting connection", zap.Error(err))
		}
		return nil, err
	}

	stream.Close()

	// nwaku does not return a response if there are no results due to the way their
	// protobuffer library works. this condition once they have proper proto3 support
	if historyResponseRPC.Response == nil {
		// Empty response
		return &pb.HistoryResponse{
			PagingInfo: &pb.PagingInfo{},
		}, nil
	}

	if err := historyResponseRPC.ValidateResponse(historyRequest.RequestId); err != nil {
		return nil, err
	}

	return historyResponseRPC.Response, nil
}

func (store *WakuStore) localQuery(historyQuery *pb.HistoryRPC) (*pb.HistoryResponse, error) {
	logger := store.log
	logger.Info("querying local message history")

	if !store.started {
		return nil, errors.New("not running local store")
	}

	historyResponseRPC := &pb.HistoryRPC{
		RequestId: historyQuery.RequestId,
		Response:  store.FindMessages(historyQuery.Query),
	}

	if historyResponseRPC.Response == nil {
		// Empty response
		return &pb.HistoryResponse{
			PagingInfo: &pb.PagingInfo{},
		}, nil
	}

	return historyResponseRPC.Response, nil
}

func (store *WakuStore) Query(ctx context.Context, query Query, opts ...HistoryRequestOption) (*Result, error) {
	params := new(HistoryRequestParameters)
	params.s = store

	optList := DefaultOptions()
	optList = append(optList, opts...)
	for _, opt := range optList {
		err := opt(params)
		if err != nil {
			return nil, err
		}
	}

	if !params.localQuery {
		pubsubTopics := []string{}
		if query.PubsubTopic == "" {
			for _, cTopic := range query.ContentTopics {
				pubsubTopic, err := protocol.GetPubSubTopicFromContentTopic(cTopic)
				if err != nil {
					return nil, err
				}
				pubsubTopics = append(pubsubTopics, pubsubTopic)
			}
		} else {
			pubsubTopics = append(pubsubTopics, query.PubsubTopic)
		}

		//Add Peer to peerstore.
		if store.pm != nil && params.peerAddr != nil {
			pData, err := store.pm.AddPeer(params.peerAddr, peerstore.Static, pubsubTopics, StoreID_v20beta4)
			if err != nil {
				return nil, err
			}
			store.pm.Connect(pData)
			params.selectedPeer = pData.AddrInfo.ID
		}
		if store.pm != nil && params.selectedPeer == "" {
			var err error
			params.selectedPeer, err = store.pm.SelectPeer(
				peermanager.PeerSelectionCriteria{
					SelectionType: params.peerSelectionType,
					Proto:         StoreID_v20beta4,
					PubsubTopics:  pubsubTopics,
					SpecificPeers: params.preferredPeers,
					Ctx:           ctx,
				},
			)
			if err != nil {
				return nil, err
			}
		}
	}

	historyRequest := &pb.HistoryRPC{
		RequestId: hex.EncodeToString(params.requestID),
		Query: &pb.HistoryQuery{
			PubsubTopic:    query.PubsubTopic,
			ContentFilters: []*pb.ContentFilter{},
			StartTime:      query.StartTime,
			EndTime:        query.EndTime,
			PagingInfo:     &pb.PagingInfo{},
		},
	}

	for _, cf := range query.ContentTopics {
		historyRequest.Query.ContentFilters = append(historyRequest.Query.ContentFilters, &pb.ContentFilter{ContentTopic: cf})
	}

	if !params.localQuery && params.selectedPeer == "" {
		store.metrics.RecordError(peerNotFoundFailure)
		return nil, ErrNoPeersAvailable
	}

	if params.cursor != nil {
		historyRequest.Query.PagingInfo.Cursor = params.cursor
	}

	if params.asc {
		historyRequest.Query.PagingInfo.Direction = pb.PagingInfo_FORWARD
	} else {
		historyRequest.Query.PagingInfo.Direction = pb.PagingInfo_BACKWARD
	}

	pageSize := params.pageSize
	if pageSize == 0 {
		pageSize = DefaultPageSize
	} else if pageSize > uint64(MaxPageSize) {
		pageSize = MaxPageSize
	}
	historyRequest.Query.PagingInfo.PageSize = pageSize

	err := historyRequest.ValidateQuery()
	if err != nil {
		return nil, err
	}

	var response *pb.HistoryResponse

	if params.localQuery {
		response, err = store.localQuery(historyRequest)
	} else {
		response, err = store.queryFrom(ctx, historyRequest, params.selectedPeer)
	}
	if err != nil {
		return nil, err
	}

	if response.Error == pb.HistoryResponse_INVALID_CURSOR {
		return nil, errors.New("invalid cursor")
	}

	result := &Result{
		store:    store,
		Messages: response.Messages,
		query:    historyRequest.Query,
		peerID:   params.selectedPeer,
	}

	if response.PagingInfo != nil {
		result.cursor = response.PagingInfo.Cursor
	}

	return result, nil
}

// Find the first message that matches a criteria. criteriaCB is a function that will be invoked for each message and returns true if the message matches the criteria
func (store *WakuStore) Find(ctx context.Context, query Query, cb CriteriaFN, opts ...HistoryRequestOption) (*wpb.WakuMessage, error) {
	if cb == nil {
		return nil, errors.New("callback can't be null")
	}

	result, err := store.Query(ctx, query, opts...)
	if err != nil {
		return nil, err
	}

	for {
		for _, m := range result.Messages {
			found, err := cb(m)
			if err != nil {
				return nil, err
			}

			if found {
				return m, nil
			}
		}

		if result.IsComplete() {
			break
		}

		result, err = store.Next(ctx, result)
		if err != nil {
			return nil, err
		}
	}

	return nil, nil
}

// Next is used with to retrieve the next page of rows from a query response.
// If no more records are found, the result will not contain any messages.
// This function is useful for iterating over results without having to manually
// specify the cursor and pagination order and max number of results
func (store *WakuStore) Next(ctx context.Context, r *Result) (*Result, error) {
	if r.IsComplete() {
		return &Result{
			store:    store,
			started:  true,
			Messages: []*wpb.WakuMessage{},
			cursor:   nil,
			query:    r.query,
			peerID:   r.PeerID(),
		}, nil
	}

	historyRequest := &pb.HistoryRPC{
		RequestId: hex.EncodeToString(protocol.GenerateRequestID()),
		Query:     r.Query(),
	}
	historyRequest.Query.PagingInfo.Cursor = r.Cursor()

	response, err := store.queryFrom(ctx, historyRequest, r.PeerID())
	if err != nil {
		return nil, err
	}

	if response.Error == pb.HistoryResponse_INVALID_CURSOR {
		return nil, errors.New("invalid cursor")
	}

	result := &Result{
		started:  true,
		store:    store,
		Messages: response.Messages,
		query:    historyRequest.Query,
		peerID:   r.PeerID(),
	}

	if response.PagingInfo != nil {
		result.cursor = response.PagingInfo.Cursor
	}

	return result, nil

}
