// Copyright 2017 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package mailserver

import (
	"crypto/ecdsa"
	"encoding/binary"
	"errors"
	"fmt"
	"math/rand"
	"sync"
	"time"

	prom "github.com/prometheus/client_golang/prometheus"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/rlp"
	gethbridge "github.com/status-im/status-go/eth-node/bridge/geth"
	"github.com/status-im/status-go/eth-node/crypto"
	"github.com/status-im/status-go/eth-node/types"
	"github.com/status-im/status-go/params"
	"github.com/status-im/status-go/waku"
	wakucommon "github.com/status-im/status-go/waku/common"
)

const (
	maxQueryRange = 24 * time.Hour
	maxQueryLimit = 1000
	// When we default the upper limit, we want to extend the range a bit
	// to accommodate for envelopes with slightly higher timestamp, in seconds
	whisperTTLSafeThreshold = 60
)

var (
	errDirectoryNotProvided        = errors.New("data directory not provided")
	errDecryptionMethodNotProvided = errors.New("decryption method is not provided")
)

const (
	timestampLength        = 4
	requestLimitLength     = 4
	requestTimeRangeLength = timestampLength * 2
	processRequestTimeout  = time.Minute
)

type Config struct {
	// DataDir points to a directory where mailserver's data is stored.
	DataDir string
	// Password is used to create a symmetric key to decrypt requests.
	Password string
	// AsymKey is an asymmetric key to decrypt requests.
	AsymKey string
	// MininumPoW is a minimum PoW for requests.
	MinimumPoW float64
	// RateLimit is a maximum number of requests per second from a peer.
	RateLimit int
	// DataRetention specifies a number of days an envelope should be stored for.
	DataRetention   int
	PostgresEnabled bool
	PostgresURI     string
}

// --------------
// WakuMailServer
// --------------

type WakuMailServer struct {
	ms            *mailServer
	shh           *waku.Waku
	minRequestPoW float64

	symFilter  *wakucommon.Filter
	asymFilter *wakucommon.Filter
}

func (s *WakuMailServer) Init(waku *waku.Waku, cfg *params.WakuConfig) error {
	s.shh = waku
	s.minRequestPoW = cfg.MinimumPoW

	config := Config{
		DataDir:         cfg.DataDir,
		Password:        cfg.MailServerPassword,
		MinimumPoW:      cfg.MinimumPoW,
		DataRetention:   cfg.MailServerDataRetention,
		RateLimit:       cfg.MailServerRateLimit,
		PostgresEnabled: cfg.DatabaseConfig.PGConfig.Enabled,
		PostgresURI:     cfg.DatabaseConfig.PGConfig.URI,
	}
	var err error
	s.ms, err = newMailServer(
		config,
		&wakuAdapter{},
		&wakuService{Waku: waku},
	)
	if err != nil {
		return err
	}

	if err := s.setupDecryptor(config.Password, config.AsymKey); err != nil {
		return err
	}

	return nil
}

func (s *WakuMailServer) Close() {
	s.ms.Close()
}

func (s *WakuMailServer) Archive(env *wakucommon.Envelope) {
	s.ms.Archive(gethbridge.NewWakuEnvelope(env))
}

func (s *WakuMailServer) Deliver(peerID []byte, req wakucommon.MessagesRequest) {
	s.ms.DeliverMail(types.BytesToHash(peerID), types.BytesToHash(req.ID), MessagesRequestPayload{
		Lower:  req.From,
		Upper:  req.To,
		Bloom:  req.Bloom,
		Topics: req.Topics,
		Limit:  req.Limit,
		Cursor: req.Cursor,
		Batch:  true,
	})
}

// DEPRECATED; user Deliver instead
func (s *WakuMailServer) DeliverMail(peerID []byte, req *wakucommon.Envelope) {
	payload, err := s.decodeRequest(peerID, req)
	if err != nil {
		deliveryFailuresCounter.WithLabelValues("validation").Inc()
		log.Error(
			"[mailserver:DeliverMail] request failed validaton",
			"peerID", types.BytesToHash(peerID),
			"requestID", req.Hash().String(),
			"err", err,
		)
		s.ms.sendHistoricMessageErrorResponse(types.BytesToHash(peerID), types.Hash(req.Hash()), err)
		return
	}

	s.ms.DeliverMail(types.BytesToHash(peerID), types.Hash(req.Hash()), payload)
}

