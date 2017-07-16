package slack

import (
	"context"
	"errors"
	"net/url"
	"strings"
)

// UserGroup contains all the information of a user group
type UserGroup struct {
	ID          string         `json:"id"`
	TeamID      string         `json:"team_id"`
	IsUserGroup bool           `json:"is_usergroup"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Handle      string         `json:"handle"`
	IsExternal  bool           `json:"is_external"`
	DateCreate  JSONTime       `json:"date_create"`
	DateUpdate  JSONTime       `json:"date_update"`
	DateDelete  JSONTime       `json:"date_delete"`
	AutoType    string         `json:"auto_type"`
	CreatedBy   string         `json:"created_by"`
	UpdatedBy   string         `json:"updated_by"`
	DeletedBy   string         `json:"deleted_by"`
	Prefs       UserGroupPrefs `json:"prefs"`
	UserCount   int            `json:"user_count"`
}

// UserGroupPrefs contains default channels and groups (private channels)
type UserGroupPrefs struct {
	Channels []string `json:"channels"`
	Groups   []string `json:"groups"`
}

type userGroupResponseFull struct {
	UserGroups []UserGroup `json:"usergroups"`
	UserGroup  UserGroup   `json:"usergroup"`
	Users      []string    `json:"users"`
	SlackResponse
}

func userGroupRequest(ctx context.Context, path string, values url.Values, debug bool) (*userGroupResponseFull, error) {
	response := &userGroupResponseFull{}
	err := post(ctx, path, values, response, debug)
	if err != nil {
		return nil, err
	}
	if !response.Ok {
		return nil, errors.New(response.Error)
	}
	return response, nil
}

// CreateUserGroup creates a new user group
func (api *Client) CreateUserGroup(userGroup UserGroup) (UserGroup, error) {
	return api.CreateUserGroupContext(context.Background(), userGroup)
}

// CreateUserGroupContext creates a new user group with a custom context
func (api *Client) CreateUserGroupContext(ctx context.Context, userGroup UserGroup) (UserGroup, error) {
	values := url.Values{
		"token": {api.config.token},
		"name":  {userGroup.Name},
	}

	if userGroup.Handle != "" {
		values["handle"] = []string{userGroup.Handle}
	}

	if userGroup.Description != "" {
		values["description"] = []string{userGroup.Description}
	}

	if len(userGroup.Prefs.Channels) > 0 {
		values["channels"] = []string{strings.Join(userGroup.Prefs.Channels, ",")}
	}

	response, err := userGroupRequest(ctx, "usergroups.create", values, api.debug)
	if err != nil {
		return UserGroup{}, err
	}
	return response.UserGroup, nil
}

// DisableUserGroup disables an existing user group
func (api *Client) DisableUserGroup(userGroup string) (UserGroup, error) {
	return api.DisableUserGroupContext(context.Background(), userGroup)
}

// DisableUserGroupContext disables an existing user group with a custom context
func (api *Client) DisableUserGroupContext(ctx context.Context, userGroup string) (UserGroup, error) {
	values := url.Values{
		"token":     {api.config.token},
		"usergroup": {userGroup},
	}

	response, err := userGroupRequest(ctx, "usergroups.disable", values, api.debug)
	if err != nil {
		return UserGroup{}, err
	}
	return response.UserGroup, nil
}

// EnableUserGroup enables an existing user group
func (api *Client) EnableUserGroup(userGroup string) (UserGroup, error) {
	return api.EnableUserGroupContext(context.Background(), userGroup)
}

// EnableUserGroupContext enables an existing user group with a custom context
func (api *Client) EnableUserGroupContext(ctx context.Context, userGroup string) (UserGroup, error) {
	values := url.Values{
		"token":     {api.config.token},
		"usergroup": {userGroup},
	}

	response, err := userGroupRequest(ctx, "usergroups.enable", values, api.debug)
	if err != nil {
		return UserGroup{}, err
	}
	return response.UserGroup, nil
}

// GetUserGroups returns a list of user groups for the team
func (api *Client) GetUserGroups() ([]UserGroup, error) {
	return api.GetUserGroupsContext(context.Background())
}

// GetUserGroupsContext returns a list of user groups for the team with a custom context
func (api *Client) GetUserGroupsContext(ctx context.Context) ([]UserGroup, error) {
	values := url.Values{
		"token": {api.config.token},
	}

	response, err := userGroupRequest(ctx, "usergroups.list", values, api.debug)
	if err != nil {
		return nil, err
	}
	return response.UserGroups, nil
}

// UpdateUserGroup will update an existing user group
func (api *Client) UpdateUserGroup(userGroup UserGroup) (UserGroup, error) {
	return api.UpdateUserGroupContext(context.Background(), userGroup)
}

// UpdateUserGroupContext will update an existing user group with a custom context
func (api *Client) UpdateUserGroupContext(ctx context.Context, userGroup UserGroup) (UserGroup, error) {
	values := url.Values{
		"token":     {api.config.token},
		"usergroup": {userGroup.ID},
	}

	if userGroup.Name != "" {
		values["name"] = []string{userGroup.Name}
	}

	if userGroup.Handle != "" {
		values["handle"] = []string{userGroup.Handle}
	}

	if userGroup.Description != "" {
		values["description"] = []string{userGroup.Description}
	}

	response, err := userGroupRequest(ctx, "usergroups.update", values, api.debug)
	if err != nil {
		return UserGroup{}, err
	}
	return response.UserGroup, nil
}

// GetUserGroupMembers will retrieve the current list of users in a group
func (api *Client) GetUserGroupMembers(userGroup string) ([]string, error) {
	return api.GetUserGroupMembersContext(context.Background(), userGroup)
}

// GetUserGroupMembersContext will retrieve the current list of users in a group with a custom context
func (api *Client) GetUserGroupMembersContext(ctx context.Context, userGroup string) ([]string, error) {
	values := url.Values{
		"token":     {api.config.token},
		"usergroup": {userGroup},
	}

	response, err := userGroupRequest(ctx, "usergroups.users.list", values, api.debug)
	if err != nil {
		return []string{}, err
	}
	return response.Users, nil
}

// UpdateUserGroupMembers will update the members of an existing user group
func (api *Client) UpdateUserGroupMembers(userGroup string, members string) (UserGroup, error) {
	return api.UpdateUserGroupMembersContext(context.Background(), userGroup, members)
}

// UpdateUserGroupMembersContext will update the members of an existing user group with a custom context
func (api *Client) UpdateUserGroupMembersContext(ctx context.Context, userGroup string, members string) (UserGroup, error) {
	values := url.Values{
		"token":     {api.config.token},
		"usergroup": {userGroup},
		"users":     {members},
	}

	response, err := userGroupRequest(ctx, "usergroups.users.update", values, api.debug)
	if err != nil {
		return UserGroup{}, err
	}
	return response.UserGroup, nil
}
