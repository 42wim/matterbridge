package api // import "github.com/SevereCloud/vksdk/v2/api"

import (
	"github.com/SevereCloud/vksdk/v2/object"
)

// WidgetsGetCommentsResponse struct.
type WidgetsGetCommentsResponse struct {
	Count int                           `json:"count"`
	Posts []object.WidgetsWidgetComment `json:"posts"`
}

// WidgetsGetComments gets a list of comments for the page added through the Comments widget.
//
// https://dev.vk.com/method/widgets.getComments
func (vk *VK) WidgetsGetComments(params Params) (response WidgetsGetCommentsResponse, err error) {
	err = vk.RequestUnmarshal("widgets.getComments", &response, params)
	return
}

// WidgetsGetPagesResponse struct.
type WidgetsGetPagesResponse struct {
	Count int                        `json:"count"`
	Pages []object.WidgetsWidgetPage `json:"pages"`
}

// WidgetsGetPages gets a list of application/site pages where the Comments widget or Like widget is installed.
//
// https://dev.vk.com/method/widgets.getPages
func (vk *VK) WidgetsGetPages(params Params) (response WidgetsGetPagesResponse, err error) {
	err = vk.RequestUnmarshal("widgets.getPages", &response, params)
	return
}
