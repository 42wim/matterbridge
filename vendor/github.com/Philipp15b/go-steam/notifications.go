package steam

import (
	. "github.com/Philipp15b/go-steam/protocol"
	. "github.com/Philipp15b/go-steam/protocol/protobuf"
	. "github.com/Philipp15b/go-steam/protocol/steamlang"
)

type Notifications struct {
	// Maps notification types to their count. If a type is not present in the map,
	// its count is zero.
	notifications map[NotificationType]uint
	client        *Client
}

func newNotifications(client *Client) *Notifications {
	return &Notifications{
		make(map[NotificationType]uint),
		client,
	}
}

func (n *Notifications) HandlePacket(packet *Packet) {
	switch packet.EMsg {
	case EMsg_ClientUserNotifications:
		n.handleClientUserNotifications(packet)
	}
}

type NotificationType uint

const (
	TradeOffer NotificationType = 1
)

func (n *Notifications) handleClientUserNotifications(packet *Packet) {
	msg := new(CMsgClientUserNotifications)
	packet.ReadProtoMsg(msg)

	for _, notification := range msg.GetNotifications() {
		typ := NotificationType(*notification.UserNotificationType)
		count := uint(*notification.Count)
		n.notifications[typ] = count
		n.client.Emit(&NotificationEvent{typ, count})
	}

	// check if there is a notification in our map that isn't in the current packet
	for typ, _ := range n.notifications {
		exists := false
		for _, t := range msg.GetNotifications() {
			if NotificationType(*t.UserNotificationType) == typ {
				exists = true
				break
			}
		}

		if !exists {
			delete(n.notifications, typ)
			n.client.Emit(&NotificationEvent{typ, 0})
		}
	}
}
