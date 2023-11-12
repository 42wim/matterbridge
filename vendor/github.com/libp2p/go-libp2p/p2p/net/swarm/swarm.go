package swarm

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/libp2p/go-libp2p/core/connmgr"
	"github.com/libp2p/go-libp2p/core/event"
	"github.com/libp2p/go-libp2p/core/metrics"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/peerstore"
	"github.com/libp2p/go-libp2p/core/transport"

	logging "github.com/ipfs/go-log/v2"
	ma "github.com/multiformats/go-multiaddr"
	madns "github.com/multiformats/go-multiaddr-dns"
)

const (
	defaultDialTimeout = 15 * time.Second

	// defaultDialTimeoutLocal is the maximum duration a Dial to local network address
	// is allowed to take.
	// This includes the time between dialing the raw network connection,
	// protocol selection as well the handshake, if applicable.
	defaultDialTimeoutLocal = 5 * time.Second
)

var log = logging.Logger("swarm2")

// ErrSwarmClosed is returned when one attempts to operate on a closed swarm.
var ErrSwarmClosed = errors.New("swarm closed")

// ErrAddrFiltered is returned when trying to register a connection to a
// filtered address. You shouldn't see this error unless some underlying
// transport is misbehaving.
var ErrAddrFiltered = errors.New("address filtered")

// ErrDialTimeout is returned when one a dial times out due to the global timeout
var ErrDialTimeout = errors.New("dial timed out")

type Option func(*Swarm) error

// WithConnectionGater sets a connection gater
func WithConnectionGater(gater connmgr.ConnectionGater) Option {
	return func(s *Swarm) error {
		s.gater = gater
		return nil
	}
}

// WithMultiaddrResolver sets a custom multiaddress resolver
func WithMultiaddrResolver(maResolver *madns.Resolver) Option {
	return func(s *Swarm) error {
		s.maResolver = maResolver
		return nil
	}
}

// WithMetrics sets a metrics reporter
func WithMetrics(reporter metrics.Reporter) Option {
	return func(s *Swarm) error {
		s.bwc = reporter
		return nil
	}
}

func WithMetricsTracer(t MetricsTracer) Option {
	return func(s *Swarm) error {
		s.metricsTracer = t
		return nil
	}
}

func WithDialTimeout(t time.Duration) Option {
	return func(s *Swarm) error {
		s.dialTimeout = t
		return nil
	}
}

func WithDialTimeoutLocal(t time.Duration) Option {
	return func(s *Swarm) error {
		s.dialTimeoutLocal = t
		return nil
	}
}

func WithResourceManager(m network.ResourceManager) Option {
	return func(s *Swarm) error {
		s.rcmgr = m
		return nil
	}
}

// WithDialRanker configures swarm to use d as the DialRanker
func WithDialRanker(d network.DialRanker) Option {
	return func(s *Swarm) error {
		if d == nil {
			return errors.New("swarm: dial ranker cannot be nil")
		}
		s.dialRanker = d
		return nil
	}
}

// WithUDPBlackHoleConfig configures swarm to use c as the config for UDP black hole detection
// n is the size of the sliding window used to evaluate black hole state
// min is the minimum number of successes out of n required to not block requests
func WithUDPBlackHoleConfig(enabled bool, n, min int) Option {
	return func(s *Swarm) error {
		s.udpBlackHoleConfig = blackHoleConfig{Enabled: enabled, N: n, MinSuccesses: min}
		return nil
	}
}

// WithIPv6BlackHoleConfig configures swarm to use c as the config for IPv6 black hole detection
// n is the size of the sliding window used to evaluate black hole state
// min is the minimum number of successes out of n required to not block requests
func WithIPv6BlackHoleConfig(enabled bool, n, min int) Option {
	return func(s *Swarm) error {
		s.ipv6BlackHoleConfig = blackHoleConfig{Enabled: enabled, N: n, MinSuccesses: min}
		return nil
	}
}

