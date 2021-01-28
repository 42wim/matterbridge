package api

import "github.com/SevereCloud/vksdk/v2/object"

// DonutGetFriendsResponse struct.
type DonutGetFriendsResponse struct {
	Count int                `json:"count"`
	Items []object.UsersUser `json:"items"`
}

// DonutGetFriends method.
//
// https://vk.com/dev/donut.getFriends
func (vk *VK) DonutGetFriends(params Params) (response DonutGetFriendsResponse, err error) {
	err = vk.RequestUnmarshal("donut.getFriends", &response, params)
	return
}

// DonutGetSubscription method.
//
// https://vk.com/dev/donut.getSubscription
func (vk *VK) DonutGetSubscription(params Params) (response object.DonutDonatorSubscriptionInfo, err error) {
	err = vk.RequestUnmarshal("donut.getSubscription", &response, params)
	return
}

// DonutGetSubscriptionsResponse struct.
type DonutGetSubscriptionsResponse struct {
	Subscriptions []object.DonutDonatorSubscriptionInfo `json:"subscriptions"`
	Count         int                                   `json:"count"`
	Profiles      []object.UsersUser                    `json:"profiles"`
	Groups        []object.GroupsGroup                  `json:"groups"`
}

// DonutGetSubscriptions method.
//
// https://vk.com/dev/donut.getSubscriptions
func (vk *VK) DonutGetSubscriptions(params Params) (response DonutGetSubscriptionsResponse, err error) {
	err = vk.RequestUnmarshal("donut.getSubscriptions", &response, params)
	return
}

// DonutIsDon method.
//
// https://vk.com/dev/donut.isDon
func (vk *VK) DonutIsDon(params Params) (response int, err error) {
	err = vk.RequestUnmarshal("donut.isDon", &response, params)
	return
}
