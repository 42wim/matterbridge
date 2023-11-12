package collectibles

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/event"
	"github.com/ethereum/go-ethereum/log"

	"github.com/status-im/status-go/multiaccounts/accounts"
	"github.com/status-im/status-go/rpc/network"

	"github.com/status-im/status-go/services/wallet/async"
	"github.com/status-im/status-go/services/wallet/bigint"
	walletCommon "github.com/status-im/status-go/services/wallet/common"
	"github.com/status-im/status-go/services/wallet/community"
	"github.com/status-im/status-go/services/wallet/thirdparty"
	"github.com/status-im/status-go/services/wallet/transfer"
	"github.com/status-im/status-go/services/wallet/walletevent"
)

// These events are used to notify the UI of state changes
const (
	EventCollectiblesOwnershipUpdateStarted           walletevent.EventType = "wallet-collectibles-ownership-update-started"
	EventCollectiblesOwnershipUpdatePartial           walletevent.EventType = "wallet-collectibles-ownership-update-partial"
	EventCollectiblesOwnershipUpdateFinished          walletevent.EventType = "wallet-collectibles-ownership-update-finished"
	EventCollectiblesOwnershipUpdateFinishedWithError walletevent.EventType = "wallet-collectibles-ownership-update-finished-with-error"
	EventCommunityCollectiblesReceived                walletevent.EventType = "wallet-collectibles-community-collectibles-received"
	EventCollectiblesDataUpdated                      walletevent.EventType = "wallet-collectibles-data-updated"

	EventOwnedCollectiblesFilteringDone walletevent.EventType = "wallet-owned-collectibles-filtering-done"
	EventGetCollectiblesDetailsDone     walletevent.EventType = "wallet-get-collectibles-details-done"
)

type OwnershipUpdateMessage struct {
	Added   []thirdparty.CollectibleUniqueID `json:"added"`
	Updated []thirdparty.CollectibleUniqueID `json:"updated"`
	Removed []thirdparty.CollectibleUniqueID `json:"removed"`
}

type CollectibleDataType byte

const (
	CollectibleDataTypeUniqueID CollectibleDataType = iota
	CollectibleDataTypeHeader
	CollectibleDataTypeDetails
	CollectibleDataTypeCommunityHeader
)

type FetchType byte

const (
	FetchTypeNeverFetch FetchType = iota
	FetchTypeAlwaysFetch
	FetchTypeFetchIfNotCached
	FetchTypeFetchIfCacheOld
)

type FetchCriteria struct {
	FetchType          FetchType `json:"fetch_type"`
	MaxCacheAgeSeconds int64     `json:"max_cache_age_seconds"`
}

var (
	filterOwnedCollectiblesTask = async.TaskType{
		ID:     1,
		Policy: async.ReplacementPolicyCancelOld,
	}
	getCollectiblesDataTask = async.TaskType{
		ID:     2,
		Policy: async.ReplacementPolicyCancelOld,
	}
)

type Service struct {
	manager          *Manager
	controller       *Controller
	db               *sql.DB
	ownershipDB      *OwnershipDB
	transferDB       *transfer.Database
	communityManager *community.Manager
	walletFeed       *event.Feed
	scheduler        *async.MultiClientScheduler
	group            *async.Group
}

func NewService(
	db *sql.DB,
	walletFeed *event.Feed,
	accountsDB *accounts.Database,
	accountsFeed *event.Feed,
	settingsFeed *event.Feed,
	communityManager *community.Manager,
	networkManager *network.Manager,
	manager *Manager) *Service {
	s := &Service{
		manager:          manager,
		controller:       NewController(db, walletFeed, accountsDB, accountsFeed, settingsFeed, networkManager, manager),
		db:               db,
		ownershipDB:      NewOwnershipDB(db),
		transferDB:       transfer.NewDB(db),
		communityManager: communityManager,
		walletFeed:       walletFeed,
		scheduler:        async.NewMultiClientScheduler(),
		group:            async.NewGroup(context.Background()),
	}
	s.controller.SetOwnedCollectiblesChangeCb(s.onOwnedCollectiblesChange)
	s.controller.SetCollectiblesTransferCb(s.onCollectiblesTransfer)
	return s
}

type ErrorCode = int

const (
	ErrorCodeSuccess ErrorCode = iota + 1
	ErrorCodeTaskCanceled
	ErrorCodeFailed
)

type OwnershipStatus struct {
	State     OwnershipState `json:"state"`
	Timestamp int64          `json:"timestamp"`
}

type OwnershipStatusPerChainID = map[walletCommon.ChainID]OwnershipStatus
type OwnershipStatusPerAddressAndChainID = map[common.Address]OwnershipStatusPerChainID

