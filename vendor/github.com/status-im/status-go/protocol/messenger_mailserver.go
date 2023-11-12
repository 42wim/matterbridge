package protocol

import (
	"context"
	"fmt"
	"math"
	"sort"
	"sync"
	"time"

	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/status-im/status-go/connection"
	"github.com/status-im/status-go/eth-node/crypto"
	"github.com/status-im/status-go/eth-node/types"
	"github.com/status-im/status-go/protocol/common"
	"github.com/status-im/status-go/protocol/protobuf"
	"github.com/status-im/status-go/protocol/transport"
	"github.com/status-im/status-go/services/mailservers"
)

const (
	initialStoreNodeRequestPageSize = 4
	defaultStoreNodeRequestPageSize = 20

	// tolerance is how many seconds of potentially out-of-order messages we want to fetch
	tolerance uint32 = 60

	mailserverRequestTimeout         = 30 * time.Second
	mailserverMaxTries          uint = 2
	mailserverMaxFailedRequests uint = 2

	oneDayDuration   = 24 * time.Hour
	oneMonthDuration = 31 * oneDayDuration
)

// maxTopicsPerRequest sets the batch size to limit the number of topics per store query
var maxTopicsPerRequest int = 10

var ErrNoFiltersForChat = errors.New("no filter registered for given chat")

func (m *Messenger) shouldSync() (bool, error) {
	if m.mailserverCycle.activeMailserver == nil || !m.Online() {
		return false, nil
	}

	useMailserver, err := m.settings.CanUseMailservers()
	if err != nil {
		m.logger.Error("failed to get use mailservers", zap.Error(err))
		return false, err
	}

	return useMailserver, nil
}

func (m *Messenger) scheduleSyncChat(chat *Chat) (bool, error) {
	shouldSync, err := m.shouldSync()
	if err != nil {
		m.logger.Error("failed to get should sync", zap.Error(err))
		return false, err
	}

	if !shouldSync {
		return false, nil
	}

	go func() {
		_, err := m.performMailserverRequest(func() (*MessengerResponse, error) {
			response, err := m.syncChatWithFilters(chat.ID)

			if err != nil {
				m.logger.Error("failed to sync chat", zap.Error(err))
				return nil, err
			}

			if m.config.messengerSignalsHandler != nil {
				m.config.messengerSignalsHandler.MessengerResponse(response)
			}
			return response, nil
		})
		if err != nil {
			m.logger.Error("failed to perform mailserver request", zap.Error(err))
		}
	}()
	return true, nil
}

func (m *Messenger) connectToNewMailserverAndWait() error {
	// Handle pinned mailservers
	m.logger.Info("disconnecting mailserver")
	pinnedMailserver, err := m.getPinnedMailserver()
	if err != nil {
		m.logger.Error("could not obtain the pinned mailserver", zap.Error(err))
		return err
	}
	// If pinned mailserver is not nil, no need to disconnect and wait for it to be available
	if pinnedMailserver == nil {
		m.disconnectActiveMailserver()
	}

	return m.findNewMailserver()
}

func (m *Messenger) performMailserverRequest(fn func() (*MessengerResponse, error)) (*MessengerResponse, error) {

	m.mailserverCycle.RLock()
	defer m.mailserverCycle.RUnlock()
	var tries uint = 0
	for tries < mailserverMaxTries {
		if !m.isActiveMailserverAvailable() {
			return nil, errors.New("mailserver not available")
		}

		m.logger.Info("trying performing mailserver requests", zap.Uint("try", tries))
		activeMailserver := m.getActiveMailserver()
		// Make sure we are connected to a mailserver
		if activeMailserver == nil {
			return nil, errors.New("mailserver not available")
		}

		// Peform request
		response, err := fn()
		if err == nil {
			// Reset failed requests
			m.logger.Debug("mailserver request performed successfully",
				zap.String("mailserverID", activeMailserver.ID))
			activeMailserver.FailedRequests = 0
			return response, nil
		}

		m.logger.Error("failed to perform mailserver request",
			zap.String("mailserverID", activeMailserver.ID),
			zap.Uint("tries", tries),
			zap.Error(err),
		)

		tries++
		// Increment failed requests
		activeMailserver.FailedRequests++

		// Change mailserver
		if activeMailserver.FailedRequests >= mailserverMaxFailedRequests {
			return nil, errors.New("too many failed requests")
		}
		// Wait a couple of second not to spam
		time.Sleep(2 * time.Second)

	}
	return nil, errors.New("failed to perform mailserver request")
}

