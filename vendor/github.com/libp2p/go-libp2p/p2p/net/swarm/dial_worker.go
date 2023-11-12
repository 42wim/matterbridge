package swarm

import (
	"context"
	"math"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"

	ma "github.com/multiformats/go-multiaddr"
)

// /////////////////////////////////////////////////////////////////////////////////
// lo and behold, The Dialer
// TODO explain how all this works
// ////////////////////////////////////////////////////////////////////////////////

// dialRequest is structure used to request dials to the peer associated with a
// worker loop
type dialRequest struct {
	// ctx is the context that may be used for the request
	// if another concurrent request is made, any of the concurrent request's ctx may be used for
	// dials to the peer's addresses
	// ctx for simultaneous connect requests have higher priority than normal requests
	ctx context.Context
	// resch is the channel used to send the response for this query
	resch chan dialResponse
}

// dialResponse is the response sent to dialRequests on the request's resch channel
type dialResponse struct {
	// conn is the connection to the peer on success
	conn *Conn
	// err is the error in dialing the peer
	// nil on connection success
	err error
}

// pendRequest is used to track progress on a dialRequest.
type pendRequest struct {
	// req is the original dialRequest
	req dialRequest
	// err comprises errors of all failed dials
	err *DialError
	// addrs are the addresses on which we are waiting for pending dials
	// At the time of creation addrs is initialised to all the addresses of the peer. On a failed dial,
	// the addr is removed from the map and err is updated. On a successful dial, the dialRequest is
	// completed and response is sent with the connection
	addrs map[string]struct{}
}

// addrDial tracks dials to a particular multiaddress.
type addrDial struct {
	// addr is the address dialed
	addr ma.Multiaddr
	// ctx is the context used for dialing the address
	ctx context.Context
	// conn is the established connection on success
	conn *Conn
	// err is the err on dialing the address
	err error
	// requests is the list of pendRequests interested in this dial
	// the value in the slice is the request number assigned to this request by the dialWorker
	requests []int
	// dialed indicates whether we have triggered the dial to the address
	dialed bool
	// createdAt is the time this struct was created
	createdAt time.Time
	// dialRankingDelay is the delay in dialing this address introduced by the ranking logic
	dialRankingDelay time.Duration
}

// dialWorker synchronises concurrent dials to a peer. It ensures that we make at most one dial to a
// peer's address
type dialWorker struct {
	s    *Swarm
	peer peer.ID
	// reqch is used to send dial requests to the worker. close reqch to end the worker loop
	reqch <-chan dialRequest
	// reqno is the request number used to track different dialRequests for a peer.
	// Each incoming request is assigned a reqno. This reqno is used in pendingRequests and in
	// addrDial objects in trackedDials to track this request
	reqno int
	// pendingRequests maps reqno to the pendRequest object for a dialRequest
	pendingRequests map[int]*pendRequest
	// trackedDials tracks dials to the peers addresses. An entry here is used to ensure that
	// we dial an address at most once
	trackedDials map[string]*addrDial
	// resch is used to receive response for dials to the peers addresses.
	resch chan dialResult

	connected bool // true when a connection has been successfully established

	// for testing
	wg sync.WaitGroup
	cl Clock
}

func newDialWorker(s *Swarm, p peer.ID, reqch <-chan dialRequest, cl Clock) *dialWorker {
	if cl == nil {
		cl = RealClock{}
	}
	return &dialWorker{
		s:               s,
		peer:            p,
		reqch:           reqch,
		pendingRequests: make(map[int]*pendRequest),
		trackedDials:    make(map[string]*addrDial),
		resch:           make(chan dialResult),
		cl:              cl,
	}
}

