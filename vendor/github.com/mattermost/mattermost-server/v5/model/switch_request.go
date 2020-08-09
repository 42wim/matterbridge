// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package model

import (
	"encoding/json"
	"io"
)

type SwitchRequest struct {
	CurrentService string `json:"current_service"`
	NewService     string `json:"new_service"`
	Email          string `json:"email"`
	Password       string `json:"password"`
	NewPassword    string `json:"new_password"`
	MfaCode        string `json:"mfa_code"`
	LdapLoginId    string `json:"ldap_id"`
}

func (o *SwitchRequest) ToJson() string {
	b, _ := json.Marshal(o)
	return string(b)
}

func SwitchRequestFromJson(data io.Reader) *SwitchRequest {
	var o *SwitchRequest
	json.NewDecoder(data).Decode(&o)
	return o
}

func (o *SwitchRequest) EmailToOAuth() bool {
	return o.CurrentService == USER_AUTH_SERVICE_EMAIL &&
		(o.NewService == USER_AUTH_SERVICE_SAML ||
			o.NewService == USER_AUTH_SERVICE_GITLAB ||
			o.NewService == SERVICE_GOOGLE ||
			o.NewService == SERVICE_OFFICE365)
}

func (o *SwitchRequest) OAuthToEmail() bool {
	return (o.CurrentService == USER_AUTH_SERVICE_SAML ||
		o.CurrentService == USER_AUTH_SERVICE_GITLAB ||
		o.CurrentService == SERVICE_GOOGLE ||
		o.CurrentService == SERVICE_OFFICE365) && o.NewService == USER_AUTH_SERVICE_EMAIL
}

func (o *SwitchRequest) EmailToLdap() bool {
	return o.CurrentService == USER_AUTH_SERVICE_EMAIL && o.NewService == USER_AUTH_SERVICE_LDAP
}

func (o *SwitchRequest) LdapToEmail() bool {
	return o.CurrentService == USER_AUTH_SERVICE_LDAP && o.NewService == USER_AUTH_SERVICE_EMAIL
}
