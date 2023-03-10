package api // import "github.com/SevereCloud/vksdk/v2/api"

import (
	"github.com/SevereCloud/vksdk/v2/object"
)

// WallCheckCopyrightLink method.
//
// https://vk.com/dev/wall.checkCopyrightLink
func (vk *VK) WallCheckCopyrightLink(params Params) (response int, err error) {
	err = vk.RequestUnmarshal("wall.checkCopyrightLink", &response, params)
	return
}

// WallCloseComments turn off post commenting.
//
// https://vk.com/dev/wall.closeComments
func (vk *VK) WallCloseComments(params Params) (response int, err error) {
	err = vk.RequestUnmarshal("wall.closeComments", &response, params)
	return
}

// WallCreateCommentResponse struct.
type WallCreateCommentResponse struct {
	CommentID    int   `json:"comment_id"`
	ParentsStack []int `json:"parents_stack"`
}

// WallCreateComment adds a comment to a post on a user wall or community wall.
//
// https://vk.com/dev/wall.createComment
func (vk *VK) WallCreateComment(params Params) (response WallCreateCommentResponse, err error) {
	err = vk.RequestUnmarshal("wall.createComment", &response, params)
	return
}

// WallDelete deletes a post from a user wall or community wall.
//
// https://vk.com/dev/wall.delete
func (vk *VK) WallDelete(params Params) (response int, err error) {
	err = vk.RequestUnmarshal("wall.delete", &response, params)
	return
}

// WallDeleteComment deletes a comment on a post on a user wall or community wall.
//
// https://vk.com/dev/wall.deleteComment
func (vk *VK) WallDeleteComment(params Params) (response int, err error) {
	err = vk.RequestUnmarshal("wall.deleteComment", &response, params)
	return
}

// WallEditResponse struct.
type WallEditResponse struct {
	PostID int `json:"post_id"`
}

// WallEdit edits a post on a user wall or community wall.
//
// https://vk.com/dev/wall.edit
func (vk *VK) WallEdit(params Params) (response WallEditResponse, err error) {
	err = vk.RequestUnmarshal("wall.edit", &response, params)
	return
}

// WallEditAdsStealth allows to edit hidden post.
//
// https://vk.com/dev/wall.editAdsStealth
func (vk *VK) WallEditAdsStealth(params Params) (response int, err error) {
	err = vk.RequestUnmarshal("wall.editAdsStealth", &response, params)
	return
}

// WallEditComment edits a comment on a user wall or community wall.
//
// https://vk.com/dev/wall.editComment
func (vk *VK) WallEditComment(params Params) (response int, err error) {
	err = vk.RequestUnmarshal("wall.editComment", &response, params)
	return
}

// WallGetResponse struct.
type WallGetResponse struct {
	Count int                   `json:"count"`
	Items []object.WallWallpost `json:"items"`
}

// WallGet returns a list of posts on a user wall or community wall.
//
//	extended=0
//
// https://vk.com/dev/wall.get
func (vk *VK) WallGet(params Params) (response WallGetResponse, err error) {
	err = vk.RequestUnmarshal("wall.get", &response, params, Params{"extended": false})

	return
}

// WallGetExtendedResponse struct.
type WallGetExtendedResponse struct {
	Count int                   `json:"count"`
	Items []object.WallWallpost `json:"items"`
	object.ExtendedResponse
}

// WallGetExtended returns a list of posts on a user wall or community wall.
//
//	extended=1
//
// https://vk.com/dev/wall.get
func (vk *VK) WallGetExtended(params Params) (response WallGetExtendedResponse, err error) {
	err = vk.RequestUnmarshal("wall.get", &response, params, Params{"extended": true})

	return
}

// WallGetByIDResponse struct.
type WallGetByIDResponse []object.WallWallpost

// WallGetByID returns a list of posts from user or community walls by their IDs.
//
//	extended=0
//
// https://vk.com/dev/wall.getById
func (vk *VK) WallGetByID(params Params) (response WallGetByIDResponse, err error) {
	err = vk.RequestUnmarshal("wall.getById", &response, params, Params{"extended": false})

	return
}

// WallGetByIDExtendedResponse struct.
type WallGetByIDExtendedResponse struct {
	Items []object.WallWallpost `json:"items"`
	object.ExtendedResponse
}

