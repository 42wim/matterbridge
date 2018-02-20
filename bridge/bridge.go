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
	"github.com/42wim/matterbridge/bridge/sshchat"
	"github.com/42wim/matterbridge/bridge/steam"
	"github.com/42wim/matterbridge/bridge/telegram"
	"github.com/42wim/matterbridge/bridge/xmpp"
	log "github.com/sirupsen/logrus"

	"strings"
)

type Bridger interface {
	Send(msg config.Message) (string, error)
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
	bridgeConfig := &config.BridgeConfig{General: &cfg.General, Account: bridge.Account, Remote: c}

	// override config from environment
	config.OverrideCfgFromEnv(cfg, protocol, name)
	switch protocol {
	case "mattermost":
		bridgeConfig.Config = cfg.Mattermost[name]
		b.Bridger = bmattermost.New(bridgeConfig)
	case "irc":
		bridgeConfig.Config = cfg.IRC[name]
		b.Bridger = birc.New(bridgeConfig)
	case "gitter":
		bridgeConfig.Config = cfg.Gitter[name]
		b.Bridger = bgitter.New(bridgeConfig)
	case "slack":
		bridgeConfig.Config = cfg.Slack[name]
		b.Bridger = bslack.New(bridgeConfig)
	case "xmpp":
		bridgeConfig.Config = cfg.Xmpp[name]
		b.Bridger = bxmpp.New(bridgeConfig)
	case "discord":
		bridgeConfig.Config = cfg.Discord[name]
		b.Bridger = bdiscord.New(bridgeConfig)
	case "telegram":
		bridgeConfig.Config = cfg.Telegram[name]
		b.Bridger = btelegram.New(bridgeConfig)
	case "rocketchat":
		bridgeConfig.Config = cfg.Rocketchat[name]
		b.Bridger = brocketchat.New(bridgeConfig)
	case "matrix":
		bridgeConfig.Config = cfg.Matrix[name]
		b.Bridger = bmatrix.New(bridgeConfig)
	case "steam":
		bridgeConfig.Config = cfg.Steam[name]
		b.Bridger = bsteam.New(bridgeConfig)
	case "sshchat":
		bridgeConfig.Config = cfg.Sshchat[name]
		b.Bridger = bsshchat.New(bridgeConfig)
	case "api":
		bridgeConfig.Config = cfg.Api[name]
		b.Bridger = api.New(bridgeConfig)
	}
	b.Config = bridgeConfig.Config
	return b
}

func (b *Bridge) JoinChannels() error {
	err := b.joinChannels(b.Channels, b.Joined)
	return err
}

func (b *Bridge) joinChannels(channels map[string]config.ChannelInfo, exists map[string]bool) error {
	for ID, channel := range channels {
		if !exists[ID] {
			log.Infof("%s: joining %s (ID: %s)", b.Account, channel.Name, ID)
			err := b.JoinChannel(channel)
			if err != nil {
				return err
			}
			exists[ID] = true
		}
	}
	return nil
}
