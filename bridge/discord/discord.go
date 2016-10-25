package bdiscord

import (
	"github.com/42wim/matterbridge/bridge/config"
	log "github.com/Sirupsen/logrus"
	"github.com/bwmarrin/discordgo"
)

type bdiscord struct {
	c        *discordgo.Session
	Config   *config.Protocol
	Remote   chan config.Message
	protocol string
	origin   string
	Channels []*discordgo.Channel
	Nick     string
}

var flog *log.Entry
var protocol = "discord"

func init() {
	flog = log.WithFields(log.Fields{"module": protocol})
}

func New(config config.Protocol, origin string, c chan config.Message) *bdiscord {
	b := &bdiscord{}
	b.Config = &config
	b.Remote = c
	b.protocol = protocol
	b.origin = origin
	return b
}

func (b *bdiscord) Connect() error {
	var err error
	flog.Info("Connecting")
	b.c, err = discordgo.New(b.Config.Token)
	if err != nil {
		flog.Debugf("%#v", err)
		return err
	}
	flog.Info("Connection succeeded")
	b.c.AddHandler(b.messageCreate)
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
			if err != nil {
				flog.Debugf("%#v", err)
				return err
			}
		}
	}
	return nil
}

func (b *bdiscord) FullOrigin() string {
	return b.protocol + "." + b.origin
}

func (b *bdiscord) JoinChannel(channel string) error {
	return nil
}

func (b *bdiscord) Name() string {
	return b.protocol + "." + b.origin
}

func (b *bdiscord) Protocol() string {
	return b.protocol
}

func (b *bdiscord) Origin() string {
	return b.origin
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
	flog.Debugf("Sending message from %s on %s to gateway", m.Author.Username, b.FullOrigin())
	b.Remote <- config.Message{Username: m.Author.Username, Text: m.Content, Channel: b.getChannelName(m.ChannelID),
		Origin: b.origin, Protocol: b.protocol, FullOrigin: b.FullOrigin()}
}

func (b *bdiscord) getChannelID(name string) string {
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