// Swarm is a connection muxer, allowing connections to other peers to
// be opened and closed, while still using the same Chan for all
// communication. The Chan sends/receives Messages, which note the
// destination or source Peer.
type Swarm struct {
	nextConnID   uint64 // guarded by atomic
	nextStreamID uint64 // guarded by atomic

	// Close refcount. This allows us to fully wait for the swarm to be torn
	// down before continuing.
	refs sync.WaitGroup

	emitter event.Emitter

	rcmgr network.ResourceManager

	local peer.ID
	peers peerstore.Peerstore

	dialTimeout      time.Duration
	dialTimeoutLocal time.Duration

	conns struct {
		sync.RWMutex
		m map[peer.ID][]*Conn
	}

	listeners struct {
		sync.RWMutex

		ifaceListenAddres []ma.Multiaddr
		cacheEOL          time.Time

		m map[transport.Listener]struct{}
	}

	notifs struct {
		sync.RWMutex
		m map[network.Notifiee]struct{}
	}

	transports struct {
		sync.RWMutex
		m map[int]transport.Transport
	}

	maResolver *madns.Resolver

	// stream handlers
	streamh atomic.Pointer[network.StreamHandler]

	// dialing helpers
	dsync   *dialSync
	backf   DialBackoff
	limiter *dialLimiter
	gater   connmgr.ConnectionGater

	closeOnce sync.Once
	ctx       context.Context // is canceled when Close is called
	ctxCancel context.CancelFunc

	bwc           metrics.Reporter
	metricsTracer MetricsTracer

	dialRanker network.DialRanker

	udpBlackHoleConfig  blackHoleConfig
	ipv6BlackHoleConfig blackHoleConfig
	bhd                 *blackHoleDetector
}

// NewSwarm constructs a Swarm.
func NewSwarm(local peer.ID, peers peerstore.Peerstore, eventBus event.Bus, opts ...Option) (*Swarm, error) {
	emitter, err := eventBus.Emitter(new(event.EvtPeerConnectednessChanged))
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithCancel(context.Background())
	s := &Swarm{
		local:            local,
		peers:            peers,
		emitter:          emitter,
		ctx:              ctx,
		ctxCancel:        cancel,
		dialTimeout:      defaultDialTimeout,
		dialTimeoutLocal: defaultDialTimeoutLocal,
		maResolver:       madns.DefaultResolver,
		dialRanker:       DefaultDialRanker,

		// A black hole is a binary property. On a network if UDP dials are blocked or there is
		// no IPv6 connectivity, all dials will fail. So a low success rate of 5 out 100 dials
		// is good enough.
		udpBlackHoleConfig:  blackHoleConfig{Enabled: true, N: 100, MinSuccesses: 5},
		ipv6BlackHoleConfig: blackHoleConfig{Enabled: true, N: 100, MinSuccesses: 5},
	}

	s.conns.m = make(map[peer.ID][]*Conn)
	s.listeners.m = make(map[transport.Listener]struct{})
	s.transports.m = make(map[int]transport.Transport)
	s.notifs.m = make(map[network.Notifiee]struct{})

	for _, opt := range opts {
		if err := opt(s); err != nil {
			return nil, err
		}
	}
	if s.rcmgr == nil {
		s.rcmgr = &network.NullResourceManager{}
	}

	s.dsync = newDialSync(s.dialWorkerLoop)

	s.limiter = newDialLimiter(s.dialAddr)
	s.backf.init(s.ctx)

	s.bhd = newBlackHoleDetector(s.udpBlackHoleConfig, s.ipv6BlackHoleConfig, s.metricsTracer)

	return s, nil
}

func (s *Swarm) Close() error {
	s.closeOnce.Do(s.close)
	return nil
}

func (s *Swarm) close() {
	s.ctxCancel()

	s.emitter.Close()

	// Prevents new connections and/or listeners from being added to the swarm.
	s.listeners.Lock()
	listeners := s.listeners.m
	s.listeners.m = nil
	s.listeners.Unlock()

	s.conns.Lock()
	conns := s.conns.m
	s.conns.m = nil
	s.conns.Unlock()

	// Lots of goroutines but we might as well do this in parallel. We want to shut down as fast as
	// possible.

	for l := range listeners {
		go func(l transport.Listener) {
			if err := l.Close(); err != nil && err != transport.ErrListenerClosed {
				log.Errorf("error when shutting down listener: %s", err)
			}
		}(l)
	}

	for _, cs := range conns {
		for _, c := range cs {
			go func(c *Conn) {
				if err := c.Close(); err != nil {
					log.Errorf("error when shutting down connection: %s", err)
				}
			}(c)
		}
	}

	// Wait for everything to finish.
	s.refs.Wait()

	// Now close out any transports (if necessary). Do this after closing
	// all connections/listeners.
	s.transports.Lock()
	transports := s.transports.m
	s.transports.m = nil
	s.transports.Unlock()

	// Dedup transports that may be listening on multiple protocols
	transportsToClose := make(map[transport.Transport]struct{}, len(transports))
	for _, t := range transports {
		transportsToClose[t] = struct{}{}
	}

	var wg sync.WaitGroup
	for t := range transportsToClose {
		if closer, ok := t.(io.Closer); ok {
			wg.Add(1)
			go func(c io.Closer) {
				defer wg.Done()
				if err := closer.Close(); err != nil {
					log.Errorf("error when closing down transport %T: %s", c, err)
				}
			}(closer)
		}
	}
	wg.Wait()
}

