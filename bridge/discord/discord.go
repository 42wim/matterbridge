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

const MessageLength = 1950

type Bdiscord struct {
	*bridge.Config

	c *discordgo.Session

	nick            string
	useChannelID    bool
	guildID         string
	webhookID       string
	webhookToken    string
	canEditWebhooks bool

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
	var guildFound bool
	token := b.GetString("Token")
	b.Log.Info("Connecting")
	if b.GetString("WebhookURL") == "" {
		b.Log.Info("Connecting using token")
	} else {
		b.Log.Info("Connecting using webhookurl (for posting) and token")
	}
	if !strings.HasPrefix(b.GetString("Token"), "Bot ") {
		token = "Bot " + b.GetString("Token")
	}
	// if we have a User token, remove the `Bot` prefix
	if strings.HasPrefix(b.GetString("Token"), "User ") {
		token = strings.Replace(b.GetString("Token"), "User ", "", -1)
	}

	b.c, err = discordgo.New(token)
	if err != nil {
		return err
	}
	b.Log.Info("Connection succeeded")
	b.c.AddHandler(b.messageCreate)
	b.c.AddHandler(b.messageTyping)
	b.c.AddHandler(b.memberUpdate)
	b.c.AddHandler(b.messageUpdate)
	b.c.AddHandler(b.messageDelete)
	b.c.AddHandler(b.messageDeleteBulk)
	b.c.AddHandler(b.memberAdd)
	b.c.AddHandler(b.memberRemove)
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
			if err != nil {
				break
			}
			b.guildID = guild.ID
			guildFound = true
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
	if b.GetString("WebhookURL") == "" {
		for _, channel := range b.channels {
			b.Log.Debugf("found channel %#v", channel)
		}
	} else {
		b.canEditWebhooks = true
		for _, channel := range b.channels {
			b.Log.Debugf("found channel %#v; verifying PermissionManageWebhooks", channel)
			perms, permsErr := b.c.State.UserChannelPermissions(userinfo.ID, channel.ID)
			manageWebhooks := discordgo.PermissionManageWebhooks
			if permsErr != nil || perms&manageWebhooks != manageWebhooks {
				b.Log.Warnf("Can't manage webhooks in channel \"%s\"", channel.Name)
				b.canEditWebhooks = false
			}
		}
		if b.canEditWebhooks {
			b.Log.Info("Can manage webhooks; will edit channel for global webhook on send")
		} else {
			b.Log.Warn("Can't manage webhooks; won't edit channel for global webhook on send")
		}
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

	if msg.Event == config.EventUserTyping {
		if b.GetBool("ShowUserTyping") {
			err := b.c.ChannelTyping(channelID)
			return "", err
		}
		return "", nil
	}

	// Make a action /me of the message
	if msg.Event == config.EventUserAction {
		msg.Text = "_" + msg.Text + "_"
	}

	// use initial webhook configured for the entire Discord account
	isGlobalWebhook := true
	wID := b.webhookID
	wToken := b.webhookToken

	// check if have a channel specific webhook
	b.channelsMutex.RLock()
	if ci, ok := b.channelInfoMap[msg.Channel+b.Account]; ok {
		if ci.Options.WebhookURL != "" {
			wID, wToken = b.splitURL(ci.Options.WebhookURL)
			isGlobalWebhook = false
		}
	}
	b.channelsMutex.RUnlock()

	// Use webhook to send the message
	if wID != "" && msg.Event != config.EventMsgDelete {
		// skip events
		if msg.Event != "" && msg.Event != config.EventJoinLeave && msg.Event != config.EventTopicChange {
			return "", nil
		}

		// If we are editing a message, delete the old message
		if msg.ID != "" {
			b.Log.Debugf("Deleting edited webhook message")
			err := b.c.ChannelMessageDelete(channelID, msg.ID)
			if err != nil {
				b.Log.Errorf("Could not delete edited webhook message: %s", err)
			}
		}

		b.Log.Debugf("Broadcasting using Webhook")

		// skip empty messages
		if msg.Text == "" && (msg.Extra == nil || len(msg.Extra["file"]) == 0) {
			b.Log.Debugf("Skipping empty message %#v", msg)
			return "", nil
		}

		msg.Text = helper.ClipMessage(msg.Text, MessageLength)
		msg.Text = b.replaceUserMentions(msg.Text)
		// discord username must be [0..32] max
		if len(msg.Username) > 32 {
			msg.Username = msg.Username[0:32]
		}
		// if we have a global webhook for this Discord account, and permission
		// to modify webhooks (previously verified), then set its channel to
		// the message channel before using it
		// TODO: this isn't necessary if the last message from this webhook was
		// sent to the current channel
		if isGlobalWebhook && b.canEditWebhooks {
			b.Log.Debugf("Setting webhook channel to \"%s\"", msg.Channel)
			_, err := b.c.WebhookEdit(wID, "", "", channelID)
			if err != nil {
				b.Log.Errorf("Could not set webhook channel: %s", err)
				return "", err
			}
		}
		b.Log.Debugf("Processing webhook sending for message %#v", msg)
		msg, err := b.webhookSend(&msg, wID, wToken)
		if err != nil {
			b.Log.Errorf("Could not broadcast via webook for message %#v: %s", msg, err)
			return "", err
		}
		if msg == nil {
			return "", nil
		}
		return msg.ID, nil
	}

	b.Log.Debugf("Broadcasting using token (API)")

	// Delete message
	if msg.Event == config.EventMsgDelete {
		if msg.ID == "" {
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
				b.Log.Errorf("Could not send message %#v: %s", rmsg, err)
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
	if msg.ID != "" {
		_, err := b.c.ChannelMessageEdit(channelID, msg.ID, msg.Username+msg.Text)
		return msg.ID, err
	}

	// Post normal message
	res, err := b.c.ChannelMessageSend(channelID, msg.Username+msg.Text)
	if err != nil {
		return "", err
	}
	return res.ID, nil
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
			return "", fmt.Errorf("file upload failed: %s", err)
		}
	}
	return "", nil
}

// webhookSend send one or more message via webhook, taking care of file
// uploads (from slack, telegram or mattermost).
// Returns messageID and error.
func (b *Bdiscord) webhookSend(msg *config.Message, webhookID, token string) (*discordgo.Message, error) {
	var (
		res *discordgo.Message
		err error
	)

	// WebhookParams can have either `Content` or `File`.

	// We can't send empty messages.
	if msg.Text != "" {
		res, err = b.c.WebhookExecute(
			webhookID,
			token,
			true,
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
			_, e2 := b.c.WebhookExecute(
				webhookID,
				token,
				false,
				&discordgo.WebhookParams{
					Username:  msg.Username,
					AvatarURL: msg.Avatar,
					File:      &file,
				},
			)
			if e2 != nil {
				b.Log.Errorf("Could not send file %#v for message %#v: %s", file, msg, e2)
			}
		}
	}
	return res, err
}
