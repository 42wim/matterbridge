// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package model

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const (
	USERNAME = "Username"
)

type TeamMember struct {
	TeamId        string `json:"team_id"`
	UserId        string `json:"user_id"`
	Roles         string `json:"roles"`
	DeleteAt      int64  `json:"delete_at"`
	SchemeGuest   bool   `json:"scheme_guest"`
	SchemeUser    bool   `json:"scheme_user"`
	SchemeAdmin   bool   `json:"scheme_admin"`
	ExplicitRoles string `json:"explicit_roles"`
}

type TeamUnread struct {
	TeamId       string `json:"team_id"`
	MsgCount     int64  `json:"msg_count"`
	MentionCount int64  `json:"mention_count"`
}

type TeamMemberForExport struct {
	TeamMember
	TeamName string
}

type TeamMemberWithError struct {
	UserId string      `json:"user_id"`
	Member *TeamMember `json:"member"`
	Error  *AppError   `json:"error"`
}

type EmailInviteWithError struct {
	Email string    `json:"email"`
	Error *AppError `json:"error"`
}

type TeamMembersGetOptions struct {
	// Sort the team members. Accepts "Username", but defaults to "Id".
	Sort string

	// If true, exclude team members whose corresponding user is deleted.
	ExcludeDeletedUsers bool

	// Restrict to search in a list of teams and channels
	ViewRestrictions *ViewUsersRestrictions
}

func (o *TeamMember) ToJson() string {
	b, _ := json.Marshal(o)
	return string(b)
}

func (o *TeamUnread) ToJson() string {
	b, _ := json.Marshal(o)
	return string(b)
}

func TeamMemberFromJson(data io.Reader) *TeamMember {
	var o *TeamMember
	json.NewDecoder(data).Decode(&o)
	return o
}

func TeamUnreadFromJson(data io.Reader) *TeamUnread {
	var o *TeamUnread
	json.NewDecoder(data).Decode(&o)
	return o
}

func EmailInviteWithErrorFromJson(data io.Reader) []*EmailInviteWithError {
	var o []*EmailInviteWithError
	json.NewDecoder(data).Decode(&o)
	return o
}

func EmailInviteWithErrorToEmails(o []*EmailInviteWithError) []string {
	var ret []string
	for _, o := range o {
		if o.Error == nil {
			ret = append(ret, o.Email)
		}
	}
	return ret
}

func EmailInviteWithErrorToJson(o []*EmailInviteWithError) string {
	if b, err := json.Marshal(o); err != nil {
		return "[]"
	} else {
		return string(b)
	}
}

func EmailInviteWithErrorToString(o *EmailInviteWithError) string {
	return fmt.Sprintf("%s:%s", o.Email, o.Error.Error())
}

func TeamMembersWithErrorToTeamMembers(o []*TeamMemberWithError) []*TeamMember {
	var ret []*TeamMember
	for _, o := range o {
		if o.Error == nil {
			ret = append(ret, o.Member)
		}
	}
	return ret
}

func TeamMembersWithErrorToJson(o []*TeamMemberWithError) string {
	if b, err := json.Marshal(o); err != nil {
		return "[]"
	} else {
		return string(b)
	}
}

func TeamMemberWithErrorToString(o *TeamMemberWithError) string {
	return fmt.Sprintf("%s:%s", o.UserId, o.Error.Error())
}

func TeamMembersWithErrorFromJson(data io.Reader) []*TeamMemberWithError {
	var o []*TeamMemberWithError
	json.NewDecoder(data).Decode(&o)
	return o
}

func TeamMembersToJson(o []*TeamMember) string {
	if b, err := json.Marshal(o); err != nil {
		return "[]"
	} else {
		return string(b)
	}
}

func TeamMembersFromJson(data io.Reader) []*TeamMember {
	var o []*TeamMember
	json.NewDecoder(data).Decode(&o)
	return o
}

func TeamsUnreadToJson(o []*TeamUnread) string {
	if b, err := json.Marshal(o); err != nil {
		return "[]"
	} else {
		return string(b)
	}
}

func TeamsUnreadFromJson(data io.Reader) []*TeamUnread {
	var o []*TeamUnread
	json.NewDecoder(data).Decode(&o)
	return o
}

func (o *TeamMember) IsValid() *AppError {

	if !IsValidId(o.TeamId) {
		return NewAppError("TeamMember.IsValid", "model.team_member.is_valid.team_id.app_error", nil, "", http.StatusBadRequest)
	}

	if !IsValidId(o.UserId) {
		return NewAppError("TeamMember.IsValid", "model.team_member.is_valid.user_id.app_error", nil, "", http.StatusBadRequest)
	}

	return nil
}

func (o *TeamMember) PreUpdate() {
}

func (o *TeamMember) GetRoles() []string {
	return strings.Fields(o.Roles)
}
