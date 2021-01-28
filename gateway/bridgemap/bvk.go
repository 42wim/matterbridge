// +build !novk

package bridgemap

import (
	bvk "github.com/42wim/matterbridge/bridge/vk"
)

func init() {
	FullMap["vk"] = bvk.New
}
