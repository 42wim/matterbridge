package store

import (
	"context"

	"go.mau.fi/libsignal/keys/identity"
	"go.mau.fi/libsignal/protocol"
)

// IdentityKey provides an interface to identity information.
type IdentityKey interface {
	// Get the local client's identity key pair.
	GetIdentityKeyPair() *identity.KeyPair

	// Return the local client's registration ID.
	//
	// Clients should maintain a registration ID, a random number between 1 and 16380
	// that's generated once at install time.
	GetLocalRegistrationID() uint32

	// Save a remote client's identity key in our identity store.
	SaveIdentity(ctx context.Context, address *protocol.SignalAddress, identityKey *identity.Key) error

	// Verify a remote client's identity key.
	//
	// Determine whether a remote client's identity is trusted. Trust is based on
	// 'trust on first use'. This means that an identity key is considered 'trusted'
	// if there is no entry for the recipient in the local store, or if it matches the
	// saved key for a recipient in the local store.
	IsTrustedIdentity(ctx context.Context, address *protocol.SignalAddress, identityKey *identity.Key) (bool, error)
}
