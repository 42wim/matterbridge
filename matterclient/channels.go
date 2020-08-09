package matterclient

import (
	"errors"
	"strings"

	"github.com/mattermost/mattermost-server/v5/model"
)

// GetChannels returns all channels we're members off
func (m *MMClient) GetChannels() []*model.Channel {
	m.RLock()
	defer m.RUnlock()
	var channels []*model.Channel
	// our primary team channels first
	channels = append(channels, m.Team.Channels...)
	for _, t := range m.OtherTeams {
		if t.Id != m.Team.Id {
			channels = append(channels, t.Channels...)
		}
	}
	return channels
}

func (m *MMClient) GetChannelHeader(channelId string) string { //nolint:golint
	m.RLock()
	defer m.RUnlock()
	for _, t := range m.OtherTeams {
		for _, channel := range append(t.Channels, t.MoreChannels...) {
			if channel.Id == channelId {
				return channel.Header
			}

		}
	}
	return ""
}

func getNormalisedName(channel *model.Channel) string {
	if channel.Type == model.CHANNEL_GROUP {
		// (deprecated in favor of ReplaceAll in go 1.12)
		res := strings.Replace(channel.DisplayName, ", ", "-", -1) //nolint: gocritic
		res = strings.Replace(res, " ", "_", -1)                   //nolint: gocritic
		return res
	}
	return channel.Name
}

func (m *MMClient) GetChannelId(name string, teamId string) string { //nolint:golint
	m.RLock()
	defer m.RUnlock()
	if teamId != "" {
		return m.getChannelIdTeam(name, teamId)
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

func (m *MMClient) getChannelIdTeam(name string, teamId string) string { //nolint:golint
	for _, t := range m.OtherTeams {
		if t.Id == teamId {
			for _, channel := range append(t.Channels, t.MoreChannels...) {
				if getNormalisedName(channel) == name {
					return channel.Id
				}
			}
		}
	}
	return ""
}

func (m *MMClient) GetChannelName(channelId string) string { //nolint:golint
	m.RLock()
	defer m.RUnlock()
	for _, t := range m.OtherTeams {
		if t == nil {
			continue
		}
		for _, channel := range append(t.Channels, t.MoreChannels...) {
			if channel.Id == channelId {
				return getNormalisedName(channel)
			}
		}
	}
	return ""
}

func (m *MMClient) GetChannelTeamId(id string) string { //nolint:golint
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

func (m *MMClient) GetLastViewedAt(channelId string) int64 { //nolint:golint
	m.RLock()
	defer m.RUnlock()
	res, resp := m.Client.GetChannelMember(channelId, m.User.Id, "")
	if resp.Error != nil {
		return model.GetMillis()
	}
	return res.LastViewedAt
}

// GetMoreChannels returns existing channels where we're not a member off.
func (m *MMClient) GetMoreChannels() []*model.Channel {
	m.RLock()
	defer m.RUnlock()
	var channels []*model.Channel
	for _, t := range m.OtherTeams {
		channels = append(channels, t.MoreChannels...)
	}
	return channels
}

// GetTeamFromChannel returns teamId belonging to channel (DM channels have no teamId).
func (m *MMClient) GetTeamFromChannel(channelId string) string { //nolint:golint
	m.RLock()
	defer m.RUnlock()
	var channels []*model.Channel
	for _, t := range m.OtherTeams {
		channels = append(channels, t.Channels...)
		if t.MoreChannels != nil {
			channels = append(channels, t.MoreChannels...)
		}
		for _, c := range channels {
			if c.Id == channelId {
				if c.Type == model.CHANNEL_GROUP {
					return "G"
				}
				return t.Id
			}
		}
		channels = nil
	}
	return ""
}

func (m *MMClient) JoinChannel(channelId string) error { //nolint:golint
	m.RLock()
	defer m.RUnlock()
	for _, c := range m.Team.Channels {
		if c.Id == channelId {
			m.logger.Debug("Not joining ", channelId, " already joined.")
			return nil
		}
	}
	m.logger.Debug("Joining ", channelId)
	_, resp := m.Client.AddChannelMember(channelId, m.User.Id)
	if resp.Error != nil {
		return resp.Error
	}
	return nil
}

func (m *MMClient) UpdateChannelsTeam(teamID string) error {
	mmchannels, resp := m.Client.GetChannelsForTeamForUser(teamID, m.User.Id, false, "")
	if resp.Error != nil {
		return errors.New(resp.Error.DetailedError)
	}
	for idx, t := range m.OtherTeams {
		if t.Id == teamID {
			m.Lock()
			m.OtherTeams[idx].Channels = mmchannels
			m.Unlock()
		}
	}

	mmchannels, resp = m.Client.GetPublicChannelsForTeam(teamID, 0, 5000, "")
	if resp.Error != nil {
		return errors.New(resp.Error.DetailedError)
	}
	for idx, t := range m.OtherTeams {
		if t.Id == teamID {
			m.Lock()
			m.OtherTeams[idx].MoreChannels = mmchannels
			m.Unlock()
		}
	}
	return nil
}

func (m *MMClient) UpdateChannels() error {
	if err := m.UpdateChannelsTeam(m.Team.Id); err != nil {
		return err
	}
	for _, t := range m.OtherTeams {
		if err := m.UpdateChannelsTeam(t.Id); err != nil {
			return err
		}
	}
	return nil
}

func (m *MMClient) UpdateChannelHeader(channelId string, header string) { //nolint:golint
	channel := &model.Channel{Id: channelId, Header: header}
	m.logger.Debugf("updating channelheader %#v, %#v", channelId, header)
	_, resp := m.Client.UpdateChannel(channel)
	if resp.Error != nil {
		m.logger.Error(resp.Error)
	}
}

func (m *MMClient) UpdateLastViewed(channelId string) error { //nolint:golint
	m.logger.Debugf("posting lastview %#v", channelId)
	view := &model.ChannelView{ChannelId: channelId}
	_, resp := m.Client.ViewChannel(m.User.Id, view)
	if resp.Error != nil {
		m.logger.Errorf("ChannelView update for %s failed: %s", channelId, resp.Error)
		return resp.Error
	}
	return nil
}
