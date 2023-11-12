package holepunch

import (
	"context"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"

	ma "github.com/multiformats/go-multiaddr"
)

const (
	tracerGCInterval    = 2 * time.Minute
	tracerCacheDuration = 5 * time.Minute
)

// WithTracer enables holepunch tracing with EventTracer et
func WithTracer(et EventTracer) Option {
	return func(hps *Service) error {
		hps.tracer = &tracer{
			et:   et,
			mt:   nil,
			self: hps.host.ID(),
			peers: make(map[peer.ID]struct {
				counter int
				last    time.Time
			}),
		}
		return nil
	}
}

// WithMetricsTracer enables holepunch Tracing with MetricsTracer mt
func WithMetricsTracer(mt MetricsTracer) Option {
	return func(hps *Service) error {
		hps.tracer = &tracer{
			et:   nil,
			mt:   mt,
			self: hps.host.ID(),
			peers: make(map[peer.ID]struct {
				counter int
				last    time.Time
			}),
		}
		return nil
	}
}

// WithMetricsAndEventTracer enables holepunch tracking with MetricsTracer and EventTracer
func WithMetricsAndEventTracer(mt MetricsTracer, et EventTracer) Option {
	return func(hps *Service) error {
		hps.tracer = &tracer{
			et:   et,
			mt:   mt,
			self: hps.host.ID(),
			peers: make(map[peer.ID]struct {
				counter int
				last    time.Time
			}),
		}
		return nil
	}
}

type tracer struct {
	et   EventTracer
	mt   MetricsTracer
	self peer.ID

	refCount  sync.WaitGroup
	ctx       context.Context
	ctxCancel context.CancelFunc

	mutex sync.Mutex
	peers map[peer.ID]struct {
		counter int
		last    time.Time
	}
}

type EventTracer interface {
	Trace(evt *Event)
}

type Event struct {
	Timestamp int64       // UNIX nanos
	Peer      peer.ID     // local peer ID
	Remote    peer.ID     // remote peer ID
	Type      string      // event type
	Evt       interface{} // the actual event
}

// Event Types
const (
	DirectDialEvtT       = "DirectDial"
	ProtocolErrorEvtT    = "ProtocolError"
	StartHolePunchEvtT   = "StartHolePunch"
	EndHolePunchEvtT     = "EndHolePunch"
	HolePunchAttemptEvtT = "HolePunchAttempt"
)

// Event Objects
type DirectDialEvt struct {
	Success      bool
	EllapsedTime time.Duration
	Error        string `json:",omitempty"`
}

type ProtocolErrorEvt struct {
	Error string
}

type StartHolePunchEvt struct {
	RemoteAddrs []string
	RTT         time.Duration
}

type EndHolePunchEvt struct {
	Success      bool
	EllapsedTime time.Duration
	Error        string `json:",omitempty"`
}

type HolePunchAttemptEvt struct {
	Attempt int
}

// tracer interface
func (t *tracer) DirectDialSuccessful(p peer.ID, dt time.Duration) {
	if t == nil {
		return
	}

	if t.et != nil {
		t.et.Trace(&Event{
			Timestamp: time.Now().UnixNano(),
			Peer:      t.self,
			Remote:    p,
			Type:      DirectDialEvtT,
			Evt: &DirectDialEvt{
				Success:      true,
				EllapsedTime: dt,
			},
		})
	}

	if t.mt != nil {
		t.mt.DirectDialFinished(true)
	}
}

func (t *tracer) DirectDialFailed(p peer.ID, dt time.Duration, err error) {
	if t == nil {
		return
	}

	if t.et != nil {
		t.et.Trace(&Event{
			Timestamp: time.Now().UnixNano(),
			Peer:      t.self,
			Remote:    p,
			Type:      DirectDialEvtT,
			Evt: &DirectDialEvt{
				Success:      false,
				EllapsedTime: dt,
				Error:        err.Error(),
			},
		})
	}

	if t.mt != nil {
		t.mt.DirectDialFinished(false)
	}
}

func (t *tracer) ProtocolError(p peer.ID, err error) {
	if t != nil && t.et != nil {
		t.et.Trace(&Event{
			Timestamp: time.Now().UnixNano(),
			Peer:      t.self,
			Remote:    p,
			Type:      ProtocolErrorEvtT,
			Evt: &ProtocolErrorEvt{
				Error: err.Error(),
			},
		})
	}
}

func (t *tracer) StartHolePunch(p peer.ID, obsAddrs []ma.Multiaddr, rtt time.Duration) {
	if t != nil && t.et != nil {
		addrs := make([]string, 0, len(obsAddrs))
		for _, a := range obsAddrs {
			addrs = append(addrs, a.String())
		}

		t.et.Trace(&Event{
			Timestamp: time.Now().UnixNano(),
			Peer:      t.self,
			Remote:    p,
			Type:      StartHolePunchEvtT,
			Evt: &StartHolePunchEvt{
				RemoteAddrs: addrs,
				RTT:         rtt,
			},
		})
	}
}

func (t *tracer) EndHolePunch(p peer.ID, dt time.Duration, err error) {
	if t != nil && t.et != nil {
		evt := &EndHolePunchEvt{
			Success:      err == nil,
			EllapsedTime: dt,
		}
		if err != nil {
			evt.Error = err.Error()
		}

		t.et.Trace(&Event{
			Timestamp: time.Now().UnixNano(),
			Peer:      t.self,
			Remote:    p,
			Type:      EndHolePunchEvtT,
			Evt:       evt,
		})
	}
}

func (t *tracer) HolePunchFinished(side string, numAttempts int, theirAddrs []ma.Multiaddr, ourAddrs []ma.Multiaddr, directConn network.Conn) {
	if t != nil && t.mt != nil {
		t.mt.HolePunchFinished(side, numAttempts, theirAddrs, ourAddrs, directConn)
	}
}

func (t *tracer) HolePunchAttempt(p peer.ID) {
	if t != nil && t.et != nil {
		now := time.Now()
		t.mutex.Lock()
		attempt := t.peers[p]
		attempt.counter++
		counter := attempt.counter
		attempt.last = now
		t.peers[p] = attempt
		t.mutex.Unlock()

		t.et.Trace(&Event{
			Timestamp: now.UnixNano(),
			Peer:      t.self,
			Remote:    p,
			Type:      HolePunchAttemptEvtT,
			Evt:       &HolePunchAttemptEvt{Attempt: counter},
		})
	}
}

// gc cleans up the peers map. This is only run when tracer is initialised with a non nil
// EventTracer
func (t *tracer) gc() {
	defer t.refCount.Done()
	timer := time.NewTicker(tracerGCInterval)
	defer timer.Stop()

	for {
		select {
		case <-timer.C:
			now := time.Now()
			t.mutex.Lock()
			for id, entry := range t.peers {
				if entry.last.Before(now.Add(-tracerCacheDuration)) {
					delete(t.peers, id)
				}
			}
			t.mutex.Unlock()
		case <-t.ctx.Done():
			return
		}
	}
}

func (t *tracer) Start() {
	if t != nil && t.et != nil {
		t.ctx, t.ctxCancel = context.WithCancel(context.Background())
		t.refCount.Add(1)
		go t.gc()
	}
}

func (t *tracer) Close() error {
	if t != nil && t.et != nil {
		t.ctxCancel()
		t.refCount.Wait()
	}
	return nil
}
