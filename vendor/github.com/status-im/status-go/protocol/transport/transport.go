package transport

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"database/sql"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/ethereum/go-ethereum/common"
	"github.com/status-im/status-go/connection"
	"github.com/status-im/status-go/eth-node/crypto"
	"github.com/status-im/status-go/eth-node/types"
)

var (
	// ErrNoMailservers returned if there is no configured mailservers that can be used.
	ErrNoMailservers = errors.New("no configured mailservers")
)

type transportKeysManager struct {
	waku types.Waku

	// Identity of the current user.
	privateKey *ecdsa.PrivateKey

	passToSymKeyMutex sync.RWMutex
	passToSymKeyCache map[string]string
}

func (m *transportKeysManager) AddOrGetKeyPair(priv *ecdsa.PrivateKey) (string, error) {
	// caching is handled in waku
	return m.waku.AddKeyPair(priv)
}

func (m *transportKeysManager) AddOrGetSymKeyFromPassword(password string) (string, error) {
	m.passToSymKeyMutex.Lock()
	defer m.passToSymKeyMutex.Unlock()

	if val, ok := m.passToSymKeyCache[password]; ok {
		return val, nil
	}

	id, err := m.waku.AddSymKeyFromPassword(password)
	if err != nil {
		return id, err
	}

	m.passToSymKeyCache[password] = id

	return id, nil
}

func (m *transportKeysManager) RawSymKey(id string) ([]byte, error) {
	return m.waku.GetSymKey(id)
}

type Option func(*Transport) error

// Transport is a transport based on Whisper service.
type Transport struct {
	waku        types.Waku
	api         types.PublicWakuAPI // only PublicWakuAPI implements logic to send messages
	keysManager *transportKeysManager
	filters     *FiltersManager
	logger      *zap.Logger
	cache       *ProcessedMessageIDsCache

	mailservers      []string
	envelopesMonitor *EnvelopesMonitor
	quit             chan struct{}
}

// NewTransport returns a new Transport.
// TODO: leaving a chat should verify that for a given public key
//
//	there are no other chats. It may happen that we leave a private chat
//	but still have a public chat for a given public key.
func NewTransport(
	waku types.Waku,
	privateKey *ecdsa.PrivateKey,
	db *sql.DB,
	sqlitePersistenceTableName string,
	mailservers []string,
	envelopesMonitorConfig *EnvelopesMonitorConfig,
	logger *zap.Logger,
	opts ...Option,
) (*Transport, error) {
	filtersManager, err := NewFiltersManager(newSQLitePersistence(db, sqlitePersistenceTableName), waku, privateKey, logger)
	if err != nil {
		return nil, err
	}

	var envelopesMonitor *EnvelopesMonitor
	if envelopesMonitorConfig != nil {
		envelopesMonitor = NewEnvelopesMonitor(waku, *envelopesMonitorConfig)
		envelopesMonitor.Start()
	}

	var api types.PublicWhisperAPI
	if waku != nil {
		api = waku.PublicWakuAPI()
	}
	t := &Transport{
		waku:             waku,
		api:              api,
		cache:            NewProcessedMessageIDsCache(db),
		envelopesMonitor: envelopesMonitor,
		quit:             make(chan struct{}),
		keysManager: &transportKeysManager{
			waku:              waku,
			privateKey:        privateKey,
			passToSymKeyCache: make(map[string]string),
		},
		filters:     filtersManager,
		mailservers: mailservers,
		logger:      logger.With(zap.Namespace("Transport")),
	}

	for _, opt := range opts {
		if err := opt(t); err != nil {
			return nil, err
		}
	}

	t.cleanFiltersLoop()

	return t, nil
}

func (t *Transport) InitFilters(chatIDs []FiltersToInitialize, publicKeys []*ecdsa.PublicKey) ([]*Filter, error) {
	return t.filters.Init(chatIDs, publicKeys)
}

func (t *Transport) InitPublicFilters(filtersToInit []FiltersToInitialize) ([]*Filter, error) {
	return t.filters.InitPublicFilters(filtersToInit)
}

func (t *Transport) Filters() []*Filter {
	return t.filters.Filters()
}

func (t *Transport) FilterByChatID(chatID string) *Filter {
	return t.filters.FilterByChatID(chatID)
}

func (t *Transport) FilterByTopic(topic []byte) *Filter {
	return t.filters.FilterByTopic(topic)
}

func (t *Transport) FiltersByIdentities(identities []string) []*Filter {
	return t.filters.FiltersByIdentities(identities)
}

