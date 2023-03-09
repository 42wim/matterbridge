package api // import "github.com/SevereCloud/vksdk/v2/api"

import (
	"github.com/SevereCloud/vksdk/v2/object"
)

// MarketAddResponse struct.
type MarketAddResponse struct {
	MarketItemID int `json:"market_item_id"` // Item ID
}

// MarketAdd adds a new item to the market.
//
// https://vk.com/dev/market.add
func (vk *VK) MarketAdd(params Params) (response MarketAddResponse, err error) {
	err = vk.RequestUnmarshal("market.add", &response, params)
	return
}

// MarketAddAlbumResponse struct.
type MarketAddAlbumResponse struct {
	MarketAlbumID int `json:"market_album_id"` // Album ID
	AlbumsCount   int `json:"albums_count"`
}

// MarketAddAlbum creates new collection of items.
//
// https://vk.com/dev/market.addAlbum
func (vk *VK) MarketAddAlbum(params Params) (response MarketAddAlbumResponse, err error) {
	err = vk.RequestUnmarshal("market.addAlbum", &response, params)
	return
}

// MarketAddToAlbum adds an item to one or multiple collections.
//
// https://vk.com/dev/market.addToAlbum
func (vk *VK) MarketAddToAlbum(params Params) (response int, err error) {
	err = vk.RequestUnmarshal("market.addToAlbum", &response, params)
	return
}

// MarketCreateComment creates a new comment for an item.
//
// https://vk.com/dev/market.createComment
func (vk *VK) MarketCreateComment(params Params) (response int, err error) {
	err = vk.RequestUnmarshal("market.createComment", &response, params)
	return
}

// MarketDelete deletes an item.
//
// https://vk.com/dev/market.delete
func (vk *VK) MarketDelete(params Params) (response int, err error) {
	err = vk.RequestUnmarshal("market.delete", &response, params)
	return
}

// MarketDeleteAlbum deletes a collection of items.
//
// https://vk.com/dev/market.deleteAlbum
func (vk *VK) MarketDeleteAlbum(params Params) (response int, err error) {
	err = vk.RequestUnmarshal("market.deleteAlbum", &response, params)
	return
}

// MarketDeleteComment deletes an item's comment.
//
// https://vk.com/dev/market.deleteComment
func (vk *VK) MarketDeleteComment(params Params) (response int, err error) {
	err = vk.RequestUnmarshal("market.deleteComment", &response, params)
	return
}

// MarketEdit edits an item.
//
// https://vk.com/dev/market.edit
func (vk *VK) MarketEdit(params Params) (response int, err error) {
	err = vk.RequestUnmarshal("market.edit", &response, params)
	return
}

// MarketEditAlbum edits a collection of items.
//
// https://vk.com/dev/market.editAlbum
func (vk *VK) MarketEditAlbum(params Params) (response int, err error) {
	err = vk.RequestUnmarshal("market.editAlbum", &response, params)
	return
}

// MarketEditComment changes item comment's text.
//
// https://vk.com/dev/market.editComment
func (vk *VK) MarketEditComment(params Params) (response int, err error) {
	err = vk.RequestUnmarshal("market.editComment", &response, params)
	return
}

// MarketEditOrder edits an order.
//
// https://vk.com/dev/market.editOrder
func (vk *VK) MarketEditOrder(params Params) (response int, err error) {
	err = vk.RequestUnmarshal("market.editOrder", &response, params)
	return
}

// MarketGetResponse struct.
type MarketGetResponse struct {
	Count int                       `json:"count"`
	Items []object.MarketMarketItem `json:"items"`
}

// MarketGet returns items list for a community.
//
// https://vk.com/dev/market.get
func (vk *VK) MarketGet(params Params) (response MarketGetResponse, err error) {
	err = vk.RequestUnmarshal("market.get", &response, params)
	return
}

// MarketGetAlbumByIDResponse struct.
type MarketGetAlbumByIDResponse struct {
	Count int                        `json:"count"`
	Items []object.MarketMarketAlbum `json:"items"`
}

// MarketGetAlbumByID returns items album's data.
//
// https://vk.com/dev/market.getAlbumById
func (vk *VK) MarketGetAlbumByID(params Params) (response MarketGetAlbumByIDResponse, err error) {
	err = vk.RequestUnmarshal("market.getAlbumById", &response, params)
	return
}

// MarketGetAlbumsResponse struct.
type MarketGetAlbumsResponse struct {
	Count int                        `json:"count"`
	Items []object.MarketMarketAlbum `json:"items"`
}

// MarketGetAlbums returns community's collections list.
//
// https://vk.com/dev/market.getAlbums
func (vk *VK) MarketGetAlbums(params Params) (response MarketGetAlbumsResponse, err error) {
	err = vk.RequestUnmarshal("market.getAlbums", &response, params)
	return
}

// MarketGetByIDResponse struct.
type MarketGetByIDResponse struct {
	Count int                       `json:"count"`
	Items []object.MarketMarketItem `json:"items"`
}

// MarketGetByID returns information about market items by their iDs.
//
// https://vk.com/dev/market.getById
func (vk *VK) MarketGetByID(params Params) (response MarketGetByIDResponse, err error) {
	err = vk.RequestUnmarshal("market.getById", &response, params)
	return
}

// MarketGetCategoriesResponse struct.
type MarketGetCategoriesResponse struct {
	Count int                           `json:"count"`
	Items []object.MarketMarketCategory `json:"items"`
}

