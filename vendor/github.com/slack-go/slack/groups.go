package slack

import (
	"context"
	"net/url"
	"strconv"
)

// Group contains all the information for a group
type Group struct {
	GroupConversation
	IsGroup bool `json:"is_group"`
}

type groupResponseFull struct {
	Group          Group   `json:"group"`
	Groups         []Group `json:"groups"`
	Purpose        string  `json:"purpose"`
	Topic          string  `json:"topic"`
	NotInGroup     bool    `json:"not_in_group"`
	NoOp           bool    `json:"no_op"`
	AlreadyClosed  bool    `json:"already_closed"`
	AlreadyOpen    bool    `json:"already_open"`
	AlreadyInGroup bool    `json:"already_in_group"`
	Channel        Channel `json:"channel"`
	History
	SlackResponse
}

func (api *Client) groupRequest(ctx context.Context, path string, values url.Values) (*groupResponseFull, error) {
	response := &groupResponseFull{}
	err := api.postMethod(ctx, path, values, response)
	if err != nil {
		return nil, err
	}

	return response, response.Err()
}

// ArchiveGroup archives a private group
func (api *Client) ArchiveGroup(group string) error {
	return api.ArchiveGroupContext(context.Background(), group)
}

// ArchiveGroupContext archives a private group
func (api *Client) ArchiveGroupContext(ctx context.Context, group string) error {
	values := url.Values{
		"token":   {api.token},
		"channel": {group},
	}

	_, err := api.groupRequest(ctx, "groups.archive", values)
	return err
}

// UnarchiveGroup unarchives a private group
func (api *Client) UnarchiveGroup(group string) error {
	return api.UnarchiveGroupContext(context.Background(), group)
}

// UnarchiveGroupContext unarchives a private group
func (api *Client) UnarchiveGroupContext(ctx context.Context, group string) error {
	values := url.Values{
		"token":   {api.token},
		"channel": {group},
	}

	_, err := api.groupRequest(ctx, "groups.unarchive", values)
	return err
}

// CreateGroup creates a private group
func (api *Client) CreateGroup(group string) (*Group, error) {
	return api.CreateGroupContext(context.Background(), group)
}

// CreateGroupContext creates a private group
func (api *Client) CreateGroupContext(ctx context.Context, group string) (*Group, error) {
	values := url.Values{
		"token": {api.token},
		"name":  {group},
	}

	response, err := api.groupRequest(ctx, "groups.create", values)
	if err != nil {
		return nil, err
	}
	return &response.Group, nil
}

// CreateChildGroup creates a new private group archiving the old one
// This method takes an existing private group and performs the following steps:
//   1. Renames the existing group (from "example" to "example-archived").
//   2. Archives the existing group.
//   3. Creates a new group with the name of the existing group.
//   4. Adds all members of the existing group to the new group.
func (api *Client) CreateChildGroup(group string) (*Group, error) {
	return api.CreateChildGroupContext(context.Background(), group)
}

// CreateChildGroupContext creates a new private group archiving the old one with a custom context
// For more information see CreateChildGroup
func (api *Client) CreateChildGroupContext(ctx context.Context, group string) (*Group, error) {
	values := url.Values{
		"token":   {api.token},
		"channel": {group},
	}

	response, err := api.groupRequest(ctx, "groups.createChild", values)
	if err != nil {
		return nil, err
	}
	return &response.Group, nil
}

// GetGroupHistory fetches all the history for a private group
func (api *Client) GetGroupHistory(group string, params HistoryParameters) (*History, error) {
	return api.GetGroupHistoryContext(context.Background(), group, params)
}

