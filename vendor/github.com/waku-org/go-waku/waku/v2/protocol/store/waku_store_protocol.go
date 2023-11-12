package store

import (
	"context"
	"encoding/hex"
	"errors"
	"math"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-msgio/pbio"
	"go.uber.org/zap"

	"github.com/waku-org/go-waku/logging"
	"github.com/waku-org/go-waku/waku/persistence"
	"github.com/waku-org/go-waku/waku/v2/protocol"
	wpb "github.com/waku-org/go-waku/waku/v2/protocol/pb"
	"github.com/waku-org/go-waku/waku/v2/protocol/relay"
	"github.com/waku-org/go-waku/waku/v2/protocol/store/pb"
	"github.com/waku-org/go-waku/waku/v2/timesource"
)

func findMessages(query *pb.HistoryQuery, msgProvider MessageProvider) ([]*wpb.WakuMessage, *pb.PagingInfo, error) {
	if query.PagingInfo == nil {
		query.PagingInfo = &pb.PagingInfo{
			Direction: pb.PagingInfo_FORWARD,
		}
	}

	if query.PagingInfo.PageSize == 0 {
		query.PagingInfo.PageSize = DefaultPageSize
	} else if query.PagingInfo.PageSize > uint64(MaxPageSize) {
		query.PagingInfo.PageSize = MaxPageSize
	}

	cursor, queryResult, err := msgProvider.Query(query)
	if err != nil {
		return nil, nil, err
	}

	if len(queryResult) == 0 { // no pagination is needed for an empty list
		return nil, &pb.PagingInfo{Cursor: nil}, nil
	}

	resultMessages := make([]*wpb.WakuMessage, len(queryResult))
	for i := range queryResult {
		resultMessages[i] = queryResult[i].Message
	}

	return resultMessages, &pb.PagingInfo{Cursor: cursor}, nil
}

func (store *WakuStore) FindMessages(query *pb.HistoryQuery) *pb.HistoryResponse {
	result := new(pb.HistoryResponse)

	messages, newPagingInfo, err := findMessages(query, store.msgProvider)
	if err != nil {
		if err == persistence.ErrInvalidCursor {
			result.Error = pb.HistoryResponse_INVALID_CURSOR
		} else {
			// TODO: return error in pb.HistoryResponse
			store.log.Error("obtaining messages from db", zap.Error(err))
		}
	}

	result.Messages = messages
	result.PagingInfo = newPagingInfo
	return result
}

type MessageProvider interface {
	GetAll() ([]persistence.StoredMessage, error)
	Query(query *pb.HistoryQuery) (*pb.Index, []persistence.StoredMessage, error)
	Validate(env *protocol.Envelope) error
	Put(env *protocol.Envelope) error
	MostRecentTimestamp() (int64, error)
	Start(ctx context.Context, timesource timesource.Timesource) error
	Stop()
	Count() (int, error)
}

type Store interface {
	SetHost(h host.Host)
	Start(context.Context, *relay.Subscription) error
	Query(ctx context.Context, query Query, opts ...HistoryRequestOption) (*Result, error)
	Find(ctx context.Context, query Query, cb CriteriaFN, opts ...HistoryRequestOption) (*wpb.WakuMessage, error)
	Next(ctx context.Context, r *Result) (*Result, error)
	Resume(ctx context.Context, pubsubTopic string, peerList []peer.ID) (int, error)
	Stop()
}

// SetMessageProvider allows switching the message provider used with a WakuStore
func (store *WakuStore) SetMessageProvider(p MessageProvider) {
	store.msgProvider = p
}

// Sets the host to be able to mount or consume a protocol
func (store *WakuStore) SetHost(h host.Host) {
	store.h = h
}

