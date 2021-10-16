package matterclient

import (
	"github.com/mattermost/mattermost-server/v6/model"
)

func (m *Client) GetNickName(userID string) string {
	if user := m.GetUser(userID); user != nil {
		return user.Nickname
	}

	return ""
}

func (m *Client) GetStatus(userID string) string {
	res, _, err := m.Client.GetUserStatus(userID, "")
	if err != nil {
		return ""
	}

	if res.Status == model.StatusAway {
		return "away"
	}

	if res.Status == model.StatusOnline {
		return "online"
	}

	return "offline"
}

func (m *Client) GetStatuses() map[string]string {
	var ids []string

	statuses := make(map[string]string)

	for id := range m.Users {
		ids = append(ids, id)
	}

	res, _, err := m.Client.GetUsersStatusesByIds(ids)
	if err != nil {
		return statuses
	}

	for _, status := range res {
		statuses[status.UserId] = "offline"
		if status.Status == model.StatusAway {
			statuses[status.UserId] = "away"
		}

		if status.Status == model.StatusOnline {
			statuses[status.UserId] = "online"
		}
	}

	return statuses
}

func (m *Client) GetTeamID() string {
	return m.Team.ID
}

// GetTeamName returns the name of the specified teamId
func (m *Client) GetTeamName(teamID string) string {
	m.RLock()
	defer m.RUnlock()

	for _, t := range m.OtherTeams {
		if t.ID == teamID {
			return t.Team.Name
		}
	}

	return ""
}

func (m *Client) GetUser(userID string) *model.User {
	m.Lock()
	defer m.Unlock()

	_, ok := m.Users[userID]
	if !ok {
		res, _, err := m.Client.GetUser(userID, "")
		if err != nil {
			return nil
		}

		m.Users[userID] = res
	}

	return m.Users[userID]
}

func (m *Client) GetUserName(userID string) string {
	if user := m.GetUser(userID); user != nil {
		return user.Username
	}

	return ""
}

func (m *Client) GetUsers() map[string]*model.User {
	users := make(map[string]*model.User)

	m.RLock()
	defer m.RUnlock()

	for k, v := range m.Users {
		users[k] = v
	}

	return users
}

func (m *Client) UpdateUsers() error {
	idx := 0
	max := 200

	var (
		mmusers []*model.User
		resp    *model.Response
		err     error
	)

	for {
		mmusers, resp, err = m.Client.GetUsers(idx, max, "")
		if err == nil {
			break
		}

		if err = m.HandleRatelimit("GetUsers", resp); err != nil {
			return err
		}
	}

	for len(mmusers) > 0 {
		m.Lock()

		for _, user := range mmusers {
			m.Users[user.Id] = user
		}

		m.Unlock()

		for {
			mmusers, resp, err = m.Client.GetUsers(idx, max, "")
			if err == nil {
				idx++

				break
			}

			if err := m.HandleRatelimit("GetUsers", resp); err != nil {
				return err
			}
		}
	}

	return nil
}

func (m *Client) UpdateUserNick(nick string) error {
	user := m.User
	user.Nickname = nick

	_, _, err := m.Client.UpdateUser(user)
	if err != nil {
		return err
	}

	return nil
}

func (m *Client) UsernamesInChannel(channelID string) []string {
	res, _, err := m.Client.GetChannelMembers(channelID, 0, 50000, "")
	if err != nil {
		m.logger.Errorf("UsernamesInChannel(%s) failed: %s", channelID, err)

		return []string{}
	}

	allusers := m.GetUsers()
	result := []string{}

	for _, member := range res {
		result = append(result, allusers[member.UserId].Nickname)
	}

	return result
}

func (m *Client) UpdateStatus(userID string, status string) error {
	_, _, err := m.Client.UpdateUserStatus(userID, &model.Status{Status: status})
	if err != nil {
		return err
	}

	return nil
}

func (m *Client) UpdateUser(userID string) {
	m.Lock()
	defer m.Unlock()

	res, _, err := m.Client.GetUser(userID, "")
	if err != nil {
		return
	}

	m.Users[userID] = res
}
