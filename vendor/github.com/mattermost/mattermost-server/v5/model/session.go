// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package model

import (
	"encoding/json"
	"io"
	"strconv"
	"strings"

	"github.com/mattermost/mattermost-server/v5/mlog"
)

const (
	SESSION_COOKIE_TOKEN              = "MMAUTHTOKEN"
	SESSION_COOKIE_USER               = "MMUSERID"
	SESSION_COOKIE_CSRF               = "MMCSRF"
	SESSION_CACHE_SIZE                = 35000
	SESSION_PROP_PLATFORM             = "platform"
	SESSION_PROP_OS                   = "os"
	SESSION_PROP_BROWSER              = "browser"
	SESSION_PROP_TYPE                 = "type"
	SESSION_PROP_USER_ACCESS_TOKEN_ID = "user_access_token_id"
	SESSION_PROP_IS_BOT               = "is_bot"
	SESSION_PROP_IS_BOT_VALUE         = "true"
	SESSION_TYPE_USER_ACCESS_TOKEN    = "UserAccessToken"
	SESSION_PROP_IS_GUEST             = "is_guest"
	SESSION_ACTIVITY_TIMEOUT          = 1000 * 60 * 5 // 5 minutes
	SESSION_USER_ACCESS_TOKEN_EXPIRY  = 100 * 365     // 100 years
)

//msgp:tuple Session

// Session contains the user session details.
// This struct's serializer methods are auto-generated. If a new field is added/removed,
// please run make gen-serialized.
type Session struct {
	Id             string        `json:"id"`
	Token          string        `json:"token"`
	CreateAt       int64         `json:"create_at"`
	ExpiresAt      int64         `json:"expires_at"`
	LastActivityAt int64         `json:"last_activity_at"`
	UserId         string        `json:"user_id"`
	DeviceId       string        `json:"device_id"`
	Roles          string        `json:"roles"`
	IsOAuth        bool          `json:"is_oauth"`
	ExpiredNotify  bool          `json:"expired_notify"`
	Props          StringMap     `json:"props"`
	TeamMembers    []*TeamMember `json:"team_members" db:"-"`
	Local          bool          `json:"local" db:"-"`
}

// Returns true if the session is unrestricted, which should grant it
// with all permissions. This is used for local mode sessions
func (me *Session) IsUnrestricted() bool {
	return me.Local
}

func (me *Session) DeepCopy() *Session {
	copySession := *me

	if me.Props != nil {
		copySession.Props = CopyStringMap(me.Props)
	}

	if me.TeamMembers != nil {
		copySession.TeamMembers = make([]*TeamMember, len(me.TeamMembers))
		for index, tm := range me.TeamMembers {
			copySession.TeamMembers[index] = new(TeamMember)
			*copySession.TeamMembers[index] = *tm
		}
	}

	return &copySession
}

func (me *Session) ToJson() string {
	b, _ := json.Marshal(me)
	return string(b)
}

func SessionFromJson(data io.Reader) *Session {
	var me *Session
	json.NewDecoder(data).Decode(&me)
	return me
}

func (me *Session) PreSave() {
	if me.Id == "" {
		me.Id = NewId()
	}

	if me.Token == "" {
		me.Token = NewId()
	}

	me.CreateAt = GetMillis()
	me.LastActivityAt = me.CreateAt

	if me.Props == nil {
		me.Props = make(map[string]string)
	}
}

func (me *Session) Sanitize() {
	me.Token = ""
}

func (me *Session) IsExpired() bool {

	if me.ExpiresAt <= 0 {
		return false
	}

	if GetMillis() > me.ExpiresAt {
		return true
	}

	return false
}

// Deprecated: SetExpireInDays is deprecated and should not be used.
//             Use (*App).SetSessionExpireInDays instead which handles the
//			   cases where the new ExpiresAt is not relative to CreateAt.
func (me *Session) SetExpireInDays(days int) {
	if me.CreateAt == 0 {
		me.ExpiresAt = GetMillis() + (1000 * 60 * 60 * 24 * int64(days))
	} else {
		me.ExpiresAt = me.CreateAt + (1000 * 60 * 60 * 24 * int64(days))
	}
}

func (me *Session) AddProp(key string, value string) {

	if me.Props == nil {
		me.Props = make(map[string]string)
	}

	me.Props[key] = value
}

func (me *Session) GetTeamByTeamId(teamId string) *TeamMember {
	for _, team := range me.TeamMembers {
		if team.TeamId == teamId {
			return team
		}
	}

	return nil
}

func (me *Session) IsMobileApp() bool {
	return len(me.DeviceId) > 0 || me.IsMobile()
}

func (me *Session) IsMobile() bool {
	val, ok := me.Props[USER_AUTH_SERVICE_IS_MOBILE]
	if !ok {
		return false
	}
	isMobile, err := strconv.ParseBool(val)
	if err != nil {
		mlog.Error("Error parsing boolean property from Session", mlog.Err(err))
		return false
	}
	return isMobile
}

func (me *Session) IsSaml() bool {
	val, ok := me.Props[USER_AUTH_SERVICE_IS_SAML]
	if !ok {
		return false
	}
	isSaml, err := strconv.ParseBool(val)
	if err != nil {
		mlog.Error("Error parsing boolean property from Session", mlog.Err(err))
		return false
	}
	return isSaml
}

func (me *Session) IsOAuthUser() bool {
	val, ok := me.Props[USER_AUTH_SERVICE_IS_OAUTH]
	if !ok {
		return false
	}
	isOAuthUser, err := strconv.ParseBool(val)
	if err != nil {
		mlog.Error("Error parsing boolean property from Session", mlog.Err(err))
		return false
	}
	return isOAuthUser
}

func (me *Session) IsSSOLogin() bool {
	return me.IsOAuthUser() || me.IsSaml()
}

func (me *Session) GetUserRoles() []string {
	return strings.Fields(me.Roles)
}

func (me *Session) GenerateCSRF() string {
	token := NewId()
	me.AddProp("csrf", token)
	return token
}

func (me *Session) GetCSRF() string {
	if me.Props == nil {
		return ""
	}

	return me.Props["csrf"]
}

func SessionsToJson(o []*Session) string {
	if b, err := json.Marshal(o); err != nil {
		return "[]"
	} else {
		return string(b)
	}
}

func SessionsFromJson(data io.Reader) []*Session {
	var o []*Session
	json.NewDecoder(data).Decode(&o)
	return o
}