// WallGetByIDExtended returns a list of posts from user or community walls by their IDs.
//
//	extended=1
//
// https://vk.com/dev/wall.getById
func (vk *VK) WallGetByIDExtended(params Params) (response WallGetByIDExtendedResponse, err error) {
	err = vk.RequestUnmarshal("wall.getById", &response, params, Params{"extended": true})

	return
}

// WallGetCommentResponse struct.
type WallGetCommentResponse struct {
	Items             []object.WallWallComment `json:"items"`
	CanPost           object.BaseBoolInt       `json:"can_post"`
	ShowReplyButton   object.BaseBoolInt       `json:"show_reply_button"`
	GroupsCanPost     object.BaseBoolInt       `json:"groups_can_post"`
	CurrentLevelCount int                      `json:"current_level_count"`
}

// WallGetComment allows to obtain wall comment info.
//
//	extended=0
//
// https://vk.com/dev/wall.getComment
func (vk *VK) WallGetComment(params Params) (response WallGetCommentResponse, err error) {
	err = vk.RequestUnmarshal("wall.getComment", &response, params, Params{"extended": false})

	return
}

// WallGetCommentExtendedResponse struct.
type WallGetCommentExtendedResponse struct {
	Count             int                      `json:"count"`
	Items             []object.WallWallComment `json:"items"`
	CanPost           object.BaseBoolInt       `json:"can_post"`
	ShowReplyButton   object.BaseBoolInt       `json:"show_reply_button"`
	GroupsCanPost     object.BaseBoolInt       `json:"groups_can_post"`
	CurrentLevelCount int                      `json:"current_level_count"`
	Profiles          []object.UsersUser       `json:"profiles"`
	Groups            []object.GroupsGroup     `json:"groups"`
}

// WallGetCommentExtended allows to obtain wall comment info.
//
//	extended=1
//
// https://vk.com/dev/wall.getComment
func (vk *VK) WallGetCommentExtended(params Params) (response WallGetCommentExtendedResponse, err error) {
	err = vk.RequestUnmarshal("wall.getComment", &response, params, Params{"extended": true})

	return
}

// WallGetCommentsResponse struct.
type WallGetCommentsResponse struct {
	CanPost           object.BaseBoolInt       `json:"can_post"`
	ShowReplyButton   object.BaseBoolInt       `json:"show_reply_button"`
	GroupsCanPost     object.BaseBoolInt       `json:"groups_can_post"`
	CurrentLevelCount int                      `json:"current_level_count"`
	Count             int                      `json:"count"`
	Items             []object.WallWallComment `json:"items"`
}

// WallGetComments returns a list of comments on a post on a user wall or community wall.
//
//	extended=0
//
// https://vk.com/dev/wall.getComments
func (vk *VK) WallGetComments(params Params) (response WallGetCommentsResponse, err error) {
	err = vk.RequestUnmarshal("wall.getComments", &response, params, Params{"extended": false})

	return
}

// WallGetCommentsExtendedResponse struct.
type WallGetCommentsExtendedResponse struct {
	CanPost           object.BaseBoolInt       `json:"can_post"`
	ShowReplyButton   object.BaseBoolInt       `json:"show_reply_button"`
	GroupsCanPost     object.BaseBoolInt       `json:"groups_can_post"`
	CurrentLevelCount int                      `json:"current_level_count"`
	Count             int                      `json:"count"`
	Items             []object.WallWallComment `json:"items"`
	object.ExtendedResponse
}

// WallGetCommentsExtended returns a list of comments on a post on a user wall or community wall.
//
//	extended=1
//
// https://vk.com/dev/wall.getComments
func (vk *VK) WallGetCommentsExtended(params Params) (response WallGetCommentsExtendedResponse, err error) {
	err = vk.RequestUnmarshal("wall.getComments", &response, params, Params{"extended": true})

	return
}

// WallGetRepostsResponse struct.
type WallGetRepostsResponse struct {
	Items []object.WallWallpost `json:"items"`
	object.ExtendedResponse
}

// WallGetReposts returns information about reposts of a post on user wall or community wall.
//
// https://vk.com/dev/wall.getReposts
func (vk *VK) WallGetReposts(params Params) (response WallGetRepostsResponse, err error) {
	err = vk.RequestUnmarshal("wall.getReposts", &response, params)
	return
}

