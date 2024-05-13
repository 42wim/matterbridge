package protocol

import (
	"database/sql"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/status-im/status-go/eth-node/crypto"
	"github.com/status-im/status-go/protocol/common/shard"

	"go.uber.org/zap"

	"github.com/status-im/status-go/eth-node/types"
	"github.com/status-im/status-go/protocol/communities"
	"github.com/status-im/status-go/protocol/transport"
	"github.com/status-im/status-go/services/mailservers"
)

const (
	storeNodeAvailableTimeout = 30 * time.Second
)

// StoreNodeRequestStats is used in tests
type StoreNodeRequestStats struct {
	FetchedEnvelopesCount int
	FetchedPagesCount     int
}

type storeNodeRequestID struct {
	RequestType storeNodeRequestType `json:"requestType"`
	DataID      string               `json:"dataID"`
}

func (r *storeNodeRequestID) getCommunityID() string {
	switch r.RequestType {
	case storeNodeCommunityRequest:
		return r.DataID
	case storeNodeShardRequest:
		return strings.TrimSuffix(r.DataID, transport.CommunityShardInfoTopicPrefix())
	default:
		return ""
	}
}

type StoreNodeRequestManager struct {
	messenger *Messenger
	logger    *zap.Logger

	// activeRequests contain all ongoing store node requests.
	// Map is indexed with `DataID`.
	// Request might be duplicated in the map in case of contentType collisions.
	activeRequests map[storeNodeRequestID]*storeNodeRequest

	// activeRequestsLock should be locked each time activeRequests is being accessed or changed.
	activeRequestsLock sync.RWMutex

	onPerformingBatch func(MailserverBatch)
}

func NewStoreNodeRequestManager(m *Messenger) *StoreNodeRequestManager {
	return &StoreNodeRequestManager{
		messenger:          m,
		logger:             m.logger.Named("StoreNodeRequestManager"),
		activeRequests:     map[storeNodeRequestID]*storeNodeRequest{},
		activeRequestsLock: sync.RWMutex{},
		onPerformingBatch:  nil,
	}
}

// FetchCommunity makes a single request to store node for a given community id/shard pair.
// When a community is successfully fetched, a `CommunityFound` event will be emitted. If `waitForResponse == true`,
// the function will also wait for the store node response and return the fetched community.
// Automatically waits for an available store node.
// When a `nil` community and `nil` error is returned, that means the community wasn't found at the store node.
func (m *StoreNodeRequestManager) FetchCommunity(community communities.CommunityShard, opts []StoreNodeRequestOption) (*communities.Community, StoreNodeRequestStats, error) {
	cfg := buildStoreNodeRequestConfig(opts)

	m.logger.Info("requesting community from store node",
		zap.Any("community", community),
		zap.Any("config", cfg))

	requestCommunity := func(communityID string, shard *shard.Shard) (*communities.Community, StoreNodeRequestStats, error) {
		channel, err := m.subscribeToRequest(storeNodeCommunityRequest, communityID, shard, cfg)
		if err != nil {
			return nil, StoreNodeRequestStats{}, fmt.Errorf("failed to create a request for community: %w", err)
		}

		if !cfg.WaitForResponse {
			return nil, StoreNodeRequestStats{}, nil
		}

		result := <-channel
		return result.community, result.stats, result.err
	}

	// if shard was not passed or nil, request shard first
	communityShard := community.Shard
	if communityShard == nil {
		id := transport.CommunityShardInfoTopic(community.CommunityID)
		fetchedShard, err := m.subscribeToRequest(storeNodeShardRequest, id, shard.DefaultShard(), cfg)
		if err != nil {
			return nil, StoreNodeRequestStats{}, fmt.Errorf("failed to create a shard info request: %w", err)
		}

		if !cfg.WaitForResponse {
			go func() {
				shardResult := <-fetchedShard
				communityShard = shardResult.shard

				_, _, _ = requestCommunity(community.CommunityID, communityShard)
			}()
			return nil, StoreNodeRequestStats{}, nil
		}

		shardResult := <-fetchedShard
		communityShard = shardResult.shard
	}

	// request community with on shard
	return requestCommunity(community.CommunityID, communityShard)
}

