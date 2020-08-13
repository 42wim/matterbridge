package request

import (
	"io/ioutil"
	"net/http"
)

// Do send http request
func (c *Client) Do() (resp SugaredResp, err error) {
	defer resp.Close()

	if err = c.buildRequest(); err != nil {
		return
	}

	// send request and close on func call end
	if resp.resp, err = c.client.Do(c.req); err != nil {
		return
	}

	// read response data form resp
	resp.Data, err = ioutil.ReadAll(resp.resp.Body)
	resp.Code = resp.resp.StatusCode
	return
}

// Resp do request and get original http response struct
func (c *Client) Resp() (resp *http.Response, err error) {
	if err = c.buildRequest(); err != nil {
		return
	}
	return c.client.Do(c.req)
}
