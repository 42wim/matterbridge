package bridgemap

import (
	"github.com/42wim/matterbridge/bridge"
)

var (
	FullMap           = map[string]bridge.Factory{}
	UserTypingSupport = map[string]struct{}{}
)
