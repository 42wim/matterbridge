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
			for _, br := range b.Bridges {
				b.handleMessage(msg, br)
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

func (b *Bridge) getDestChannel(msg *config.Message, dest string) string {
	for _, v := range b.Channels {
		if v[msg.Origin] == msg.Channel {
			return v[dest]
		}
	}
	return ""
}

func (b *Bridge) handleMessage(msg config.Message, dest Bridger) {
	if b.ignoreMessage(&msg) {
		return
	}
	if dest.Name() != msg.Origin {
		msg.Channel = b.getDestChannel(&msg, dest.Name())
		if msg.Channel == "" {
			return
		}
		b.modifyMessage(&msg, dest.Name())
		dest.Send(msg)
	}
}

func (b *Bridge) ignoreMessage(msg *config.Message) bool {
	// should we discard messages ?
	for _, entry := range b.ignoreNicks[msg.Origin] {
		if msg.Username == entry {
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

func (b *Bridge) modifyMessage(msg *config.Message, dest string) {
	switch dest {
	case "irc":
		setNickFormat(msg, b.Config.IRC.RemoteNickFormat)
	case "xmpp":
		setNickFormat(msg, b.Config.Xmpp.RemoteNickFormat)
	case "mattermost":
		setNickFormat(msg, b.Config.Mattermost.RemoteNickFormat)
	}
}
