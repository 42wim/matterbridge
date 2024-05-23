package ldap

import (
	"errors"

	ber "github.com/go-asn1-ber/asn1-ber"
	"github.com/mattermost/mattermost/server/public/shared/mlog"
)

var (
	errRespChanClosed = errors.New("ldap: response channel closed")
	errCouldNotRetMsg = errors.New("ldap: could not retrieve message")
)

type request interface {
	appendTo(*ber.Packet) error
}

type requestFunc func(*ber.Packet) error

func (f requestFunc) appendTo(p *ber.Packet) error {
	return f(p)
}

func (l *Conn) doRequest(req request) (*messageContext, error) {
	packet := ber.Encode(ber.ClassUniversal, ber.TypeConstructed, ber.TagSequence, nil, "LDAP Request")
	packet.AppendChild(ber.NewInteger(ber.ClassUniversal, ber.TypePrimitive, ber.TagInteger, l.nextMessageID(), "MessageID"))
	if err := req.appendTo(packet); err != nil {
		return nil, err
	}

	l.Debug.Log("Sending package", PacketToField(packet))

	msgCtx, err := l.sendMessage(packet)
	if err != nil {
		return nil, err
	}

	l.Debug.Log("Send package", mlog.Int("id", msgCtx.id))
	return msgCtx, nil
}

func (l *Conn) readPacket(msgCtx *messageContext) (*ber.Packet, error) {
	l.Debug.Log("Waiting for response", mlog.Int("id", msgCtx.id))
	packetResponse, ok := <-msgCtx.responses
	if !ok {
		return nil, NewError(ErrorNetwork, errRespChanClosed)
	}
	packet, err := packetResponse.ReadPacket()
	if l.Debug.Enabled() {
		if err := addLDAPDescriptions(packet); err != nil {
			return nil, err
		}
		l.Debug.Log("Got response", mlog.Int("id", msgCtx.id), PacketToField(packet), mlog.Err(err))
	}

	if err != nil {
		return nil, err
	}

	if packet == nil {
		return nil, NewError(ErrorNetwork, errCouldNotRetMsg)
	}

	return packet, nil
}
