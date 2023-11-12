package libp2pwebtransport

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io"
	"sync"
	"sync/atomic"
	"time"

	"github.com/libp2p/go-libp2p/core/connmgr"
	ic "github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/pnet"
	tpt "github.com/libp2p/go-libp2p/core/transport"
	"github.com/libp2p/go-libp2p/p2p/security/noise"
	"github.com/libp2p/go-libp2p/p2p/security/noise/pb"
	"github.com/libp2p/go-libp2p/p2p/transport/quicreuse"

	"github.com/benbjohnson/clock"
	logging "github.com/ipfs/go-log/v2"
	ma "github.com/multiformats/go-multiaddr"
	manet "github.com/multiformats/go-multiaddr/net"
	"github.com/multiformats/go-multihash"
	"github.com/quic-go/quic-go"
	"github.com/quic-go/quic-go/http3"
	"github.com/quic-go/webtransport-go"
)

var log = logging.Logger("webtransport")

const webtransportHTTPEndpoint = "/.well-known/libp2p-webtransport"

const errorCodeConnectionGating = 0x47415445 // GATE in ASCII

const certValidity = 14 * 24 * time.Hour

type Option func(*transport) error

func WithClock(cl clock.Clock) Option {
	return func(t *transport) error {
		t.clock = cl
		return nil
	}
}

// WithTLSClientConfig sets a custom tls.Config used for dialing.
// This option is most useful for setting a custom tls.Config.RootCAs certificate pool.
// When dialing a multiaddr that contains a /certhash component, this library will set InsecureSkipVerify and
// overwrite the VerifyPeerCertificate callback.
func WithTLSClientConfig(c *tls.Config) Option {
	return func(t *transport) error {
		t.tlsClientConf = c
		return nil
	}
}

type transport struct {
	privKey ic.PrivKey
	pid     peer.ID
	clock   clock.Clock

	connManager *quicreuse.ConnManager
	rcmgr       network.ResourceManager
	gater       connmgr.ConnectionGater

	listenOnce     sync.Once
	listenOnceErr  error
	certManager    *certManager
	hasCertManager atomic.Bool // set to true once the certManager is initialized
	staticTLSConf  *tls.Config
	tlsClientConf  *tls.Config

	noise *noise.Transport

	connMx sync.Mutex
	conns  map[uint64]*conn // using quic-go's ConnectionTracingKey as map key
}

var _ tpt.Transport = &transport{}
var _ tpt.Resolver = &transport{}
var _ io.Closer = &transport{}

func New(key ic.PrivKey, psk pnet.PSK, connManager *quicreuse.ConnManager, gater connmgr.ConnectionGater, rcmgr network.ResourceManager, opts ...Option) (tpt.Transport, error) {
	if len(psk) > 0 {
		log.Error("WebTransport doesn't support private networks yet.")
		return nil, errors.New("WebTransport doesn't support private networks yet")
	}
	if rcmgr == nil {
		rcmgr = &network.NullResourceManager{}
	}
	id, err := peer.IDFromPrivateKey(key)
	if err != nil {
		return nil, err
	}
	t := &transport{
		pid:         id,
		privKey:     key,
		rcmgr:       rcmgr,
		gater:       gater,
		clock:       clock.New(),
		connManager: connManager,
		conns:       map[uint64]*conn{},
	}
	for _, opt := range opts {
		if err := opt(t); err != nil {
			return nil, err
		}
	}
	n, err := noise.New(noise.ID, key, nil)
	if err != nil {
		return nil, err
	}
	t.noise = n
	return t, nil
}

func (t *transport) Dial(ctx context.Context, raddr ma.Multiaddr, p peer.ID) (tpt.CapableConn, error) {
	scope, err := t.rcmgr.OpenConnection(network.DirOutbound, false, raddr)
	if err != nil {
		log.Debugw("resource manager blocked outgoing connection", "peer", p, "addr", raddr, "error", err)
		return nil, err
	}

	c, err := t.dialWithScope(ctx, raddr, p, scope)
	if err != nil {
		scope.Done()
		return nil, err
	}

	return c, nil
}

func (t *transport) dialWithScope(ctx context.Context, raddr ma.Multiaddr, p peer.ID, scope network.ConnManagementScope) (tpt.CapableConn, error) {
	_, addr, err := manet.DialArgs(raddr)
	if err != nil {
		return nil, err
	}
	url := fmt.Sprintf("https://%s%s?type=noise", addr, webtransportHTTPEndpoint)
	certHashes, err := extractCertHashes(raddr)
	if err != nil {
		return nil, err
	}

	if len(certHashes) == 0 {
		return nil, errors.New("can't dial webtransport without certhashes")
	}

	sni, _ := extractSNI(raddr)

	if err := scope.SetPeer(p); err != nil {
		log.Debugw("resource manager blocked outgoing connection for peer", "peer", p, "addr", raddr, "error", err)
		return nil, err
	}

	maddr, _ := ma.SplitFunc(raddr, func(c ma.Component) bool { return c.Protocol().Code == ma.P_WEBTRANSPORT })
	sess, err := t.dial(ctx, maddr, url, sni, certHashes)
	if err != nil {
		return nil, err
	}
	sconn, err := t.upgrade(ctx, sess, p, certHashes)
	if err != nil {
		sess.CloseWithError(1, "")
		return nil, err
	}
	if t.gater != nil && !t.gater.InterceptSecured(network.DirOutbound, p, sconn) {
		sess.CloseWithError(errorCodeConnectionGating, "")
		return nil, fmt.Errorf("secured connection gated")
	}
	conn := newConn(t, sess, sconn, scope)
	t.addConn(sess, conn)
	return conn, nil
}

