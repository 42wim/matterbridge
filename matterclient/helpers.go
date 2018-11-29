package matterclient

import (
	"crypto/md5"
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/jpillora/backoff"
	"github.com/mattermost/mattermost-server/model"
)

func (m *MMClient) doLogin(firstConnection bool, b *backoff.Backoff) error {
	var resp *model.Response
	var appErr *model.AppError
	var logmsg = "trying login"
	var err error
	for {
		m.log.Debugf("%s %s %s %s", logmsg, m.Credentials.Team, m.Credentials.Login, m.Credentials.Server)
		if m.Credentials.Token != "" {
			resp, err = m.doLoginToken()
			if err != nil {
				return err
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
	return nil
}

func (m *MMClient) doLoginToken() (*model.Response, error) {
	var resp *model.Response
	var logmsg = "trying login"
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
		return resp, resp.Error
	}
	if m.User == nil {
		m.log.Errorf("LOGIN TOKEN: %s is invalid", m.Credentials.Pass)
		return resp, errors.New("invalid token")
	}
	return resp, nil
}

func (m *MMClient) handleLoginToken() error {
	switch {
	case strings.Contains(m.Credentials.Pass, model.SESSION_COOKIE_TOKEN):
		token := strings.Split(m.Credentials.Pass, model.SESSION_COOKIE_TOKEN+"=")
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

func (m *MMClient) initClient(firstConnection bool, b *backoff.Backoff) error {
	uriScheme := "https://"
	if m.NoTLS {
		uriScheme = "http://"
	}
	// login to mattermost
	m.Client = model.NewAPIv4Client(uriScheme + m.Credentials.Server)
	m.Client.HttpClient.Transport = &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: m.SkipTLSVerify}, Proxy: http.ProxyFromEnvironment}
	m.Client.HttpClient.Timeout = time.Second * 10

	// handle MMAUTHTOKEN and personal token
	if err := m.handleLoginToken(); err != nil {
		return err
	}

	// check if server alive, retry until
	if err := m.serverAlive(firstConnection, b); err != nil {
		return err
	}

	return nil
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

func (m *MMClient) serverAlive(firstConnection bool, b *backoff.Backoff) error {
	defer b.Reset()
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
			return nil
		}
	}
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
