package matterclient

import (
	"crypto/md5"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/hashicorp/golang-lru"
	"github.com/jpillora/backoff"
	prefixed "github.com/matterbridge/logrus-prefixed-formatter"
	"github.com/mattermost/platform/model"
	log "github.com/sirupsen/logrus"
)

type Credentials struct {
	Login         string
	Team          string
	Pass          string
	Token         string
	CookieToken   bool
	Server        string
	NoTLS         bool
	SkipTLSVerify bool
}

type Message struct {
	Raw      *model.WebSocketEvent
	Post     *model.Post
	Team     string
	Channel  string
	Username string
	Text     string
	Type     string
	UserID   string
}

//nolint:golint
type Team struct {
	Team         *model.Team
	Id           string
	Channels     []*model.Channel
	MoreChannels []*model.Channel
	Users        map[string]*model.User
}

type MMClient struct {
	sync.RWMutex
	*Credentials
	Team          *Team
	OtherTeams    []*Team
	Client        *model.Client4
	User          *model.User
	Users         map[string]*model.User
	MessageChan   chan *Message
	log           *log.Entry
	WsClient      *websocket.Conn
	WsQuit        bool
	WsAway        bool
	WsConnected   bool
	WsSequence    int64
	WsPingChan    chan *model.WebSocketResponse
	ServerVersion string
	OnWsConnect   func()
	lruCache      *lru.Cache
}

func New(login, pass, team, server string) *MMClient {
	cred := &Credentials{Login: login, Pass: pass, Team: team, Server: server}
	mmclient := &MMClient{Credentials: cred, MessageChan: make(chan *Message, 100), Users: make(map[string]*model.User)}
	log.SetFormatter(&prefixed.TextFormatter{PrefixPadding: 13, DisableColors: true})
	mmclient.log = log.WithFields(log.Fields{"prefix": "matterclient"})
	mmclient.lruCache, _ = lru.New(500)
	return mmclient
}

func (m *MMClient) SetDebugLog() {
	log.SetFormatter(&prefixed.TextFormatter{PrefixPadding: 13, DisableColors: true, FullTimestamp: false, ForceFormatting: true})
}

func (m *MMClient) SetLogLevel(level string) {
	l, err := log.ParseLevel(level)
	if err != nil {
		log.SetLevel(log.InfoLevel)
		return
	}
	log.SetLevel(l)
}

func (m *MMClient) Login() error {
	// check if this is a first connect or a reconnection
	firstConnection := true
	if m.WsConnected {
		firstConnection = false
	}
	m.WsConnected = false
	if m.WsQuit {
		return nil
	}
	b := &backoff.Backoff{
		Min:    time.Second,
		Max:    5 * time.Minute,
		Jitter: true,
	}
	uriScheme := "https://"
	if m.NoTLS {
		uriScheme = "http://"
	}
	// login to mattermost
	m.Client = model.NewAPIv4Client(uriScheme + m.Credentials.Server)
	m.Client.HttpClient.Transport = &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: m.SkipTLSVerify}, Proxy: http.ProxyFromEnvironment}
	m.Client.HttpClient.Timeout = time.Second * 10

	if strings.Contains(m.Credentials.Pass, model.SESSION_COOKIE_TOKEN) {
		token := strings.Split(m.Credentials.Pass, model.SESSION_COOKIE_TOKEN+"=")
		if len(token) != 2 {
			return errors.New("incorrect MMAUTHTOKEN. valid input is MMAUTHTOKEN=yourtoken")
		}
		m.Credentials.Token = token[1]
		m.Credentials.CookieToken = true
	}

	if strings.Contains(m.Credentials.Pass, "token=") {
		token := strings.Split(m.Credentials.Pass, "token=")
		if len(token) != 2 {
			return errors.New("incorrect personal token. valid input is token=yourtoken")
		}
		m.Credentials.Token = token[1]
	}

	for {
		d := b.Duration()
		// bogus call to get the serverversion
		_, resp := m.Client.Logout()
		if resp.Error != nil {
			return fmt.Errorf("%#v", resp.Error.Error())
		}
		if firstConnection && !supportedVersion(resp.ServerVersion) {
			return fmt.Errorf("unsupported mattermost version: %s", resp.ServerVersion)
		}
		m.ServerVersion = resp.ServerVersion
		if m.ServerVersion == "" {
			m.log.Debugf("Server not up yet, reconnecting in %s", d)
			time.Sleep(d)
		} else {
			m.log.Infof("Found version %s", m.ServerVersion)
			break
		}
	}
	b.Reset()

	var resp *model.Response
	//var myinfo *model.Result
	var appErr *model.AppError
	var logmsg = "trying login"
	for {
		m.log.Debugf("%s %s %s %s", logmsg, m.Credentials.Team, m.Credentials.Login, m.Credentials.Server)
		if m.Credentials.Token != "" {
			m.Client.AuthType = model.HEADER_BEARER
			m.Client.AuthToken = m.Credentials.Token
			if m.Credentials.CookieToken {
				m.log.Debugf(logmsg + " with cookie (MMAUTH) token")
				m.Client.HttpClient.Jar = m.createCookieJar(m.Credentials.Token)
			} else {
				m.log.Debugf(logmsg + " with personal token")
			}
			m.User, resp = m.Client.GetMe("")
			if resp.Error != nil {
				return resp.Error
			}
			if m.User == nil {
				m.log.Errorf("LOGIN TOKEN: %s is invalid", m.Credentials.Pass)
				return errors.New("invalid token")
			}
		} else {
			m.User, resp = m.Client.Login(m.Credentials.Login, m.Credentials.Pass)
		}
		appErr = resp.Error
		if appErr != nil {
			d := b.Duration()
			m.log.Debug(appErr.DetailedError)
			if firstConnection {
				if appErr.Message == "" {
					return errors.New(appErr.DetailedError)
				}
				return errors.New(appErr.Message)
			}
			m.log.Debugf("LOGIN: %s, reconnecting in %s", appErr, d)
			time.Sleep(d)
			logmsg = "retrying login"
			continue
		}
		break
	}
	// reset timer
	b.Reset()

	err := m.initUser()
	if err != nil {
		return err
	}

	if m.Team == nil {
		validTeamNames := make([]string, len(m.OtherTeams))
		for i, t := range m.OtherTeams {
			validTeamNames[i] = t.Team.Name
		}
		return fmt.Errorf("Team '%s' not found in %v", m.Credentials.Team, validTeamNames)
	}

	m.wsConnect()

	return nil
}

