package giphy

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

// DefaultClient is the default Giphy API client
var DefaultClient = NewClient()

// PublicBetaKey is the public beta key for the Giphy API
var PublicBetaKey = "dc6zaTOxFJmzC"

// A Client communicates with the Giphy API.
type Client struct {
	// APIKey is the key used for requests to the Giphy API
	APIKey string

	// Limit is the limit used for requests to the Giphy API
	Limit int

	// Rating is the rating used for requests to the Giphy API
	Rating string

	// BaseURL is the base url for Giphy API.
	BaseURL *url.URL

	// BasePath is the base path for the gifs endpoints
	BasePath string

	// User agent used for HTTP requests to Giphy API.
	UserAgent string

	// HTTP client used to communicate with the Giphy API.
	httpClient *http.Client
}

// NewClient returns a new Giphy API client.
// If no *http.Client were provided then http.DefaultClient is used.
func NewClient(httpClients ...*http.Client) *Client {
	var httpClient *http.Client

	if len(httpClients) > 0 && httpClients[0] != nil {
		httpClient = httpClients[0]
	} else {
		cloned := *http.DefaultClient
		httpClient = &cloned
	}

	c := &Client{
		APIKey: Env("GIPHY_API_KEY", PublicBetaKey),
		Rating: Env("GIPHY_RATING", "g"),
		Limit:  EnvInt("GIPHY_LIMIT", 10),
		BaseURL: &url.URL{
			Scheme: Env("GIPHY_BASE_URL_SCHEME", "https"),
			Host:   Env("GIPHY_BASE_URL_HOST", "api.giphy.com"),
		},
		BasePath:   Env("GIPHY_BASE_PATH", "/v1"),
		UserAgent:  Env("GIPHY_USER_AGENT", "giphy.go"),
		httpClient: httpClient,
	}

	return c
}

// NewRequest creates an API request.
func (c *Client) NewRequest(s string) (*http.Request, error) {
	rel, err := url.Parse(c.BasePath + s)
	if err != nil {
		return nil, err
	}

	q := rel.Query()
	q.Set("api_key", c.APIKey)
	q.Set("rating", c.Rating)
	rel.RawQuery = q.Encode()

	u := c.BaseURL.ResolveReference(rel)

	if EnvBool("GIPHY_VERBOSE", false) {
		fmt.Println("giphy: GET", u.String())
	}

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("User-Agent", c.UserAgent)
	return req, nil
}

// Do sends an API request and returns the API response. The API response is
// decoded and stored in the value pointed to by v, or returned as an error if
// an API error has occurred.
func (c *Client) Do(req *http.Request, v interface{}) (*http.Response, error) {
	// Make sure to close the connection after replying to this request
	req.Close = true

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return resp, err
	}
	defer resp.Body.Close()

	if v != nil {
		err = json.NewDecoder(resp.Body).Decode(v)
	}

	if err != nil {
		return nil, fmt.Errorf("error reading response from %s %s: %s", req.Method, req.URL.RequestURI(), err)
	}

	return resp, nil
}
