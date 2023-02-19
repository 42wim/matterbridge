package bdiscord

import (
	"bytes"

	"github.com/42wim/matterbridge/bridge/config"
	"github.com/42wim/matterbridge/bridge/helper"
	"github.com/bwmarrin/discordgo"
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
		res  *discordgo.Message
		res2 *discordgo.Message
		err  error
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
				Content:         msg.Text,
				Username:        msg.Username,
				AvatarURL:       msg.Avatar,
				AllowedMentions: b.getAllowedMentions(),
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
			content := fi.Comment

			res2, err = b.transmitter.Send(
				channelID,
				&discordgo.WebhookParams{
					Username:        msg.Username,
					AvatarURL:       msg.Avatar,
					Files:           []*discordgo.File{&file},
					Content:         content,
					AllowedMentions: b.getAllowedMentions(),
				},
			)
			if err != nil {
				b.Log.Errorf("Could not send file %#v for message %#v: %s", file, msg, err)
			}
		}
	}

	if msg.Text == "" {
		res = res2
	}

	return res, err
}

func (b *Bdiscord) handleEventWebhook(msg *config.Message, channelID string) (string, error) {
	// skip events
	if msg.Event != "" && msg.Event != config.EventUserAction && msg.Event != config.EventJoinLeave && msg.Event != config.EventTopicChange {
		return "", nil
	}

	// skip empty messages
	if msg.Text == "" && (msg.Extra == nil || len(msg.Extra["file"]) == 0) {
		b.Log.Debugf("Skipping empty message %#v", msg)
		return "", nil
	}

	msg.Text = helper.ClipMessage(msg.Text, MessageLength, b.GetString("MessageClipped"))
	msg.Text = b.replaceUserMentions(msg.Text)
	// discord username must be [0..32] max
	if len(msg.Username) > 32 {
		msg.Username = msg.Username[0:32]
	}

	if msg.ID != "" {
		b.Log.Debugf("Editing webhook message")
		err := b.transmitter.Edit(channelID, msg.ID, &discordgo.WebhookParams{
			Content:         msg.Text,
			Username:        msg.Username,
			AllowedMentions: b.getAllowedMentions(),
		})
		if err == nil {
			return msg.ID, nil
		}
		b.Log.Errorf("Could not edit webhook message: %s", err)
	}

	b.Log.Debugf("Processing webhook sending for message %#v", msg)
	discordMsg, err := b.webhookSend(msg, channelID)
	if err != nil {
		b.Log.Errorf("Could not broadcast via webhook for message %#v: %s", msg, err)
		return "", err
	}
	if discordMsg == nil {
		return "", nil
	}

	return discordMsg.ID, nil
}
