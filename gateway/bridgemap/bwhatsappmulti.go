// +build whatsappmulti

package bridgemap

import (
	bwhatsapp "github.com/42wim/matterbridge/bridge/whatsappmulti"
)

func init() {
	FullMap["whatsapp"] = bwhatsapp.New
}
