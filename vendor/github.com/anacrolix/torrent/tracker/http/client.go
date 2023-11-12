package http

import (
	"crypto/tls"
	"net/http"
	"net/url"
)

type Client struct {
	hc   *http.Client
	url_ *url.URL
}

type ProxyFunc func(*http.Request) (*url.URL, error)

type NewClientOpts struct {
	Proxy          ProxyFunc
	ServerName     string
	AllowKeepAlive bool
}

func NewClient(url_ *url.URL, opts NewClientOpts) Client {
	return Client{
		url_: url_,
		hc: &http.Client{
			Transport: &http.Transport{
				Proxy: opts.Proxy,
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true,
					ServerName:         opts.ServerName,
				},
				// This is for S3 trackers that hold connections open.
				DisableKeepAlives: !opts.AllowKeepAlive,
			},
		},
	}
}

func (cl Client) Close() error {
	cl.hc.CloseIdleConnections()
	return nil
}
