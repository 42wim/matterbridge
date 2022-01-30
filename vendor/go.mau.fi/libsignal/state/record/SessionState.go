package record

import (
	"go.mau.fi/libsignal/ecc"
	"go.mau.fi/libsignal/kdf"
	"go.mau.fi/libsignal/keys/chain"
	"go.mau.fi/libsignal/keys/identity"
	"go.mau.fi/libsignal/keys/message"
	"go.mau.fi/libsignal/keys/root"
	"go.mau.fi/libsignal/keys/session"
	"go.mau.fi/libsignal/logger"
	"go.mau.fi/libsignal/util/errorhelper"
	"go.mau.fi/libsignal/util/optional"
)

const maxMessageKeys int = 2000
const maxReceiverChains int = 5

// StateSerializer is an interface for serializing and deserializing
// a Signal State into bytes. An implementation of this interface should be
// used to encode/decode the object into JSON, Protobuffers, etc.
type StateSerializer interface {
	Serialize(state *StateStructure) []byte
	Deserialize(serialized []byte) (*StateStructure, error)
}

// NewStateFromBytes will return a Signal State from the given
// bytes using the given serializer.
func NewStateFromBytes(serialized []byte, serializer StateSerializer) (*State, error) {
	// Use the given serializer to decode the signal message.
	stateStructure, err := serializer.Deserialize(serialized)
	if err != nil {
		return nil, err
	}

	return NewStateFromStructure(stateStructure, serializer)
}

// NewState returns a new session state.
func NewState(serializer StateSerializer) *State {
	return &State{serializer: serializer}
}

// NewStateFromStructure will return a new session state with the
// given state structure.
func NewStateFromStructure(structure *StateStructure, serializer StateSerializer) (*State, error) {
	// Keep a list of errors, so they can be handled once.
	errors := errorhelper.NewMultiError()

	// Convert our ecc keys from bytes into object form.
	localIdentityPublic, err := ecc.DecodePoint(structure.LocalIdentityPublic, 0)
	errors.Add(err)
	remoteIdentityPublic, err := ecc.DecodePoint(structure.RemoteIdentityPublic, 0)
	errors.Add(err)
	senderBaseKey, err := ecc.DecodePoint(structure.SenderBaseKey, 0)
	errors.Add(err)
	var pendingPreKey *PendingPreKey
	if structure.PendingPreKey != nil {
		pendingPreKey, err = NewPendingPreKeyFromStruct(structure.PendingPreKey)
		errors.Add(err)
	}
	senderChain, err := NewChainFromStructure(structure.SenderChain)
	errors.Add(err)

	// Build our receiver chains from structure.
	receiverChains := make([]*Chain, len(structure.ReceiverChains))
	for i := range structure.ReceiverChains {
		receiverChains[i], err = NewChainFromStructure(structure.ReceiverChains[i])
		errors.Add(err)
	}

	// Handle any errors. The first error will always be returned if there are multiple.
	if errors.HasErrors() {
		return nil, errors
	}

	// Build our state object.
	state := &State{
		localIdentityPublic:  identity.NewKey(localIdentityPublic),
		localRegistrationID:  structure.LocalRegistrationID,
		needsRefresh:         structure.NeedsRefresh,
		pendingKeyExchange:   NewPendingKeyExchangeFromStruct(structure.PendingKeyExchange),
		pendingPreKey:        pendingPreKey,
		previousCounter:      structure.PreviousCounter,
		receiverChains:       receiverChains,
		remoteIdentityPublic: identity.NewKey(remoteIdentityPublic),
		remoteRegistrationID: structure.RemoteRegistrationID,
		rootKey:              root.NewKey(kdf.DeriveSecrets, structure.RootKey),
		senderBaseKey:        senderBaseKey,
		senderChain:          senderChain,
		serializer:           serializer,
		sessionVersion:       structure.SessionVersion,
	}

	return state, nil
}

