package birc

import (
	"crypto/tls"
	"github.com/42wim/matterbridge/bridge/config"
	log "github.com/Sirupsen/logrus"
	ircm "github.com/sorcix/irc"
	"github.com/thoj/go-ircevent"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

type Birc struct {
	i        *irc.Connection
	Nick     string
	names    map[string][]string
	Config   *config.Protocol
	origin   string
	protocol string
	Remote   chan config.Message
}

var flog *log.Entry
var protocol = "irc"

func init() {
	flog = log.WithFields(log.Fields{"module": protocol})
}

func New(config config.Protocol, origin string, c chan config.Message) *Birc {
	b := &Birc{}
	b.Config = &config
	b.Nick = b.Config.Nick
	b.Remote = c
	b.names = make(map[string][]string)
	b.origin = origin
	b.protocol = protocol
	return b
}

func (b *Birc) Command(msg *config.Message) string {
	switch msg.Text {
	case "!users":
		b.i.SendRaw("NAMES " + msg.Channel)
	}
	return ""
}

func (b *Birc) Connect() error {
	flog.Infof("Connecting %s", b.Config.Server)
	i := irc.IRC(b.Config.Nick, b.Config.Nick)
	i.UseTLS = b.Config.UseTLS
	i.UseSASL = b.Config.UseSASL
	i.SASLLogin = b.Config.NickServNick
	i.SASLPassword = b.Config.NickServPassword
	i.TLSConfig = &tls.Config{InsecureSkipVerify: b.Config.SkipTLSVerify}
	if b.Config.Password != "" {
		i.Password = b.Config.Password
	}
	i.AddCallback(ircm.RPL_WELCOME, b.handleNewConnection)
	err := i.Connect(b.Config.Server)
	if err != nil {
		return err
	}
	flog.Info("Connection succeeded")
	b.i = i
	return nil
}

func (b *Birc) FullOrigin() string {
	return b.protocol + "." + b.origin
}

func (b *Birc) JoinChannel(channel string) error {
	b.i.Join(channel)
	return nil
}

func (b *Birc) Name() string {
	return b.protocol + "." + b.origin
}

func (b *Birc) Protocol() string {
	return b.protocol
}

func (b *Birc) Origin() string {
	return b.origin
}

func (b *Birc) Send(msg config.Message) error {
	flog.Debugf("Receiving %#v", msg)
	if msg.FullOrigin == b.FullOrigin() {
		return nil
	}
	if strings.HasPrefix(msg.Text, "!") {
		b.Command(&msg)
		return nil
	}
	for _, text := range strings.Split(msg.Text, "\n") {
		b.i.Privmsg(msg.Channel, msg.Username+text)
	}
	return nil
}

func (b *Birc) endNames(event *irc.Event) {
	channel := event.Arguments[1]
	sort.Strings(b.names[channel])
	maxNamesPerPost := (300 / b.nicksPerRow()) * b.nicksPerRow()
	continued := false
	for len(b.names[channel]) > maxNamesPerPost {
		b.Remote <- config.Message{Username: b.Nick, Text: b.formatnicks(b.names[channel][0:maxNamesPerPost], continued),
			Channel: channel, Origin: b.origin, Protocol: b.protocol, FullOrigin: b.FullOrigin()}
		b.names[channel] = b.names[channel][maxNamesPerPost:]
		continued = true
	}
	b.Remote <- config.Message{Username: b.Nick, Text: b.formatnicks(b.names[channel], continued), Channel: channel,
		Origin: b.origin, Protocol: b.protocol, FullOrigin: b.FullOrigin()}
	b.names[channel] = nil
}

func (b *Birc) handleNewConnection(event *irc.Event) {
	flog.Debug("Registering callbacks")
	i := b.i
	b.Nick = event.Arguments[0]
	i.AddCallback("PRIVMSG", b.handlePrivMsg)
	i.AddCallback("CTCP_ACTION", b.handlePrivMsg)
	i.AddCallback(ircm.RPL_TOPICWHOTIME, b.handleTopicWhoTime)
	i.AddCallback(ircm.RPL_ENDOFNAMES, b.endNames)
	i.AddCallback(ircm.RPL_NAMREPLY, b.storeNames)
	i.AddCallback(ircm.NOTICE, b.handleNotice)
	//i.AddCallback(ircm.RPL_MYINFO, func(e *irc.Event) { flog.Infof("%s: %s", e.Code, strings.Join(e.Arguments[1:], " ")) })
	i.AddCallback("PING", func(e *irc.Event) {
		i.SendRaw("PONG :" + e.Message())
		flog.Debugf("PING/PONG")
	})
	i.AddCallback("*", b.handleOther)
}

func (b *Birc) handleNotice(event *irc.Event) {
	if strings.Contains(event.Message(), "This nickname is registered") {
		b.i.Privmsg(b.Config.NickServNick, "IDENTIFY "+b.Config.NickServPassword)
	}
}

func (b *Birc) handleOther(event *irc.Event) {
	switch event.Code {
	case "372", "375", "376", "250", "251", "252", "253", "254", "255", "265", "266", "002", "003", "004", "005":
		return
	}
	flog.Debugf("%#v", event.Raw)
}

func (b *Birc) handlePrivMsg(event *irc.Event) {
	flog.Debugf("handlePrivMsg() %s %s", event.Nick, event.Message())
	msg := ""
	if event.Code == "CTCP_ACTION" {
		msg = event.Nick + " "
	}
	msg += event.Message()
	// strip IRC colors
	re := regexp.MustCompile(`[[:cntrl:]](\d+,|)\d+`)
	msg = re.ReplaceAllString(msg, "")
	flog.Debugf("Sending message from %s on %s to gateway", event.Arguments[0], b.FullOrigin())
	b.Remote <- config.Message{Username: event.Nick, Text: msg, Channel: event.Arguments[0], Origin: b.origin, Protocol: b.protocol, FullOrigin: b.FullOrigin()}
}

func (b *Birc) handleTopicWhoTime(event *irc.Event) {
	parts := strings.Split(event.Arguments[2], "!")
	t, err := strconv.ParseInt(event.Arguments[3], 10, 64)
	if err != nil {
		flog.Errorf("Invalid time stamp: %s", event.Arguments[3])
	}
	user := parts[0]
	if len(parts) > 1 {
		user += " [" + parts[1] + "]"
	}
	flog.Debugf("%s: Topic set by %s [%s]", event.Code, user, time.Unix(t, 0))
}

func (b *Birc) nicksPerRow() int {
	return 4
	/*
		if b.Config.Mattermost.NicksPerRow < 1 {
			return 4
		}
		return b.Config.Mattermost.NicksPerRow
	*/
}

func (b *Birc) storeNames(event *irc.Event) {
	channel := event.Arguments[2]
	b.names[channel] = append(
		b.names[channel],
		strings.Split(strings.TrimSpace(event.Message()), " ")...)
}

func (b *Birc) formatnicks(nicks []string, continued bool) string {
	return plainformatter(nicks, b.nicksPerRow())
	/*
		switch b.Config.Mattermost.NickFormatter {
		case "table":
			return tableformatter(nicks, b.nicksPerRow(), continued)
		default:
			return plainformatter(nicks, b.nicksPerRow())
		}
	*/
}