func (m *MMClient) wsConnect() {
	b := &backoff.Backoff{
		Min:    time.Second,
		Max:    5 * time.Minute,
		Jitter: true,
	}

	m.WsConnected = false
	wsScheme := "wss://"
	if m.NoTLS {
		wsScheme = "ws://"
	}

	// setup websocket connection
	wsurl := wsScheme + m.Credentials.Server + model.API_URL_SUFFIX_V4 + "/websocket"
	header := http.Header{}
	header.Set(model.HEADER_AUTH, "BEARER "+m.Client.AuthToken)

	m.log.Debugf("WsClient: making connection: %s", wsurl)
	for {
		wsDialer := &websocket.Dialer{Proxy: http.ProxyFromEnvironment, TLSClientConfig: &tls.Config{InsecureSkipVerify: m.SkipTLSVerify}}
		var err error
		m.WsClient, _, err = wsDialer.Dial(wsurl, header)
		if err != nil {
			d := b.Duration()
			m.log.Debugf("WSS: %s, reconnecting in %s", err, d)
			time.Sleep(d)
			continue
		}
		break
	}

	m.log.Debug("WsClient: connected")
	m.WsSequence = 1
	m.WsPingChan = make(chan *model.WebSocketResponse)
	// only start to parse WS messages when login is completely done
	m.WsConnected = true
}

func (m *MMClient) Logout() error {
	m.log.Debugf("logout as %s (team: %s) on %s", m.Credentials.Login, m.Credentials.Team, m.Credentials.Server)
	m.WsQuit = true
	m.WsClient.Close()
	m.WsClient.UnderlyingConn().Close()
	if strings.Contains(m.Credentials.Pass, model.SESSION_COOKIE_TOKEN) {
		m.log.Debug("Not invalidating session in logout, credential is a token")
		return nil
	}
	_, resp := m.Client.Logout()
	if resp.Error != nil {
		return resp.Error
	}
	return nil
}

