package bdiscord

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/42wim/matterbridge/bridge"
	"github.com/42wim/matterbridge/bridge/config"
	"github.com/42wim/matterbridge/bridge/helper"
	"github.com/bwmarrin/discordgo"
)

const (
	MessageLength               = 1950
	webhookPlaceholderMessageID = "-webhook-message-"
)

type Bdiscord struct {
	*bridge.Config

	c *discordgo.Session

	nick         string
	useChannelID bool
	guildID      string
	webhookID    string
	webhookToken string

	channelsMutex  sync.RWMutex
	channels       []*discordgo.Channel
	channelInfoMap map[string]*config.ChannelInfo

	membersMutex  sync.RWMutex
	userMemberMap map[string]*discordgo.Member
	nickMemberMap map[string]*discordgo.Member
}

func New(cfg *bridge.Config) bridge.Bridger {
	b := &Bdiscord{Config: cfg}
	b.userMemberMap = make(map[string]*discordgo.Member)
	b.nickMemberMap = make(map[string]*discordgo.Member)
	b.channelInfoMap = make(map[string]*config.ChannelInfo)
	if b.GetString("WebhookURL") != "" {
		b.Log.Debug("Configuring Discord Incoming Webhook")
		b.webhookID, b.webhookToken = b.splitURL(b.GetString("WebhookURL"))
	}
	return b
}

func (b *Bdiscord) Connect() error {
	var err error
	var token string
	var guildFound bool
	b.Log.Info("Connecting")
	if b.GetString("WebhookURL") == "" {
		b.Log.Info("Connecting using token")
	} else {
		b.Log.Info("Connecting using webhookurl (for posting) and token")
	}
	if !strings.HasPrefix(b.GetString("Token"), "Bot ") {
		token = "Bot " + b.GetString("Token")
	}
	b.c, err = discordgo.New(token)
	if err != nil {
		return err
	}
	b.Log.Info("Connection succeeded")
	b.c.AddHandler(b.messageCreate)
	b.c.AddHandler(b.memberUpdate)
	b.c.AddHandler(b.messageUpdate)
	b.c.AddHandler(b.messageDelete)
	err = b.c.Open()
	if err != nil {
		return err
	}
	guilds, err := b.c.UserGuilds(100, "", "")
	if err != nil {
		return err
	}
	userinfo, err := b.c.User("@me")
	if err != nil {
		return err
	}
	serverName := strings.Replace(b.GetString("Server"), "ID:", "", -1)
	b.nick = userinfo.Username
	b.channelsMutex.Lock()
	for _, guild := range guilds {
		if guild.Name == serverName || guild.ID == serverName {
			b.channels, err = b.c.GuildChannels(guild.ID)
			b.guildID = guild.ID
			guildFound = true
			if err != nil {
				break
			}
		}
	}
	b.channelsMutex.Unlock()
	if !guildFound {
		msg := fmt.Sprintf("Server \"%s\" not found", b.GetString("Server"))
		err = errors.New(msg)
		b.Log.Error(msg)
		b.Log.Info("Possible values:")
		for _, guild := range guilds {
			b.Log.Infof("Server=\"%s\" # Server name", guild.Name)
			b.Log.Infof("Server=\"%s\" # Server ID", guild.ID)
		}
	}

	if err != nil {
		return err
	}
	b.channelsMutex.RLock()
	for _, channel := range b.channels {
		b.Log.Debugf("found channel %#v", channel)
	}
	b.channelsMutex.RUnlock()

	// Obtaining guild members and initializing nickname mapping.
	b.membersMutex.Lock()
	defer b.membersMutex.Unlock()
	members, err := b.c.GuildMembers(b.guildID, "", 1000)
	if err != nil {
		b.Log.Error("Error obtaining server members: ", err)
		return err
	}
	for _, member := range members {
		if member == nil {
			b.Log.Warnf("Skipping missing information for a user.")
			continue
		}
		b.userMemberMap[member.User.ID] = member
		b.nickMemberMap[member.User.Username] = member
		if member.Nick != "" {
			b.nickMemberMap[member.Nick] = member
		}
	}
	return nil
}

func (b *Bdiscord) Disconnect() error {
	return b.c.Close()
}

func (b *Bdiscord) JoinChannel(channel config.ChannelInfo) error {
	b.channelsMutex.Lock()
	defer b.channelsMutex.Unlock()

	b.channelInfoMap[channel.ID] = &channel
	idcheck := strings.Split(channel.Name, "ID:")
	if len(idcheck) > 1 {
		b.useChannelID = true
	}
	return nil
}

