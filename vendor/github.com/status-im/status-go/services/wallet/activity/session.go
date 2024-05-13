package activity

import (
	"context"
	"errors"
	"strconv"
	"time"

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
	id          common.MultiTransactionIDType
}

func (e EntryIdentity) same(a EntryIdentity) bool {
	return a.payloadType == e.payloadType &&
		((a.transaction == nil && e.transaction == nil) ||
			(a.transaction.ChainID == e.transaction.ChainID &&
				a.transaction.Hash == e.transaction.Hash &&
				a.transaction.Address == e.transaction.Address)) &&
		a.id == e.id
}

func (e EntryIdentity) key() string {
	txID := nilStr
	if e.transaction != nil {
		txID = strconv.FormatUint(uint64(e.transaction.ChainID), 10) + e.transaction.Hash.Hex() + e.transaction.Address.Hex()
	}
	return strconv.Itoa(e.payloadType) + txID + strconv.FormatInt(int64(e.id), 16)
}

type SessionID int32

// Session stores state related to a filter session
// The user happy flow is:
// 1. StartFilterSession to get a new SessionID and client be notified by the current state
// 2. GetMoreForFilterSession anytime to get more entries after the first page
// 3. UpdateFilterForSession to update the filter and get the new state or clean the filter and get the newer entries
// 4. ResetFilterSession in case client receives SessionUpdate with HasNewOnTop = true to get the latest state
// 5. StopFilterSession to stop the session when no used (user changed from activity screens or changed addresses and chains)
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
	// noFilterModel is a mirror of the data model presentation has when filter is empty
	noFilterModel map[string]EntryIdentity
	// new holds the new entries until user requests update by calling ResetFilterSession
	new []EntryIdentity
}

type EntryUpdate struct {
	Pos   int    `json:"pos"`
	Entry *Entry `json:"entry"`
}