type GetOwnedCollectiblesResponse struct {
	Collectibles []Collectible `json:"collectibles"`
	Offset       int           `json:"offset"`
	// Used to indicate that there might be more collectibles that were not returned
	// based on a simple heuristic
	HasMore         bool                                `json:"hasMore"`
	OwnershipStatus OwnershipStatusPerAddressAndChainID `json:"ownershipStatus"`
	ErrorCode       ErrorCode                           `json:"errorCode"`
}

type GetCollectiblesByUniqueIDResponse struct {
	Collectibles []Collectible `json:"collectibles"`
	ErrorCode    ErrorCode     `json:"errorCode"`
}

type GetOwnedCollectiblesReturnType struct {
	collectibles    []Collectible
	hasMore         bool
	ownershipStatus OwnershipStatusPerAddressAndChainID
}

type GetCollectiblesByUniqueIDReturnType struct {
	collectibles []Collectible
}

func (s *Service) GetOwnedCollectibles(
	ctx context.Context,
	chainIDs []walletCommon.ChainID,
	addresses []common.Address,
	filter Filter,
	offset int,
	limit int,
	dataType CollectibleDataType,
	fetchCriteria FetchCriteria) (*GetOwnedCollectiblesReturnType, error) {
	err := s.fetchOwnedCollectiblesIfNeeded(ctx, chainIDs, addresses, fetchCriteria)
	if err != nil {
		return nil, err
	}

	ids, hasMore, err := s.FilterOwnedCollectibles(chainIDs, addresses, filter, offset, limit)
	if err != nil {
		return nil, err
	}

	collectibles, err := s.collectibleIDsToDataType(ctx, ids, dataType)
	if err != nil {
		return nil, err
	}

	ownershipStatus, err := s.GetOwnershipStatus(chainIDs, addresses)
	if err != nil {
		return nil, err
	}

	return &GetOwnedCollectiblesReturnType{
		collectibles:    collectibles,
		hasMore:         hasMore,
		ownershipStatus: ownershipStatus,
	}, err
}

func (s *Service) needsToFetch(chainID walletCommon.ChainID, address common.Address, fetchCriteria FetchCriteria) (bool, error) {
	mustFetch := false
	switch fetchCriteria.FetchType {
	case FetchTypeAlwaysFetch:
		mustFetch = true
	case FetchTypeNeverFetch:
		mustFetch = false
	case FetchTypeFetchIfNotCached, FetchTypeFetchIfCacheOld:
		timestamp, err := s.ownershipDB.GetOwnershipUpdateTimestamp(address, chainID)
		if err != nil {
			return false, err
		}
		if timestamp == InvalidTimestamp ||
			(fetchCriteria.FetchType == FetchTypeFetchIfCacheOld && timestamp+fetchCriteria.MaxCacheAgeSeconds < time.Now().Unix()) {
			mustFetch = true
		}
	}
	return mustFetch, nil
}

func (s *Service) fetchOwnedCollectiblesIfNeeded(ctx context.Context, chainIDs []walletCommon.ChainID, addresses []common.Address, fetchCriteria FetchCriteria) error {
	if fetchCriteria.FetchType == FetchTypeNeverFetch {
		return nil
	}

	group := async.NewGroup(ctx)
	for _, address := range addresses {
		for _, chainID := range chainIDs {
			mustFetch, err := s.needsToFetch(chainID, address, fetchCriteria)
			if err != nil {
				return err
			}
			if mustFetch {
				command := newLoadOwnedCollectiblesCommand(s.manager, s.ownershipDB, s.walletFeed, chainID, address, nil)
				group.Add(command.Command())
			}
		}
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-group.WaitAsync():
		return nil
	}
}

// GetOwnedCollectiblesAsync allows only one filter task to run at a time
// and it cancels the current one if a new one is started
// All calls will trigger an EventOwnedCollectiblesFilteringDone event with the result of the filtering
func (s *Service) GetOwnedCollectiblesAsync(
	requestID int32,
	chainIDs []walletCommon.ChainID,
	addresses []common.Address,
	filter Filter,
	offset int,
	limit int,
	dataType CollectibleDataType,
	fetchCriteria FetchCriteria) {
	s.scheduler.Enqueue(requestID, filterOwnedCollectiblesTask, func(ctx context.Context) (interface{}, error) {
		return s.GetOwnedCollectibles(ctx, chainIDs, addresses, filter, offset, limit, dataType, fetchCriteria)
	}, func(result interface{}, taskType async.TaskType, err error) {
		res := GetOwnedCollectiblesResponse{
			ErrorCode: ErrorCodeFailed,
		}

		if errors.Is(err, context.Canceled) || errors.Is(err, async.ErrTaskOverwritten) {
			res.ErrorCode = ErrorCodeTaskCanceled
		} else if err == nil {
			fnRet := result.(*GetOwnedCollectiblesReturnType)

			if err == nil {
				res.Collectibles = fnRet.collectibles
				res.Offset = offset
				res.HasMore = fnRet.hasMore
				res.OwnershipStatus = fnRet.ownershipStatus
				res.ErrorCode = ErrorCodeSuccess
			}
		}

		s.sendResponseEvent(&requestID, EventOwnedCollectiblesFilteringDone, res, err)
	})
}

