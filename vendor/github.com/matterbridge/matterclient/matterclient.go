package matterclient

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	lru "github.com/hashicorp/golang-lru"
	"github.com/jpillora/backoff"
	prefixed "github.com/matterbridge/logrus-prefixed-formatter"
	"github.com/mattermost/mattermost-server/v6/model"
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
	MFAToken         string
}

type Team struct {
	Team         *model.Team
	ID           string
	Channels     []*model.Channel
	MoreChannels []*model.Channel
	Users        map[string]*model.User
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

type Client struct {
	sync.RWMutex
	*Credentials

	Team          *Team
	OtherTeams    []*Team
	Client        *model.Client4
	User          *model.User
	Users         map[string]*model.User
	MessageChan   chan *Message
	WsClient      *model.WebSocketClient
	AntiIdle      bool
	AntiIdleChan  string
	AntiIdleIntvl int
	WsQuit        bool
	WsConnected   bool
	OnWsConnect   func()
	reconnectBusy bool
	Timeout       int

	logger      *logrus.Entry
	rootLogger  *logrus.Logger
	lruCache    *lru.Cache
	aliveChan   chan bool
	loginCancel context.CancelFunc
	lastPong    time.Time
}

var Matterircd bool

func New(login string, pass string, team string, server string, mfatoken string) *Client {
	rootLogger := logrus.New()
	rootLogger.SetFormatter(&prefixed.TextFormatter{
		PrefixPadding: 13,
		DisableColors: true,
		FullTimestamp: true,
	})

	cred := &Credentials{
		Login:    login,
		Pass:     pass,
		Team:     team,
		Server:   server,
		MFAToken: mfatoken,
	}

	cache, _ := lru.New(500)

	return &Client{
		Credentials: cred,
		MessageChan: make(chan *Message, 100),
		Users:       make(map[string]*model.User),
		rootLogger:  rootLogger,
		lruCache:    cache,
		logger:      rootLogger.WithFields(logrus.Fields{"prefix": "matterclient"}),
		aliveChan:   make(chan bool),
	}
}

// Login tries to connect the client with the loging details with which it was initialized.
func (m *Client) Login() error {
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
	if err := m.initClient(b); err != nil {
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

	if err := m.initUserChannels(); err != nil {
		return err
	}

	// connect websocket
	m.wsConnect()

	ctx, loginCancel := context.WithCancel(context.Background())
	m.loginCancel = loginCancel

	m.logger.Debug("starting wsreceiver")

	go m.WsReceiver(ctx)

	if m.OnWsConnect != nil {
		m.logger.Debug("executing OnWsConnect()")

		go m.OnWsConnect()
	}

	go m.checkConnection(ctx)

	if m.AntiIdle {
		if m.AntiIdleChan == "" {
			// do anti idle on town-square, every installation should have this channel
			m.AntiIdleChan = "town-square"
		}

		channels := m.GetChannels()
		for _, channel := range channels {
			if channel.Name == m.AntiIdleChan {
				go m.antiIdle(ctx, channel.Id, m.AntiIdleIntvl)

				continue
			}
		}
	}

	return nil
}

func (m *Client) Reconnect() {
	if m.reconnectBusy {
		return
	}

	m.reconnectBusy = true

	m.logger.Info("reconnect: logout")
	m.reconnectLogout()

	for {
		m.logger.Info("reconnect: login")

		err := m.Login()
		if err != nil {
			m.logger.Errorf("reconnect: login failed: %s, retrying in 10 seconds", err)
			time.Sleep(time.Second * 10)

			continue
		}

		break
	}

	m.logger.Info("reconnect successful")

	m.reconnectBusy = false
}

func (m *Client) initClient(b *backoff.Backoff) error {
	uriScheme := "https://"
	if m.NoTLS {
		uriScheme = "http://"
	}
	// login to mattermost
	m.Client = model.NewAPIv4Client(uriScheme + m.Credentials.Server)
	m.Client.HTTPClient.Transport = &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: m.SkipTLSVerify, //nolint:gosec
		},
		Proxy: http.ProxyFromEnvironment,
	}

	if m.Timeout == 0 {
		m.Timeout = 10
	}

	m.Client.HTTPClient.Timeout = time.Second * time.Duration(m.Timeout)

	// handle MMAUTHTOKEN and personal token
	if err := m.handleLoginToken(); err != nil {
		return err
	}

	// check if server alive, retry until
	if err := m.serverAlive(b); err != nil {
		return err
	}

	return nil
}

