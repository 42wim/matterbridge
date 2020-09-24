package gumbleutil

import (
	"layeh.com/gumble/gumble"
)

// UserGroups fetches the group names the given user belongs to in the given
// channel. The slice of group names sent via the returned channel. On error,
// the returned channel is closed without without sending a slice.
func UserGroups(client *gumble.Client, user *gumble.User, channel *gumble.Channel) <-chan []string {
	ch := make(chan []string)

	if !user.IsRegistered() {
		close(ch)
		return ch
	}

	var detacher gumble.Detacher
	listener := Listener{
		Disconnect: func(e *gumble.DisconnectEvent) {
			detacher.Detach()
			close(ch)
		},
		ChannelChange: func(e *gumble.ChannelChangeEvent) {
			if e.Channel == channel && e.Type.Has(gumble.ChannelChangeRemoved) {
				detacher.Detach()
				close(ch)
			}
		},
		PermissionDenied: func(e *gumble.PermissionDeniedEvent) {
			if e.Channel == channel && e.Type == gumble.PermissionDeniedPermission && (e.Permission&gumble.PermissionWrite) != 0 {
				detacher.Detach()
				close(ch)
			}
		},
		ACL: func(e *gumble.ACLEvent) {
			if e.ACL.Channel != channel {
				return
			}
			var names []string
			for _, g := range e.ACL.Groups {
				if (g.UsersAdd[user.UserID] != nil || g.UsersInherited[user.UserID] != nil) && g.UsersRemove[user.UserID] == nil {
					names = append(names, g.Name)
				}
			}
			detacher.Detach()
			ch <- names
			close(ch)
		},
	}
	detacher = client.Config.Attach(&listener)
	channel.RequestACL()

	return ch
}