func (b *Bdiscord) Send(msg config.Message) (string, error) {
	b.Log.Debugf("=> Receiving %#v", msg)

	channelID := b.getChannelID(msg.Channel)
	if channelID == "" {
		return "", fmt.Errorf("Could not find channelID for %v", msg.Channel)
	}

	// Make a action /me of the message
	if msg.Event == config.EventUserAction {
		msg.Text = "_" + msg.Text + "_"
	}

	// use initial webhook
	wID := b.webhookID
	wToken := b.webhookToken

	// check if have a channel specific webhook
	b.channelsMutex.RLock()
	if ci, ok := b.channelInfoMap[msg.Channel+b.Account]; ok {
		if ci.Options.WebhookURL != "" {
			wID, wToken = b.splitURL(ci.Options.WebhookURL)
		}
	}
	b.channelsMutex.RUnlock()

	// Use webhook to send the message
	if wID != "" {
		// skip events
		if msg.Event != "" && msg.Event != config.EventJoinLeave && msg.Event != config.EventTopicChange {
			return "", nil
		}
		b.Log.Debugf("Broadcasting using Webhook")
		for _, f := range msg.Extra["file"] {
			fi := f.(config.FileInfo)
			if fi.Comment != "" {
				msg.Text += fi.Comment + ": "
			}
			if fi.URL != "" {
				msg.Text = fi.URL
				if fi.Comment != "" {
					msg.Text = fi.Comment + ": " + fi.URL
				}
			}
		}
		// skip empty messages
		if msg.Text == "" {
			return "", nil
		}

		if msg.ID == webhookPlaceholderMessageID {
			// Received an edit for a message that was previously sent via a
			// webhook.  We can't edit the original message, but we can mark
			// the new version with a pencil icon.
			msg.Text = "[:pencil:] " + msg.Text
		}
		msg.Text = helper.ClipMessage(msg.Text, MessageLength)
		msg.Text = b.replaceUserMentions(msg.Text)
		// discord username must be [0..32] max
		if len(msg.Username) > 32 {
			msg.Username = msg.Username[0:32]
		}
		err := b.c.WebhookExecute(
			wID,
			wToken,
			true,
			&discordgo.WebhookParams{
				Content:   msg.Text,
				Username:  msg.Username,
				AvatarURL: msg.Avatar,
			})
		// Replace with real ID after https://github.com/bwmarrin/discordgo/issues/622
		return webhookPlaceholderMessageID, err
	}

	b.Log.Debugf("Broadcasting using token (API)")

	// Delete message
	if msg.Event == config.EventMsgDelete {
		if msg.ID == "" || msg.ID == webhookPlaceholderMessageID {
			return "", nil
		}
		err := b.c.ChannelMessageDelete(channelID, msg.ID)
		return "", err
	}

	// Upload a file if it exists
	if msg.Extra != nil {
		for _, rmsg := range helper.HandleExtra(&msg, b.General) {
			rmsg.Text = helper.ClipMessage(rmsg.Text, MessageLength)
			if _, err := b.c.ChannelMessageSend(channelID, rmsg.Username+rmsg.Text); err != nil {
				b.Log.Errorf("Could not send message %#v: %v", rmsg, err)
			}
		}
		// check if we have files to upload (from slack, telegram or mattermost)
		if len(msg.Extra["file"]) > 0 {
			return b.handleUploadFile(&msg, channelID)
		}
	}

	msg.Text = helper.ClipMessage(msg.Text, MessageLength)
	msg.Text = b.replaceUserMentions(msg.Text)

	// Edit message
	if msg.ID != "" && msg.ID != webhookPlaceholderMessageID {
		// Note: cannot edit messages made by a webhook.  This should never
		// happen though, because if we get here then it means we don't have a
		// usable webhook.
		_, err := b.c.ChannelMessageEdit(channelID, msg.ID, msg.Username+msg.Text)
		return msg.ID, err
	}

	// Post normal message
	res, err := b.c.ChannelMessageSend(channelID, msg.Username+msg.Text)
	if err != nil {
		return "", err
	}
	return res.ID, err
}

// useWebhook returns true if we have a webhook defined somewhere
func (b *Bdiscord) useWebhook() bool {
	if b.GetString("WebhookURL") != "" {
		return true
	}

	b.channelsMutex.RLock()
	defer b.channelsMutex.RUnlock()

	for _, channel := range b.channelInfoMap {
		if channel.Options.WebhookURL != "" {
			return true
		}
	}
	return false
}

// isWebhookID returns true if the specified id is used in a defined webhook
func (b *Bdiscord) isWebhookID(id string) bool {
	if b.GetString("WebhookURL") != "" {
		wID, _ := b.splitURL(b.GetString("WebhookURL"))
		if wID == id {
			return true
		}
	}

	b.channelsMutex.RLock()
	defer b.channelsMutex.RUnlock()

	for _, channel := range b.channelInfoMap {
		if channel.Options.WebhookURL != "" {
			wID, _ := b.splitURL(channel.Options.WebhookURL)
			if wID == id {
				return true
			}
		}
	}
	return false
}

// handleUploadFile handles native upload of files
func (b *Bdiscord) handleUploadFile(msg *config.Message, channelID string) (string, error) {
	var err error
	for _, f := range msg.Extra["file"] {
		fi := f.(config.FileInfo)
		file := discordgo.File{
			Name:        fi.Name,
			ContentType: "",
			Reader:      bytes.NewReader(*fi.Data),
		}
		m := discordgo.MessageSend{
			Content: msg.Username + fi.Comment,
			Files:   []*discordgo.File{&file},
		}
		_, err = b.c.ChannelMessageSendComplex(channelID, &m)
		if err != nil {
			return "", fmt.Errorf("file upload failed: %#v", err)
		}
	}
	return "", nil
}
