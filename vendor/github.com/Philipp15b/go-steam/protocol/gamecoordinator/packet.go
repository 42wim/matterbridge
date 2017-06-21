package gamecoordinator

import (
	"bytes"
	. "github.com/Philipp15b/go-steam/protocol"
	. "github.com/Philipp15b/go-steam/protocol/protobuf"
	. "github.com/Philipp15b/go-steam/protocol/steamlang"
	"github.com/golang/protobuf/proto"
)

// An incoming, partially unread message from the Game Coordinator.
type GCPacket struct {
	AppId       uint32
	MsgType     uint32
	IsProto     bool
	GCName      string
	Body        []byte
	TargetJobId JobId
}

func NewGCPacket(wrapper *CMsgGCClient) (*GCPacket, error) {
	packet := &GCPacket{
		AppId:   wrapper.GetAppid(),
		MsgType: wrapper.GetMsgtype(),
		GCName:  wrapper.GetGcname(),
	}

	r := bytes.NewReader(wrapper.GetPayload())
	if IsProto(wrapper.GetMsgtype()) {
		packet.MsgType = packet.MsgType & EMsgMask
		packet.IsProto = true

		header := NewMsgGCHdrProtoBuf()
		err := header.Deserialize(r)
		if err != nil {
			return nil, err
		}
		packet.TargetJobId = JobId(header.Proto.GetJobidTarget())
	} else {
		header := NewMsgGCHdr()
		err := header.Deserialize(r)
		if err != nil {
			return nil, err
		}
		packet.TargetJobId = JobId(header.TargetJobID)
	}

	body := make([]byte, r.Len())
	r.Read(body)
	packet.Body = body

	return packet, nil
}

func (g *GCPacket) ReadProtoMsg(body proto.Message) {
	proto.Unmarshal(g.Body, body)
}

func (g *GCPacket) ReadMsg(body MessageBody) {
	body.Deserialize(bytes.NewReader(g.Body))
}