func (t *Transport) LoadFilters(filters []*Filter) ([]*Filter, error) {
	return t.filters.InitWithFilters(filters)
}

func (t *Transport) InitCommunityFilters(communityFiltersToInitialize []CommunityFilterToInitialize, useShards bool) ([]*Filter, error) {
	return t.filters.InitCommunityFilters(communityFiltersToInitialize, useShards)
}

func (t *Transport) RemoveFilters(filters []*Filter) error {
	return t.filters.Remove(context.Background(), filters...)
}

func (t *Transport) RemoveFilterByChatID(chatID string) (*Filter, error) {
	return t.filters.RemoveFilterByChatID(chatID)
}

func (t *Transport) ResetFilters(ctx context.Context) error {
	return t.filters.Reset(ctx)
}

func (t *Transport) ProcessNegotiatedSecret(secret types.NegotiatedSecret) (*Filter, error) {
	filter, err := t.filters.LoadNegotiated(secret)
	if err != nil {
		return nil, err
	}
	return filter, nil
}

func (t *Transport) JoinPublic(chatID string) (*Filter, error) {
	return t.filters.LoadPublic(chatID, "")
}

func (t *Transport) LeavePublic(chatID string) error {
	chat := t.filters.Filter(chatID)
	if chat != nil {
		return nil
	}
	return t.filters.Remove(context.Background(), chat)
}

func (t *Transport) JoinPrivate(publicKey *ecdsa.PublicKey) (*Filter, error) {
	return t.filters.LoadContactCode(publicKey)
}

func (t *Transport) JoinGroup(publicKeys []*ecdsa.PublicKey) ([]*Filter, error) {
	var filters []*Filter
	for _, pk := range publicKeys {
		f, err := t.filters.LoadContactCode(pk)
		if err != nil {
			return nil, err
		}
		filters = append(filters, f)

	}
	return filters, nil
}

func (t *Transport) GetStats() types.StatsSummary {
	return t.waku.GetStats()
}

func (t *Transport) RetrieveRawAll() (map[Filter][]*types.Message, error) {
	result := make(map[Filter][]*types.Message)
	logger := t.logger.With(zap.String("site", "retrieveRawAll"))

	for _, filter := range t.filters.Filters() {
		msgs, err := t.api.GetFilterMessages(filter.FilterID)
		if err != nil {
			logger.Warn("failed to fetch messages", zap.Error(err))
			continue
		}
		// Don't pull from filters we don't listen to
		if !filter.Listen {
			for _, msg := range msgs {
				t.waku.MarkP2PMessageAsProcessed(common.BytesToHash(msg.Hash))
			}
			continue
		}

		if len(msgs) == 0 {
			continue
		}

		ids := make([]string, len(msgs))
		for i := range msgs {
			id := types.EncodeHex(msgs[i].Hash)
			ids[i] = id
		}

		hits, err := t.cache.Hits(ids)
		if err != nil {
			logger.Error("failed to check messages exists", zap.Error(err))
			return nil, err
		}

		for i := range msgs {
			// Exclude anything that is a cache hit
			if !hits[types.EncodeHex(msgs[i].Hash)] {
				result[*filter] = append(result[*filter], msgs[i])
				logger.Debug("message not cached", zap.String("hash", types.EncodeHex(msgs[i].Hash)))
			} else {
				logger.Debug("message cached", zap.String("hash", types.EncodeHex(msgs[i].Hash)))
				t.waku.MarkP2PMessageAsProcessed(common.BytesToHash(msgs[i].Hash))
			}
		}

	}

	return result, nil
}

// SendPublic sends a new message using the Whisper service.
// For public filters, chat name is used as an ID as well as
// a topic.
func (t *Transport) SendPublic(ctx context.Context, newMessage *types.NewMessage, chatName string) ([]byte, error) {
	if err := t.addSig(newMessage); err != nil {
		return nil, err
	}

	filter, err := t.filters.LoadPublic(chatName, newMessage.PubsubTopic)
	if err != nil {
		return nil, err
	}

	newMessage.SymKeyID = filter.SymKeyID
	newMessage.Topic = filter.ContentTopic
	newMessage.PubsubTopic = filter.PubsubTopic

	return t.api.Post(ctx, *newMessage)
}

func (t *Transport) SendPrivateWithSharedSecret(ctx context.Context, newMessage *types.NewMessage, publicKey *ecdsa.PublicKey, secret []byte) ([]byte, error) {
	if err := t.addSig(newMessage); err != nil {
		return nil, err
	}

	filter, err := t.filters.LoadNegotiated(types.NegotiatedSecret{
		PublicKey: publicKey,
		Key:       secret,
	})
	if err != nil {
		return nil, err
	}

	newMessage.SymKeyID = filter.SymKeyID
	newMessage.Topic = filter.ContentTopic
	newMessage.PubsubTopic = filter.PubsubTopic
	newMessage.PublicKey = nil

	return t.api.Post(ctx, *newMessage)
}