// loop implements the core dial worker loop. Requests are received on w.reqch.
// The loop exits when w.reqch is closed.
func (w *dialWorker) loop() {
	w.wg.Add(1)
	defer w.wg.Done()
	defer w.s.limiter.clearAllPeerDials(w.peer)

	// dq is used to pace dials to different addresses of the peer
	dq := newDialQueue()
	// dialsInFlight is the number of dials in flight.
	dialsInFlight := 0

	startTime := w.cl.Now()
	// dialTimer is the dialTimer used to trigger dials
	dialTimer := w.cl.InstantTimer(startTime.Add(math.MaxInt64))
	timerRunning := true
	// scheduleNextDial updates timer for triggering the next dial
	scheduleNextDial := func() {
		if timerRunning && !dialTimer.Stop() {
			<-dialTimer.Ch()
		}
		timerRunning = false
		if dq.len() > 0 {
			if dialsInFlight == 0 && !w.connected {
				// if there are no dials in flight, trigger the next dials immediately
				dialTimer.Reset(startTime)
			} else {
				dialTimer.Reset(startTime.Add(dq.top().Delay))
			}
			timerRunning = true
		}
	}

	// totalDials is used to track number of dials made by this worker for metrics
	totalDials := 0
loop:
	for {
		// The loop has three parts
		//  1. Input requests are received on w.reqch. If a suitable connection is not available we create
		//     a pendRequest object to track the dialRequest and add the addresses to dq.
		//  2. Addresses from the dialQueue are dialed at appropriate time intervals depending on delay logic.
		//     We are notified of the completion of these dials on w.resch.
		//  3. Responses for dials are received on w.resch. On receiving a response, we updated the pendRequests
		//     interested in dials on this address.

		select {
		case req, ok := <-w.reqch:
			if !ok {
				if w.s.metricsTracer != nil {
					w.s.metricsTracer.DialCompleted(w.connected, totalDials)
				}
				return
			}
			// We have received a new request. If we do not have a suitable connection,
			// track this dialRequest with a pendRequest.
			// Enqueue the peer's addresses relevant to this request in dq and
			// track dials to the addresses relevant to this request.

			c, err := w.s.bestAcceptableConnToPeer(req.ctx, w.peer)
			if c != nil || err != nil {
				req.resch <- dialResponse{conn: c, err: err}
				continue loop
			}

			addrs, err := w.s.addrsForDial(req.ctx, w.peer)
			if err != nil {
				req.resch <- dialResponse{err: err}
				continue loop
			}

			// get the delays to dial these addrs from the swarms dialRanker
			simConnect, _, _ := network.GetSimultaneousConnect(req.ctx)
			addrRanking := w.rankAddrs(addrs, simConnect)
			addrDelay := make(map[string]time.Duration, len(addrRanking))

			// create the pending request object
			pr := &pendRequest{
				req:   req,
				err:   &DialError{Peer: w.peer},
				addrs: make(map[string]struct{}, len(addrRanking)),
			}
			for _, adelay := range addrRanking {
				pr.addrs[string(adelay.Addr.Bytes())] = struct{}{}
				addrDelay[string(adelay.Addr.Bytes())] = adelay.Delay
			}

			// Check if dials to any of the addrs have completed already
			// If they have errored, record the error in pr. If they have succeeded,
			// respond with the connection.
			// If they are pending, add them to tojoin.
			// If we haven't seen any of the addresses before, add them to todial.
			var todial []ma.Multiaddr
			var tojoin []*addrDial

			for _, adelay := range addrRanking {
				ad, ok := w.trackedDials[string(adelay.Addr.Bytes())]
				if !ok {
					todial = append(todial, adelay.Addr)
					continue
				}

				if ad.conn != nil {
					// dial to this addr was successful, complete the request
					req.resch <- dialResponse{conn: ad.conn}
					continue loop
				}

				if ad.err != nil {
					// dial to this addr errored, accumulate the error
					pr.err.recordErr(ad.addr, ad.err)
					delete(pr.addrs, string(ad.addr.Bytes()))
					continue
				}

				// dial is still pending, add to the join list
				tojoin = append(tojoin, ad)
			}

			if len(todial) == 0 && len(tojoin) == 0 {
				// all request applicable addrs have been dialed, we must have errored
				req.resch <- dialResponse{err: pr.err}
				continue loop
			}

			// The request has some pending or new dials. We assign this request a request number.
			// This value of w.reqno is used to track this request in all the structures
			w.reqno++
			w.pendingRequests[w.reqno] = pr

			for _, ad := range tojoin {
				if !ad.dialed {
					// we haven't dialed this address. update the ad.ctx to have simultaneous connect values
					// set correctly
					if simConnect, isClient, reason := network.GetSimultaneousConnect(req.ctx); simConnect {
						if simConnect, _, _ := network.GetSimultaneousConnect(ad.ctx); !simConnect {
							ad.ctx = network.WithSimultaneousConnect(ad.ctx, isClient, reason)
							// update the element in dq to use the simultaneous connect delay.
							dq.Add(network.AddrDelay{
								Addr:  ad.addr,
								Delay: addrDelay[string(ad.addr.Bytes())],
							})
						}
					}
				}
				// add the request to the addrDial
				ad.requests = append(ad.requests, w.reqno)
			}

			if len(todial) > 0 {
				now := time.Now()
				// these are new addresses, track them and add them to dq
				for _, a := range todial {
					w.trackedDials[string(a.Bytes())] = &addrDial{
						addr:      a,
						ctx:       req.ctx,
						requests:  []int{w.reqno},
						createdAt: now,
					}
					dq.Add(network.AddrDelay{Addr: a, Delay: addrDelay[string(a.Bytes())]})
				}
			}
			// setup dialTimer for updates to dq
			scheduleNextDial()

		case <-dialTimer.Ch():
			// It's time to dial the next batch of addresses.
			// We don't check the delay of the addresses received from the queue here
			// because if the timer triggered before the delay, it means that all
			// the inflight dials have errored and we should dial the next batch of
			// addresses
			now := time.Now()
			for _, adelay := range dq.NextBatch() {
				// spawn the dial
				ad, ok := w.trackedDials[string(adelay.Addr.Bytes())]
				if !ok {
					log.Errorf("SWARM BUG: no entry for address %s in trackedDials", adelay.Addr)
					continue
				}
				ad.dialed = true
				ad.dialRankingDelay = now.Sub(ad.createdAt)
				err := w.s.dialNextAddr(ad.ctx, w.peer, ad.addr, w.resch)
				if err != nil {
					// Errored without attempting a dial. This happens in case of
					// backoff or black hole.
					w.dispatchError(ad, err)
				} else {
					// the dial was successful. update inflight dials
					dialsInFlight++
					totalDials++
				}
			}
			timerRunning = false
			// schedule more dials
			scheduleNextDial()

		case res := <-w.resch:
			// A dial to an address has completed.
			// Update all requests waiting on this address. On success, complete the request.
			// On error, record the error

			dialsInFlight--
			ad, ok := w.trackedDials[string(res.Addr.Bytes())]
			if !ok {
				log.Errorf("SWARM BUG: no entry for address %s in trackedDials", res.Addr)
				if res.Conn != nil {
					res.Conn.Close()
				}
				continue
			}

			if res.Conn != nil {
				// we got a connection, add it to the swarm
				conn, err := w.s.addConn(res.Conn, network.DirOutbound)
				if err != nil {
					// oops no, we failed to add it to the swarm
					res.Conn.Close()
					w.dispatchError(ad, err)
					continue loop
				}

				// request succeeded, respond to all pending requests
				for _, reqno := range ad.requests {
					pr, ok := w.pendingRequests[reqno]
					if !ok {
						// some other dial for this request succeeded before this one
						continue
					}
					pr.req.resch <- dialResponse{conn: conn}
					delete(w.pendingRequests, reqno)
				}

				ad.conn = conn
				ad.requests = nil

				if !w.connected {
					w.connected = true
					if w.s.metricsTracer != nil {
						w.s.metricsTracer.DialRankingDelay(ad.dialRankingDelay)
					}
				}

				continue loop
			}

			// it must be an error -- add backoff if applicable and dispatch
			// ErrDialRefusedBlackHole shouldn't end up here, just a safety check
			if res.Err != ErrDialRefusedBlackHole && res.Err != context.Canceled && !w.connected {
				// we only add backoff if there has not been a successful connection
				// for consistency with the old dialer behavior.
				w.s.backf.AddBackoff(w.peer, res.Addr)
			} else if res.Err == ErrDialRefusedBlackHole {
				log.Errorf("SWARM BUG: unexpected ErrDialRefusedBlackHole while dialing peer %s to addr %s",
					w.peer, res.Addr)
			}

			w.dispatchError(ad, res.Err)
			// Only schedule next dial on error.
			// If we scheduleNextDial on success, we will end up making one dial more than
			// required because the final successful dial will spawn one more dial
			scheduleNextDial()
		}
	}
}

