package request

import (
	"crypto/tls"
	"net/http"
	"time"
)

// ContentType Content-Type
type ContentType string

// Method http method
// TODO:
type Method string

// Client Method
/*
     Method         = "OPTIONS"                ; Section 9.2
                    | "GET"                    ; Section 9.3
                    | "HEAD"                   ; Section 9.4
                    | "POST"                   ; Section 9.5
                    | "PUT"                    ; Section 9.6
                    | "DELETE"                 ; Section 9.7
                    | "TRACE"                  ; Section 9.8
                    | "CONNECT"                ; Section 9.9
                    | extension-method
   extension-method = token
     token          = 1*<any CHAR except CTLs or separators>
*/
type Client struct {
	URL         string
	Method      string
	Header      map[string]string
	Params      map[string]string
	Body        []byte
	BasicAuth   BasicAuth
	Timeout     time.Duration // second
	ProxyURL    string
	ContentType ContentType
	Cookies     []*http.Cookie
	TLSConfig   *tls.Config
	Transport   *http.Transport

	// private
	client     *http.Client
	requestURL requestURL
	req        *http.Request
}

// BasicAuth Add Username:Password as Basic Auth
type BasicAuth struct {
	Username string
	Password string
}

// SugaredResp Sugared response with status code and body data
type SugaredResp struct {
	Data []byte
	Code int

	// private
	resp *http.Response
}
