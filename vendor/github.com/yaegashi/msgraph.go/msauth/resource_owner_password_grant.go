package msauth

import (
	"context"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/microsoft"
)

// ResourceOwnerPasswordGrant preforms OAuth 2.0 client resource owner password grant and returns a token.
func (m *Manager) ResourceOwnerPasswordGrant(ctx context.Context, tenantID, clientID, clientSecret, username, password string, scopes []string) (oauth2.TokenSource, error) {
	endpoint := microsoft.AzureADEndpoint(tenantID)
	endpoint.AuthStyle = oauth2.AuthStyleInParams
	config := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Endpoint:     endpoint,
		Scopes:       scopes,
	}
	t, err := config.PasswordCredentialsToken(ctx, username, password)
	if err != nil {
		return nil, err
	}
	ts := config.TokenSource(ctx, t)
	return ts, nil
}
