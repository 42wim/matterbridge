package gumble

import (
	"net"
	"time"

	"github.com/golang/protobuf/proto"
	"layeh.com/gumble/gumble/MumbleProto"
)

// BanList is a list of server ban entries.
//
// Whenever a ban is changed, it does not come into effect until the ban list
// is sent back to the server.
type BanList []*Ban

// Add creates a new ban list entry with the given parameters.
func (b *BanList) Add(address net.IP, mask net.IPMask, reason string, duration time.Duration) *Ban {
	ban := &Ban{
		Address:  address,
		Mask:     mask,
		Reason:   reason,
		Duration: duration,
	}
	*b = append(*b, ban)
	return ban
}

// Ban represents an entry in the server ban list.
//
// This type should not be initialized manually. Instead, create new ban
// entries using BanList.Add().
type Ban struct {
	// The banned IP address.
	Address net.IP
	// The IP mask that the ban applies to.
	Mask net.IPMask
	// The name of the banned user.
	Name string
	// The certificate hash of the banned user.
	Hash string
	// The reason for the ban.
	Reason string
	// The start time from which the ban applies.
	Start time.Time
	// How long the ban is for.
	Duration time.Duration

	unban bool
}

// SetAddress sets the banned IP address.
func (b *Ban) SetAddress(address net.IP) {
	b.Address = address
}

// SetMask sets the IP mask that the ban applies to.
func (b *Ban) SetMask(mask net.IPMask) {
	b.Mask = mask
}

// SetReason changes the reason for the ban.
func (b *Ban) SetReason(reason string) {
	b.Reason = reason
}

// SetDuration changes the duration of the ban.
func (b *Ban) SetDuration(duration time.Duration) {
	b.Duration = duration
}

// Unban will unban the user from the server.
func (b *Ban) Unban() {
	b.unban = true
}

// Ban will ban the user from the server. This is only useful if Unban() was
// called on the ban entry.
func (b *Ban) Ban() {
	b.unban = false
}

func (b BanList) writeMessage(client *Client) error {
	packet := MumbleProto.BanList{
		Query: proto.Bool(false),
	}

	for _, ban := range b {
		if !ban.unban {
			maskSize, _ := ban.Mask.Size()
			packet.Bans = append(packet.Bans, &MumbleProto.BanList_BanEntry{
				Address:  ban.Address,
				Mask:     proto.Uint32(uint32(maskSize)),
				Reason:   &ban.Reason,
				Duration: proto.Uint32(uint32(ban.Duration / time.Second)),
			})
		}
	}

	return client.Conn.WriteProto(&packet)
}
