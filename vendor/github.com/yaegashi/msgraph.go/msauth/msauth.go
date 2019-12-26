// Package msauth implements a library to authorize against Microsoft identity platform:
// https://docs.microsoft.com/en-us/azure/active-directory/develop/
//
// It utilizes v2.0 endpoint
// so it can authorize users with both personal (Microsoft) and organizational (Azure AD) account.
package msauth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"sync"
	"time"

	"golang.org/x/oauth2"
)

const (
	// DefaultMSGraphScope is the default scope for MS Graph API
	DefaultMSGraphScope = "https://graph.microsoft.com/.default"
	endpointURLFormat   = "https://login.microsoftonline.com/%s/oauth2/v2.0/%s"
)

// TokenError is returned on failed authentication
type TokenError struct {
	ErrorObject      string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

// Error implements error interface
func (t *TokenError) Error() string {
	return fmt.Sprintf("%s: %s", t.ErrorObject, t.ErrorDescription)
}

func generateKey(tenantID, clientID string) string {
	return fmt.Sprintf("%s:%s", tenantID, clientID)
}

func deviceCodeURL(tenantID string) string {
	return fmt.Sprintf(endpointURLFormat, tenantID, "devicecode")
}

func tokenURL(tenantID string) string {
	return fmt.Sprintf(endpointURLFormat, tenantID, "token")
}

type tokenJSON struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
}

func (e *tokenJSON) expiry() (t time.Time) {
	if v := e.ExpiresIn; v != 0 {
		return time.Now().Add(time.Duration(v) * time.Second)
	}
	return
}

// Manager is oauth2 token cache manager
type Manager struct {
	mu         sync.Mutex
	TokenCache map[string]*oauth2.Token
}

// NewManager returns a new Manager instance
func NewManager() *Manager {
	return &Manager{TokenCache: map[string]*oauth2.Token{}}
}

// LoadBytes loads token cache from opaque bytes (it's actually JSON)
func (m *Manager) LoadBytes(b []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	return json.Unmarshal(b, &m.TokenCache)
}

// SaveBytes saves token cache to opaque bytes (it's actually JSON)
func (m *Manager) SaveBytes() ([]byte, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return json.Marshal(m.TokenCache)
}

// LoadFile loads token cache from file
func (m *Manager) LoadFile(path string) error {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	return m.LoadBytes(b)
}

// SaveFile saves token cache to file
func (m *Manager) SaveFile(path string) error {
	b, err := m.SaveBytes()
	if err != nil {
		return err
	}
	return ioutil.WriteFile(path, b, 0644)
}

// Cache stores a token into token cache
func (m *Manager) Cache(tenantID, clientID string, token *oauth2.Token) {
	m.TokenCache[generateKey(tenantID, clientID)] = token
}

// requestToken requests a token from the token endpoint
// TODO(ctx): use http client from ctx
func (m *Manager) requestToken(ctx context.Context, tenantID, clientID string, values url.Values) (*oauth2.Token, error) {
	res, err := http.PostForm(tokenURL(tenantID), values)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		var terr *TokenError
		err = json.Unmarshal(b, &terr)
		if err != nil {
			return nil, err
		}
		return nil, terr
	}
	var tj *tokenJSON
	err = json.Unmarshal(b, &tj)
	if err != nil {
		return nil, err
	}
	token := &oauth2.Token{
		AccessToken:  tj.AccessToken,
		TokenType:    tj.TokenType,
		RefreshToken: tj.RefreshToken,
		Expiry:       tj.expiry(),
	}
	if token.AccessToken == "" {
		return nil, errors.New("msauth: server response missing access_token")
	}
	return token, nil
}
