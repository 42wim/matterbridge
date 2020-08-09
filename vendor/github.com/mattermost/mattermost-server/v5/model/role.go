// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package model

import (
	"encoding/json"
	"io"
	"strings"
)

var BuiltInSchemeManagedRoleIDs []string

func init() {
	BuiltInSchemeManagedRoleIDs = []string{
		SYSTEM_GUEST_ROLE_ID,
		SYSTEM_USER_ROLE_ID,
		SYSTEM_ADMIN_ROLE_ID,
		SYSTEM_POST_ALL_ROLE_ID,
		SYSTEM_POST_ALL_PUBLIC_ROLE_ID,
		SYSTEM_USER_ACCESS_TOKEN_ROLE_ID,

		TEAM_GUEST_ROLE_ID,
		TEAM_USER_ROLE_ID,
		TEAM_ADMIN_ROLE_ID,
		TEAM_POST_ALL_ROLE_ID,
		TEAM_POST_ALL_PUBLIC_ROLE_ID,

		CHANNEL_GUEST_ROLE_ID,
		CHANNEL_USER_ROLE_ID,
		CHANNEL_ADMIN_ROLE_ID,
	}
}

type RoleType string
type RoleScope string

const (
	SYSTEM_GUEST_ROLE_ID             = "system_guest"
	SYSTEM_USER_ROLE_ID              = "system_user"
	SYSTEM_ADMIN_ROLE_ID             = "system_admin"
	SYSTEM_POST_ALL_ROLE_ID          = "system_post_all"
	SYSTEM_POST_ALL_PUBLIC_ROLE_ID   = "system_post_all_public"
	SYSTEM_USER_ACCESS_TOKEN_ROLE_ID = "system_user_access_token"

	TEAM_GUEST_ROLE_ID           = "team_guest"
	TEAM_USER_ROLE_ID            = "team_user"
	TEAM_ADMIN_ROLE_ID           = "team_admin"
	TEAM_POST_ALL_ROLE_ID        = "team_post_all"
	TEAM_POST_ALL_PUBLIC_ROLE_ID = "team_post_all_public"

	CHANNEL_GUEST_ROLE_ID = "channel_guest"
	CHANNEL_USER_ROLE_ID  = "channel_user"
	CHANNEL_ADMIN_ROLE_ID = "channel_admin"

	ROLE_NAME_MAX_LENGTH         = 64
	ROLE_DISPLAY_NAME_MAX_LENGTH = 128
	ROLE_DESCRIPTION_MAX_LENGTH  = 1024

	RoleScopeSystem  RoleScope = "System"
	RoleScopeTeam    RoleScope = "Team"
	RoleScopeChannel RoleScope = "Channel"

	RoleTypeGuest RoleType = "Guest"
	RoleTypeUser  RoleType = "User"
	RoleTypeAdmin RoleType = "Admin"
)

type Role struct {
	Id            string   `json:"id"`
	Name          string   `json:"name"`
	DisplayName   string   `json:"display_name"`
	Description   string   `json:"description"`
	CreateAt      int64    `json:"create_at"`
	UpdateAt      int64    `json:"update_at"`
	DeleteAt      int64    `json:"delete_at"`
	Permissions   []string `json:"permissions"`
	SchemeManaged bool     `json:"scheme_managed"`
	BuiltIn       bool     `json:"built_in"`
}

type RolePatch struct {
	Permissions *[]string `json:"permissions"`
}

type RolePermissions struct {
	RoleID      string
	Permissions []string
}

func (r *Role) ToJson() string {
	b, _ := json.Marshal(r)
	return string(b)
}

func RoleFromJson(data io.Reader) *Role {
	var r *Role
	json.NewDecoder(data).Decode(&r)
	return r
}

func RoleListToJson(r []*Role) string {
	b, _ := json.Marshal(r)
	return string(b)
}

func RoleListFromJson(data io.Reader) []*Role {
	var roles []*Role
	json.NewDecoder(data).Decode(&roles)
	return roles
}

func (r *RolePatch) ToJson() string {
	b, _ := json.Marshal(r)
	return string(b)
}