func (t *Transport) SendPrivateWithPartitioned(ctx context.Context, newMessage *types.NewMessage, publicKey *ecdsa.PublicKey) ([]byte, error) {
	if err := t.addSig(newMessage); err != nil {
		return nil, err
	}

	filter, err := t.filters.LoadPartitioned(publicKey, t.keysManager.privateKey, false)
	if err != nil {
		return nil, err
	}

	newMessage.PubsubTopic = filter.PubsubTopic
	newMessage.Topic = filter.ContentTopic
	newMessage.PublicKey = crypto.FromECDSAPub(publicKey)

	return t.api.Post(ctx, *newMessage)
}

func (t *Transport) SendPrivateOnPersonalTopic(ctx context.Context, newMessage *types.NewMessage, publicKey *ecdsa.PublicKey) ([]byte, error) {
	if err := t.addSig(newMessage); err != nil {
		return nil, err
	}

	filter, err := t.filters.LoadPersonal(publicKey, t.keysManager.privateKey, false)
	if err != nil {
		return nil, err
	}

	newMessage.PubsubTopic = filter.PubsubTopic
	newMessage.Topic = filter.ContentTopic
	newMessage.PublicKey = crypto.FromECDSAPub(publicKey)

	return t.api.Post(ctx, *newMessage)
}

func (t *Transport) PersonalTopicFilter() *Filter {
	return t.filters.PersonalTopicFilter()
}

func (t *Transport) LoadKeyFilters(key *ecdsa.PrivateKey) (*Filter, error) {
	return t.filters.LoadEphemeral(&key.PublicKey, key, true)
}

func (t *Transport) SendCommunityMessage(ctx context.Context, newMessage *types.NewMessage, publicKey *ecdsa.PublicKey) ([]byte, error) {
	if err := t.addSig(newMessage); err != nil {
		return nil, err
	}

	// We load the filter to make sure we can post on it
	filter, err := t.filters.LoadPublic(PubkeyToHex(publicKey)[2:], newMessage.PubsubTopic)
	if err != nil {
		return nil, err
	}

	newMessage.PubsubTopic = filter.PubsubTopic
	newMessage.Topic = filter.ContentTopic
	newMessage.PublicKey = crypto.FromECDSAPub(publicKey)

	t.logger.Debug("SENDING message", zap.Binary("topic", filter.ContentTopic[:]))

	return t.api.Post(ctx, *newMessage)
}

func (t *Transport) cleanFilters() error {
	return t.filters.RemoveNoListenFilters()
}

func (t *Transport) addSig(newMessage *types.NewMessage) error {
	sigID, err := t.keysManager.AddOrGetKeyPair(t.keysManager.privateKey)
	if err != nil {
		return err
	}
	newMessage.SigID = sigID
	return nil
}

func (t *Transport) Track(identifier []byte, hashes [][]byte, newMessages []*types.NewMessage) {
	t.TrackMany([][]byte{identifier}, hashes, newMessages)
}

func (t *Transport) TrackMany(identifiers [][]byte, hashes [][]byte, newMessages []*types.NewMessage) {
	if t.envelopesMonitor == nil {
		return
	}

	envelopeHashes := make([]types.Hash, len(hashes))
	for i, hash := range hashes {
		envelopeHashes[i] = types.BytesToHash(hash)
	}

	err := t.envelopesMonitor.Add(identifiers, envelopeHashes, newMessages)
	if err != nil {
		t.logger.Error("failed to track messages", zap.Error(err))
	}
}

// GetCurrentTime returns the current unix timestamp in milliseconds
func (t *Transport) GetCurrentTime() uint64 {
	return uint64(t.waku.GetCurrentTime().UnixNano() / int64(time.Millisecond))
}

func (t *Transport) MaxMessageSize() uint32 {
	return t.waku.MaxMessageSize()
}

func (t *Transport) Stop() error {
	close(t.quit)
	if t.envelopesMonitor != nil {
		t.envelopesMonitor.Stop()
	}
	return nil
}

// cleanFiltersLoop cleans up the topic we create for the only purpose
// of sending messages.
// Whenever we send a message we also need to listen to that particular topic
// but in case of asymettric topics, we are not interested in listening to them.
// We therefore periodically clean them up so we don't receive unnecessary data.