// FetchCommunities makes a FetchCommunity for each element in given `communities` list.
// For each successfully fetched community, a `CommunityFound` event will be emitted. Ability to subscribe
// to results is not provided, because it's not needed and would complicate the code. `FetchCommunity` can
// be called directly if such functionality is needed.
//
// This function intentionally doesn't fetch multiple content topics in a single store node request. For now
// FetchCommunities is only used for regular (once in 2 minutes) fetching of curated communities. If one of
// those content topics is spammed with to many envelopes, then on each iteration we will have to fetch all
// of this spam first to get the envelopes in other content topics. To avoid this we keep independent requests
// for each content topic.
func (m *StoreNodeRequestManager) FetchCommunities(communities []communities.CommunityShard, opts []StoreNodeRequestOption) error {
	m.logger.Info("requesting communities from store node", zap.Any("communities", communities))

	// when fetching multiple communities we don't wait for the response
	opts = append(opts, WithWaitForResponseOption(false))

	var outErr error

	for _, community := range communities {
		_, _, err := m.FetchCommunity(community, opts)
		if err != nil {
			outErr = fmt.Errorf("%sfailed to create a request for community %s: %w", outErr, community.CommunityID, err)
		}
	}

	return outErr
}

func (m *StoreNodeRequestManager) FetchContact(contactID string, opts []StoreNodeRequestOption) (*Contact, StoreNodeRequestStats, error) {

	cfg := buildStoreNodeRequestConfig(opts)

	m.logger.Info("requesting contact from store node",
		zap.Any("contactID", contactID),
		zap.Any("config", cfg))

	channel, err := m.subscribeToRequest(storeNodeContactRequest, contactID, nil, cfg)
	if err != nil {
		return nil, StoreNodeRequestStats{}, fmt.Errorf("failed to create a request for community: %w", err)
	}

	if !cfg.WaitForResponse {
		return nil, StoreNodeRequestStats{}, nil
	}

	result := <-channel
	return result.contact, result.stats, result.err
}

// subscribeToRequest checks if a request for given community/contact is already in progress, creates and installs
// a new one if not found, and returns a subscription to the result of the found/started request.
// The subscription can then be used to get the result of the request, this could be either a community/contact or an error.
func (m *StoreNodeRequestManager) subscribeToRequest(requestType storeNodeRequestType, dataID string, shard *shard.Shard, cfg StoreNodeRequestConfig) (storeNodeResponseSubscription, error) {
	// It's important to unlock only after getting the subscription channel.
	// We also lock `activeRequestsLock` during finalizing the requests. This ensures that the subscription
	// created in this function will get the result even if the requests proceeds faster than this function ends.
	m.activeRequestsLock.Lock()
	defer m.activeRequestsLock.Unlock()

	requestID := storeNodeRequestID{
		RequestType: requestType,
		DataID:      dataID,
	}

	request, requestFound := m.activeRequests[requestID]

	if !requestFound {
		// Create corresponding filter
		var err error
		var filter *transport.Filter
		filterCreated := false

		filter, filterCreated, err = m.getFilter(requestType, dataID, shard)
		if err != nil {
			if filterCreated {
				m.forgetFilter(filter)
			}
			return nil, fmt.Errorf("failed to create community filter: %w", err)
		}

		request = m.newStoreNodeRequest()
		request.config = cfg
		request.pubsubTopic = filter.PubsubTopic
		request.requestID = requestID
		request.contentTopic = filter.ContentTopic
		if filterCreated {
			request.filterToForget = filter
		}

		m.activeRequests[requestID] = request
		request.start()
	}

	return request.subscribe(), nil
}

// newStoreNodeRequest creates a new storeNodeRequest struct
func (m *StoreNodeRequestManager) newStoreNodeRequest() *storeNodeRequest {
	return &storeNodeRequest{
		manager:       m,
		subscriptions: make([]storeNodeResponseSubscription, 0),
	}
}

// getFilter checks if a filter for a given community is already created and creates one of not found.
// Returns the found/created filter, a flag if the filter was created by the function and an error.
func (m *StoreNodeRequestManager) getFilter(requestType storeNodeRequestType, dataID string, shard *shard.Shard) (*transport.Filter, bool, error) {
	// First check if such filter already exists.
	filter := m.messenger.transport.FilterByChatID(dataID)
	if filter != nil {
		//we don't remember filter id associated with community because it was already installed
		return filter, false, nil
	}

	switch requestType {
	case storeNodeShardRequest, storeNodeCommunityRequest:
		// If filter wasn't installed we create it and
		// remember for uninstalling after response is received
		filters, err := m.messenger.transport.InitPublicFilters([]transport.FiltersToInitialize{{
			ChatID:      dataID,
			PubsubTopic: shard.PubsubTopic(),
		}})

		if err != nil {
			m.logger.Error("failed to install filter for community", zap.Error(err))
			return nil, false, err
		}

		if len(filters) != 1 {
			m.logger.Error("Unexpected number of filters created")
			return nil, true, fmt.Errorf("unexepcted number of filters created")
		}

		filter = filters[0]
	case storeNodeContactRequest:
		publicKeyBytes, err := types.DecodeHex(dataID)
		if err != nil {
			return nil, false, fmt.Errorf("failed to decode contact id: %w", err)
		}

		publicKey, err := crypto.UnmarshalPubkey(publicKeyBytes)
		if err != nil {
			return nil, false, fmt.Errorf("failed to unmarshal public key: %w", err)
		}

		filter, err = m.messenger.transport.JoinPrivate(publicKey)

		if err != nil {
			return nil, false, fmt.Errorf("failed to install filter for contact: %w", err)
		}

	default:
		return nil, false, fmt.Errorf("invalid store node request type: %d", requestType)
	}

	return filter, true, nil
}