func RolePatchFromJson(data io.Reader) *RolePatch {
	var rolePatch *RolePatch
	json.NewDecoder(data).Decode(&rolePatch)
	return rolePatch
}

func (r *Role) Patch(patch *RolePatch) {
	if patch.Permissions != nil {
		r.Permissions = *patch.Permissions
	}
}

// MergeChannelHigherScopedPermissions is meant to be invoked on a channel scheme's role and merges the higher-scoped
// channel role's permissions.
func (r *Role) MergeChannelHigherScopedPermissions(higherScopedPermissions *RolePermissions) {
	mergedPermissions := []string{}

	higherScopedPermissionsMap := AsStringBoolMap(higherScopedPermissions.Permissions)
	rolePermissionsMap := AsStringBoolMap(r.Permissions)

	for _, cp := range ALL_PERMISSIONS {
		if cp.Scope != PERMISSION_SCOPE_CHANNEL {
			continue
		}

		_, presentOnHigherScope := higherScopedPermissionsMap[cp.Id]

		// For the channel admin role always look to the higher scope to determine if the role has ther permission.
		// The channel admin is a special case because they're not part of the UI to be "channel moderated", only
		// channel members and channel guests are.
		if higherScopedPermissions.RoleID == CHANNEL_ADMIN_ROLE_ID && presentOnHigherScope {
			mergedPermissions = append(mergedPermissions, cp.Id)
			continue
		}

		_, permissionIsModerated := CHANNEL_MODERATED_PERMISSIONS_MAP[cp.Id]
		if permissionIsModerated {
			_, presentOnRole := rolePermissionsMap[cp.Id]
			if presentOnRole && presentOnHigherScope {
				mergedPermissions = append(mergedPermissions, cp.Id)
			}
		} else {
			if presentOnHigherScope {
				mergedPermissions = append(mergedPermissions, cp.Id)
			}
		}
	}

	r.Permissions = mergedPermissions
}

// Returns an array of permissions that are in either role.Permissions
// or patch.Permissions, but not both.
func PermissionsChangedByPatch(role *Role, patch *RolePatch) []string {
	var result []string

	if patch.Permissions == nil {
		return result
	}

	roleMap := make(map[string]bool)
	patchMap := make(map[string]bool)

	for _, permission := range role.Permissions {
		roleMap[permission] = true
	}

	for _, permission := range *patch.Permissions {
		patchMap[permission] = true
	}

	for _, permission := range role.Permissions {
		if !patchMap[permission] {
			result = append(result, permission)
		}
	}

	for _, permission := range *patch.Permissions {
		if !roleMap[permission] {
			result = append(result, permission)
		}
	}

	return result
}

func ChannelModeratedPermissionsChangedByPatch(role *Role, patch *RolePatch) []string {
	var result []string

	if role == nil {
		return result
	}

	if patch.Permissions == nil {
		return result
	}

	roleMap := make(map[string]bool)
	patchMap := make(map[string]bool)

	for _, permission := range role.Permissions {
		if channelModeratedPermissionName, found := CHANNEL_MODERATED_PERMISSIONS_MAP[permission]; found {
			roleMap[channelModeratedPermissionName] = true
		}
	}

	for _, permission := range *patch.Permissions {
		if channelModeratedPermissionName, found := CHANNEL_MODERATED_PERMISSIONS_MAP[permission]; found {
			patchMap[channelModeratedPermissionName] = true
		}
	}

	for permissionKey := range roleMap {
		if !patchMap[permissionKey] {
			result = append(result, permissionKey)
		}
	}

	for permissionKey := range patchMap {
		if !roleMap[permissionKey] {
			result = append(result, permissionKey)
		}
	}

	return result
}

