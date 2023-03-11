package bdiscord

import (
	"bytes"
	"fmt"
	"strings"
	"sync"

	"github.com/42wim/matterbridge/bridge"
	"github.com/42wim/matterbridge/bridge/config"
	"github.com/42wim/matterbridge/bridge/discord/transmitter"
	"github.com/42wim/matterbridge/bridge/helper"
	"github.com/bwmarrin/discordgo"
	lru "github.com/hashicorp/golang-lru"
)

const (
	MessageLength = 1950
	cFileUpload   = "file_upload"
)

type Bdiscord struct {
	*bridge.Config

	c *discordgo.Session

	nick    string
	userID  string
	guildID string

	channelsMutex  sync.RWMutex
	channels       []*discordgo.Channel
	channelInfoMap map[string]*config.ChannelInfo

	membersMutex  sync.RWMutex
	userMemberMap map[string]*discordgo.Member
	nickMemberMap map[string]*discordgo.Member

	// Webhook specific logic
	useAutoWebhooks bool
	transmitter     *transmitter.Transmitter
	cache           *lru.Cache
}

func New(cfg *bridge.Config) bridge.Bridger {
	newCache, err := lru.New(5000)
	if err != nil {
		cfg.Log.Fatalf("Could not create LRU cache: %v", err)
	}

	b := &Bdiscord{
		Config: cfg,
		cache:  newCache,
	}

	b.userMemberMap = make(map[string]*discordgo.Member)
	b.nickMemberMap = make(map[string]*discordgo.Member)
	b.channelInfoMap = make(map[string]*config.ChannelInfo)

	b.useAutoWebhooks = b.GetBool("AutoWebhooks")
	if b.useAutoWebhooks {
		b.Log.Debug("Using automatic webhooks")
	}
	return b
}

