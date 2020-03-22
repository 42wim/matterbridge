// +build !nomattermost

package bridgemap

import (
	bmattermost "github.com/42wim/matterbridge/bridge/mattermost"
)

func init() {
	FullMap["mattermost"] = bmattermost.New
}