func (s *Service) GetCollectiblesByUniqueID(
	ctx context.Context,
	uniqueIDs []thirdparty.CollectibleUniqueID,
	dataType CollectibleDataType) (*GetCollectiblesByUniqueIDReturnType, error) {
	collectibles, err := s.collectibleIDsToDataType(ctx, uniqueIDs, dataType)
	if err != nil {
		return nil, err
	}
	return &GetCollectiblesByUniqueIDReturnType{
		collectibles: collectibles,
	}, err
}

func (s *Service) GetCollectiblesByUniqueIDAsync(
	requestID int32,
	uniqueIDs []thirdparty.CollectibleUniqueID,
	dataType CollectibleDataType) {
	s.scheduler.Enqueue(requestID, getCollectiblesDataTask, func(ctx context.Context) (interface{}, error) {
		return s.GetCollectiblesByUniqueID(ctx, uniqueIDs, dataType)
	}, func(result interface{}, taskType async.TaskType, err error) {
		res := GetCollectiblesByUniqueIDResponse{
			ErrorCode: ErrorCodeFailed,
		}

		if errors.Is(err, context.Canceled) || errors.Is(err, async.ErrTaskOverwritten) {
			res.ErrorCode = ErrorCodeTaskCanceled
		} else if err == nil {
			fnRet := result.(*GetCollectiblesByUniqueIDReturnType)

			if err == nil {
				res.Collectibles = fnRet.collectibles
				res.ErrorCode = ErrorCodeSuccess
			}
		}

		s.sendResponseEvent(&requestID, EventGetCollectiblesDetailsDone, res, err)
	})
}

func (s *Service) RefetchOwnedCollectibles() {
	s.controller.RefetchOwnedCollectibles()
}

func (s *Service) Start() {
	s.controller.Start()
}

func (s *Service) Stop() {
	s.controller.Stop()

	s.scheduler.Stop()
}

func (s *Service) sendResponseEvent(requestID *int32, eventType walletevent.EventType, payloadObj interface{}, resErr error) {
	payload, err := json.Marshal(payloadObj)
	if err != nil {
		log.Error("Error marshaling response: %v; result error: %w", err, resErr)
	} else {
		err = resErr
	}

	log.Debug("wallet.api.collectibles.Service RESPONSE", "requestID", requestID, "eventType", eventType, "error", err, "payload.len", len(payload))

	event := walletevent.Event{
		Type:    eventType,
		Message: string(payload),
	}

	if requestID != nil {
		event.RequestID = new(int)
		*event.RequestID = int(*requestID)
	}

	s.walletFeed.Send(event)
}

func (s *Service) FilterOwnedCollectibles(chainIDs []walletCommon.ChainID, owners []common.Address, filter Filter, offset int, limit int) ([]thirdparty.CollectibleUniqueID, bool, error) {
	ctx := context.Background()
	// Request one more than limit, to check if DB has more available
	ids, err := filterOwnedCollectibles(ctx, s.db, chainIDs, owners, filter, offset, limit+1)
	if err != nil {
		return nil, false, err
	}

	hasMore := len(ids) > limit
	if hasMore {
		ids = ids[:limit]
	}

	return ids, hasMore, nil
}

func (s *Service) GetOwnedCollectible(chainID walletCommon.ChainID, owner common.Address, contractAddress common.Address, tokenID *big.Int) (*thirdparty.CollectibleUniqueID, error) {
	return s.ownershipDB.GetOwnedCollectible(chainID, owner, contractAddress, tokenID)
}

func (s *Service) GetOwnershipStatus(chainIDs []walletCommon.ChainID, owners []common.Address) (OwnershipStatusPerAddressAndChainID, error) {
	ret := make(OwnershipStatusPerAddressAndChainID)
	for _, address := range owners {
		ret[address] = make(OwnershipStatusPerChainID)
		for _, chainID := range chainIDs {
			timestamp, err := s.ownershipDB.GetOwnershipUpdateTimestamp(address, chainID)
			if err != nil {
				return nil, err
			}
			ret[address][chainID] = OwnershipStatus{
				State:     s.controller.GetCommandState(chainID, address),
				Timestamp: timestamp,
			}
		}
	}

	return ret, nil
}

