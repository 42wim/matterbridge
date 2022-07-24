package mautrix

import (
	"maunium.net/go/mautrix/id"
)

// SyncStore is an interface which must be satisfied to store client data.
//
// You can either write a struct which persists this data to disk, or you can use the
// provided "MemorySyncStore" which just keeps data around in-memory which is lost on
// restarts.
type SyncStore interface {
	SaveFilterID(userID id.UserID, filterID string)
	LoadFilterID(userID id.UserID) string
	SaveNextBatch(userID id.UserID, nextBatchToken string)
	LoadNextBatch(userID id.UserID) string
}

// Deprecated: renamed to SyncStore
type Storer = SyncStore

// MemorySyncStore implements the Storer interface.
//
// Everything is persisted in-memory as maps. It is not safe to load/save filter IDs
// or next batch tokens on any goroutine other than the syncing goroutine: the one
// which called Client.Sync().
type MemorySyncStore struct {
	Filters   map[id.UserID]string
	NextBatch map[id.UserID]string
}

// SaveFilterID to memory.
func (s *MemorySyncStore) SaveFilterID(userID id.UserID, filterID string) {
	s.Filters[userID] = filterID
}

// LoadFilterID from memory.
func (s *MemorySyncStore) LoadFilterID(userID id.UserID) string {
	return s.Filters[userID]
}

// SaveNextBatch to memory.
func (s *MemorySyncStore) SaveNextBatch(userID id.UserID, nextBatchToken string) {
	s.NextBatch[userID] = nextBatchToken
}

// LoadNextBatch from memory.
func (s *MemorySyncStore) LoadNextBatch(userID id.UserID) string {
	return s.NextBatch[userID]
}

// NewMemorySyncStore constructs a new MemorySyncStore.
func NewMemorySyncStore() *MemorySyncStore {
	return &MemorySyncStore{
		Filters:   make(map[id.UserID]string),
		NextBatch: make(map[id.UserID]string),
	}
}

// AccountDataStore uses account data to store the next batch token, and stores the filter ID in memory
// (as filters can be safely recreated every startup).
type AccountDataStore struct {
	FilterID  string
	EventType string
	client    *Client
}

type accountData struct {
	NextBatch string `json:"next_batch"`
}

func (s *AccountDataStore) SaveFilterID(userID id.UserID, filterID string) {
	if userID.String() != s.client.UserID.String() {
		panic("AccountDataStore must only be used with a single account")
	}
	s.FilterID = filterID
}

func (s *AccountDataStore) LoadFilterID(userID id.UserID) string {
	if userID.String() != s.client.UserID.String() {
		panic("AccountDataStore must only be used with a single account")
	}
	return s.FilterID
}

func (s *AccountDataStore) SaveNextBatch(userID id.UserID, nextBatchToken string) {
	if userID.String() != s.client.UserID.String() {
		panic("AccountDataStore must only be used with a single account")
	}

	data := accountData{
		NextBatch: nextBatchToken,
	}

	err := s.client.SetAccountData(s.EventType, data)
	if err != nil {
		s.client.Log.Warn().Err(err).Msg("Failed to save next batch token to account data")
	}
}

func (s *AccountDataStore) LoadNextBatch(userID id.UserID) string {
	if userID.String() != s.client.UserID.String() {
		panic("AccountDataStore must only be used with a single account")
	}

	data := &accountData{}

	err := s.client.GetAccountData(s.EventType, data)
	if err != nil {
		s.client.Log.Warn().Err(err).Msg("Failed to load next batch token from account data")
		return ""
	}

	return data.NextBatch
}

// NewAccountDataStore returns a new AccountDataStore, which stores
// the next_batch token as a custom event in account data in the
// homeserver.
//
// AccountDataStore is only appropriate for bots, not appservices.
//
// The event type should be a reversed DNS name like tld.domain.sub.internal and
// must be unique for a client. The data stored in it is considered internal
// and must not be modified through outside means. You should also add a filter
// for account data changes of this event type, to avoid ending up in a sync
// loop:
//
//	filter := mautrix.Filter{
//		AccountData: mautrix.FilterPart{
//			Limit: 20,
//			NotTypes: []event.Type{
//				event.NewEventType(eventType),
//			},
//		},
//	}
//	// If you use a custom Syncer, set the filter there, not like this
//	client.Syncer.(*mautrix.DefaultSyncer).FilterJSON = &filter
//	client.Store = mautrix.NewAccountDataStore("com.example.mybot.store", client)
//	go func() {
//		err := client.Sync()
//		// don't forget to check err
//	}()
func NewAccountDataStore(eventType string, client *Client) *AccountDataStore {
	return &AccountDataStore{
		EventType: eventType,
		client:    client,
	}
}
