//go:build !noharmony
// +build !noharmony

package bridgemap

import (
	bharmony "github.com/42wim/matterbridge/bridge/harmony"
)

func init() {
	FullMap["harmony"] = bharmony.New
}
