package transport

import (
	"errors"
	math "math/rand"
	"sync"
	"time"

	"github.com/status-im/mvds/protobuf"
	"github.com/status-im/mvds/state"
)

// ChannelTransport implements a basic MVDS transport using channels for basic testing purposes.
type ChannelTransport struct {
	sync.Mutex

	offline int

	in  <-chan Packet
	out map[state.PeerID]chan<- Packet
}

func NewChannelTransport(offline int, in <-chan Packet) *ChannelTransport {
	return &ChannelTransport{
		offline: offline,
		in:      in,
		out:     make(map[state.PeerID]chan<- Packet),
	}
}

func (t *ChannelTransport) AddOutput(id state.PeerID, c chan<- Packet) {
	t.out[id] = c
}

func (t *ChannelTransport) Watch() Packet {
	return <-t.in
}

func (t *ChannelTransport) Send(sender state.PeerID, peer state.PeerID, payload *protobuf.Payload) error {
	// @todo we can do this better, we put node onlineness into a goroutine where we just stop the nodes for x seconds
	// outside of this class
	math.Seed(time.Now().UnixNano())
	if math.Intn(100) < t.offline {
		return nil
	}

	c, ok := t.out[peer]
	if !ok {
		return errors.New("peer unknown")
	}

	c <- Packet{Sender: sender, Payload: payload}
	return nil
}
