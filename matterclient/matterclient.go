package matterclient

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	lru "github.com/hashicorp/golang-lru"
	"github.com/jpillora/backoff"
	prefixed "github.com/matterbridge/logrus-prefixed-formatter"
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/sirupsen/logrus"
)

type Credentials struct {
	Login            string
	Team             string
	Pass             string
	Token            string
	CookieToken      bool
	Server           string
	NoTLS            bool
	SkipTLSVerify    bool
	SkipVersionCheck bool
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
	WsClient      *websocket.Conn
	WsQuit        bool
	WsAway        bool
	WsConnected   bool
	WsSequence    int64
	WsPingChan    chan *model.WebSocketResponse
	ServerVersion string
	OnWsConnect   func()

	logger     *logrus.Entry
	rootLogger *logrus.Logger
	lruCache   *lru.Cache
	allevents  bool
}

// New will instantiate a new Matterclient with the specified login details without connecting.
func New(login string, pass string, team string, server string) *MMClient {
	rootLogger := logrus.New()
	rootLogger.SetFormatter(&prefixed.TextFormatter{
		PrefixPadding: 13,
		DisableColors: true,
	})

	cred := &Credentials{
		Login:  login,
		Pass:   pass,
		Team:   team,
		Server: server,
	}

	cache, _ := lru.New(500)
	return &MMClient{
		Credentials: cred,
		MessageChan: make(chan *Message, 100),
		Users:       make(map[string]*model.User),
		rootLogger:  rootLogger,
		lruCache:    cache,
		logger:      rootLogger.WithFields(logrus.Fields{"prefix": "matterclient"}),
	}
}

// SetDebugLog activates debugging logging on all Matterclient log output.
func (m *MMClient) SetDebugLog() {
	m.rootLogger.SetFormatter(&prefixed.TextFormatter{
		PrefixPadding:   13,
		DisableColors:   true,
		FullTimestamp:   false,
		ForceFormatting: true,
	})
}

// SetLogLevel tries to parse the specified level and if successful sets
// the log level accordingly. Accepted levels are: 'debug', 'info', 'warn',
// 'error', 'fatal' and 'panic'.
func (m *MMClient) SetLogLevel(level string) {
	l, err := logrus.ParseLevel(level)
	if err != nil {
		m.logger.Warnf("Failed to parse specified log-level '%s': %#v", level, err)
	} else {
		m.rootLogger.SetLevel(l)
	}
}

func (m *MMClient) EnableAllEvents() {
	m.allevents = true
}

// Login tries to connect the client with the loging details with which it was initialized.
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

// Logout disconnects the client from the chat server.
func (m *MMClient) Logout() error {
	m.logger.Debugf("logout as %s (team: %s) on %s", m.Credentials.Login, m.Credentials.Team, m.Credentials.Server)
	m.WsQuit = true
	m.WsClient.Close()
	m.WsClient.UnderlyingConn().Close()
	if strings.Contains(m.Credentials.Pass, model.SESSION_COOKIE_TOKEN) {
		m.logger.Debug("Not invalidating session in logout, credential is a token")
		return nil
	}
	_, resp := m.Client.Logout()
	if resp.Error != nil {
		return resp.Error
	}
	return nil
}

// WsReceiver implements the core loop that manages the connection to the chat server. In
// case of a disconnect it will try to reconnect. A call to this method is blocking until
// the 'WsQuite' field of the MMClient object is set to 'true'.
func (m *MMClient) WsReceiver() {
	for {
		var rawMsg json.RawMessage
		var err error

		if m.WsQuit {
			m.logger.Debug("exiting WsReceiver")
			return
		}

		if !m.WsConnected {
			time.Sleep(time.Millisecond * 100)
			continue
		}

		if _, rawMsg, err = m.WsClient.ReadMessage(); err != nil {
			m.logger.Error("error:", err)
			// reconnect
			m.wsConnect()
		}

		var event model.WebSocketEvent
		if err := json.Unmarshal(rawMsg, &event); err == nil && event.IsValid() {
			m.logger.Debugf("WsReceiver event: %#v", event)
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
					continue
				}
			}
			if m.allevents {
				m.MessageChan <- msg
				continue
			}
			switch msg.Raw.Event {
			case model.WEBSOCKET_EVENT_USER_ADDED,
				model.WEBSOCKET_EVENT_USER_REMOVED,
				model.WEBSOCKET_EVENT_CHANNEL_CREATED,
				model.WEBSOCKET_EVENT_CHANNEL_DELETED:
				m.MessageChan <- msg
				continue
			}
		}

		var response model.WebSocketResponse
		if err := json.Unmarshal(rawMsg, &response); err == nil && response.IsValid() {
			m.logger.Debugf("WsReceiver response: %#v", response)
			m.parseResponse(response)
		}
	}
}

// StatusLoop implements a ping-cycle that ensures that the connection to the chat servers
// remains alive. In case of a disconnect it will try to reconnect. A call to this method
// is blocking until the 'WsQuite' field of the MMClient object is set to 'true'.
func (m *MMClient) StatusLoop() {
	retries := 0
	backoff := time.Second * 60
	if m.OnWsConnect != nil {
		m.OnWsConnect()
	}
	m.logger.Debug("StatusLoop:", m.OnWsConnect != nil)
	for {
		if m.WsQuit {
			return
		}
		if m.WsConnected {
			if err := m.checkAlive(); err != nil {
				m.logger.Errorf("Connection is not alive: %#v", err)
			}
			select {
			case <-m.WsPingChan:
				m.logger.Debug("WS PONG received")
				backoff = time.Second * 60
			case <-time.After(time.Second * 5):
				if retries > 3 {
					m.logger.Debug("StatusLoop() timeout")
					m.Logout()
					m.WsQuit = false
					err := m.Login()
					if err != nil {
						m.logger.Errorf("Login failed: %#v", err)
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
