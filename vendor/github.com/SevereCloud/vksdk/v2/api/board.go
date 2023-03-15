package api // import "github.com/SevereCloud/vksdk/v2/api"

import (
	"github.com/SevereCloud/vksdk/v2/object"
)

// BoardAddTopic creates a new topic on a community's discussion board.
//
// https://vk.com/dev/board.addTopic
func (vk *VK) BoardAddTopic(params Params) (response int, err error) {
	err = vk.RequestUnmarshal("board.addTopic", &response, params)
	return
}

// BoardCloseTopic closes a topic on a community's discussion board so that comments cannot be posted.
//
// https://vk.com/dev/board.closeTopic
func (vk *VK) BoardCloseTopic(params Params) (response int, err error) {
	err = vk.RequestUnmarshal("board.closeTopic", &response, params)
	return
}

// BoardCreateComment adds a comment on a topic on a community's discussion board.
//
// https://vk.com/dev/board.createComment
func (vk *VK) BoardCreateComment(params Params) (response int, err error) {
	err = vk.RequestUnmarshal("board.createComment", &response, params)
	return
}

// BoardDeleteComment deletes a comment on a topic on a community's discussion board.
//
// https://vk.com/dev/board.deleteComment
func (vk *VK) BoardDeleteComment(params Params) (response int, err error) {
	err = vk.RequestUnmarshal("board.deleteComment", &response, params)
	return
}

// BoardDeleteTopic deletes a topic from a community's discussion board.
//
// https://vk.com/dev/board.deleteTopic
func (vk *VK) BoardDeleteTopic(params Params) (response int, err error) {
	err = vk.RequestUnmarshal("board.deleteTopic", &response, params)
	return
}

// BoardEditComment edits a comment on a topic on a community's discussion board.
//
// https://vk.com/dev/board.editComment
func (vk *VK) BoardEditComment(params Params) (response int, err error) {
	err = vk.RequestUnmarshal("board.editComment", &response, params)
	return
}

// BoardEditTopic edits the title of a topic on a community's discussion board.
//
// https://vk.com/dev/board.editTopic
func (vk *VK) BoardEditTopic(params Params) (response int, err error) {
	err = vk.RequestUnmarshal("board.editTopic", &response, params)
	return
}

// BoardFixTopic pins a topic (fixes its place) to the top of a community's discussion board.
//
// https://vk.com/dev/board.fixTopic
func (vk *VK) BoardFixTopic(params Params) (response int, err error) {
	err = vk.RequestUnmarshal("board.fixTopic", &response, params)
	return
}

// BoardGetCommentsResponse struct.
type BoardGetCommentsResponse struct {
	Count      int                        `json:"count"`
	Items      []object.BoardTopicComment `json:"items"`
	Poll       object.BoardTopicPoll      `json:"poll"`
	RealOffset int                        `json:"real_offset"`
}

// BoardGetComments returns a list of comments on a topic on a community's discussion board.
//
//	extended=0
//
// https://vk.com/dev/board.getComments
func (vk *VK) BoardGetComments(params Params) (response BoardGetCommentsResponse, err error) {
	err = vk.RequestUnmarshal("board.getComments", &response, params, Params{"extended": false})

	return
}

// BoardGetCommentsExtendedResponse struct.
type BoardGetCommentsExtendedResponse struct {
	Count      int                        `json:"count"`
	Items      []object.BoardTopicComment `json:"items"`
	Poll       object.BoardTopicPoll      `json:"poll"`
	RealOffset int                        `json:"real_offset"`
	Profiles   []object.UsersUser         `json:"profiles"`
	Groups     []object.GroupsGroup       `json:"groups"`
}

// BoardGetCommentsExtended returns a list of comments on a topic on a community's discussion board.
//
//	extended=1
//
// https://vk.com/dev/board.getComments
func (vk *VK) BoardGetCommentsExtended(params Params) (response BoardGetCommentsExtendedResponse, err error) {
	err = vk.RequestUnmarshal("board.getComments", &response, params, Params{"extended": true})

	return
}

// BoardGetTopicsResponse struct.
type BoardGetTopicsResponse struct {
	Count        int                 `json:"count"`
	Items        []object.BoardTopic `json:"items"`
	DefaultOrder float64             `json:"default_order"` // BUG(VK): default_order int https://vk.com/bug136682
	CanAddTopics object.BaseBoolInt  `json:"can_add_topics"`
}

// BoardGetTopics returns a list of topics on a community's discussion board.
//
//	extended=0
//
// https://vk.com/dev/board.getTopics
func (vk *VK) BoardGetTopics(params Params) (response BoardGetTopicsResponse, err error) {
	err = vk.RequestUnmarshal("board.getTopics", &response, params, Params{"extended": false})

	return
}

// BoardGetTopicsExtendedResponse struct.
type BoardGetTopicsExtendedResponse struct {
	Count        int                  `json:"count"`
	Items        []object.BoardTopic  `json:"items"`
	DefaultOrder float64              `json:"default_order"` // BUG(VK): default_order int https://vk.com/bug136682
	CanAddTopics object.BaseBoolInt   `json:"can_add_topics"`
	Profiles     []object.UsersUser   `json:"profiles"`
	Groups       []object.GroupsGroup `json:"groups"`
}

// BoardGetTopicsExtended returns a list of topics on a community's discussion board.
//
//	extended=1
//
// https://vk.com/dev/board.getTopics
func (vk *VK) BoardGetTopicsExtended(params Params) (response BoardGetTopicsExtendedResponse, err error) {
	err = vk.RequestUnmarshal("board.getTopics", &response, params, Params{"extended": true})

	return
}

// BoardOpenTopic re-opens a previously closed topic on a community's discussion board.
//
// https://vk.com/dev/board.openTopic
func (vk *VK) BoardOpenTopic(params Params) (response int, err error) {
	err = vk.RequestUnmarshal("board.openTopic", &response, params)
	return
}

// BoardRestoreComment restores a comment deleted from a topic on a community's discussion board.
//
// https://vk.com/dev/board.restoreComment
func (vk *VK) BoardRestoreComment(params Params) (response int, err error) {
	err = vk.RequestUnmarshal("board.restoreComment", &response, params)
	return
}

// BoardUnfixTopic unpins a pinned topic from the top of a community's discussion board.
//
// https://vk.com/dev/board.unfixTopic
func (vk *VK) BoardUnfixTopic(params Params) (response int, err error) {
	err = vk.RequestUnmarshal("board.unfixTopic", &response, params)
	return
}
