package bxmpp

import (
	"github.com/42wim/matterbridge/bridge/config"
	"github.com/42wim/matterbridge/bridge/helper"
	"github.com/matterbridge/go-xmpp"
)

// handleDownloadAvatar downloads the avatar of userid from channel
// sends a EVENT_AVATAR_DOWNLOAD message to the gateway if successful.
// logs an error message if it fails
func (b *Bxmpp) handleDownloadAvatar(avatar xmpp.AvatarData) {
	rmsg := config.Message{
		Username: "system",
		Text:     "avatar",
		Channel:  b.parseChannel(avatar.From),
		Account:  b.Account,
		UserID:   avatar.From,
		Event:    config.EventAvatarDownload,
		Extra:    make(map[string][]interface{}),
	}
	if _, ok := b.avatarMap[avatar.From]; !ok {
		b.Log.Debugf("Avatar.From: %s", avatar.From)

		err := helper.HandleDownloadSize(b.Log, &rmsg, avatar.From+".png", int64(len(avatar.Data)), b.General)
		if err != nil {
			b.Log.Error(err)
			return
		}
		helper.HandleDownloadData(b.Log, &rmsg, avatar.From+".png", rmsg.Text, "", &avatar.Data, b.General)
		b.Log.Debugf("Avatar download complete")
		b.Remote <- rmsg
	}
}
