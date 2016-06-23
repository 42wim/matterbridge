package matterclient

import (
	"crypto/tls"
	"errors"
	log "github.com/Sirupsen/logrus"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"

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
	Raw      *model.Message
	Post     *model.Post
	Team     string
	Channel  string
	Username string
	Text     string
}

type MMClient struct {
	*Credentials
	Client       *model.Client
	WsClient     *websocket.Conn
	WsQuit       bool
	WsAway       bool
	Channels     *model.ChannelList
	MoreChannels *model.ChannelList
	User         *model.User
	Users        map[string]*model.User
	MessageChan  chan *Message
	Team         *model.Team
	log          *log.Entry
}

func New(login, pass, team, server string) *MMClient {
	cred := &Credentials{Login: login, Pass: pass, Team: team, Server: server}
	mmclient := &MMClient{Credentials: cred, MessageChan: make(chan *Message, 100)}
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
	m.Client.HttpClient.Transport = &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: m.SkipTLSVerify}}
	var myinfo *model.Result
	var appErr *model.AppError
	var logmsg = "trying login"
	for {
		m.log.Debugf("%s %s %s %s", logmsg, m.Credentials.Team, m.Credentials.Login, m.Credentials.Server)
		if strings.Contains(m.Credentials.Pass, model.SESSION_COOKIE_TOKEN) {
			m.log.Debugf(logmsg+" with ", model.SESSION_COOKIE_TOKEN)
			token := strings.Split(m.Credentials.Pass, model.SESSION_COOKIE_TOKEN+"=")
			m.Client.HttpClient.Jar = m.createCookieJar(token[1])
			m.Client.MockSession(token[1])
			myinfo, appErr = m.Client.GetMe("")
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
			if !strings.Contains(appErr.DetailedError, "connection refused") &&
				!strings.Contains(appErr.DetailedError, "invalid character") {
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

	initLoad, _ := m.Client.GetInitialLoad()
	initData := initLoad.Data.(*model.InitialLoad)
	m.User = initData.User
	for _, v := range initData.Teams {
		m.log.Debugf("trying %s (id: %s)", v.Name, v.Id)
		if v.Name == m.Credentials.Team {
			m.Client.SetTeamId(v.Id)
			m.Team = v
			m.log.Debugf("GetallTeamListings: found id %s for team %s", v.Id, v.Name)
			break
		}
	}
	if m.Team == nil {
		return errors.New("team not found")
	}

	// setup websocket connection
	wsurl := wsScheme + m.Credentials.Server + "/api/v3/users/websocket"
	header := http.Header{}
	header.Set(model.HEADER_AUTH, "BEARER "+m.Client.AuthToken)

	m.log.Debug("WsClient: making connection")
	var err error
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

	// populating users
	m.UpdateUsers()

	// populating channels
	m.UpdateChannels()

	return nil
}

func (m *MMClient) WsReceiver() {
	var rmsg model.Message
	for {
		if m.WsQuit {
			m.log.Debug("exiting WsReceiver")
			return
		}
		if err := m.WsClient.ReadJSON(&rmsg); err != nil {
			m.log.Error("error:", err)
			// reconnect
			m.Login()
		}
		if rmsg.Action == "ping" {
			m.handleWsPing()
			continue
		}
		msg := &Message{Raw: &rmsg, Team: m.Credentials.Team}
		m.parseMessage(msg)
		m.MessageChan <- msg
	}

}

func (m *MMClient) handleWsPing() {
	m.log.Debug("Ws PING")
	if !m.WsQuit && !m.WsAway {
		m.log.Debug("Ws PONG")
		m.WsClient.WriteMessage(websocket.PongMessage, []byte{})
	}
}

func (m *MMClient) parseMessage(rmsg *Message) {
	switch rmsg.Raw.Action {
	case model.ACTION_POSTED:
		m.parseActionPost(rmsg)
		/*
			case model.ACTION_USER_REMOVED:
				m.handleWsActionUserRemoved(&rmsg)
			case model.ACTION_USER_ADDED:
				m.handleWsActionUserAdded(&rmsg)
		*/
	}
}

func (m *MMClient) parseActionPost(rmsg *Message) {
	data := model.PostFromJson(strings.NewReader(rmsg.Raw.Props["post"]))
	//	log.Println("receiving userid", data.UserId)
	// we don't have the user, refresh the userlist
	if m.Users[data.UserId] == nil {
		m.UpdateUsers()
	}
	rmsg.Username = m.Users[data.UserId].Username
	rmsg.Channel = m.GetChannelName(data.ChannelId)
	// direct message
	if strings.Contains(rmsg.Channel, "__") {
		//log.Println("direct message")
		rcvusers := strings.Split(rmsg.Channel, "__")
		if rcvusers[0] != m.User.Id {
			rmsg.Channel = m.Users[rcvusers[0]].Username
		} else {
			rmsg.Channel = m.Users[rcvusers[1]].Username
		}
	}
	rmsg.Text = data.Message
	rmsg.Post = data
	return
}

func (m *MMClient) UpdateUsers() error {
	mmusers, _ := m.Client.GetProfilesForDirectMessageList(m.Team.Id)
	m.Users = mmusers.Data.(map[string]*model.User)
	return nil
}

func (m *MMClient) UpdateChannels() error {
	mmchannels, _ := m.Client.GetChannels("")
	m.Channels = mmchannels.Data.(*model.ChannelList)
	mmchannels, _ = m.Client.GetMoreChannels("")
	m.MoreChannels = mmchannels.Data.(*model.ChannelList)
	return nil
}

func (m *MMClient) GetChannelName(id string) string {
	for _, channel := range append(m.Channels.Channels, m.MoreChannels.Channels...) {
		if channel.Id == id {
			return channel.Name
		}
	}
	// not found? could be a new direct message from mattermost. Try to update and check again
	m.UpdateChannels()
	for _, channel := range append(m.Channels.Channels, m.MoreChannels.Channels...) {
		if channel.Id == id {
			return channel.Name
		}
	}
	return ""
}

func (m *MMClient) GetChannelId(name string) string {
	for _, channel := range append(m.Channels.Channels, m.MoreChannels.Channels...) {
		if channel.Name == name {
			return channel.Id
		}
	}
	return ""
}

func (m *MMClient) GetChannelHeader(id string) string {
	for _, channel := range append(m.Channels.Channels, m.MoreChannels.Channels...) {
		if channel.Id == id {
			return channel.Header
		}
	}
	return ""
}

func (m *MMClient) PostMessage(channel string, text string) {
	post := &model.Post{ChannelId: m.GetChannelId(channel), Message: text}
	m.Client.CreatePost(post)
}

func (m *MMClient) JoinChannel(channel string) error {
	cleanChan := strings.Replace(channel, "#", "", 1)
	if m.GetChannelId(cleanChan) == "" {
		return errors.New("failed to join")
	}
	for _, c := range m.Channels.Channels {
		if c.Name == cleanChan {
			m.log.Debug("Not joining ", cleanChan, " already joined.")
			return nil
		}
	}
	m.log.Debug("Joining ", cleanChan)
	_, err := m.Client.JoinChannel(m.GetChannelId(cleanChan))
	if err != nil {
		return errors.New("failed to join")
	}
	//	m.SyncChannel(m.getMMChannelId(strings.Replace(channel, "#", "", 1)), strings.Replace(channel, "#", "", 1))
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
	return res.Data.(string)
}

func (m *MMClient) GetPublicLinks(filenames []string) []string {
	var output []string
	for _, f := range filenames {
		res, err := m.Client.GetPublicLink(f)
		if err != nil {
			continue
		}
		output = append(output, res.Data.(string))
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
	_, err := m.Client.UpdateLastViewedAt(channelId)
	if err != nil {
		m.log.Error(err)
	}
}

func (m *MMClient) UsernamesInChannel(channelName string) []string {
	ceiRes, err := m.Client.GetChannelExtraInfo(m.GetChannelId(channelName), 5000, "")
	if err != nil {
		m.log.Errorf("UsernamesInChannel(%s) failed: %s", channelName, err)
		return []string{}
	}
	extra := ceiRes.Data.(*model.ChannelExtra)
	result := []string{}
	for _, member := range extra.Members {
		result = append(result, member.Username)
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

func (m *MMClient) GetOtherUserDM(channel string) *model.User {
	m.UpdateUsers()
	var rcvuser *model.User
	if strings.Contains(channel, "__") {
		rcvusers := strings.Split(channel, "__")
		if rcvusers[0] != m.User.Id {
			rcvuser = m.Users[rcvusers[0]]
		} else {
			rcvuser = m.Users[rcvusers[1]]
		}
	}
	return rcvuser
}

func (m *MMClient) SendDirectMessage(toUserId string, msg string) {
	m.log.Debugf("SendDirectMessage to %s, msg %s", toUserId, msg)
	var channel string
	// We don't have a DM with this user yet.
	if m.GetChannelId(toUserId+"__"+m.User.Id) == "" && m.GetChannelId(m.User.Id+"__"+toUserId) == "" {
		// create DM channel
		_, err := m.Client.CreateDirectChannel(toUserId)
		if err != nil {
			m.log.Debugf("SendDirectMessage to %#v failed: %s", toUserId, err)
		}
		// update our channels
		mmchannels, _ := m.Client.GetChannels("")
		m.Channels = mmchannels.Data.(*model.ChannelList)
	}

	// build the channel name
	if toUserId > m.User.Id {
		channel = m.User.Id + "__" + toUserId
	} else {
		channel = toUserId + "__" + m.User.Id
	}
	// build & send the message
	msg = strings.Replace(msg, "\r", "", -1)
	post := &model.Post{ChannelId: m.GetChannelId(channel), Message: msg}
	m.Client.CreatePost(post)
}
