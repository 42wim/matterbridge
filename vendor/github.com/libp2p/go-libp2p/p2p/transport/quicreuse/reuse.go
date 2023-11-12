package quicreuse

import (
	"context"
	"crypto/tls"
	"net"
	"sync"
	"time"

	"github.com/google/gopacket/routing"
	"github.com/libp2p/go-netroute"
	"github.com/quic-go/quic-go"
)

type refCountedQuicTransport interface {
	LocalAddr() net.Addr

	// Used to send packets directly around QUIC. Useful for hole punching.
	WriteTo([]byte, net.Addr) (int, error)

	Close() error

	// count transport reference
	DecreaseCount()
	IncreaseCount()

	Dial(ctx context.Context, addr net.Addr, tlsConf *tls.Config, conf *quic.Config) (quic.Connection, error)
	Listen(tlsConf *tls.Config, conf *quic.Config) (*quic.Listener, error)
}

type singleOwnerTransport struct {
	quic.Transport

	// Used to write packets directly around QUIC.
	packetConn net.PacketConn
}

func (c *singleOwnerTransport) IncreaseCount() {}
func (c *singleOwnerTransport) DecreaseCount() {
	c.Transport.Close()
}

func (c *singleOwnerTransport) LocalAddr() net.Addr {
	return c.Transport.Conn.LocalAddr()
}

func (c *singleOwnerTransport) Close() error {
	// TODO(when we drop support for go 1.19) use errors.Join
	c.Transport.Close()
	return c.packetConn.Close()
}

func (c *singleOwnerTransport) WriteTo(b []byte, addr net.Addr) (int, error) {
	// Safe because we called quic.OptimizeConn ourselves.
	return c.packetConn.WriteTo(b, addr)
}

// Constant. Defined as variables to simplify testing.
var (
	garbageCollectInterval = 30 * time.Second
	maxUnusedDuration      = 10 * time.Second
)

type refcountedTransport struct {
	quic.Transport

	// Used to write packets directly around QUIC.
	packetConn net.PacketConn

	mutex       sync.Mutex
	refCount    int
	unusedSince time.Time
}

func (c *refcountedTransport) IncreaseCount() {
	c.mutex.Lock()
	c.refCount++
	c.unusedSince = time.Time{}
	c.mutex.Unlock()
}

func (c *refcountedTransport) Close() error {
	// TODO(when we drop support for go 1.19) use errors.Join
	c.Transport.Close()
	return c.packetConn.Close()
}

func (c *refcountedTransport) WriteTo(b []byte, addr net.Addr) (int, error) {
	// Safe because we called quic.OptimizeConn ourselves.
	return c.packetConn.WriteTo(b, addr)
}

func (c *refcountedTransport) LocalAddr() net.Addr {
	return c.Transport.Conn.LocalAddr()
}

func (c *refcountedTransport) DecreaseCount() {
	c.mutex.Lock()
	c.refCount--
	if c.refCount == 0 {
		c.unusedSince = time.Now()
	}
	c.mutex.Unlock()
}

func (c *refcountedTransport) ShouldGarbageCollect(now time.Time) bool {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return !c.unusedSince.IsZero() && c.unusedSince.Add(maxUnusedDuration).Before(now)
}

type reuse struct {
	mutex sync.Mutex

	closeChan  chan struct{}
	gcStopChan chan struct{}

	routes  routing.Router
	unicast map[string] /* IP.String() */ map[int] /* port */ *refcountedTransport
	// globalListeners contains transports that are listening on 0.0.0.0 / ::
	globalListeners map[int]*refcountedTransport
	// globalDialers contains transports that we've dialed out from. These transports are listening on 0.0.0.0 / ::
	// On Dial, transports are reused from this map if no transport is available in the globalListeners
	// On Listen, transports are reused from this map if the requested port is 0, and then moved to globalListeners
	globalDialers map[int]*refcountedTransport

	statelessResetKey *quic.StatelessResetKey
	metricsTracer     *metricsTracer
}

func newReuse(srk *quic.StatelessResetKey, mt *metricsTracer) *reuse {
	r := &reuse{
		unicast:           make(map[string]map[int]*refcountedTransport),
		globalListeners:   make(map[int]*refcountedTransport),
		globalDialers:     make(map[int]*refcountedTransport),
		closeChan:         make(chan struct{}),
		gcStopChan:        make(chan struct{}),
		statelessResetKey: srk,
		metricsTracer:     mt,
	}
	go r.gc()
	return r
}

func (r *reuse) gc() {
	defer func() {
		r.mutex.Lock()
		for _, tr := range r.globalListeners {
			tr.Close()
		}
		for _, tr := range r.globalDialers {
			tr.Close()
		}
		for _, trs := range r.unicast {
			for _, tr := range trs {
				tr.Close()
			}
		}
		r.mutex.Unlock()
		close(r.gcStopChan)
	}()
	ticker := time.NewTicker(garbageCollectInterval)
	defer ticker.Stop()

	for {
		select {
		case <-r.closeChan:
			return
		case <-ticker.C:
			now := time.Now()
			r.mutex.Lock()
			for key, tr := range r.globalListeners {
				if tr.ShouldGarbageCollect(now) {
					tr.Close()
					delete(r.globalListeners, key)
				}
			}
			for key, tr := range r.globalDialers {
				if tr.ShouldGarbageCollect(now) {
					tr.Close()
					delete(r.globalDialers, key)
				}
			}
			for ukey, trs := range r.unicast {
				for key, tr := range trs {
					if tr.ShouldGarbageCollect(now) {
						tr.Close()
						delete(trs, key)
					}
				}
				if len(trs) == 0 {
					delete(r.unicast, ukey)
					// If we've dropped all transports with a unicast binding,
					// assume our routes may have changed.
					if len(r.unicast) == 0 {
						r.routes = nil
					} else {
						// Ignore the error, there's nothing we can do about
						// it.
						r.routes, _ = netroute.New()
					}
				}
			}
			r.mutex.Unlock()
		}
	}
}

