package bridge

import (
	//"fmt"
	"github.com/42wim/matterbridge/bridge/config"
	"github.com/42wim/matterbridge/bridge/irc"
	"github.com/42wim/matterbridge/bridge/mattermost"
	"github.com/42wim/matterbridge/bridge/xmpp"
	log "github.com/Sirupsen/logrus"
	"strings"
)

type Bridge struct {
	*config.Config
	Source      string
	Bridges     []Bridger
	Channels    []map[string]string
	ignoreNicks map[string][]string
}

type Bridger interface {
	Send(msg config.Message) error
	Name() string
	Connect() error
	//Command(cmd string) string
}

func NewBridge(cfg *config.Config) error {
	c := make(chan config.Message)
	b := &Bridge{}
	b.Config = cfg
	if cfg.IRC.Enable {
		b.Bridges = append(b.Bridges, birc.New(cfg, c))
	}
	if cfg.Mattermost.Enable {
		b.Bridges = append(b.Bridges, bmattermost.New(cfg, c))
	}
	if cfg.Xmpp.Enable {
		b.Bridges = append(b.Bridges, bxmpp.New(cfg, c))
	}
	if len(b.Bridges) < 2 {
		log.Fatalf("only %d sections enabled. Need at least 2 sections enabled (eg [IRC] and [mattermost]", len(b.Bridges))
	}
	for _, br := range b.Bridges {
		br.Connect()
	}
	b.mapChannels()
	b.mapIgnores()
	b.handleReceive(c)
	return nil
}

func (b *Bridge) handleReceive(c chan config.Message) {
	for {
		select {
		case msg := <-c:
			m := b.getChannel(msg.Origin, msg.Channel)
			if m == nil {
				continue
			}
			for _, br := range b.Bridges {
				if b.ignoreMessage(msg.Username, msg.Text, msg.Origin) {
					continue
				}
				// do not send to originated bridge
				if br.Name() != msg.Origin {
					msg.Channel = m[br.Name()]
					msgmod := b.modifyMessage(msg, br.Name())
					br.Send(msgmod)
				}
			}
		}
	}
}

func (b *Bridge) mapChannels() error {
	for _, val := range b.Config.Channel {
		m := make(map[string]string)
		m["irc"] = val.IRC
		m["mattermost"] = val.Mattermost
		m["xmpp"] = val.Xmpp
		b.Channels = append(b.Channels, m)
	}
	return nil
}

func (b *Bridge) mapIgnores() {
	m := make(map[string][]string)
	m["irc"] = strings.Fields(b.Config.IRC.IgnoreNicks)
	m["mattermost"] = strings.Fields(b.Config.Mattermost.IgnoreNicks)
	m["xmpp"] = strings.Fields(b.Config.Mattermost.IgnoreNicks)
	b.ignoreNicks = m
}

func (b *Bridge) getChannel(src, name string) map[string]string {
	for _, v := range b.Channels {
		if v[src] == name {
			return v
		}
	}
	return nil
}

func (b *Bridge) ignoreMessage(nick string, message string, protocol string) bool {
	// should we discard messages ?
	for _, entry := range b.ignoreNicks[protocol] {
		if nick == entry {
			return true
		}
	}
	return false
}

func setNickFormat(msg *config.Message, format string) {
	if format == "" {
		msg.Username = msg.Origin + "-" + msg.Username + ": "
		return
	}
	msg.Username = strings.Replace(format, "{NICK}", msg.Username, -1)
	msg.Username = strings.Replace(msg.Username, "{BRIDGE}", msg.Origin, -1)
}

func (b *Bridge) modifyMessage(msg config.Message, dest string) config.Message {
	switch dest {
	case "irc":
		setNickFormat(&msg, b.Config.IRC.RemoteNickFormat)
	case "xmpp":
		setNickFormat(&msg, b.Config.Xmpp.RemoteNickFormat)
	case "mattermost":
		setNickFormat(&msg, b.Config.Mattermost.RemoteNickFormat)
	}
	return msg
}
