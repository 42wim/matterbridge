package store

import (
	"go.mau.fi/libsignal/protocol"
	"go.mau.fi/libsignal/state/record"
)

// Session store is an interface for the persistent storage of session
// state information for remote clients.
type Session interface {
	LoadSession(address *protocol.SignalAddress) *record.Session
	GetSubDeviceSessions(name string) []uint32
	StoreSession(remoteAddress *protocol.SignalAddress, record *record.Session)
	ContainsSession(remoteAddress *protocol.SignalAddress) bool
	DeleteSession(remoteAddress *protocol.SignalAddress)
	DeleteAllSessions()
}
