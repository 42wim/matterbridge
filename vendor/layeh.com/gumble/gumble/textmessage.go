package gumble

import (
	"layeh.com/gumble/gumble/MumbleProto"
)

// TextMessage is a chat message that can be received from and sent to the
// server.
type TextMessage struct {
	// User who sent the message (can be nil).
	Sender *User
	// Users that receive the message.
	Users []*User
	// Channels that receive the message.
	Channels []*Channel
	// Channels that receive the message and send it recursively to sub-channels.
	Trees []*Channel
	// Chat message.
	Message string
}

func (t *TextMessage) writeMessage(client *Client) error {
	packet := MumbleProto.TextMessage{
		Message: &t.Message,
	}
	if t.Users != nil {
		packet.Session = make([]uint32, len(t.Users))
		for i, user := range t.Users {
			packet.Session[i] = user.Session
		}
	}
	if t.Channels != nil {
		packet.ChannelId = make([]uint32, len(t.Channels))
		for i, channel := range t.Channels {
			packet.ChannelId[i] = channel.ID
		}
	}
	if t.Trees != nil {
		packet.TreeId = make([]uint32, len(t.Trees))
		for i, channel := range t.Trees {
			packet.TreeId[i] = channel.ID
		}
	}
	return client.Conn.WriteProto(&packet)
}