// StateStructure is the structure of a session state. Fields are public
// to be used for serialization and deserialization.
type StateStructure struct {
	LocalIdentityPublic  []byte
	LocalRegistrationID  uint32
	NeedsRefresh         bool
	PendingKeyExchange   *PendingKeyExchangeStructure
	PendingPreKey        *PendingPreKeyStructure
	PreviousCounter      uint32
	ReceiverChains       []*ChainStructure
	RemoteIdentityPublic []byte
	RemoteRegistrationID uint32
	RootKey              []byte
	SenderBaseKey        []byte
	SenderChain          *ChainStructure
	SessionVersion       int
}

// State is a session state that contains the structure for
// all sessions. Session states are contained inside session records.
// The session state is implemented as a struct rather than protobuffers
// to allow other serialization methods.
type State struct {
	localIdentityPublic  *identity.Key
	localRegistrationID  uint32
	needsRefresh         bool
	pendingKeyExchange   *PendingKeyExchange
	pendingPreKey        *PendingPreKey
	previousCounter      uint32
	receiverChains       []*Chain
	remoteIdentityPublic *identity.Key
	remoteRegistrationID uint32
	rootKey              *root.Key
	senderBaseKey        ecc.ECPublicKeyable
	senderChain          *Chain
	serializer           StateSerializer
	sessionVersion       int
}

// SenderBaseKey returns the sender's base key in bytes.
func (s *State) SenderBaseKey() []byte {
	if s.senderBaseKey == nil {
		return nil
	}
	return s.senderBaseKey.Serialize()
}

// SetSenderBaseKey sets the sender's base key with the given bytes.
func (s *State) SetSenderBaseKey(senderBaseKey []byte) {
	s.senderBaseKey, _ = ecc.DecodePoint(senderBaseKey, 0)
}

// Version returns the session's version.
func (s *State) Version() int {
	return s.sessionVersion
}

// SetVersion sets the session state's version number.
func (s *State) SetVersion(version int) {
	s.sessionVersion = version
}

// RemoteIdentityKey returns the identity key of the remote user.
func (s *State) RemoteIdentityKey() *identity.Key {
	return s.remoteIdentityPublic
}

// SetRemoteIdentityKey sets this session's identity key for the remote
// user.
func (s *State) SetRemoteIdentityKey(identityKey *identity.Key) {
	s.remoteIdentityPublic = identityKey
}

// LocalIdentityKey returns the session's identity key for the local
// user.
func (s *State) LocalIdentityKey() *identity.Key {
	return s.localIdentityPublic
}

// SetLocalIdentityKey sets the session's identity key for the local
// user.
func (s *State) SetLocalIdentityKey(identityKey *identity.Key) {
	s.localIdentityPublic = identityKey
}

// PreviousCounter returns the counter of the previous message.
func (s *State) PreviousCounter() uint32 {
	return s.previousCounter
}

// SetPreviousCounter sets the counter for the previous message.
func (s *State) SetPreviousCounter(previousCounter uint32) {
	s.previousCounter = previousCounter
}

// RootKey returns the root key for the session.
func (s *State) RootKey() session.RootKeyable {
	return s.rootKey
}

// SetRootKey sets the root key for the session.
func (s *State) SetRootKey(rootKey session.RootKeyable) {
	s.rootKey = rootKey.(*root.Key)
}

// SenderRatchetKey returns the public ratchet key of the sender.
func (s *State) SenderRatchetKey() ecc.ECPublicKeyable {
	return s.senderChain.senderRatchetKeyPair.PublicKey()
}

// SenderRatchetKeyPair returns the public/private ratchet key pair
// of the sender.
func (s *State) SenderRatchetKeyPair() *ecc.ECKeyPair {
	return s.senderChain.senderRatchetKeyPair
}

// HasReceiverChain will check to see if the session state has
// the given ephemeral key.
func (s *State) HasReceiverChain(senderEphemeral ecc.ECPublicKeyable) bool {
	return s.receiverChain(senderEphemeral) != nil
}

