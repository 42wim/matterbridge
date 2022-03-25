package ratchet

import (
	"go.mau.fi/libsignal/ecc"
	"go.mau.fi/libsignal/keys/identity"
)

// NewSenderParameters creates a structure with all the keys needed to construct
// a new session when we are sending a message to a recipient for the first time.
func NewSenderParameters(ourIdentityKey *identity.KeyPair, ourBaseKey *ecc.ECKeyPair,
	theirIdentityKey *identity.Key, theirSignedPreKey ecc.ECPublicKeyable,
	theirRatchetKey ecc.ECPublicKeyable, theirOneTimePreKey ecc.ECPublicKeyable) *SenderParameters {

	senderParams := SenderParameters{
		ourIdentityKeyPair: ourIdentityKey,
		ourBaseKey:         ourBaseKey,
		theirIdentityKey:   theirIdentityKey,
		theirSignedPreKey:  theirSignedPreKey,
		theirOneTimePreKey: theirOneTimePreKey,
		theirRatchetKey:    theirRatchetKey,
	}

	return &senderParams
}

// NewEmptySenderParameters creates an empty structure with the sender parameters
// needed to create a session. You should use the `set` functions to set all the
// necessary keys needed to build a session.
func NewEmptySenderParameters() *SenderParameters {
	senderParams := SenderParameters{}

	return &senderParams
}

// SenderParameters describes the session parameters if we are sending the
// recipient a message for the first time. These parameters are used as the
// basis for deriving a shared secret with a recipient.
type SenderParameters struct {
	ourIdentityKeyPair *identity.KeyPair
	ourBaseKey         *ecc.ECKeyPair

	theirIdentityKey   *identity.Key
	theirSignedPreKey  ecc.ECPublicKeyable
	theirOneTimePreKey ecc.ECPublicKeyable
	theirRatchetKey    ecc.ECPublicKeyable
}

// OurIdentityKey returns the identity key pair of the sender.
func (s *SenderParameters) OurIdentityKey() *identity.KeyPair {
	return s.ourIdentityKeyPair
}

// OurBaseKey returns the base ECC key pair of the sender.
func (s *SenderParameters) OurBaseKey() *ecc.ECKeyPair {
	return s.ourBaseKey
}

// TheirIdentityKey returns the identity public key of the receiver.
func (s *SenderParameters) TheirIdentityKey() *identity.Key {
	return s.theirIdentityKey
}

// TheirSignedPreKey returns the signed pre key of the receiver.
func (s *SenderParameters) TheirSignedPreKey() ecc.ECPublicKeyable {
	return s.theirSignedPreKey
}

// TheirOneTimePreKey returns the receiver's one time prekey.
func (s *SenderParameters) TheirOneTimePreKey() ecc.ECPublicKeyable {
	return s.theirOneTimePreKey
}

// TheirRatchetKey returns the receiver's ratchet key.
func (s *SenderParameters) TheirRatchetKey() ecc.ECPublicKeyable {
	return s.theirRatchetKey
}

// SetOurIdentityKey sets the identity key pair of the sender.
func (s *SenderParameters) SetOurIdentityKey(ourIdentityKey *identity.KeyPair) {
	s.ourIdentityKeyPair = ourIdentityKey
}

// SetOurBaseKey sets the base ECC key pair of the sender.
func (s *SenderParameters) SetOurBaseKey(ourBaseKey *ecc.ECKeyPair) {
	s.ourBaseKey = ourBaseKey
}

// SetTheirIdentityKey sets the identity public key of the receiver.
func (s *SenderParameters) SetTheirIdentityKey(theirIdentityKey *identity.Key) {
	s.theirIdentityKey = theirIdentityKey
}

// SetTheirSignedPreKey sets the signed pre key of the receiver.
func (s *SenderParameters) SetTheirSignedPreKey(theirSignedPreKey ecc.ECPublicKeyable) {
	s.theirSignedPreKey = theirSignedPreKey
}

// SetTheirOneTimePreKey sets the receiver's one time prekey.
func (s *SenderParameters) SetTheirOneTimePreKey(theirOneTimePreKey ecc.ECPublicKeyable) {
	s.theirOneTimePreKey = theirOneTimePreKey
}

// SetTheirRatchetKey sets the receiver's ratchet key.
func (s *SenderParameters) SetTheirRatchetKey(theirRatchetKey ecc.ECPublicKeyable) {
	s.theirRatchetKey = theirRatchetKey
}