// Start initializes the WakuStore by enabling the protocol and fetching records from a message provider
func (store *WakuStore) Start(ctx context.Context, sub *relay.Subscription) error {
	if store.started {
		return nil
	}

	if store.msgProvider == nil {
		store.log.Info("Store protocol started (no message provider)")
		return nil
	}

	err := store.msgProvider.Start(ctx, store.timesource) // TODO: store protocol should not start a message provider
	if err != nil {
		store.log.Error("Error starting message provider", zap.Error(err))
		return err
	}

	store.started = true
	store.ctx, store.cancel = context.WithCancel(ctx)
	store.MsgC = sub

	store.h.SetStreamHandlerMatch(StoreID_v20beta4, protocol.PrefixTextMatch(string(StoreID_v20beta4)), store.onRequest)

	store.wg.Add(1)
	go store.storeIncomingMessages(store.ctx)

	store.log.Info("Store protocol started")

	return nil
}

func (store *WakuStore) storeMessage(env *protocol.Envelope) error {

	if env.Message().GetEphemeral() {
		return nil
	}

	err := store.msgProvider.Validate(env)
	if err != nil {
		return err
	}

	err = store.msgProvider.Put(env)
	if err != nil {
		store.log.Error("storing message", zap.Error(err))
		store.metrics.RecordError(storeFailure)
		return err
	}

	return nil
}

func (store *WakuStore) storeIncomingMessages(ctx context.Context) {
	defer store.wg.Done()
	for envelope := range store.MsgC.Ch {
		go func(env *protocol.Envelope) {
			_ = store.storeMessage(env)
		}(envelope)
	}
}

func (store *WakuStore) onRequest(stream network.Stream) {
	logger := store.log.With(logging.HostID("peer", stream.Conn().RemotePeer()))
	historyRPCRequest := &pb.HistoryRPC{}

	writer := pbio.NewDelimitedWriter(stream)
	reader := pbio.NewDelimitedReader(stream, math.MaxInt32)

	err := reader.ReadMsg(historyRPCRequest)
	if err != nil {
		logger.Error("reading request", zap.Error(err))
		store.metrics.RecordError(decodeRPCFailure)
		if err := stream.Reset(); err != nil {
			store.log.Error("resetting connection", zap.Error(err))
		}
		return
	}

	if err := historyRPCRequest.ValidateQuery(); err != nil {
		logger.Error("invalid request received", zap.Error(err))
		store.metrics.RecordError(decodeRPCFailure)
		if err := stream.Reset(); err != nil {
			store.log.Error("resetting connection", zap.Error(err))
		}

		// TODO: If store protocol is updated to include error messages
		//       `err.Error()` can be returned as a response
		return
	}

	logger = logger.With(zap.String("id", historyRPCRequest.RequestId))
	if query := historyRPCRequest.Query; query != nil {
		logger = logger.With(logging.Filters(query.GetContentFilters()))
	} else {
		logger.Error("reading request", zap.Error(err))
		store.metrics.RecordError(emptyRPCQueryFailure)
		if err := stream.Reset(); err != nil {
			store.log.Error("resetting connection", zap.Error(err))
		}
		return
	}

	logger.Info("received history query")
	store.metrics.RecordQuery()

	historyResponseRPC := &pb.HistoryRPC{}
	historyResponseRPC.RequestId = historyRPCRequest.RequestId
	historyResponseRPC.Response = store.FindMessages(historyRPCRequest.Query)

	logger = logger.With(zap.Int("messages", len(historyResponseRPC.Response.Messages)))
	err = writer.WriteMsg(historyResponseRPC)
	if err != nil {
		logger.Error("writing response", zap.Error(err), logging.PagingInfo(historyResponseRPC.Response.PagingInfo))
		store.metrics.RecordError(writeResponseFailure)
		if err := stream.Reset(); err != nil {
			store.log.Error("resetting connection", zap.Error(err))
		}
		return
	}

	logger.Info("response sent")
	stream.Close()
}

// Stop closes the store message channel and removes the protocol stream handler
func (store *WakuStore) Stop() {
	if store.cancel == nil {
		return
	}

	store.cancel()

	store.started = false

	store.MsgC.Unsubscribe()

	if store.msgProvider != nil {
		store.msgProvider.Stop() // TODO: StoreProtocol should not stop a message provider
	}

	if store.h != nil {
		store.h.RemoveStreamHandler(StoreID_v20beta4)
	}

	store.wg.Wait()
}

type queryLoopCandidateResponse struct {
	peerID   peer.ID
	response *pb.HistoryResponse
	err      error
}