// HasSenderChain will check to see if the session state has a
// sender chain.
func (s *State) HasSenderChain() bool {
	return s.senderChain != nil
}

// receiverChain will loop through the session state's receiver chains
// and compare the given ephemeral key. If it is found, then the chain
// and index will be returned as a pair.
func (s *State) receiverChain(senderEphemeral ecc.ECPublicKeyable) *ReceiverChainPair {
	receiverChains := s.receiverChains

	for i, receiverChain := range receiverChains {
		chainSenderRatchetKey, err := ecc.DecodePoint(receiverChain.senderRatchetKeyPair.PublicKey().Serialize(), 0)
		if err != nil {
			logger.Error("Error getting receiverchain: ", err)
		}

		// If the chainSenderRatchetKey equals our senderEphemeral key, return it.
		if chainSenderRatchetKey.PublicKey() == senderEphemeral.PublicKey() {
			return NewReceiverChainPair(receiverChain, i)
		}
	}

	return nil
}

// ReceiverChainKey will use the given ephemeral key to generate a new
// chain key.
func (s *State) ReceiverChainKey(senderEphemeral ecc.ECPublicKeyable) *chain.Key {
	receiverChainAndIndex := s.receiverChain(senderEphemeral)
	receiverChain := receiverChainAndIndex.ReceiverChain

	if receiverChainAndIndex == nil || receiverChain == nil {
		return nil
	}

	return chain.NewKey(
		kdf.DeriveSecrets,
		receiverChain.chainKey.Key(),
		receiverChain.chainKey.Index(),
	)
}

// AddReceiverChain will add the given ratchet key and chain key to the session
// state.
func (s *State) AddReceiverChain(senderRatchetKey ecc.ECPublicKeyable, chainKey session.ChainKeyable) {
	// Create a keypair structure with our sender ratchet key.
	senderKey := ecc.NewECKeyPair(senderRatchetKey, nil)

	// Create a Chain state object that will hold our sender key, chain key, and
	// message keys.
	chain := NewChain(senderKey, chainKey.(*chain.Key), []*message.Keys{})

	// Add the Chain state to our list of receiver chain states.
	s.receiverChains = append(s.receiverChains, chain)

	// If our list of receiver chains is too big, delete the oldest entry.
	if len(s.receiverChains) > maxReceiverChains {
		i := 0
		s.receiverChains = append(s.receiverChains[:i], s.receiverChains[i+1:]...)
	}
}

// SetSenderChain will set the given ratchet key pair and chain key for this session
// state.
func (s *State) SetSenderChain(senderRatchetKeyPair *ecc.ECKeyPair, chainKey session.ChainKeyable) {
	// Create a Chain state object that will hold our sender key, chain key, and
	// message keys.
	chain := NewChain(senderRatchetKeyPair, chainKey.(*chain.Key), []*message.Keys{})

	// Set the sender chain.
	s.senderChain = chain
}

// SenderChainKey will return the chain key of the session state.
func (s *State) SenderChainKey() session.ChainKeyable {
	chainKey := s.senderChain.chainKey
	return chain.NewKey(kdf.DeriveSecrets, chainKey.Key(), chainKey.Index())
}

// SetSenderChainKey will set the chain key in the chain state for this session to
// the given chain key.
func (s *State) SetSenderChainKey(nextChainKey session.ChainKeyable) {
	senderChain := s.senderChain
	senderChain.SetChainKey(nextChainKey.(*chain.Key))
}

// HasMessageKeys returns true if we have message keys associated with the given
// sender key and counter.
func (s *State) HasMessageKeys(senderEphemeral ecc.ECPublicKeyable, counter uint32) bool {
	// Get our chain state that has our chain key.
	chainAndIndex := s.receiverChain(senderEphemeral)
	receiverChain := chainAndIndex.ReceiverChain

	// If the chain is empty, we don't have any message keys.
	if receiverChain == nil {
		return false
	}

	// Get our message keys from our receiver chain.
	messageKeyList := receiverChain.MessageKeys()

	// Loop through our message keys and compare its index with the
	// given counter.
	for _, messageKey := range messageKeyList {
		if messageKey.Index() == counter {
			return true
		}
	}

	return false
}

