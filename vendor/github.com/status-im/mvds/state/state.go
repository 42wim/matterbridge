// Package state contains everything related to the synchronization state for MVDS.
package state

// RecordType is the type for a specific record, either `OFFER`, `REQUEST` or `MESSAGE`.
type RecordType int

const (
	OFFER RecordType = iota
	REQUEST
	MESSAGE
)

// State is a struct used to store a records [state](https://github.com/status-im/bigbrother-specs/blob/master/data_sync/mvds.md#state).
type State struct {
	Type      RecordType
	SendCount uint64
	SendEpoch int64
	// GroupID is optional, thus nullable
	GroupID   *GroupID
	PeerID    PeerID
	MessageID MessageID
}

type SyncState interface {
	Add(newState State) error
	Remove(id MessageID, peer PeerID) error
	All(epoch int64) ([]State, error)
	Map(epoch int64, process func(State) State) error
}
