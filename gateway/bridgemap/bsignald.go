// +build !nosignald

package bridgemap

import (
	bsignald "github.com/42wim/matterbridge/bridge/signald"
)

func init() {
	FullMap["signald"] = bsignald.New
}