// GetGroupHistoryContext fetches all the history for a private group with a custom context
func (api *Client) GetGroupHistoryContext(ctx context.Context, group string, params HistoryParameters) (*History, error) {
	values := url.Values{
		"token":   {api.token},
		"channel": {group},
	}
	if params.Latest != DEFAULT_HISTORY_LATEST {
		values.Add("latest", params.Latest)
	}
	if params.Oldest != DEFAULT_HISTORY_OLDEST {
		values.Add("oldest", params.Oldest)
	}
	if params.Count != DEFAULT_HISTORY_COUNT {
		values.Add("count", strconv.Itoa(params.Count))
	}
	if params.Inclusive != DEFAULT_HISTORY_INCLUSIVE {
		if params.Inclusive {
			values.Add("inclusive", "1")
		} else {
			values.Add("inclusive", "0")
		}
	}
	if params.Unreads != DEFAULT_HISTORY_UNREADS {
		if params.Unreads {
			values.Add("unreads", "1")
		} else {
			values.Add("unreads", "0")
		}
	}

	response, err := api.groupRequest(ctx, "groups.history", values)
	if err != nil {
		return nil, err
	}
	return &response.History, nil
}

// InviteUserToGroup invites a specific user to a private group
func (api *Client) InviteUserToGroup(group, user string) (*Group, bool, error) {
	return api.InviteUserToGroupContext(context.Background(), group, user)
}

// InviteUserToGroupContext invites a specific user to a private group with a custom context
func (api *Client) InviteUserToGroupContext(ctx context.Context, group, user string) (*Group, bool, error) {
	values := url.Values{
		"token":   {api.token},
		"channel": {group},
		"user":    {user},
	}

	response, err := api.groupRequest(ctx, "groups.invite", values)
	if err != nil {
		return nil, false, err
	}
	return &response.Group, response.AlreadyInGroup, nil
}

// LeaveGroup makes authenticated user leave the group
func (api *Client) LeaveGroup(group string) error {
	return api.LeaveGroupContext(context.Background(), group)
}

// LeaveGroupContext makes authenticated user leave the group with a custom context
func (api *Client) LeaveGroupContext(ctx context.Context, group string) (err error) {
	values := url.Values{
		"token":   {api.token},
		"channel": {group},
	}

	_, err = api.groupRequest(ctx, "groups.leave", values)
	return err
}

// KickUserFromGroup kicks a user from a group
func (api *Client) KickUserFromGroup(group, user string) error {
	return api.KickUserFromGroupContext(context.Background(), group, user)
}

// KickUserFromGroupContext kicks a user from a group with a custom context
func (api *Client) KickUserFromGroupContext(ctx context.Context, group, user string) (err error) {
	values := url.Values{
		"token":   {api.token},
		"channel": {group},
		"user":    {user},
	}

	_, err = api.groupRequest(ctx, "groups.kick", values)
	return err
}

// GetGroups retrieves all groups
func (api *Client) GetGroups(excludeArchived bool) ([]Group, error) {
	return api.GetGroupsContext(context.Background(), excludeArchived)
}

// GetGroupsContext retrieves all groups with a custom context
func (api *Client) GetGroupsContext(ctx context.Context, excludeArchived bool) ([]Group, error) {
	values := url.Values{
		"token": {api.token},
	}
	if excludeArchived {
		values.Add("exclude_archived", "1")
	}

	response, err := api.groupRequest(ctx, "groups.list", values)
	if err != nil {
		return nil, err
	}
	return response.Groups, nil
}

// GetGroupInfo retrieves the given group
func (api *Client) GetGroupInfo(group string) (*Group, error) {
	return api.GetGroupInfoContext(context.Background(), group)
}

// GetGroupInfoContext retrieves the given group with a custom context
func (api *Client) GetGroupInfoContext(ctx context.Context, group string) (*Group, error) {
	values := url.Values{
		"token":          {api.token},
		"channel":        {group},
		"include_locale": {strconv.FormatBool(true)},
	}

	response, err := api.groupRequest(ctx, "groups.info", values)
	if err != nil {
		return nil, err
	}
	return &response.Group, nil
}

// SetGroupReadMark sets the read mark on a private group
// Clients should try to avoid making this call too often. When needing to mark a read position, a client should set a
// timer before making the call. In this way, any further updates needed during the timeout will not generate extra
// calls (just one per channel). This is useful for when reading scroll-back history, or following a busy live
// channel. A timeout of 5 seconds is a good starting point. Be sure to flush these calls on shutdown/logout.
func (api *Client) SetGroupReadMark(group, ts string) error {
	return api.SetGroupReadMarkContext(context.Background(), group, ts)
}