// SessionUpdate payload for EventActivitySessionUpdated
type SessionUpdate struct {
	HasNewOnTop *bool           `json:"hasNewOnTop,omitempty"`
	New         []*EntryUpdate  `json:"new,omitempty"`
	Removed     []EntryIdentity `json:"removed,omitempty"`
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

// mirrorIdentities for update use
func mirrorIdentities(entries []Entry) []EntryIdentity {
	model := make([]EntryIdentity, 0, len(entries))
	for _, a := range entries {
		model = append(model, EntryIdentity{
			payloadType: a.payloadType,
			transaction: a.transaction,
			id:          a.id,
		})
	}
	return model
}

func (s *Service) internalFilterForSession(session *Session, firstPageCount int) {
	s.internalFilter(
		fullFilterParams{
			sessionID:    session.id,
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

			session.model = mirrorIdentities(entries)

			return 0
		},
	)
}

func (s *Service) StartFilterSession(addresses []eth.Address, allAddresses bool, chainIDs []common.ChainID, filter Filter, firstPageCount int) SessionID {
	sessionID := s.nextSessionID()

	session := &Session{
		id: sessionID,

		addresses:    addresses,
		allAddresses: allAddresses,
		chainIDs:     chainIDs,
		filter:       filter,

		model: make([]EntryIdentity, 0, firstPageCount),
	}

	s.sessionsRWMutex.Lock()
	subscribeToEvents := len(s.sessions) == 0

	s.sessions[sessionID] = session

	if subscribeToEvents {
		s.subscribeToEvents()
	}
	s.sessionsRWMutex.Unlock()

	s.internalFilterForSession(session, firstPageCount)

	return sessionID
}

// UpdateFilterForSession is to be called for updating the filter of a specific session
// After calling this method to set a filter all the incoming changes will be reported with
// Entry.isNew = true when filter is reset to empty
func (s *Service) UpdateFilterForSession(id SessionID, filter Filter, firstPageCount int) error {
	s.sessionsRWMutex.RLock()
	session, found := s.sessions[id]
	if !found {
		s.sessionsRWMutex.RUnlock()
		return errors.New("session not found")
	}

	prevFilterEmpty := session.filter.IsEmpty()
	newFilerEmpty := filter.IsEmpty()
	s.sessionsRWMutex.RUnlock()

	s.sessionsRWMutex.Lock()

	session.new = nil

	session.filter = filter

	if prevFilterEmpty && !newFilerEmpty {
		// Session is moving from empty to non-empty filter
		// Take a snapshot of the current model
		session.noFilterModel = entryIdsToMap(session.model)

		session.model = make([]EntryIdentity, 0, firstPageCount)

		// In this case there is nothing to flag so we request the first page
		s.internalFilterForSession(session, firstPageCount)
	} else if !prevFilterEmpty && newFilerEmpty {
		// Session is moving from non-empty to empty filter
		// In this case we need to flag all the new entries that are not in the noFilterModel
		s.internalFilter(
			fullFilterParams{
				sessionID:    session.id,
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
				for i, a := range entries {
					_, found := session.noFilterModel[a.getIdentity().key()]
					entries[i].isNew = !found
				}

				// Mirror identities for update use
				session.model = mirrorIdentities(entries)
				session.noFilterModel = nil
				return 0
			},
		)
	} else {
		// Else act as a normal filter update
		s.internalFilterForSession(session, firstPageCount)
	}
	s.sessionsRWMutex.Unlock()

	return nil
}

// ResetFilterSession is to be called when SessionUpdate.HasNewOnTop == true to
// update client with the latest state including new on top entries
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

			if session.noFilterModel != nil {
				// Add reported new entries to mark them as seen
				for _, a := range newMap {
					session.noFilterModel[a.key()] = a
				}
			}

			// Mirror client identities for checking updates
			session.model = mirrorIdentities(entries)

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

// processEvents runs only if more than one session is active
func (s *Service) processEvents() {
	eventCount := 0
	lastUpdate := time.Now().UnixMilli()
	for event := range s.ch {
		if event.Type == transactions.EventPendingTransactionUpdate ||
			event.Type == transactions.EventPendingTransactionStatusChanged ||
			event.Type == transfer.EventNewTransfers {
			eventCount++
		}
		// debounce events updates
		if eventCount > 0 &&
			(time.Duration(time.Now().UnixMilli()-lastUpdate)*time.Millisecond) >= s.debounceDuration {
			s.detectNew(eventCount)
			eventCount = 0
			lastUpdate = time.Now().UnixMilli()
		}
	}
}

func (s *Service) detectNew(changeCount int) {
	for sessionID := range s.sessions {
		session := s.sessions[sessionID]

		fetchLen := len(session.model) + changeCount
		activities, err := getActivityEntries(context.Background(), s.getDeps(), session.addresses, session.allAddresses, session.chainIDs, session.filter, 0, fetchLen)
		if err != nil {
			log.Error("Error getting activity entries", "error", err)
			continue
		}

		s.sessionsRWMutex.RLock()
		allData := append(session.new, session.model...)
		new, _ /*removed*/ := findUpdates(allData, activities)
		s.sessionsRWMutex.RUnlock()

		s.sessionsRWMutex.Lock()
		lastProcessed := -1
		onTop := true
		var mixed []*EntryUpdate
		for i, idRes := range new {
			// Detect on top
			if onTop {
				// mixedIdentityResult.newPos includes session.new, therefore compensate for it
				if ((idRes.newPos - len(session.new)) - lastProcessed) > 1 {
					// From now on the events are not on top and continuous but mixed between existing entries
					onTop = false
					mixed = make([]*EntryUpdate, 0, len(new)-i)
				}
				lastProcessed = idRes.newPos
			}

			if onTop {
				if session.new == nil {
					session.new = make([]EntryIdentity, 0, len(new))
				}
				session.new = append(session.new, idRes.id)
			} else {
				modelPos := idRes.newPos - len(session.new)
				entry := activities[idRes.newPos]
				entry.isNew = true
				mixed = append(mixed, &EntryUpdate{
					Pos:   modelPos,
					Entry: &entry,
				})
				// Insert in session model at modelPos index
				session.model = append(session.model[:modelPos], append([]EntryIdentity{{payloadType: entry.payloadType, transaction: entry.transaction, id: entry.id}}, session.model[modelPos:]...)...)
			}
		}

		s.sessionsRWMutex.Unlock()

		if len(session.new) > 0 || len(mixed) > 0 {
			go notify(s.eventFeed, sessionID, len(session.new) > 0, mixed)
		}
	}
}

func notify(eventFeed *event.Feed, id SessionID, hasNewOnTop bool, mixed []*EntryUpdate) {
	payload := SessionUpdate{
		New: mixed,
	}

	if hasNewOnTop {
		payload.HasNewOnTop = &hasNewOnTop
	}

	sendResponseEvent(eventFeed, (*int32)(&id), EventActivitySessionUpdated, payload, nil)
}

// unsubscribeFromEvents should be called with sessionsRWMutex locked for writing
func (s *Service) unsubscribeFromEvents() {
	s.subscriptions.Unsubscribe()
	close(s.ch)
	s.ch = nil
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
	if len(updated) == 0 {
		return
	}

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

		if len(identities) > 0 && entry.getIdentity().same(identities[len(identities)-1]) {
			break
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
