// +build !nokeybase

package bridgemap

import (
	bkeybase "github.com/42wim/matterbridge/bridge/keybase"
)

func init() {
	FullMap["keybase"] = bkeybase.New
}