func (m *Client) handleLoginToken() error {
	switch {
	case strings.Contains(m.Credentials.Pass, model.SessionCookieToken):
		token := strings.Split(m.Credentials.Pass, model.SessionCookieToken+"=")
		if len(token) != 2 {
			return errors.New("incorrect MMAUTHTOKEN. valid input is MMAUTHTOKEN=yourtoken")
		}

		m.Credentials.Token = token[1]
		m.Credentials.CookieToken = true
	case strings.Contains(m.Credentials.Pass, "token="):
		token := strings.Split(m.Credentials.Pass, "token=")
		if len(token) != 2 {
			return errors.New("incorrect personal token. valid input is token=yourtoken")
		}

		m.Credentials.Token = token[1]
	}

	return nil
}

func (m *Client) serverAlive(b *backoff.Backoff) error {
	defer b.Reset()

	for {
		d := b.Duration()
		// bogus call to get the serverversion
		resp, err := m.Client.Logout()
		if err != nil {
			return err
		}

		if resp.ServerVersion == "" {
			m.logger.Debugf("Server not up yet, reconnecting in %s", d)
			time.Sleep(d)
		} else {
			m.logger.Infof("Found version %s", resp.ServerVersion)

			return nil
		}
	}
}

// initialize user and teams
// nolint:funlen
func (m *Client) initUser() error {
	m.Lock()
	defer m.Unlock()
	// we only load all team data on initial login.
	// all other updates are for channels from our (primary) team only.
	teams, _, err := m.Client.GetTeamsForUser(m.User.Id, "")
	if err != nil {
		return err
	}

	for _, team := range teams {
		idx := 0
		max := 200
		usermap := make(map[string]*model.User)

		mmusers, _, err := m.Client.GetUsersInTeam(team.Id, idx, max, "")
		if err != nil {
			return err
		}

		for len(mmusers) > 0 {
			for _, user := range mmusers {
				usermap[user.Id] = user
			}

			mmusers, _, err = m.Client.GetUsersInTeam(team.Id, idx, max, "")
			if err != nil {
				return err
			}

			idx++

			time.Sleep(time.Millisecond * 200)
		}
		m.logger.Debugf("found %d users in team %s", len(usermap), team.Name)
		// add all users
		for k, v := range usermap {
			m.Users[k] = v
		}

		t := &Team{
			Team:  team,
			Users: usermap,
			ID:    team.Id,
		}

		m.OtherTeams = append(m.OtherTeams, t)

		if team.Name == m.Credentials.Team {
			m.Team = t
			m.logger.Debugf("initUser(): found our team %s (id: %s)", team.Name, team.Id)
		}
	}

	return nil
}

func (m *Client) initUserChannels() error {
	if err := m.UpdateChannels(); err != nil {
		return err
	}

	for _, t := range m.OtherTeams {
		m.logger.Debugf("found %d channels for user in team %s", len(t.Channels), t.Team.Name)
		m.logger.Debugf("found %d public channels in team %s", len(t.MoreChannels), t.Team.Name)
	}

	return nil
}

func (m *Client) doLogin(firstConnection bool, b *backoff.Backoff) error {
	var (
		logmsg = "trying login"
		err    error
		user   *model.User
	)

	for {
		m.logger.Debugf("%s %s %s %s", logmsg, m.Credentials.Team, m.Credentials.Login, m.Credentials.Server)

		switch {
		case m.Credentials.Token != "":
			user, _, err = m.doLoginToken()
			if err != nil {
				return err
			}
		case m.Credentials.MFAToken != "":
			user, _, err = m.Client.LoginWithMFA(m.Credentials.Login, m.Credentials.Pass, m.Credentials.MFAToken)
		default:
			user, _, err = m.Client.Login(m.Credentials.Login, m.Credentials.Pass)
		}

		if err != nil {
			d := b.Duration()

			m.logger.Debug(err)

			if firstConnection {
				return err
			}

			m.logger.Debugf("LOGIN: %s, reconnecting in %s", err, d)

			time.Sleep(d)

			logmsg = "retrying login"

			continue
		}

		m.User = user

		break
	}
	// reset timer
	b.Reset()

	return nil
}

