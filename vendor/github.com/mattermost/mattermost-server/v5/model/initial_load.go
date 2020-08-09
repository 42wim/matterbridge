// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package model

import (
	"encoding/json"
	"io"
)

type InitialLoad struct {
	User        *User             `json:"user"`
	TeamMembers []*TeamMember     `json:"team_members"`
	Teams       []*Team           `json:"teams"`
	Preferences Preferences       `json:"preferences"`
	ClientCfg   map[string]string `json:"client_cfg"`
	LicenseCfg  map[string]string `json:"license_cfg"`
	NoAccounts  bool              `json:"no_accounts"`
}

func (me *InitialLoad) ToJson() string {
	b, _ := json.Marshal(me)
	return string(b)
}

func InitialLoadFromJson(data io.Reader) *InitialLoad {
	var o *InitialLoad
	json.NewDecoder(data).Decode(&o)
	return o
}