// MarketGetCategories returns a list of market categories.
//
// https://vk.com/dev/market.getCategories
func (vk *VK) MarketGetCategories(params Params) (response MarketGetCategoriesResponse, err error) {
	err = vk.RequestUnmarshal("market.getCategories", &response, params)
	return
}

// MarketGetCommentsResponse struct.
type MarketGetCommentsResponse struct {
	Count int                      `json:"count"`
	Items []object.WallWallComment `json:"items"`
}

// MarketGetComments returns comments list for an item.
//
//	extended=0
//
// https://vk.com/dev/market.getComments
func (vk *VK) MarketGetComments(params Params) (response MarketGetCommentsResponse, err error) {
	err = vk.RequestUnmarshal("market.getComments", &response, params, Params{"extended": false})

	return
}

// MarketGetCommentsExtendedResponse struct.
type MarketGetCommentsExtendedResponse struct {
	Count int                      `json:"count"`
	Items []object.WallWallComment `json:"items"`
	object.ExtendedResponse
}

// MarketGetCommentsExtended returns comments list for an item.
//
//	extended=1
//
// https://vk.com/dev/market.getComments
func (vk *VK) MarketGetCommentsExtended(params Params) (response MarketGetCommentsExtendedResponse, err error) {
	err = vk.RequestUnmarshal("market.getComments", &response, params, Params{"extended": true})

	return
}

// MarketGetGroupOrdersResponse struct.
type MarketGetGroupOrdersResponse struct {
	Count int                  `json:"count"`
	Items []object.MarketOrder `json:"items"`
}

// MarketGetGroupOrders returns community's orders list.
//
// https://vk.com/dev/market.getGroupOrders
func (vk *VK) MarketGetGroupOrders(params Params) (response MarketGetGroupOrdersResponse, err error) {
	err = vk.RequestUnmarshal("market.getGroupOrders", &response, params)
	return
}

// MarketGetOrderByIDResponse struct.
type MarketGetOrderByIDResponse struct {
	Order object.MarketOrder `json:"order"`
}

// MarketGetOrderByID returns order by id.
//
// https://vk.com/dev/market.getOrderById
func (vk *VK) MarketGetOrderByID(params Params) (response MarketGetOrderByIDResponse, err error) {
	err = vk.RequestUnmarshal("market.getOrderById", &response, params)
	return
}

// MarketGetOrderItemsResponse struct.
type MarketGetOrderItemsResponse struct {
	Count int                      `json:"count"`
	Items []object.MarketOrderItem `json:"items"`
}

// MarketGetOrderItems returns items of an order.
//
// https://vk.com/dev/market.getOrderItems
func (vk *VK) MarketGetOrderItems(params Params) (response MarketGetOrderItemsResponse, err error) {
	err = vk.RequestUnmarshal("market.getOrderItems", &response, params)
	return
}

// MarketRemoveFromAlbum removes an item from one or multiple collections.
//
// https://vk.com/dev/market.removeFromAlbum
func (vk *VK) MarketRemoveFromAlbum(params Params) (response int, err error) {
	err = vk.RequestUnmarshal("market.removeFromAlbum", &response, params)
	return
}

// MarketReorderAlbums reorders the collections list.
//
// https://vk.com/dev/market.reorderAlbums
func (vk *VK) MarketReorderAlbums(params Params) (response int, err error) {
	err = vk.RequestUnmarshal("market.reorderAlbums", &response, params)
	return
}

// MarketReorderItems changes item place in a collection.
//
// https://vk.com/dev/market.reorderItems
func (vk *VK) MarketReorderItems(params Params) (response int, err error) {
	err = vk.RequestUnmarshal("market.reorderItems", &response, params)
	return
}

// MarketReport sends a complaint to the item.
//
// https://vk.com/dev/market.report
func (vk *VK) MarketReport(params Params) (response int, err error) {
	err = vk.RequestUnmarshal("market.report", &response, params)
	return
}

// MarketReportComment sends a complaint to the item's comment.
//
// https://vk.com/dev/market.reportComment
func (vk *VK) MarketReportComment(params Params) (response int, err error) {
	err = vk.RequestUnmarshal("market.reportComment", &response, params)
	return
}

// MarketRestore restores recently deleted item.
//
// https://vk.com/dev/market.restore
func (vk *VK) MarketRestore(params Params) (response int, err error) {
	err = vk.RequestUnmarshal("market.restore", &response, params)
	return
}

// MarketRestoreComment restores a recently deleted comment.
//
// https://vk.com/dev/market.restoreComment
func (vk *VK) MarketRestoreComment(params Params) (response int, err error) {
	err = vk.RequestUnmarshal("market.restoreComment", &response, params)
	return
}

// MarketSearchResponse struct.
type MarketSearchResponse struct {
	Count    int                       `json:"count"`
	Items    []object.MarketMarketItem `json:"items"`
	ViewType int                       `json:"view_type"`
}

// MarketSearch searches market items in a community's catalog.
//
// https://vk.com/dev/market.search
func (vk *VK) MarketSearch(params Params) (response MarketSearchResponse, err error) {
	err = vk.RequestUnmarshal("market.search", &response, params)
	return
}

// MarketSearchItemsResponse struct.
type MarketSearchItemsResponse struct {
	Count    int                       `json:"count"`
	ViewType int                       `json:"view_type"`
	Items    []object.MarketMarketItem `json:"items"`
	Groups   []object.GroupsGroup      `json:"groups,omitempty"`
}

// MarketSearchItems method.
//
// https://vk.com/dev/market.searchItems
func (vk *VK) MarketSearchItems(params Params) (response MarketSearchItemsResponse, err error) {
	err = vk.RequestUnmarshal("market.searchItems", &response, params)
	return
}
