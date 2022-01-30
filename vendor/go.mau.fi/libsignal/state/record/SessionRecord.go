package record

import (
	"bytes"
)

// archivedStatesMaxLength describes how many previous session
// states we should keep track of.
const archivedStatesMaxLength int = 40

// SessionSerializer is an interface for serializing and deserializing
// a Signal Session into bytes. An implementation of this interface should be
// used to encode/decode the object into JSON, Protobuffers, etc.
type SessionSerializer interface {
	Serialize(state *SessionStructure) []byte
	Deserialize(serialized []byte) (*SessionStructure, error)
}

// NewSessionFromBytes will return a Signal Session from the given
// bytes using the given serializer.
func NewSessionFromBytes(serialized []byte, serializer SessionSerializer, stateSerializer StateSerializer) (*Session, error) {
	// Use the given serializer to decode the session.
	sessionStructure, err := serializer.Deserialize(serialized)
	if err != nil {
		return nil, err
	}

	return NewSessionFromStructure(sessionStructure, serializer, stateSerializer)
}

// NewSession creates a new session record and uses the given session and state
// serializers to convert the object into storeable bytes.
func NewSession(serializer SessionSerializer, stateSerializer StateSerializer) *Session {
	record := Session{
		sessionState:   NewState(stateSerializer),
		previousStates: []*State{},
		fresh:          true,
		serializer:     serializer,
	}

	return &record
}

// NewSessionFromStructure will return a new Signal Session from the given
// session structure and serializer.
func NewSessionFromStructure(structure *SessionStructure, serializer SessionSerializer,
	stateSerializer StateSerializer) (*Session, error) {

	// Build our previous states from structure.
	previousStates := make([]*State, len(structure.PreviousStates))
	for i := range structure.PreviousStates {
		var err error
		previousStates[i], err = NewStateFromStructure(structure.PreviousStates[i], stateSerializer)
		if err != nil {
			return nil, err
		}
	}

	// Build our current state from structure.
	sessionState, err := NewStateFromStructure(structure.SessionState, stateSerializer)
	if err != nil {
		return nil, err
	}

	// Build and return our session.
	session := &Session{
		previousStates: previousStates,
		sessionState:   sessionState,
		serializer:     serializer,
		fresh:          false,
	}

	return session, nil
}

// NewSessionFromState creates a new session record from the given
// session state.
func NewSessionFromState(sessionState *State, serializer SessionSerializer) *Session {
	record := Session{
		sessionState:   sessionState,
		previousStates: []*State{},
		fresh:          false,
		serializer:     serializer,
	}

	return &record
}

// SessionStructure is a public, serializeable structure for Signal
// Sessions. The states defined in the session are immuteable, as
// they should not be changed by anyone but the serializer.
type SessionStructure struct {
	SessionState   *StateStructure
	PreviousStates []*StateStructure
}

// Session encapsulates the state of an ongoing session.
type Session struct {
	serializer     SessionSerializer
	sessionState   *State
	previousStates []*State
	fresh          bool
}

// SetState sets the session record's current state to the given
// one.
func (r *Session) SetState(sessionState *State) {
	r.sessionState = sessionState
}

// IsFresh is used to determine if this is a brand new session
// or if a session record has already existed.
func (r *Session) IsFresh() bool {
	return r.fresh
}

// SessionState returns the session state object of the current
// session record.
func (r *Session) SessionState() *State {
	return r.sessionState
}

// PreviousSessionStates returns a list of all currently maintained
// "previous" session states.
func (r *Session) PreviousSessionStates() []*State {
	return r.previousStates
}

// HasSessionState will check this record to see if the sender's
// base key exists in the current and previous states.
func (r *Session) HasSessionState(version int, senderBaseKey []byte) bool {
	// Ensure the session state version is identical to this one.
	if r.sessionState.Version() == version && (bytes.Compare(senderBaseKey, r.sessionState.SenderBaseKey()) == 0) {
		return true
	}

	// Loop through all of our previous states and see if this
	// exists in our state.
	for i := range r.previousStates {
		if r.previousStates[i].Version() == version && bytes.Compare(senderBaseKey, r.previousStates[i].SenderBaseKey()) == 0 {
			return true
		}
	}

	return false
}

// ArchiveCurrentState moves the current session state into the list
// of "previous" session states, and replaces the current session state
// with a fresh reset instance.
func (r *Session) ArchiveCurrentState() {
	r.PromoteState(NewState(r.sessionState.serializer))
}

// PromoteState takes the given session state and replaces it with the
// current state, pushing the previous current state to "previousStates".
func (r *Session) PromoteState(promotedState *State) {
	r.previousStates = r.prependStates(r.previousStates, r.sessionState)
	r.sessionState = promotedState

	// Remove the last state if it has reached our maximum length
	if len(r.previousStates) > archivedStatesMaxLength {
		r.previousStates = r.removeLastState(r.previousStates)
	}
}

// Serialize will return the session as serialized bytes so it can be
// persistently stored.
func (r *Session) Serialize() []byte {
	return r.serializer.Serialize(r.Structure())
}

// prependStates takes an array/slice of states and prepends it with
// the given session state.
func (r *Session) prependStates(states []*State, sessionState *State) []*State {
	return append([]*State{sessionState}, states...)
}

// removeLastState takes an array/slice of states and removes the
// last element from it.
func (r *Session) removeLastState(states []*State) []*State {
	return states[:len(states)-1]
}

// Structure will return a simple serializable session structure
// from the given structure. This is used for serialization to persistently
// store a session record.
func (r *Session) Structure() *SessionStructure {
	previousStates := make([]*StateStructure, len(r.previousStates))
	for i := range r.previousStates {
		previousStates[i] = r.previousStates[i].structure()
	}
	return &SessionStructure{
		SessionState:   r.sessionState.structure(),
		PreviousStates: previousStates,
	}
}
