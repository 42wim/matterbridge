package gumble

import (
	"layeh.com/gumble/gumble/MumbleProto"
)

// ContextActionType is a bitmask of contexts where a ContextAction can be
// triggered.
type ContextActionType int

// Supported ContextAction contexts.
const (
	ContextActionServer  ContextActionType = ContextActionType(MumbleProto.ContextActionModify_Server)
	ContextActionChannel ContextActionType = ContextActionType(MumbleProto.ContextActionModify_Channel)
	ContextActionUser    ContextActionType = ContextActionType(MumbleProto.ContextActionModify_User)
)

// ContextAction is an triggerable item that has been added by a server-side
// plugin.
type ContextAction struct {
	// The context action type.
	Type ContextActionType
	// The name of the context action.
	Name string
	// The user-friendly description of the context action.
	Label string

	client *Client
}

// Trigger will trigger the context action in the context of the server.
func (c *ContextAction) Trigger() {
	packet := MumbleProto.ContextAction{
		Action: &c.Name,
	}
	c.client.Conn.WriteProto(&packet)
}

// TriggerUser will trigger the context action in the context of the given
// user.
func (c *ContextAction) TriggerUser(user *User) {
	packet := MumbleProto.ContextAction{
		Session: &user.Session,
		Action:  &c.Name,
	}
	c.client.Conn.WriteProto(&packet)
}

// TriggerChannel will trigger the context action in the context of the given
// channel.
func (c *ContextAction) TriggerChannel(channel *Channel) {
	packet := MumbleProto.ContextAction{
		ChannelId: &channel.ID,
		Action:    &c.Name,
	}
	c.client.Conn.WriteProto(&packet)
}
