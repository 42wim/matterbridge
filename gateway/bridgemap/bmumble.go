// +build !nomumble

package bridgemap

import (
	bmumble "github.com/42wim/matterbridge/bridge/mumble"
)

func init() {
	FullMap["mumble"] = bmumble.New
}
