package ddp

import (
	"fmt"
	"sync"
	"time"

	"github.com/apex/log"
)

// Contains common utility types.

// -------------------------------------------------------------------

// KeyManager provides simple incrementing IDs for ddp messages.
type KeyManager struct {
	// nextID is the next ID for API calls
	nextID uint64
	// idMutex is a mutex to protect ID updates
	idMutex *sync.Mutex
}

// NewKeyManager creates a new instance and sets up resources.
func NewKeyManager() *KeyManager {
	return &KeyManager{idMutex: new(sync.Mutex)}
}

// Next issues a new ID for use in calls.
func (id *KeyManager) Next() string {
	id.idMutex.Lock()
	next := id.nextID
	id.nextID++
	id.idMutex.Unlock()
	return fmt.Sprintf("%x", next)
}

// -------------------------------------------------------------------

// PingTracker tracks in-flight pings.
type PingTracker struct {
	handler func(error)
	timeout time.Duration
	timer   *time.Timer
}

// -------------------------------------------------------------------

// Call represents an active RPC call.
type Call struct {
	ID            string      // The uuid for this method call
	ServiceMethod string      // The name of the service and method to call.
	Args          interface{} // The argument to the function (*struct).
	Reply         interface{} // The reply from the function (*struct).
	Error         error       // After completion, the error status.
	Done          chan *Call  // Strobes when call is complete.
	Owner         *Client     // Client that owns the method call
}

// done removes the call from any owners and strobes the done channel with itself.
func (call *Call) done() {
	delete(call.Owner.calls, call.ID)
	select {
	case call.Done <- call:
		// ok
	default:
		// We don't want to block here.  It is the caller's responsibility to make
		// sure the channel has enough buffer space. See comment in Go().
		log.Debug("rpc: discarding Call reply due to insufficient Done chan capacity")
	}
}

// IgnoreErr logs an error if it occurs and ignores it.
func IgnoreErr(err error, msg string) {
	if err != nil {
		log.WithError(err).Debug(msg)
	}
}