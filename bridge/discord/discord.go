package bdiscord

import (
	"bytes"
	"fmt"
	"github.com/42wim/matterbridge/bridge/config"
	"github.com/42wim/matterbridge/bridge/helper"
	"github.com/bwmarrin/discordgo"
	log "github.com/sirupsen/logrus"
	"regexp"
	"strings"
	"sync"
)

type Bdiscord struct {
	c              *discordgo.Session
	Channels       []*discordgo.Channel
	Nick           string
	UseChannelID   bool
	userMemberMap  map[string]*discordgo.Member
	guildID        string
	webhookID      string
	webhookToken   string
	channelInfoMap map[string]*config.ChannelInfo
	sync.RWMutex
	*config.BridgeConfig
}

var flog *log.Entry
var protocol = "discord"

func init() {
	flog = log.WithFields(log.Fields{"prefix": protocol})
}

func New(cfg *config.BridgeConfig) *Bdiscord {
	b := &Bdiscord{BridgeConfig: cfg}
	b.userMemberMap = make(map[string]*discordgo.Member)
	b.channelInfoMap = make(map[string]*config.ChannelInfo)
	if b.Config.WebhookURL != "" {
		flog.Debug("Configuring Discord Incoming Webhook")
		b.webhookID, b.webhookToken = b.splitURL(b.Config.WebhookURL)
	}
	return b
}

