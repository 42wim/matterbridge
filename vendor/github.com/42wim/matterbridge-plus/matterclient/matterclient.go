package matterclient

import (
	"errors"
	log "github.com/Sirupsen/logrus"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/jpillora/backoff"
	"github.com/mattermost/platform/model"
)

type Credentials struct {
	Login  string
	Team   string
	Pass   string
	Server string
	NoTLS  bool
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
	var myinfo *model.Result
	var appErr *model.AppError
	var logmsg = "trying login"
	for {
		m.log.Debugf(logmsg+" %s %s %s", m.Credentials.Team, m.Credentials.Login, m.Credentials.Server)
		myinfo, appErr = m.Client.LoginByEmail(m.Credentials.Team, m.Credentials.Login, m.Credentials.Pass)
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
			m.log.Debug("LOGIN: %s, reconnecting in %s", appErr, d)
			time.Sleep(d)
			logmsg = "retrying login"
			continue
		}
		break
	}
	// reset timer
	b.Reset()
	m.User = myinfo.Data.(*model.User)
	myinfo, _ = m.Client.GetMyTeam("")
	m.Team = myinfo.Data.(*model.Team)

	// setup websocket connection
	wsurl := wsScheme + m.Credentials.Server + "/api/v1/websocket"
	header := http.Header{}
	header.Set(model.HEADER_AUTH, "BEARER "+m.Client.AuthToken)

	var WsClient *websocket.Conn
	var err error
	for {
		WsClient, _, err = websocket.DefaultDialer.Dial(wsurl, header)
		if err != nil {
			d := b.Duration()
			log.Printf("WSS: %s, reconnecting in %s", err, d)
			time.Sleep(d)
			continue
		}
		break
	}
	b.Reset()

	m.WsClient = WsClient

	// populating users
	m.UpdateUsers()

	// populating channels
	m.UpdateChannels()

	return nil
}

func (m *MMClient) WsReceiver() {
	var rmsg model.Message
	for {
		if err := m.WsClient.ReadJSON(&rmsg); err != nil {
			log.Println("error:", err)
			// reconnect
			m.Login()
		}
		//log.Printf("WsReceiver: %#v", rmsg)
		msg := &Message{Raw: &rmsg, Team: m.Credentials.Team}
		m.parseMessage(msg)
		m.MessageChan <- msg
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
	mmusers, _ := m.Client.GetProfiles(m.User.TeamId, "")
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
	res, err := m.Client.SearchPosts(query)
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

func (m *MMClient) UpdateChannelHeader(channelId string, header string) {
	data := make(map[string]string)
	data["channel_id"] = channelId
	data["channel_header"] = header
	log.Printf("updating channelheader %#v, %#v", channelId, header)
	_, err := m.Client.UpdateChannelHeader(data)
	if err != nil {
		log.Print(err)
	}
}

func (m *MMClient) UpdateLastViewed(channelId string) {
	log.Printf("posting lastview %#v", channelId)
	_, err := m.Client.UpdateLastViewedAt(channelId)
	if err != nil {
		log.Print(err)
	}
}

func (m *MMClient) UsernamesInChannel(channelName string) []string {
	ceiRes, err := m.Client.GetChannelExtraInfo(m.GetChannelId(channelName), 5000, "")
	if err != nil {
		log.Errorf("UsernamesInChannel(%s) failed: %s", channelName, err)
		return []string{}
	}
	extra := ceiRes.Data.(*model.ChannelExtra)
	result := []string{}
	for _, member := range extra.Members {
		result = append(result, member.Username)
	}
	return result
}
