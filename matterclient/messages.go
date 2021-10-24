package matterclient

import (
	"strings"

	"github.com/mattermost/mattermost-server/v5/model"
)

func (m *MMClient) parseActionPost(rmsg *Message) {
	// add post to cache, if it already exists don't relay this again.
	// this should fix reposts
	if ok, _ := m.lruCache.ContainsOrAdd(digestString(rmsg.Raw.Data["post"].(string)), true); ok && rmsg.Raw.Event != model.WEBSOCKET_EVENT_POST_DELETED {
		m.logger.Debugf("message %#v in cache, not processing again", rmsg.Raw.Data["post"].(string))
		rmsg.Text = ""
		return
	}
	data := model.PostFromJson(strings.NewReader(rmsg.Raw.Data["post"].(string)))
	// we don't have the user, refresh the userlist
	if m.GetUser(data.UserId) == nil {
		m.logger.Infof("User '%v' is not known, ignoring message '%#v'",
			data.UserId, data)
		return
	}
	rmsg.Username = m.GetUserName(data.UserId)
	rmsg.Channel = m.GetChannelName(data.ChannelId)
	rmsg.UserID = data.UserId
	rmsg.Type = data.Type
	teamid, _ := rmsg.Raw.Data["team_id"].(string)
	// edit messsages have no team_id for some reason
	if teamid == "" {
		// we can find the team_id from the channelid
		teamid = m.GetChannelTeamId(data.ChannelId)
		rmsg.Raw.Data["team_id"] = teamid
	}
	if teamid != "" {
		rmsg.Team = m.GetTeamName(teamid)
	}
	// direct message
	if rmsg.Raw.Data["channel_type"] == "D" {
		rmsg.Channel = m.GetUser(data.UserId).Username
	}
	rmsg.Text = data.Message
	rmsg.Post = data
}

func (m *MMClient) parseMessage(rmsg *Message) {
	switch rmsg.Raw.Event {
	case model.WEBSOCKET_EVENT_POSTED, model.WEBSOCKET_EVENT_POST_EDITED, model.WEBSOCKET_EVENT_POST_DELETED:
		m.parseActionPost(rmsg)
	case "user_updated":
		user := rmsg.Raw.Data["user"].(map[string]interface{})
		if _, ok := user["id"].(string); ok {
			m.UpdateUser(user["id"].(string))
		}
	case "group_added":
		if err := m.UpdateChannels(); err != nil {
			m.logger.Errorf("failed to update channels: %#v", err)
		}
		/*
			case model.ACTION_USER_REMOVED:
				m.handleWsActionUserRemoved(&rmsg)
			case model.ACTION_USER_ADDED:
				m.handleWsActionUserAdded(&rmsg)
		*/
	}
}

func (m *MMClient) parseResponse(rmsg model.WebSocketResponse) {
	if rmsg.Data != nil {
		// ping reply
		if rmsg.Data["text"].(string) == "pong" {
			m.WsPingChan <- &rmsg
		}
	}
}

func (m *MMClient) DeleteMessage(postId string) error { //nolint:golint
	_, resp := m.Client.DeletePost(postId)
	if resp.Error != nil {
		return resp.Error
	}
	return nil
}

func (m *MMClient) EditMessage(postId string, text string) (string, error) { //nolint:golint
	post := &model.Post{Message: text, Id: postId}
	res, resp := m.Client.UpdatePost(postId, post)
	if resp.Error != nil {
		return "", resp.Error
	}
	return res.Id, nil
}

func (m *MMClient) GetFileLinks(filenames []string) []string {
	uriScheme := "https://"
	if m.NoTLS {
		uriScheme = "http://"
	}

	var output []string
	for _, f := range filenames {
		res, resp := m.Client.GetFileLink(f)
		if resp.Error != nil {
			// public links is probably disabled, create the link ourselves
			output = append(output, uriScheme+m.Credentials.Server+model.API_URL_SUFFIX_V4+"/files/"+f)
			continue
		}
		output = append(output, res)
	}
	return output
}

func (m *MMClient) GetPosts(channelId string, limit int) *model.PostList { //nolint:golint
	res, resp := m.Client.GetPostsForChannel(channelId, 0, limit, "", true)
	if resp.Error != nil {
		return nil
	}
	return res
}

func (m *MMClient) GetPostsSince(channelId string, time int64) *model.PostList { //nolint:golint
	res, resp := m.Client.GetPostsSince(channelId, time, true)
	if resp.Error != nil {
		return nil
	}
	return res
}

func (m *MMClient) GetPublicLink(filename string) string {
	res, resp := m.Client.GetFileLink(filename)
	if resp.Error != nil {
		return ""
	}
	return res
}

func (m *MMClient) GetPublicLinks(filenames []string) []string {
	var output []string
	for _, f := range filenames {
		res, resp := m.Client.GetFileLink(f)
		if resp.Error != nil {
			continue
		}
		output = append(output, res)
	}
	return output
}

func (m *MMClient) PostMessage(channelId string, text string, rootId string) (string, error) { //nolint:golint
	post := &model.Post{ChannelId: channelId, Message: text, RootId: rootId}
	res, resp := m.Client.CreatePost(post)
	if resp.Error != nil {
		return "", resp.Error
	}
	return res.Id, nil
}

func (m *MMClient) PostMessageWithFiles(channelId string, text string, rootId string, fileIds []string) (string, error) { //nolint:golint
	post := &model.Post{ChannelId: channelId, Message: text, RootId: rootId, FileIds: fileIds}
	res, resp := m.Client.CreatePost(post)
	if resp.Error != nil {
		return "", resp.Error
	}
	return res.Id, nil
}

func (m *MMClient) SearchPosts(query string) *model.PostList {
	res, resp := m.Client.SearchPosts(m.Team.Id, query, false)
	if resp.Error != nil {
		return nil
	}
	return res
}

// SendDirectMessage sends a direct message to specified user
func (m *MMClient) SendDirectMessage(toUserId string, msg string, rootId string) { //nolint:golint
	m.SendDirectMessageProps(toUserId, msg, rootId, nil)
}

func (m *MMClient) SendDirectMessageProps(toUserId string, msg string, rootId string, props map[string]interface{}) { //nolint:golint
	m.logger.Debugf("SendDirectMessage to %s, msg %s", toUserId, msg)
	// create DM channel (only happens on first message)
	_, resp := m.Client.CreateDirectChannel(m.User.Id, toUserId)
	if resp.Error != nil {
		m.logger.Debugf("SendDirectMessage to %#v failed: %s", toUserId, resp.Error)
		return
	}
	channelName := model.GetDMNameFromIds(toUserId, m.User.Id)

	// update our channels
	if err := m.UpdateChannels(); err != nil {
		m.logger.Errorf("failed to update channels: %#v", err)
	}

	// build & send the message
	msg = strings.Replace(msg, "\r", "", -1)
	post := &model.Post{ChannelId: m.GetChannelId(channelName, m.Team.Id), Message: msg, RootId: rootId, Props: props}
	m.Client.CreatePost(post)
}

func (m *MMClient) UploadFile(data []byte, channelId string, filename string) (string, error) { //nolint:golint
	f, resp := m.Client.UploadFile(data, channelId, filename)
	if resp.Error != nil {
		return "", resp.Error
	}
	return f.FileInfos[0].Id, nil
}