func (m *MMClient) WsReceiver() {
	for {
		var rawMsg json.RawMessage
		var err error

		if m.WsQuit {
			m.log.Debug("exiting WsReceiver")
			return
		}

		if !m.WsConnected {
			time.Sleep(time.Millisecond * 100)
			continue
		}

		if _, rawMsg, err = m.WsClient.ReadMessage(); err != nil {
			m.log.Error("error:", err)
			// reconnect
			m.wsConnect()
		}

		var event model.WebSocketEvent
		if err := json.Unmarshal(rawMsg, &event); err == nil && event.IsValid() {
			m.log.Debugf("WsReceiver event: %#v", event)
			msg := &Message{Raw: &event, Team: m.Credentials.Team}
			m.parseMessage(msg)
			// check if we didn't empty the message
			if msg.Text != "" {
				m.MessageChan <- msg
				continue
			}
			// if we have file attached but the message is empty, also send it
			if msg.Post != nil {
				if msg.Text != "" || len(msg.Post.FileIds) > 0 || msg.Post.Type == "slack_attachment" {
					m.MessageChan <- msg
				}
			}
			continue
		}

		var response model.WebSocketResponse
		if err := json.Unmarshal(rawMsg, &response); err == nil && response.IsValid() {
			m.log.Debugf("WsReceiver response: %#v", response)
			m.parseResponse(response)
			continue
		}
	}
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
		m.UpdateChannels()
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

func (m *MMClient) parseActionPost(rmsg *Message) {
	// add post to cache, if it already exists don't relay this again.
	// this should fix reposts
	if ok, _ := m.lruCache.ContainsOrAdd(digestString(rmsg.Raw.Data["post"].(string)), true); ok {
		m.log.Debugf("message %#v in cache, not processing again", rmsg.Raw.Data["post"].(string))
		rmsg.Text = ""
		return
	}
	data := model.PostFromJson(strings.NewReader(rmsg.Raw.Data["post"].(string)))
	// we don't have the user, refresh the userlist
	if m.GetUser(data.UserId) == nil {
		m.log.Infof("User '%v' is not known, ignoring message '%#v'",
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

func (m *MMClient) UpdateUsers() error {
	mmusers, resp := m.Client.GetUsers(0, 50000, "")
	if resp.Error != nil {
		return errors.New(resp.Error.DetailedError)
	}
	m.Lock()
	for _, user := range mmusers {
		m.Users[user.Id] = user
	}
	m.Unlock()
	return nil
}

func (m *MMClient) UpdateChannels() error {
	mmchannels, resp := m.Client.GetChannelsForTeamForUser(m.Team.Id, m.User.Id, "")
	if resp.Error != nil {
		return errors.New(resp.Error.DetailedError)
	}
	m.Lock()
	m.Team.Channels = mmchannels
	m.Unlock()

	mmchannels, resp = m.Client.GetPublicChannelsForTeam(m.Team.Id, 0, 5000, "")
	if resp.Error != nil {
		return errors.New(resp.Error.DetailedError)
	}

	m.Lock()
	m.Team.MoreChannels = mmchannels
	m.Unlock()
	return nil
}

func (m *MMClient) GetChannelName(channelId string) string { //nolint:golint
	m.RLock()
	defer m.RUnlock()
	for _, t := range m.OtherTeams {
		if t == nil {
			continue
		}
		if t.Channels != nil {
			for _, channel := range t.Channels {
				if channel.Id == channelId {
					if channel.Type == model.CHANNEL_GROUP {
						res := strings.Replace(channel.DisplayName, ", ", "-", -1)
						res = strings.Replace(res, " ", "_", -1)
						return res
					}
					return channel.Name
				}
			}
		}
		if t.MoreChannels != nil {
			for _, channel := range t.MoreChannels {
				if channel.Id == channelId {
					if channel.Type == model.CHANNEL_GROUP {
						res := strings.Replace(channel.DisplayName, ", ", "-", -1)
						res = strings.Replace(res, " ", "_", -1)
						return res
					}
					return channel.Name
				}
			}
		}
	}
	return ""
}

func (m *MMClient) GetChannelId(name string, teamId string) string { //nolint:golint
	m.RLock()
	defer m.RUnlock()
	if teamId == "" {
		for _, t := range m.OtherTeams {
			for _, channel := range append(t.Channels, t.MoreChannels...) {
				if channel.Type == model.CHANNEL_GROUP {
					res := strings.Replace(channel.DisplayName, ", ", "-", -1)
					res = strings.Replace(res, " ", "_", -1)
					if res == name {
						return channel.Id
					}
				}

			}
		}
	}

	for _, t := range m.OtherTeams {
		if t.Id == teamId {
			for _, channel := range append(t.Channels, t.MoreChannels...) {
				if channel.Name == name {
					return channel.Id
				}
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

func (m *MMClient) PostMessage(channelId string, text string) (string, error) { //nolint:golint
	post := &model.Post{ChannelId: channelId, Message: text}
	res, resp := m.Client.CreatePost(post)
	if resp.Error != nil {
		return "", resp.Error
	}
	return res.Id, nil
}

func (m *MMClient) PostMessageWithFiles(channelId string, text string, fileIds []string) (string, error) { //nolint:golint
	post := &model.Post{ChannelId: channelId, Message: text, FileIds: fileIds}
	res, resp := m.Client.CreatePost(post)
	if resp.Error != nil {
		return "", resp.Error
	}
	return res.Id, nil
}

func (m *MMClient) EditMessage(postId string, text string) (string, error) { //nolint:golint
	post := &model.Post{Message: text}
	res, resp := m.Client.UpdatePost(postId, post)
	if resp.Error != nil {
		return "", resp.Error
	}
	return res.Id, nil
}

func (m *MMClient) DeleteMessage(postId string) error { //nolint:golint
	_, resp := m.Client.DeletePost(postId)
	if resp.Error != nil {
		return resp.Error
	}
	return nil
}

func (m *MMClient) JoinChannel(channelId string) error { //nolint:golint
	m.RLock()
	defer m.RUnlock()
	for _, c := range m.Team.Channels {
		if c.Id == channelId {
			m.log.Debug("Not joining ", channelId, " already joined.")
			return nil
		}
	}
	m.log.Debug("Joining ", channelId)
	_, resp := m.Client.AddChannelMember(channelId, m.User.Id)
	if resp.Error != nil {
		return resp.Error
	}
	return nil
}

func (m *MMClient) GetPostsSince(channelId string, time int64) *model.PostList { //nolint:golint
	res, resp := m.Client.GetPostsSince(channelId, time)
	if resp.Error != nil {
		return nil
	}
	return res
}

func (m *MMClient) SearchPosts(query string) *model.PostList {
	res, resp := m.Client.SearchPosts(m.Team.Id, query, false)
	if resp.Error != nil {
		return nil
	}
	return res
}

func (m *MMClient) GetPosts(channelId string, limit int) *model.PostList { //nolint:golint
	res, resp := m.Client.GetPostsForChannel(channelId, 0, limit, "")
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

func (m *MMClient) UpdateChannelHeader(channelId string, header string) { //nolint:golint
	channel := &model.Channel{Id: channelId, Header: header}
	m.log.Debugf("updating channelheader %#v, %#v", channelId, header)
	_, resp := m.Client.UpdateChannel(channel)
	if resp.Error != nil {
		log.Error(resp.Error)
	}
}

func (m *MMClient) UpdateLastViewed(channelId string) error { //nolint:golint
	m.log.Debugf("posting lastview %#v", channelId)
	view := &model.ChannelView{ChannelId: channelId}
	_, resp := m.Client.ViewChannel(m.User.Id, view)
	if resp.Error != nil {
		m.log.Errorf("ChannelView update for %s failed: %s", channelId, resp.Error)
		return resp.Error
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
		m.log.Errorf("UsernamesInChannel(%s) failed: %s", channelId, resp.Error)
		return []string{}
	}
	allusers := m.GetUsers()
	result := []string{}
	for _, member := range *res {
		result = append(result, allusers[member.UserId].Nickname)
	}
	return result
}

func (m *MMClient) createCookieJar(token string) *cookiejar.Jar {
	var cookies []*http.Cookie
	jar, _ := cookiejar.New(nil)
	firstCookie := &http.Cookie{
		Name:   "MMAUTHTOKEN",
		Value:  token,
		Path:   "/",
		Domain: m.Credentials.Server,
	}
	cookies = append(cookies, firstCookie)
	cookieURL, _ := url.Parse("https://" + m.Credentials.Server)
	jar.SetCookies(cookieURL, cookies)
	return jar
}

// SendDirectMessage sends a direct message to specified user
func (m *MMClient) SendDirectMessage(toUserId string, msg string) { //nolint:golint
	m.SendDirectMessageProps(toUserId, msg, nil)
}

func (m *MMClient) SendDirectMessageProps(toUserId string, msg string, props map[string]interface{}) { //nolint:golint
	m.log.Debugf("SendDirectMessage to %s, msg %s", toUserId, msg)
	// create DM channel (only happens on first message)
	_, resp := m.Client.CreateDirectChannel(m.User.Id, toUserId)
	if resp.Error != nil {
		m.log.Debugf("SendDirectMessage to %#v failed: %s", toUserId, resp.Error)
		return
	}
	channelName := model.GetDMNameFromIds(toUserId, m.User.Id)

	// update our channels
	m.UpdateChannels()

	// build & send the message
	msg = strings.Replace(msg, "\r", "", -1)
	post := &model.Post{ChannelId: m.GetChannelId(channelName, m.Team.Id), Message: msg, Props: props}
	m.Client.CreatePost(post)
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

func (m *MMClient) GetLastViewedAt(channelId string) int64 { //nolint:golint
	m.RLock()
	defer m.RUnlock()
	res, resp := m.Client.GetChannelMember(channelId, m.User.Id, "")
	if resp.Error != nil {
		return model.GetMillis()
	}
	return res.LastViewedAt
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

func (m *MMClient) UpdateUser(userId string) { //nolint:golint
	m.Lock()
	defer m.Unlock()
	res, resp := m.Client.GetUser(userId, "")
	if resp.Error != nil {
		return
	}
	m.Users[userId] = res
}

func (m *MMClient) GetUserName(userId string) string { //nolint:golint
	user := m.GetUser(userId)
	if user != nil {
		return user.Username
	}
	return ""
}

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

func (m *MMClient) UpdateStatus(userId string, status string) error { //nolint:golint
	_, resp := m.Client.UpdateUserStatus(userId, &model.Status{Status: status})
	if resp.Error != nil {
		return resp.Error
	}
	return nil
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

func (m *MMClient) UploadFile(data []byte, channelId string, filename string) (string, error) { //nolint:golint
	f, resp := m.Client.UploadFile(data, channelId, filename)
	if resp.Error != nil {
		return "", resp.Error
	}
	return f.FileInfos[0].Id, nil
}

func (m *MMClient) StatusLoop() {
	retries := 0
	backoff := time.Second * 60
	if m.OnWsConnect != nil {
		m.OnWsConnect()
	}
	m.log.Debug("StatusLoop:", m.OnWsConnect != nil)
	for {
		if m.WsQuit {
			return
		}
		if m.WsConnected {
			m.checkAlive()
			select {
			case <-m.WsPingChan:
				m.log.Debug("WS PONG received")
				backoff = time.Second * 60
			case <-time.After(time.Second * 5):
				if retries > 3 {
					m.log.Debug("StatusLoop() timeout")
					m.Logout()
					m.WsQuit = false
					err := m.Login()
					if err != nil {
						log.Errorf("Login failed: %#v", err)
						break
					}
					if m.OnWsConnect != nil {
						m.OnWsConnect()
					}
					go m.WsReceiver()
				} else {
					retries++
					backoff = time.Second * 5
				}
			}
		}
		time.Sleep(backoff)
	}
}

// initialize user and teams
func (m *MMClient) initUser() error {
	m.Lock()
	defer m.Unlock()
	// we only load all team data on initial login.
	// all other updates are for channels from our (primary) team only.
	//m.log.Debug("initUser(): loading all team data")
	teams, resp := m.Client.GetTeamsForUser(m.User.Id, "")
	if resp.Error != nil {
		return resp.Error
	}
	for _, team := range teams {
		mmusers, resp := m.Client.GetUsersInTeam(team.Id, 0, 50000, "")
		if resp.Error != nil {
			return errors.New(resp.Error.DetailedError)
		}
		usermap := make(map[string]*model.User)
		for _, user := range mmusers {
			usermap[user.Id] = user
		}

		t := &Team{Team: team, Users: usermap, Id: team.Id}

		mmchannels, resp := m.Client.GetChannelsForTeamForUser(team.Id, m.User.Id, "")
		if resp.Error != nil {
			return resp.Error
		}
		t.Channels = mmchannels
		mmchannels, resp = m.Client.GetPublicChannelsForTeam(team.Id, 0, 5000, "")
		if resp.Error != nil {
			return resp.Error
		}
		t.MoreChannels = mmchannels
		m.OtherTeams = append(m.OtherTeams, t)
		if team.Name == m.Credentials.Team {
			m.Team = t
			m.log.Debugf("initUser(): found our team %s (id: %s)", team.Name, team.Id)
		}
		// add all users
		for k, v := range t.Users {
			m.Users[k] = v
		}
	}
	return nil
}

func (m *MMClient) checkAlive() error {
	// check if session still is valid
	_, resp := m.Client.GetMe("")
	if resp.Error != nil {
		return resp.Error
	}
	m.log.Debug("WS PING")
	return m.sendWSRequest("ping", nil)
}

func (m *MMClient) sendWSRequest(action string, data map[string]interface{}) error {
	req := &model.WebSocketRequest{}
	req.Seq = m.WsSequence
	req.Action = action
	req.Data = data
	m.WsSequence++
	m.log.Debugf("sendWsRequest %#v", req)
	m.WsClient.WriteJSON(req)
	return nil
}

func supportedVersion(version string) bool {
	if strings.HasPrefix(version, "3.8.0") ||
		strings.HasPrefix(version, "3.9.0") ||
		strings.HasPrefix(version, "3.10.0") ||
		strings.HasPrefix(version, "4.") ||
		strings.HasPrefix(version, "5.") {
		return true
	}
	return false
}

func digestString(s string) string {
	return fmt.Sprintf("%x", md5.Sum([]byte(s)))
}
