// +build !noslack

package bridgemap

import (
	bslack "github.com/42wim/matterbridge/bridge/slack"
)

func init() {
	FullMap["slack-legacy"] = bslack.NewLegacy
	FullMap["slack"] = bslack.New
	UserTypingSupport["slack"] = struct{}{}
}