func (t *transport) dial(ctx context.Context, addr ma.Multiaddr, url, sni string, certHashes []multihash.DecodedMultihash) (*webtransport.Session, error) {
	var tlsConf *tls.Config
	if t.tlsClientConf != nil {
		tlsConf = t.tlsClientConf.Clone()
	} else {
		tlsConf = &tls.Config{}
	}
	tlsConf.NextProtos = append(tlsConf.NextProtos, http3.NextProtoH3)

	if sni != "" {
		tlsConf.ServerName = sni
	}

	if len(certHashes) > 0 {
		// This is not insecure. We verify the certificate ourselves.
		// See https://www.w3.org/TR/webtransport/#certificate-hashes.
		tlsConf.InsecureSkipVerify = true
		tlsConf.VerifyPeerCertificate = func(rawCerts [][]byte, _ [][]*x509.Certificate) error {
			return verifyRawCerts(rawCerts, certHashes)
		}
	}
	conn, err := t.connManager.DialQUIC(ctx, addr, tlsConf, t.allowWindowIncrease)
	if err != nil {
		return nil, err
	}
	dialer := webtransport.Dialer{
		RoundTripper: &http3.RoundTripper{
			Dial: func(ctx context.Context, addr string, tlsCfg *tls.Config, cfg *quic.Config) (quic.EarlyConnection, error) {
				return conn.(quic.EarlyConnection), nil
			},
		},
	}
	rsp, sess, err := dialer.Dial(ctx, url, nil)
	if err != nil {
		return nil, err
	}
	if rsp.StatusCode < 200 || rsp.StatusCode > 299 {
		return nil, fmt.Errorf("invalid response status code: %d", rsp.StatusCode)
	}
	return sess, err
}

func (t *transport) upgrade(ctx context.Context, sess *webtransport.Session, p peer.ID, certHashes []multihash.DecodedMultihash) (*connSecurityMultiaddrs, error) {
	local, err := toWebtransportMultiaddr(sess.LocalAddr())
	if err != nil {
		return nil, fmt.Errorf("error determining local addr: %w", err)
	}
	remote, err := toWebtransportMultiaddr(sess.RemoteAddr())
	if err != nil {
		return nil, fmt.Errorf("error determining remote addr: %w", err)
	}

	str, err := sess.OpenStreamSync(ctx)
	if err != nil {
		return nil, err
	}
	defer str.Close()

	// Now run a Noise handshake (using early data) and get all the certificate hashes from the server.
	// We will verify that the certhashes we used to dial is a subset of the certhashes we received from the server.
	var verified bool
	n, err := t.noise.WithSessionOptions(noise.EarlyData(newEarlyDataReceiver(func(b *pb.NoiseExtensions) error {
		decodedCertHashes, err := decodeCertHashesFromProtobuf(b.WebtransportCerthashes)
		if err != nil {
			return err
		}
		for _, sent := range certHashes {
			var found bool
			for _, rcvd := range decodedCertHashes {
				if sent.Code == rcvd.Code && bytes.Equal(sent.Digest, rcvd.Digest) {
					found = true
					break
				}
			}
			if !found {
				return fmt.Errorf("missing cert hash: %v", sent)
			}
		}
		verified = true
		return nil
	}), nil))
	if err != nil {
		return nil, fmt.Errorf("failed to create Noise transport: %w", err)
	}
	c, err := n.SecureOutbound(ctx, &webtransportStream{Stream: str, wsess: sess}, p)
	if err != nil {
		return nil, err
	}
	defer c.Close()
	// The Noise handshake _should_ guarantee that our verification callback is called.
	// Double-check just in case.
	if !verified {
		return nil, errors.New("didn't verify")
	}
	return &connSecurityMultiaddrs{
		ConnSecurity:   c,
		ConnMultiaddrs: &connMultiaddrs{local: local, remote: remote},
	}, nil
}

func decodeCertHashesFromProtobuf(b [][]byte) ([]multihash.DecodedMultihash, error) {
	hashes := make([]multihash.DecodedMultihash, 0, len(b))
	for _, h := range b {
		dh, err := multihash.Decode(h)
		if err != nil {
			return nil, fmt.Errorf("failed to decode hash: %w", err)
		}
		hashes = append(hashes, *dh)
	}
	return hashes, nil
}