func (b *Bdiscord) Connect() error {
	var err error
	flog.Info("Connecting")
	if b.Config.WebhookURL == "" {
		flog.Info("Connecting using token")
	} else {
		flog.Info("Connecting using webhookurl (for posting) and token")
	}
	if !strings.HasPrefix(b.Config.Token, "Bot ") {
		b.Config.Token = "Bot " + b.Config.Token
	}
	b.c, err = discordgo.New(b.Config.Token)
	if err != nil {
		return err
	}
	flog.Info("Connection succeeded")
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
	b.Nick = userinfo.Username
	for _, guild := range guilds {
		if guild.Name == b.Config.Server {
			b.Channels, err = b.c.GuildChannels(guild.ID)
			b.guildID = guild.ID
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (b *Bdiscord) Disconnect() error {
	return nil
}

func (b *Bdiscord) JoinChannel(channel config.ChannelInfo) error {
	b.channelInfoMap[channel.ID] = &channel
	idcheck := strings.Split(channel.Name, "ID:")
	if len(idcheck) > 1 {
		b.UseChannelID = true
	}
	return nil
}

func (b *Bdiscord) Send(msg config.Message) (string, error) {
	flog.Debugf("Receiving %#v", msg)

	channelID := b.getChannelID(msg.Channel)
	if channelID == "" {
		return "", fmt.Errorf("Could not find channelID for %v", msg.Channel)
	}

	// Make a action /me of the message
	if msg.Event == config.EVENT_USER_ACTION {
		msg.Text = "_" + msg.Text + "_"
	}

	// use initial webhook
	wID := b.webhookID
	wToken := b.webhookToken

	// check if have a channel specific webhook
	if ci, ok := b.channelInfoMap[msg.Channel+b.Account]; ok {
		if ci.Options.WebhookURL != "" {
			wID, wToken = b.splitURL(ci.Options.WebhookURL)
		}
	}

	// Use webhook to send the message
	if wID != "" {
		// skip events
		if msg.Event != "" {
			return "", nil
		}
		flog.Debugf("Broadcasting using Webhook")
		err := b.c.WebhookExecute(
			wID,
			wToken,
			true,
			&discordgo.WebhookParams{
				Content:   msg.Text,
				Username:  msg.Username,
				AvatarURL: msg.Avatar,
			})
		return "", err
	}

	flog.Debugf("Broadcasting using token (API)")

	// Delete message
	if msg.Event == config.EVENT_MSG_DELETE {
		if msg.ID == "" {
			return "", nil
		}
		err := b.c.ChannelMessageDelete(channelID, msg.ID)
		return "", err
	}

	// Upload a file if it exists
	if msg.Extra != nil {
		for _, rmsg := range helper.HandleExtra(&msg, b.General) {
			b.c.ChannelMessageSend(channelID, rmsg.Username+rmsg.Text)
		}
		// check if we have files to upload (from slack, telegram or mattermost)
		if len(msg.Extra["file"]) > 0 {
			return b.handleUploadFile(&msg, channelID)
		}
	}

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
	return res.ID, err
}

func (b *Bdiscord) messageDelete(s *discordgo.Session, m *discordgo.MessageDelete) {
	rmsg := config.Message{Account: b.Account, ID: m.ID, Event: config.EVENT_MSG_DELETE, Text: config.EVENT_MSG_DELETE}
	rmsg.Channel = b.getChannelName(m.ChannelID)
	if b.UseChannelID {
		rmsg.Channel = "ID:" + m.ChannelID
	}
	flog.Debugf("Sending message from %s to gateway", b.Account)
	flog.Debugf("Message is %#v", rmsg)
	b.Remote <- rmsg
}

func (b *Bdiscord) messageUpdate(s *discordgo.Session, m *discordgo.MessageUpdate) {
	if b.Config.EditDisable {
		return
	}
	// only when message is actually edited
	if m.Message.EditedTimestamp != "" {
		flog.Debugf("Sending edit message")
		m.Content = m.Content + b.Config.EditSuffix
		b.messageCreate(s, (*discordgo.MessageCreate)(m))
	}
}

func (b *Bdiscord) messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	var err error

	// not relay our own messages
	if m.Author.Username == b.Nick {
		return
	}
	// if using webhooks, do not relay if it's ours
	if b.useWebhook() && m.Author.Bot && b.isWebhookID(m.Author.ID) {
		return
	}

	// add the url of the attachments to content
	if len(m.Attachments) > 0 {
		for _, attach := range m.Attachments {
			m.Content = m.Content + "\n" + attach.URL
		}
	}

	rmsg := config.Message{Account: b.Account, Avatar: "https://cdn.discordapp.com/avatars/" + m.Author.ID + "/" + m.Author.Avatar + ".jpg", UserID: m.Author.ID, ID: m.ID}

	if m.Content != "" {
		flog.Debugf("Receiving message %#v", m.Message)
		m.Message.Content = b.stripCustomoji(m.Message.Content)
		m.Message.Content = b.replaceChannelMentions(m.Message.Content)
		rmsg.Text, err = m.ContentWithMoreMentionsReplaced(b.c)
		if err != nil {
			flog.Errorf("ContentWithMoreMentionsReplaced failed: %s", err)
			rmsg.Text = m.ContentWithMentionsReplaced()
		}
	}

	// set channel name
	rmsg.Channel = b.getChannelName(m.ChannelID)
	if b.UseChannelID {
		rmsg.Channel = "ID:" + m.ChannelID
	}

	// set username
	if !b.Config.UseUserName {
		rmsg.Username = b.getNick(m.Author)
	} else {
		rmsg.Username = m.Author.Username
	}

	// if we have embedded content add it to text
	if b.Config.ShowEmbeds && m.Message.Embeds != nil {
		for _, embed := range m.Message.Embeds {
			rmsg.Text = rmsg.Text + "embed: " + embed.Title + " - " + embed.Description + " - " + embed.URL + "\n"
		}
	}

	// no empty messages
	if rmsg.Text == "" {
		return
	}

	// do we have a /me action
	var ok bool
	rmsg.Text, ok = b.replaceAction(rmsg.Text)
	if ok {
		rmsg.Event = config.EVENT_USER_ACTION
	}

	flog.Debugf("Sending message from %s on %s to gateway", m.Author.Username, b.Account)
	flog.Debugf("Message is %#v", rmsg)
	b.Remote <- rmsg
}

func (b *Bdiscord) memberUpdate(s *discordgo.Session, m *discordgo.GuildMemberUpdate) {
	b.Lock()
	if _, ok := b.userMemberMap[m.Member.User.ID]; ok {
		flog.Debugf("%s: memberupdate: user %s (nick %s) changes nick to %s", b.Account, m.Member.User.Username, b.userMemberMap[m.Member.User.ID].Nick, m.Member.Nick)
	}
	b.userMemberMap[m.Member.User.ID] = m.Member
	b.Unlock()
}

func (b *Bdiscord) getNick(user *discordgo.User) string {
	var err error
	b.Lock()
	defer b.Unlock()
	if _, ok := b.userMemberMap[user.ID]; ok {
		if b.userMemberMap[user.ID] != nil {
			if b.userMemberMap[user.ID].Nick != "" {
				// only return if nick is set
				return b.userMemberMap[user.ID].Nick
			}
			// otherwise return username
			return user.Username
		}
	}
	// if we didn't find nick, search for it
	member, err := b.c.GuildMember(b.guildID, user.ID)
	if err != nil {
		return user.Username
	}
	b.userMemberMap[user.ID] = member
	// only return if nick is set
	if b.userMemberMap[user.ID].Nick != "" {
		return b.userMemberMap[user.ID].Nick
	}
	return user.Username
}

func (b *Bdiscord) getChannelID(name string) string {
	idcheck := strings.Split(name, "ID:")
	if len(idcheck) > 1 {
		return idcheck[1]
	}
	for _, channel := range b.Channels {
		if channel.Name == name {
			return channel.ID
		}
	}
	return ""
}

func (b *Bdiscord) getChannelName(id string) string {
	for _, channel := range b.Channels {
		if channel.ID == id {
			return channel.Name
		}
	}
	return ""
}

func (b *Bdiscord) replaceChannelMentions(text string) string {
	var err error
	re := regexp.MustCompile("<#[0-9]+>")
	text = re.ReplaceAllStringFunc(text, func(m string) string {
		channel := b.getChannelName(m[2 : len(m)-1])
		// if at first don't succeed, try again
		if channel == "" {
			b.Channels, err = b.c.GuildChannels(b.guildID)
			if err != nil {
				return "#unknownchannel"
			}
			channel = b.getChannelName(m[2 : len(m)-1])
			return "#" + channel
		}
		return "#" + channel
	})
	return text
}

func (b *Bdiscord) replaceAction(text string) (string, bool) {
	if strings.HasPrefix(text, "_") && strings.HasSuffix(text, "_") {
		return strings.Replace(text, "_", "", -1), true
	}
	return text, false
}

func (b *Bdiscord) stripCustomoji(text string) string {
	// <:doge:302803592035958784>
	re := regexp.MustCompile("<(:.*?:)[0-9]+>")
	return re.ReplaceAllString(text, `$1`)
}

// splitURL splits a webhookURL and returns the id and token
func (b *Bdiscord) splitURL(url string) (string, string) {
	webhookURLSplit := strings.Split(url, "/")
	if len(webhookURLSplit) != 7 {
		log.Fatalf("%s is no correct discord WebhookURL", url)
	}
	return webhookURLSplit[len(webhookURLSplit)-2], webhookURLSplit[len(webhookURLSplit)-1]
}

// useWebhook returns true if we have a webhook defined somewhere
func (b *Bdiscord) useWebhook() bool {
	if b.Config.WebhookURL != "" {
		return true
	}
	for _, channel := range b.channelInfoMap {
		if channel.Options.WebhookURL != "" {
			return true
		}
	}
	return false
}

// isWebhookID returns true if the specified id is used in a defined webhook
func (b *Bdiscord) isWebhookID(id string) bool {
	if b.Config.WebhookURL != "" {
		wID, _ := b.splitURL(b.Config.WebhookURL)
		if wID == id {
			return true
		}
	}
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
		files := []*discordgo.File{}
		files = append(files, &discordgo.File{fi.Name, "", bytes.NewReader(*fi.Data)})
		_, err = b.c.ChannelMessageSendComplex(channelID, &discordgo.MessageSend{Content: msg.Username + fi.Comment, Files: files})
		if err != nil {
			return "", fmt.Errorf("file upload failed: %#v", err)
		}
	}
	return "", nil
}