// dispatches an error to a specific addr dial
func (w *dialWorker) dispatchError(ad *addrDial, err error) {
	ad.err = err
	for _, reqno := range ad.requests {
		pr, ok := w.pendingRequests[reqno]
		if !ok {
			// some other dial for this request succeeded before this one
			continue
		}

		// accumulate the error
		pr.err.recordErr(ad.addr, err)

		delete(pr.addrs, string(ad.addr.Bytes()))
		if len(pr.addrs) == 0 {
			// all addrs have erred, dispatch dial error
			// but first do a last one check in case an acceptable connection has landed from
			// a simultaneous dial that started later and added new acceptable addrs
			c, _ := w.s.bestAcceptableConnToPeer(pr.req.ctx, w.peer)
			if c != nil {
				pr.req.resch <- dialResponse{conn: c}
			} else {
				pr.req.resch <- dialResponse{err: pr.err}
			}
			delete(w.pendingRequests, reqno)
		}
	}

	ad.requests = nil

	// if it was a backoff, clear the address dial so that it doesn't inhibit new dial requests.
	// this is necessary to support active listen scenarios, where a new dial comes in while
	// another dial is in progress, and needs to do a direct connection without inhibitions from
	// dial backoff.
	if err == ErrDialBackoff {
		delete(w.trackedDials, string(ad.addr.Bytes()))
	}
}

