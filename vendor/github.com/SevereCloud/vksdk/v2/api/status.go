package api // import "github.com/SevereCloud/vksdk/v2/api"

import (
	"github.com/SevereCloud/vksdk/v2/object"
)

// StatusGetResponse struct.
type StatusGetResponse struct {
	Audio object.AudioAudio `json:"audio"`
	Text  string            `json:"text"`
}

// StatusGet returns data required to show the status of a user or community.
func (vk *VK) StatusGet(params Params) (response StatusGetResponse, err error) {
	err = vk.RequestUnmarshal("status.get", &response, params)
	return
}

// StatusSet sets a new status for the current user.
func (vk *VK) StatusSet(params Params) (response int, err error) {
	err = vk.RequestUnmarshal("status.set", &response, params)
	return
}
