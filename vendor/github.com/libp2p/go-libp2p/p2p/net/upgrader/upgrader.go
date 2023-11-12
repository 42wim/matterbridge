package upgrader

import (
	"context"
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/libp2p/go-libp2p/core/connmgr"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	ipnet "github.com/libp2p/go-libp2p/core/pnet"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/libp2p/go-libp2p/core/sec"
	"github.com/libp2p/go-libp2p/core/transport"
	"github.com/libp2p/go-libp2p/p2p/net/pnet"

	manet "github.com/multiformats/go-multiaddr/net"
	mss "github.com/multiformats/go-multistream"
)

// ErrNilPeer is returned when attempting to upgrade an outbound connection
// without specifying a peer ID.
var ErrNilPeer = errors.New("nil peer")

// AcceptQueueLength is the number of connections to fully setup before not accepting any new connections
var AcceptQueueLength = 16

const (
	defaultAcceptTimeout    = 15 * time.Second
	defaultNegotiateTimeout = 60 * time.Second
)

type Option func(*upgrader) error

func WithAcceptTimeout(t time.Duration) Option {
	return func(u *upgrader) error {
		u.acceptTimeout = t
		return nil
	}
}

type StreamMuxer struct {
	ID    protocol.ID
	Muxer network.Multiplexer
}

// Upgrader is a multistream upgrader that can upgrade an underlying connection
// to a full transport connection (secure and multiplexed).
type upgrader struct {
	psk       ipnet.PSK
	connGater connmgr.ConnectionGater
	rcmgr     network.ResourceManager

	muxerMuxer *mss.MultistreamMuxer[protocol.ID]
	muxers     []StreamMuxer
	muxerIDs   []protocol.ID

	security      []sec.SecureTransport
	securityMuxer *mss.MultistreamMuxer[protocol.ID]
	securityIDs   []protocol.ID

	// AcceptTimeout is the maximum duration an Accept is allowed to take.
	// This includes the time between accepting the raw network connection,
	// protocol selection as well as the handshake, if applicable.
	//
	// If unset, the default value (15s) is used.
	acceptTimeout time.Duration
}

var _ transport.Upgrader = &upgrader{}

func New(security []sec.SecureTransport, muxers []StreamMuxer, psk ipnet.PSK, rcmgr network.ResourceManager, connGater connmgr.ConnectionGater, opts ...Option) (transport.Upgrader, error) {
	u := &upgrader{
		acceptTimeout: defaultAcceptTimeout,
		rcmgr:         rcmgr,
		connGater:     connGater,
		psk:           psk,
		muxerMuxer:    mss.NewMultistreamMuxer[protocol.ID](),
		muxers:        muxers,
		security:      security,
		securityMuxer: mss.NewMultistreamMuxer[protocol.ID](),
	}
	for _, opt := range opts {
		if err := opt(u); err != nil {
			return nil, err
		}
	}
	if u.rcmgr == nil {
		u.rcmgr = &network.NullResourceManager{}
	}
	u.muxerIDs = make([]protocol.ID, 0, len(muxers))
	for _, m := range muxers {
		u.muxerMuxer.AddHandler(m.ID, nil)
		u.muxerIDs = append(u.muxerIDs, m.ID)
	}
	u.securityIDs = make([]protocol.ID, 0, len(security))
	for _, s := range security {
		u.securityMuxer.AddHandler(s.ID(), nil)
		u.securityIDs = append(u.securityIDs, s.ID())
	}
	return u, nil
}

// UpgradeListener upgrades the passed multiaddr-net listener into a full libp2p-transport listener.
func (u *upgrader) UpgradeListener(t transport.Transport, list manet.Listener) transport.Listener {
	ctx, cancel := context.WithCancel(context.Background())
	l := &listener{
		Listener:  list,
		upgrader:  u,
		transport: t,
		rcmgr:     u.rcmgr,
		threshold: newThreshold(AcceptQueueLength),
		incoming:  make(chan transport.CapableConn),
		cancel:    cancel,
		ctx:       ctx,
	}
	go l.handleIncoming()
	return l
}