func (t *transport) CanDial(addr ma.Multiaddr) bool {
	ok, _ := IsWebtransportMultiaddr(addr)
	return ok
}

func (t *transport) Listen(laddr ma.Multiaddr) (tpt.Listener, error) {
	isWebTransport, certhashCount := IsWebtransportMultiaddr(laddr)
	if !isWebTransport {
		return nil, fmt.Errorf("cannot listen on non-WebTransport addr: %s", laddr)
	}
	if certhashCount > 0 {
		return nil, fmt.Errorf("cannot listen on a specific certhash non-WebTransport addr: %s", laddr)
	}
	if t.staticTLSConf == nil {
		t.listenOnce.Do(func() {
			t.certManager, t.listenOnceErr = newCertManager(t.privKey, t.clock)
			t.hasCertManager.Store(true)
		})
		if t.listenOnceErr != nil {
			return nil, t.listenOnceErr
		}
	} else {
		return nil, errors.New("static TLS config not supported on WebTransport")
	}
	tlsConf := t.staticTLSConf.Clone()
	if tlsConf == nil {
		tlsConf = &tls.Config{GetConfigForClient: func(*tls.ClientHelloInfo) (*tls.Config, error) {
			return t.certManager.GetConfig(), nil
		}}
	}
	tlsConf.NextProtos = append(tlsConf.NextProtos, http3.NextProtoH3)

	ln, err := t.connManager.ListenQUIC(laddr, tlsConf, t.allowWindowIncrease)
	if err != nil {
		return nil, err
	}
	return newListener(ln, t, t.staticTLSConf != nil)
}

func (t *transport) Protocols() []int {
	return []int{ma.P_WEBTRANSPORT}
}

func (t *transport) Proxy() bool {
	return false
}

func (t *transport) Close() error {
	t.listenOnce.Do(func() {})
	if t.certManager != nil {
		return t.certManager.Close()
	}
	return nil
}

func (t *transport) allowWindowIncrease(conn quic.Connection, size uint64) bool {
	t.connMx.Lock()
	defer t.connMx.Unlock()

	c, ok := t.conns[conn.Context().Value(quic.ConnectionTracingKey).(uint64)]
	if !ok {
		return false
	}
	return c.allowWindowIncrease(size)
}

func (t *transport) addConn(sess *webtransport.Session, c *conn) {
	t.connMx.Lock()
	t.conns[sess.Context().Value(quic.ConnectionTracingKey).(uint64)] = c
	t.connMx.Unlock()
}

func (t *transport) removeConn(sess *webtransport.Session) {
	t.connMx.Lock()
	delete(t.conns, sess.Context().Value(quic.ConnectionTracingKey).(uint64))
	t.connMx.Unlock()
}

// extractSNI returns what the SNI should be for the given maddr. If there is an
// SNI component in the multiaddr, then it will be returned and
// foundSniComponent will be true. If there's no SNI component, but there is a
// DNS-like component, then that will be returned for the sni and
// foundSniComponent will be false (since we didn't find an actual sni component).
func extractSNI(maddr ma.Multiaddr) (sni string, foundSniComponent bool) {
	ma.ForEach(maddr, func(c ma.Component) bool {
		switch c.Protocol().Code {
		case ma.P_SNI:
			sni = c.Value()
			foundSniComponent = true
			return false
		case ma.P_DNS, ma.P_DNS4, ma.P_DNS6, ma.P_DNSADDR:
			sni = c.Value()
			// Keep going in case we find an `sni` component
			return true
		}
		return true
	})
	return sni, foundSniComponent
}

// Resolve implements transport.Resolver
func (t *transport) Resolve(_ context.Context, maddr ma.Multiaddr) ([]ma.Multiaddr, error) {
	sni, foundSniComponent := extractSNI(maddr)

	if foundSniComponent || sni == "" {
		// The multiaddr already had an sni field, we can keep using it. Or we don't have any sni like thing
		return []ma.Multiaddr{maddr}, nil
	}

	beforeQuicMA, afterIncludingQuicMA := ma.SplitFunc(maddr, func(c ma.Component) bool {
		return c.Protocol().Code == ma.P_QUIC_V1
	})
	quicComponent, afterQuicMA := ma.SplitFirst(afterIncludingQuicMA)
	sniComponent, err := ma.NewComponent(ma.ProtocolWithCode(ma.P_SNI).Name, sni)
	if err != nil {
		return nil, err
	}
	return []ma.Multiaddr{beforeQuicMA.Encapsulate(quicComponent).Encapsulate(sniComponent).Encapsulate(afterQuicMA)}, nil
}

// AddCertHashes adds the current certificate hashes to a multiaddress.
// If called before Listen, it's a no-op.
func (t *transport) AddCertHashes(m ma.Multiaddr) (ma.Multiaddr, bool) {
	if !t.hasCertManager.Load() {
		return m, false
	}
	return m.Encapsulate(t.certManager.AddrComponent()), true
}
