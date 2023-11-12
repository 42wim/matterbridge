package datasync

import (
	"errors"
	"math"
	"math/rand"
	"time"

	"github.com/status-im/mvds/protobuf"
	"github.com/status-im/mvds/state"
	"github.com/status-im/mvds/transport"
	"go.uber.org/zap"
)

const backoffInterval = 30

var errNotInitialized = errors.New("datasync transport not initialized")
var DatasyncTicker = 300 * time.Millisecond

// It's easier to calculate nextEpoch if we consider seconds as a unit rather than
// 300 ms, so we multiply the result by the ratio
var offsetToSecond = uint64(time.Second / DatasyncTicker)

type NodeTransport struct {
	packets  chan transport.Packet
	logger   *zap.Logger
	dispatch func(state.PeerID, *protobuf.Payload) error
}

func NewNodeTransport() *NodeTransport {
	return &NodeTransport{
		packets: make(chan transport.Packet),
	}
}

func (t *NodeTransport) Init(dispatch func(state.PeerID, *protobuf.Payload) error, logger *zap.Logger) {
	t.dispatch = dispatch
	t.logger = logger
}

func (t *NodeTransport) AddPacket(p transport.Packet) {
	t.packets <- p
}

func (t *NodeTransport) Watch() transport.Packet {
	return <-t.packets
}

func (t *NodeTransport) Send(_ state.PeerID, peer state.PeerID, payload *protobuf.Payload) error {
	if t.dispatch == nil {
		return errNotInitialized
	}

	// We don't return an error otherwise datasync will keep
	// re-trying sending at each epoch
	err := t.dispatch(peer, payload)
	if err != nil {
		t.logger.Error("failed to send message", zap.Error(err))
		return nil
	}

	return nil
}

// CalculateSendTime calculates the next epoch
// at which a message should be sent.
// We randomize it a bit so that not all messages are sent on the same epoch
func CalculateSendTime(count uint64, time int64) int64 {
	return time + int64(uint64(math.Exp2(float64(count-1)))*backoffInterval*offsetToSecond) + int64(rand.Intn(30)) // nolint: gosec
}
