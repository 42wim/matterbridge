// +build !nomsteams

package bridgemap

import (
	bmsteams "github.com/42wim/matterbridge/bridge/msteams"
)

func init() {
	FullMap["msteams"] = bmsteams.New
}
