// +build !norocketchat

package bridgemap

import (
	brocketchat "github.com/42wim/matterbridge/bridge/rocketchat"
)

func init() {
	FullMap["rocketchat"] = brocketchat.New
}