func (m *Messenger) scheduleSyncFilters(filters []*transport.Filter) (bool, error) {
	shouldSync, err := m.shouldSync()
	if err != nil {
		m.logger.Error("failed to get shouldSync", zap.Error(err))
		return false, err
	}

	if !shouldSync {
		return false, nil
	}

	go func() {
		_, err := m.performMailserverRequest(func() (*MessengerResponse, error) {
			response, err := m.syncFilters(filters)

			if err != nil {
				m.logger.Error("failed to sync filter", zap.Error(err))
				return nil, err
			}

			if m.config.messengerSignalsHandler != nil {
				m.config.messengerSignalsHandler.MessengerResponse(response)
			}
			return response, nil
		})
		if err != nil {
			m.logger.Error("failed to perform mailserver request", zap.Error(err))
		}

	}()
	return true, nil
}

func (m *Messenger) calculateMailserverTo() uint32 {
	seconds := float64(m.GetCurrentTimeInMillis()) / 1000
	return uint32(math.Ceil(seconds))
}

func (m *Messenger) calculateMailserverTimeBounds(duration time.Duration) (uint32, uint32) {
	now := float64(m.GetCurrentTimeInMillis()) / 1000
	to := uint32(math.Ceil(now))
	from := uint32(math.Floor(now)) - uint32(duration.Seconds())
	return from, to
}

func (m *Messenger) filtersForChat(chatID string) ([]*transport.Filter, error) {
	chat, ok := m.allChats.Load(chatID)
	if !ok {
		return nil, ErrChatNotFound
	}
	var filters []*transport.Filter

	if chat.OneToOne() {
		// We sync our own topic and any eventual negotiated
		publicKeys := []string{common.PubkeyToHex(&m.identity.PublicKey), chatID}

		filters = m.transport.FiltersByIdentities(publicKeys)

	} else if chat.PrivateGroupChat() {
		var publicKeys []string
		for _, m := range chat.Members {
			publicKeys = append(publicKeys, m.ID)
		}

		filters = m.transport.FiltersByIdentities(publicKeys)

	} else {
		filter := m.transport.FilterByChatID(chatID)
		if filter == nil {
			return nil, ErrNoFiltersForChat
		}
		filters = []*transport.Filter{filter}
	}

	return filters, nil
}

func (m *Messenger) topicsForChat(chatID string) (string, []types.TopicType, error) {
	filters, err := m.filtersForChat(chatID)
	if err != nil {
		return "", nil, err
	}

	var contentTopics []types.TopicType

	for _, filter := range filters {
		contentTopics = append(contentTopics, filter.ContentTopic)
	}

	return filters[0].PubsubTopic, contentTopics, nil
}

func (m *Messenger) syncChatWithFilters(chatID string) (*MessengerResponse, error) {
	filters, err := m.filtersForChat(chatID)
	if err != nil {
		return nil, err
	}
	return m.syncFilters(filters)
}

func (m *Messenger) syncBackup() error {

	filter := m.transport.PersonalTopicFilter()
	if filter == nil {
		return errors.New("personal topic filter not loaded")
	}

	from, to := m.calculateMailserverTimeBounds(oneMonthDuration)

	batch := MailserverBatch{From: from, To: to, Topics: []types.TopicType{filter.ContentTopic}}
	err := m.processMailserverBatch(batch)
	if err != nil {
		return err
	}
	return m.settings.SetBackupFetched(true)
}

