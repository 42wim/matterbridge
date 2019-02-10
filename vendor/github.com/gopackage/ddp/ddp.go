// Package ddp implements the MeteorJS DDP protocol over websockets. Fallback
// to longpolling is NOT supported (and is not planned on ever being supported
// by this library). We will try to model the library after `net/http` - right
// now the library is barebones and doesn't provide the pluggability of http.
// However, that's the goal for the package eventually.
package ddp

import (
	"fmt"
	"log"
	"sync"
	"time"
)

// debugLog is true if we should log debugging information about the connection
var debugLog = true

// The main file contains common utility types.

// -------------------------------------------------------------------

// idManager provides simple incrementing IDs for ddp messages.
type idManager struct {
	// nextID is the next ID for API calls
	nextID uint64
	// idMutex is a mutex to protect ID updates
	idMutex *sync.Mutex
}

// newidManager creates a new instance and sets up resources.
func newidManager() *idManager {
	return &idManager{idMutex: new(sync.Mutex)}
}

// newID issues a new ID for use in calls.
func (id *idManager) newID() string {
	id.idMutex.Lock()
	next := id.nextID
	id.nextID++
	id.idMutex.Unlock()
	return fmt.Sprintf("%x", next)
}

// -------------------------------------------------------------------

// pingTracker tracks in-flight pings.
type pingTracker struct {
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
		if debugLog {
			log.Println("rpc: discarding Call reply due to insufficient Done chan capacity")
		}
	}
}
