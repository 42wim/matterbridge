package gumble

import (
	"time"

	"layeh.com/gumble/gumble/MumbleProto"
)

// RegisteredUser represents a registered user on the server.
type RegisteredUser struct {
	// The registered user's ID.
	UserID uint32
	// The registered user's name.
	Name string
	// The last time the user was seen by the server.
	LastSeen time.Time
	// The last channel the user was seen in.
	LastChannel *Channel

	changed    bool
	deregister bool
}

// SetName sets the new name for the user.
func (r *RegisteredUser) SetName(name string) {
	r.Name = name
	r.changed = true
}

// Deregister will remove the registered user from the server.
func (r *RegisteredUser) Deregister() {
	r.deregister = true
}

// Register will keep the user registered on the server. This is only useful if
// Deregister() was called on the registered user.
func (r *RegisteredUser) Register() {
	r.deregister = false
}

// ACLUser returns an ACLUser for the given registered user.
func (r *RegisteredUser) ACLUser() *ACLUser {
	return &ACLUser{
		UserID: r.UserID,
		Name:   r.Name,
	}
}

// RegisteredUsers is a list of users who are registered on the server.
//
// Whenever a registered user is changed, it does not come into effect until
// the registered user list is sent back to the server.
type RegisteredUsers []*RegisteredUser

func (r RegisteredUsers) writeMessage(client *Client) error {
	packet := MumbleProto.UserList{}

	for _, user := range r {
		if user.deregister || user.changed {
			userListUser := &MumbleProto.UserList_User{
				UserId: &user.UserID,
			}
			if !user.deregister {
				userListUser.Name = &user.Name
			}
			packet.Users = append(packet.Users, userListUser)
		}
	}

	if len(packet.Users) <= 0 {
		return nil
	}
	return client.Conn.WriteProto(&packet)
}
