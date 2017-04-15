package bdiscord

import (
	"github.com/42wim/matterbridge/bridge/config"
	log "github.com/Sirupsen/logrus"
	"github.com/bwmarrin/discordgo"
	"regexp"
	"strings"
	"sync"
)

type bdiscord struct {
	c             *discordgo.Session
	Config        *config.Protocol
	Remote        chan config.Message
	Account       string
	Channels      []*discordgo.Channel
	Nick          string
	UseChannelID  bool
	userMemberMap map[string]*discordgo.Member
	guildID       string
	sync.RWMutex
}

var flog *log.Entry
var protocol = "discord"

func init() {
	flog = log.WithFields(log.Fields{"module": protocol})
}

func New(cfg config.Protocol, account string, c chan config.Message) *bdiscord {
	b := &bdiscord{}
	b.Config = &cfg
	b.Remote = c
	b.Account = account
	b.userMemberMap = make(map[string]*discordgo.Member)
	return b
}

func (b *bdiscord) Connect() error {
	var err error
	flog.Info("Connecting")
	if !strings.HasPrefix(b.Config.Token, "Bot ") {
		b.Config.Token = "Bot " + b.Config.Token
	}
	b.c, err = discordgo.New(b.Config.Token)
	if err != nil {
		flog.Debugf("%#v", err)
		return err
	}
	flog.Info("Connection succeeded")
	b.c.AddHandler(b.messageCreate)
	b.c.AddHandler(b.memberUpdate)
	b.c.AddHandler(b.messageUpdate)
	err = b.c.Open()
	if err != nil {
		flog.Debugf("%#v", err)
		return err
	}
	guilds, err := b.c.UserGuilds()
	if err != nil {
		flog.Debugf("%#v", err)
		return err
	}
	userinfo, err := b.c.User("@me")
	if err != nil {
		flog.Debugf("%#v", err)
		return err
	}
	b.Nick = userinfo.Username
	for _, guild := range guilds {
		if guild.Name == b.Config.Server {
			b.Channels, err = b.c.GuildChannels(guild.ID)
			b.guildID = guild.ID
			if err != nil {
				flog.Debugf("%#v", err)
				return err
			}
		}
	}
	return nil
}

func (b *bdiscord) Disconnect() error {
	return nil
}

func (b *bdiscord) JoinChannel(channel string) error {
	idcheck := strings.Split(channel, "ID:")
	if len(idcheck) > 1 {
		b.UseChannelID = true
	}
	return nil
}

func (b *bdiscord) Send(msg config.Message) error {
	flog.Debugf("Receiving %#v", msg)
	channelID := b.getChannelID(msg.Channel)
	if channelID == "" {
		flog.Errorf("Could not find channelID for %v", msg.Channel)
		return nil
	}
	b.c.ChannelMessageSend(channelID, msg.Username+msg.Text)
	return nil
}

func (b *bdiscord) messageUpdate(s *discordgo.Session, m *discordgo.MessageUpdate) {
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

func (b *bdiscord) messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	// not relay our own messages
	if m.Author.Username == b.Nick {
		return
	}
	if len(m.Attachments) > 0 {
		for _, attach := range m.Attachments {
			m.Content = m.Content + "\n" + attach.URL
		}
	}
	if m.Content == "" {
		return
	}
	flog.Debugf("Sending message from %s on %s to gateway", m.Author.Username, b.Account)
	channelName := b.getChannelName(m.ChannelID)
	if b.UseChannelID {
		channelName = "ID:" + m.ChannelID
	}
	username := b.getNick(m.Author)
	if len(m.MentionRoles) > 0 {
		m.Message.Content = b.replaceRoleMentions(m.Message.Content)
	}
	m.Message.Content = b.stripCustomoji(m.Message.Content)
	b.Remote <- config.Message{Username: username, Text: m.ContentWithMentionsReplaced(), Channel: channelName,
		Account: b.Account, Avatar: "https://cdn.discordapp.com/avatars/" + m.Author.ID + "/" + m.Author.Avatar + ".jpg"}
}

func (b *bdiscord) memberUpdate(s *discordgo.Session, m *discordgo.GuildMemberUpdate) {
	b.Lock()
	if _, ok := b.userMemberMap[m.Member.User.ID]; ok {
		flog.Debugf("%s: memberupdate: user %s (nick %s) changes nick to %s", b.Account, m.Member.User.Username, b.userMemberMap[m.Member.User.ID].Nick, m.Member.Nick)
	}
	b.userMemberMap[m.Member.User.ID] = m.Member
	b.Unlock()
}

func (b *bdiscord) getNick(user *discordgo.User) string {
	var err error
	b.Lock()
	defer b.Unlock()
	if _, ok := b.userMemberMap[user.ID]; ok {
		if b.userMemberMap[user.ID].Nick != "" {
			// only return if nick is set
			return b.userMemberMap[user.ID].Nick
		}
		// otherwise return username
		return user.Username
	}
	// if we didn't find nick, search for it
	b.userMemberMap[user.ID], err = b.c.GuildMember(b.guildID, user.ID)
	if err != nil {
		return user.Username
	}
	// only return if nick is set
	if b.userMemberMap[user.ID].Nick != "" {
		return b.userMemberMap[user.ID].Nick
	}
	return user.Username
}

func (b *bdiscord) getChannelID(name string) string {
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

func (b *bdiscord) getChannelName(id string) string {
	for _, channel := range b.Channels {
		if channel.ID == id {
			return channel.Name
		}
	}
	return ""
}

func (b *bdiscord) replaceRoleMentions(text string) string {
	roles, err := b.c.GuildRoles(b.guildID)
	if err != nil {
		flog.Debugf("%#v", string(err.(*discordgo.RESTError).ResponseBody))
		return text
	}
	for _, role := range roles {
		text = strings.Replace(text, "<@&"+role.ID+">", "@"+role.Name, -1)
	}
	return text
}

func (b *bdiscord) stripCustomoji(text string) string {
	// <:doge:302803592035958784>
	re := regexp.MustCompile("<(:.*?:)[0-9]+>")
	return re.ReplaceAllString(text, `$1`)
}
