package msauth

import (
	"context"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
	"golang.org/x/oauth2/microsoft"
)

// ClientCredentialsGrant performs OAuth 2.0 client credentials grant and returns auto-refreshing TokenSource
func (m *Manager) ClientCredentialsGrant(ctx context.Context, tenantID, clientID, clientSecret string, scopes []string) (oauth2.TokenSource, error) {
	config := &clientcredentials.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		TokenURL:     microsoft.AzureADEndpoint(tenantID).TokenURL,
		Scopes:       scopes,
		AuthStyle:    oauth2.AuthStyleInParams,
	}
	var err error
	ts := config.TokenSource(ctx)
	_, err = ts.Token()
	if err != nil {
		return nil, err
	}
	return ts, nil
}
