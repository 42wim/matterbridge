package activity

import (
	"context"
	"errors"
	"strconv"

	"golang.org/x/exp/slices"

	eth "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/event"
	"github.com/ethereum/go-ethereum/log"
	"github.com/status-im/status-go/services/wallet/async"
	"github.com/status-im/status-go/services/wallet/common"
	"github.com/status-im/status-go/services/wallet/transfer"
	"github.com/status-im/status-go/services/wallet/walletevent"
	"github.com/status-im/status-go/transactions"
)

const nilStr = "nil"

type EntryIdentity struct {
	payloadType PayloadType
	transaction *transfer.TransactionIdentity
	id          transfer.MultiTransactionIDType
}

// func (e EntryIdentity) same(a EntryIdentity) bool {
// 	return a.payloadType == e.payloadType && (a.transaction == e.transaction && (a.transaction == nil || (a.transaction.ChainID == e.transaction.ChainID &&
// 		a.transaction.Hash == e.transaction.Hash &&
// 		a.transaction.Address == e.transaction.Address))) && a.id == e.id
// }

func (e EntryIdentity) key() string {
	txID := nilStr
	if e.transaction != nil {
		txID = strconv.FormatUint(uint64(e.transaction.ChainID), 10) + e.transaction.Hash.Hex() + e.transaction.Address.Hex()
	}
	return strconv.Itoa(e.payloadType) + txID + strconv.FormatInt(int64(e.id), 16)
}

type SessionID int32

type Session struct {
	id SessionID

	// Filter info
	//
	addresses    []eth.Address
	allAddresses bool
	chainIDs     []common.ChainID
	filter       Filter

	// model is a mirror of the data model presentation has (sent by EventActivityFilteringDone)
	model []EntryIdentity
	// new holds the new entries until user requests update by calling ResetFilterSession
	new []EntryIdentity
}

// SessionUpdate payload for EventActivitySessionUpdated
type SessionUpdate struct {
	HasNewEntries *bool           `json:"hasNewEntries,omitempty"`
	Removed       []EntryIdentity `json:"removed,omitempty"`
	Updated       []Entry         `json:"updated,omitempty"`
}

type fullFilterParams struct {
	sessionID    SessionID
	addresses    []eth.Address
	allAddresses bool
	chainIDs     []common.ChainID
	filter       Filter
}

func (s *Service) internalFilter(f fullFilterParams, offset int, count int, processResults func(entries []Entry) (offsetOverride int)) {
	s.scheduler.Enqueue(int32(f.sessionID), filterTask, func(ctx context.Context) (interface{}, error) {
		activities, err := getActivityEntries(ctx, s.getDeps(), f.addresses, f.allAddresses, f.chainIDs, f.filter, offset, count)
		return activities, err
	}, func(result interface{}, taskType async.TaskType, err error) {
		res := FilterResponse{
			ErrorCode: ErrorCodeFailed,
		}

		if errors.Is(err, context.Canceled) || errors.Is(err, async.ErrTaskOverwritten) {
			res.ErrorCode = ErrorCodeTaskCanceled
		} else if err == nil {
			activities := result.([]Entry)
			res.Activities = activities
			res.HasMore = len(activities) == count
			res.ErrorCode = ErrorCodeSuccess

			res.Offset = processResults(activities)
		}

		int32SessionID := int32(f.sessionID)
		sendResponseEvent(s.eventFeed, &int32SessionID, EventActivityFilteringDone, res, err)

		s.getActivityDetailsAsync(int32SessionID, res.Activities)
	})
}