// bloomFromReceivedMessage for a given whisper.ReceivedMessage it extracts the
// used bloom filter.
func (s *WakuMailServer) bloomFromReceivedMessage(msg *wakucommon.ReceivedMessage) ([]byte, error) {
	payloadSize := len(msg.Payload)

	if payloadSize < 8 {
		return nil, errors.New("Undersized p2p request")
	} else if payloadSize == 8 {
		return wakucommon.MakeFullNodeBloom(), nil
	} else if payloadSize < 8+wakucommon.BloomFilterSize {
		return nil, errors.New("Undersized bloom filter in p2p request")
	}

	return msg.Payload[8 : 8+wakucommon.BloomFilterSize], nil
}

func (s *WakuMailServer) decompositeRequest(peerID []byte, request *wakucommon.Envelope) (MessagesRequestPayload, error) {
	var (
		payload MessagesRequestPayload
		err     error
	)

	if s.minRequestPoW > 0.0 && request.PoW() < s.minRequestPoW {
		return payload, fmt.Errorf("PoW() is too low")
	}

	decrypted := s.openEnvelope(request)
	if decrypted == nil {
		return payload, fmt.Errorf("failed to decrypt p2p request")
	}

	if err := checkMsgSignature(decrypted.Src, peerID); err != nil {
		return payload, err
	}

	payload.Bloom, err = s.bloomFromReceivedMessage(decrypted)
	if err != nil {
		return payload, err
	}

	payload.Lower = binary.BigEndian.Uint32(decrypted.Payload[:4])
	payload.Upper = binary.BigEndian.Uint32(decrypted.Payload[4:8])

	if payload.Upper < payload.Lower {
		err := fmt.Errorf("query range is invalid: from > to (%d > %d)", payload.Lower, payload.Upper)
		return payload, err
	}

	lowerTime := time.Unix(int64(payload.Lower), 0)
	upperTime := time.Unix(int64(payload.Upper), 0)
	if upperTime.Sub(lowerTime) > maxQueryRange {
		err := fmt.Errorf("query range too big for peer %s", string(peerID))
		return payload, err
	}

	if len(decrypted.Payload) >= requestTimeRangeLength+wakucommon.BloomFilterSize+requestLimitLength {
		payload.Limit = binary.BigEndian.Uint32(decrypted.Payload[requestTimeRangeLength+wakucommon.BloomFilterSize:])
	}

	if len(decrypted.Payload) == requestTimeRangeLength+wakucommon.BloomFilterSize+requestLimitLength+DBKeyLength {
		payload.Cursor = decrypted.Payload[requestTimeRangeLength+wakucommon.BloomFilterSize+requestLimitLength:]
	}

	return payload, nil
}

func (s *WakuMailServer) setupDecryptor(password, asymKey string) error {
	s.symFilter = nil
	s.asymFilter = nil

	if password != "" {
		keyID, err := s.shh.AddSymKeyFromPassword(password)
		if err != nil {
			return fmt.Errorf("create symmetric key: %v", err)
		}

		symKey, err := s.shh.GetSymKey(keyID)
		if err != nil {
			return fmt.Errorf("save symmetric key: %v", err)
		}

		s.symFilter = &wakucommon.Filter{KeySym: symKey}
	}

	if asymKey != "" {
		keyAsym, err := crypto.HexToECDSA(asymKey)
		if err != nil {
			return err
		}
		s.asymFilter = &wakucommon.Filter{KeyAsym: keyAsym}
	}

	return nil
}

// openEnvelope tries to decrypt an envelope, first based on asymetric key (if
// provided) and second on the symetric key (if provided)
func (s *WakuMailServer) openEnvelope(request *wakucommon.Envelope) *wakucommon.ReceivedMessage {
	if s.asymFilter != nil {
		if d := request.Open(s.asymFilter); d != nil {
			return d
		}
	}
	if s.symFilter != nil {
		if d := request.Open(s.symFilter); d != nil {
			return d
		}
	}
	return nil
}

