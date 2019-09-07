package slack

import (
	"context"
	"net/url"
)

// OAuthResponseIncomingWebhook ...
type OAuthResponseIncomingWebhook struct {
	URL              string `json:"url"`
	Channel          string `json:"channel"`
	ChannelID        string `json:"channel_id,omitempty"`
	ConfigurationURL string `json:"configuration_url"`
}

// OAuthResponseBot ...
type OAuthResponseBot struct {
	BotUserID      string `json:"bot_user_id"`
	BotAccessToken string `json:"bot_access_token"`
}

// OAuthResponse ...
type OAuthResponse struct {
	AccessToken     string                       `json:"access_token"`
	Scope           string                       `json:"scope"`
	TeamName        string                       `json:"team_name"`
	TeamID          string                       `json:"team_id"`
	IncomingWebhook OAuthResponseIncomingWebhook `json:"incoming_webhook"`
	Bot             OAuthResponseBot             `json:"bot"`
	UserID          string                       `json:"user_id,omitempty"`
	SlackResponse
}

// GetOAuthToken retrieves an AccessToken
func GetOAuthToken(client httpClient, clientID, clientSecret, code, redirectURI string) (accessToken string, scope string, err error) {
	return GetOAuthTokenContext(context.Background(), client, clientID, clientSecret, code, redirectURI)
}

// GetOAuthTokenContext retrieves an AccessToken with a custom context
func GetOAuthTokenContext(ctx context.Context, client httpClient, clientID, clientSecret, code, redirectURI string) (accessToken string, scope string, err error) {
	response, err := GetOAuthResponseContext(ctx, client, clientID, clientSecret, code, redirectURI)
	if err != nil {
		return "", "", err
	}
	return response.AccessToken, response.Scope, nil
}

func GetOAuthResponse(client httpClient, clientID, clientSecret, code, redirectURI string) (resp *OAuthResponse, err error) {
	return GetOAuthResponseContext(context.Background(), client, clientID, clientSecret, code, redirectURI)
}

func GetOAuthResponseContext(ctx context.Context, client httpClient, clientID, clientSecret, code, redirectURI string) (resp *OAuthResponse, err error) {
	values := url.Values{
		"client_id":     {clientID},
		"client_secret": {clientSecret},
		"code":          {code},
		"redirect_uri":  {redirectURI},
	}
	response := &OAuthResponse{}
	if err = postForm(ctx, client, APIURL+"oauth.access", values, response, discard{}); err != nil {
		return nil, err
	}
	return response, response.Err()
}
