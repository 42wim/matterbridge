package shibshib

import chatv1 "github.com/harmony-development/shibshib/gen/chat/v1"

type LocatedMessage struct {
	chatv1.MessageWithId

	GuildID   uint64
	ChannelID uint64
}