func (store *WakuStore) queryLoop(ctx context.Context, query *pb.HistoryQuery, candidateList []peer.ID) ([]*queryLoopCandidateResponse, error) {
	err := query.Validate()
	if err != nil {
		return nil, err
	}

	queryWg := sync.WaitGroup{}
	queryWg.Add(len(candidateList))

	resultChan := make(chan *queryLoopCandidateResponse, len(candidateList))

	// loops through the candidateList in order and sends the query to each until one of the query gets resolved successfully
	// returns the number of retrieved messages, or error if all the requests fail
	for _, peer := range candidateList {
		func() {
			defer queryWg.Done()

			historyRequest := &pb.HistoryRPC{
				RequestId: hex.EncodeToString(protocol.GenerateRequestID()),
				Query:     query,
			}

			result := &queryLoopCandidateResponse{
				peerID: peer,
			}

			response, err := store.queryFrom(ctx, historyRequest, peer)
			if err != nil {
				store.log.Error("resuming history", logging.HostID("peer", peer), zap.Error(err))
				result.err = err
			} else {
				result.response = response
			}

			resultChan <- result
		}()
	}

	queryWg.Wait()
	close(resultChan)

	var queryLoopResults []*queryLoopCandidateResponse
	for result := range resultChan {
		queryLoopResults = append(queryLoopResults, result)
	}

	return queryLoopResults, nil
}

func (store *WakuStore) findLastSeen() (int64, error) {
	return store.msgProvider.MostRecentTimestamp()
}

func max(x, y int64) int64 {
	if x > y {
		return x
	}
	return y
}

// Resume retrieves the history of waku messages published on the default waku pubsub topic since the last time the waku store node has been online
// messages are stored in the store node's messages field and in the message db
// the offline time window is measured as the difference between the current time and the timestamp of the most recent persisted waku message
// an offset of 20 second is added to the time window to count for nodes asynchrony
// the history is fetched from one of the peers persisted in the waku store node's peer manager unit
// peerList indicates the list of peers to query from. The history is fetched from the first available peer in this list. Such candidates should be found through a discovery method (to be developed).
// if no peerList is passed, one of the peers in the underlying peer manager unit of the store protocol is picked randomly to fetch the history from. The history gets fetched successfully if the dialed peer has been online during the queried time window.
// the resume proc returns the number of retrieved messages if no error occurs, otherwise returns the error string
func (store *WakuStore) Resume(ctx context.Context, pubsubTopic string, peerList []peer.ID) (int, error) {
	if !store.started {
		return 0, errors.New("can't resume: store has not started")
	}

	lastSeenTime, err := store.findLastSeen()
	if err != nil {
		return 0, err
	}

	offset := int64(20 * time.Nanosecond)
	currentTime := store.timesource.Now().UnixNano() + offset
	lastSeenTime = max(lastSeenTime-offset, 0)

	rpc := &pb.HistoryQuery{
		PubsubTopic: pubsubTopic,
		StartTime:   &lastSeenTime,
		EndTime:     &currentTime,
		PagingInfo: &pb.PagingInfo{
			PageSize:  0,
			Direction: pb.PagingInfo_BACKWARD,
		},
	}

	if len(peerList) == 0 {
		return -1, ErrNoPeersAvailable
	}

	queryLoopResults, err := store.queryLoop(ctx, rpc, peerList)
	if err != nil {
		store.log.Error("resuming history", zap.Error(err))
		return -1, ErrFailedToResumeHistory
	}

	msgCount := 0
	for _, r := range queryLoopResults {
		if err == nil && r.response.GetError() != pb.HistoryResponse_NONE {
			r.err = errors.New("invalid cursor")
		}

		if r.err != nil {
			store.log.Warn("could not resume message history", zap.Error(r.err), logging.HostID("peer", r.peerID))
			continue
		}

		for _, msg := range r.response.Messages {
			if err = store.storeMessage(protocol.NewEnvelope(msg, store.timesource.Now().UnixNano(), pubsubTopic)); err == nil {
				msgCount++
			}
		}
	}

	store.log.Info("retrieved messages since the last online time", zap.Int("messages", msgCount))

	return msgCount, nil
}
