package bridge

import (
	"github.com/42wim/matterbridge/bridge/api"
	"github.com/42wim/matterbridge/bridge/config"
	"github.com/42wim/matterbridge/bridge/discord"
	"github.com/42wim/matterbridge/bridge/gitter"
	"github.com/42wim/matterbridge/bridge/irc"
	"github.com/42wim/matterbridge/bridge/matrix"
	"github.com/42wim/matterbridge/bridge/mattermost"
	"github.com/42wim/matterbridge/bridge/rocketchat"
	"github.com/42wim/matterbridge/bridge/slack"
	"github.com/42wim/matterbridge/bridge/steam"
	"github.com/42wim/matterbridge/bridge/telegram"
	"github.com/42wim/matterbridge/bridge/xmpp"
	log "github.com/Sirupsen/logrus"

	"strings"
)

type Bridger interface {
	Send(msg config.Message) error
	Connect() error
	JoinChannel(channel config.ChannelInfo) error
	Disconnect() error
}

type Bridge struct {
	Config config.Protocol
	Bridger
	Name     string
	Account  string
	Protocol string
	Channels map[string]config.ChannelInfo
	Joined   map[string]bool
}

func New(cfg *config.Config, bridge *config.Bridge, c chan config.Message) *Bridge {
	b := new(Bridge)
	b.Channels = make(map[string]config.ChannelInfo)
	accInfo := strings.Split(bridge.Account, ".")
	protocol := accInfo[0]
	name := accInfo[1]
	b.Name = name
	b.Protocol = protocol
	b.Account = bridge.Account
	b.Joined = make(map[string]bool)

	// override config from environment
	config.OverrideCfgFromEnv(cfg, protocol, name)
	switch protocol {
	case "mattermost":
		b.Config = cfg.Mattermost[name]
		b.Bridger = bmattermost.New(cfg.Mattermost[name], bridge.Account, c)
	case "irc":
		b.Config = cfg.IRC[name]
		b.Bridger = birc.New(cfg.IRC[name], bridge.Account, c)
	case "gitter":
		b.Config = cfg.Gitter[name]
		b.Bridger = bgitter.New(cfg.Gitter[name], bridge.Account, c)
	case "slack":
		b.Config = cfg.Slack[name]
		b.Bridger = bslack.New(cfg.Slack[name], bridge.Account, c)
	case "xmpp":
		b.Config = cfg.Xmpp[name]
		b.Bridger = bxmpp.New(cfg.Xmpp[name], bridge.Account, c)
	case "discord":
		b.Config = cfg.Discord[name]
		b.Bridger = bdiscord.New(cfg.Discord[name], bridge.Account, c)
	case "telegram":
		b.Config = cfg.Telegram[name]
		b.Bridger = btelegram.New(cfg.Telegram[name], bridge.Account, c)
	case "rocketchat":
		b.Config = cfg.Rocketchat[name]
		b.Bridger = brocketchat.New(cfg.Rocketchat[name], bridge.Account, c)
	case "matrix":
		b.Config = cfg.Matrix[name]
		b.Bridger = bmatrix.New(cfg.Matrix[name], bridge.Account, c)
	case "steam":
		b.Config = cfg.Steam[name]
		b.Bridger = bsteam.New(cfg.Steam[name], bridge.Account, c)
	case "api":
		b.Config = cfg.Api[name]
		b.Bridger = api.New(cfg.Api[name], bridge.Account, c)
	}
	return b
}

func (b *Bridge) JoinChannels() error {
	err := b.joinChannels(b.Channels, b.Joined)
	return err
}

func (b *Bridge) joinChannels(channels map[string]config.ChannelInfo, exists map[string]bool) error {
	for ID, channel := range channels {
		if !exists[ID] {
			log.Infof("%s: joining %s (%s)", b.Account, channel.Name, ID)
			err := b.JoinChannel(channel)
			if err != nil {
				return err
			}
			exists[ID] = true
		}
	}
	return nil
}
