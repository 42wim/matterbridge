package protocol

import (
	"github.com/golang/protobuf/proto"
	. "github.com/Philipp15b/go-steam/protocol/steamlang"
	. "github.com/Philipp15b/go-steam/steamid"
	"io"
)

// Interface for all messages, typically outgoing. They can also be created by
// using the Read* methods in a PacketMsg.
type IMsg interface {
	Serializer
	IsProto() bool
	GetMsgType() EMsg
	GetTargetJobId() JobId
	SetTargetJobId(JobId)
	GetSourceJobId() JobId
	SetSourceJobId(JobId)
}

// Interface for client messages, i.e. messages that are sent after logging in.
// ClientMsgProtobuf and ClientMsg implement this.
type IClientMsg interface {
	IMsg
	GetSessionId() int32
	SetSessionId(int32)
	GetSteamId() SteamId
	SetSteamId(SteamId)
}

// Represents a protobuf backed client message with session data.
type ClientMsgProtobuf struct {
	Header *MsgHdrProtoBuf
	Body   proto.Message
}

func NewClientMsgProtobuf(eMsg EMsg, body proto.Message) *ClientMsgProtobuf {
	hdr := NewMsgHdrProtoBuf()
	hdr.Msg = eMsg
	return &ClientMsgProtobuf{
		Header: hdr,
		Body:   body,
	}
}

func (c *ClientMsgProtobuf) IsProto() bool {
	return true
}

func (c *ClientMsgProtobuf) GetMsgType() EMsg {
	return NewEMsg(uint32(c.Header.Msg))
}

func (c *ClientMsgProtobuf) GetSessionId() int32 {
	return c.Header.Proto.GetClientSessionid()
}

func (c *ClientMsgProtobuf) SetSessionId(session int32) {
	c.Header.Proto.ClientSessionid = &session
}

func (c *ClientMsgProtobuf) GetSteamId() SteamId {
	return SteamId(c.Header.Proto.GetSteamid())
}

func (c *ClientMsgProtobuf) SetSteamId(s SteamId) {
	c.Header.Proto.Steamid = proto.Uint64(uint64(s))
}

func (c *ClientMsgProtobuf) GetTargetJobId() JobId {
	return JobId(c.Header.Proto.GetJobidTarget())
}

func (c *ClientMsgProtobuf) SetTargetJobId(job JobId) {
	c.Header.Proto.JobidTarget = proto.Uint64(uint64(job))
}

func (c *ClientMsgProtobuf) GetSourceJobId() JobId {
	return JobId(c.Header.Proto.GetJobidSource())
}

func (c *ClientMsgProtobuf) SetSourceJobId(job JobId) {
	c.Header.Proto.JobidSource = proto.Uint64(uint64(job))
}

func (c *ClientMsgProtobuf) Serialize(w io.Writer) error {
	err := c.Header.Serialize(w)
	if err != nil {
		return err
	}
	body, err := proto.Marshal(c.Body)
	if err != nil {
		return err
	}
	_, err = w.Write(body)
	return err
}

// Represents a struct backed client message.
type ClientMsg struct {
	Header  *ExtendedClientMsgHdr
	Body    MessageBody
	Payload []byte
}

func NewClientMsg(body MessageBody, payload []byte) *ClientMsg {
	hdr := NewExtendedClientMsgHdr()
	hdr.Msg = body.GetEMsg()
	return &ClientMsg{
		Header:  hdr,
		Body:    body,
		Payload: payload,
	}
}

func (c *ClientMsg) IsProto() bool {
	return true
}

func (c *ClientMsg) GetMsgType() EMsg {
	return c.Header.Msg
}

func (c *ClientMsg) GetSessionId() int32 {
	return c.Header.SessionID
}

func (c *ClientMsg) SetSessionId(session int32) {
	c.Header.SessionID = session
}

func (c *ClientMsg) GetSteamId() SteamId {
	return c.Header.SteamID
}

func (c *ClientMsg) SetSteamId(s SteamId) {
	c.Header.SteamID = s
}

func (c *ClientMsg) GetTargetJobId() JobId {
	return JobId(c.Header.TargetJobID)
}

func (c *ClientMsg) SetTargetJobId(job JobId) {
	c.Header.TargetJobID = uint64(job)
}

func (c *ClientMsg) GetSourceJobId() JobId {
	return JobId(c.Header.SourceJobID)
}

func (c *ClientMsg) SetSourceJobId(job JobId) {
	c.Header.SourceJobID = uint64(job)
}

func (c *ClientMsg) Serialize(w io.Writer) error {
	err := c.Header.Serialize(w)
	if err != nil {
		return err
	}
	err = c.Body.Serialize(w)
	if err != nil {
		return err
	}
	_, err = w.Write(c.Payload)
	return err
}

type Msg struct {
	Header  *MsgHdr
	Body    MessageBody
	Payload []byte
}

func NewMsg(body MessageBody, payload []byte) *Msg {
	hdr := NewMsgHdr()
	hdr.Msg = body.GetEMsg()
	return &Msg{
		Header:  hdr,
		Body:    body,
		Payload: payload,
	}
}

func (m *Msg) GetMsgType() EMsg {
	return m.Header.Msg
}

func (m *Msg) IsProto() bool {
	return false
}

func (m *Msg) GetTargetJobId() JobId {
	return JobId(m.Header.TargetJobID)
}

func (m *Msg) SetTargetJobId(job JobId) {
	m.Header.TargetJobID = uint64(job)
}

func (m *Msg) GetSourceJobId() JobId {
	return JobId(m.Header.SourceJobID)
}

func (m *Msg) SetSourceJobId(job JobId) {
	m.Header.SourceJobID = uint64(job)
}

func (m *Msg) Serialize(w io.Writer) error {
	err := m.Header.Serialize(w)
	if err != nil {
		return err
	}
	err = m.Body.Serialize(w)
	if err != nil {
		return err
	}
	_, err = w.Write(m.Payload)
	return err
}
