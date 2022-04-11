package api // import "github.com/SevereCloud/vksdk/v2/api"

import (
	"github.com/SevereCloud/vksdk/v2/object"
)

// StoreAddStickersToFavorite add stickers to favorite.
//
// https://vk.com/dev/store.addStickersToFavorite
func (vk *VK) StoreAddStickersToFavorite(params Params) (response int, err error) {
	err = vk.RequestUnmarshal("store.addStickersToFavorite", &response, params)
	return
}

// StoreGetFavoriteStickersResponse struct.
type StoreGetFavoriteStickersResponse struct {
	Count int                  `json:"count"`
	Items []object.BaseSticker `json:"items"`
}

// StoreGetFavoriteStickers return favorite stickers.
//
// https://vk.com/dev/store.getFavoriteStickers
func (vk *VK) StoreGetFavoriteStickers(params Params) (response StoreGetFavoriteStickersResponse, err error) {
	err = vk.RequestUnmarshal("store.getFavoriteStickers", &response, params)
	return
}

// StoreRemoveStickersFromFavorite remove stickers from favorite.
//
// https://vk.com/dev/store.removeStickersFromFavorite
func (vk *VK) StoreRemoveStickersFromFavorite(params Params) (response int, err error) {
	err = vk.RequestUnmarshal("store.removeStickersFromFavorite", &response, params)
	return
}
