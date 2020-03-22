// +build !nodiscord

package bridgemap

import (
	bdiscord "github.com/42wim/matterbridge/bridge/discord"
)

func init() {
	FullMap["discord"] = bdiscord.New
	UserTypingSupport["discord"] = struct{}{}
}
