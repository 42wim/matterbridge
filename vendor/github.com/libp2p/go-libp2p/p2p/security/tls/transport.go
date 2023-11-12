package libp2ptls

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"os"
	"runtime/debug"

	"github.com/libp2p/go-libp2p/core/canonicallog"
	ci "github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/libp2p/go-libp2p/core/sec"
	tptu "github.com/libp2p/go-libp2p/p2p/net/upgrader"

	manet "github.com/multiformats/go-multiaddr/net"
)

// ID is the protocol ID (used when negotiating with multistream)
const ID = "/tls/1.0.0"

// Transport constructs secure communication sessions for a peer.
type Transport struct {
	identity *Identity

	localPeer  peer.ID
	privKey    ci.PrivKey
	muxers     []protocol.ID
	protocolID protocol.ID
}

var _ sec.SecureTransport = &Transport{}

// New creates a TLS encrypted transport
func New(id protocol.ID, key ci.PrivKey, muxers []tptu.StreamMuxer) (*Transport, error) {
	localPeer, err := peer.IDFromPrivateKey(key)
	if err != nil {
		return nil, err
	}
	muxerIDs := make([]protocol.ID, 0, len(muxers))
	for _, m := range muxers {
		muxerIDs = append(muxerIDs, m.ID)
	}
	t := &Transport{
		protocolID: id,
		localPeer:  localPeer,
		privKey:    key,
		muxers:     muxerIDs,
	}

	identity, err := NewIdentity(key)
	if err != nil {
		return nil, err
	}
	t.identity = identity
	return t, nil
}

// SecureInbound runs the TLS handshake as a server.
// If p is empty, connections from any peer are accepted.
func (t *Transport) SecureInbound(ctx context.Context, insecure net.Conn, p peer.ID) (sec.SecureConn, error) {
	config, keyCh := t.identity.ConfigForPeer(p)
	muxers := make([]string, 0, len(t.muxers))
	for _, muxer := range t.muxers {
		muxers = append(muxers, string(muxer))
	}
	// TLS' ALPN selection lets the server select the protocol, preferring the server's preferences.
	// We want to prefer the client's preference though.
	getConfigForClient := config.GetConfigForClient
	config.GetConfigForClient = func(info *tls.ClientHelloInfo) (*tls.Config, error) {
	alpnLoop:
		for _, proto := range info.SupportedProtos {
			for _, m := range muxers {
				if m == proto {
					// Match found. Select this muxer, as it's the client's preference.
					// There's no need to add the "libp2p" entry here.
					config.NextProtos = []string{proto}
					break alpnLoop
				}
			}
		}
		if getConfigForClient != nil {
			return getConfigForClient(info)
		}
		return config, nil
	}
	config.NextProtos = append(muxers, config.NextProtos...)
	cs, err := t.handshake(ctx, tls.Server(insecure, config), keyCh)
	if err != nil {
		addr, maErr := manet.FromNetAddr(insecure.RemoteAddr())
		if maErr == nil {
			canonicallog.LogPeerStatus(100, p, addr, "handshake_failure", "tls", "err", err.Error())
		}
		insecure.Close()
	}
	return cs, err
}

// SecureOutbound runs the TLS handshake as a client.
// Note that SecureOutbound will not return an error if the server doesn't
// accept the certificate. This is due to the fact that in TLS 1.3, the client
// sends its certificate and the ClientFinished in the same flight, and can send
// application data immediately afterwards.
// If the handshake fails, the server will close the connection. The client will
// notice this after 1 RTT when calling Read.
func (t *Transport) SecureOutbound(ctx context.Context, insecure net.Conn, p peer.ID) (sec.SecureConn, error) {
	config, keyCh := t.identity.ConfigForPeer(p)
	muxers := make([]string, 0, len(t.muxers))
	for _, muxer := range t.muxers {
		muxers = append(muxers, (string)(muxer))
	}
	// Prepend the prefered muxers list to TLS config.
	config.NextProtos = append(muxers, config.NextProtos...)
	cs, err := t.handshake(ctx, tls.Client(insecure, config), keyCh)
	if err != nil {
		insecure.Close()
	}
	return cs, err
}

func (t *Transport) handshake(ctx context.Context, tlsConn *tls.Conn, keyCh <-chan ci.PubKey) (_sconn sec.SecureConn, err error) {
	defer func() {
		if rerr := recover(); rerr != nil {
			fmt.Fprintf(os.Stderr, "panic in TLS handshake: %s\n%s\n", rerr, debug.Stack())
			err = fmt.Errorf("panic in TLS handshake: %s", rerr)

		}
	}()

	// handshaking...
	if err := tlsConn.HandshakeContext(ctx); err != nil {
		return nil, err
	}

	// Should be ready by this point, don't block.
	var remotePubKey ci.PubKey
	select {
	case remotePubKey = <-keyCh:
	default:
	}
	if remotePubKey == nil {
		return nil, errors.New("go-libp2p tls BUG: expected remote pub key to be set")
	}

	return t.setupConn(tlsConn, remotePubKey)
}

func (t *Transport) setupConn(tlsConn *tls.Conn, remotePubKey ci.PubKey) (sec.SecureConn, error) {
	remotePeerID, err := peer.IDFromPublicKey(remotePubKey)
	if err != nil {
		return nil, err
	}

	nextProto := tlsConn.ConnectionState().NegotiatedProtocol
	// The special ALPN extension value "libp2p" is used by libp2p versions
	// that don't support early muxer negotiation. If we see this sepcial
	// value selected, that means we are handshaking with a version that does
	// not support early muxer negotiation. In this case return empty nextProto
	// to indicate no muxer is selected.
	if nextProto == "libp2p" {
		nextProto = ""
	}

	return &conn{
		Conn:         tlsConn,
		localPeer:    t.localPeer,
		remotePeer:   remotePeerID,
		remotePubKey: remotePubKey,
		connectionState: network.ConnectionState{
			StreamMultiplexer:         protocol.ID(nextProto),
			UsedEarlyMuxerNegotiation: nextProto != "",
		},
	}, nil
}

func (t *Transport) ID() protocol.ID {
	return t.protocolID
}