func (r *reuse) TransportForDial(network string, raddr *net.UDPAddr) (*refcountedTransport, error) {
	var ip *net.IP

	// Only bother looking up the source address if we actually _have_ non 0.0.0.0 listeners.
	// Otherwise, save some time.

	r.mutex.Lock()
	router := r.routes
	r.mutex.Unlock()

	if router != nil {
		_, _, src, err := router.Route(raddr.IP)
		if err == nil && !src.IsUnspecified() {
			ip = &src
		}
	}

	r.mutex.Lock()
	defer r.mutex.Unlock()

	tr, err := r.transportForDialLocked(network, ip)
	if err != nil {
		return nil, err
	}
	tr.IncreaseCount()
	return tr, nil
}

func (r *reuse) transportForDialLocked(network string, source *net.IP) (*refcountedTransport, error) {
	if source != nil {
		// We already have at least one suitable transport...
		if trs, ok := r.unicast[source.String()]; ok {
			// ... we don't care which port we're dialing from. Just use the first.
			for _, tr := range trs {
				return tr, nil
			}
		}
	}

	// Use a transport listening on 0.0.0.0 (or ::).
	// Again, we don't care about the port number.
	for _, tr := range r.globalListeners {
		return tr, nil
	}

	// Use a transport we've previously dialed from
	for _, tr := range r.globalDialers {
		return tr, nil
	}

	// We don't have a transport that we can use for dialing.
	// Dial a new connection from a random port.
	var addr *net.UDPAddr
	switch network {
	case "udp4":
		addr = &net.UDPAddr{IP: net.IPv4zero, Port: 0}
	case "udp6":
		addr = &net.UDPAddr{IP: net.IPv6zero, Port: 0}
	}
	conn, err := listenAndOptimize(network, addr)
	if err != nil {
		return nil, err
	}
	tr := &refcountedTransport{Transport: quic.Transport{
		Conn:              conn,
		StatelessResetKey: r.statelessResetKey,
	}, packetConn: conn}
	if r.metricsTracer != nil {
		tr.Transport.Tracer = r.metricsTracer
	}
	r.globalDialers[conn.LocalAddr().(*net.UDPAddr).Port] = tr
	return tr, nil
}

func (r *reuse) TransportForListen(network string, laddr *net.UDPAddr) (*refcountedTransport, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	// Check if we can reuse a transport we have already dialed out from.
	// We reuse a transport from globalDialers when the requested port is 0 or the requested
	// port is already in the globalDialers.
	// If we are reusing a transport from globalDialers, we move the globalDialers entry to
	// globalListeners
	if laddr.IP.IsUnspecified() {
		var rTr *refcountedTransport
		var localAddr *net.UDPAddr

		if laddr.Port == 0 {
			// the requested port is 0, we can reuse any transport
			for _, tr := range r.globalDialers {
				rTr = tr
				localAddr = rTr.LocalAddr().(*net.UDPAddr)
				delete(r.globalDialers, localAddr.Port)
				break
			}
		} else if _, ok := r.globalDialers[laddr.Port]; ok {
			rTr = r.globalDialers[laddr.Port]
			localAddr = rTr.LocalAddr().(*net.UDPAddr)
			delete(r.globalDialers, localAddr.Port)
		}
		// found a match
		if rTr != nil {
			rTr.IncreaseCount()
			r.globalListeners[localAddr.Port] = rTr
			return rTr, nil
		}
	}

	conn, err := listenAndOptimize(network, laddr)
	if err != nil {
		return nil, err
	}
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	tr := &refcountedTransport{Transport: quic.Transport{
		Conn:              conn,
		StatelessResetKey: r.statelessResetKey,
	}, packetConn: conn}
	if r.metricsTracer != nil {
		tr.Transport.Tracer = r.metricsTracer
	}

	tr.IncreaseCount()

	// Deal with listen on a global address
	if localAddr.IP.IsUnspecified() {
		// The kernel already checked that the laddr is not already listen
		// so we need not check here (when we create ListenUDP).
		r.globalListeners[localAddr.Port] = tr
		return tr, nil
	}

	// Deal with listen on a unicast address
	if _, ok := r.unicast[localAddr.IP.String()]; !ok {
		r.unicast[localAddr.IP.String()] = make(map[int]*refcountedTransport)
		// Assume the system's routes may have changed if we're adding a new listener.
		// Ignore the error, there's nothing we can do.
		r.routes, _ = netroute.New()
	}

	// The kernel already checked that the laddr is not already listen
	// so we need not check here (when we create ListenUDP).
	r.unicast[localAddr.IP.String()][localAddr.Port] = tr
	return tr, nil
}

func (r *reuse) Close() error {
	close(r.closeChan)
	<-r.gcStopChan
	return nil
}
