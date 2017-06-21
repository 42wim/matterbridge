package netutil

import (
	"net/http"
	"net/url"
	"strings"
)

// Version of http.Client.PostForm that returns a new request instead of executing it directly.
func NewPostForm(url string, data url.Values) *http.Request {
	req, err := http.NewRequest("POST", url, strings.NewReader(data.Encode()))
	if err != nil {
		panic(err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return req
}
