package noise

import (
	"context"
	"net"

	"github.com/libp2p/go-libp2p/core/canonicallog"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/libp2p/go-libp2p/core/sec"
	tptu "github.com/libp2p/go-libp2p/p2p/net/upgrader"
	"github.com/libp2p/go-libp2p/p2p/security/noise/pb"

	manet "github.com/multiformats/go-multiaddr/net"
)

// ID is the protocol ID for noise
const ID = "/noise"
const maxProtoNum = 100

type Transport struct {
	protocolID protocol.ID
	localID    peer.ID
	privateKey crypto.PrivKey
	muxers     []protocol.ID
}

var _ sec.SecureTransport = &Transport{}

// New creates a new Noise transport using the given private key as its
// libp2p identity key.
func New(id protocol.ID, privkey crypto.PrivKey, muxers []tptu.StreamMuxer) (*Transport, error) {
	localID, err := peer.IDFromPrivateKey(privkey)
	if err != nil {
		return nil, err
	}

	muxerIDs := make([]protocol.ID, 0, len(muxers))
	for _, m := range muxers {
		muxerIDs = append(muxerIDs, m.ID)
	}

	return &Transport{
		protocolID: id,
		localID:    localID,
		privateKey: privkey,
		muxers:     muxerIDs,
	}, nil
}

// SecureInbound runs the Noise handshake as the responder.
// If p is empty, connections from any peer are accepted.
func (t *Transport) SecureInbound(ctx context.Context, insecure net.Conn, p peer.ID) (sec.SecureConn, error) {
	responderEDH := newTransportEDH(t)
	c, err := newSecureSession(t, ctx, insecure, p, nil, nil, responderEDH, false, p != "")
	if err != nil {
		addr, maErr := manet.FromNetAddr(insecure.RemoteAddr())
		if maErr == nil {
			canonicallog.LogPeerStatus(100, p, addr, "handshake_failure", "noise", "err", err.Error())
		}
	}
	return SessionWithConnState(c, responderEDH.MatchMuxers(false)), err
}

// SecureOutbound runs the Noise handshake as the initiator.
func (t *Transport) SecureOutbound(ctx context.Context, insecure net.Conn, p peer.ID) (sec.SecureConn, error) {
	initiatorEDH := newTransportEDH(t)
	c, err := newSecureSession(t, ctx, insecure, p, nil, initiatorEDH, nil, true, true)
	if err != nil {
		return c, err
	}
	return SessionWithConnState(c, initiatorEDH.MatchMuxers(true)), err
}

func (t *Transport) WithSessionOptions(opts ...SessionOption) (*SessionTransport, error) {
	st := &SessionTransport{t: t, protocolID: t.protocolID}
	for _, opt := range opts {
		if err := opt(st); err != nil {
			return nil, err
		}
	}
	return st, nil
}

func (t *Transport) ID() protocol.ID {
	return t.protocolID
}

func matchMuxers(initiatorMuxers, responderMuxers []protocol.ID) protocol.ID {
	for _, initMuxer := range initiatorMuxers {
		for _, respMuxer := range responderMuxers {
			if initMuxer == respMuxer {
				return initMuxer
			}
		}
	}
	return ""
}

type transportEarlyDataHandler struct {
	transport      *Transport
	receivedMuxers []protocol.ID
}

var _ EarlyDataHandler = &transportEarlyDataHandler{}

func newTransportEDH(t *Transport) *transportEarlyDataHandler {
	return &transportEarlyDataHandler{transport: t}
}

func (i *transportEarlyDataHandler) Send(context.Context, net.Conn, peer.ID) *pb.NoiseExtensions {
	return &pb.NoiseExtensions{
		StreamMuxers: protocol.ConvertToStrings(i.transport.muxers),
	}
}

func (i *transportEarlyDataHandler) Received(_ context.Context, _ net.Conn, extension *pb.NoiseExtensions) error {
	// Discard messages with size or the number of protocols exceeding extension limit for security.
	if extension != nil && len(extension.StreamMuxers) <= maxProtoNum {
		i.receivedMuxers = protocol.ConvertFromStrings(extension.GetStreamMuxers())
	}
	return nil
}

func (i *transportEarlyDataHandler) MatchMuxers(isInitiator bool) protocol.ID {
	if isInitiator {
		return matchMuxers(i.transport.muxers, i.receivedMuxers)
	}
	return matchMuxers(i.receivedMuxers, i.transport.muxers)
}
