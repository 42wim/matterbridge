package ratchet

import (
	"go.mau.fi/libsignal/ecc"
	"go.mau.fi/libsignal/keys/identity"
)

// NewReceiverParameters creates a structure with all the keys needed to construct
// a new session when we are receiving a message from a user for the first time.
func NewReceiverParameters(ourIdentityKey *identity.KeyPair, ourSignedPreKey *ecc.ECKeyPair,
	ourOneTimePreKey *ecc.ECKeyPair, ourRatchetKey *ecc.ECKeyPair,
	theirBaseKey ecc.ECPublicKeyable, theirIdentityKey *identity.Key) *ReceiverParameters {

	receiverParams := ReceiverParameters{
		ourIdentityKeyPair: ourIdentityKey,
		ourSignedPreKey:    ourSignedPreKey,
		ourOneTimePreKey:   ourOneTimePreKey,
		ourRatchetKey:      ourRatchetKey,
		theirBaseKey:       theirBaseKey,
		theirIdentityKey:   theirIdentityKey,
	}

	return &receiverParams
}

// NewEmptyReceiverParameters creates an empty structure with the receiver parameters
// needed to create a session. You should use the `set` functions to set all the
// necessary keys needed to build a session.
func NewEmptyReceiverParameters() *ReceiverParameters {
	receiverParams := ReceiverParameters{}

	return &receiverParams
}

// ReceiverParameters describes the session parameters if we are receiving
// a message from someone for the first time. These parameters are used as
// the basis for deriving a shared secret with the sender.
type ReceiverParameters struct {
	ourIdentityKeyPair *identity.KeyPair
	ourSignedPreKey    *ecc.ECKeyPair
	ourOneTimePreKey   *ecc.ECKeyPair
	ourRatchetKey      *ecc.ECKeyPair

	theirBaseKey     ecc.ECPublicKeyable
	theirIdentityKey *identity.Key
}

// OurIdentityKeyPair returns the identity key of the receiver.
func (r *ReceiverParameters) OurIdentityKeyPair() *identity.KeyPair {
	return r.ourIdentityKeyPair
}

// OurSignedPreKey returns the signed prekey of the receiver.
func (r *ReceiverParameters) OurSignedPreKey() *ecc.ECKeyPair {
	return r.ourSignedPreKey
}

// OurOneTimePreKey returns the one time prekey of the receiver.
func (r *ReceiverParameters) OurOneTimePreKey() *ecc.ECKeyPair {
	return r.ourOneTimePreKey
}

// OurRatchetKey returns the ratchet key of the receiver.
func (r *ReceiverParameters) OurRatchetKey() *ecc.ECKeyPair {
	return r.ourRatchetKey
}

// TheirBaseKey returns the base key of the sender.
func (r *ReceiverParameters) TheirBaseKey() ecc.ECPublicKeyable {
	return r.theirBaseKey
}

// TheirIdentityKey returns the identity key of the sender.
func (r *ReceiverParameters) TheirIdentityKey() *identity.Key {
	return r.theirIdentityKey
}

// SetOurIdentityKeyPair sets the identity key of the receiver.
func (r *ReceiverParameters) SetOurIdentityKeyPair(ourIdentityKey *identity.KeyPair) {
	r.ourIdentityKeyPair = ourIdentityKey
}

// SetOurSignedPreKey sets the signed prekey of the receiver.
func (r *ReceiverParameters) SetOurSignedPreKey(ourSignedPreKey *ecc.ECKeyPair) {
	r.ourSignedPreKey = ourSignedPreKey
}

// SetOurOneTimePreKey sets the one time prekey of the receiver.
func (r *ReceiverParameters) SetOurOneTimePreKey(ourOneTimePreKey *ecc.ECKeyPair) {
	r.ourOneTimePreKey = ourOneTimePreKey
}

// SetOurRatchetKey sets the ratchet key of the receiver.
func (r *ReceiverParameters) SetOurRatchetKey(ourRatchetKey *ecc.ECKeyPair) {
	r.ourRatchetKey = ourRatchetKey
}

// SetTheirBaseKey sets the base key of the sender.
func (r *ReceiverParameters) SetTheirBaseKey(theirBaseKey ecc.ECPublicKeyable) {
	r.theirBaseKey = theirBaseKey
}

// SetTheirIdentityKey sets the identity key of the sender.
func (r *ReceiverParameters) SetTheirIdentityKey(theirIdentityKey *identity.Key) {
	r.theirIdentityKey = theirIdentityKey
}