// forgetFilter uninstalls the given filter
func (m *StoreNodeRequestManager) forgetFilter(filter *transport.Filter) {
	err := m.messenger.transport.RemoveFilters([]*transport.Filter{filter})
	if err != nil {
		m.logger.Warn("failed to remove filter", zap.Error(err))
	}
}

type storeNodeRequestType int

const (
	storeNodeCommunityRequest storeNodeRequestType = iota
	storeNodeContactRequest
	storeNodeShardRequest
)

// storeNodeRequest represents a single store node batch request.
// For a valid storeNodeRequest to be performed, the user must set all the struct fields and call start method.
type storeNodeRequest struct {
	requestID storeNodeRequestID

	// request parameters
	pubsubTopic      string
	contentTopic     types.TopicType
	minimumDataClock uint64
	config           StoreNodeRequestConfig

	// request corresponding metadata to be used in finalize
	filterToForget *transport.Filter

	// internal fields
	manager       *StoreNodeRequestManager
	subscriptions []storeNodeResponseSubscription
	result        storeNodeRequestResult
}

// storeNodeRequestResult contains result of a single storeNodeRequest
// Further by using `data` we mean community/contact, depending on request type.
// If any error occurs during the request, err field will be set.
// If data was successfully fetched, data field will contain the fetched information.
// If data wasn't found in store node, then a data will be set to `nil`.
// stats will contain information about the performed request that might be useful for testing.
type storeNodeRequestResult struct {
	err   error
	stats StoreNodeRequestStats
	// One of data fields (community or contact) will be present depending on request type
	community *communities.Community
	contact   *Contact
	shard     *shard.Shard
}

type storeNodeResponseSubscription = chan storeNodeRequestResult

func (r *storeNodeRequest) subscribe() storeNodeResponseSubscription {
	channel := make(storeNodeResponseSubscription, 100)
	r.subscriptions = append(r.subscriptions, channel)
	return channel
}

func (r *storeNodeRequest) finalize() {
	r.manager.activeRequestsLock.Lock()
	defer r.manager.activeRequestsLock.Unlock()

	r.manager.logger.Info("request finished",
		zap.Any("requestID", r.requestID),
		zap.Bool("communityFound", r.result.community != nil),
		zap.Bool("contactFound", r.result.contact != nil),
		zap.Bool("shardFound", r.result.shard != nil),
		zap.Error(r.result.err))

	// Send the result to subscribers
	// It's important that this is done with `activeRequestsLock` locked.
	for _, s := range r.subscriptions {
		s <- r.result
		close(s)
	}

	if r.result.community != nil {
		r.manager.messenger.passStoredCommunityInfoToSignalHandler(r.result.community)
	}

	delete(r.manager.activeRequests, r.requestID)

	if r.filterToForget != nil {
		r.manager.forgetFilter(r.filterToForget)
	}
}

