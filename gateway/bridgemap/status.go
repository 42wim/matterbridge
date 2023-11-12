//go:build !nostatus
// +build !nostatus

package bridgemap

import (
	bstatus "github.com/42wim/matterbridge/bridge/status"
)

func init() {
	FullMap["status"] = bstatus.New
}
