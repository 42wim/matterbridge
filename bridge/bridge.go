package bridge

import (
	"github.com/42wim/matterbridge/bridge/config"
	"github.com/42wim/matterbridge/bridge/gitter"
	"github.com/42wim/matterbridge/bridge/irc"
	"github.com/42wim/matterbridge/bridge/mattermost"
	"github.com/42wim/matterbridge/bridge/slack"
	"github.com/42wim/matterbridge/bridge/xmpp"
	"strings"
)

type Bridge interface {
	Send(msg config.Message) error
	Name() string
	Connect() error
	FullOrigin() string
	Origin() string
	Protocol() string
	JoinChannel(channel string) error
}

func New(cfg *config.Config, bridge *config.Bridge, c chan config.Message) Bridge {
	accInfo := strings.Split(bridge.Account, ".")
	protocol := accInfo[0]
	name := accInfo[1]
	switch protocol {
	case "mattermost":
		return bmattermost.New(cfg.Mattermost[name], name, c)
	case "irc":
		return birc.New(cfg.IRC[name], name, c)
	case "gitter":
		return bgitter.New(cfg.Gitter[name], name, c)
	case "slack":
		return bslack.New(cfg.Slack[name], name, c)
	case "xmpp":
		return bxmpp.New(cfg.Xmpp[name], name, c)
	}
	return nil
}
