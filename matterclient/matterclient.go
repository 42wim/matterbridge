package matterclient

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/hashicorp/golang-lru"
	"github.com/jpillora/backoff"
	prefixed "github.com/matterbridge/logrus-prefixed-formatter"
	"github.com/mattermost/mattermost-server/model"
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

	// do initialization setup
	if err := m.initClient(firstConnection, b); err != nil {
		return err
	}

	if err := m.doLogin(firstConnection, b); err != nil {
		return err
	}

	if err := m.initUser(); err != nil {
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
