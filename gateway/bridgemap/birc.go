// +build !noirc

package bridgemap

import (
	birc "github.com/42wim/matterbridge/bridge/irc"
)

func init() {
	FullMap["irc"] = birc.New
}
