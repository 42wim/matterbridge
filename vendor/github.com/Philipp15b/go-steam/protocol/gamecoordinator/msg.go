package gamecoordinator

import (
	"io"

	. "github.com/Philipp15b/go-steam/protocol"
	. "github.com/Philipp15b/go-steam/protocol/steamlang"
	"github.com/golang/protobuf/proto"
)

// An outgoing message to the Game Coordinator.
type IGCMsg interface {
	Serializer
	IsProto() bool
	GetAppId() uint32
	GetMsgType() uint32

	GetTargetJobId() JobId
	SetTargetJobId(JobId)
	GetSourceJobId() JobId
	SetSourceJobId(JobId)
}

type GCMsgProtobuf struct {
	AppId  uint32
	Header *MsgGCHdrProtoBuf
	Body   proto.Message
}

func NewGCMsgProtobuf(appId, msgType uint32, body proto.Message) *GCMsgProtobuf {
	hdr := NewMsgGCHdrProtoBuf()
	hdr.Msg = msgType
	return &GCMsgProtobuf{
		AppId:  appId,
		Header: hdr,
		Body:   body,
	}
}

func (g *GCMsgProtobuf) IsProto() bool {
	return true
}

func (g *GCMsgProtobuf) GetAppId() uint32 {
	return g.AppId
}

func (g *GCMsgProtobuf) GetMsgType() uint32 {
	return g.Header.Msg
}

func (g *GCMsgProtobuf) GetTargetJobId() JobId {
	return JobId(g.Header.Proto.GetJobidTarget())
}

func (g *GCMsgProtobuf) SetTargetJobId(job JobId) {
	g.Header.Proto.JobidTarget = proto.Uint64(uint64(job))
}

func (g *GCMsgProtobuf) GetSourceJobId() JobId {
	return JobId(g.Header.Proto.GetJobidSource())
}

func (g *GCMsgProtobuf) SetSourceJobId(job JobId) {
	g.Header.Proto.JobidSource = proto.Uint64(uint64(job))
}

func (g *GCMsgProtobuf) Serialize(w io.Writer) error {
	err := g.Header.Serialize(w)
	if err != nil {
		return err
	}
	body, err := proto.Marshal(g.Body)
	if err != nil {
		return err
	}
	_, err = w.Write(body)
	return err
}

type GCMsg struct {
	AppId   uint32
	MsgType uint32
	Header  *MsgGCHdr
	Body    Serializer
}

func NewGCMsg(appId, msgType uint32, body Serializer) *GCMsg {
	return &GCMsg{
		AppId:   appId,
		MsgType: msgType,
		Header:  NewMsgGCHdr(),
		Body:    body,
	}
}

func (g *GCMsg) GetMsgType() uint32 {
	return g.MsgType
}

func (g *GCMsg) GetAppId() uint32 {
	return g.AppId
}

func (g *GCMsg) IsProto() bool {
	return false
}

func (g *GCMsg) GetTargetJobId() JobId {
	return JobId(g.Header.TargetJobID)
}

func (g *GCMsg) SetTargetJobId(job JobId) {
	g.Header.TargetJobID = uint64(job)
}

func (g *GCMsg) GetSourceJobId() JobId {
	return JobId(g.Header.SourceJobID)
}

func (g *GCMsg) SetSourceJobId(job JobId) {
	g.Header.SourceJobID = uint64(job)
}

func (g *GCMsg) Serialize(w io.Writer) error {
	err := g.Header.Serialize(w)
	if err != nil {
		return err
	}
	err = g.Body.Serialize(w)
	return err
}