// WallOpenComments includes posting comments.
//
// https://vk.com/dev/wall.openComments
func (vk *VK) WallOpenComments(params Params) (response int, err error) {
	err = vk.RequestUnmarshal("wall.openComments", &response, params)
	return
}

// WallPin pins the post on wall.
//
// https://vk.com/dev/wall.pin
func (vk *VK) WallPin(params Params) (response int, err error) {
	err = vk.RequestUnmarshal("wall.pin", &response, params)
	return
}

// WallPostResponse struct.
type WallPostResponse struct {
	PostID int `json:"post_id"`
}

// WallPost adds a new post on a user wall or community wall.Can also be used to publish suggested or scheduled posts.
//
// https://vk.com/dev/wall.post
func (vk *VK) WallPost(params Params) (response WallPostResponse, err error) {
	err = vk.RequestUnmarshal("wall.post", &response, params)
	return
}

// WallPostAdsStealthResponse struct.
type WallPostAdsStealthResponse struct {
	PostID int `json:"post_id"`
}

// WallPostAdsStealth allows to create hidden post which will
// not be shown on the community's wall and can be used for creating
// an ad with type "Community post".
//
// https://vk.com/dev/wall.postAdsStealth
func (vk *VK) WallPostAdsStealth(params Params) (response WallPostAdsStealthResponse, err error) {
	err = vk.RequestUnmarshal("wall.postAdsStealth", &response, params)
	return
}

// WallReportComment reports (submits a complaint about) a comment on a post on a user wall or community wall.
//
// https://vk.com/dev/wall.reportComment
func (vk *VK) WallReportComment(params Params) (response int, err error) {
	err = vk.RequestUnmarshal("wall.reportComment", &response, params)
	return
}

// WallReportPost reports (submits a complaint about) a post on a user wall or community wall.
//
// https://vk.com/dev/wall.reportPost
func (vk *VK) WallReportPost(params Params) (response int, err error) {
	err = vk.RequestUnmarshal("wall.reportPost", &response, params)
	return
}

// WallRepostResponse struct.
type WallRepostResponse struct {
	Success         int `json:"success"`
	PostID          int `json:"post_id"`
	RepostsCount    int `json:"reposts_count"`
	LikesCount      int `json:"likes_count"`
	WallRepostCount int `json:"wall_repost_count"`
	MailRepostCount int `json:"mail_repost_count"`
}

// WallRepost reposts ( copies) an object to a user wall or community wall.
//
// https://vk.com/dev/wall.repost
func (vk *VK) WallRepost(params Params) (response WallRepostResponse, err error) {
	err = vk.RequestUnmarshal("wall.repost", &response, params)
	return
}

// WallRestore restores a post deleted from a user wall or community wall.
//
// https://vk.com/dev/wall.restore
func (vk *VK) WallRestore(params Params) (response int, err error) {
	err = vk.RequestUnmarshal("wall.restore", &response, params)
	return
}

// WallRestoreComment restores a comment deleted from a user wall or community wall.
//
// https://vk.com/dev/wall.restoreComment
func (vk *VK) WallRestoreComment(params Params) (response int, err error) {
	err = vk.RequestUnmarshal("wall.restoreComment", &response, params)
	return
}

// WallSearchResponse struct.
type WallSearchResponse struct {
	Count int                   `json:"count"`
	Items []object.WallWallpost `json:"items"`
}

// WallSearch allows to search posts on user or community walls.
//
//	extended=0
//
// https://vk.com/dev/wall.search
func (vk *VK) WallSearch(params Params) (response WallSearchResponse, err error) {
	err = vk.RequestUnmarshal("wall.search", &response, params, Params{"extended": false})

	return
}

// WallSearchExtendedResponse struct.
type WallSearchExtendedResponse struct {
	Count int                   `json:"count"`
	Items []object.WallWallpost `json:"items"`
	object.ExtendedResponse
}

// WallSearchExtended allows to search posts on user or community walls.
//
//	extended=1
//
// https://vk.com/dev/wall.search
func (vk *VK) WallSearchExtended(params Params) (response WallSearchExtendedResponse, err error) {
	err = vk.RequestUnmarshal("wall.search", &response, params, Params{"extended": true})

	return
}

// WallUnpin unpins the post on wall.
//
// https://vk.com/dev/wall.unpin
func (vk *VK) WallUnpin(params Params) (response int, err error) {
	err = vk.RequestUnmarshal("wall.unpin", &response, params)
	return
}
