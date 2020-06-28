package request

import (
	"bytes"
	"crypto/tls"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"time"
)

// Do send http request
func (c *Client) Do() (resp SugaredResp, err error) {
	defer resp.Close()

	if err := c.buildRequest(); err != nil {
		return resp, err
	}

	// send request and close on func call end
	if resp.resp, err = c.client.Do(c.req); err != nil {
		return resp, err
	}

	// read response data form resp
	resp.Data, err = ioutil.ReadAll(resp.resp.Body)
	resp.Code = resp.resp.StatusCode
	return resp, err
}

func (c *Client) buildRequest() (err error) {

	// encode requestURL.httpURL like https://google.com?hello=world&package=request
	ru := requestURL{
		urlString:  c.URL,
		parameters: c.Params,
	}
	if err := ru.EncodeURL(); err != nil {
		return err
	}

	// build request
	c.req, err = http.NewRequest(c.Method, ru.string(), bytes.NewReader(c.Body))
	if err != nil {
		return err
	}

	// apply Header to request
	if c.Method == "POST" {
		if c.ContentType == "" {
			c.ContentType = ApplicationJSON
		}
		c.req.Header.Set("Content-Type", string(c.ContentType))
	}
	for k, v := range c.Header {
		c.req.Header.Add(k, v)
	}

	// apply basic Auth of request header
	if c.BasicAuth.Username != "" && c.BasicAuth.Password != "" {
		c.req.SetBasicAuth(c.BasicAuth.Username, c.BasicAuth.Password)
	}

	c.client = &http.Client{}

	// apply timeout
	if c.Timeout > 0 {
		c.client.Timeout = c.Timeout * time.Second
	}

	// apply cookies
	if c.Cookies != nil {
		jar, _ := cookiejar.New(nil)
		jar.SetCookies(&url.URL{Scheme: ru.scheme(), Host: ru.host()}, c.Cookies)
		c.client.Jar = jar
	}

	// apply proxy
	if c.ProxyURL != "" {
		if proxy, err := url.Parse(c.ProxyURL); err == nil && proxy != nil {
			c.client.Transport = &http.Transport{
				Proxy:           http.ProxyURL(proxy),
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			}
		}
	}

	return err
}

// Resp do request and get original http response struct
func (c *Client) Resp() (resp *http.Response, err error) {
	if err = c.buildRequest(); err != nil {
		return resp, err
	}
	return c.client.Do(c.req)
}
