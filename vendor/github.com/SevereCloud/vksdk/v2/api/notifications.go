package api // import "github.com/SevereCloud/vksdk/v2/api"

import (
	"github.com/SevereCloud/vksdk/v2/object"
)

// NotificationsGetResponse struct.
type NotificationsGetResponse struct {
	Count      int                                `json:"count"`
	Items      []object.NotificationsNotification `json:"items"`
	Profiles   []object.UsersUser                 `json:"profiles"`
	Groups     []object.GroupsGroup               `json:"groups"`
	Photos     []object.PhotosPhoto               `json:"photos"`
	Videos     []object.VideoVideo                `json:"videos"`
	Apps       []object.AppsApp                   `json:"apps"`
	LastViewed int                                `json:"last_viewed"`
	NextFrom   string                             `json:"next_from"`
	TTL        int                                `json:"ttl"`
}

// NotificationsGet returns a list of notifications about other users' feedback to the current user's wall posts.
//
// https://vk.com/dev/notifications.get
func (vk *VK) NotificationsGet(params Params) (response NotificationsGetResponse, err error) {
	err = vk.RequestUnmarshal("notifications.get", &response, params)
	return
}

// NotificationsMarkAsViewed resets the counter of new notifications
// about other users' feedback to the current user's wall posts.
//
// https://vk.com/dev/notifications.markAsViewed
func (vk *VK) NotificationsMarkAsViewed(params Params) (response int, err error) {
	err = vk.RequestUnmarshal("notifications.markAsViewed", &response, params)
	return
}

// NotificationsSendMessageResponse struct.
type NotificationsSendMessageResponse []struct {
	UserID int                `json:"user_id"`
	Status object.BaseBoolInt `json:"status"`
	Error  struct {
		Code        int    `json:"code"`
		Description string `json:"description"`
	} `json:"error"`
}

// NotificationsSendMessage sends notification to the VK Apps user.
//
// https://vk.com/dev/notifications.sendMessage
func (vk *VK) NotificationsSendMessage(params Params) (response NotificationsSendMessageResponse, err error) {
	err = vk.RequestUnmarshal("notifications.sendMessage", &response, params)
	return
}
