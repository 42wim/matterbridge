package msauth

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/microsoft"
)

const (
	deviceCodeGrantType       = "urn:ietf:params:oauth:grant-type:device_code"
	authorizationPendingError = "authorization_pending"
)

// DeviceCode is returned on device auth initiation
type DeviceCode struct {
	DeviceCode      string `json:"device_code"`
	UserCode        string `json:"user_code"`
	VerificationURL string `json:"verification_url"`
	ExpiresIn       int    `json:"expires_in"`
	Interval        int    `json:"interval"`
	Message         string `json:"message"`
}

// DeviceAuthorizationGrant performs OAuth 2.0 device authorization grant and returns auto-refreshing TokenSource
func (m *Manager) DeviceAuthorizationGrant(ctx context.Context, tenantID, clientID string, scopes []string, callback func(*DeviceCode) error) (oauth2.TokenSource, error) {
	endpoint := microsoft.AzureADEndpoint(tenantID)
	endpoint.AuthStyle = oauth2.AuthStyleInParams
	config := &oauth2.Config{
		ClientID: clientID,
		Endpoint: endpoint,
		Scopes:   scopes,
	}
	if t, ok := m.GetToken(CacheKey(tenantID, clientID)); ok {
		tt, err := config.TokenSource(ctx, t).Token()
		if err == nil {
			m.PutToken(CacheKey(tenantID, clientID), tt)
			return config.TokenSource(ctx, tt), nil
		}
		if _, ok := err.(*oauth2.RetrieveError); !ok {
			return nil, err
		}
	}
	scope := strings.Join(scopes, " ")
	res, err := http.PostForm(deviceCodeURL(tenantID), url.Values{"client_id": {clientID}, "scope": {scope}})
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		b, _ := ioutil.ReadAll(res.Body)
		return nil, fmt.Errorf("%s: %s", res.Status, string(b))
	}
	dc := &DeviceCode{}
	dec := json.NewDecoder(res.Body)
	err = dec.Decode(&dc)
	if err != nil {
		return nil, err
	}
	if callback != nil {
		err = callback(dc)
		if err != nil {
			return nil, err
		}
	} else {
		fmt.Fprintln(os.Stderr, dc.Message)
	}
	values := url.Values{
		"client_id":   {clientID},
		"grant_type":  {deviceCodeGrantType},
		"device_code": {dc.DeviceCode},
	}
	interval := dc.Interval
	if interval == 0 {
		interval = 5
	}
	for {
		time.Sleep(time.Second * time.Duration(interval))
		token, err := m.requestToken(ctx, tenantID, clientID, values)
		if err == nil {
			m.PutToken(CacheKey(tenantID, clientID), token)
			return config.TokenSource(ctx, token), nil
		}
		tokenError, ok := err.(*TokenError)
		if !ok || tokenError.ErrorObject != authorizationPendingError {
			return nil, err
		}
	}
}