func (s *Service) collectibleIDsToDataType(ctx context.Context, ids []thirdparty.CollectibleUniqueID, dataType CollectibleDataType) ([]Collectible, error) {
	switch dataType {
	case CollectibleDataTypeUniqueID:
		return idsToCollectibles(ids), nil
	case CollectibleDataTypeHeader, CollectibleDataTypeDetails, CollectibleDataTypeCommunityHeader:
		collectibles, err := s.manager.FetchAssetsByCollectibleUniqueID(ctx, ids, true)
		if err != nil {
			return nil, err
		}
		switch dataType {
		case CollectibleDataTypeHeader:
			return fullCollectiblesDataToHeaders(collectibles), nil
		case CollectibleDataTypeDetails:
			return fullCollectiblesDataToDetails(collectibles), nil
		case CollectibleDataTypeCommunityHeader:
			return fullCollectiblesDataToCommunityHeader(collectibles), nil
		}
	}
	return nil, errors.New("unknown data type")
}

func (s *Service) onOwnedCollectiblesChange(ownedCollectiblesChange OwnedCollectiblesChange) {
	// Try to find a matching transfer for newly added/updated collectibles
	switch ownedCollectiblesChange.changeType {
	case OwnedCollectiblesChangeTypeAdded, OwnedCollectiblesChangeTypeUpdated:
		// For recently added/updated collectibles, try to find a matching transfer
		s.lookupTransferForCollectibles(ownedCollectiblesChange.ownedCollectibles)
		s.notifyCommunityCollectiblesReceived(ownedCollectiblesChange.ownedCollectibles)
	}
}

func (s *Service) onCollectiblesTransfer(account common.Address, chainID walletCommon.ChainID, transfers []transfer.Transfer) {
	for _, transfer := range transfers {
		// If Collectible is already in the DB, update transfer ID with the latest detected transfer
		id := thirdparty.CollectibleUniqueID{
			ContractID: thirdparty.ContractID{
				ChainID: chainID,
				Address: transfer.Log.Address,
			},
			TokenID: &bigint.BigInt{Int: transfer.TokenID},
		}
		err := s.manager.SetCollectibleTransferID(account, id, transfer.ID, true)
		if err != nil {
			log.Error("Error setting transfer ID for collectible", "error", err)
		}
	}
}

func (s *Service) lookupTransferForCollectibles(ownedCollectibles OwnedCollectibles) {
	// There are some limitations to this approach:
	// - Collectibles ownership and transfers are not in sync and might represent the state at different moments.
	// - We have no way of knowing if the latest collectible transfer we've detected is actually the latest one, so the timestamp we
	// use might be older than the real one.
	// - There might be detected transfers that are temporarily not reflected in the collectibles ownership.
	// - For ERC721 tokens we should only look for incoming transfers. For ERC1155 tokens we should look for both incoming and outgoing transfers.
	// We need to get the contract standard for each collectible to know which approach to take.
	for _, id := range ownedCollectibles.ids {
		transfer, err := s.transferDB.GetLatestCollectibleTransfer(ownedCollectibles.account, id)
		if err != nil {
			log.Error("Error fetching latest collectible transfer", "error", err)
			continue
		}
		if transfer != nil {
			err = s.manager.SetCollectibleTransferID(ownedCollectibles.account, id, transfer.ID, false)
			if err != nil {
				log.Error("Error setting transfer ID for collectible", "error", err)
			}
		}
	}
}

func (s *Service) notifyCommunityCollectiblesReceived(ownedCollectibles OwnedCollectibles) {
	ctx := context.Background()

	collectiblesData, err := s.manager.FetchAssetsByCollectibleUniqueID(ctx, ownedCollectibles.ids, false)
	if err != nil {
		log.Error("Error fetching collectibles data", "error", err)
		return
	}

	communityCollectibles := fullCollectiblesDataToCommunityHeader(collectiblesData)

	if len(communityCollectibles) == 0 {
		return
	}

	encodedMessage, err := json.Marshal(communityCollectibles)
	if err != nil {
		return
	}

	s.walletFeed.Send(walletevent.Event{
		Type:    EventCommunityCollectiblesReceived,
		ChainID: uint64(ownedCollectibles.chainID),
		Accounts: []common.Address{
			ownedCollectibles.account,
		},
		Message: string(encodedMessage),
	})
}
