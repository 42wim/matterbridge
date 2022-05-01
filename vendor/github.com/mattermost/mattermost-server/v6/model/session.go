// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package model

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/mattermost/mattermost-server/v6/shared/mlog"
)

const (
	SessionCookieToken            = "MMAUTHTOKEN"
	SessionCookieUser             = "MMUSERID"
	SessionCookieCsrf             = "MMCSRF"
	SessionCookieCloudUrl         = "MMCLOUDURL"
	SessionCacheSize              = 35000
	SessionPropPlatform           = "platform"
	SessionPropOs                 = "os"
	SessionPropBrowser            = "browser"
	SessionPropType               = "type"
	SessionPropUserAccessTokenId  = "user_access_token_id"
	SessionPropIsBot              = "is_bot"
	SessionPropIsBotValue         = "true"
	SessionPropOAuthAppID         = "oauth_app_id"
	SessionPropMattermostAppID    = "mattermost_app_id"
	SessionTypeUserAccessToken    = "UserAccessToken"
	SessionTypeCloudKey           = "CloudKey"
	SessionTypeRemoteclusterToken = "RemoteClusterToken"
	SessionPropIsGuest            = "is_guest"
	SessionActivityTimeout        = 1000 * 60 * 5 // 5 minutes
	SessionUserAccessTokenExpiry  = 100 * 365     // 100 years
)

//msgp StringMap
type StringMap map[string]string

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
func (s *Session) IsUnrestricted() bool {
	return s.Local
}

func (s *Session) DeepCopy() *Session {
	copySession := *s

	if s.Props != nil {
		copySession.Props = CopyStringMap(s.Props)
	}

	if s.TeamMembers != nil {
		copySession.TeamMembers = make([]*TeamMember, len(s.TeamMembers))
		for index, tm := range s.TeamMembers {
			copySession.TeamMembers[index] = new(TeamMember)
			*copySession.TeamMembers[index] = *tm
		}
	}

	return &copySession
}

func (s *Session) IsValid() *AppError {
	if !IsValidId(s.Id) {
		return NewAppError("Session.IsValid", "model.session.is_valid.id.app_error", nil, "", http.StatusBadRequest)
	}

	if !IsValidId(s.UserId) {
		return NewAppError("Session.IsValid", "model.session.is_valid.user_id.app_error", nil, "", http.StatusBadRequest)
	}

	if s.CreateAt == 0 {
		return NewAppError("Session.IsValid", "model.session.is_valid.create_at.app_error", nil, "", http.StatusBadRequest)
	}

	if len(s.Roles) > UserRolesMaxLength {
		return NewAppError("Session.IsValid", "model.session.is_valid.roles_limit.app_error",
			map[string]interface{}{"Limit": UserRolesMaxLength}, "session_id="+s.Id, http.StatusBadRequest)
	}

	return nil
}

func (s *Session) PreSave() {
	if s.Id == "" {
		s.Id = NewId()
	}

	if s.Token == "" {
		s.Token = NewId()
	}

	s.CreateAt = GetMillis()
	s.LastActivityAt = s.CreateAt

	if s.Props == nil {
		s.Props = make(map[string]string)
	}
}

func (s *Session) Sanitize() {
	s.Token = ""
}

func (s *Session) IsExpired() bool {

	if s.ExpiresAt <= 0 {
		return false
	}

	if GetMillis() > s.ExpiresAt {
		return true
	}

	return false
}

func (s *Session) AddProp(key string, value string) {

	if s.Props == nil {
		s.Props = make(map[string]string)
	}

	s.Props[key] = value
}

func (s *Session) GetTeamByTeamId(teamId string) *TeamMember {
	for _, team := range s.TeamMembers {
		if team.TeamId == teamId {
			return team
		}
	}

	return nil
}

func (s *Session) IsMobileApp() bool {
	return s.DeviceId != "" || s.IsMobile()
}

func (s *Session) IsMobile() bool {
	val, ok := s.Props[UserAuthServiceIsMobile]
	if !ok {
		return false
	}
	isMobile, err := strconv.ParseBool(val)
	if err != nil {
		mlog.Debug("Error parsing boolean property from Session", mlog.Err(err))
		return false
	}
	return isMobile
}

func (s *Session) IsSaml() bool {
	val, ok := s.Props[UserAuthServiceIsSaml]
	if !ok {
		return false
	}
	isSaml, err := strconv.ParseBool(val)
	if err != nil {
		mlog.Debug("Error parsing boolean property from Session", mlog.Err(err))
		return false
	}
	return isSaml
}

func (s *Session) IsOAuthUser() bool {
	val, ok := s.Props[UserAuthServiceIsOAuth]
	if !ok {
		return false
	}
	isOAuthUser, err := strconv.ParseBool(val)
	if err != nil {
		mlog.Debug("Error parsing boolean property from Session", mlog.Err(err))
		return false
	}
	return isOAuthUser
}

func (s *Session) IsSSOLogin() bool {
	return s.IsOAuthUser() || s.IsSaml()
}

func (s *Session) GetUserRoles() []string {
	return strings.Fields(s.Roles)
}

func (s *Session) GenerateCSRF() string {
	token := NewId()
	s.AddProp("csrf", token)
	return token
}

func (s *Session) GetCSRF() string {
	if s.Props == nil {
		return ""
	}

	return s.Props["csrf"]
}

func (s *Session) CreateAt_() float64 {
	return float64(s.CreateAt)
}

func (s *Session) ExpiresAt_() float64 {
	return float64(s.ExpiresAt)
}