func (m *Messenger) defaultSyncPeriodFromNow() (uint32, error) {
	defaultSyncPeriod, err := m.settings.GetDefaultSyncPeriod()
	if err != nil {
		return 0, err
	}
	return uint32(m.getTimesource().GetCurrentTime()/1000) - defaultSyncPeriod, nil
}

// capToDefaultSyncPeriod caps the sync period to the default
func (m *Messenger) capToDefaultSyncPeriod(period uint32) (uint32, error) {
	d, err := m.defaultSyncPeriodFromNow()
	if err != nil {
		return 0, err
	}
	if d > period {
		return d, nil
	}
	return period - tolerance, nil
}

func (m *Messenger) updateFiltersPriority(filters []*transport.Filter) {
	for _, filter := range filters {
		chatID := filter.ChatID
		chat := m.Chat(chatID)
		if chat != nil {
			filter.Priority = chat.ReadMessagesAtClockValue
		}
	}
}

func (m *Messenger) resetFiltersPriority(filters []*transport.Filter) {
	for _, filter := range filters {
		filter.Priority = 0
	}
}

func (m *Messenger) RequestAllHistoricMessagesWithRetries(forceFetchingBackup bool) (*MessengerResponse, error) {
	return m.performMailserverRequest(func() (*MessengerResponse, error) {
		return m.RequestAllHistoricMessages(forceFetchingBackup)
	})
}

