package gumbleutil

import (
	"layeh.com/gumble/gumble"
)

// ChannelPath returns a slice of channel names, starting from the root channel
// to the given channel.
func ChannelPath(channel *gumble.Channel) []string {
	var pieces []string
	for ; channel != nil; channel = channel.Parent {
		pieces = append(pieces, channel.Name)
	}
	for i := 0; i < (len(pieces) / 2); i++ {
		pieces[len(pieces)-1-i], pieces[i] = pieces[i], pieces[len(pieces)-1-i]
	}
	return pieces
}
