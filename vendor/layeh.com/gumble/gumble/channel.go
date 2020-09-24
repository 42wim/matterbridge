package gumble

import (
	"github.com/golang/protobuf/proto"
	"layeh.com/gumble/gumble/MumbleProto"
)

// Channel represents a channel in the server's channel tree.
type Channel struct {
	// The channel's unique ID.
	ID uint32
	// The channel's name.
	Name string
	// The channel's parent. nil if the channel is the root channel.
	Parent *Channel
	// The channels directly underneath the channel.
	Children Channels
	// The channels that are linked to the channel.
	Links Channels
	// The users currently in the channel.
	Users Users
	// The channel's description. Contains the empty string if the channel does
	// not have a description, or if it needs to be requested.
	Description string
	// The channel's description hash. nil if Channel.Description has
	// been populated.
	DescriptionHash []byte
	// The maximum number of users allowed in the channel. If the value is zero,
	// the maximum number of users per-channel is dictated by the server's
	// "usersperchannel" setting.
	MaxUsers uint32
	// The position at which the channel should be displayed in an ordered list.
	Position int32
	// Is the channel temporary?
	Temporary bool

	client *Client
}

// IsRoot returns true if the channel is the server's root channel.
func (c *Channel) IsRoot() bool {
	return c.ID == 0
}

// Add will add a sub-channel to the given channel.
func (c *Channel) Add(name string, temporary bool) {
	packet := MumbleProto.ChannelState{
		Parent:    &c.ID,
		Name:      &name,
		Temporary: &temporary,
	}
	c.client.Conn.WriteProto(&packet)
}

// Remove will remove the given channel and all sub-channels from the server's
// channel tree.
func (c *Channel) Remove() {
	packet := MumbleProto.ChannelRemove{
		ChannelId: &c.ID,
	}
	c.client.Conn.WriteProto(&packet)
}

// SetName will set the name of the channel. This will have no effect if the
// channel is the server's root channel.
func (c *Channel) SetName(name string) {
	packet := MumbleProto.ChannelState{
		ChannelId: &c.ID,
		Name:      &name,
	}
	c.client.Conn.WriteProto(&packet)
}

// SetDescription will set the description of the channel.
func (c *Channel) SetDescription(description string) {
	packet := MumbleProto.ChannelState{
		ChannelId:   &c.ID,
		Description: &description,
	}
	c.client.Conn.WriteProto(&packet)
}

// SetPosition will set the position of the channel.
func (c *Channel) SetPosition(position int32) {
	packet := MumbleProto.ChannelState{
		ChannelId: &c.ID,
		Position:  &position,
	}
	c.client.Conn.WriteProto(&packet)
}

// SetMaxUsers will set the maximum number of users allowed in the channel.
func (c *Channel) SetMaxUsers(maxUsers uint32) {
	packet := MumbleProto.ChannelState{
		ChannelId: &c.ID,
		MaxUsers:  &maxUsers,
	}
	c.client.Conn.WriteProto(&packet)
}

// Find returns a channel whose path (by channel name) from the current channel
// is equal to the arguments passed.
//
// For example, given the following server channel tree:
//         Root
//           Child 1
//           Child 2
//             Child 2.1
//             Child 2.2
//               Child 2.2.1
//           Child 3
// To get the "Child 2.2.1" channel:
//         root.Find("Child 2", "Child 2.2", "Child 2.2.1")
func (c *Channel) Find(names ...string) *Channel {
	if len(names) == 0 {
		return c
	}
	for _, child := range c.Children {
		if child.Name == names[0] {
			return child.Find(names[1:]...)
		}
	}
	return nil
}

// RequestDescription requests that the actual channel description
// (i.e. non-hashed) be sent to the client.
func (c *Channel) RequestDescription() {
	packet := MumbleProto.RequestBlob{
		ChannelDescription: []uint32{c.ID},
	}
	c.client.Conn.WriteProto(&packet)
}

// RequestACL requests that the channel's ACL to be sent to the client.
func (c *Channel) RequestACL() {
	packet := MumbleProto.ACL{
		ChannelId: &c.ID,
		Query:     proto.Bool(true),
	}
	c.client.Conn.WriteProto(&packet)
}

// RequestPermission requests that the channel's permission information to be
// sent to the client.
//
// Note: the server will not reply to the request if the client has up-to-date
// permission information.
func (c *Channel) RequestPermission() {
	packet := MumbleProto.PermissionQuery{
		ChannelId: &c.ID,
	}
	c.client.Conn.WriteProto(&packet)
}

// Send will send a text message to the channel.
func (c *Channel) Send(message string, recursive bool) {
	textMessage := TextMessage{
		Message: message,
	}
	if recursive {
		textMessage.Trees = []*Channel{c}
	} else {
		textMessage.Channels = []*Channel{c}
	}
	c.client.Send(&textMessage)
}

// Permission returns the permissions the user has in the channel, or nil if
// the permissions are unknown.
func (c *Channel) Permission() *Permission {
	return c.client.permissions[c.ID]
}

// Link links the given channels to the channel.
func (c *Channel) Link(channel ...*Channel) {
	packet := MumbleProto.ChannelState{
		ChannelId: &c.ID,
		LinksAdd:  make([]uint32, len(channel)),
	}
	for i, ch := range channel {
		packet.LinksAdd[i] = ch.ID
	}
	c.client.Conn.WriteProto(&packet)
}

// Unlink unlinks the given channels from the channel. If no arguments are
// passed, all linked channels are unlinked.
func (c *Channel) Unlink(channel ...*Channel) {
	packet := MumbleProto.ChannelState{
		ChannelId: &c.ID,
	}
	if len(channel) == 0 {
		packet.LinksRemove = make([]uint32, len(c.Links))
		i := 0
		for channelID := range c.Links {
			packet.LinksRemove[i] = channelID
			i++
		}
	} else {
		packet.LinksRemove = make([]uint32, len(channel))
		for i, ch := range channel {
			packet.LinksRemove[i] = ch.ID
		}
	}
	c.client.Conn.WriteProto(&packet)
}
