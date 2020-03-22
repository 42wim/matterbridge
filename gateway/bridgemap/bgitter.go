// +build !nogitter

package bridgemap

import (
	bgitter "github.com/42wim/matterbridge/bridge/gitter"
)

func init() {
	FullMap["gitter"] = bgitter.New
}
