package torrent

import (
	"fmt"
	"net"
	"time"

	"github.com/anacrolix/torrent/webtorrent"
	"github.com/pion/datachannel"
	"github.com/pion/webrtc/v3"
)

const webrtcNetwork = "webrtc"

type webrtcNetConn struct {
	datachannel.ReadWriteCloser
	webtorrent.DataChannelContext
}

type webrtcNetAddr struct {
	*webrtc.ICECandidate
}

var _ net.Addr = webrtcNetAddr{}

func (webrtcNetAddr) Network() string {
	// Now that we have the ICE candidate, we can tell if it's over udp or tcp. But should we use
	// that for the network?
	return webrtcNetwork
}

func (me webrtcNetAddr) String() string {
	// Probably makes sense to return the IP:port expected of most net.Addrs. I'm not sure if
	// Address would be quoted for IPv6 already. If not, net.JoinHostPort might be appropriate.
	return fmt.Sprintf("%s:%d", me.Address, me.Port)
}

func (me webrtcNetConn) LocalAddr() net.Addr {
	// I'm not sure if this evolves over time. It might also be unavailable if the PeerConnection is
	// closed or closes itself. The same concern applies to RemoteAddr.
	pair, err := me.DataChannelContext.GetSelectedIceCandidatePair()
	if err != nil {
		panic(err)
	}
	return webrtcNetAddr{pair.Local}
}

func (me webrtcNetConn) RemoteAddr() net.Addr {
	// See comments on LocalAddr.
	pair, err := me.DataChannelContext.GetSelectedIceCandidatePair()
	if err != nil {
		panic(err)
	}
	return webrtcNetAddr{pair.Remote}
}

// Do we need these for WebRTC connections exposed as net.Conns? Can we set them somewhere inside
// PeerConnection or on the channel or some transport?

func (w webrtcNetConn) SetDeadline(t time.Time) error {
	return nil
}

func (w webrtcNetConn) SetReadDeadline(t time.Time) error {
	return nil
}

func (w webrtcNetConn) SetWriteDeadline(t time.Time) error {
	return nil
}
