package api // import "github.com/SevereCloud/vksdk/v2/api"

import (
	"github.com/SevereCloud/vksdk/v2/object"
)

// DownloadedGamesGetPaidStatusResponse struct.
type DownloadedGamesGetPaidStatusResponse struct {
	IsPaid object.BaseBoolInt `json:"is_paid"`
}

// DownloadedGamesGetPaidStatus method.
//
// https://vk.com/dev/downloadedGames.getPaidStatus
func (vk *VK) DownloadedGamesGetPaidStatus(params Params) (response DownloadedGamesGetPaidStatusResponse, err error) {
	err = vk.RequestUnmarshal("downloadedGames.getPaidStatus", &response, params, Params{"extended": false})

	return
}
