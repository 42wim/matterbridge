package slack

import (
	"context"
	"net/url"
)

// AuthRevokeResponse contains our Auth response from the auth.revoke endpoint
type AuthRevokeResponse struct {
	SlackResponse      // Contains the "ok", and "Error", if any
	Revoked       bool `json:"revoked,omitempty"`
}

// authRequest sends the actual request, and unmarshals the response
func authRequest(ctx context.Context, client httpClient, path string, values url.Values, d debug) (*AuthRevokeResponse, error) {
	response := &AuthRevokeResponse{}
	err := postSlackMethod(ctx, client, path, values, response, d)
	if err != nil {
		return nil, err
	}

	return response, response.Err()
}

// SendAuthRevoke will send a revocation for our token
func (api *Client) SendAuthRevoke(token string) (*AuthRevokeResponse, error) {
	return api.SendAuthRevokeContext(context.Background(), token)
}

// SendAuthRevokeContext will retrieve the satus from api.test
func (api *Client) SendAuthRevokeContext(ctx context.Context, token string) (*AuthRevokeResponse, error) {
	if token == "" {
		token = api.token
	}
	values := url.Values{
		"token": {token},
	}

	return authRequest(ctx, api.httpclient, "auth.revoke", values, api)
}