func (s *Swarm) addConn(tc transport.CapableConn, dir network.Direction) (*Conn, error) {
	var (
		p    = tc.RemotePeer()
		addr = tc.RemoteMultiaddr()
	)

	// create the Stat object, initializing with the underlying connection Stat if available
	var stat network.ConnStats
	if cs, ok := tc.(network.ConnStat); ok {
		stat = cs.Stat()
	}
	stat.Direction = dir
	stat.Opened = time.Now()

	// Wrap and register the connection.
	c := &Conn{
		conn:  tc,
		swarm: s,
		stat:  stat,
		id:    atomic.AddUint64(&s.nextConnID, 1),
	}

	// we ONLY check upgraded connections here so we can send them a Disconnect message.
	// If we do this in the Upgrader, we will not be able to do this.
	if s.gater != nil {
		if allow, _ := s.gater.InterceptUpgraded(c); !allow {
			// TODO Send disconnect with reason here
			err := tc.Close()
			if err != nil {
				log.Warnf("failed to close connection with peer %s and addr %s; err: %s", p.Pretty(), addr, err)
			}
			return nil, ErrGaterDisallowedConnection
		}
	}

	// Add the public key.
	if pk := tc.RemotePublicKey(); pk != nil {
		s.peers.AddPubKey(p, pk)
	}

	// Clear any backoffs
	s.backf.Clear(p)

	// Finally, add the peer.
	s.conns.Lock()
	// Check if we're still online
	if s.conns.m == nil {
		s.conns.Unlock()
		tc.Close()
		return nil, ErrSwarmClosed
	}

	c.streams.m = make(map[*Stream]struct{})
	isFirstConnection := len(s.conns.m[p]) == 0
	s.conns.m[p] = append(s.conns.m[p], c)

	// Add two swarm refs:
	// * One will be decremented after the close notifications fire in Conn.doClose
	// * The other will be decremented when Conn.start exits.
	s.refs.Add(2)

	// Take the notification lock before releasing the conns lock to block
	// Disconnect notifications until after the Connect notifications done.
	c.notifyLk.Lock()
	s.conns.Unlock()

	// Emit event after releasing `s.conns` lock so that a consumer can still
	// use swarm methods that need the `s.conns` lock.
	if isFirstConnection {
		s.emitter.Emit(event.EvtPeerConnectednessChanged{
			Peer:          p,
			Connectedness: network.Connected,
		})
	}

	s.notifyAll(func(f network.Notifiee) {
		f.Connected(s, c)
	})
	c.notifyLk.Unlock()

	c.start()
	return c, nil
}

// Peerstore returns this swarms internal Peerstore.
func (s *Swarm) Peerstore() peerstore.Peerstore {
	return s.peers
}

// SetStreamHandler assigns the handler for new streams.
func (s *Swarm) SetStreamHandler(handler network.StreamHandler) {
	s.streamh.Store(&handler)
}

// StreamHandler gets the handler for new streams.
func (s *Swarm) StreamHandler() network.StreamHandler {
	handler := s.streamh.Load()
	if handler == nil {
		return nil
	}
	return *handler
}