// RequestAllHistoricMessages requests all the historic messages for any topic
func (m *Messenger) RequestAllHistoricMessages(forceFetchingBackup bool) (*MessengerResponse, error) {
	shouldSync, err := m.shouldSync()
	if err != nil {
		return nil, err
	}

	if !shouldSync {
		return nil, nil
	}

	backupFetched, err := m.settings.BackupFetched()
	if err != nil {
		return nil, err
	}

	if forceFetchingBackup || !backupFetched {
		m.logger.Info("fetching backup")
		err := m.syncBackup()
		if err != nil {
			return nil, err
		}
		m.logger.Info("backup fetched")
	}

	filters := m.transport.Filters()
	m.updateFiltersPriority(filters)
	defer m.resetFiltersPriority(filters)
	response, err := m.syncFilters(filters)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func getPrioritizedBatches() []int {
	return []int{1, 5, 10}
}

func (m *Messenger) syncFiltersFrom(filters []*transport.Filter, lastRequest uint32) (*MessengerResponse, error) {
	response := &MessengerResponse{}
	topicInfo, err := m.mailserversDatabase.Topics()
	if err != nil {
		return nil, err
	}

	topicsData := make(map[string]mailservers.MailserverTopic)
	for _, topic := range topicInfo {
		topicsData[fmt.Sprintf("%s-%s", topic.PubsubTopic, topic.ContentTopic)] = topic
	}

	batches := make(map[string]map[int]MailserverBatch)

	to := m.calculateMailserverTo()
	var syncedTopics []mailservers.MailserverTopic

	sort.Slice(filters[:], func(i, j int) bool {
		p1 := filters[i].Priority
		p2 := filters[j].Priority
		return p1 > p2
	})
	prioritizedBatches := getPrioritizedBatches()
	currentBatch := 0

	if len(filters) == 0 || filters[0].Priority == 0 {
		currentBatch = len(prioritizedBatches)
	}

	defaultPeriodFromNow, err := m.defaultSyncPeriodFromNow()
	if err != nil {
		return nil, err
	}

	contentTopicsPerPubsubTopic := make(map[string]map[string]*transport.Filter)
	for _, filter := range filters {
		if !filter.Listen || filter.Ephemeral {
			continue
		}

		contentTopics, ok := contentTopicsPerPubsubTopic[filter.PubsubTopic]
		if !ok {
			contentTopics = make(map[string]*transport.Filter)
		}
		contentTopics[filter.ContentTopic.String()] = filter
		contentTopicsPerPubsubTopic[filter.PubsubTopic] = contentTopics
	}

	for pubsubTopic, contentTopics := range contentTopicsPerPubsubTopic {
		if _, ok := batches[pubsubTopic]; !ok {
			batches[pubsubTopic] = make(map[int]MailserverBatch)
		}

		for _, filter := range contentTopics {
			var chatID string
			// If the filter has an identity, we use it as a chatID, otherwise is a public chat/community chat filter
			if len(filter.Identity) != 0 {
				chatID = filter.Identity
			} else {
				chatID = filter.ChatID
			}

			topicData, ok := topicsData[filter.PubsubTopic+filter.ContentTopic.String()]
			var capToDefaultSyncPeriod = true
			if !ok {
				if lastRequest == 0 {
					lastRequest = defaultPeriodFromNow
				}
				topicData = mailservers.MailserverTopic{
					PubsubTopic:  filter.PubsubTopic,
					ContentTopic: filter.ContentTopic.String(),
					LastRequest:  int(defaultPeriodFromNow),
				}
			} else if lastRequest != 0 {
				topicData.LastRequest = int(lastRequest)
				capToDefaultSyncPeriod = false
			}

			batchID := topicData.LastRequest

			if currentBatch < len(prioritizedBatches) {
				batch, ok := batches[pubsubTopic][currentBatch]
				if ok {
					prevTopicData, ok := topicsData[batch.PubsubTopic+batch.Topics[0].String()]
					if (!ok && topicData.LastRequest != int(defaultPeriodFromNow)) ||
						(ok && prevTopicData.LastRequest != topicData.LastRequest) {
						currentBatch++
					}
				}
				if currentBatch < len(prioritizedBatches) {
					batchID = currentBatch
					currentBatchCap := prioritizedBatches[currentBatch] - 1
					if currentBatchCap == 0 {
						currentBatch++
					} else {
						prioritizedBatches[currentBatch] = currentBatchCap
					}
				}
			}

			batch, ok := batches[pubsubTopic][batchID]
			if !ok {
				from := uint32(topicData.LastRequest)
				if capToDefaultSyncPeriod {
					from, err = m.capToDefaultSyncPeriod(uint32(topicData.LastRequest))
					if err != nil {
						return nil, err
					}
				}
				batch = MailserverBatch{From: from, To: to}
			}

			batch.ChatIDs = append(batch.ChatIDs, chatID)
			batch.PubsubTopic = pubsubTopic
			batch.Topics = append(batch.Topics, filter.ContentTopic)
			batches[pubsubTopic][batchID] = batch

			// Set last request to the new `to`
			topicData.LastRequest = int(to)
			syncedTopics = append(syncedTopics, topicData)
		}
	}

	if m.config.messengerSignalsHandler != nil {
		m.config.messengerSignalsHandler.HistoryRequestStarted(len(batches))
	}

	var batches24h []MailserverBatch
	for pubsubTopic := range batches {
		batchKeys := make([]int, 0, len(batches[pubsubTopic]))
		for k := range batches[pubsubTopic] {
			batchKeys = append(batchKeys, k)
		}
		sort.Ints(batchKeys)

		keysToIterate := append([]int{}, batchKeys...)
		for {
			// For all batches
			var tmpKeysToIterate []int
			for _, k := range keysToIterate {
				batch := batches[pubsubTopic][k]

				dayBatch := MailserverBatch{
					To:          batch.To,
					Cursor:      batch.Cursor,
					PubsubTopic: batch.PubsubTopic,
					Topics:      batch.Topics,
					ChatIDs:     batch.ChatIDs,
				}

				from := batch.To - uint32(oneDayDuration.Seconds())
				if from > batch.From {
					dayBatch.From = from
					batches24h = append(batches24h, dayBatch)

					// Replace og batch with new dates
					batch.To = from
					batches[pubsubTopic][k] = batch
					tmpKeysToIterate = append(tmpKeysToIterate, k)
				} else {
					batches24h = append(batches24h, batch)
				}
			}

			if len(tmpKeysToIterate) == 0 {
				break
			}
			keysToIterate = tmpKeysToIterate
		}
	}

	i := 0
	for _, batch := range batches24h {
		i++
		err := m.processMailserverBatch(batch)
		if err != nil {
			m.logger.Error("error syncing topics", zap.Error(err))
			return nil, err
		}
	}

	m.logger.Debug("topics synced")
	if m.config.messengerSignalsHandler != nil {
		m.config.messengerSignalsHandler.HistoryRequestCompleted()
	}

	err = m.mailserversDatabase.AddTopics(syncedTopics)
	if err != nil {
		return nil, err
	}

	var messagesToBeSaved []*common.Message
	for _, batches := range batches {
		for _, batch := range batches {
			for _, id := range batch.ChatIDs {
				chat, ok := m.allChats.Load(id)
				if !ok || !chat.Active || chat.Timeline() || chat.ProfileUpdates() {
					continue
				}
				gap, err := m.calculateGapForChat(chat, batch.From)
				if err != nil {
					return nil, err
				}
				if chat.SyncedFrom == 0 || chat.SyncedFrom > batch.From {
					chat.SyncedFrom = batch.From
				}

				chat.SyncedTo = to

				err = m.persistence.SetSyncTimestamps(chat.SyncedFrom, chat.SyncedTo, chat.ID)
				if err != nil {
					return nil, err
				}

				response.AddChat(chat)
				if gap != nil {
					response.AddMessage(gap)
					messagesToBeSaved = append(messagesToBeSaved, gap)
				}
			}
		}
	}

	if len(messagesToBeSaved) > 0 {
		err := m.persistence.SaveMessages(messagesToBeSaved)
		if err != nil {
			return nil, err
		}
	}
	return response, nil
}

func (m *Messenger) syncFilters(filters []*transport.Filter) (*MessengerResponse, error) {
	return m.syncFiltersFrom(filters, 0)
}

func (m *Messenger) calculateGapForChat(chat *Chat, from uint32) (*common.Message, error) {
	// Chat was never synced, no gap necessary
	if chat.SyncedTo == 0 {
		return nil, nil
	}

	// If we filled the gap, nothing to do
	if chat.SyncedTo >= from {
		return nil, nil
	}

	timestamp := m.getTimesource().GetCurrentTime()

	message := &common.Message{
		ChatMessage: &protobuf.ChatMessage{
			ChatId:      chat.ID,
			Text:        "Gap message",
			MessageType: protobuf.MessageType_SYSTEM_MESSAGE_GAP,
			ContentType: protobuf.ChatMessage_SYSTEM_MESSAGE_GAP,
			Clock:       uint64(from) * 1000,
			Timestamp:   timestamp,
		},
		GapParameters: &common.GapParameters{
			From: chat.SyncedTo,
			To:   from,
		},
		From:             common.PubkeyToHex(&m.identity.PublicKey),
		WhisperTimestamp: timestamp,
		LocalChatID:      chat.ID,
		Seen:             true,
		ID:               types.EncodeHex(crypto.Keccak256([]byte(fmt.Sprintf("%s-%d-%d", chat.ID, chat.SyncedTo, from)))),
	}

	return message, m.persistence.SaveMessages([]*common.Message{message})
}

type work struct {
	pubsubTopic   string
	contentTopics []types.TopicType
	cursor        []byte
	storeCursor   *types.StoreRequestCursor
	limit         uint32
}

type messageRequester interface {
	SendMessagesRequestForTopics(
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
	) (cursor []byte, storeCursor *types.StoreRequestCursor, envelopesCount int, err error)
}

func processMailserverBatch(
	ctx context.Context,
	messageRequester messageRequester,
	batch MailserverBatch,
	mailserverID []byte,
	logger *zap.Logger,
	pageLimit uint32,
	shouldProcessNextPage func(int) (bool, uint32),
	processEnvelopes bool,
) error {

	var topicStrings []string
	for _, t := range batch.Topics {
		topicStrings = append(topicStrings, t.String())
	}
	logger = logger.With(zap.Any("chatIDs", batch.ChatIDs),
		zap.String("fromString", time.Unix(int64(batch.From), 0).Format(time.RFC3339)),
		zap.String("toString", time.Unix(int64(batch.To), 0).Format(time.RFC3339)),
		zap.Any("topic", topicStrings),
		zap.Int64("from", int64(batch.From)),
		zap.Int64("to", int64(batch.To)))

	logger.Info("syncing topic")

	wg := sync.WaitGroup{}
	workWg := sync.WaitGroup{}
	workCh := make(chan work, 1000)       // each batch item is split in 10 topics bunch and sent to this channel
	workCompleteCh := make(chan struct{}) // once all batch items are processed, this channel is triggered
	semaphore := make(chan int, 3)        // limit the number of concurrent queries
	errCh := make(chan error)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Producer
	wg.Add(1)
	go func() {
		defer func() {
			logger.Debug("mailserver batch producer complete")
			wg.Done()
		}()

		allWorks := int(math.Ceil(float64(len(batch.Topics)) / float64(maxTopicsPerRequest)))
		workWg.Add(allWorks)

		for i := 0; i < len(batch.Topics); i += maxTopicsPerRequest {
			j := i + maxTopicsPerRequest
			if j > len(batch.Topics) {
				j = len(batch.Topics)
			}

			select {
			case <-ctx.Done():
				logger.Debug("processBatch producer - context done")
				return
			default:
				logger.Debug("processBatch producer - creating work")
				workCh <- work{
					pubsubTopic:   batch.PubsubTopic,
					contentTopics: batch.Topics[i:j],
					limit:         pageLimit,
				}
				time.Sleep(50 * time.Millisecond)
			}
		}

		go func() {
			workWg.Wait()
			workCompleteCh <- struct{}{}
		}()

		logger.Debug("processBatch producer complete")
	}()

	var result error

loop:
	for {
		select {
		case <-ctx.Done():
			logger.Debug("processBatch cleanup - context done")
			result = ctx.Err()
			if errors.Is(result, context.Canceled) {
				result = nil
			}
			break loop
		case w, ok := <-workCh:
			if !ok {
				continue
			}

			logger.Debug("processBatch - received work")
			semaphore <- 1
			go func(w work) { // Consumer
				defer func() {
					workWg.Done()
					<-semaphore
				}()

				queryCtx, queryCancel := context.WithTimeout(ctx, mailserverRequestTimeout)
				cursor, storeCursor, envelopesCount, err := messageRequester.SendMessagesRequestForTopics(queryCtx, mailserverID, batch.From, batch.To, w.cursor, w.storeCursor, w.pubsubTopic, w.contentTopics, w.limit, true, processEnvelopes)

				queryCancel()

				if err != nil {
					logger.Debug("failed to send request", zap.Error(err))
					errCh <- err
					return
				}

				processNextPage := true
				nextPageLimit := pageLimit

				if shouldProcessNextPage != nil {
					processNextPage, nextPageLimit = shouldProcessNextPage(envelopesCount)
				}

				if !processNextPage {
					return
				}

				// Check the cursor after calling `shouldProcessNextPage`.
				// The app might use process the fetched envelopes in the callback for own needs.
				if len(cursor) == 0 && storeCursor == nil {
					return
				}

				logger.Debug("processBatch producer - creating work (cursor)")

				workWg.Add(1)
				workCh <- work{
					pubsubTopic:   w.pubsubTopic,
					contentTopics: w.contentTopics,
					cursor:        cursor,
					storeCursor:   storeCursor,
					limit:         nextPageLimit,
				}
			}(w)
		case err := <-errCh:
			logger.Debug("processBatch - received error", zap.Error(err))
			cancel() // Kill go routines
			return err
		case <-workCompleteCh:
			logger.Debug("processBatch - all jobs complete")
			cancel() // Kill go routines
		}
	}

	wg.Wait()

	// NOTE(camellos): Disabling for now, not critical and I'd rather take a bit more time
	// to test it
	//logger.Info("waiting until message processed")
	//m.waitUntilP2PMessagesProcessed()

	logger.Info("synced topic", zap.NamedError("hasError", result))
	return result
}

func (m *Messenger) processMailserverBatch(batch MailserverBatch) error {
	if m.featureFlags.StoreNodesDisabled {
		return nil
	}

	mailserverID, err := m.activeMailserverID()
	if err != nil {
		return err
	}

	return processMailserverBatch(m.ctx, m.transport, batch, mailserverID, m.logger, defaultStoreNodeRequestPageSize, nil, false)
}

func (m *Messenger) processMailserverBatchWithOptions(batch MailserverBatch, pageLimit uint32, shouldProcessNextPage func(int) (bool, uint32), processEnvelopes bool) error {
	if m.featureFlags.StoreNodesDisabled {
		return nil
	}

	mailserverID, err := m.activeMailserverID()
	if err != nil {
		return err
	}

	return processMailserverBatch(m.ctx, m.transport, batch, mailserverID, m.logger, pageLimit, shouldProcessNextPage, processEnvelopes)
}

type MailserverBatch struct {
	From        uint32
	To          uint32
	Cursor      string
	PubsubTopic string
	Topics      []types.TopicType
	ChatIDs     []string
}

func (m *Messenger) SyncChatFromSyncedFrom(chatID string) (uint32, error) {
	var from uint32
	_, err := m.performMailserverRequest(func() (*MessengerResponse, error) {
		pubsubTopic, topics, err := m.topicsForChat(chatID)
		if err != nil {
			return nil, nil
		}

		chat, ok := m.allChats.Load(chatID)
		if !ok {
			return nil, ErrChatNotFound
		}

		defaultSyncPeriod, err := m.settings.GetDefaultSyncPeriod()
		if err != nil {
			return nil, err
		}

		batch := MailserverBatch{
			ChatIDs:     []string{chatID},
			To:          chat.SyncedFrom,
			From:        chat.SyncedFrom - defaultSyncPeriod,
			PubsubTopic: pubsubTopic,
			Topics:      topics,
		}
		if m.config.messengerSignalsHandler != nil {
			m.config.messengerSignalsHandler.HistoryRequestStarted(1)
		}

		err = m.processMailserverBatch(batch)
		if err != nil {
			return nil, err
		}

		if m.config.messengerSignalsHandler != nil {
			m.config.messengerSignalsHandler.HistoryRequestCompleted()
		}
		if chat.SyncedFrom == 0 || chat.SyncedFrom > batch.From {
			chat.SyncedFrom = batch.From
		}

		m.logger.Debug("setting sync timestamps", zap.Int64("from", int64(batch.From)), zap.Int64("to", int64(chat.SyncedTo)), zap.String("chatID", chatID))

		err = m.persistence.SetSyncTimestamps(batch.From, chat.SyncedTo, chat.ID)
		from = batch.From
		return nil, err
	})
	if err != nil {
		return 0, err
	}

	return from, nil
}

func (m *Messenger) FillGaps(chatID string, messageIDs []string) error {
	messages, err := m.persistence.MessagesByIDs(messageIDs)
	if err != nil {
		return err
	}

	_, ok := m.allChats.Load(chatID)
	if !ok {
		return errors.New("chat not existing")
	}

	pubsubTopic, topics, err := m.topicsForChat(chatID)
	if err != nil {
		return err
	}

	var lowestFrom, highestTo uint32

	for _, message := range messages {
		if message.GapParameters == nil {
			return errors.New("can't sync non-gap message")
		}

		if lowestFrom == 0 || lowestFrom > message.GapParameters.From {
			lowestFrom = message.GapParameters.From
		}

		if highestTo < message.GapParameters.To {
			highestTo = message.GapParameters.To
		}
	}

	batch := MailserverBatch{
		ChatIDs:     []string{chatID},
		To:          highestTo,
		From:        lowestFrom,
		PubsubTopic: pubsubTopic,
		Topics:      topics,
	}

	if m.config.messengerSignalsHandler != nil {
		m.config.messengerSignalsHandler.HistoryRequestStarted(1)
	}

	err = m.processMailserverBatch(batch)
	if err != nil {
		return err
	}

	if m.config.messengerSignalsHandler != nil {
		m.config.messengerSignalsHandler.HistoryRequestCompleted()
	}

	return m.persistence.DeleteMessages(messageIDs)
}

func (m *Messenger) waitUntilP2PMessagesProcessed() { // nolint: unused

	ticker := time.NewTicker(50 * time.Millisecond)

	for { //nolint: gosimple
		select {
		case <-ticker.C:
			if !m.transport.ProcessingP2PMessages() {
				ticker.Stop()
				return
			}
		}
	}
}

func (m *Messenger) LoadFilters(filters []*transport.Filter) ([]*transport.Filter, error) {
	return m.transport.LoadFilters(filters)
}

func (m *Messenger) ToggleUseMailservers(value bool) error {
	m.mailserverCycle.Lock()
	defer m.mailserverCycle.Unlock()

	err := m.settings.SetUseMailservers(value)
	if err != nil {
		return err
	}

	if value {
		m.cycleMailservers()
		return nil
	}

	m.disconnectActiveMailserver()
	return nil
}

func (m *Messenger) SetPinnedMailservers(mailservers map[string]string) error {
	err := m.settings.SetPinnedMailservers(mailservers)
	if err != nil {
		return err
	}

	m.cycleMailservers()
	return nil
}

func (m *Messenger) RemoveFilters(filters []*transport.Filter) error {
	return m.transport.RemoveFilters(filters)
}

func (m *Messenger) ConnectionChanged(state connection.State) {
	m.transport.ConnectionChanged(state)
	if !m.connectionState.Offline && state.Offline {
		m.sender.StopDatasync()
	}

	if m.connectionState.Offline && !state.Offline {
		err := m.sender.StartDatasync(m.sendDataSync)
		if err != nil {
			m.logger.Error("failed to start datasync", zap.Error(err))
		}
	}

	m.connectionState = state
}

func (m *Messenger) fetchMessages(chatID string, duration time.Duration) (uint32, error) {
	from, to := m.calculateMailserverTimeBounds(duration)

	_, err := m.performMailserverRequest(func() (*MessengerResponse, error) {
		pubsubTopic, topics, err := m.topicsForChat(chatID)
		if err != nil {
			return nil, nil
		}

		chat, ok := m.allChats.Load(chatID)
		if !ok {
			return nil, ErrChatNotFound
		}

		batch := MailserverBatch{
			ChatIDs:     []string{chatID},
			From:        from,
			To:          to,
			PubsubTopic: pubsubTopic,
			Topics:      topics,
		}
		if m.config.messengerSignalsHandler != nil {
			m.config.messengerSignalsHandler.HistoryRequestStarted(1)
		}

		err = m.processMailserverBatch(batch)
		if err != nil {
			return nil, err
		}

		if m.config.messengerSignalsHandler != nil {
			m.config.messengerSignalsHandler.HistoryRequestCompleted()
		}
		if chat.SyncedFrom == 0 || chat.SyncedFrom > batch.From {
			chat.SyncedFrom = batch.From
		}

		m.logger.Debug("setting sync timestamps", zap.Int64("from", int64(batch.From)), zap.Int64("to", int64(chat.SyncedTo)), zap.String("chatID", chatID))

		err = m.persistence.SetSyncTimestamps(batch.From, chat.SyncedTo, chat.ID)
		from = batch.From
		return nil, err
	})
	if err != nil {
		return 0, err
	}

	return from, nil
}
