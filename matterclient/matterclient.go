package matterclient

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/gorilla/websocket"
	"github.com/jpillora/backoff"
	"github.com/mattermost/platform/model"
)

type Credentials struct {
	Login         string
	Team          string
	Pass          string
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
}

type Team struct {
	Team         *model.Team
	Id           string
	Channels     *model.ChannelList
	MoreChannels *model.ChannelList
	Users        map[string]*model.User
}

type MMClient struct {
	sync.RWMutex
	*Credentials
	Team          *Team
	OtherTeams    []*Team
	Client        *model.Client
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
}

func New(login, pass, team, server string) *MMClient {
	cred := &Credentials{Login: login, Pass: pass, Team: team, Server: server}
	mmclient := &MMClient{Credentials: cred, MessageChan: make(chan *Message, 100), Users: make(map[string]*model.User)}
	mmclient.log = log.WithFields(log.Fields{"module": "matterclient"})
	log.SetFormatter(&log.TextFormatter{FullTimestamp: true})
	return mmclient
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
	if m.WsConnected == true {
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
	wsScheme := "wss://"
	if m.NoTLS {
		uriScheme = "http://"
		wsScheme = "ws://"
	}
	// login to mattermost
	m.Client = model.NewClient(uriScheme + m.Credentials.Server)
	m.Client.HttpClient.Transport = &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: m.SkipTLSVerify}, Proxy: http.ProxyFromEnvironment}
	m.Client.HttpClient.Timeout = time.Second * 10
	// bogus call to get the serverversion
	m.Client.GetClientProperties()
	if firstConnection && !supportedVersion(m.Client.ServerVersion) {
		return fmt.Errorf("unsupported mattermost version: %s", m.Client.ServerVersion)
	}
	m.ServerVersion = m.Client.ServerVersion
	m.log.Infof("Found version %s", m.ServerVersion)

	var myinfo *model.Result
	var appErr *model.AppError
	var logmsg = "trying login"
	for {
		m.log.Debugf("%s %s %s %s", logmsg, m.Credentials.Team, m.Credentials.Login, m.Credentials.Server)
		if strings.Contains(m.Credentials.Pass, model.SESSION_COOKIE_TOKEN) {
			m.log.Debugf(logmsg+" with %s", model.SESSION_COOKIE_TOKEN)
			token := strings.Split(m.Credentials.Pass, model.SESSION_COOKIE_TOKEN+"=")
			if len(token) != 2 {
				return errors.New("incorrect MMAUTHTOKEN. valid input is MMAUTHTOKEN=yourtoken")
			}
			m.Client.HttpClient.Jar = m.createCookieJar(token[1])
			m.Client.MockSession(token[1])
			myinfo, appErr = m.Client.GetMe("")
			if appErr != nil {
				return errors.New(appErr.DetailedError)
			}
			if myinfo.Data.(*model.User) == nil {
				m.log.Errorf("LOGIN TOKEN: %s is invalid", m.Credentials.Pass)
				return errors.New("invalid " + model.SESSION_COOKIE_TOKEN)
			}
		} else {
			myinfo, appErr = m.Client.Login(m.Credentials.Login, m.Credentials.Pass)
		}
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
		return errors.New("team not found")
	}
	// set our team id as default route
	m.Client.SetTeamId(m.Team.Id)

	// setup websocket connection
	wsurl := wsScheme + m.Credentials.Server + model.API_URL_SUFFIX_V3 + "/users/websocket"
	header := http.Header{}
	header.Set(model.HEADER_AUTH, "BEARER "+m.Client.AuthToken)

	m.log.Debugf("WsClient: making connection: %s", wsurl)
	for {
		wsDialer := &websocket.Dialer{Proxy: http.ProxyFromEnvironment, TLSClientConfig: &tls.Config{InsecureSkipVerify: m.SkipTLSVerify}}
		m.WsClient, _, err = wsDialer.Dial(wsurl, header)
		if err != nil {
			d := b.Duration()
			m.log.Debugf("WSS: %s, reconnecting in %s", err, d)
			time.Sleep(d)
			continue
		}
		break
	}
	b.Reset()

	m.log.Debug("WsClient: connected")
	m.WsSequence = 1
	m.WsPingChan = make(chan *model.WebSocketResponse)
	// only start to parse WS messages when login is completely done
	m.WsConnected = true

	return nil
}

