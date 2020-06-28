// +build !nonctalk

package bridgemap

import (
	btalk "github.com/42wim/matterbridge/bridge/nctalk"
)

func init() {
	FullMap["nctalk"] = btalk.New
}
