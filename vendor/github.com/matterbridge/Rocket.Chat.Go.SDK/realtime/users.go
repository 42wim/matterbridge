package realtime

import (
	"crypto/sha256"
	"encoding/hex"
	"strconv"

	"github.com/Jeffail/gabs"
	"github.com/matterbridge/Rocket.Chat.Go.SDK/models"
)

type ddpLoginRequest struct {
	User     ddpUser     `json:"user"`
	Password ddpPassword `json:"password"`
}

type ddpTokenLoginRequest struct {
	Token string `json:"resume"`
}

type ddpUser struct {
	Email string `json:"email"`
}

type ddpPassword struct {
	Digest    string `json:"digest"`
	Algorithm string `json:"algorithm"`
}

// RegisterUser a new user on the server. This function does not need a logged in user. The registered user gets logged in
// to set its username.
func (c *Client) RegisterUser(credentials *models.UserCredentials) (*models.User, error) {

	if _, err := c.ddp.Call("registerUser", credentials); err != nil {
		return nil, err
	}

	user, err := c.Login(credentials)
	if err != nil {
		return nil, err
	}

	if _, err := c.ddp.Call("setUsername", credentials.Name); err != nil {
		return nil, err
	}

	return user, nil
}

// Login a user.
// token shouldn't be nil, otherwise the password and the email are not allowed to be nil.
//
// https://rocket.chat/docs/developer-guides/realtime-api/method-calls/login/
func (c *Client) Login(credentials *models.UserCredentials) (*models.User, error) {
	var request interface{}
	if credentials.Token != "" {
		request = ddpTokenLoginRequest{
			Token: credentials.Token,
		}
	} else {
		digest := sha256.Sum256([]byte(credentials.Password))
		request = ddpLoginRequest{
			User: ddpUser{Email: credentials.Email},
			Password: ddpPassword{
				Digest:    hex.EncodeToString(digest[:]),
				Algorithm: "sha-256",
			},
		}
	}

	rawResponse, err := c.ddp.Call("login", request)
	if err != nil {
		return nil, err
	}

	user := getUserFromData(rawResponse.(map[string]interface{}))
	if credentials.Token == "" {
		credentials.ID, credentials.Token = user.ID, user.Token
	}

	return user, nil
}

func getUserFromData(data interface{}) *models.User {
	document, _ := gabs.Consume(data)

	expires, _ := strconv.ParseFloat(stringOrZero(document.Path("tokenExpires.$date").Data()), 64)
	return &models.User{
		ID:           stringOrZero(document.Path("id").Data()),
		Token:        stringOrZero(document.Path("token").Data()),
		TokenExpires: int64(expires),
	}
}

// SetPresence set user presence
func (c *Client) SetPresence(status string) error {
	_, err := c.ddp.Call("UserPresence:setDefaultStatus", status)
	if err != nil {
		return err
	}

	return nil
}