func (t *Transport) cleanFiltersLoop() {

	ticker := time.NewTicker(5 * time.Minute)
	go func() {
		for {
			select {
			case <-t.quit:
				ticker.Stop()
				return
			case <-ticker.C:
				err := t.cleanFilters()
				if err != nil {
					t.logger.Error("failed to clean up topics", zap.Error(err))
				}
			}
		}
	}()
}

func (t *Transport) WakuVersion() uint {
	return t.waku.Version()
}

func (t *Transport) PeerCount() int {
	return t.waku.PeerCount()
}

func (t *Transport) Peers() map[string]types.WakuV2Peer {
	return t.waku.Peers()
}

func (t *Transport) createMessagesRequestV1(
	ctx context.Context,
	peerID []byte,
	from, to uint32,
	previousCursor []byte,
	topics []types.TopicType,
	waitForResponse bool,
) (cursor []byte, err error) {
	r := createMessagesRequest(from, to, previousCursor, nil, "", topics, 1000)

	events := make(chan types.EnvelopeEvent, 10)
	sub := t.waku.SubscribeEnvelopeEvents(events)
	defer sub.Unsubscribe()

	err = t.waku.SendMessagesRequest(peerID, r)
	if err != nil {
		return
	}

	if !waitForResponse {
		return
	}

	var resp *types.MailServerResponse
	resp, err = t.waitForRequestCompleted(ctx, r.ID, events)
	if err == nil && resp != nil && resp.Error != nil {
		err = resp.Error
	} else if err == nil && resp != nil {
		cursor = resp.Cursor
	}

	return
}

func (t *Transport) createMessagesRequestV2(
	ctx context.Context,
	peerID []byte,
	from, to uint32,
	previousStoreCursor *types.StoreRequestCursor,
	pubsubTopic string,
	contentTopics []types.TopicType,
	limit uint32,
	waitForResponse bool,
	processEnvelopes bool,
) (storeCursor *types.StoreRequestCursor, envelopesCount int, err error) {
	r := createMessagesRequest(from, to, nil, previousStoreCursor, pubsubTopic, contentTopics, limit)

	if waitForResponse {
		resultCh := make(chan struct {
			storeCursor    *types.StoreRequestCursor
			envelopesCount int
			err            error
		})

		go func() {
			storeCursor, envelopesCount, err = t.waku.RequestStoreMessages(ctx, peerID, r, processEnvelopes)
			resultCh <- struct {
				storeCursor    *types.StoreRequestCursor
				envelopesCount int
				err            error
			}{storeCursor, envelopesCount, err}
		}()

		select {
		case result := <-resultCh:
			return result.storeCursor, result.envelopesCount, result.err
		case <-ctx.Done():
			return nil, 0, ctx.Err()
		}
	} else {
		go func() {
			_, _, err = t.waku.RequestStoreMessages(ctx, peerID, r, false)
			if err != nil {
				t.logger.Error("failed to request store messages", zap.Error(err))
			}
		}()
	}

	return
}

func (t *Transport) SendMessagesRequestForTopics(
	ctx context.Context,
	peerID []byte,
	from, to uint32,
	previousCursor []byte,
	previousStoreCursor *types.StoreRequestCursor,
	pubsubTopic string,
	contentTopics []types.TopicType,
	limit uint32,
	waitForResponse bool,
	processEnvelopes bool,
) (cursor []byte, storeCursor *types.StoreRequestCursor, envelopesCount int, err error) {
	switch t.waku.Version() {
	case 2:
		storeCursor, envelopesCount, err = t.createMessagesRequestV2(ctx, peerID, from, to, previousStoreCursor, pubsubTopic, contentTopics, limit, waitForResponse, processEnvelopes)
	case 1:
		cursor, err = t.createMessagesRequestV1(ctx, peerID, from, to, previousCursor, contentTopics, waitForResponse)
	default:
		err = fmt.Errorf("unsupported version %d", t.waku.Version())
	}
	return
}

func createMessagesRequest(from, to uint32, cursor []byte, storeCursor *types.StoreRequestCursor, pubsubTopic string, topics []types.TopicType, limit uint32) types.MessagesRequest {
	aUUID := uuid.New()
	// uuid is 16 bytes, converted to hex it's 32 bytes as expected by types.MessagesRequest
	id := []byte(hex.EncodeToString(aUUID[:]))
	var topicBytes [][]byte
	for idx := range topics {
		topicBytes = append(topicBytes, topics[idx][:])
	}
	return types.MessagesRequest{
		ID:            id,
		From:          from,
		To:            to,
		Limit:         limit,
		Cursor:        cursor,
		PubsubTopic:   pubsubTopic,
		ContentTopics: topicBytes,
		StoreCursor:   storeCursor,
	}
}