func (b *Bdiscord) Connect() error {
	var err error
	token := b.GetString("Token")
	b.Log.Info("Connecting")
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
	// Add privileged intent for guild member tracking. This is needed to track nicks
	// for display names and @mention translation
	b.c.Identify.Intents = discordgo.MakeIntent(discordgo.IntentsAllWithoutPrivileged |
		discordgo.IntentsGuildMembers)

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
	b.userID = userinfo.ID

	// Try and find this account's guild, and populate channels
	b.channelsMutex.Lock()
	for _, guild := range guilds {
		// Skip, if the server name does not match the visible name or the ID
		if guild.Name != serverName && guild.ID != serverName {
			continue
		}

		// Complain about an ambiguous Server setting. Two Discord servers could have the same title!
		// For IDs, practically this will never happen. It would only trigger if some server's name is also an ID.
		if b.guildID != "" {
			return fmt.Errorf("found multiple Discord servers with the same name %#v, expected to see only one", serverName)
		}

		// Getting this guild's channel could result in a permission error
		b.channels, err = b.c.GuildChannels(guild.ID)
		if err != nil {
			return fmt.Errorf("could not get %#v's channels: %w", b.GetString("Server"), err)
		}

		b.guildID = guild.ID
	}
	b.channelsMutex.Unlock()

	// If we couldn't find a guild, we print extra debug information and return a nice error
	if b.guildID == "" {
		err = fmt.Errorf("could not find Discord server %#v", b.GetString("Server"))
		b.Log.Error(err.Error())

		// Print all of the possible server values
		b.Log.Info("Possible server values:")
		for _, guild := range guilds {
			b.Log.Infof("\t- Server=%#v # by name", guild.Name)
			b.Log.Infof("\t- Server=%#v # by ID", guild.ID)
		}

		// If there are no results, we should say that
		if len(guilds) == 0 {
			b.Log.Info("\t- (none found)")
		}

		return err
	}

	// Legacy note: WebhookURL used to have an actual webhook URL that we would edit,
	// but we stopped doing that due to Discord making rate limits more aggressive.
	//
	// Even older: the same WebhookURL used to be used by every channel, which is usually unexpected.
	// This is no longer possible.
	if b.GetString("WebhookURL") != "" {
		message := "The global WebhookURL setting has been removed. "
		message += "You can get similar \"webhook editing\" behaviour by replacing this line with `AutoWebhooks=true`. "
		message += "If you rely on the old-OLD (non-editing) behaviour, can move the WebhookURL to specific channel sections."
		b.Log.Errorln(message)
		return fmt.Errorf("use of removed WebhookURL setting")
	}

	if b.GetInt("debuglevel") == 2 {
		b.Log.Debug("enabling even more discord debug")
		b.c.Debug = true
	}

	// Initialise webhook management
	b.transmitter = transmitter.New(b.c, b.guildID, "matterbridge", b.useAutoWebhooks)
	b.transmitter.Log = b.Log

	var webhookChannelIDs []string
	for _, channel := range b.Channels {
		channelID := b.getChannelID(channel.Name) // note(qaisjp): this readlocks channelsMutex

		// If a WebhookURL was not explicitly provided for this channel,
		// there are two options: just a regular bot message (ugly) or this is should be webhook sent
		if channel.Options.WebhookURL == "" {
			// If it should be webhook sent, we should enforce this via the transmitter
			if b.useAutoWebhooks {
				webhookChannelIDs = append(webhookChannelIDs, channelID)
			}
			continue
		}

		whID, whToken, ok := b.splitURL(channel.Options.WebhookURL)
		if !ok {
			return fmt.Errorf("failed to parse WebhookURL %#v for channel %#v", channel.Options.WebhookURL, channel.ID)
		}

		b.transmitter.AddWebhook(channelID, &discordgo.Webhook{
			ID:        whID,
			Token:     whToken,
			GuildID:   b.guildID,
			ChannelID: channelID,
		})
	}

	if b.useAutoWebhooks {
		err = b.transmitter.RefreshGuildWebhooks(webhookChannelIDs)
		if err != nil {
			b.Log.WithError(err).Println("transmitter could not refresh guild webhooks")
			return err
		}
	}

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

	b.c.AddHandler(b.messageCreate)
	b.c.AddHandler(b.messageTyping)
	b.c.AddHandler(b.messageUpdate)
	b.c.AddHandler(b.messageDelete)
	b.c.AddHandler(b.messageDeleteBulk)
	b.c.AddHandler(b.memberAdd)
	b.c.AddHandler(b.memberRemove)
	b.c.AddHandler(b.memberUpdate)
	if b.GetInt("debuglevel") == 1 {
		b.c.AddHandler(b.messageEvent)
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

	// Handle prefix hint for unthreaded messages.
	if msg.ParentNotFound() {
		msg.ParentID = ""
	}

	// Use webhook to send the message
	useWebhooks := b.shouldMessageUseWebhooks(&msg)
	if useWebhooks && msg.Event != config.EventMsgDelete && msg.ParentID == "" {
		return b.handleEventWebhook(&msg, channelID)
	}

	return b.handleEventBotUser(&msg, channelID)
}

// handleEventDirect handles events via the bot user
func (b *Bdiscord) handleEventBotUser(msg *config.Message, channelID string) (string, error) {
	b.Log.Debugf("Broadcasting using token (API)")

	// Delete message
	if msg.Event == config.EventMsgDelete {
		if msg.ID == "" {
			return "", nil
		}
		err := b.c.ChannelMessageDelete(channelID, msg.ID)
		return "", err
	}

	// Delete a file
	if msg.Event == config.EventFileDelete {
		if msg.ID == "" {
			return "", nil
		}

		if fi, ok := b.cache.Get(cFileUpload + msg.ID); ok {
			err := b.c.ChannelMessageDelete(channelID, fi.(string)) // nolint:forcetypeassert
			b.cache.Remove(cFileUpload + msg.ID)
			return "", err
		}

		return "", fmt.Errorf("file %s not found", msg.ID)
	}

	// Upload a file if it exists
	if msg.Extra != nil {
		for _, rmsg := range helper.HandleExtra(msg, b.General) {
			rmsg.Text = helper.ClipMessage(rmsg.Text, MessageLength, b.GetString("MessageClipped"))
			if _, err := b.c.ChannelMessageSend(channelID, rmsg.Username+rmsg.Text); err != nil {
				b.Log.Errorf("Could not send message %#v: %s", rmsg, err)
			}
		}
		// check if we have files to upload (from slack, telegram or mattermost)
		if len(msg.Extra["file"]) > 0 {
			return b.handleUploadFile(msg, channelID)
		}
	}

	msg.Text = helper.ClipMessage(msg.Text, MessageLength, b.GetString("MessageClipped"))
	msg.Text = b.replaceUserMentions(msg.Text)

	// Edit message
	if msg.ID != "" {
		_, err := b.c.ChannelMessageEdit(channelID, msg.ID, msg.Username+msg.Text)
		return msg.ID, err
	}

	m := discordgo.MessageSend{
		Content:         msg.Username + msg.Text,
		AllowedMentions: b.getAllowedMentions(),
	}

	if msg.ParentValid() {
		m.Reference = &discordgo.MessageReference{
			MessageID: msg.ParentID,
			ChannelID: channelID,
			GuildID:   b.guildID,
		}
	}

	// Post normal message
	res, err := b.c.ChannelMessageSendComplex(channelID, &m)
	if err != nil {
		return "", err
	}

	return res.ID, nil
}

// handleUploadFile handles native upload of files
func (b *Bdiscord) handleUploadFile(msg *config.Message, channelID string) (string, error) {
	for _, f := range msg.Extra["file"] {
		fi := f.(config.FileInfo)
		file := discordgo.File{
			Name:        fi.Name,
			ContentType: "",
			Reader:      bytes.NewReader(*fi.Data),
		}
		m := discordgo.MessageSend{
			Content:         msg.Username + fi.Comment,
			Files:           []*discordgo.File{&file},
			AllowedMentions: b.getAllowedMentions(),
		}
		res, err := b.c.ChannelMessageSendComplex(channelID, &m)
		if err != nil {
			return "", fmt.Errorf("file upload failed: %s", err)
		}

		// link file_upload_nativeID (file ID from the original bridge) to our upload id
		// so that we can remove this later when it eg needs to be deleted
		b.cache.Add(cFileUpload+fi.NativeID, res.ID)
	}

	return "", nil
}