// RemoveMessageKeys removes the message key with the given sender key and
// counter. It will return the removed message key.
func (s *State) RemoveMessageKeys(senderEphemeral ecc.ECPublicKeyable, counter uint32) *message.Keys {
	// Get our chain state that has our chain key.
	chainAndIndex := s.receiverChain(senderEphemeral)
	chainKey := chainAndIndex.ReceiverChain

	// If the chain is empty, we don't have any message keys.
	if chainKey == nil {
		return nil
	}

	// Get our message keys from our receiver chain.
	messageKeyList := chainKey.MessageKeys()

	// Loop through our message keys and compare its index with the
	// given counter. When we find a match, remove it from our list.
	var rmIndex int
	for i, messageKey := range messageKeyList {
		if messageKey.Index() == counter {
			rmIndex = i
			break
		}
	}

	// Retrive the message key
	messageKey := chainKey.messageKeys[rmIndex]

	// Delete the message key from the given position.
	chainKey.messageKeys = append(chainKey.messageKeys[:rmIndex], chainKey.messageKeys[rmIndex+1:]...)

	return message.NewKeys(
		messageKey.CipherKey(),
		messageKey.MacKey(),
		messageKey.Iv(),
		messageKey.Index(),
	)
}

// SetMessageKeys will update the chain associated with the given sender key with
// the given message keys.
func (s *State) SetMessageKeys(senderEphemeral ecc.ECPublicKeyable, messageKeys *message.Keys) {
	chainAndIndex := s.receiverChain(senderEphemeral)
	chainState := chainAndIndex.ReceiverChain

	// Add the message keys to our chain state.
	chainState.AddMessageKeys(
		message.NewKeys(
			messageKeys.CipherKey(),
			messageKeys.MacKey(),
			messageKeys.Iv(),
			messageKeys.Index(),
		),
	)

	if len(chainState.MessageKeys()) > maxMessageKeys {
		chainState.PopFirstMessageKeys()
	}
}

// SetReceiverChainKey sets the session's receiver chain key with the given chain key
// associated with the given senderEphemeral key.
func (s *State) SetReceiverChainKey(senderEphemeral ecc.ECPublicKeyable, chainKey session.ChainKeyable) {
	chainAndIndex := s.receiverChain(senderEphemeral)
	chainState := chainAndIndex.ReceiverChain
	chainState.SetChainKey(chainKey.(*chain.Key))
}

// SetPendingKeyExchange will set the session's pending key exchange state to the given
// sequence and key pairs.
func (s *State) SetPendingKeyExchange(sequence uint32, ourBaseKey, ourRatchetKey *ecc.ECKeyPair,
	ourIdentityKey *identity.KeyPair) {

	s.pendingKeyExchange = NewPendingKeyExchange(
		sequence,
		ourBaseKey,
		ourRatchetKey,
		ourIdentityKey,
	)
}

// PendingKeyExchangeSequence will return the session's pending key exchange sequence
// number.
func (s *State) PendingKeyExchangeSequence() uint32 {
	return s.pendingKeyExchange.sequence
}

// PendingKeyExchangeBaseKeyPair will return the session's pending key exchange base keypair.
func (s *State) PendingKeyExchangeBaseKeyPair() *ecc.ECKeyPair {
	return s.pendingKeyExchange.localBaseKeyPair
}

// PendingKeyExchangeRatchetKeyPair will return the session's pending key exchange ratchet
// keypair.
func (s *State) PendingKeyExchangeRatchetKeyPair() *ecc.ECKeyPair {
	return s.pendingKeyExchange.localRatchetKeyPair
}

