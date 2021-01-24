package api // import "github.com/SevereCloud/vksdk/v2/api"

import (
	"github.com/SevereCloud/vksdk/v2/object"
)

// OrdersCancelSubscription allows to cancel subscription.
//
// https://vk.com/dev/orders.cancelSubscription
func (vk *VK) OrdersCancelSubscription(params Params) (response int, err error) {
	err = vk.RequestUnmarshal("orders.cancelSubscription", &response, params)
	return
}

// OrdersChangeStateResponse struct.
type OrdersChangeStateResponse string // New state

// OrdersChangeState changes order status.
//
// https://vk.com/dev/orders.changeState
func (vk *VK) OrdersChangeState(params Params) (response OrdersChangeStateResponse, err error) {
	err = vk.RequestUnmarshal("orders.changeState", &response, params)
	return
}

// OrdersGetResponse struct.
type OrdersGetResponse []object.OrdersOrder

// OrdersGet returns a list of orders.
//
// https://vk.com/dev/orders.get
func (vk *VK) OrdersGet(params Params) (response OrdersGetResponse, err error) {
	err = vk.RequestUnmarshal("orders.get", &response, params)
	return
}

// OrdersGetAmountResponse struct.
type OrdersGetAmountResponse []object.OrdersAmount

// OrdersGetAmount returns the cost of votes in the user's consent.
//
// https://vk.com/dev/orders.getAmount
func (vk *VK) OrdersGetAmount(params Params) (response OrdersGetAmountResponse, err error) {
	err = vk.RequestUnmarshal("orders.getAmount", &response, params)
	return
}

// OrdersGetByIDResponse struct.
type OrdersGetByIDResponse []object.OrdersOrder

// OrdersGetByID returns information about orders by their IDs.
//
// https://vk.com/dev/orders.getByID
func (vk *VK) OrdersGetByID(params Params) (response OrdersGetByIDResponse, err error) {
	err = vk.RequestUnmarshal("orders.getById", &response, params)
	return
}

// OrdersGetUserSubscriptionByIDResponse struct.
type OrdersGetUserSubscriptionByIDResponse object.OrdersSubscription

// OrdersGetUserSubscriptionByID allows to get subscription by its ID.
//
// https://vk.com/dev/orders.getUserSubscriptionById
func (vk *VK) OrdersGetUserSubscriptionByID(params Params) (response OrdersGetUserSubscriptionByIDResponse, err error) {
	err = vk.RequestUnmarshal("orders.getUserSubscriptionById", &response, params)
	return
}

// OrdersGetUserSubscriptionsResponse struct.
type OrdersGetUserSubscriptionsResponse struct {
	Count int                         `json:"count"` // Total number
	Items []object.OrdersSubscription `json:"items"`
}

// OrdersGetUserSubscriptions allows to get user's active subscriptions.
//
// https://vk.com/dev/orders.getUserSubscriptions
func (vk *VK) OrdersGetUserSubscriptions(params Params) (response OrdersGetUserSubscriptionsResponse, err error) {
	err = vk.RequestUnmarshal("orders.getUserSubscriptions", &response, params)
	return
}

// OrdersUpdateSubscription allows to update subscription price.
//
// https://vk.com/dev/orders.updateSubscription
func (vk *VK) OrdersUpdateSubscription(params Params) (response int, err error) {
	err = vk.RequestUnmarshal("orders.updateSubscription", &response, params)
	return
}
