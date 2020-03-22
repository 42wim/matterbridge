// +build !notelegram

package bridgemap

import (
	btelegram "github.com/42wim/matterbridge/bridge/telegram"
)

func init() {
	FullMap["telegram"] = btelegram.New
}