func (s *WakuMailServer) decodeRequest(peerID []byte, request *wakucommon.Envelope) (MessagesRequestPayload, error) {
	var payload MessagesRequestPayload

	if s.minRequestPoW > 0.0 && request.PoW() < s.minRequestPoW {
		return payload, errors.New("PoW too low")
	}

	decrypted := s.openEnvelope(request)
	if decrypted == nil {
		log.Warn("Failed to decrypt p2p request")
		return payload, errors.New("failed to decrypt p2p request")
	}

	if err := checkMsgSignature(decrypted.Src, peerID); err != nil {
		log.Warn("Check message signature failed", "err", err.Error())
		return payload, fmt.Errorf("check message signature failed: %v", err)
	}

	if err := rlp.DecodeBytes(decrypted.Payload, &payload); err != nil {
		return payload, fmt.Errorf("failed to decode data: %v", err)
	}

	if payload.Upper == 0 {
		payload.Upper = uint32(time.Now().Unix() + whisperTTLSafeThreshold)
	}

	if payload.Upper < payload.Lower {
		log.Error("Query range is invalid: lower > upper", "lower", payload.Lower, "upper", payload.Upper)
		return payload, errors.New("query range is invalid: lower > upper")
	}

	return payload, nil
}

// -------
// adapter
// -------

type adapter interface {
	CreateRequestFailedPayload(reqID types.Hash, err error) []byte
	CreateRequestCompletedPayload(reqID, lastEnvelopeHash types.Hash, cursor []byte) []byte
	CreateSyncResponse(envelopes []types.Envelope, cursor []byte, final bool, err string) interface{}
	CreateRawSyncResponse(envelopes []rlp.RawValue, cursor []byte, final bool, err string) interface{}
}

// -----------
// wakuAdapter
// -----------

type wakuAdapter struct{}

var _ adapter = (*wakuAdapter)(nil)

func (wakuAdapter) CreateRequestFailedPayload(reqID types.Hash, err error) []byte {
	return waku.CreateMailServerRequestFailedPayload(common.Hash(reqID), err)
}

func (wakuAdapter) CreateRequestCompletedPayload(reqID, lastEnvelopeHash types.Hash, cursor []byte) []byte {
	return waku.CreateMailServerRequestCompletedPayload(common.Hash(reqID), common.Hash(lastEnvelopeHash), cursor)
}

func (wakuAdapter) CreateSyncResponse(_ []types.Envelope, _ []byte, _ bool, _ string) interface{} {
	return nil
}

func (wakuAdapter) CreateRawSyncResponse(_ []rlp.RawValue, _ []byte, _ bool, _ string) interface{} {
	return nil
}

// -------
// service
// -------

type service interface {
	SendHistoricMessageResponse(peerID []byte, payload []byte) error
	SendRawP2PDirect(peerID []byte, envelopes ...rlp.RawValue) error
	MaxMessageSize() uint32
	SendRawSyncResponse(peerID []byte, data interface{}) error // optional
	SendSyncResponse(peerID []byte, data interface{}) error    // optional
}

// -----------
// wakuService
// -----------

type wakuService struct {
	*waku.Waku
}

func (s *wakuService) SendRawSyncResponse(peerID []byte, data interface{}) error {
	return errors.New("syncing mailservers is not support by Waku")
}

func (s *wakuService) SendSyncResponse(peerID []byte, data interface{}) error {
	return errors.New("syncing mailservers is not support by Waku")
}

// ----------
// mailServer
// ----------

type mailServer struct {
	adapter       adapter
	service       service
	db            DB
	cleaner       *dbCleaner // removes old envelopes
	muRateLimiter sync.RWMutex
	rateLimiter   *rateLimiter
}

