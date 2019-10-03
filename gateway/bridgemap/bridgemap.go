package bridgemap

import (
	"github.com/42wim/matterbridge/bridge"
	"github.com/42wim/matterbridge/bridge/api"
	bdiscord "github.com/42wim/matterbridge/bridge/discord"
	bgitter "github.com/42wim/matterbridge/bridge/gitter"
	birc "github.com/42wim/matterbridge/bridge/irc"
	bkeybase "github.com/42wim/matterbridge/bridge/keybase"
	bmatrix "github.com/42wim/matterbridge/bridge/matrix"
	bmattermost "github.com/42wim/matterbridge/bridge/mattermost"
	brocketchat "github.com/42wim/matterbridge/bridge/rocketchat"
	bslack "github.com/42wim/matterbridge/bridge/slack"
	bsshchat "github.com/42wim/matterbridge/bridge/sshchat"
	bsteam "github.com/42wim/matterbridge/bridge/steam"
	btelegram "github.com/42wim/matterbridge/bridge/telegram"
	bwhatsapp "github.com/42wim/matterbridge/bridge/whatsapp"
	bxmpp "github.com/42wim/matterbridge/bridge/xmpp"
	bzulip "github.com/42wim/matterbridge/bridge/zulip"
)

var (
	FullMap = map[string]bridge.Factory{
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
		"whatsapp":     bwhatsapp.New,
		"xmpp":         bxmpp.New,
		"zulip":        bzulip.New,
		"keybase":      bkeybase.New,
	}

	UserTypingSupport = map[string]struct{}{
		"slack":   {},
		"discord": {},
	}
)