func (m *Client) doLoginToken() (*model.User, *model.Response, error) {
	var (
		resp   *model.Response
		logmsg = "trying login"
		user   *model.User
		err    error
	)

	m.Client.AuthType = model.HeaderBearer
	m.Client.AuthToken = m.Credentials.Token

	if m.Credentials.CookieToken {
		m.logger.Debugf(logmsg + " with cookie (MMAUTH) token")
		m.Client.HTTPClient.Jar = m.createCookieJar(m.Credentials.Token)
	} else {
		m.logger.Debugf(logmsg + " with personal token")
	}

	user, resp, err = m.Client.GetMe("")
	if err != nil {
		return user, resp, err
	}

	if user == nil {
		m.logger.Errorf("LOGIN TOKEN: %s is invalid", m.Credentials.Pass)

		return user, resp, errors.New("invalid token")
	}

	return user, resp, nil
}

func (m *Client) createCookieJar(token string) *cookiejar.Jar {
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

func (m *Client) wsConnect() {
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
	wsurl := wsScheme + m.Credentials.Server
	// + model.API_URL_SUFFIX_V4
	// + "/websocket"
	header := http.Header{}
	header.Set(model.HeaderAuth, "BEARER "+m.Client.AuthToken)

	m.logger.Debugf("WsClient: making connection: %s", wsurl)

	for {
		wsDialer := &websocket.Dialer{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: m.SkipTLSVerify, //nolint:gosec
			},
			Proxy: http.ProxyFromEnvironment,
		}

		var err error

		m.WsClient, err = model.NewWebSocketClientWithDialer(wsDialer, wsurl, m.Client.AuthToken)
		if err != nil {
			d := b.Duration()

			m.logger.Debugf("WSS: %s, reconnecting in %s", err, d)

			time.Sleep(d)

			continue
		}

		break
	}

	m.WsClient.Listen()

	m.lastPong = time.Now()

	m.logger.Debug("WsClient: connected")

	// only start to parse WS messages when login is completely done
	m.WsConnected = true
}

func (m *Client) doCheckAlive() error {
	if _, _, err := m.Client.GetPing(); err != nil {
		return err
	}

	if m.reconnectBusy {
		return nil
	}

	if m.WsClient.ListenError == nil {
		m.WsClient.SendMessage("ping", nil)
	} else {
		m.logger.Errorf("got a listen error: %#v", m.WsClient.ListenError)

		return m.WsClient.ListenError
	}

	if time.Since(m.lastPong) > 90*time.Second {
		return errors.New("no pong received in 90 seconds")
	}

	return nil
}

func (m *Client) checkAlive(ctx context.Context) {
	ticker := time.NewTicker(time.Second * 45)

	for {
		select {
		case <-ctx.Done():
			m.logger.Debugf("checkAlive: ctx.Done() triggered")

			return
		case <-ticker.C:
			// check if session still is valid
			err := m.doCheckAlive()
			if err != nil {
				m.logger.Errorf("connection not alive: %s", err)
				m.aliveChan <- false
			}

			m.aliveChan <- true
		}
	}
}

func (m *Client) checkConnection(ctx context.Context) {
	go m.checkAlive(ctx)

	for {
		select {
		case alive := <-m.aliveChan:
			if !alive {
				time.Sleep(time.Second * 10)

				if m.doCheckAlive() != nil {
					m.Reconnect()
				}
			}
		case <-ctx.Done():
			m.logger.Debug("checkConnection: ctx.Done() triggered, exiting")

			return
		}
	}
}

