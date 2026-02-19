package store

import (
	"context"

	"go.mau.fi/libsignal/protocol"
	"go.mau.fi/libsignal/state/record"
)

// Session store is an interface for the persistent storage of session
// state information for remote clients.
type Session interface {
	LoadSession(ctx context.Context, address *protocol.SignalAddress) (*record.Session, error)
	GetSubDeviceSessions(ctx context.Context, name string) ([]uint32, error)
	StoreSession(ctx context.Context, remoteAddress *protocol.SignalAddress, record *record.Session) error
	ContainsSession(ctx context.Context, remoteAddress *protocol.SignalAddress) (bool, error)
	DeleteSession(ctx context.Context, remoteAddress *protocol.SignalAddress) error
	DeleteAllSessions(ctx context.Context) error
}