// SetGroupReadMarkContext sets the read mark on a private group with a custom context
// For more details see SetGroupReadMark
func (api *Client) SetGroupReadMarkContext(ctx context.Context, group, ts string) (err error) {
	values := url.Values{
		"token":   {api.token},
		"channel": {group},
		"ts":      {ts},
	}

	_, err = api.groupRequest(ctx, "groups.mark", values)
	return err
}

// OpenGroup opens a private group
func (api *Client) OpenGroup(group string) (bool, bool, error) {
	return api.OpenGroupContext(context.Background(), group)
}

// OpenGroupContext opens a private group with a custom context
func (api *Client) OpenGroupContext(ctx context.Context, group string) (bool, bool, error) {
	values := url.Values{
		"token":   {api.token},
		"channel": {group},
	}

	response, err := api.groupRequest(ctx, "groups.open", values)
	if err != nil {
		return false, false, err
	}
	return response.NoOp, response.AlreadyOpen, nil
}

// RenameGroup renames a group
// XXX: They return a channel, not a group. What is this crap? :(
// Inconsistent api it seems.
func (api *Client) RenameGroup(group, name string) (*Channel, error) {
	return api.RenameGroupContext(context.Background(), group, name)
}

// RenameGroupContext renames a group with a custom context
func (api *Client) RenameGroupContext(ctx context.Context, group, name string) (*Channel, error) {
	values := url.Values{
		"token":   {api.token},
		"channel": {group},
		"name":    {name},
	}

	// XXX: the created entry in this call returns a string instead of a number
	// so I may have to do some workaround to solve it.
	response, err := api.groupRequest(ctx, "groups.rename", values)
	if err != nil {
		return nil, err
	}
	return &response.Channel, nil
}

// SetGroupPurpose sets the group purpose
func (api *Client) SetGroupPurpose(group, purpose string) (string, error) {
	return api.SetGroupPurposeContext(context.Background(), group, purpose)
}

// SetGroupPurposeContext sets the group purpose with a custom context
func (api *Client) SetGroupPurposeContext(ctx context.Context, group, purpose string) (string, error) {
	values := url.Values{
		"token":   {api.token},
		"channel": {group},
		"purpose": {purpose},
	}

	response, err := api.groupRequest(ctx, "groups.setPurpose", values)
	if err != nil {
		return "", err
	}
	return response.Purpose, nil
}

// SetGroupTopic sets the group topic
func (api *Client) SetGroupTopic(group, topic string) (string, error) {
	return api.SetGroupTopicContext(context.Background(), group, topic)
}

// SetGroupTopicContext sets the group topic with a custom context
func (api *Client) SetGroupTopicContext(ctx context.Context, group, topic string) (string, error) {
	values := url.Values{
		"token":   {api.token},
		"channel": {group},
		"topic":   {topic},
	}

	response, err := api.groupRequest(ctx, "groups.setTopic", values)
	if err != nil {
		return "", err
	}
	return response.Topic, nil
}

// GetGroupReplies gets an entire thread (a message plus all the messages in reply to it).
// see https://api.slack.com/methods/groups.replies
func (api *Client) GetGroupReplies(channelID, thread_ts string) ([]Message, error) {
	return api.GetGroupRepliesContext(context.Background(), channelID, thread_ts)
}

// GetGroupRepliesContext gets an entire thread (a message plus all the messages in reply to it) with a custom context
// see https://api.slack.com/methods/groups.replies
func (api *Client) GetGroupRepliesContext(ctx context.Context, channelID, thread_ts string) ([]Message, error) {
	values := url.Values{
		"token":     {api.token},
		"channel":   {channelID},
		"thread_ts": {thread_ts},
	}
	response, err := api.groupRequest(ctx, "groups.replies", values)
	if err != nil {
		return nil, err
	}
	return response.History.Messages, nil
}
