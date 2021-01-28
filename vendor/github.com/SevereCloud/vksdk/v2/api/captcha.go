package api

// CaptchaForce api method.
func (vk *VK) CaptchaForce(params Params) (response int, err error) {
	err = vk.RequestUnmarshal("captcha.force", &response, params)
	return
}
