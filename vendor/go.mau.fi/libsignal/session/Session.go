// Package session provides the methods necessary to build sessions
package session

import (
	"context"
	"fmt"

	"go.mau.fi/libsignal/ecc"
	"go.mau.fi/libsignal/keys/prekey"
	"go.mau.fi/libsignal/logger"
	"go.mau.fi/libsignal/protocol"
	"go.mau.fi/libsignal/ratchet"
	"go.mau.fi/libsignal/serialize"
	"go.mau.fi/libsignal/signalerror"
	"go.mau.fi/libsignal/state/record"
	"go.mau.fi/libsignal/state/store"
	"go.mau.fi/libsignal/util/medium"
	"go.mau.fi/libsignal/util/optional"
)

// NewBuilder constructs a session builder.
func NewBuilder(sessionStore store.Session, preKeyStore store.PreKey,
	signedStore store.SignedPreKey, identityStore store.IdentityKey,
	remoteAddress *protocol.SignalAddress, serializer *serialize.Serializer) *Builder {

	builder := Builder{
		sessionStore:      sessionStore,
		preKeyStore:       preKeyStore,
		signedPreKeyStore: signedStore,
		identityKeyStore:  identityStore,
		remoteAddress:     remoteAddress,
		serializer:        serializer,
	}

	return &builder
}

// NewBuilderFromSignal Store constructs a session builder using a
// SignalProtocol Store.
func NewBuilderFromSignal(signalStore store.SignalProtocol,
	remoteAddress *protocol.SignalAddress, serializer *serialize.Serializer) *Builder {

	builder := Builder{
		sessionStore:      signalStore,
		preKeyStore:       signalStore,
		signedPreKeyStore: signalStore,
		identityKeyStore:  signalStore,
		remoteAddress:     remoteAddress,
		serializer:        serializer,
	}

	return &builder
}

// Builder is responsible for setting up encrypted sessions.
// Once a session has been established, SessionCipher can be
// used to encrypt/decrypt messages in that session.
//
// Sessions are built from one of three different vectors:
//   - PreKeyBundle retrieved from a server.
//   - PreKeySignalMessage received from a client.
//   - KeyExchangeMessage sent to or received from a client.
//
// Sessions are constructed per recipientId + deviceId tuple.
// Remote logical users are identified by their recipientId,
// and each logical recipientId can have multiple physical
// devices.
type Builder struct {
	sessionStore      store.Session
	preKeyStore       store.PreKey
	signedPreKeyStore store.SignedPreKey
	identityKeyStore  store.IdentityKey
	remoteAddress     *protocol.SignalAddress
	serializer        *serialize.Serializer
}

// Process builds a new session from a session record and pre
// key signal message.
func (b *Builder) Process(ctx context.Context, sessionRecord *record.Session, message *protocol.PreKeySignalMessage) (unsignedPreKeyID *optional.Uint32, err error) {

	// Check to see if the keys are trusted.
	theirIdentityKey := message.IdentityKey()
	trusted, err := b.identityKeyStore.IsTrustedIdentity(ctx, b.remoteAddress, theirIdentityKey)
	if err != nil {
		return nil, err
	}
	if !trusted {
		return nil, signalerror.ErrUntrustedIdentity
	}

	// Use version 3 of the signal/axolotl protocol.
	unsignedPreKeyID, err = b.processV3(ctx, sessionRecord, message)
	if err != nil {
		return nil, err
	}

	// Save the identity key to our identity store.
	if err := b.identityKeyStore.SaveIdentity(ctx, b.remoteAddress, theirIdentityKey); err != nil {
		return nil, err
	}

	// Return the unsignedPreKeyID
	return unsignedPreKeyID, nil
}

