package bridgemap

import (
	"github.com/42wim/matterbridge/bridge"
	"github.com/42wim/matterbridge/bridge/api"
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
	"github.com/42wim/matterbridge/bridge/zulip"
)

var FullMap = map[string]bridge.Factory{
	"api":          api.New,
	"discord":      bdiscord.New,
	"gitter":       bgitter.New,
	"irc":          birc.New,
	"mattermost":   bmattermost.New,
	"matrix":       bmatrix.New,
	"rocketchat":   brocketchat.New,
	"slack-legacy": bslack.NewLegacy,
	"slack":        bslack.New,
	"sshchat":      bsshchat.New,
	"steam":        bsteam.New,
	"telegram":     btelegram.New,
	"xmpp":         bxmpp.New,
	"zulip":        bzulip.New,
}
