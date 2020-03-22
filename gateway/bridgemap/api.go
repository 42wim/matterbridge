// +build !noapi

package bridgemap

import (
	"github.com/42wim/matterbridge/bridge/api"
)

func init() {
	FullMap["api"] = api.New
}
