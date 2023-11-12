// Package transport contains transport related logic for MVDS.
package transport

import (
	"github.com/status-im/mvds/protobuf"
	"github.com/status-im/mvds/state"
)

type Packet struct {
	Sender  state.PeerID
	Payload *protobuf.Payload
}

// Transport defines an interface allowing for agnostic transport implementations.
type Transport interface {
	Watch() Packet
	Send(sender state.PeerID, peer state.PeerID, payload *protobuf.Payload) error
}
