package api // import "github.com/SevereCloud/vksdk/v2/api"

import "github.com/SevereCloud/vksdk/v2/object"

// SearchGetHintsResponse struct.
type SearchGetHintsResponse struct {
	Count int                 `json:"count"`
	Items []object.SearchHint `json:"items"`
}

// SearchGetHints allows the programmer to do a quick search for any substring.
//
// https://vk.com/dev/search.getHints
func (vk *VK) SearchGetHints(params Params) (response SearchGetHintsResponse, err error) {
	err = vk.RequestUnmarshal("search.getHints", &response, params)
	return
}
