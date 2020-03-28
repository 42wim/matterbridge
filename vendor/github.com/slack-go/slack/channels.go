package slack

import (
	"context"
	"net/url"
	"strconv"
	"time"
)

type channelResponseFull struct {
	Channel      Channel   `json:"channel"`
	Channels     []Channel `json:"channels"`
	Purpose      string    `json:"purpose"`
	Topic        string    `json:"topic"`
	NotInChannel bool      `json:"not_in_channel"`
	History
	SlackResponse
	Metadata ResponseMetadata `json:"response_metadata"`
}

// Channel contains information about the channel
type Channel struct {
	GroupConversation
	IsChannel bool   `json:"is_channel"`
	IsGeneral bool   `json:"is_general"`
	IsMember  bool   `json:"is_member"`
	Locale    string `json:"locale"`
}

func (api *Client) channelRequest(ctx context.Context, path string, values url.Values) (*channelResponseFull, error) {
	response := &channelResponseFull{}
	err := postForm(ctx, api.httpclient, api.endpoint+path, values, response, api)
	if err != nil {
		return nil, err
	}

	return response, response.Err()
}

// GetChannelsOption option provided when getting channels.
type GetChannelsOption func(*ChannelPagination) error

// GetChannelsOptionExcludeMembers excludes the members collection from each channel.
func GetChannelsOptionExcludeMembers() GetChannelsOption {
	return func(p *ChannelPagination) error {
		p.excludeMembers = true
		return nil
	}
}

// GetChannelsOptionExcludeArchived excludes archived channels from results.
func GetChannelsOptionExcludeArchived() GetChannelsOption {
	return func(p *ChannelPagination) error {
		p.excludeArchived = true
		return nil
	}
}

// ArchiveChannel archives the given channel
// see https://api.slack.com/methods/channels.archive
func (api *Client) ArchiveChannel(channelID string) error {
	return api.ArchiveChannelContext(context.Background(), channelID)
}

// ArchiveChannelContext archives the given channel with a custom context
// see https://api.slack.com/methods/channels.archive
func (api *Client) ArchiveChannelContext(ctx context.Context, channelID string) (err error) {
	values := url.Values{
		"token":   {api.token},
		"channel": {channelID},
	}

	_, err = api.channelRequest(ctx, "channels.archive", values)
	return err
}

// UnarchiveChannel unarchives the given channel
// see https://api.slack.com/methods/channels.unarchive
func (api *Client) UnarchiveChannel(channelID string) error {
	return api.UnarchiveChannelContext(context.Background(), channelID)
}

// UnarchiveChannelContext unarchives the given channel with a custom context
// see https://api.slack.com/methods/channels.unarchive
func (api *Client) UnarchiveChannelContext(ctx context.Context, channelID string) (err error) {
	values := url.Values{
		"token":   {api.token},
		"channel": {channelID},
	}

	_, err = api.channelRequest(ctx, "channels.unarchive", values)
	return err
}

// CreateChannel creates a channel with the given name and returns a *Channel
// see https://api.slack.com/methods/channels.create
func (api *Client) CreateChannel(channelName string) (*Channel, error) {
	return api.CreateChannelContext(context.Background(), channelName)
}

// CreateChannelContext creates a channel with the given name and returns a *Channel with a custom context
// see https://api.slack.com/methods/channels.create
func (api *Client) CreateChannelContext(ctx context.Context, channelName string) (*Channel, error) {
	values := url.Values{
		"token": {api.token},
		"name":  {channelName},
	}

	response, err := api.channelRequest(ctx, "channels.create", values)
	if err != nil {
		return nil, err
	}
	return &response.Channel, nil
}

// GetChannelHistory retrieves the channel history
// see https://api.slack.com/methods/channels.history
func (api *Client) GetChannelHistory(channelID string, params HistoryParameters) (*History, error) {
	return api.GetChannelHistoryContext(context.Background(), channelID, params)
}

