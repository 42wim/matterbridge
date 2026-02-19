package store

import (
	"context"

	"go.mau.fi/libsignal/state/record"
)

// SignedPreKey store is an interface that describes how to persistently
// store signed PreKeys.
type SignedPreKey interface {
	// LoadSignedPreKey loads a local SignedPreKeyRecord
	LoadSignedPreKey(ctx context.Context, signedPreKeyID uint32) (*record.SignedPreKey, error)

	// LoadSignedPreKeys loads all local SignedPreKeyRecords
	LoadSignedPreKeys(ctx context.Context) ([]*record.SignedPreKey, error)

	// Store a local SignedPreKeyRecord
	StoreSignedPreKey(ctx context.Context, signedPreKeyID uint32, record *record.SignedPreKey) error

	// Check to see if store contains the given record
	ContainsSignedPreKey(ctx context.Context, signedPreKeyID uint32) (bool, error)

	// Delete a SignedPreKeyRecord from local storage
	RemoveSignedPreKey(ctx context.Context, signedPreKeyID uint32) error
}
