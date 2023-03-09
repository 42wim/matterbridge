package api // import "github.com/SevereCloud/vksdk/v2/api"

import (
	"github.com/SevereCloud/vksdk/v2/object"
)

// StoriesBanOwner allows to hide stories from chosen sources from current user's feed.
//
// https://vk.com/dev/stories.banOwner
func (vk *VK) StoriesBanOwner(params Params) (response int, err error) {
	err = vk.RequestUnmarshal("stories.banOwner", &response, params)
	return
}

// StoriesDelete allows to delete story.
//
// https://vk.com/dev/stories.delete
func (vk *VK) StoriesDelete(params Params) (response int, err error) {
	err = vk.RequestUnmarshal("stories.delete", &response, params)
	return
}

// StoriesGetResponse struct.
type StoriesGetResponse struct {
	Count            int                      `json:"count"`
	Items            []object.StoriesFeedItem `json:"items"`
	PromoData        object.StoriesPromoData  `json:"promo_data"`
	NeedUploadScreen object.BaseBoolInt       `json:"need_upload_screen"`
}

// StoriesGet returns stories available for current user.
//
//	extended=0
//
// https://vk.com/dev/stories.get
func (vk *VK) StoriesGet(params Params) (response StoriesGetResponse, err error) {
	err = vk.RequestUnmarshal("stories.get", &response, params, Params{"extended": false})

	return
}

// StoriesGetExtendedResponse struct.
type StoriesGetExtendedResponse struct {
	Count            int                      `json:"count"`
	Items            []object.StoriesFeedItem `json:"items"`
	PromoData        object.StoriesPromoData  `json:"promo_data"`
	NeedUploadScreen object.BaseBoolInt       `json:"need_upload_screen"`
	object.ExtendedResponse
}

// StoriesGetExtended returns stories available for current user.
//
//	extended=1
//
// https://vk.com/dev/stories.get
func (vk *VK) StoriesGetExtended(params Params) (response StoriesGetExtendedResponse, err error) {
	err = vk.RequestUnmarshal("stories.get", &response, params, Params{"extended": true})

	return
}

// StoriesGetBannedResponse struct.
type StoriesGetBannedResponse struct {
	Count int   `json:"count"`
	Items []int `json:"items"`
}

// StoriesGetBanned returns list of sources hidden from current user's feed.
//
//	extended=0
//
// https://vk.com/dev/stories.getBanned
func (vk *VK) StoriesGetBanned(params Params) (response StoriesGetBannedResponse, err error) {
	err = vk.RequestUnmarshal("stories.getBanned", &response, params, Params{"extended": false})

	return
}

// StoriesGetBannedExtendedResponse struct.
type StoriesGetBannedExtendedResponse struct {
	Count int   `json:"count"`
	Items []int `json:"items"`
	object.ExtendedResponse
}

// StoriesGetBannedExtended returns list of sources hidden from current user's feed.
//
//	extended=1
//
// https://vk.com/dev/stories.getBanned
func (vk *VK) StoriesGetBannedExtended(params Params) (response StoriesGetBannedExtendedResponse, err error) {
	err = vk.RequestUnmarshal("stories.getBanned", &response, params, Params{"extended": true})

	return
}

// StoriesGetByIDResponse struct.
type StoriesGetByIDResponse struct {
	Count int                   `json:"count"`
	Items []object.StoriesStory `json:"items"`
}

// StoriesGetByID returns story by its ID.
//
//	extended=0
//
// https://vk.com/dev/stories.getById
func (vk *VK) StoriesGetByID(params Params) (response StoriesGetByIDResponse, err error) {
	err = vk.RequestUnmarshal("stories.getById", &response, params, Params{"extended": false})

	return
}

// StoriesGetByIDExtendedResponse struct.
type StoriesGetByIDExtendedResponse struct {
	Count int                   `json:"count"`
	Items []object.StoriesStory `json:"items"`
	object.ExtendedResponse
}

// StoriesGetByIDExtended returns story by its ID.
//
//	extended=1
//
// https://vk.com/dev/stories.getById
func (vk *VK) StoriesGetByIDExtended(params Params) (response StoriesGetByIDExtendedResponse, err error) {
	err = vk.RequestUnmarshal("stories.getById", &response, params, Params{"extended": true})

	return
}

// StoriesGetPhotoUploadServerResponse struct.
type StoriesGetPhotoUploadServerResponse struct {
	UploadURL string `json:"upload_url"`
	PeerIDs   []int  `json:"peer_ids"`
	UserIDs   []int  `json:"user_ids"`
}

// StoriesGetPhotoUploadServer returns URL for uploading a story with photo.
//
// https://vk.com/dev/stories.getPhotoUploadServer
func (vk *VK) StoriesGetPhotoUploadServer(params Params) (response StoriesGetPhotoUploadServerResponse, err error) {
	err = vk.RequestUnmarshal("stories.getPhotoUploadServer", &response, params)
	return
}

// StoriesGetRepliesResponse struct.
type StoriesGetRepliesResponse struct {
	Count int                      `json:"count"`
	Items []object.StoriesFeedItem `json:"items"`
}