func (s *Service) StartFilterSession(addresses []eth.Address, allAddresses bool, chainIDs []common.ChainID, filter Filter, firstPageCount int) SessionID {
	sessionID := s.nextSessionID()

	// TODO #12120: sort rest of the filters
	// TODO #12120: prettyfy this
	slices.SortFunc(addresses, func(a eth.Address, b eth.Address) bool {
		return a.Hex() < b.Hex()
	})
	slices.Sort(chainIDs)
	slices.SortFunc(filter.CounterpartyAddresses, func(a eth.Address, b eth.Address) bool {
		return a.Hex() < b.Hex()
	})

	s.sessionsRWMutex.Lock()
	subscribeToEvents := len(s.sessions) == 0
	session := &Session{
		id: sessionID,

		addresses:    addresses,
		allAddresses: allAddresses,
		chainIDs:     chainIDs,
		filter:       filter,

		model: make([]EntryIdentity, 0, firstPageCount),
	}
	s.sessions[sessionID] = session

	if subscribeToEvents {
		s.subscribeToEvents()
	}
	s.sessionsRWMutex.Unlock()

	s.internalFilter(
		fullFilterParams{
			sessionID:    sessionID,
			addresses:    addresses,
			allAddresses: allAddresses,
			chainIDs:     chainIDs,
			filter:       filter,
		},
		0,
		firstPageCount,
		func(entries []Entry) (offset int) {
			// Mirror identities for update use
			s.sessionsRWMutex.Lock()
			defer s.sessionsRWMutex.Unlock()

			session.model = make([]EntryIdentity, 0, len(entries))
			for _, a := range entries {
				session.model = append(session.model, EntryIdentity{
					payloadType: a.payloadType,
					transaction: a.transaction,
					id:          a.id,
				})
			}
			return 0
		},
	)

	return sessionID
}

func (s *Service) ResetFilterSession(id SessionID, firstPageCount int) error {
	session, found := s.sessions[id]
	if !found {
		return errors.New("session not found")
	}

	s.internalFilter(
		fullFilterParams{
			sessionID:    id,
			addresses:    session.addresses,
			allAddresses: session.allAddresses,
			chainIDs:     session.chainIDs,
			filter:       session.filter,
		},
		0,
		firstPageCount,
		func(entries []Entry) (offset int) {
			s.sessionsRWMutex.Lock()
			defer s.sessionsRWMutex.Unlock()

			// Mark new entries
			newMap := entryIdsToMap(session.new)
			for i, a := range entries {
				_, isNew := newMap[a.getIdentity().key()]
				entries[i].isNew = isNew
			}
			session.new = nil

			// Mirror client identities for checking updates
			session.model = make([]EntryIdentity, 0, len(entries))
			for _, a := range entries {
				session.model = append(session.model, EntryIdentity{
					payloadType: a.payloadType,
					transaction: a.transaction,
					id:          a.id,
				})
			}
			return 0
		},
	)
	return nil
}

func (s *Service) GetMoreForFilterSession(id SessionID, pageCount int) error {
	session, found := s.sessions[id]
	if !found {
		return errors.New("session not found")
	}

	prevModelLen := len(session.model)
	s.internalFilter(
		fullFilterParams{
			sessionID:    id,
			addresses:    session.addresses,
			allAddresses: session.allAddresses,
			chainIDs:     session.chainIDs,
			filter:       session.filter,
		},
		prevModelLen+len(session.new),
		pageCount,
		func(entries []Entry) (offset int) {
			s.sessionsRWMutex.Lock()
			defer s.sessionsRWMutex.Unlock()

			// Mirror client identities for checking updates
			for _, a := range entries {
				session.model = append(session.model, EntryIdentity{
					payloadType: a.payloadType,
					transaction: a.transaction,
					id:          a.id,
				})
			}

			// Overwrite the offset to account for new entries
			return prevModelLen
		},
	)
	return nil
}

// subscribeToEvents should be called with sessionsRWMutex locked for writing
func (s *Service) subscribeToEvents() {
	s.ch = make(chan walletevent.Event, 100)
	s.subscriptions = s.eventFeed.Subscribe(s.ch)
	go s.processEvents()
}

