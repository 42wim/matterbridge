// +build !nowhatsapp
// +build !whatsappmulti

package bridgemap

import (
	bwhatsapp "github.com/42wim/matterbridge/bridge/whatsapp"
)

func init() {
	FullMap["whatsapp"] = bwhatsapp.New
}
