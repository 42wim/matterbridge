package matterclient

import (
	"strings"

	"github.com/mattermost/mattermost-server/v6/model"
)

// GetChannels returns all channels we're members off
func (m *Client) GetChannels() []*model.Channel {
	m.RLock()
	defer m.RUnlock()

	var channels []*model.Channel
	// our primary team channels first
	channels = append(channels, m.Team.Channels...)

	for _, t := range m.OtherTeams {
		if t.ID != m.Team.ID {
			channels = append(channels, t.Channels...)
		}
	}

	return channels
}

func (m *Client) GetChannelHeader(channelID string) string {
	m.RLock()
	defer m.RUnlock()

	for _, t := range m.OtherTeams {
		for _, channel := range append(t.Channels, t.MoreChannels...) {
			if channel.Id == channelID {
				return channel.Header
			}
		}
	}

	return ""
}

func getNormalisedName(channel *model.Channel) string {
	if channel.Type == model.ChannelTypeGroup {
		res := strings.ReplaceAll(channel.DisplayName, ", ", "-")
		res = strings.ReplaceAll(res, " ", "_")

		return res
	}

	return channel.Name
}

func (m *Client) GetChannelID(name string, teamID string) string {
	m.RLock()
	defer m.RUnlock()

	if teamID != "" {
		return m.getChannelIDTeam(name, teamID)
	}

	for _, t := range m.OtherTeams {
		for _, channel := range append(t.Channels, t.MoreChannels...) {
			if getNormalisedName(channel) == name {
				return channel.Id
			}
		}
	}

	return ""
}

func (m *Client) getChannelIDTeam(name string, teamID string) string {
	for _, t := range m.OtherTeams {
		if t.ID == teamID {
			for _, channel := range append(t.Channels, t.MoreChannels...) {
				if getNormalisedName(channel) == name {
					return channel.Id
				}
			}
		}
	}

	return ""
}

func (m *Client) GetChannelName(channelID string) string {
	m.RLock()
	defer m.RUnlock()

	for _, t := range m.OtherTeams {
		if t == nil {
			continue
		}

		for _, channel := range append(t.Channels, t.MoreChannels...) {
			if channel.Id == channelID {
				return getNormalisedName(channel)
			}
		}
	}

	return ""
}

func (m *Client) GetChannelTeamID(id string) string {
	m.RLock()
	defer m.RUnlock()

	for _, t := range append(m.OtherTeams, m.Team) {
		for _, channel := range append(t.Channels, t.MoreChannels...) {
			if channel.Id == id {
				return channel.TeamId
			}
		}
	}

	return ""
}

func (m *Client) GetLastViewedAt(channelID string) int64 {
	m.RLock()
	defer m.RUnlock()

	for {
		res, resp, err := m.Client.GetChannelMember(channelID, m.User.Id, "")
		if err == nil {
			return res.LastViewedAt
		}

		if err := m.HandleRatelimit("GetChannelMember", resp); err != nil {
			return model.GetMillis()
		}
	}
}

// GetMoreChannels returns existing channels where we're not a member off.
func (m *Client) GetMoreChannels() []*model.Channel {
	m.RLock()
	defer m.RUnlock()

	var channels []*model.Channel
	for _, t := range m.OtherTeams {
		channels = append(channels, t.MoreChannels...)
	}

	return channels
}

// GetTeamFromChannel returns teamId belonging to channel (DM channels have no teamId).
func (m *Client) GetTeamFromChannel(channelID string) string {
	m.RLock()
	defer m.RUnlock()

	var channels []*model.Channel

	for _, t := range m.OtherTeams {
		channels = append(channels, t.Channels...)

		if t.MoreChannels != nil {
			channels = append(channels, t.MoreChannels...)
		}

		for _, c := range channels {
			if c.Id == channelID {
				if c.Type == model.ChannelTypeGroup {
					return "G"
				}

				return t.ID
			}
		}

		channels = nil
	}

	return ""
}

func (m *Client) JoinChannel(channelID string) error {
	m.RLock()
	defer m.RUnlock()

	for _, c := range m.Team.Channels {
		if c.Id == channelID {
			m.logger.Debug("Not joining ", channelID, " already joined.")

			return nil
		}
	}

	m.logger.Debug("Joining ", channelID)

	_, _, err := m.Client.AddChannelMember(channelID, m.User.Id)
	if err != nil {
		return err
	}

	return nil
}

func (m *Client) UpdateChannelsTeam(teamID string) error {
	var (
		mmchannels []*model.Channel
		resp       *model.Response
		err        error
	)

	for {
		mmchannels, resp, err = m.Client.GetChannelsForTeamForUser(teamID, m.User.Id, false, "")
		if err == nil {
			break
		}

		if err = m.HandleRatelimit("GetChannelsForTeamForUser", resp); err != nil {
			return err
		}
	}

	for idx, t := range m.OtherTeams {
		if t.ID == teamID {
			m.Lock()
			m.OtherTeams[idx].Channels = mmchannels
			m.Unlock()
		}
	}

	for {
		mmchannels, resp, err = m.Client.GetPublicChannelsForTeam(teamID, 0, 5000, "")
		if err == nil {
			break
		}

		if err := m.HandleRatelimit("GetPublicChannelsForTeam", resp); err != nil {
			return err
		}
	}

	for idx, t := range m.OtherTeams {
		if t.ID == teamID {
			m.Lock()
			m.OtherTeams[idx].MoreChannels = mmchannels
			m.Unlock()
		}
	}

	return nil
}

func (m *Client) UpdateChannels() error {
	if err := m.UpdateChannelsTeam(m.Team.ID); err != nil {
		return err
	}

	for _, t := range m.OtherTeams {
		if err := m.UpdateChannelsTeam(t.ID); err != nil {
			return err
		}
	}

	return nil
}

func (m *Client) UpdateChannelHeader(channelID string, header string) {
	channel := &model.Channel{Id: channelID, Header: header}

	m.logger.Debugf("updating channelheader %#v, %#v", channelID, header)

	_, _, err := m.Client.UpdateChannel(channel)
	if err != nil {
		m.logger.Error(err)
	}
}

func (m *Client) UpdateLastViewed(channelID string) error {
	m.logger.Debugf("posting lastview %#v", channelID)

	view := &model.ChannelView{ChannelId: channelID}

	for {
		_, resp, err := m.Client.ViewChannel(m.User.Id, view)
		if err == nil {
			return nil
		}

		if err := m.HandleRatelimit("ViewChannel", resp); err != nil {
			m.logger.Errorf("ChannelView update for %s failed: %s", channelID, err)

			return err
		}
	}
}