// rankAddrs ranks addresses for dialing. if it's a simConnect request we
// dial all addresses immediately without any delay
func (w *dialWorker) rankAddrs(addrs []ma.Multiaddr, isSimConnect bool) []network.AddrDelay {
	if isSimConnect {
		return NoDelayDialRanker(addrs)
	}
	return w.s.dialRanker(addrs)
}

// dialQueue is a priority queue used to schedule dials
type dialQueue struct {
	// q contains dials ordered by delay
	q []network.AddrDelay
}

// newDialQueue returns a new dialQueue
func newDialQueue() *dialQueue {
	return &dialQueue{q: make([]network.AddrDelay, 0, 16)}
}

// Add adds adelay to the queue. If another element exists in the queue with
// the same address, it replaces that element.
func (dq *dialQueue) Add(adelay network.AddrDelay) {
	for i := 0; i < dq.len(); i++ {
		if dq.q[i].Addr.Equal(adelay.Addr) {
			if dq.q[i].Delay == adelay.Delay {
				// existing element is the same. nothing to do
				return
			}
			// remove the element
			copy(dq.q[i:], dq.q[i+1:])
			dq.q = dq.q[:len(dq.q)-1]
			break
		}
	}

	for i := 0; i < dq.len(); i++ {
		if dq.q[i].Delay > adelay.Delay {
			dq.q = append(dq.q, network.AddrDelay{}) // extend the slice
			copy(dq.q[i+1:], dq.q[i:])
			dq.q[i] = adelay
			return
		}
	}
	dq.q = append(dq.q, adelay)
}

// NextBatch returns all the elements in the queue with the highest priority
func (dq *dialQueue) NextBatch() []network.AddrDelay {
	if dq.len() == 0 {
		return nil
	}

	// i is the index of the second highest priority element
	var i int
	for i = 0; i < dq.len(); i++ {
		if dq.q[i].Delay != dq.q[0].Delay {
			break
		}
	}
	res := dq.q[:i]
	dq.q = dq.q[i:]
	return res
}

// top returns the top element of the queue
func (dq *dialQueue) top() network.AddrDelay {
	return dq.q[0]
}

// len returns the number of elements in the queue
func (dq *dialQueue) len() int {
	return len(dq.q)
}