// ProcessV3 builds a new session from a session record and pre key
// signal message. After a session is constructed in this way, the embedded
// SignalMessage can be decrypted.
func (b *Builder) processV3(ctx context.Context, sessionRecord *record.Session,
	message *protocol.PreKeySignalMessage) (unsignedPreKeyID *optional.Uint32, err error) {

	logger.Debug("Processing message with PreKeyID: ", message.PreKeyID())
	// Check to see if we've already set up a session for this V3 message.
	sessionExists := sessionRecord.HasSessionState(
		message.MessageVersion(),
		message.BaseKey().Serialize(),
	)
	if sessionExists {
		logger.Debug("We've already setup a session for this V3 message, letting bundled message fall through...")
		return optional.NewEmptyUint32(), nil
	}

	// Load our signed prekey from our signed prekey store.
	ourSignedPreKeyRecord, err := b.signedPreKeyStore.LoadSignedPreKey(ctx, message.SignedPreKeyID())
	if err != nil {
		return nil, err
	}
	if ourSignedPreKeyRecord == nil {
		return nil, fmt.Errorf("%w with ID %d", signalerror.ErrNoSignedPreKey, message.SignedPreKeyID())
	}
	ourSignedPreKey := ourSignedPreKeyRecord.KeyPair()

	// Build the parameters of the session.
	parameters := ratchet.NewEmptyReceiverParameters()
	parameters.SetTheirBaseKey(message.BaseKey())
	parameters.SetTheirIdentityKey(message.IdentityKey())
	parameters.SetOurIdentityKeyPair(b.identityKeyStore.GetIdentityKeyPair())
	parameters.SetOurSignedPreKey(ourSignedPreKey)
	parameters.SetOurRatchetKey(ourSignedPreKey)

	// Set our one time pre key with the one from our prekey store
	// if the message contains a valid pre key id
	if !message.PreKeyID().IsEmpty {
		oneTimePreKey, err := b.preKeyStore.LoadPreKey(ctx, message.PreKeyID().Value)
		if err != nil {
			return nil, err
		}
		if oneTimePreKey == nil {
			return nil, fmt.Errorf("%w with ID %d", signalerror.ErrNoOneTimeKeyFound, message.PreKeyID().Value)
		}
		parameters.SetOurOneTimePreKey(oneTimePreKey.KeyPair())
	} else {
		parameters.SetOurOneTimePreKey(nil)
	}

	// If this is a fresh record, archive our current state.
	if !sessionRecord.IsFresh() {
		sessionRecord.ArchiveCurrentState()
	}

	///////// Initialize our session /////////
	sessionState := sessionRecord.SessionState()
	derivedKeys, sessionErr := ratchet.CalculateReceiverSession(parameters)
	if sessionErr != nil {
		return nil, sessionErr
	}
	sessionState.SetVersion(protocol.CurrentVersion)
	sessionState.SetRemoteIdentityKey(parameters.TheirIdentityKey())
	sessionState.SetLocalIdentityKey(parameters.OurIdentityKeyPair().PublicKey())
	sessionState.SetSenderChain(parameters.OurRatchetKey(), derivedKeys.ChainKey)
	sessionState.SetRootKey(derivedKeys.RootKey)

	// Set the session's registration ids and base key
	sessionState.SetLocalRegistrationID(b.identityKeyStore.GetLocalRegistrationID())
	sessionState.SetRemoteRegistrationID(message.RegistrationID())
	sessionState.SetSenderBaseKey(message.BaseKey().Serialize())

	// Remove the PreKey from our store and return the message prekey id if it is valid.
	if message.PreKeyID() != nil && message.PreKeyID().Value != medium.MaxValue {
		return message.PreKeyID(), nil
	}
	return optional.NewEmptyUint32(), nil
}