func newMailServer(cfg Config, adapter adapter, service service) (*mailServer, error) {
	if len(cfg.DataDir) == 0 {
		return nil, errDirectoryNotProvided
	}

	// TODO: move out
	if len(cfg.Password) == 0 && len(cfg.AsymKey) == 0 {
		return nil, errDecryptionMethodNotProvided
	}

	s := mailServer{
		adapter: adapter,
		service: service,
	}

	if cfg.RateLimit > 0 {
		s.setupRateLimiter(time.Duration(cfg.RateLimit) * time.Second)
	}

	// Open database in the last step in order not to init with error
	// and leave the database open by accident.
	if cfg.PostgresEnabled {
		log.Info("Connecting to postgres database")
		database, err := NewPostgresDB(cfg.PostgresURI)
		if err != nil {
			return nil, fmt.Errorf("open DB: %s", err)
		}
		s.db = database
		log.Info("Connected to postgres database")
	} else {
		// Defaults to LevelDB
		database, err := NewLevelDB(cfg.DataDir)
		if err != nil {
			return nil, fmt.Errorf("open DB: %s", err)
		}
		s.db = database
	}

	if cfg.DataRetention > 0 {
		// MailServerDataRetention is a number of days.
		s.setupCleaner(time.Duration(cfg.DataRetention) * time.Hour * 24)
	}

	return &s, nil
}

// setupRateLimiter in case limit is bigger than 0 it will setup an automated
// limit db cleanup.
func (s *mailServer) setupRateLimiter(limit time.Duration) {
	s.rateLimiter = newRateLimiter(limit)
	s.rateLimiter.Start()
}

func (s *mailServer) setupCleaner(retention time.Duration) {
	s.cleaner = newDBCleaner(s.db, retention)
	s.cleaner.Start()
}

func (s *mailServer) Archive(env types.Envelope) {
	err := s.db.SaveEnvelope(env)
	if err != nil {
		log.Error("Could not save envelope", "hash", env.Hash().String())
	}
}

func (s *mailServer) DeliverMail(peerID, reqID types.Hash, req MessagesRequestPayload) {
	timer := prom.NewTimer(mailDeliveryDuration)
	defer timer.ObserveDuration()

	deliveryAttemptsCounter.Inc()
	log.Info(
		"[mailserver:DeliverMail] delivering mail",
		"peerID", peerID.String(),
		"requestID", reqID.String(),
	)

	req.SetDefaults()

	log.Info(
		"[mailserver:DeliverMail] processing request",
		"peerID", peerID.String(),
		"requestID", reqID.String(),
		"lower", req.Lower,
		"upper", req.Upper,
		"bloom", req.Bloom,
		"topics", req.Topics,
		"limit", req.Limit,
		"cursor", req.Cursor,
		"batch", req.Batch,
	)

	if err := req.Validate(); err != nil {
		syncFailuresCounter.WithLabelValues("req_invalid").Inc()
		log.Error(
			"[mailserver:DeliverMail] request invalid",
			"peerID", peerID.String(),
			"requestID", reqID.String(),
			"err", err,
		)
		s.sendHistoricMessageErrorResponse(peerID, reqID, fmt.Errorf("request is invalid: %v", err))
		return
	}

	if s.exceedsPeerRequests(peerID) {
		deliveryFailuresCounter.WithLabelValues("peer_req_limit").Inc()
		log.Error(
			"[mailserver:DeliverMail] peer exceeded the limit",
			"peerID", peerID.String(),
			"requestID", reqID.String(),
		)
		s.sendHistoricMessageErrorResponse(peerID, reqID, fmt.Errorf("rate limit exceeded"))
		return
	}

	if req.Batch {
		requestsBatchedCounter.Inc()
	}

	iter, err := s.createIterator(req)
	if err != nil {
		log.Error(
			"[mailserver:DeliverMail] request failed",
			"peerID", peerID.String(),
			"requestID", reqID.String(),
			"err", err,
		)
		return
	}
	defer func() { _ = iter.Release() }()

	bundles := make(chan []rlp.RawValue, 5)
	errCh := make(chan error)
	cancelProcessing := make(chan struct{})

	go func() {
		counter := 0
		for bundle := range bundles {
			if err := s.sendRawEnvelopes(peerID, bundle, req.Batch); err != nil {
				close(cancelProcessing)
				errCh <- err
				break
			}
			counter++
		}
		close(errCh)
		log.Info(
			"[mailserver:DeliverMail] finished sending bundles",
			"peerID", peerID,
			"requestID", reqID.String(),
			"counter", counter,
		)
	}()

	nextPageCursor, lastEnvelopeHash := s.processRequestInBundles(
		iter,
		req.Bloom,
		req.Topics,
		int(req.Limit),
		processRequestTimeout,
		reqID.String(),
		bundles,
		cancelProcessing,
	)

	// Wait for the goroutine to finish the work. It may return an error.
	if err := <-errCh; err != nil {
		deliveryFailuresCounter.WithLabelValues("process").Inc()
		log.Error(
			"[mailserver:DeliverMail] error while processing",
			"err", err,
			"peerID", peerID,
			"requestID", reqID,
		)
		s.sendHistoricMessageErrorResponse(peerID, reqID, err)
		return
	}

	// Processing of the request could be finished earlier due to iterator error.
	if err := iter.Error(); err != nil {
		deliveryFailuresCounter.WithLabelValues("iterator").Inc()
		log.Error(
			"[mailserver:DeliverMail] iterator failed",
			"err", err,
			"peerID", peerID,
			"requestID", reqID,
		)
		s.sendHistoricMessageErrorResponse(peerID, reqID, err)
		return
	}

	log.Info(
		"[mailserver:DeliverMail] sending historic message response",
		"peerID", peerID,
		"requestID", reqID,
		"last", lastEnvelopeHash,
		"next", nextPageCursor,
	)

	s.sendHistoricMessageResponse(peerID, reqID, lastEnvelopeHash, nextPageCursor)
}

