package quicreuse

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net"
	"strings"
	"sync"

	"github.com/libp2p/go-libp2p/core/transport"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/quic-go/quic-go"
)

type Listener interface {
	Accept(context.Context) (quic.Connection, error)
	Addr() net.Addr
	Multiaddrs() []ma.Multiaddr
	io.Closer
}

type protoConf struct {
	ln                  *listener
	tlsConf             *tls.Config
	allowWindowIncrease func(conn quic.Connection, delta uint64) bool
}

type quicListener struct {
	l         *quic.Listener
	transport refCountedQuicTransport
	running   chan struct{}
	addrs     []ma.Multiaddr

	protocolsMu sync.Mutex
	protocols   map[string]protoConf
}

func newQuicListener(tr refCountedQuicTransport, quicConfig *quic.Config, enableDraft29 bool) (*quicListener, error) {
	localMultiaddrs := make([]ma.Multiaddr, 0, 2)
	a, err := ToQuicMultiaddr(tr.LocalAddr(), quic.Version1)
	if err != nil {
		return nil, err
	}
	localMultiaddrs = append(localMultiaddrs, a)
	if enableDraft29 {
		a, err := ToQuicMultiaddr(tr.LocalAddr(), quic.VersionDraft29)
		if err != nil {
			return nil, err
		}
		localMultiaddrs = append(localMultiaddrs, a)
	}
	cl := &quicListener{
		protocols: map[string]protoConf{},
		running:   make(chan struct{}),
		transport: tr,
		addrs:     localMultiaddrs,
	}
	tlsConf := &tls.Config{
		GetConfigForClient: func(info *tls.ClientHelloInfo) (*tls.Config, error) {
			cl.protocolsMu.Lock()
			defer cl.protocolsMu.Unlock()
			for _, proto := range info.SupportedProtos {
				if entry, ok := cl.protocols[proto]; ok {
					conf := entry.tlsConf
					if conf.GetConfigForClient != nil {
						return conf.GetConfigForClient(info)
					}
					return conf, nil
				}
			}
			return nil, fmt.Errorf("no supported protocol found. offered: %+v", info.SupportedProtos)
		},
	}
	quicConf := quicConfig.Clone()
	quicConf.AllowConnectionWindowIncrease = cl.allowWindowIncrease
	ln, err := tr.Listen(tlsConf, quicConf)
	if err != nil {
		return nil, err
	}
	cl.l = ln
	go cl.Run() // This go routine shuts down once the underlying quic.Listener is closed (or returns an error).
	return cl, nil
}

func (l *quicListener) allowWindowIncrease(conn quic.Connection, delta uint64) bool {
	l.protocolsMu.Lock()
	defer l.protocolsMu.Unlock()

	conf, ok := l.protocols[conn.ConnectionState().TLS.ConnectionState.NegotiatedProtocol]
	if !ok {
		return false
	}
	return conf.allowWindowIncrease(conn, delta)
}

func (l *quicListener) Add(tlsConf *tls.Config, allowWindowIncrease func(conn quic.Connection, delta uint64) bool, onRemove func()) (Listener, error) {
	l.protocolsMu.Lock()
	defer l.protocolsMu.Unlock()

	if len(tlsConf.NextProtos) == 0 {
		return nil, errors.New("no ALPN found in tls.Config")
	}

	for _, proto := range tlsConf.NextProtos {
		if _, ok := l.protocols[proto]; ok {
			return nil, fmt.Errorf("already listening for protocol %s", proto)
		}
	}

	ln := newSingleListener(l.l.Addr(), l.addrs, func() {
		l.protocolsMu.Lock()
		for _, proto := range tlsConf.NextProtos {
			delete(l.protocols, proto)
		}
		l.protocolsMu.Unlock()
		onRemove()
	}, l.running)
	for _, proto := range tlsConf.NextProtos {
		l.protocols[proto] = protoConf{
			ln:                  ln,
			tlsConf:             tlsConf,
			allowWindowIncrease: allowWindowIncrease,
		}
	}
	return ln, nil
}

func (l *quicListener) Run() error {
	defer close(l.running)
	defer l.transport.DecreaseCount()
	for {
		conn, err := l.l.Accept(context.Background())
		if err != nil {
			if errors.Is(err, quic.ErrServerClosed) || strings.Contains(err.Error(), "use of closed network connection") {
				return transport.ErrListenerClosed
			}
			return err
		}
		proto := conn.ConnectionState().TLS.NegotiatedProtocol

		l.protocolsMu.Lock()
		ln, ok := l.protocols[proto]
		if !ok {
			l.protocolsMu.Unlock()
			return fmt.Errorf("negotiated unknown protocol: %s", proto)
		}
		ln.ln.add(conn)
		l.protocolsMu.Unlock()
	}
}

func (l *quicListener) Close() error {
	err := l.l.Close()
	<-l.running // wait for Run to return
	return err
}

const queueLen = 16

// A listener for a single ALPN protocol (set).
type listener struct {
	queue             chan quic.Connection
	acceptLoopRunning chan struct{}
	addr              net.Addr
	addrs             []ma.Multiaddr
	remove            func()
	closeOnce         sync.Once
}

var _ Listener = &listener{}

func newSingleListener(addr net.Addr, addrs []ma.Multiaddr, remove func(), running chan struct{}) *listener {
	return &listener{
		queue:             make(chan quic.Connection, queueLen),
		acceptLoopRunning: running,
		remove:            remove,
		addr:              addr,
		addrs:             addrs,
	}
}

func (l *listener) add(c quic.Connection) {
	select {
	case l.queue <- c:
	default:
		c.CloseWithError(1, "queue full")
	}
}

func (l *listener) Accept(ctx context.Context) (quic.Connection, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-l.acceptLoopRunning:
		return nil, transport.ErrListenerClosed
	case c, ok := <-l.queue:
		if !ok {
			return nil, transport.ErrListenerClosed
		}
		return c, nil
	}
}

func (l *listener) Addr() net.Addr {
	return l.addr
}

func (l *listener) Multiaddrs() []ma.Multiaddr {
	return l.addrs
}

func (l *listener) Close() error {
	l.closeOnce.Do(func() {
		l.remove()
		close(l.queue)
		// drain the queue
		for conn := range l.queue {
			conn.CloseWithError(1, "closing")
		}
	})
	return nil
}
