package steam

import (
	"github.com/Philipp15b/go-steam/netutil"
)

// When this event is emitted by the Client, the connection is automatically closed.
// This may be caused by a network error, for example.
type FatalErrorEvent error

type ConnectedEvent struct{}

type DisconnectedEvent struct{}

// A list of connection manager addresses to connect to in the future.
// You should always save them and then select one of these
// instead of the builtin ones for the next connection.
type ClientCMListEvent struct {
	Addresses []*netutil.PortAddr
}