// WsReceiver implements the core loop that manages the connection to the chat server. In
// case of a disconnect it will try to reconnect. A call to this method is blocking until
// the 'WsQuite' field of the MMClient object is set to 'true'.
func (m *Client) WsReceiver(ctx context.Context) {
	m.logger.Debug("starting WsReceiver")

	ticker := time.NewTicker(time.Second * 10)

	for {
		select {
		case event := <-m.WsClient.EventChannel:
			if event == nil {
				return
			}

			if !event.IsValid() {
				continue
			}

			m.logger.Debugf("WsReceiver event: %#v", event)

			msg := &Message{
				Raw:  event,
				Team: m.Credentials.Team,
			}

			if !Matterircd {
				m.parseMessage(msg)
			}

			m.MessageChan <- msg
		case response := <-m.WsClient.ResponseChannel:
			if response == nil || !response.IsValid() {
				continue
			}

			m.logger.Debugf("WsReceiver response: %#v", response)

			if text, ok := response.Data["text"].(string); ok {
				if text == "pong" {
					m.lastPong = time.Now()
				}
			}

			m.parseResponse(response)
		case <-m.WsClient.PingTimeoutChannel:
			m.logger.Error("got a ping timeout")
			m.Reconnect()

			return
		case <-ticker.C:
			if m.WsClient.ListenError != nil {
				m.logger.Errorf("%#v", m.WsClient.ListenError)
				m.Reconnect()

				return
			}
		case <-ctx.Done():
			m.logger.Debugf("wsReceiver: ctx.Done() triggered")

			return
		}
	}
}

// Logout disconnects the client from the chat server.
func (m *Client) reconnectLogout() error {
	err := m.Logout()
	m.WsQuit = false

	if err != nil {
		return err
	}

	return nil
}

// Logout disconnects the client from the chat server.
func (m *Client) Logout() error {
	m.logger.Debug("logout running loginCancel to exit goroutines")
	m.loginCancel()

	m.logger.Debugf("logout as %s (team: %s) on %s", m.Credentials.Login, m.Credentials.Team, m.Credentials.Server)
	m.WsQuit = true
	// close the websocket
	m.logger.Debug("closing websocket")
	m.WsClient.Close()

	if strings.Contains(m.Credentials.Pass, model.SessionCookieToken) {
		m.logger.Debug("Not invalidating session in logout, credential is a token")

		return nil
	}

	// actually log out
	m.logger.Debug("running m.Client.Logout")

	if _, err := m.Client.Logout(); err != nil {
		return err
	}

	m.logger.Debug("exiting Logout()")

	return nil
}

// SetLogLevel tries to parse the specified level and if successful sets
// the log level accordingly. Accepted levels are: 'debug', 'info', 'warn',
// 'error', 'fatal' and 'panic'.
func (m *Client) SetLogLevel(level string) {
	l, err := logrus.ParseLevel(level)
	if err != nil {
		m.logger.Warnf("Failed to parse specified log-level '%s': %#v", level, err)
	} else {
		m.rootLogger.SetLevel(l)
	}
}

func (m *Client) HandleRatelimit(name string, resp *model.Response) error {
	if resp == nil {
		return fmt.Errorf("Got a nil model response from %s", name)
	}

	if resp.StatusCode != 429 {
		return fmt.Errorf("StatusCode error: %d", resp.StatusCode)
	}

	waitTime, err := strconv.Atoi(resp.Header.Get("X-RateLimit-Reset"))
	if err != nil {
		return err
	}

	m.logger.Warnf("Ratelimited on %s for %d", name, waitTime)

	time.Sleep(time.Duration(waitTime) * time.Second)

	return nil
}

func (m *Client) antiIdle(ctx context.Context, channelID string, interval int) {
	if interval == 0 {
		interval = 60
	}

	m.logger.Debugf("starting antiIdle for %s every %d secs", channelID, interval)
	ticker := time.NewTicker(time.Second * time.Duration(interval))

	for {
		select {
		case <-ctx.Done():
			m.logger.Debugf("antiIlde: ctx.Done() triggered, exiting for %s", channelID)

			return
		case <-ticker.C:
			m.logger.Tracef("antiIdle %s", channelID)

			m.UpdateLastViewed(channelID)
		}
	}
}
