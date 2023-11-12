package datasync

import (
	"crypto/ecdsa"
	"errors"

	"github.com/golang/protobuf/proto"
	datasyncnode "github.com/status-im/mvds/node"
	"github.com/status-im/mvds/protobuf"
	datasyncproto "github.com/status-im/mvds/protobuf"
	datasynctransport "github.com/status-im/mvds/transport"
	"go.uber.org/zap"

	datasyncpeer "github.com/status-im/status-go/protocol/datasync/peer"
)

type DataSync struct {
	*datasyncnode.Node
	// NodeTransport is the implementation of the datasync transport interface.
	*NodeTransport
	logger         *zap.Logger
	sendingEnabled bool
}

func New(node *datasyncnode.Node, transport *NodeTransport, sendingEnabled bool, logger *zap.Logger) *DataSync {
	return &DataSync{Node: node, NodeTransport: transport, sendingEnabled: sendingEnabled, logger: logger}
}

// Unwrap tries to unwrap datasync message and passes back the message to datasync in order to acknowledge any potential message and mark messages as acknowledged
func (d *DataSync) Unwrap(sender *ecdsa.PublicKey, payload []byte) (*protobuf.Payload, error) {
	logger := d.logger.With(zap.String("site", "Handle"))

	datasyncMessage, err := unwrap(payload)
	// If it failed to decode is not a protobuf message, if it successfully decoded but body is empty, is likedly a protobuf wrapped message
	if err != nil {
		logger.Debug("Unwrapping datasync message failed", zap.Error(err))
		return nil, err
	} else if !datasyncMessage.IsValid() {
		return nil, errors.New("handling non-datasync message")
	} else {
		logger.Debug("handling datasync message")
		if d.sendingEnabled {
			d.add(sender, &datasyncMessage)
		}
	}

	return &datasyncMessage, nil
}

func (d *DataSync) Stop() {
	d.Node.Stop()
}

func (d *DataSync) add(publicKey *ecdsa.PublicKey, datasyncMessage *datasyncproto.Payload) {
	packet := datasynctransport.Packet{
		Sender:  datasyncpeer.PublicKeyToPeerID(*publicKey),
		Payload: datasyncMessage,
	}
	d.NodeTransport.AddPacket(packet)
}

func unwrap(payload []byte) (datasyncPayload datasyncproto.Payload, err error) {
	err = proto.Unmarshal(payload, &datasyncPayload)
	return
}
