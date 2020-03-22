// +build !noxmpp

package bridgemap

import (
	bxmpp "github.com/42wim/matterbridge/bridge/xmpp"
)

func init() {
	FullMap["xmpp"] = bxmpp.New
}