func (t *Transport) waitForRequestCompleted(ctx context.Context, requestID []byte, events chan types.EnvelopeEvent) (*types.MailServerResponse, error) {
	for {
		select {
		case ev := <-events:
			if !bytes.Equal(ev.Hash.Bytes(), requestID) {
				continue
			}
			if ev.Event != types.EventMailServerRequestCompleted {
				continue
			}
			data, ok := ev.Data.(*types.MailServerResponse)
			if ok {
				return data, nil
			}
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
}

// ConfirmMessagesProcessed marks the messages as processed in the cache so
// they won't be passed to the next layer anymore
func (t *Transport) ConfirmMessagesProcessed(ids []string, timestamp uint64) error {
	t.logger.Debug("confirming message processed", zap.Any("ids", ids), zap.Any("timestamp", timestamp))
	return t.cache.Add(ids, timestamp)
}

// CleanMessagesProcessed clears the messages that are older than timestamp
func (t *Transport) CleanMessagesProcessed(timestamp uint64) error {
	return t.cache.Clean(timestamp)
}

func (t *Transport) SetEnvelopeEventsHandler(handler EnvelopeEventsHandler) error {
	if t.envelopesMonitor == nil {
		return errors.New("Current transport has no envelopes monitor")
	}
	t.envelopesMonitor.handler = handler
	return nil
}

func (t *Transport) ClearProcessedMessageIDsCache() error {
	t.logger.Debug("clearing processed messages cache")
	t.waku.ClearEnvelopesCache()
	return t.cache.Clear()
}

func (t *Transport) BloomFilter() []byte {
	return t.api.BloomFilter()
}

func PubkeyToHex(key *ecdsa.PublicKey) string {
	return types.EncodeHex(crypto.FromECDSAPub(key))
}

func (t *Transport) StartDiscV5() error {
	return t.waku.StartDiscV5()
}

func (t *Transport) StopDiscV5() error {
	return t.waku.StopDiscV5()
}

func (t *Transport) ListenAddresses() ([]string, error) {
	return t.waku.ListenAddresses()
}

func (t *Transport) AddStorePeer(address string) (peer.ID, error) {
	return t.waku.AddStorePeer(address)
}

func (t *Transport) AddRelayPeer(address string) (peer.ID, error) {
	return t.waku.AddRelayPeer(address)
}

func (t *Transport) DialPeer(address string) error {
	return t.waku.DialPeer(address)
}

func (t *Transport) DialPeerByID(peerID string) error {
	return t.waku.DialPeerByID(peerID)
}

func (t *Transport) DropPeer(peerID string) error {
	return t.waku.DropPeer(peerID)
}

func (t *Transport) ProcessingP2PMessages() bool {
	return t.waku.ProcessingP2PMessages()
}

func (t *Transport) MarkP2PMessageAsProcessed(hash common.Hash) {
	t.waku.MarkP2PMessageAsProcessed(hash)
}

func (t *Transport) SubscribeToConnStatusChanges() (*types.ConnStatusSubscription, error) {
	return t.waku.SubscribeToConnStatusChanges()
}

func (t *Transport) ConnectionChanged(state connection.State) {
	t.waku.ConnectionChanged(state)
}

// Subscribe to a pubsub topic, passing an optional public key if the pubsub topic is protected
func (t *Transport) SubscribeToPubsubTopic(topic string, optPublicKey *ecdsa.PublicKey) error {
	if t.waku.Version() == 2 {
		return t.waku.SubscribeToPubsubTopic(topic, optPublicKey)
	}
	return nil
}

// Unsubscribe from a pubsub topic
func (t *Transport) UnsubscribeFromPubsubTopic(topic string) error {
	if t.waku.Version() == 2 {
		return t.waku.UnsubscribeFromPubsubTopic(topic)
	}
	return nil
}

func (t *Transport) StorePubsubTopicKey(topic string, privKey *ecdsa.PrivateKey) error {
	return t.waku.StorePubsubTopicKey(topic, privKey)
}

func (t *Transport) RetrievePubsubTopicKey(topic string) (*ecdsa.PrivateKey, error) {
	return t.waku.RetrievePubsubTopicKey(topic)
}

func (t *Transport) RemovePubsubTopicKey(topic string) error {
	if t.waku.Version() == 2 {
		return t.waku.RemovePubsubTopicKey(topic)
	}
	return nil
}