// StoriesGetReplies returns replies to the story.
//
//	extended=0
//
// https://vk.com/dev/stories.getReplies
func (vk *VK) StoriesGetReplies(params Params) (response StoriesGetRepliesResponse, err error) {
	err = vk.RequestUnmarshal("stories.getReplies", &response, params, Params{"extended": false})

	return
}

// StoriesGetRepliesExtendedResponse struct.
type StoriesGetRepliesExtendedResponse struct {
	Count int                      `json:"count"`
	Items []object.StoriesFeedItem `json:"items"`
	object.ExtendedResponse
}

// StoriesGetRepliesExtended returns replies to the story.
//
//	extended=1
//
// https://vk.com/dev/stories.getReplies
func (vk *VK) StoriesGetRepliesExtended(params Params) (response StoriesGetRepliesExtendedResponse, err error) {
	err = vk.RequestUnmarshal("stories.getReplies", &response, params, Params{"extended": true})

	return
}

// StoriesGetStatsResponse struct.
type StoriesGetStatsResponse object.StoriesStoryStats

// StoriesGetStats return statistics data for the story.
//
// https://vk.com/dev/stories.getStats
func (vk *VK) StoriesGetStats(params Params) (response StoriesGetStatsResponse, err error) {
	err = vk.RequestUnmarshal("stories.getStats", &response, params)
	return
}

// StoriesGetVideoUploadServerResponse struct.
type StoriesGetVideoUploadServerResponse struct {
	UploadURL string `json:"upload_url"`
	PeerIDs   []int  `json:"peer_ids"`
	UserIDs   []int  `json:"user_ids"`
}

// StoriesGetVideoUploadServer allows to receive URL for uploading story with video.
//
// https://vk.com/dev/stories.getVideoUploadServer
func (vk *VK) StoriesGetVideoUploadServer(params Params) (response StoriesGetVideoUploadServerResponse, err error) {
	err = vk.RequestUnmarshal("stories.getVideoUploadServer", &response, params)
	return
}

// StoriesGetViewersResponse struct.
type StoriesGetViewersResponse struct {
	Count int                    `json:"count"`
	Items []object.StoriesViewer `json:"items"`
}

// StoriesGetViewers returns a list of story viewers.
//
//	extended=0
//
// https://vk.com/dev/stories.getViewers
func (vk *VK) StoriesGetViewers(params Params) (response StoriesGetViewersResponse, err error) {
	err = vk.RequestUnmarshal("stories.getViewers", &response, params)

	return
}

// StoriesHideAllReplies hides all replies in the last 24 hours from the user to current user's stories.
//
// https://vk.com/dev/stories.hideAllReplies
func (vk *VK) StoriesHideAllReplies(params Params) (response int, err error) {
	err = vk.RequestUnmarshal("stories.hideAllReplies", &response, params)
	return
}

// StoriesHideReply hides the reply to the current user's story.
//
// https://vk.com/dev/stories.hideReply
func (vk *VK) StoriesHideReply(params Params) (response int, err error) {
	err = vk.RequestUnmarshal("stories.hideReply", &response, params)
	return
}

// StoriesSaveResponse struct.
type StoriesSaveResponse struct {
	Count int                   `json:"count"`
	Items []object.StoriesStory `json:"items"`
	object.ExtendedResponse
}

// StoriesSave method.
//
// https://vk.com/dev/stories.save
func (vk *VK) StoriesSave(params Params) (response StoriesSaveResponse, err error) {
	err = vk.RequestUnmarshal("stories.save", &response, params)
	return
}

// StoriesSearchResponse struct.
type StoriesSearchResponse struct {
	Count int                      `json:"count"`
	Items []object.StoriesFeedItem `json:"items"`
}

// StoriesSearch returns search results for stories.
//
//	extended=0
//
// https://vk.com/dev/stories.search
func (vk *VK) StoriesSearch(params Params) (response StoriesSearchResponse, err error) {
	err = vk.RequestUnmarshal("stories.search", &response, params, Params{"extended": false})

	return
}

// StoriesSearchExtendedResponse struct.
type StoriesSearchExtendedResponse struct {
	Count int                      `json:"count"`
	Items []object.StoriesFeedItem `json:"items"`
	object.ExtendedResponse
}

// StoriesSearchExtended returns search results for stories.
//
//	extended=1
//
// https://vk.com/dev/stories.search
func (vk *VK) StoriesSearchExtended(params Params) (response StoriesSearchExtendedResponse, err error) {
	err = vk.RequestUnmarshal("stories.search", &response, params, Params{"extended": true})

	return
}

// StoriesSendInteraction sends feedback to the story.
//
// Available for applications with type VK Mini Apps. The default method is
// not available to applications.
//
// https://vk.com/dev/stories.sendInteraction
func (vk *VK) StoriesSendInteraction(params Params) (response int, err error) {
	err = vk.RequestUnmarshal("stories.sendInteraction", &response, params)
	return
}

// StoriesUnbanOwner allows to show stories from hidden sources in current user's feed.
//
// https://vk.com/dev/stories.unbanOwner
func (vk *VK) StoriesUnbanOwner(params Params) (response int, err error) {
	err = vk.RequestUnmarshal("stories.unbanOwner", &response, params)
	return
}
