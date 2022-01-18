// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package model

import (
	"fmt"
	"regexp"
)

const (
	SchemeDisplayNameMaxLength = 128
	SchemeNameMaxLength        = 64
	SchemeDescriptionMaxLength = 1024
	SchemeScopeTeam            = "team"
	SchemeScopeChannel         = "channel"
	SchemeScopePlaybook        = "playbook"
	SchemeScopeRun             = "run"
)

type Scheme struct {
	Id                        string `json:"id"`
	Name                      string `json:"name"`
	DisplayName               string `json:"display_name"`
	Description               string `json:"description"`
	CreateAt                  int64  `json:"create_at"`
	UpdateAt                  int64  `json:"update_at"`
	DeleteAt                  int64  `json:"delete_at"`
	Scope                     string `json:"scope"`
	DefaultTeamAdminRole      string `json:"default_team_admin_role"`
	DefaultTeamUserRole       string `json:"default_team_user_role"`
	DefaultChannelAdminRole   string `json:"default_channel_admin_role"`
	DefaultChannelUserRole    string `json:"default_channel_user_role"`
	DefaultTeamGuestRole      string `json:"default_team_guest_role"`
	DefaultChannelGuestRole   string `json:"default_channel_guest_role"`
	DefaultPlaybookAdminRole  string `json:"default_playbook_admin_role"`
	DefaultPlaybookMemberRole string `json:"default_playbook_member_role"`
	DefaultRunAdminRole       string `json:"default_run_admin_role"`
	DefaultRunMemberRole      string `json:"default_run_member_role"`
}

type SchemePatch struct {
	Name        *string `json:"name"`
	DisplayName *string `json:"display_name"`
	Description *string `json:"description"`
}

type SchemeIDPatch struct {
	SchemeID *string `json:"scheme_id"`
}

// SchemeConveyor is used for importing and exporting a Scheme and its associated Roles.
type SchemeConveyor struct {
	Name           string  `json:"name"`
	DisplayName    string  `json:"display_name"`
	Description    string  `json:"description"`
	Scope          string  `json:"scope"`
	TeamAdmin      string  `json:"default_team_admin_role"`
	TeamUser       string  `json:"default_team_user_role"`
	TeamGuest      string  `json:"default_team_guest_role"`
	ChannelAdmin   string  `json:"default_channel_admin_role"`
	ChannelUser    string  `json:"default_channel_user_role"`
	ChannelGuest   string  `json:"default_channel_guest_role"`
	PlaybookAdmin  string  `json:"default_playbook_admin_role"`
	PlaybookMember string  `json:"default_playbook_member_role"`
	RunAdmin       string  `json:"default_run_admin_role"`
	RunMember      string  `json:"default_run_member_role"`
	Roles          []*Role `json:"roles"`
}

func (sc *SchemeConveyor) Scheme() *Scheme {
	return &Scheme{
		DisplayName:               sc.DisplayName,
		Name:                      sc.Name,
		Description:               sc.Description,
		Scope:                     sc.Scope,
		DefaultTeamAdminRole:      sc.TeamAdmin,
		DefaultTeamUserRole:       sc.TeamUser,
		DefaultTeamGuestRole:      sc.TeamGuest,
		DefaultChannelAdminRole:   sc.ChannelAdmin,
		DefaultChannelUserRole:    sc.ChannelUser,
		DefaultChannelGuestRole:   sc.ChannelGuest,
		DefaultPlaybookAdminRole:  sc.PlaybookAdmin,
		DefaultPlaybookMemberRole: sc.PlaybookMember,
		DefaultRunAdminRole:       sc.RunAdmin,
		DefaultRunMemberRole:      sc.RunMember,
	}
}

type SchemeRoles struct {
	SchemeAdmin bool `json:"scheme_admin"`
	SchemeUser  bool `json:"scheme_user"`
	SchemeGuest bool `json:"scheme_guest"`
}

func (scheme *Scheme) IsValid() bool {
	if !IsValidId(scheme.Id) {
		return false
	}

	return scheme.IsValidForCreate()
}

func (scheme *Scheme) IsValidForCreate() bool {
	if scheme.DisplayName == "" || len(scheme.DisplayName) > SchemeDisplayNameMaxLength {
		return false
	}

	if !IsValidSchemeName(scheme.Name) {
		return false
	}

	if len(scheme.Description) > SchemeDescriptionMaxLength {
		return false
	}

	switch scheme.Scope {
	case SchemeScopeTeam, SchemeScopeChannel, SchemeScopePlaybook, SchemeScopeRun:
	default:
		return false
	}

	if !IsValidRoleName(scheme.DefaultChannelAdminRole) {
		return false
	}

	if !IsValidRoleName(scheme.DefaultChannelUserRole) {
		return false
	}

	if !IsValidRoleName(scheme.DefaultChannelGuestRole) {
		return false
	}

	if scheme.Scope == SchemeScopeTeam {
		if !IsValidRoleName(scheme.DefaultTeamAdminRole) {
			return false
		}

		if !IsValidRoleName(scheme.DefaultTeamUserRole) {
			return false
		}

		if !IsValidRoleName(scheme.DefaultTeamGuestRole) {
			return false
		}

		if !IsValidRoleName(scheme.DefaultPlaybookAdminRole) {
			return false
		}

		if !IsValidRoleName(scheme.DefaultPlaybookMemberRole) {
			return false
		}

		if !IsValidRoleName(scheme.DefaultRunAdminRole) {
			return false
		}

		if !IsValidRoleName(scheme.DefaultRunMemberRole) {
			return false
		}
	}

	if scheme.Scope == SchemeScopeChannel {
		if scheme.DefaultTeamAdminRole != "" {
			return false
		}

		if scheme.DefaultTeamUserRole != "" {
			return false
		}

		if scheme.DefaultTeamGuestRole != "" {
			return false
		}
	}

	return true
}

func (scheme *Scheme) Patch(patch *SchemePatch) {
	if patch.DisplayName != nil {
		scheme.DisplayName = *patch.DisplayName
	}
	if patch.Name != nil {
		scheme.Name = *patch.Name
	}
	if patch.Description != nil {
		scheme.Description = *patch.Description
	}
}

func IsValidSchemeName(name string) bool {
	re := regexp.MustCompile(fmt.Sprintf("^[a-z0-9_]{2,%d}$", SchemeNameMaxLength))
	return re.MatchString(name)
}