func (s *mailServer) SyncMail(peerID types.Hash, req MessagesRequestPayload) error {
	log.Info("Started syncing envelopes", "peer", peerID.String(), "req", req)

	requestID := fmt.Sprintf("%d-%d", time.Now().UnixNano(), rand.Intn(1000)) // nolint: gosec

	syncAttemptsCounter.Inc()

	// Check rate limiting for a requesting peer.
	if s.exceedsPeerRequests(peerID) {
		syncFailuresCounter.WithLabelValues("req_per_sec_limit").Inc()
		log.Error("Peer exceeded request per seconds limit", "peerID", peerID.String())
		return fmt.Errorf("requests per seconds limit exceeded")
	}

	req.SetDefaults()

	if err := req.Validate(); err != nil {
		syncFailuresCounter.WithLabelValues("req_invalid").Inc()
		return fmt.Errorf("request is invalid: %v", err)
	}

	iter, err := s.createIterator(req)
	if err != nil {
		syncFailuresCounter.WithLabelValues("iterator").Inc()
		return err
	}
	defer func() { _ = iter.Release() }()

	bundles := make(chan []rlp.RawValue, 5)
	errCh := make(chan error)
	cancelProcessing := make(chan struct{})

	go func() {
		for bundle := range bundles {
			resp := s.adapter.CreateRawSyncResponse(bundle, nil, false, "")
			if err := s.service.SendRawSyncResponse(peerID.Bytes(), resp); err != nil {
				close(cancelProcessing)
				errCh <- fmt.Errorf("failed to send sync response: %v", err)
				break
			}
		}
		close(errCh)
	}()

	nextCursor, _ := s.processRequestInBundles(
		iter,
		req.Bloom,
		req.Topics,
		int(req.Limit),
		processRequestTimeout,
		requestID,
		bundles,
		cancelProcessing,
	)

	// Wait for the goroutine to finish the work. It may return an error.
	if err := <-errCh; err != nil {
		syncFailuresCounter.WithLabelValues("routine").Inc()
		_ = s.service.SendSyncResponse(
			peerID.Bytes(),
			s.adapter.CreateSyncResponse(nil, nil, false, "failed to send a response"),
		)
		return err
	}

	// Processing of the request could be finished earlier due to iterator error.
	if err := iter.Error(); err != nil {
		syncFailuresCounter.WithLabelValues("iterator").Inc()
		_ = s.service.SendSyncResponse(
			peerID.Bytes(),
			s.adapter.CreateSyncResponse(nil, nil, false, "failed to process all envelopes"),
		)
		return fmt.Errorf("LevelDB iterator failed: %v", err)
	}

	log.Info("Finished syncing envelopes", "peer", peerID.String())

	err = s.service.SendSyncResponse(
		peerID.Bytes(),
		s.adapter.CreateSyncResponse(nil, nextCursor, true, ""),
	)
	if err != nil {
		syncFailuresCounter.WithLabelValues("response_send").Inc()
		return fmt.Errorf("failed to send the final sync response: %v", err)
	}

	return nil
}

