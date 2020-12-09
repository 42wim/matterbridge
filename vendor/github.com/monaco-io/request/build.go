package request

import (
	"bytes"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"time"
)

// TODO: func unit test coverage
func (c *Client) buildRequest() (err error) {
	if err = c.applyRequest(); err != nil {
		return
	}

	c.applyHTTPHeader()
	c.applyBasicAuth()
	c.applyClient()
	c.applyTimeout()
	c.applyCookies()
	// Apply transport needs to be called before TLSConfig as TLSConfig modifies
	// the http transport
	c.applyTransport()
	c.applyTLSConfig()
	err = c.applyProxy()

	c.client.Transport = c.Transport
	return
}

func (c *Client) applyRequest() (err error) {
	// encode requestURL.httpURL like https://google.com?hello=world&package=request
	c.requestURL = requestURL{
		urlString:  c.URL,
		parameters: c.Params,
	}
	if err = c.requestURL.EncodeURL(); err != nil {
		return
	}
	c.req, err = http.NewRequest(c.Method, c.requestURL.string(), bytes.NewReader(c.Body))
	return
}

func (c *Client) applyHTTPHeader() {
	if c.Method == POST {
		if c.ContentType == emptyString {
			c.ContentType = ApplicationJSON
		}
		c.req.Header.Set(contentType, string(c.ContentType))
	}
	for k, v := range c.Header {
		c.req.Header.Add(k, v)
	}
}

func (c *Client) applyBasicAuth() {
	if c.BasicAuth.Username != emptyString && c.BasicAuth.Password != emptyString {
		c.req.SetBasicAuth(c.BasicAuth.Username, c.BasicAuth.Password)
	}
}

func (c *Client) applyClient() {
	c.client = &http.Client{}
}

func (c *Client) applyTimeout() {
	if c.Timeout > 0 {
		c.client.Timeout = c.Timeout * time.Second
	}
}

func (c *Client) applyCookies() {
	if c.Cookies != nil {
		jar, _ := cookiejar.New(nil)
		jar.SetCookies(&url.URL{Scheme: c.requestURL.scheme(), Host: c.requestURL.host()}, c.Cookies)
		c.client.Jar = jar
	}
}

// TODO: test case
func (c *Client) applyProxy() (err error) {
	if c.ProxyURL != emptyString {
		var proxy *url.URL
		if proxy, err = url.Parse(c.ProxyURL); err != nil {
			return
		} else if proxy != nil {
			c.Transport.Proxy = http.ProxyURL(proxy)
		}
	}
	return
}

func (c *Client) applyTLSConfig() {
	if c.TLSConfig != nil {
		c.Transport.TLSClientConfig = c.TLSConfig
	}
}

func (c *Client) applyTransport() {
	if c.Transport == nil {
		c.Transport = &http.Transport{}
	}
}
