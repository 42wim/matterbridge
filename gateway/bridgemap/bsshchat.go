// +build !nosshchat

package bridgemap

import (
	bsshchat "github.com/42wim/matterbridge/bridge/sshchat"
)

func init() {
	FullMap["sshchat"] = bsshchat.New
}
