package matterclient

import (
	"errors"
	"time"

	"github.com/mattermost/mattermost-server/v5/model"
)

func (m *MMClient) GetNickName(userId string) string { //nolint:golint
	user := m.GetUser(userId)
	if user != nil {
		return user.Nickname
	}
	return ""
}

func (m *MMClient) GetStatus(userId string) string { //nolint:golint
	res, resp := m.Client.GetUserStatus(userId, "")
	if resp.Error != nil {
		return ""
	}
	if res.Status == model.STATUS_AWAY {
		return "away"
	}
	if res.Status == model.STATUS_ONLINE {
		return "online"
	}
	return "offline"
}

func (m *MMClient) GetStatuses() map[string]string {
	var ids []string
	statuses := make(map[string]string)
	for id := range m.Users {
		ids = append(ids, id)
	}
	res, resp := m.Client.GetUsersStatusesByIds(ids)
	if resp.Error != nil {
		return statuses
	}
	for _, status := range res {
		statuses[status.UserId] = "offline"
		if status.Status == model.STATUS_AWAY {
			statuses[status.UserId] = "away"
		}
		if status.Status == model.STATUS_ONLINE {
			statuses[status.UserId] = "online"
		}
	}
	return statuses
}

func (m *MMClient) GetTeamId() string { //nolint:golint
	return m.Team.Id
}

// GetTeamName returns the name of the specified teamId
func (m *MMClient) GetTeamName(teamId string) string { //nolint:golint
	m.RLock()
	defer m.RUnlock()
	for _, t := range m.OtherTeams {
		if t.Id == teamId {
			return t.Team.Name
		}
	}
	return ""
}

func (m *MMClient) GetUser(userId string) *model.User { //nolint:golint
	m.Lock()
	defer m.Unlock()
	_, ok := m.Users[userId]
	if !ok {
		res, resp := m.Client.GetUser(userId, "")
		if resp.Error != nil {
			return nil
		}
		m.Users[userId] = res
	}
	return m.Users[userId]
}

func (m *MMClient) GetUserName(userId string) string { //nolint:golint
	user := m.GetUser(userId)
	if user != nil {
		return user.Username
	}
	return ""
}

func (m *MMClient) GetUsers() map[string]*model.User {
	users := make(map[string]*model.User)
	m.RLock()
	defer m.RUnlock()
	for k, v := range m.Users {
		users[k] = v
	}
	return users
}

func (m *MMClient) UpdateUsers() error {
	idx := 0
	max := 200
	mmusers, resp := m.Client.GetUsers(idx, max, "")
	if resp.Error != nil {
		return errors.New(resp.Error.DetailedError)
	}
	for len(mmusers) > 0 {
		m.Lock()
		for _, user := range mmusers {
			m.Users[user.Id] = user
		}
		m.Unlock()
		mmusers, resp = m.Client.GetUsers(idx, max, "")
		time.Sleep(time.Millisecond * 300)
		if resp.Error != nil {
			return errors.New(resp.Error.DetailedError)
		}
		idx++
	}
	return nil
}

func (m *MMClient) UpdateUserNick(nick string) error {
	user := m.User
	user.Nickname = nick
	_, resp := m.Client.UpdateUser(user)
	if resp.Error != nil {
		return resp.Error
	}
	return nil
}

func (m *MMClient) UsernamesInChannel(channelId string) []string { //nolint:golint
	res, resp := m.Client.GetChannelMembers(channelId, 0, 50000, "")
	if resp.Error != nil {
		m.logger.Errorf("UsernamesInChannel(%s) failed: %s", channelId, resp.Error)
		return []string{}
	}
	allusers := m.GetUsers()
	result := []string{}
	for _, member := range *res {
		result = append(result, allusers[member.UserId].Nickname)
	}
	return result
}

func (m *MMClient) UpdateStatus(userId string, status string) error { //nolint:golint
	_, resp := m.Client.UpdateUserStatus(userId, &model.Status{Status: status})
	if resp.Error != nil {
		return resp.Error
	}
	return nil
}

func (m *MMClient) UpdateUser(userId string) { //nolint:golint
	m.Lock()
	defer m.Unlock()
	res, resp := m.Client.GetUser(userId, "")
	if resp.Error != nil {
		return
	}
	m.Users[userId] = res
}
