// +build !nozulip

package bridgemap

import (
	bzulip "github.com/42wim/matterbridge/bridge/zulip"
)

func init() {
	FullMap["zulip"] = bzulip.New
}