// Upgrade upgrades the multiaddr/net connection into a full libp2p-transport connection.
func (u *upgrader) Upgrade(ctx context.Context, t transport.Transport, maconn manet.Conn, dir network.Direction, p peer.ID, connScope network.ConnManagementScope) (transport.CapableConn, error) {
	c, err := u.upgrade(ctx, t, maconn, dir, p, connScope)
	if err != nil {
		connScope.Done()
		return nil, err
	}
	return c, nil
}

func (u *upgrader) upgrade(ctx context.Context, t transport.Transport, maconn manet.Conn, dir network.Direction, p peer.ID, connScope network.ConnManagementScope) (transport.CapableConn, error) {
	if dir == network.DirOutbound && p == "" {
		return nil, ErrNilPeer
	}
	var stat network.ConnStats
	if cs, ok := maconn.(network.ConnStat); ok {
		stat = cs.Stat()
	}

	var conn net.Conn = maconn
	if u.psk != nil {
		pconn, err := pnet.NewProtectedConn(u.psk, conn)
		if err != nil {
			conn.Close()
			return nil, fmt.Errorf("failed to setup private network protector: %w", err)
		}
		conn = pconn
	} else if ipnet.ForcePrivateNetwork {
		log.Error("tried to dial with no Private Network Protector but usage of Private Networks is forced by the environment")
		return nil, ipnet.ErrNotInPrivateNetwork
	}

	sconn, security, server, err := u.setupSecurity(ctx, conn, p, dir)
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to negotiate security protocol: %w", err)
	}

	// call the connection gater, if one is registered.
	if u.connGater != nil && !u.connGater.InterceptSecured(dir, sconn.RemotePeer(), maconn) {
		if err := maconn.Close(); err != nil {
			log.Errorw("failed to close connection", "peer", p, "addr", maconn.RemoteMultiaddr(), "error", err)
		}
		return nil, fmt.Errorf("gater rejected connection with peer %s and addr %s with direction %d",
			sconn.RemotePeer(), maconn.RemoteMultiaddr(), dir)
	}
	// Only call SetPeer if it hasn't already been set -- this can happen when we don't know
	// the peer in advance and in some bug scenarios.
	if connScope.PeerScope() == nil {
		if err := connScope.SetPeer(sconn.RemotePeer()); err != nil {
			log.Debugw("resource manager blocked connection for peer", "peer", sconn.RemotePeer(), "addr", conn.RemoteAddr(), "error", err)
			if err := maconn.Close(); err != nil {
				log.Errorw("failed to close connection", "peer", p, "addr", maconn.RemoteMultiaddr(), "error", err)
			}
			return nil, fmt.Errorf("resource manager connection with peer %s and addr %s with direction %d",
				sconn.RemotePeer(), maconn.RemoteMultiaddr(), dir)
		}
	}

	muxer, smconn, err := u.setupMuxer(ctx, sconn, server, connScope.PeerScope())
	if err != nil {
		sconn.Close()
		return nil, fmt.Errorf("failed to negotiate stream multiplexer: %w", err)
	}

	tc := &transportConn{
		MuxedConn:                 smconn,
		ConnMultiaddrs:            maconn,
		ConnSecurity:              sconn,
		transport:                 t,
		stat:                      stat,
		scope:                     connScope,
		muxer:                     muxer,
		security:                  security,
		usedEarlyMuxerNegotiation: sconn.ConnState().UsedEarlyMuxerNegotiation,
	}
	return tc, nil
}

func (u *upgrader) setupSecurity(ctx context.Context, conn net.Conn, p peer.ID, dir network.Direction) (sec.SecureConn, protocol.ID, bool, error) {
	isServer := dir == network.DirInbound
	var st sec.SecureTransport
	var err error
	st, isServer, err = u.negotiateSecurity(ctx, conn, isServer)
	if err != nil {
		return nil, "", false, err
	}
	if isServer {
		sconn, err := st.SecureInbound(ctx, conn, p)
		return sconn, st.ID(), true, err
	}
	sconn, err := st.SecureOutbound(ctx, conn, p)
	return sconn, st.ID(), false, err
}