// NewStream creates a new stream on any available connection to peer, dialing
// if necessary.
func (s *Swarm) NewStream(ctx context.Context, p peer.ID) (network.Stream, error) {
	log.Debugf("[%s] opening stream to peer [%s]", s.local, p)

	// Algorithm:
	// 1. Find the best connection, otherwise, dial.
	// 2. Try opening a stream.
	// 3. If the underlying connection is, in fact, closed, close the outer
	//    connection and try again. We do this in case we have a closed
	//    connection but don't notice it until we actually try to open a
	//    stream.
	//
	// Note: We only dial once.
	//
	// TODO: Try all connections even if we get an error opening a stream on
	// a non-closed connection.
	dials := 0
	for {
		// will prefer direct connections over relayed connections for opening streams
		c, err := s.bestAcceptableConnToPeer(ctx, p)
		if err != nil {
			return nil, err
		}

		if c == nil {
			if nodial, _ := network.GetNoDial(ctx); nodial {
				return nil, network.ErrNoConn
			}

			if dials >= DialAttempts {
				return nil, errors.New("max dial attempts exceeded")
			}
			dials++

			var err error
			c, err = s.dialPeer(ctx, p)
			if err != nil {
				return nil, err
			}
		}

		s, err := c.NewStream(ctx)
		if err != nil {
			if c.conn.IsClosed() {
				continue
			}
			return nil, err
		}
		return s, nil
	}
}

// ConnsToPeer returns all the live connections to peer.
func (s *Swarm) ConnsToPeer(p peer.ID) []network.Conn {
	// TODO: Consider sorting the connection list best to worst. Currently,
	// it's sorted oldest to newest.
	s.conns.RLock()
	defer s.conns.RUnlock()
	conns := s.conns.m[p]
	output := make([]network.Conn, len(conns))
	for i, c := range conns {
		output[i] = c
	}
	return output
}

func isBetterConn(a, b *Conn) bool {
	// If one is transient and not the other, prefer the non-transient connection.
	aTransient := a.Stat().Transient
	bTransient := b.Stat().Transient
	if aTransient != bTransient {
		return !aTransient
	}

	// If one is direct and not the other, prefer the direct connection.
	aDirect := isDirectConn(a)
	bDirect := isDirectConn(b)
	if aDirect != bDirect {
		return aDirect
	}

	// Otherwise, prefer the connection with more open streams.
	a.streams.Lock()
	aLen := len(a.streams.m)
	a.streams.Unlock()

	b.streams.Lock()
	bLen := len(b.streams.m)
	b.streams.Unlock()

	if aLen != bLen {
		return aLen > bLen
	}

	// finally, pick the last connection.
	return true
}

// bestConnToPeer returns the best connection to peer.
func (s *Swarm) bestConnToPeer(p peer.ID) *Conn {

	// TODO: Prefer some transports over others.
	// For now, prefers direct connections over Relayed connections.
	// For tie-breaking, select the newest non-closed connection with the most streams.
	s.conns.RLock()
	defer s.conns.RUnlock()

	var best *Conn
	for _, c := range s.conns.m[p] {
		if c.conn.IsClosed() {
			// We *will* garbage collect this soon anyways.
			continue
		}
		if best == nil || isBetterConn(c, best) {
			best = c
		}
	}
	return best
}

// - Returns the best "acceptable" connection, if available.
// - Returns nothing if no such connection exists, but if we should try dialing anyways.
// - Returns an error if no such connection exists, but we should not try dialing.
func (s *Swarm) bestAcceptableConnToPeer(ctx context.Context, p peer.ID) (*Conn, error) {
	conn := s.bestConnToPeer(p)
	if conn == nil {
		return nil, nil
	}

	forceDirect, _ := network.GetForceDirectDial(ctx)
	if forceDirect && !isDirectConn(conn) {
		return nil, nil
	}

	useTransient, _ := network.GetUseTransient(ctx)
	if useTransient || !conn.Stat().Transient {
		return conn, nil
	}

	return nil, network.ErrTransientConn
}

func isDirectConn(c *Conn) bool {
	return c != nil && !c.conn.Transport().Proxy()
}

// Connectedness returns our "connectedness" state with the given peer.
//
// To check if we have an open connection, use `s.Connectedness(p) ==
// network.Connected`.
func (s *Swarm) Connectedness(p peer.ID) network.Connectedness {
	if s.bestConnToPeer(p) != nil {
		return network.Connected
	}
	return network.NotConnected
}

// Conns returns a slice of all connections.
func (s *Swarm) Conns() []network.Conn {
	s.conns.RLock()
	defer s.conns.RUnlock()

	conns := make([]network.Conn, 0, len(s.conns.m))
	for _, cs := range s.conns.m {
		for _, c := range cs {
			conns = append(conns, c)
		}
	}
	return conns
}

