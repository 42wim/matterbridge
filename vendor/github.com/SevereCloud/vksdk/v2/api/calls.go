package api // import "github.com/SevereCloud/vksdk/v2/api"

// CallsStartResponse struct.
type CallsStartResponse struct {
	JoinLink string `json:"join_link"`
	CallID   string `json:"call_id"`
}

// CallsStart method.
//
// https://dev.vk.com/method/calls.start
func (vk *VK) CallsStart(params Params) (response CallsStartResponse, err error) {
	err = vk.RequestUnmarshal("calls.start", &response, params)
	return
}

// CallsForceFinish method.
//
// https://dev.vk.com/method/calls.forceFinish
func (vk *VK) CallsForceFinish(params Params) (response int, err error) {
	err = vk.RequestUnmarshal("calls.forceFinish", &response, params)
	return
}
