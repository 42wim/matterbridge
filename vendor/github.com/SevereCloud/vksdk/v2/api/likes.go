package api // import "github.com/SevereCloud/vksdk/v2/api"

import (
	"github.com/SevereCloud/vksdk/v2/object"
)

// LikesAddResponse struct.
type LikesAddResponse struct {
	Likes int `json:"likes"`
}

// LikesAdd adds the specified object to the Likes list of the current user.
//
// https://vk.com/dev/likes.add
func (vk *VK) LikesAdd(params Params) (response LikesAddResponse, err error) {
	err = vk.RequestUnmarshal("likes.add", &response, params)
	return
}

// LikesDeleteResponse struct.
type LikesDeleteResponse struct {
	Likes int `json:"likes"`
}

// LikesDelete deletes the specified object from the Likes list of the current user.
//
// https://vk.com/dev/likes.delete
func (vk *VK) LikesDelete(params Params) (response LikesDeleteResponse, err error) {
	err = vk.RequestUnmarshal("likes.delete", &response, params)
	return
}

// LikesGetListResponse struct.
type LikesGetListResponse struct {
	Count int   `json:"count"`
	Items []int `json:"items"`
}

// LikesGetList likes.getList returns a list of IDs of users who added the specified object to their Likes list.
//
//	extended=0
//
// https://vk.com/dev/likes.getList
func (vk *VK) LikesGetList(params Params) (response LikesGetListResponse, err error) {
	err = vk.RequestUnmarshal("likes.getList", &response, params, Params{"extended": false})

	return
}

// LikesGetListExtendedResponse struct.
type LikesGetListExtendedResponse struct {
	Count int                `json:"count"`
	Items []object.UsersUser `json:"items"`
}

// LikesGetListExtended likes.getList returns a list of IDs of users who added the specified object to their Likes list.
//
//	extended=1
//
// https://vk.com/dev/likes.getList
func (vk *VK) LikesGetListExtended(params Params) (response LikesGetListExtendedResponse, err error) {
	err = vk.RequestUnmarshal("likes.getList", &response, params, Params{"extended": true})

	return
}

// LikesIsLikedResponse struct.
type LikesIsLikedResponse struct {
	Liked  object.BaseBoolInt `json:"liked"`
	Copied object.BaseBoolInt `json:"copied"`
}

// LikesIsLiked checks for the object in the Likes list of the specified user.
//
// https://vk.com/dev/likes.isLiked
func (vk *VK) LikesIsLiked(params Params) (response LikesIsLikedResponse, err error) {
	err = vk.RequestUnmarshal("likes.isLiked", &response, params)
	return
}