// GetChannelModeratedPermissions returns a map of channel moderated permissions that the role has access to
func (r *Role) GetChannelModeratedPermissions(channelType string) map[string]bool {
	moderatedPermissions := make(map[string]bool)
	for _, permission := range r.Permissions {
		if _, found := CHANNEL_MODERATED_PERMISSIONS_MAP[permission]; !found {
			continue
		}

		for moderated, moderatedPermissionValue := range CHANNEL_MODERATED_PERMISSIONS_MAP {
			// the moderated permission has already been found to be true so skip this iteration
			if moderatedPermissions[moderatedPermissionValue] {
				continue
			}

			if moderated == permission {
				// Special case where the channel moderated permission for `manage_members` is different depending on whether the channel is private or public
				if moderated == PERMISSION_MANAGE_PUBLIC_CHANNEL_MEMBERS.Id || moderated == PERMISSION_MANAGE_PRIVATE_CHANNEL_MEMBERS.Id {
					canManagePublic := channelType == CHANNEL_OPEN && moderated == PERMISSION_MANAGE_PUBLIC_CHANNEL_MEMBERS.Id
					canManagePrivate := channelType == CHANNEL_PRIVATE && moderated == PERMISSION_MANAGE_PRIVATE_CHANNEL_MEMBERS.Id
					moderatedPermissions[moderatedPermissionValue] = canManagePublic || canManagePrivate
				} else {
					moderatedPermissions[moderatedPermissionValue] = true
				}
			}
		}
	}

	return moderatedPermissions
}

// RolePatchFromChannelModerationsPatch Creates and returns a RolePatch based on a slice of ChannelModerationPatchs, roleName is expected to be either "members" or "guests".
func (r *Role) RolePatchFromChannelModerationsPatch(channelModerationsPatch []*ChannelModerationPatch, roleName string) *RolePatch {
	permissionsToAddToPatch := make(map[string]bool)

	// Iterate through the list of existing permissions on the role and append permissions that we want to keep.
	for _, permission := range r.Permissions {
		// Permission is not moderated so dont add it to the patch and skip the channelModerationsPatch
		if _, isModerated := CHANNEL_MODERATED_PERMISSIONS_MAP[permission]; !isModerated {
			continue
		}

		permissionEnabled := true
		// Check if permission has a matching moderated permission name inside the channel moderation patch
		for _, channelModerationPatch := range channelModerationsPatch {
			if *channelModerationPatch.Name == CHANNEL_MODERATED_PERMISSIONS_MAP[permission] {
				// Permission key exists in patch with a value of false so skip over it
				if roleName == "members" {
					if channelModerationPatch.Roles.Members != nil && !*channelModerationPatch.Roles.Members {
						permissionEnabled = false
					}
				} else if roleName == "guests" {
					if channelModerationPatch.Roles.Guests != nil && !*channelModerationPatch.Roles.Guests {
						permissionEnabled = false
					}
				}
			}
		}

		if permissionEnabled {
			permissionsToAddToPatch[permission] = true
		}
	}

	// Iterate through the patch and add any permissions that dont already exist on the role
	for _, channelModerationPatch := range channelModerationsPatch {
		for permission, moderatedPermissionName := range CHANNEL_MODERATED_PERMISSIONS_MAP {
			if roleName == "members" && channelModerationPatch.Roles.Members != nil && *channelModerationPatch.Roles.Members && *channelModerationPatch.Name == moderatedPermissionName {
				permissionsToAddToPatch[permission] = true
			}

			if roleName == "guests" && channelModerationPatch.Roles.Guests != nil && *channelModerationPatch.Roles.Guests && *channelModerationPatch.Name == moderatedPermissionName {
				permissionsToAddToPatch[permission] = true
			}
		}
	}

	patchPermissions := make([]string, 0, len(permissionsToAddToPatch))
	for permission := range permissionsToAddToPatch {
		patchPermissions = append(patchPermissions, permission)
	}

	return &RolePatch{Permissions: &patchPermissions}
}

func (r *Role) IsValid() bool {
	if !IsValidId(r.Id) {
		return false
	}

	return r.IsValidWithoutId()
}

func (r *Role) IsValidWithoutId() bool {
	if !IsValidRoleName(r.Name) {
		return false
	}

	if len(r.DisplayName) == 0 || len(r.DisplayName) > ROLE_DISPLAY_NAME_MAX_LENGTH {
		return false
	}

	if len(r.Description) > ROLE_DESCRIPTION_MAX_LENGTH {
		return false
	}

	for _, permission := range r.Permissions {
		permissionValidated := false
		for _, p := range ALL_PERMISSIONS {
			if permission == p.Id {
				permissionValidated = true
				break
			}
		}

		if !permissionValidated {
			return false
		}
	}

	return true
}