// ProcessBundle builds a new session from a PreKeyBundle retrieved
// from a server.
func (b *Builder) ProcessBundle(ctx context.Context, preKey *prekey.Bundle) error {
	// Check to see if the keys are trusted.
	trusted, err := b.identityKeyStore.IsTrustedIdentity(ctx, b.remoteAddress, preKey.IdentityKey())
	if err != nil {
		return err
	}
	if !trusted {
		return signalerror.ErrUntrustedIdentity
	}

	// Check to see if the bundle has a signed pre key.
	if preKey.SignedPreKey() == nil {
		return signalerror.ErrNoSignedPreKey
	}

	// Verify the signature of the pre key
	preKeyPublic := preKey.IdentityKey().PublicKey()
	preKeyBytes := preKey.SignedPreKey().Serialize()
	preKeySignature := preKey.SignedPreKeySignature()
	if !ecc.VerifySignature(preKeyPublic, preKeyBytes, preKeySignature) {
		return signalerror.ErrInvalidSignature
	}

	// Load our session and generate keys.
	sessionRecord, err := b.sessionStore.LoadSession(ctx, b.remoteAddress)
	if err != nil {
		return err
	}
	if sessionRecord == nil {
		return fmt.Errorf("LoadSession returned nil")
	}
	ourBaseKey, err := ecc.GenerateKeyPair()
	if err != nil {
		return err
	}
	theirSignedPreKey := preKey.SignedPreKey()
	theirOneTimePreKey := preKey.PreKey()
	theirOneTimePreKeyID := preKey.PreKeyID()

	// Build the parameters of the session
	parameters := ratchet.NewEmptySenderParameters()
	parameters.SetOurBaseKey(ourBaseKey)
	parameters.SetOurIdentityKey(b.identityKeyStore.GetIdentityKeyPair())
	parameters.SetTheirIdentityKey(preKey.IdentityKey())
	parameters.SetTheirSignedPreKey(theirSignedPreKey)
	parameters.SetTheirRatchetKey(theirSignedPreKey)
	parameters.SetTheirOneTimePreKey(theirOneTimePreKey)

	// If this is a fresh record, archive our current state.
	if !sessionRecord.IsFresh() {
		sessionRecord.ArchiveCurrentState()
	}

	///////// Initialize our session /////////
	sessionState := sessionRecord.SessionState()
	derivedKeys, sessionErr := ratchet.CalculateSenderSession(parameters)
	if sessionErr != nil {
		return sessionErr
	}
	// Generate an ephemeral "ratchet" key that will be advertised to
	// the receiving user.
	sendingRatchetKey, keyErr := ecc.GenerateKeyPair()
	if keyErr != nil {
		return keyErr
	}
	sendingChain, chainErr := derivedKeys.RootKey.CreateChain(
		parameters.TheirRatchetKey(),
		sendingRatchetKey,
	)
	if chainErr != nil {
		return chainErr
	}

	// Calculate the sender session.
	sessionState.SetVersion(protocol.CurrentVersion)
	sessionState.SetRemoteIdentityKey(parameters.TheirIdentityKey())
	sessionState.SetLocalIdentityKey(parameters.OurIdentityKey().PublicKey())
	sessionState.AddReceiverChain(parameters.TheirRatchetKey(), derivedKeys.ChainKey.Current())
	sessionState.SetSenderChain(sendingRatchetKey, sendingChain.ChainKey)
	sessionState.SetRootKey(sendingChain.RootKey)

	// Update our session record with the unackowledged prekey message
	sessionState.SetUnacknowledgedPreKeyMessage(
		theirOneTimePreKeyID,
		preKey.SignedPreKeyID(),
		ourBaseKey.PublicKey(),
	)

	// Set the local registration ID based on the registration id in our identity key store.
	sessionState.SetLocalRegistrationID(
		b.identityKeyStore.GetLocalRegistrationID(),
	)

	// Set the remote registration ID based on the given prekey bundle registrationID.
	sessionState.SetRemoteRegistrationID(
		preKey.RegistrationID(),
	)

	// Set the sender base key in our session record state.
	sessionState.SetSenderBaseKey(
		ourBaseKey.PublicKey().Serialize(),
	)

	// Store the session in our session store and save the identity in our identity store.
	if err := b.sessionStore.StoreSession(ctx, b.remoteAddress, sessionRecord); err != nil {
		return err
	}
	if err := b.identityKeyStore.SaveIdentity(ctx, b.remoteAddress, preKey.IdentityKey()); err != nil {
		return err
	}

	return nil
}