// ClosePeer closes all connections to the given peer.
func (s *Swarm) ClosePeer(p peer.ID) error {
	conns := s.ConnsToPeer(p)
	switch len(conns) {
	case 0:
		return nil
	case 1:
		return conns[0].Close()
	default:
		errCh := make(chan error)
		for _, c := range conns {
			go func(c network.Conn) {
				errCh <- c.Close()
			}(c)
		}

		var errs []string
		for range conns {
			err := <-errCh
			if err != nil {
				errs = append(errs, err.Error())
			}
		}
		if len(errs) > 0 {
			return fmt.Errorf("when disconnecting from peer %s: %s", p, strings.Join(errs, ", "))
		}
		return nil
	}
}

// Peers returns a copy of the set of peers swarm is connected to.
func (s *Swarm) Peers() []peer.ID {
	s.conns.RLock()
	defer s.conns.RUnlock()
	peers := make([]peer.ID, 0, len(s.conns.m))
	for p := range s.conns.m {
		peers = append(peers, p)
	}

	return peers
}

// LocalPeer returns the local peer swarm is associated to.
func (s *Swarm) LocalPeer() peer.ID {
	return s.local
}

// Backoff returns the DialBackoff object for this swarm.
func (s *Swarm) Backoff() *DialBackoff {
	return &s.backf
}

// notifyAll sends a signal to all Notifiees
func (s *Swarm) notifyAll(notify func(network.Notifiee)) {
	s.notifs.RLock()
	for f := range s.notifs.m {
		notify(f)
	}
	s.notifs.RUnlock()
}

// Notify signs up Notifiee to receive signals when events happen
func (s *Swarm) Notify(f network.Notifiee) {
	s.notifs.Lock()
	s.notifs.m[f] = struct{}{}
	s.notifs.Unlock()
}

// StopNotify unregisters Notifiee fromr receiving signals
func (s *Swarm) StopNotify(f network.Notifiee) {
	s.notifs.Lock()
	delete(s.notifs.m, f)
	s.notifs.Unlock()
}

func (s *Swarm) removeConn(c *Conn) {
	p := c.RemotePeer()

	s.conns.Lock()

	cs := s.conns.m[p]

	if len(cs) == 1 {
		delete(s.conns.m, p)
		s.conns.Unlock()

		// Emit event after releasing `s.conns` lock so that a consumer can still
		// use swarm methods that need the `s.conns` lock.
		s.emitter.Emit(event.EvtPeerConnectednessChanged{
			Peer:          p,
			Connectedness: network.NotConnected,
		})
		return
	}

	defer s.conns.Unlock()

	for i, ci := range cs {
		if ci == c {
			// NOTE: We're intentionally preserving order.
			// This way, connections to a peer are always
			// sorted oldest to newest.
			copy(cs[i:], cs[i+1:])
			cs[len(cs)-1] = nil
			s.conns.m[p] = cs[:len(cs)-1]
			break
		}
	}
}

// String returns a string representation of Network.
func (s *Swarm) String() string {
	return fmt.Sprintf("<Swarm %s>", s.LocalPeer())
}

func (s *Swarm) ResourceManager() network.ResourceManager {
	return s.rcmgr
}

// Swarm is a Network.
var _ network.Network = (*Swarm)(nil)
var _ transport.TransportNetwork = (*Swarm)(nil)

type connWithMetrics struct {
	transport.CapableConn
	opened        time.Time
	dir           network.Direction
	metricsTracer MetricsTracer
}

func wrapWithMetrics(capableConn transport.CapableConn, metricsTracer MetricsTracer, opened time.Time, dir network.Direction) connWithMetrics {
	c := connWithMetrics{CapableConn: capableConn, opened: opened, dir: dir, metricsTracer: metricsTracer}
	c.metricsTracer.OpenedConnection(c.dir, capableConn.RemotePublicKey(), capableConn.ConnState(), capableConn.LocalMultiaddr())
	return c
}

func (c connWithMetrics) completedHandshake() {
	c.metricsTracer.CompletedHandshake(time.Since(c.opened), c.ConnState(), c.LocalMultiaddr())
}

func (c connWithMetrics) Close() error {
	c.metricsTracer.ClosedConnection(c.dir, time.Since(c.opened), c.ConnState(), c.LocalMultiaddr())
	return c.CapableConn.Close()
}

func (c connWithMetrics) Stat() network.ConnStats {
	if cs, ok := c.CapableConn.(network.ConnStat); ok {
		return cs.Stat()
	}
	return network.ConnStats{}
}

var _ network.ConnStat = connWithMetrics{}
