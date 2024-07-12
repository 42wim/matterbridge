// +build !nosoulseek

package bridgemap

import (
	bsoulseek "github.com/42wim/matterbridge/bridge/soulseek"
)

func init() {
	FullMap["soulseek"] = bsoulseek.New
}
