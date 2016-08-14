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
	kind        string
	Channels    []map[string]string
	ignoreNicks map[string][]string
}

type FancyLog struct {
	irc  *log.Entry
	mm   *log.Entry
	xmpp *log.Entry
}

type Bridger interface {
	Send(msg config.Message) error
	Name() string
	//Command(cmd string) string
}

var flog FancyLog

const Legacy = "legacy"

func initFLog() {
	flog.irc = log.WithFields(log.Fields{"module": "irc"})
	flog.mm = log.WithFields(log.Fields{"module": "mattermost"})
	flog.xmpp = log.WithFields(log.Fields{"module": "xmpp"})
}

func NewBridge(name string, cfg *config.Config, kind string) *Bridge {
	c := make(chan config.Message)
	initFLog()
	b := &Bridge{}
	b.Config = cfg
	if cfg.General.Irc {
		b.Bridges = append(b.Bridges, birc.New(cfg, c))
	}
	if cfg.General.Mattermost {
		b.Bridges = append(b.Bridges, bmattermost.New(cfg, c))
	}
	if cfg.General.Xmpp {
		b.Bridges = append(b.Bridges, bxmpp.New(cfg, c))
	}
	b.mapChannels()
	b.mapIgnores()
	go b.handleReceive(c)
	return b
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
					br.Send(msg)
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
