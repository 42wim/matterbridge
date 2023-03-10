package api // import "github.com/SevereCloud/vksdk/v2/api"

import (
	"github.com/SevereCloud/vksdk/v2/object"
)

// AuthCheckPhone checks a user's phone number for correctness.
//
// https://vk.com/dev/auth.checkPhone
//
// Deprecated: This method is deprecated and may be disabled soon, please avoid
// using it.
func (vk *VK) AuthCheckPhone(params Params) (response int, err error) {
	err = vk.RequestUnmarshal("auth.checkPhone", &response, params)
	return
}

// AuthRestoreResponse struct.
type AuthRestoreResponse struct {
	Success int    `json:"success"`
	SID     string `json:"sid"`
}

// AuthRestore allows to restore account access using a code received via SMS.
//
// https://vk.com/dev/auth.restore
func (vk *VK) AuthRestore(params Params) (response AuthRestoreResponse, err error) {
	err = vk.RequestUnmarshal("auth.restore", &response, params)
	return
}

// AuthGetProfileInfoBySilentTokenResponse struct.
type AuthGetProfileInfoBySilentTokenResponse struct {
	Success []object.AuthSilentTokenProfile `json:"success"`
	Errors  []AuthSilentTokenError          `json:"errors"`
}

// AuthGetProfileInfoBySilentToken method.
//
// https://platform.vk.com/?p=DocsDashboard&docs=tokens_silent-token
func (vk *VK) AuthGetProfileInfoBySilentToken(params Params) (response AuthGetProfileInfoBySilentTokenResponse, err error) {
	err = vk.RequestUnmarshal("auth.getProfileInfoBySilentToken", &response, params)
	return
}

// ExchangeSilentTokenSource call conditions exchangeSilentToken.
//
//	0	Unknown
//	1	Silent authentication
//	2	Auth by login and password
//	3	Extended registration
//	4	Auth by exchange token
//	5	Auth by exchange token on reset password
//	6	Auth by exchange token on unblock
//	7	Auth by exchange token on reset session
//	8	Auth by exchange token on change password
//	9	Finish phone validation on authentication
//	10	Auth by code
//	11	Auth by external oauth
//	12	Reactivation
//	15	Auth by SDK temporary access-token
type ExchangeSilentTokenSource int

// AuthExchangeSilentAuthTokenResponse struct.
type AuthExchangeSilentAuthTokenResponse struct {
	AccessToken              string                    `json:"access_token"`
	AccessTokenID            string                    `json:"access_token_id"`
	UserID                   int                       `json:"user_id"`
	Phone                    string                    `json:"phone"`
	PhoneValidated           interface{}               `json:"phone_validated"`
	IsPartial                bool                      `json:"is_partial"`
	IsService                bool                      `json:"is_service"`
	AdditionalSignupRequired bool                      `json:"additional_signup_required"`
	Email                    string                    `json:"email"`
	Source                   ExchangeSilentTokenSource `json:"source"`
	SourceDescription        string                    `json:"source_description"`
}

// AuthExchangeSilentAuthToken method.
//
// https://platform.vk.com/?p=DocsDashboard&docs=tokens_access-token
func (vk *VK) AuthExchangeSilentAuthToken(params Params) (response AuthExchangeSilentAuthTokenResponse, err error) {
	err = vk.RequestUnmarshal("auth.exchangeSilentAuthToken", &response, params)
	return
}