// Close the mailserver and its associated db connection.
func (s *mailServer) Close() {
	if s.db != nil {
		if err := s.db.Close(); err != nil {
			log.Error("closing database failed", "err", err)
		}
	}
	if s.rateLimiter != nil {
		s.rateLimiter.Stop()
	}
	if s.cleaner != nil {
		s.cleaner.Stop()
	}
}

func (s *mailServer) exceedsPeerRequests(peerID types.Hash) bool {
	s.muRateLimiter.RLock()
	defer s.muRateLimiter.RUnlock()

	if s.rateLimiter == nil {
		return false
	}

	if s.rateLimiter.IsAllowed(peerID.String()) {
		s.rateLimiter.Add(peerID.String())
		return false
	}

	log.Info("peerID exceeded the number of requests per second", "peerID", peerID.String())
	return true
}

func (s *mailServer) createIterator(req MessagesRequestPayload) (Iterator, error) {
	var (
		emptyHash  types.Hash
		emptyTopic types.TopicType
		ku, kl     *DBKey
	)

	ku = NewDBKey(req.Upper+1, emptyTopic, emptyHash)
	kl = NewDBKey(req.Lower, emptyTopic, emptyHash)

	query := CursorQuery{
		start:  kl.Bytes(),
		end:    ku.Bytes(),
		cursor: req.Cursor,
		topics: req.Topics,
		bloom:  req.Bloom,
		limit:  req.Limit,
	}
	return s.db.BuildIterator(query)
}

func (s *mailServer) processRequestInBundles(
	iter Iterator,
	bloom []byte,
	topics [][]byte,
	limit int,
	timeout time.Duration,
	requestID string,
	output chan<- []rlp.RawValue,
	cancel <-chan struct{},
) ([]byte, types.Hash) {
	timer := prom.NewTimer(requestsInBundlesDuration)
	defer timer.ObserveDuration()

	var (
		bundle                 []rlp.RawValue
		bundleSize             uint32
		batches                [][]rlp.RawValue
		processedEnvelopes     int
		processedEnvelopesSize int64
		nextCursor             []byte
		lastEnvelopeHash       types.Hash
	)

	log.Info(
		"[mailserver:processRequestInBundles] processing request",
		"requestID", requestID,
		"limit", limit,
	)

	var topicsMap map[types.TopicType]bool

	if len(topics) != 0 {
		topicsMap = make(map[types.TopicType]bool)
		for _, t := range topics {
			topicsMap[types.BytesToTopic(t)] = true
		}
	}

	// We iterate over the envelopes.
	// We collect envelopes in batches.
	// If there still room and we haven't reached the limit
	// append and continue.
	// Otherwise publish what you have so far, reset the bundle to the
	// current envelope, and leave if we hit the limit
	for iter.Next() {
		var rawValue []byte
		var err error
		if len(topicsMap) != 0 {
			rawValue, err = iter.GetEnvelopeByTopicsMap(topicsMap)

		} else if len(bloom) != 0 {
			rawValue, err = iter.GetEnvelopeByBloomFilter(bloom)
		} else {
			err = errors.New("either topics or bloom must be specified")
		}
		if err != nil {
			log.Error(
				"[mailserver:processRequestInBundles]Failed to get envelope from iterator",
				"err", err,
				"requestID", requestID,
			)
			continue
		}

		if rawValue == nil {
			continue
		}

		key, err := iter.DBKey()
		if err != nil {
			log.Error(
				"[mailserver:processRequestInBundles] failed getting key",
				"requestID", requestID,
			)
			break

		}

		// TODO(adam): this is invalid code. If the limit is 1000,
		// it will only send 999 items and send a cursor.
		lastEnvelopeHash = key.EnvelopeHash()
		processedEnvelopes++
		envelopeSize := uint32(len(rawValue))
		limitReached := processedEnvelopes >= limit
		newSize := bundleSize + envelopeSize

		// If we still have some room for messages, add and continue
		if !limitReached && newSize < s.service.MaxMessageSize() {
			bundle = append(bundle, rawValue)
			bundleSize = newSize
			continue
		}

		// Publish if anything is in the bundle (there should always be
		// something unless limit = 1)
		if len(bundle) != 0 {
			batches = append(batches, bundle)
			processedEnvelopesSize += int64(bundleSize)
		}

		// Reset the bundle with the current envelope
		bundle = []rlp.RawValue{rawValue}
		bundleSize = envelopeSize

		// Leave if we reached the limit
		if limitReached {
			nextCursor = key.Cursor()
			break
		}
	}

	if len(bundle) > 0 {
		batches = append(batches, bundle)
		processedEnvelopesSize += int64(bundleSize)
	}

	log.Info(
		"[mailserver:processRequestInBundles] publishing envelopes",
		"requestID", requestID,
		"batchesCount", len(batches),
		"envelopeCount", processedEnvelopes,
		"processedEnvelopesSize", processedEnvelopesSize,
		"cursor", nextCursor,
	)

	// Publish
batchLoop:
	for _, batch := range batches {
		select {
		case output <- batch:
		// It might happen that during producing the batches,
		// the connection with the peer goes down and
		// the consumer of `output` channel exits prematurely.
		// In such a case, we should stop pushing batches and exit.
		case <-cancel:
			log.Info(
				"[mailserver:processRequestInBundles] failed to push all batches",
				"requestID", requestID,
			)
			break batchLoop
		case <-time.After(timeout):
			log.Error(
				"[mailserver:processRequestInBundles] timed out pushing a batch",
				"requestID", requestID,
			)
			break batchLoop
		}
	}

	envelopesCounter.Inc()
	sentEnvelopeBatchSizeMeter.Observe(float64(processedEnvelopesSize))

	log.Info(
		"[mailserver:processRequestInBundles] envelopes published",
		"requestID", requestID,
	)
	close(output)

	return nextCursor, lastEnvelopeHash
}

