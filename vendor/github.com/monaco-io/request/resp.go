package request

// StatusCode get response status code
func (s *SugaredResp) StatusCode() (code int) {
	return s.resp.StatusCode
}

// Status get response status code and text, like 200 ok
func (s *SugaredResp) Status() (status string) {
	return s.resp.Status
}

// Close close response body
func (s *SugaredResp) Close() {
	if s.resp != nil {
		_ = s.resp.Body.Close()
	}
}