func IsValidRoleName(roleName string) bool {
	if len(roleName) <= 0 || len(roleName) > ROLE_NAME_MAX_LENGTH {
		return false
	}

	if strings.TrimLeft(roleName, "abcdefghijklmnopqrstuvwxyz0123456789_") != "" {
		return false
	}

	return true
}

func MakeDefaultRoles() map[string]*Role {
	roles := make(map[string]*Role)

	roles[CHANNEL_GUEST_ROLE_ID] = &Role{
		Name:        "channel_guest",
		DisplayName: "authentication.roles.channel_guest.name",
		Description: "authentication.roles.channel_guest.description",
		Permissions: []string{
			PERMISSION_READ_CHANNEL.Id,
			PERMISSION_ADD_REACTION.Id,
			PERMISSION_REMOVE_REACTION.Id,
			PERMISSION_UPLOAD_FILE.Id,
			PERMISSION_EDIT_POST.Id,
			PERMISSION_CREATE_POST.Id,
			PERMISSION_USE_CHANNEL_MENTIONS.Id,
			PERMISSION_USE_SLASH_COMMANDS.Id,
		},
		SchemeManaged: true,
		BuiltIn:       true,
	}

	roles[CHANNEL_USER_ROLE_ID] = &Role{
		Name:        "channel_user",
		DisplayName: "authentication.roles.channel_user.name",
		Description: "authentication.roles.channel_user.description",
		Permissions: []string{
			PERMISSION_READ_CHANNEL.Id,
			PERMISSION_ADD_REACTION.Id,
			PERMISSION_REMOVE_REACTION.Id,
			PERMISSION_MANAGE_PUBLIC_CHANNEL_MEMBERS.Id,
			PERMISSION_UPLOAD_FILE.Id,
			PERMISSION_GET_PUBLIC_LINK.Id,
			PERMISSION_CREATE_POST.Id,
			PERMISSION_USE_CHANNEL_MENTIONS.Id,
			PERMISSION_USE_SLASH_COMMANDS.Id,
		},
		SchemeManaged: true,
		BuiltIn:       true,
	}

	roles[CHANNEL_ADMIN_ROLE_ID] = &Role{
		Name:        "channel_admin",
		DisplayName: "authentication.roles.channel_admin.name",
		Description: "authentication.roles.channel_admin.description",
		Permissions: []string{
			PERMISSION_MANAGE_CHANNEL_ROLES.Id,
			PERMISSION_USE_GROUP_MENTIONS.Id,
		},
		SchemeManaged: true,
		BuiltIn:       true,
	}

	roles[TEAM_GUEST_ROLE_ID] = &Role{
		Name:        "team_guest",
		DisplayName: "authentication.roles.team_guest.name",
		Description: "authentication.roles.team_guest.description",
		Permissions: []string{
			PERMISSION_VIEW_TEAM.Id,
		},
		SchemeManaged: true,
		BuiltIn:       true,
	}

	roles[TEAM_USER_ROLE_ID] = &Role{
		Name:        "team_user",
		DisplayName: "authentication.roles.team_user.name",
		Description: "authentication.roles.team_user.description",
		Permissions: []string{
			PERMISSION_LIST_TEAM_CHANNELS.Id,
			PERMISSION_JOIN_PUBLIC_CHANNELS.Id,
			PERMISSION_READ_PUBLIC_CHANNEL.Id,
			PERMISSION_VIEW_TEAM.Id,
		},
		SchemeManaged: true,
		BuiltIn:       true,
	}

	roles[TEAM_POST_ALL_ROLE_ID] = &Role{
		Name:        "team_post_all",
		DisplayName: "authentication.roles.team_post_all.name",
		Description: "authentication.roles.team_post_all.description",
		Permissions: []string{
			PERMISSION_CREATE_POST.Id,
			PERMISSION_USE_CHANNEL_MENTIONS.Id,
		},
		SchemeManaged: false,
		BuiltIn:       true,
	}

	roles[TEAM_POST_ALL_PUBLIC_ROLE_ID] = &Role{
		Name:        "team_post_all_public",
		DisplayName: "authentication.roles.team_post_all_public.name",
		Description: "authentication.roles.team_post_all_public.description",
		Permissions: []string{
			PERMISSION_CREATE_POST_PUBLIC.Id,
			PERMISSION_USE_CHANNEL_MENTIONS.Id,
		},
		SchemeManaged: false,
		BuiltIn:       true,
	}

	roles[TEAM_ADMIN_ROLE_ID] = &Role{
		Name:        "team_admin",
		DisplayName: "authentication.roles.team_admin.name",
		Description: "authentication.roles.team_admin.description",
		Permissions: []string{
			PERMISSION_REMOVE_USER_FROM_TEAM.Id,
			PERMISSION_MANAGE_TEAM.Id,
			PERMISSION_IMPORT_TEAM.Id,
			PERMISSION_MANAGE_TEAM_ROLES.Id,
			PERMISSION_MANAGE_CHANNEL_ROLES.Id,
			PERMISSION_MANAGE_OTHERS_INCOMING_WEBHOOKS.Id,
			PERMISSION_MANAGE_OTHERS_OUTGOING_WEBHOOKS.Id,
			PERMISSION_MANAGE_SLASH_COMMANDS.Id,
			PERMISSION_MANAGE_OTHERS_SLASH_COMMANDS.Id,
			PERMISSION_MANAGE_INCOMING_WEBHOOKS.Id,
			PERMISSION_MANAGE_OUTGOING_WEBHOOKS.Id,
		},
		SchemeManaged: true,
		BuiltIn:       true,
	}

	roles[SYSTEM_GUEST_ROLE_ID] = &Role{
		Name:        "system_guest",
		DisplayName: "authentication.roles.global_guest.name",
		Description: "authentication.roles.global_guest.description",
		Permissions: []string{
			PERMISSION_CREATE_DIRECT_CHANNEL.Id,
			PERMISSION_CREATE_GROUP_CHANNEL.Id,
		},
		SchemeManaged: true,
		BuiltIn:       true,
	}

	roles[SYSTEM_USER_ROLE_ID] = &Role{
		Name:        "system_user",
		DisplayName: "authentication.roles.global_user.name",
		Description: "authentication.roles.global_user.description",
		Permissions: []string{
			PERMISSION_LIST_PUBLIC_TEAMS.Id,
			PERMISSION_JOIN_PUBLIC_TEAMS.Id,
			PERMISSION_CREATE_DIRECT_CHANNEL.Id,
			PERMISSION_CREATE_GROUP_CHANNEL.Id,
			PERMISSION_VIEW_MEMBERS.Id,
		},
		SchemeManaged: true,
		BuiltIn:       true,
	}

	roles[SYSTEM_POST_ALL_ROLE_ID] = &Role{
		Name:        "system_post_all",
		DisplayName: "authentication.roles.system_post_all.name",
		Description: "authentication.roles.system_post_all.description",
		Permissions: []string{
			PERMISSION_CREATE_POST.Id,
			PERMISSION_USE_CHANNEL_MENTIONS.Id,
		},
		SchemeManaged: false,
		BuiltIn:       true,
	}

	roles[SYSTEM_POST_ALL_PUBLIC_ROLE_ID] = &Role{
		Name:        "system_post_all_public",
		DisplayName: "authentication.roles.system_post_all_public.name",
		Description: "authentication.roles.system_post_all_public.description",
		Permissions: []string{
			PERMISSION_CREATE_POST_PUBLIC.Id,
			PERMISSION_USE_CHANNEL_MENTIONS.Id,
		},
		SchemeManaged: false,
		BuiltIn:       true,
	}

	roles[SYSTEM_USER_ACCESS_TOKEN_ROLE_ID] = &Role{
		Name:        "system_user_access_token",
		DisplayName: "authentication.roles.system_user_access_token.name",
		Description: "authentication.roles.system_user_access_token.description",
		Permissions: []string{
			PERMISSION_CREATE_USER_ACCESS_TOKEN.Id,
			PERMISSION_READ_USER_ACCESS_TOKEN.Id,
			PERMISSION_REVOKE_USER_ACCESS_TOKEN.Id,
		},
		SchemeManaged: false,
		BuiltIn:       true,
	}

	roles[SYSTEM_ADMIN_ROLE_ID] = &Role{
		Name:        "system_admin",
		DisplayName: "authentication.roles.global_admin.name",
		Description: "authentication.roles.global_admin.description",
		// System admins can do anything channel and team admins can do
		// plus everything members of teams and channels can do to all teams
		// and channels on the system
		Permissions: append(
			append(
				append(
					append(
						[]string{
							PERMISSION_ASSIGN_SYSTEM_ADMIN_ROLE.Id,
							PERMISSION_MANAGE_SYSTEM.Id,
							PERMISSION_MANAGE_ROLES.Id,
							PERMISSION_MANAGE_PUBLIC_CHANNEL_PROPERTIES.Id,
							PERMISSION_MANAGE_PUBLIC_CHANNEL_MEMBERS.Id,
							PERMISSION_MANAGE_PRIVATE_CHANNEL_MEMBERS.Id,
							PERMISSION_DELETE_PUBLIC_CHANNEL.Id,
							PERMISSION_CREATE_PUBLIC_CHANNEL.Id,
							PERMISSION_MANAGE_PRIVATE_CHANNEL_PROPERTIES.Id,
							PERMISSION_DELETE_PRIVATE_CHANNEL.Id,
							PERMISSION_CREATE_PRIVATE_CHANNEL.Id,
							PERMISSION_MANAGE_SYSTEM_WIDE_OAUTH.Id,
							PERMISSION_MANAGE_OTHERS_INCOMING_WEBHOOKS.Id,
							PERMISSION_MANAGE_OTHERS_OUTGOING_WEBHOOKS.Id,
							PERMISSION_EDIT_OTHER_USERS.Id,
							PERMISSION_EDIT_OTHERS_POSTS.Id,
							PERMISSION_MANAGE_OAUTH.Id,
							PERMISSION_INVITE_USER.Id,
							PERMISSION_INVITE_GUEST.Id,
							PERMISSION_PROMOTE_GUEST.Id,
							PERMISSION_DEMOTE_TO_GUEST.Id,
							PERMISSION_DELETE_POST.Id,
							PERMISSION_DELETE_OTHERS_POSTS.Id,
							PERMISSION_CREATE_TEAM.Id,
							PERMISSION_ADD_USER_TO_TEAM.Id,
							PERMISSION_LIST_USERS_WITHOUT_TEAM.Id,
							PERMISSION_MANAGE_JOBS.Id,
							PERMISSION_CREATE_POST_PUBLIC.Id,
							PERMISSION_CREATE_POST_EPHEMERAL.Id,
							PERMISSION_CREATE_USER_ACCESS_TOKEN.Id,
							PERMISSION_READ_USER_ACCESS_TOKEN.Id,
							PERMISSION_REVOKE_USER_ACCESS_TOKEN.Id,
							PERMISSION_CREATE_BOT.Id,
							PERMISSION_READ_BOTS.Id,
							PERMISSION_READ_OTHERS_BOTS.Id,
							PERMISSION_MANAGE_BOTS.Id,
							PERMISSION_MANAGE_OTHERS_BOTS.Id,
							PERMISSION_REMOVE_OTHERS_REACTIONS.Id,
							PERMISSION_LIST_PRIVATE_TEAMS.Id,
							PERMISSION_JOIN_PRIVATE_TEAMS.Id,
							PERMISSION_VIEW_MEMBERS.Id,
						},
						roles[TEAM_USER_ROLE_ID].Permissions...,
					),
					roles[CHANNEL_USER_ROLE_ID].Permissions...,
				),
				roles[TEAM_ADMIN_ROLE_ID].Permissions...,
			),
			roles[CHANNEL_ADMIN_ROLE_ID].Permissions...,
		),
		SchemeManaged: true,
		BuiltIn:       true,
	}

	return roles
}