// PendingKeyExchangeIdentityKeyPair will return the session's pending key exchange identity
// keypair.
func (s *State) PendingKeyExchangeIdentityKeyPair() *identity.KeyPair {
	return s.pendingKeyExchange.localIdentityKeyPair
}

// HasPendingKeyExchange will return true if there is a valid pending key exchange waiting.
func (s *State) HasPendingKeyExchange() bool {
	return s.pendingKeyExchange != nil
}

// SetUnacknowledgedPreKeyMessage will return unacknowledged pre key message with the
// given key ids and base key.
func (s *State) SetUnacknowledgedPreKeyMessage(preKeyID *optional.Uint32, signedPreKeyID uint32, baseKey ecc.ECPublicKeyable) {
	s.pendingPreKey = NewPendingPreKey(
		preKeyID,
		signedPreKeyID,
		baseKey,
	)
}

// HasUnacknowledgedPreKeyMessage will return true if this session has an unacknowledged
// pre key message.
func (s *State) HasUnacknowledgedPreKeyMessage() bool {
	return s.pendingPreKey != nil
}

// UnackPreKeyMessageItems will return the session's unacknowledged pre key messages.
func (s *State) UnackPreKeyMessageItems() (*UnackPreKeyMessageItems, error) {
	preKeyID := s.pendingPreKey.preKeyID
	signedPreKeyID := s.pendingPreKey.signedPreKeyID
	baseKey, err := ecc.DecodePoint(s.pendingPreKey.baseKey.Serialize(), 0)
	if err != nil {
		return nil, err
	}
	return NewUnackPreKeyMessageItems(preKeyID, signedPreKeyID, baseKey), nil
}

// ClearUnackPreKeyMessage will clear the session's pending pre key.
func (s *State) ClearUnackPreKeyMessage() {
	s.pendingPreKey = nil
}

// SetRemoteRegistrationID sets the remote user's registration id.
func (s *State) SetRemoteRegistrationID(registrationID uint32) {
	s.remoteRegistrationID = registrationID
}

// RemoteRegistrationID returns the remote user's registration id.
func (s *State) RemoteRegistrationID() uint32 {
	return s.remoteRegistrationID
}

// SetLocalRegistrationID sets the local user's registration id.
func (s *State) SetLocalRegistrationID(registrationID uint32) {
	s.localRegistrationID = registrationID
}

// LocalRegistrationID returns the local user's registration id.
func (s *State) LocalRegistrationID() uint32 {
	return s.localRegistrationID
}

// Serialize will return the state as bytes using the given serializer.
func (s *State) Serialize() []byte {
	return s.serializer.Serialize(s.structure())
}

// structure will return a serializable structure of the
// the given state so it can be persistently stored.
func (s *State) structure() *StateStructure {
	// Convert our receiver chains into a serializeable structure
	receiverChains := make([]*ChainStructure, len(s.receiverChains))
	for i := range s.receiverChains {
		receiverChains[i] = s.receiverChains[i].structure()
	}

	// Convert our pending key exchange into a serializeable structure
	var pendingKeyExchange *PendingKeyExchangeStructure
	if s.pendingKeyExchange != nil {
		pendingKeyExchange = s.pendingKeyExchange.structure()
	}

	// Build and return our state structure.
	return &StateStructure{
		LocalIdentityPublic:  s.localIdentityPublic.Serialize(),
		LocalRegistrationID:  s.localRegistrationID,
		NeedsRefresh:         s.needsRefresh,
		PendingKeyExchange:   pendingKeyExchange,
		PendingPreKey:        s.pendingPreKey.structure(),
		PreviousCounter:      s.previousCounter,
		ReceiverChains:       receiverChains,
		RemoteIdentityPublic: s.remoteIdentityPublic.Serialize(),
		RemoteRegistrationID: s.remoteRegistrationID,
		RootKey:              s.rootKey.Bytes(),
		SenderBaseKey:        s.senderBaseKey.Serialize(),
		SenderChain:          s.senderChain.structure(),
		SessionVersion:       s.sessionVersion,
	}
}
