package bdiscord

import (
	"bytes"

	"github.com/42wim/matterbridge/bridge/config"
	"github.com/matterbridge/discordgo"
)

// shouldMessageUseWebhooks checks if have a channel specific webhook, if we're not using auto webhooks
func (b *Bdiscord) shouldMessageUseWebhooks(msg *config.Message) bool {
	if b.useAutoWebhooks {
		return true
	}

	b.channelsMutex.RLock()
	defer b.channelsMutex.RUnlock()
	if ci, ok := b.channelInfoMap[msg.Channel+b.Account]; ok {
		if ci.Options.WebhookURL != "" {
			return true
		}
	}
	return false
}

// maybeGetLocalAvatar checks if UseLocalAvatar contains the message's
// account or protocol, and if so, returns the Discord avatar (if exists)
func (b *Bdiscord) maybeGetLocalAvatar(msg *config.Message) string {
	for _, val := range b.GetStringSlice("UseLocalAvatar") {
		if msg.Protocol != val && msg.Account != val {
			continue
		}

		member, err := b.getGuildMemberByNick(msg.Username)
		if err != nil {
			return ""
		}

		return member.User.AvatarURL("")
	}
	return ""
}

// webhookSend send one or more message via webhook, taking care of file
// uploads (from slack, telegram or mattermost).
// Returns messageID and error.
func (b *Bdiscord) webhookSend(msg *config.Message, channelID string) (*discordgo.Message, error) {
	var (
		res *discordgo.Message
		err error
	)

	// If avatar is unset, mutate the message to include the local avatar (but only if settings say we should do this)
	if msg.Avatar == "" {
		msg.Avatar = b.maybeGetLocalAvatar(msg)
	}

	// WebhookParams can have either `Content` or `File`.

	// We can't send empty messages.
	if msg.Text != "" {
		res, err = b.transmitter.Send(
			channelID,
			&discordgo.WebhookParams{
				Content:   msg.Text,
				Username:  msg.Username,
				AvatarURL: msg.Avatar,
			},
		)
		if err != nil {
			b.Log.Errorf("Could not send text (%s) for message %#v: %s", msg.Text, msg, err)
		}
	}

	if msg.Extra != nil {
		for _, f := range msg.Extra["file"] {
			fi := f.(config.FileInfo)
			file := discordgo.File{
				Name:        fi.Name,
				ContentType: "",
				Reader:      bytes.NewReader(*fi.Data),
			}
			content := ""
			if msg.Text == "" {
				content = fi.Comment
			}
			_, e2 := b.transmitter.Send(
				channelID,
				&discordgo.WebhookParams{
					Username:  msg.Username,
					AvatarURL: msg.Avatar,
					File:      &file,
					Content:   content,
				},
			)
			if e2 != nil {
				b.Log.Errorf("Could not send file %#v for message %#v: %s", file, msg, e2)
			}
		}
	}
	return res, err
}