func (u *upgrader) negotiateMuxer(nc net.Conn, isServer bool) (*StreamMuxer, error) {
	if err := nc.SetDeadline(time.Now().Add(defaultNegotiateTimeout)); err != nil {
		return nil, err
	}

	var proto protocol.ID
	if isServer {
		selected, _, err := u.muxerMuxer.Negotiate(nc)
		if err != nil {
			return nil, err
		}
		proto = selected
	} else {
		selected, err := mss.SelectOneOf(u.muxerIDs, nc)
		if err != nil {
			return nil, err
		}
		proto = selected
	}

	if err := nc.SetDeadline(time.Time{}); err != nil {
		return nil, err
	}

	if m := u.getMuxerByID(proto); m != nil {
		return m, nil
	}
	return nil, fmt.Errorf("selected protocol we don't have a transport for")
}

func (u *upgrader) getMuxerByID(id protocol.ID) *StreamMuxer {
	for _, m := range u.muxers {
		if m.ID == id {
			return &m
		}
	}
	return nil
}

func (u *upgrader) setupMuxer(ctx context.Context, conn sec.SecureConn, server bool, scope network.PeerScope) (protocol.ID, network.MuxedConn, error) {
	muxerSelected := conn.ConnState().StreamMultiplexer
	// Use muxer selected from security handshake if available. Otherwise fall back to multistream-selection.
	if len(muxerSelected) > 0 {
		m := u.getMuxerByID(muxerSelected)
		if m == nil {
			return "", nil, fmt.Errorf("selected a muxer we don't know: %s", muxerSelected)
		}
		c, err := m.Muxer.NewConn(conn, server, scope)
		if err != nil {
			return "", nil, err
		}
		return muxerSelected, c, nil
	}

	type result struct {
		smconn  network.MuxedConn
		muxerID protocol.ID
		err     error
	}

	done := make(chan result, 1)
	// TODO: The muxer should take a context.
	go func() {
		m, err := u.negotiateMuxer(conn, server)
		if err != nil {
			done <- result{err: err}
			return
		}
		smconn, err := m.Muxer.NewConn(conn, server, scope)
		done <- result{smconn: smconn, muxerID: m.ID, err: err}
	}()

	select {
	case r := <-done:
		return r.muxerID, r.smconn, r.err
	case <-ctx.Done():
		// interrupt this process
		conn.Close()
		// wait to finish
		<-done
		return "", nil, ctx.Err()
	}
}

func (u *upgrader) getSecurityByID(id protocol.ID) sec.SecureTransport {
	for _, s := range u.security {
		if s.ID() == id {
			return s
		}
	}
	return nil
}

func (u *upgrader) negotiateSecurity(ctx context.Context, insecure net.Conn, server bool) (sec.SecureTransport, bool, error) {
	type result struct {
		proto     protocol.ID
		iamserver bool
		err       error
	}

	done := make(chan result, 1)
	go func() {
		if server {
			var r result
			r.iamserver = true
			r.proto, _, r.err = u.securityMuxer.Negotiate(insecure)
			done <- r
			return
		}
		var r result
		r.proto, r.iamserver, r.err = mss.SelectWithSimopenOrFail(u.securityIDs, insecure)
		done <- r
	}()

	select {
	case r := <-done:
		if r.err != nil {
			return nil, false, r.err
		}
		if s := u.getSecurityByID(r.proto); s != nil {
			return s, r.iamserver, nil
		}
		return nil, false, fmt.Errorf("selected unknown security transport: %s", r.proto)
	case <-ctx.Done():
		// We *must* do this. We have outstanding work on the connection
		// and it's no longer safe to use.
		insecure.Close()
		<-done // wait to stop using the connection.
		return nil, false, ctx.Err()
	}
}