// TODO #12120: check that it exits on channel close
func (s *Service) processEvents() {
	for event := range s.ch {
		// TODO #12120: process rest of the events
		// TODO #12120: debounce for 1s
		if event.Type == transactions.EventPendingTransactionUpdate {
			for sessionID := range s.sessions {
				session := s.sessions[sessionID]
				activities, err := getActivityEntries(context.Background(), s.getDeps(), session.addresses, session.allAddresses, session.chainIDs, session.filter, 0, len(session.model))
				if err != nil {
					log.Error("Error getting activity entries", "error", err)
					continue
				}

				s.sessionsRWMutex.RLock()
				allData := append(session.model, session.new...)
				new, _ /*removed*/ := findUpdates(allData, activities)
				s.sessionsRWMutex.RUnlock()

				s.sessionsRWMutex.Lock()
				lastProcessed := -1
				for i, idRes := range new {
					if i-lastProcessed > 1 {
						// The events are not continuous, therefore these are not on top but mixed between existing entries
						break
					}
					lastProcessed = idRes.newPos
					// TODO #12120: make it more generic to follow the detection function
					// TODO #12120: hold the first few and only send mixed and removed
					if session.new == nil {
						session.new = make([]EntryIdentity, 0, len(new))
					}
					session.new = append(session.new, idRes.id)
				}

				// TODO #12120: mixed

				s.sessionsRWMutex.Unlock()

				go notify(s.eventFeed, sessionID, len(session.new) > 0)
			}
		}
	}
}

func notify(eventFeed *event.Feed, id SessionID, hasNewEntries bool) {
	payload := SessionUpdate{}
	if hasNewEntries {
		payload.HasNewEntries = &hasNewEntries
	}

	sendResponseEvent(eventFeed, (*int32)(&id), EventActivitySessionUpdated, payload, nil)
}

// unsubscribeFromEvents should be called with sessionsRWMutex locked for writing
func (s *Service) unsubscribeFromEvents() {
	s.subscriptions.Unsubscribe()
	s.subscriptions = nil
}

func (s *Service) StopFilterSession(id SessionID) {
	s.sessionsRWMutex.Lock()
	delete(s.sessions, id)
	if len(s.sessions) == 0 {
		s.unsubscribeFromEvents()
	}
	s.sessionsRWMutex.Unlock()

	// Cancel any pending or ongoing task
	s.scheduler.Enqueue(int32(id), filterTask, func(ctx context.Context) (interface{}, error) {
		return nil, nil
	}, func(result interface{}, taskType async.TaskType, err error) {})
}

func (s *Service) getActivityDetailsAsync(requestID int32, entries []Entry) {
	if len(entries) == 0 {
		return
	}

	ctx := context.Background()

	go func() {
		activityData, err := s.getActivityDetails(ctx, entries)
		if len(activityData) != 0 {
			sendResponseEvent(s.eventFeed, &requestID, EventActivityFilteringUpdate, activityData, err)
		}
	}()
}

type mixedIdentityResult struct {
	newPos int
	id     EntryIdentity
}

func entryIdsToMap(ids []EntryIdentity) map[string]EntryIdentity {
	idsMap := make(map[string]EntryIdentity, len(ids))
	for _, id := range ids {
		idsMap[id.key()] = id
	}
	return idsMap
}

func entriesToMap(entries []Entry) map[string]Entry {
	entryMap := make(map[string]Entry, len(entries))
	for _, entry := range entries {
		updatedIdentity := entry.getIdentity()
		entryMap[updatedIdentity.key()] = entry
	}
	return entryMap
}

// FindUpdates returns changes in updated entries compared to the identities
//
// expects identities and entries to be sorted by timestamp
//
// the returned newer are entries that are newer than the first identity
// the returned mixed are entries that are older than the first identity (sorted by timestamp)
// the returned removed are identities that are not present in the updated entries (sorted by timestamp)
//
// implementation assumes the order of each identity doesn't change from old state (identities) and new state (updated); we have either add or removed.
func findUpdates(identities []EntryIdentity, updated []Entry) (new []mixedIdentityResult, removed []EntryIdentity) {

	idsMap := entryIdsToMap(identities)
	updatedMap := entriesToMap(updated)

	for newIndex, entry := range updated {
		id := entry.getIdentity()
		if _, found := idsMap[id.key()]; !found {
			new = append(new, mixedIdentityResult{
				newPos: newIndex,
				id:     id,
			})
		}
	}

	// Account for new entries
	for i := 0; i < len(identities); i++ {
		id := identities[i]
		if _, found := updatedMap[id.key()]; !found {
			removed = append(removed, id)
		}
	}
	return
}
