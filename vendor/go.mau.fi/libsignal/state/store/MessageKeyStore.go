package store

import (
	"context"

	"go.mau.fi/libsignal/keys/message"
)

// MessageKey store is an interface describing the optional local storage
// of message keys.
type MessageKey interface {
	// Load a local message key by id
	LoadMessageKey(ctx context.Context, keyID uint32) (*message.Keys, error)

	// Store a local message key
	StoreMessageKey(ctx context.Context, keyID uint32, key *message.Keys) error

	// Check to see if the store contains a message key with id.
	ContainsMessageKey(ctx context.Context, keyID uint32) (bool, error)

	// Delete a message key from local storage.
	RemoveMessageKey(ctx context.Context, keyID uint32) error
}