func (m *MMClient) Logout() error {
	m.log.Debugf("logout as %s (team: %s) on %s", m.Credentials.Login, m.Credentials.Team, m.Credentials.Server)
	m.WsQuit = true
	m.WsClient.Close()
	m.WsClient.UnderlyingConn().Close()
	_, err := m.Client.Logout()
	if err != nil {
		return err
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
			m.Login()
		}

		var event model.WebSocketEvent
		if err := json.Unmarshal(rawMsg, &event); err == nil && event.IsValid() {
			m.log.Debugf("WsReceiver: %#v", event)
			msg := &Message{Raw: &event, Team: m.Credentials.Team}
			m.parseMessage(msg)
			m.MessageChan <- msg
			continue
		}

		var response model.WebSocketResponse
		if err := json.Unmarshal(rawMsg, &response); err == nil && response.IsValid() {
			m.log.Debugf("WsReceiver: %#v", response)
			m.parseResponse(response)
			continue
		}
	}
}

func (m *MMClient) parseMessage(rmsg *Message) {
	switch rmsg.Raw.Event {
	case model.WEBSOCKET_EVENT_POSTED, model.WEBSOCKET_EVENT_POST_EDITED:
		m.parseActionPost(rmsg)
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
	data := model.PostFromJson(strings.NewReader(rmsg.Raw.Data["post"].(string)))
	// we don't have the user, refresh the userlist
	if m.GetUser(data.UserId) == nil {
		m.UpdateUsers()
	}
	rmsg.Username = m.GetUser(data.UserId).Username
	rmsg.Channel = m.GetChannelName(data.ChannelId)
	rmsg.Type = data.Type
	teamid, _ := rmsg.Raw.Data["team_id"].(string)
	// edit messsages have no team_id for some reason
	if teamid == "" {
		// we can find the team_id from the channelid
		result, _ := m.Client.GetChannel(data.ChannelId, "")
		teamid = result.Data.(*model.ChannelData).Channel.TeamId
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
	return
}

func (m *MMClient) UpdateUsers() error {
	mmusers, err := m.Client.GetProfiles(0, 50000, "")
	if err != nil {
		return errors.New(err.DetailedError)
	}
	m.Lock()
	m.Users = mmusers.Data.(map[string]*model.User)
	m.Unlock()
	return nil
}

func (m *MMClient) UpdateChannels() error {
	mmchannels, err := m.Client.GetChannels("")
	if err != nil {
		return errors.New(err.DetailedError)
	}
	var mmchannels2 *model.Result
	if m.mmVersion() >= 3.8 {
		mmchannels2, err = m.Client.GetMoreChannelsPage(0, 5000)
	} else {
		mmchannels2, err = m.Client.GetMoreChannels("")
	}
	if err != nil {
		return errors.New(err.DetailedError)
	}
	m.Lock()
	m.Team.Channels = mmchannels.Data.(*model.ChannelList)
	m.Team.MoreChannels = mmchannels2.Data.(*model.ChannelList)
	m.Unlock()
	return nil
}

func (m *MMClient) GetChannelName(channelId string) string {
	m.RLock()
	defer m.RUnlock()
	for _, t := range m.OtherTeams {
		for _, channel := range append(*t.Channels, *t.MoreChannels...) {
			if channel.Id == channelId {
				return channel.Name
			}
		}
	}
	return ""
}

func (m *MMClient) GetChannelId(name string, teamId string) string {
	m.RLock()
	defer m.RUnlock()
	if teamId == "" {
		teamId = m.Team.Id
	}
	for _, t := range m.OtherTeams {
		if t.Id == teamId {
			for _, channel := range append(*t.Channels, *t.MoreChannels...) {
				if channel.Name == name {
					return channel.Id
				}
			}
		}
	}
	return ""
}

func (m *MMClient) GetChannelHeader(channelId string) string {
	m.RLock()
	defer m.RUnlock()
	for _, t := range m.OtherTeams {
		for _, channel := range append(*t.Channels, *t.MoreChannels...) {
			if channel.Id == channelId {
				return channel.Header
			}

		}
	}
	return ""
}

func (m *MMClient) PostMessage(channelId string, text string) {
	post := &model.Post{ChannelId: channelId, Message: text}
	m.Client.CreatePost(post)
}

func (m *MMClient) JoinChannel(channelId string) error {
	m.RLock()
	defer m.RUnlock()
	for _, c := range *m.Team.Channels {
		if c.Id == channelId {
			m.log.Debug("Not joining ", channelId, " already joined.")
			return nil
		}
	}
	m.log.Debug("Joining ", channelId)
	_, err := m.Client.JoinChannel(channelId)
	if err != nil {
		return errors.New("failed to join")
	}
	return nil
}

func (m *MMClient) GetPostsSince(channelId string, time int64) *model.PostList {
	res, err := m.Client.GetPostsSince(channelId, time)
	if err != nil {
		return nil
	}
	return res.Data.(*model.PostList)
}

func (m *MMClient) SearchPosts(query string) *model.PostList {
	res, err := m.Client.SearchPosts(query, false)
	if err != nil {
		return nil
	}
	return res.Data.(*model.PostList)
}

func (m *MMClient) GetPosts(channelId string, limit int) *model.PostList {
	res, err := m.Client.GetPosts(channelId, 0, limit, "")
	if err != nil {
		return nil
	}
	return res.Data.(*model.PostList)
}

func (m *MMClient) GetPublicLink(filename string) string {
	res, err := m.Client.GetPublicLink(filename)
	if err != nil {
		return ""
	}
	return res
}

func (m *MMClient) GetPublicLinks(filenames []string) []string {
	var output []string
	for _, f := range filenames {
		res, err := m.Client.GetPublicLink(f)
		if err != nil {
			continue
		}
		output = append(output, res)
	}
	return output
}

func (m *MMClient) UpdateChannelHeader(channelId string, header string) {
	data := make(map[string]string)
	data["channel_id"] = channelId
	data["channel_header"] = header
	m.log.Debugf("updating channelheader %#v, %#v", channelId, header)
	_, err := m.Client.UpdateChannelHeader(data)
	if err != nil {
		log.Error(err)
	}
}

func (m *MMClient) UpdateLastViewed(channelId string) {
	m.log.Debugf("posting lastview %#v", channelId)
	if m.mmVersion() >= 3.8 {
		view := model.ChannelView{ChannelId: channelId}
		res, _ := m.Client.ViewChannel(view)
		if res == false {
			m.log.Errorf("ChannelView update for %s failed", channelId)
		}
		return
	}
	_, err := m.Client.UpdateLastViewedAt(channelId, true)
	if err != nil {
		m.log.Error(err)
	}
}

func (m *MMClient) UsernamesInChannel(channelId string) []string {
	res, err := m.Client.GetMyChannelMembers()
	if err != nil {
		m.log.Errorf("UsernamesInChannel(%s) failed: %s", channelId, err)
		return []string{}
	}
	members := res.Data.(*model.ChannelMembers)
	result := []string{}
	for _, channel := range *members {
		if channel.ChannelId == channelId {
			result = append(result, m.GetUser(channel.UserId).Username)
		}
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
func (m *MMClient) SendDirectMessage(toUserId string, msg string) {
	m.log.Debugf("SendDirectMessage to %s, msg %s", toUserId, msg)
	// create DM channel (only happens on first message)
	_, err := m.Client.CreateDirectChannel(toUserId)
	if err != nil {
		m.log.Debugf("SendDirectMessage to %#v failed: %s", toUserId, err)
		return
	}
	channelName := model.GetDMNameFromIds(toUserId, m.User.Id)

	// update our channels
	mmchannels, err := m.Client.GetChannels("")
	if err != nil {
		m.log.Debug("SendDirectMessage: Couldn't update channels")
		return
	}
	m.Lock()
	m.Team.Channels = mmchannels.Data.(*model.ChannelList)
	m.Unlock()

	// build & send the message
	msg = strings.Replace(msg, "\r", "", -1)
	post := &model.Post{ChannelId: m.GetChannelId(channelName, ""), Message: msg}
	m.Client.CreatePost(post)
}

// GetTeamName returns the name of the specified teamId
func (m *MMClient) GetTeamName(teamId string) string {
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
	channels = append(channels, *m.Team.Channels...)
	for _, t := range m.OtherTeams {
		if t.Id != m.Team.Id {
			channels = append(channels, *t.Channels...)
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
		channels = append(channels, *t.MoreChannels...)
	}
	return channels
}

// GetTeamFromChannel returns teamId belonging to channel (DM channels have no teamId).
func (m *MMClient) GetTeamFromChannel(channelId string) string {
	m.RLock()
	defer m.RUnlock()
	var channels []*model.Channel
	for _, t := range m.OtherTeams {
		channels = append(channels, *t.Channels...)
		channels = append(channels, *t.MoreChannels...)
		for _, c := range channels {
			if c.Id == channelId {
				return t.Id
			}
		}
	}
	return ""
}

func (m *MMClient) GetLastViewedAt(channelId string) int64 {
	m.RLock()
	defer m.RUnlock()
	res, err := m.Client.GetChannel(channelId, "")
	if err != nil {
		return model.GetMillis()
	}
	data := res.Data.(*model.ChannelData)
	return data.Member.LastViewedAt
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

func (m *MMClient) GetUser(userId string) *model.User {
	m.RLock()
	defer m.RUnlock()
	return m.Users[userId]
}

func (m *MMClient) GetStatus(userId string) string {
	res, err := m.Client.GetStatuses()
	if err != nil {
		return ""
	}
	status := res.Data.(map[string]string)
	if status[userId] == model.STATUS_AWAY {
		return "away"
	}
	if status[userId] == model.STATUS_ONLINE {
		return "online"
	}
	return "offline"
}

func (m *MMClient) GetStatuses() map[string]string {
	var ok bool
	statuses := make(map[string]string)
	res, err := m.Client.GetStatuses()
	if err != nil {
		return statuses
	}
	if statuses, ok = res.Data.(map[string]string); ok {
		for userId, status := range statuses {
			statuses[userId] = "offline"
			if status == model.STATUS_AWAY {
				statuses[userId] = "away"
			}
			if status == model.STATUS_ONLINE {
				statuses[userId] = "online"
			}
		}
	}
	return statuses
}

func (m *MMClient) GetTeamId() string {
	return m.Team.Id
}

func (m *MMClient) StatusLoop() {
	for {
		if m.WsQuit {
			return
		}
		if m.WsConnected {
			m.log.Debug("WS PING")
			m.sendWSRequest("ping", nil)
			select {
			case <-m.WsPingChan:
				m.log.Debug("WS PONG received")
			case <-time.After(time.Second * 5):
				m.Logout()
				m.WsQuit = false
				m.Login()
				go m.WsReceiver()
			}
		}
		time.Sleep(time.Second * 60)
	}
}

// initialize user and teams
func (m *MMClient) initUser() error {
	m.Lock()
	defer m.Unlock()
	initLoad, err := m.Client.GetInitialLoad()
	if err != nil {
		return err
	}
	initData := initLoad.Data.(*model.InitialLoad)
	m.User = initData.User
	// we only load all team data on initial login.
	// all other updates are for channels from our (primary) team only.
	//m.log.Debug("initUser(): loading all team data")
	for _, v := range initData.Teams {
		m.Client.SetTeamId(v.Id)
		mmusers, err := m.Client.GetProfiles(0, 50000, "")
		if err != nil {
			return errors.New(err.DetailedError)
		}
		t := &Team{Team: v, Users: mmusers.Data.(map[string]*model.User), Id: v.Id}
		mmchannels, err := m.Client.GetChannels("")
		if err != nil {
			return errors.New(err.DetailedError)
		}
		t.Channels = mmchannels.Data.(*model.ChannelList)
		if m.mmVersion() >= 3.8 {
			mmchannels, err = m.Client.GetMoreChannelsPage(0, 5000)
		} else {
			mmchannels, err = m.Client.GetMoreChannels("")
		}
		if err != nil {
			return errors.New(err.DetailedError)
		}
		t.MoreChannels = mmchannels.Data.(*model.ChannelList)
		m.OtherTeams = append(m.OtherTeams, t)
		if v.Name == m.Credentials.Team {
			m.Team = t
			m.log.Debugf("initUser(): found our team %s (id: %s)", v.Name, v.Id)
		}
		// add all users
		for k, v := range t.Users {
			m.Users[k] = v
		}
	}
	return nil
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

func (m *MMClient) mmVersion() float64 {
	v, _ := strconv.ParseFloat(m.ServerVersion[0:3], 64)
	return v
}

func supportedVersion(version string) bool {
	if strings.HasPrefix(version, "3.5.0") ||
		strings.HasPrefix(version, "3.6.0") ||
		strings.HasPrefix(version, "3.7.0") ||
		strings.HasPrefix(version, "3.8.0") {
		return true
	}
	return false
}
