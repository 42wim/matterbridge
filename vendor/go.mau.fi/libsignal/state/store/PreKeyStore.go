package store

import (
	"context"

	"go.mau.fi/libsignal/state/record"
)

// PreKey store is an interface describing the local storage
// of PreKeyRecords
type PreKey interface {
	// Load a local PreKeyRecord
	LoadPreKey(ctx context.Context, preKeyID uint32) (*record.PreKey, error)

	// Store a local PreKeyRecord
	StorePreKey(ctx context.Context, preKeyID uint32, preKeyRecord *record.PreKey) error

	// Check to see if the store contains a PreKeyRecord
	ContainsPreKey(ctx context.Context, preKeyID uint32) (bool, error)

	// Delete a PreKeyRecord from local storage.
	RemovePreKey(ctx context.Context, preKeyID uint32) error
}