func (r *storeNodeRequest) shouldFetchNextPage(envelopesCount int) (bool, uint32) {
	logger := r.manager.logger.With(
		zap.Any("requestID", r.requestID),
		zap.Int("envelopesCount", envelopesCount))

	r.result.stats.FetchedEnvelopesCount += envelopesCount
	r.result.stats.FetchedPagesCount++

	// Force all received envelopes to be processed
	r.manager.messenger.ProcessAllMessages()

	// Try to get community from database
	switch r.requestID.RequestType {
	case storeNodeCommunityRequest:
		communityID, err := types.DecodeHex(r.requestID.DataID)
		if err != nil {
			logger.Error("failed to decode community ID",
				zap.String("communityID", r.requestID.DataID),
				zap.Error(err))
			r.result = storeNodeRequestResult{
				community: nil,
				err:       fmt.Errorf("failed to decode community ID: %w", err),
			}
			return false, 0 // failed to decode community ID, no sense to continue the procedure
		}

		// check if community is waiting for a verification and do a verification manually
		_, err = r.manager.messenger.communitiesManager.ValidateCommunityByID(communityID)
		if err != nil {
			logger.Error("failed to validate community by ID",
				zap.String("communityID", r.requestID.DataID),
				zap.Error(err))
			r.result = storeNodeRequestResult{
				community: nil,
				err:       fmt.Errorf("failed to validate community by ID: %w", err),
			}
			return false, 0 // failed to validate community, no sense to continue the procedure
		}

		community, err := r.manager.messenger.communitiesManager.GetByID(communityID)

		if err != nil && err != communities.ErrOrgNotFound {
			logger.Error("failed to read community from database",
				zap.String("communityID", r.requestID.DataID),
				zap.Error(err))
			r.result = storeNodeRequestResult{
				community: nil,
				err:       fmt.Errorf("failed to read community from database: %w", err),
			}
			return false, 0 // failed to read from database, no sense to continue the procedure
		}

		if community == nil {
			// community not found in the database, request next page
			logger.Debug("community still not fetched")
			return true, r.config.FurtherPageSize
		}

		// We check here if the community was fetched actually fetched and updated, because it
		// could be that the community was already in the database when we started the fetching.
		//
		// Would be perfect if we could track that the community was in these particular envelopes,
		// but I don't think that's possible right now. We check if clock was updated instead.

		if community.Clock() <= r.minimumDataClock {
			logger.Debug("local community description is not newer than existing",
				zap.Any("existingClock", community.Clock()),
				zap.Any("minimumDataClock", r.minimumDataClock),
			)
			return true, r.config.FurtherPageSize
		}

		logger.Debug("community found",
			zap.String("displayName", community.Name()))

		r.result.community = community

	case storeNodeShardRequest:
		communityIDStr := strings.TrimSuffix(r.requestID.DataID, transport.CommunityShardInfoTopicPrefix())
		communityID, err := types.DecodeHex(communityIDStr)
		if err != nil {
			logger.Error("decode community ID failed",
				zap.String("communityID", communityIDStr),
				zap.Error(err))
			return false, 0
		}
		shardResult, err := r.manager.messenger.communitiesManager.GetCommunityShard(communityID)
		if err != nil {
			if err != sql.ErrNoRows {
				logger.Error("failed to read from database",
					zap.String("communityID", communityIDStr),
					zap.Error(err))
				r.result = storeNodeRequestResult{
					shard: nil,
					err:   fmt.Errorf("failed to read from database: %w", err),
				}
				return false, 0 // failed to read from database, no sense to continue the procedure
			}
		}

		logger.Debug("shard found",
			zap.String("community", communityIDStr),
			zap.Any("shard", shardResult),
		)

		r.result.shard = shardResult

	case storeNodeContactRequest:
		contact := r.manager.messenger.GetContactByID(r.requestID.DataID)

		if contact == nil {
			// contact not found in the database, request next page
			logger.Debug("contact still not fetched")
			return true, r.config.FurtherPageSize
		}

		logger.Debug("contact found",
			zap.String("displayName", contact.DisplayName))

		r.result.contact = contact
	}

	return !r.config.StopWhenDataFound, r.config.FurtherPageSize
}

func (r *storeNodeRequest) routine() {
	r.manager.logger.Info("starting store node request",
		zap.Any("requestID", r.requestID),
		zap.String("pubsubTopic", r.pubsubTopic),
		zap.Any("contentTopic", r.contentTopic),
	)

	// Return a nil community and no error when request was
	// performed successfully, but no community/contact found.
	r.result = storeNodeRequestResult{
		err:       nil,
		community: nil,
		contact:   nil,
		shard:     nil,
	}

	defer func() {
		r.finalize()
	}()

	communityID := r.requestID.getCommunityID()

	if r.requestID.RequestType != storeNodeCommunityRequest || !r.manager.messenger.communityStorenodes.HasStorenodeSetup(communityID) {
		if !r.manager.messenger.waitForAvailableStoreNode(storeNodeAvailableTimeout) {
			r.result.err = fmt.Errorf("store node is not available")
			return
		}
	}

	storeNode := r.manager.messenger.getActiveMailserver(communityID)

	// Check if community already exists locally and get Clock.
	if r.requestID.RequestType == storeNodeCommunityRequest {
		localCommunity, _ := r.manager.messenger.communitiesManager.GetByIDString(communityID)
		if localCommunity != nil {
			r.minimumDataClock = localCommunity.Clock()
		}
	}

	// Start store node request
	from, to := r.manager.messenger.calculateMailserverTimeBounds(oneMonthDuration)

	_, err := r.manager.messenger.performMailserverRequest(storeNode, func(ms mailservers.Mailserver) (*MessengerResponse, error) {
		batch := MailserverBatch{
			From:        from,
			To:          to,
			PubsubTopic: r.pubsubTopic,
			Topics:      []types.TopicType{r.contentTopic},
		}
		r.manager.logger.Info("perform store node request", zap.Any("batch", batch))
		if r.manager.onPerformingBatch != nil {
			r.manager.onPerformingBatch(batch)
		}

		return nil, r.manager.messenger.processMailserverBatchWithOptions(ms, batch, r.config.InitialPageSize, r.shouldFetchNextPage, true)
	})

	r.result.err = err
}

func (r *storeNodeRequest) start() {
	go r.routine()
}