func (s *mailServer) sendRawEnvelopes(peerID types.Hash, envelopes []rlp.RawValue, batch bool) error {
	timer := prom.NewTimer(sendRawEnvelopeDuration)
	defer timer.ObserveDuration()

	if batch {
		return s.service.SendRawP2PDirect(peerID.Bytes(), envelopes...)
	}

	for _, env := range envelopes {
		if err := s.service.SendRawP2PDirect(peerID.Bytes(), env); err != nil {
			return err
		}
	}

	return nil
}

func (s *mailServer) sendHistoricMessageResponse(peerID, reqID, lastEnvelopeHash types.Hash, cursor []byte) {
	payload := s.adapter.CreateRequestCompletedPayload(reqID, lastEnvelopeHash, cursor)
	err := s.service.SendHistoricMessageResponse(peerID.Bytes(), payload)
	if err != nil {
		deliveryFailuresCounter.WithLabelValues("historic_msg_resp").Inc()
		log.Error(
			"[mailserver:DeliverMail] error sending historic message response",
			"err", err,
			"peerID", peerID,
			"requestID", reqID,
		)
	}
}

func (s *mailServer) sendHistoricMessageErrorResponse(peerID, reqID types.Hash, errorToReport error) {
	payload := s.adapter.CreateRequestFailedPayload(reqID, errorToReport)
	err := s.service.SendHistoricMessageResponse(peerID.Bytes(), payload)
	// if we can't report an error, probably something is wrong with p2p connection,
	// so we just print a log entry to document this sad fact
	if err != nil {
		log.Error("Error while reporting error response", "err", err, "peerID", peerID.String())
	}
}

func extractBloomFromEncodedEnvelope(rawValue rlp.RawValue) ([]byte, error) {
	var envelope wakucommon.Envelope
	decodeErr := rlp.DecodeBytes(rawValue, &envelope)
	if decodeErr != nil {
		return nil, decodeErr
	}
	return envelope.Bloom(), nil
}

// checkMsgSignature returns an error in case the message is not correctly signed.
func checkMsgSignature(reqSrc *ecdsa.PublicKey, id []byte) error {
	src := crypto.FromECDSAPub(reqSrc)
	if len(src)-len(id) == 1 {
		src = src[1:]
	}

	// if you want to check the signature, you can do it here. e.g.:
	// if !bytes.Equal(peerID, src) {
	if src == nil {
		return errors.New("wrong signature of p2p request")
	}

	return nil
}
