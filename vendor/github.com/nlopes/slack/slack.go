package slack

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
)

// APIURL a dded as a var so that we can change this for testing purposes
var APIURL = "https://slack.com/api/"

// WEBAPIURLFormat ...
const WEBAPIURLFormat = "https://%s.slack.com/api/users.admin.%s?t=%d"

// httpClient defines the minimal interface needed for an http.Client to be implemented.
type httpClient interface {
	Do(*http.Request) (*http.Response, error)
}

// ResponseMetadata holds pagination metadata
type ResponseMetadata struct {
	Cursor string `json:"next_cursor"`
}

func (t *ResponseMetadata) initialize() *ResponseMetadata {
	if t != nil {
		return t
	}

	return &ResponseMetadata{}
}

// AuthTestResponse ...
type AuthTestResponse struct {
	URL    string `json:"url"`
	Team   string `json:"team"`
	User   string `json:"user"`
	TeamID string `json:"team_id"`
	UserID string `json:"user_id"`
}

type authTestResponseFull struct {
	SlackResponse
	AuthTestResponse
}

// Client for the slack api.
type Client struct {
	token      string
	info       Info
	debug      bool
	log        ilogger
	httpclient httpClient
}

// Option defines an option for a Client
type Option func(*Client)

// OptionHTTPClient - provide a custom http client to the slack client.
func OptionHTTPClient(client httpClient) func(*Client) {
	return func(c *Client) {
		c.httpclient = client
	}
}

// OptionDebug enable debugging for the client
func OptionDebug(b bool) func(*Client) {
	return func(c *Client) {
		c.debug = b
	}
}

// OptionLog set logging for client.
func OptionLog(l logger) func(*Client) {
	return func(c *Client) {
		c.log = internalLog{logger: l}
	}
}

// New builds a slack client from the provided token and options.
func New(token string, options ...Option) *Client {
	s := &Client{
		token:      token,
		httpclient: &http.Client{},
		log:        log.New(os.Stderr, "nlopes/slack", log.LstdFlags|log.Lshortfile),
	}

	for _, opt := range options {
		opt(s)
	}

	return s
}

// AuthTest tests if the user is able to do authenticated requests or not
func (api *Client) AuthTest() (response *AuthTestResponse, error error) {
	return api.AuthTestContext(context.Background())
}

// AuthTestContext tests if the user is able to do authenticated requests or not with a custom context
func (api *Client) AuthTestContext(ctx context.Context) (response *AuthTestResponse, err error) {
	api.Debugf("Challenging auth...")
	responseFull := &authTestResponseFull{}
	err = postSlackMethod(ctx, api.httpclient, "auth.test", url.Values{"token": {api.token}}, responseFull, api)
	if err != nil {
		return nil, err
	}

	return &responseFull.AuthTestResponse, responseFull.Err()
}

// Debugf print a formatted debug line.
func (api *Client) Debugf(format string, v ...interface{}) {
	if api.debug {
		api.log.Output(2, fmt.Sprintf(format, v...))
	}
}

// Debugln print a debug line.
func (api *Client) Debugln(v ...interface{}) {
	if api.debug {
		api.log.Output(2, fmt.Sprintln(v...))
	}
}

// Debug returns if debug is enabled.
func (api *Client) Debug() bool {
	return api.debug
}
