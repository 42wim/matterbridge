package rendezvous

import (
	"math/rand"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/discovery/backoff"
)

// RendezvousPoint is a structure that represent a node that can be used to discover new peers
type RendezvousPoint struct {
	sync.RWMutex

	id     peer.ID
	cookie []byte

	bkf     backoff.BackoffStrategy
	nextTry time.Time
}

// NewRendezvousPoint is used to create a RendezvousPoint
func NewRendezvousPoint(peerID peer.ID) *RendezvousPoint {
	rngSrc := rand.NewSource(rand.Int63())
	minBackoff, maxBackoff := time.Second*30, time.Hour
	bkf := backoff.NewExponentialBackoff(minBackoff, maxBackoff, backoff.FullJitter, time.Second, 5.0, 0, rand.New(rngSrc))

	now := time.Now()

	rp := &RendezvousPoint{
		id:      peerID,
		nextTry: now,
		bkf:     bkf(),
	}

	return rp
}

// Delay is used to indicate that the connection to a rendezvous point failed
func (rp *RendezvousPoint) Delay() {
	rp.Lock()
	defer rp.Unlock()

	rp.nextTry = time.Now().Add(rp.bkf.Delay())
}

// SetSuccess is used to indicate that a connection to a rendezvous point was succesful
func (rp *RendezvousPoint) SetSuccess(cookie []byte) {
	rp.Lock()
	defer rp.Unlock()

	rp.bkf.Reset()
	rp.nextTry = time.Now()
	rp.cookie = cookie
}

// NextTry returns when can a rendezvous point be used again
func (rp *RendezvousPoint) NextTry() time.Time {
	rp.RLock()
	defer rp.RUnlock()
	return rp.nextTry
}
