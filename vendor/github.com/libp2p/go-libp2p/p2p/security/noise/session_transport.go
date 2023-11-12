package noise

import (
	"context"
	"net"

	"github.com/libp2p/go-libp2p/core/canonicallog"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/libp2p/go-libp2p/core/sec"
	"github.com/libp2p/go-libp2p/p2p/security/noise/pb"

	manet "github.com/multiformats/go-multiaddr/net"
)

type SessionOption = func(*SessionTransport) error

// Prologue sets a prologue for the Noise session.
// The handshake will only complete successfully if both parties set the same prologue.
// See https://noiseprotocol.org/noise.html#prologue for details.
func Prologue(prologue []byte) SessionOption {
	return func(s *SessionTransport) error {
		s.prologue = prologue
		return nil
	}
}

// EarlyDataHandler defines what the application payload is for either the second
// (if responder) or third (if initiator) handshake message, and defines the
// logic for handling the other side's early data. Note the early data in the
// second handshake message is encrypted, but the peer is not authenticated at that point.
type EarlyDataHandler interface {
	// Send for the initiator is called for the client before sending the third
	// handshake message. Defines the application payload for the third message.
	// Send for the responder is called before sending the second handshake message.
	Send(context.Context, net.Conn, peer.ID) *pb.NoiseExtensions
	// Received for the initiator is called when the second handshake message
	// from the responder is received.
	// Received for the responder is called when the third handshake message
	// from the initiator is received.
	Received(context.Context, net.Conn, *pb.NoiseExtensions) error
}

// EarlyData sets the `EarlyDataHandler` for the initiator and responder roles.
// See `EarlyDataHandler` for more details.
func EarlyData(initiator, responder EarlyDataHandler) SessionOption {
	return func(s *SessionTransport) error {
		s.initiatorEarlyDataHandler = initiator
		s.responderEarlyDataHandler = responder
		return nil
	}
}

// DisablePeerIDCheck disables checking the remote peer ID for a noise connection.
// For outbound connections, this is the equivalent of calling `SecureInbound` with an empty
// peer ID. This is susceptible to MITM attacks since we do not verify the identity of the remote
// peer.
func DisablePeerIDCheck() SessionOption {
	return func(s *SessionTransport) error {
		s.disablePeerIDCheck = true
		return nil
	}
}

var _ sec.SecureTransport = &SessionTransport{}

// SessionTransport can be used
// to provide per-connection options
type SessionTransport struct {
	t *Transport
	// options
	prologue           []byte
	disablePeerIDCheck bool

	protocolID protocol.ID

	initiatorEarlyDataHandler, responderEarlyDataHandler EarlyDataHandler
}

// SecureInbound runs the Noise handshake as the responder.
// If p is empty, connections from any peer are accepted.
func (i *SessionTransport) SecureInbound(ctx context.Context, insecure net.Conn, p peer.ID) (sec.SecureConn, error) {
	checkPeerID := !i.disablePeerIDCheck && p != ""
	c, err := newSecureSession(i.t, ctx, insecure, p, i.prologue, i.initiatorEarlyDataHandler, i.responderEarlyDataHandler, false, checkPeerID)
	if err != nil {
		addr, maErr := manet.FromNetAddr(insecure.RemoteAddr())
		if maErr == nil {
			canonicallog.LogPeerStatus(100, p, addr, "handshake_failure", "noise", "err", err.Error())
		}
	}
	return c, err
}

// SecureOutbound runs the Noise handshake as the initiator.
func (i *SessionTransport) SecureOutbound(ctx context.Context, insecure net.Conn, p peer.ID) (sec.SecureConn, error) {
	return newSecureSession(i.t, ctx, insecure, p, i.prologue, i.initiatorEarlyDataHandler, i.responderEarlyDataHandler, true, !i.disablePeerIDCheck)
}

func (i *SessionTransport) ID() protocol.ID {
	return i.protocolID
}