// GetChannelHistoryContext retrieves the channel history with a custom context
// see https://api.slack.com/methods/channels.history
func (api *Client) GetChannelHistoryContext(ctx context.Context, channelID string, params HistoryParameters) (*History, error) {
	values := url.Values{
		"token":   {api.token},
		"channel": {channelID},
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

	response, err := api.channelRequest(ctx, "channels.history", values)
	if err != nil {
		return nil, err
	}
	return &response.History, nil
}

// GetChannelInfo retrieves the given channel
// see https://api.slack.com/methods/channels.info
func (api *Client) GetChannelInfo(channelID string) (*Channel, error) {
	return api.GetChannelInfoContext(context.Background(), channelID)
}

// GetChannelInfoContext retrieves the given channel with a custom context
// see https://api.slack.com/methods/channels.info
func (api *Client) GetChannelInfoContext(ctx context.Context, channelID string) (*Channel, error) {
	values := url.Values{
		"token":          {api.token},
		"channel":        {channelID},
		"include_locale": {strconv.FormatBool(true)},
	}

	response, err := api.channelRequest(ctx, "channels.info", values)
	if err != nil {
		return nil, err
	}
	return &response.Channel, nil
}

// InviteUserToChannel invites a user to a given channel and returns a *Channel
// see https://api.slack.com/methods/channels.invite
func (api *Client) InviteUserToChannel(channelID, user string) (*Channel, error) {
	return api.InviteUserToChannelContext(context.Background(), channelID, user)
}

// InviteUserToChannelContext invites a user to a given channel and returns a *Channel with a custom context
// see https://api.slack.com/methods/channels.invite
func (api *Client) InviteUserToChannelContext(ctx context.Context, channelID, user string) (*Channel, error) {
	values := url.Values{
		"token":   {api.token},
		"channel": {channelID},
		"user":    {user},
	}

	response, err := api.channelRequest(ctx, "channels.invite", values)
	if err != nil {
		return nil, err
	}
	return &response.Channel, nil
}

// JoinChannel joins the currently authenticated user to a channel
// see https://api.slack.com/methods/channels.join
func (api *Client) JoinChannel(channelName string) (*Channel, error) {
	return api.JoinChannelContext(context.Background(), channelName)
}

// JoinChannelContext joins the currently authenticated user to a channel with a custom context
// see https://api.slack.com/methods/channels.join
func (api *Client) JoinChannelContext(ctx context.Context, channelName string) (*Channel, error) {
	values := url.Values{
		"token": {api.token},
		"name":  {channelName},
	}

	response, err := api.channelRequest(ctx, "channels.join", values)
	if err != nil {
		return nil, err
	}
	return &response.Channel, nil
}

// LeaveChannel makes the authenticated user leave the given channel
// see https://api.slack.com/methods/channels.leave
func (api *Client) LeaveChannel(channelID string) (bool, error) {
	return api.LeaveChannelContext(context.Background(), channelID)
}

// LeaveChannelContext makes the authenticated user leave the given channel with a custom context
// see https://api.slack.com/methods/channels.leave
func (api *Client) LeaveChannelContext(ctx context.Context, channelID string) (bool, error) {
	values := url.Values{
		"token":   {api.token},
		"channel": {channelID},
	}

	response, err := api.channelRequest(ctx, "channels.leave", values)
	if err != nil {
		return false, err
	}

	return response.NotInChannel, nil
}

// KickUserFromChannel kicks a user from a given channel
// see https://api.slack.com/methods/channels.kick
func (api *Client) KickUserFromChannel(channelID, user string) error {
	return api.KickUserFromChannelContext(context.Background(), channelID, user)
}

// KickUserFromChannelContext kicks a user from a given channel with a custom context
// see https://api.slack.com/methods/channels.kick
func (api *Client) KickUserFromChannelContext(ctx context.Context, channelID, user string) (err error) {
	values := url.Values{
		"token":   {api.token},
		"channel": {channelID},
		"user":    {user},
	}

	_, err = api.channelRequest(ctx, "channels.kick", values)
	return err
}

func newChannelPagination(c *Client, options ...GetChannelsOption) (cp ChannelPagination) {
	cp = ChannelPagination{
		c:     c,
		limit: 200, // per slack api documentation.
	}

	for _, opt := range options {
		opt(&cp)
	}

	return cp
}

// ChannelPagination allows for paginating over the channels
type ChannelPagination struct {
	Channels        []Channel
	limit           int
	excludeArchived bool
	excludeMembers  bool
	previousResp    *ResponseMetadata
	c               *Client
}

// Done checks if the pagination has completed
func (ChannelPagination) Done(err error) bool {
	return err == errPaginationComplete
}

// Failure checks if pagination failed.
func (t ChannelPagination) Failure(err error) error {
	if t.Done(err) {
		return nil
	}

	return err
}

func (t ChannelPagination) Next(ctx context.Context) (_ ChannelPagination, err error) {
	var (
		resp *channelResponseFull
	)

	if t.c == nil || (t.previousResp != nil && t.previousResp.Cursor == "") {
		return t, errPaginationComplete
	}

	t.previousResp = t.previousResp.initialize()

	values := url.Values{
		"limit":            {strconv.Itoa(t.limit)},
		"exclude_archived": {strconv.FormatBool(t.excludeArchived)},
		"exclude_members":  {strconv.FormatBool(t.excludeMembers)},
		"token":            {t.c.token},
		"cursor":           {t.previousResp.Cursor},
	}

	if resp, err = t.c.channelRequest(ctx, "channels.list", values); err != nil {
		return t, err
	}

	t.c.Debugf("GetChannelsContext: got %d channels; metadata %v", len(resp.Channels), resp.Metadata)
	t.Channels = resp.Channels
	t.previousResp = &resp.Metadata

	return t, nil
}

// GetChannelsPaginated fetches channels in a paginated fashion, see GetChannelsContext for usage.
func (api *Client) GetChannelsPaginated(options ...GetChannelsOption) ChannelPagination {
	return newChannelPagination(api, options...)
}

// GetChannels retrieves all the channels
// see https://api.slack.com/methods/channels.list
func (api *Client) GetChannels(excludeArchived bool, options ...GetChannelsOption) ([]Channel, error) {
	return api.GetChannelsContext(context.Background(), excludeArchived, options...)
}

// GetChannelsContext retrieves all the channels with a custom context
// see https://api.slack.com/methods/channels.list
func (api *Client) GetChannelsContext(ctx context.Context, excludeArchived bool, options ...GetChannelsOption) (results []Channel, err error) {
	if excludeArchived {
		options = append(options, GetChannelsOptionExcludeArchived())
	}

	p := api.GetChannelsPaginated(options...)
	for err == nil {
		p, err = p.Next(ctx)
		if err == nil {
			results = append(results, p.Channels...)
		} else if rateLimitedError, ok := err.(*RateLimitedError); ok {
			select {
			case <-ctx.Done():
				err = ctx.Err()
			case <-time.After(rateLimitedError.RetryAfter):
				err = nil
			}
		}
	}

	return results, p.Failure(err)
}

// SetChannelReadMark sets the read mark of a given channel to a specific point
// Clients should try to avoid making this call too often. When needing to mark a read position, a client should set a
// timer before making the call. In this way, any further updates needed during the timeout will not generate extra calls
// (just one per channel). This is useful for when reading scroll-back history, or following a busy live channel. A
// timeout of 5 seconds is a good starting point. Be sure to flush these calls on shutdown/logout.
// see https://api.slack.com/methods/channels.mark
func (api *Client) SetChannelReadMark(channelID, ts string) error {
	return api.SetChannelReadMarkContext(context.Background(), channelID, ts)
}

// SetChannelReadMarkContext sets the read mark of a given channel to a specific point with a custom context
// For more details see SetChannelReadMark documentation
// see https://api.slack.com/methods/channels.mark
func (api *Client) SetChannelReadMarkContext(ctx context.Context, channelID, ts string) (err error) {
	values := url.Values{
		"token":   {api.token},
		"channel": {channelID},
		"ts":      {ts},
	}

	_, err = api.channelRequest(ctx, "channels.mark", values)
	return err
}

// RenameChannel renames a given channel
// see https://api.slack.com/methods/channels.rename
func (api *Client) RenameChannel(channelID, name string) (*Channel, error) {
	return api.RenameChannelContext(context.Background(), channelID, name)
}

// RenameChannelContext renames a given channel with a custom context
// see https://api.slack.com/methods/channels.rename
func (api *Client) RenameChannelContext(ctx context.Context, channelID, name string) (*Channel, error) {
	values := url.Values{
		"token":   {api.token},
		"channel": {channelID},
		"name":    {name},
	}

	// XXX: the created entry in this call returns a string instead of a number
	// so I may have to do some workaround to solve it.
	response, err := api.channelRequest(ctx, "channels.rename", values)
	if err != nil {
		return nil, err
	}
	return &response.Channel, nil
}

// SetChannelPurpose sets the channel purpose and returns the purpose that was successfully set
// see https://api.slack.com/methods/channels.setPurpose
func (api *Client) SetChannelPurpose(channelID, purpose string) (string, error) {
	return api.SetChannelPurposeContext(context.Background(), channelID, purpose)
}

// SetChannelPurposeContext sets the channel purpose and returns the purpose that was successfully set with a custom context
// see https://api.slack.com/methods/channels.setPurpose
func (api *Client) SetChannelPurposeContext(ctx context.Context, channelID, purpose string) (string, error) {
	values := url.Values{
		"token":   {api.token},
		"channel": {channelID},
		"purpose": {purpose},
	}

	response, err := api.channelRequest(ctx, "channels.setPurpose", values)
	if err != nil {
		return "", err
	}
	return response.Purpose, nil
}

// SetChannelTopic sets the channel topic and returns the topic that was successfully set
// see https://api.slack.com/methods/channels.setTopic
func (api *Client) SetChannelTopic(channelID, topic string) (string, error) {
	return api.SetChannelTopicContext(context.Background(), channelID, topic)
}

// SetChannelTopicContext sets the channel topic and returns the topic that was successfully set with a custom context
// see https://api.slack.com/methods/channels.setTopic
func (api *Client) SetChannelTopicContext(ctx context.Context, channelID, topic string) (string, error) {
	values := url.Values{
		"token":   {api.token},
		"channel": {channelID},
		"topic":   {topic},
	}

	response, err := api.channelRequest(ctx, "channels.setTopic", values)
	if err != nil {
		return "", err
	}
	return response.Topic, nil
}

// GetChannelReplies gets an entire thread (a message plus all the messages in reply to it).
// see https://api.slack.com/methods/channels.replies
func (api *Client) GetChannelReplies(channelID, thread_ts string) ([]Message, error) {
	return api.GetChannelRepliesContext(context.Background(), channelID, thread_ts)
}

// GetChannelRepliesContext gets an entire thread (a message plus all the messages in reply to it) with a custom context
// see https://api.slack.com/methods/channels.replies
func (api *Client) GetChannelRepliesContext(ctx context.Context, channelID, thread_ts string) ([]Message, error) {
	values := url.Values{
		"token":     {api.token},
		"channel":   {channelID},
		"thread_ts": {thread_ts},
	}
	response, err := api.channelRequest(ctx, "channels.replies", values)
	if err != nil {
		return nil, err
	}
	return response.History.Messages, nil
}
