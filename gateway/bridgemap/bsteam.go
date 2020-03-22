// +build !nosteam

package bridgemap

import (
	bsteam "github.com/42wim/matterbridge/bridge/steam"
)

func init() {
	FullMap["steam"] = bsteam.New
}
