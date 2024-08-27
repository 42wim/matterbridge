package api // import "github.com/SevereCloud/vksdk/v2/api"

import (
	"github.com/SevereCloud/vksdk/v2/object"
)

// FaveAddArticle adds a link to user faves.
//
// https://dev.vk.com/method/fave.addArticle
func (vk *VK) FaveAddArticle(params Params) (response int, err error) {
	err = vk.RequestUnmarshal("fave.addArticle", &response, params)
	return
}

// FaveAddLink adds a link to user faves.
//
// https://dev.vk.com/method/fave.addLink
func (vk *VK) FaveAddLink(params Params) (response int, err error) {
	err = vk.RequestUnmarshal("fave.addLink", &response, params)
	return
}

// FaveAddPage method.
//
// https://dev.vk.com/method/fave.addPage
func (vk *VK) FaveAddPage(params Params) (response int, err error) {
	err = vk.RequestUnmarshal("fave.addPage", &response, params)
	return
}

// FaveAddPost method.
//
// https://dev.vk.com/method/fave.addPost
func (vk *VK) FaveAddPost(params Params) (response int, err error) {
	err = vk.RequestUnmarshal("fave.addPost", &response, params)
	return
}

// FaveAddProduct method.
//
// https://dev.vk.com/method/fave.addProduct
func (vk *VK) FaveAddProduct(params Params) (response int, err error) {
	err = vk.RequestUnmarshal("fave.addProduct", &response, params)
	return
}

// FaveAddTagResponse struct.
type FaveAddTagResponse object.FaveTag

// FaveAddTag method.
//
// https://dev.vk.com/method/fave.addTag
func (vk *VK) FaveAddTag(params Params) (response FaveAddTagResponse, err error) {
	err = vk.RequestUnmarshal("fave.addTag", &response, params)
	return
}

// FaveAddVideo method.
//
// https://dev.vk.com/method/fave.addVideo
func (vk *VK) FaveAddVideo(params Params) (response int, err error) {
	err = vk.RequestUnmarshal("fave.addVideo", &response, params)
	return
}

// FaveEditTag method.
//
// https://dev.vk.com/method/fave.editTag
func (vk *VK) FaveEditTag(params Params) (response int, err error) {
	err = vk.RequestUnmarshal("fave.editTag", &response, params)
	return
}

// FaveGetResponse struct.
type FaveGetResponse struct {
	Count int               `json:"count"`
	Items []object.FaveItem `json:"items"`
}

// FaveGet method.
//
//	extended=0
//
// https://dev.vk.com/method/fave.get
func (vk *VK) FaveGet(params Params) (response FaveGetResponse, err error) {
	err = vk.RequestUnmarshal("fave.get", &response, params, Params{"extended": false})

	return
}

// FaveGetExtendedResponse struct.
type FaveGetExtendedResponse struct {
	Count int               `json:"count"`
	Items []object.FaveItem `json:"items"`
	object.ExtendedResponse
}

// FaveGetExtended method.
//
//	extended=1
//
// https://dev.vk.com/method/fave.get
func (vk *VK) FaveGetExtended(params Params) (response FaveGetExtendedResponse, err error) {
	err = vk.RequestUnmarshal("fave.get", &response, params, Params{"extended": true})

	return
}

// FaveGetPagesResponse struct.
type FaveGetPagesResponse struct {
	Count int               `json:"count"`
	Items []object.FavePage `json:"items"`
}

// FaveGetPages method.
//
// https://dev.vk.com/method/fave.getPages
func (vk *VK) FaveGetPages(params Params) (response FaveGetPagesResponse, err error) {
	err = vk.RequestUnmarshal("fave.getPages", &response, params)
	return
}

// FaveGetTagsResponse struct.
type FaveGetTagsResponse struct {
	Count int              `json:"count"`
	Items []object.FaveTag `json:"items"`
}

// FaveGetTags method.
//
// https://dev.vk.com/method/fave.getTags
func (vk *VK) FaveGetTags(params Params) (response FaveGetTagsResponse, err error) {
	err = vk.RequestUnmarshal("fave.getTags", &response, params)
	return
}

// FaveMarkSeen method.
//
// https://dev.vk.com/method/fave.markSeen
func (vk *VK) FaveMarkSeen(params Params) (response int, err error) {
	err = vk.RequestUnmarshal("fave.markSeen", &response, params)
	return
}

// FaveRemoveArticle method.
//
// https://dev.vk.com/method/fave.removeArticle
func (vk *VK) FaveRemoveArticle(params Params) (response int, err error) {
	err = vk.RequestUnmarshal("fave.removeArticle", &response, params)
	return
}

// FaveRemoveLink removes link from the user's faves.
//
// https://dev.vk.com/method/fave.removeLink
func (vk *VK) FaveRemoveLink(params Params) (response int, err error) {
	err = vk.RequestUnmarshal("fave.removeLink", &response, params)
	return
}

// FaveRemovePage method.
//
// https://dev.vk.com/method/fave.removePage
func (vk *VK) FaveRemovePage(params Params) (response int, err error) {
	err = vk.RequestUnmarshal("fave.removePage", &response, params)
	return
}

// FaveRemovePost method.
//
// https://dev.vk.com/method/fave.removePost
func (vk *VK) FaveRemovePost(params Params) (response int, err error) {
	err = vk.RequestUnmarshal("fave.removePost", &response, params)
	return
}

// FaveRemoveProduct method.
//
// https://dev.vk.com/method/fave.removeProduct
func (vk *VK) FaveRemoveProduct(params Params) (response int, err error) {
	err = vk.RequestUnmarshal("fave.removeProduct", &response, params)
	return
}

// FaveRemoveTag method.
//
// https://dev.vk.com/method/fave.removeTag
func (vk *VK) FaveRemoveTag(params Params) (response int, err error) {
	err = vk.RequestUnmarshal("fave.removeTag", &response, params)
	return
}

// FaveRemoveVideo method.
//
// https://dev.vk.com/method/fave.removeVideo
func (vk *VK) FaveRemoveVideo(params Params) (response int, err error) {
	err = vk.RequestUnmarshal("fave.removeVideo", &response, params)
	return
}

// FaveReorderTags method.
//
// https://dev.vk.com/method/fave.reorderTags
func (vk *VK) FaveReorderTags(params Params) (response int, err error) {
	err = vk.RequestUnmarshal("fave.reorderTags", &response, params)
	return
}

// FaveSetPageTags method.
//
// https://dev.vk.com/method/fave.setPageTags
func (vk *VK) FaveSetPageTags(params Params) (response int, err error) {
	err = vk.RequestUnmarshal("fave.setPageTags", &response, params)
	return
}

// FaveSetTags method.
//
// https://dev.vk.com/method/fave.setTags
func (vk *VK) FaveSetTags(params Params) (response int, err error) {
	err = vk.RequestUnmarshal("fave.setTags", &response, params)
	return
}

// FaveTrackPageInteraction method.
//
// https://dev.vk.com/method/fave.trackPageInteraction
func (vk *VK) FaveTrackPageInteraction(params Params) (response int, err error) {
	err = vk.RequestUnmarshal("fave.trackPageInteraction", &response, params)
	return
}
