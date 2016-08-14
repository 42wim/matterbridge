package birc

import (
	"crypto/tls"
	"github.com/42wim/matterbridge/bridge/config"
	log "github.com/Sirupsen/logrus"
	ircm "github.com/sorcix/irc"
	"github.com/thoj/go-ircevent"
	"sort"
	"strconv"
	"strings"
	"time"
)

//type Bridge struct {
type Birc struct {
	i              *irc.Connection
	ircNick        string
	ircMap         map[string]string
	names          map[string][]string
	ircIgnoreNicks []string
	*config.Config
	kind   string
	Remote chan config.Message
}

type FancyLog struct {
	irc  *log.Entry
	mm   *log.Entry
	xmpp *log.Entry
}

var flog FancyLog

const Legacy = "legacy"

func init() {
	flog.irc = log.WithFields(log.Fields{"module": "irc"})
	flog.mm = log.WithFields(log.Fields{"module": "mattermost"})
	flog.xmpp = log.WithFields(log.Fields{"module": "xmpp"})
}

func New(config *config.Config, c chan config.Message) *Birc {
	b := &Birc{}
	b.Config = config
	b.kind = "legacy"
	b.Remote = c
	b.ircNick = b.Config.IRC.Nick
	b.ircMap = make(map[string]string)
	b.names = make(map[string][]string)
	b.ircIgnoreNicks = strings.Fields(b.Config.IRC.IgnoreNicks)
	flog.irc.Info("Trying IRC connection")
	b.i = b.connect()
	flog.irc.Info("Connection succeeded")
	return b
}

func (b *Birc) Command(msg *config.Message) string {
	switch msg.Text {
	case "!users":
		b.i.SendRaw("NAMES " + msg.Channel)
	}
	return ""
}

func (b *Birc) Name() string {
	return "irc"
}

func (b *Birc) Send(msg config.Message) error {
	if msg.Origin == "irc" {
		return nil
	}
	if strings.HasPrefix(msg.Text, "!") {
		b.Command(&msg)
		return nil
	}
	username := b.ircNickFormat(msg.Username)
	b.i.Privmsg(msg.Channel, username+msg.Text)
	return nil
}

func (b *Birc) connect() *irc.Connection {
	i := irc.IRC(b.Config.IRC.Nick, b.Config.IRC.Nick)
	i.UseTLS = b.Config.IRC.UseTLS
	i.UseSASL = b.Config.IRC.UseSASL
	i.SASLLogin = b.Config.IRC.NickServNick
	i.SASLPassword = b.Config.IRC.NickServPassword
	i.TLSConfig = &tls.Config{InsecureSkipVerify: b.Config.IRC.SkipTLSVerify}
	if b.Config.IRC.Password != "" {
		i.Password = b.Config.IRC.Password
	}
	i.AddCallback(ircm.RPL_WELCOME, b.handleNewConnection)
	err := i.Connect(b.Config.IRC.Server)
	if err != nil {
		flog.irc.Fatal(err)
	}
	return i
}

func (b *Birc) endNames(event *irc.Event) {
	channel := event.Arguments[1]
	sort.Strings(b.names[channel])
	maxNamesPerPost := (300 / b.nicksPerRow()) * b.nicksPerRow()
	continued := false
	for len(b.names[channel]) > maxNamesPerPost {
		b.Remote <- config.Message{Username: b.ircNick, Text: b.formatnicks(b.names[channel][0:maxNamesPerPost], continued), Channel: channel, Origin: "irc"}
		b.names[channel] = b.names[channel][maxNamesPerPost:]
		continued = true
	}
	b.Remote <- config.Message{Username: b.ircNick, Text: b.formatnicks(b.names[channel], continued), Channel: channel, Origin: "irc"}
	b.names[channel] = nil
}

func (b *Birc) handleNewConnection(event *irc.Event) {
	flog.irc.Info("Registering callbacks")
	i := b.i
	b.ircNick = event.Arguments[0]
	i.AddCallback("PRIVMSG", b.handlePrivMsg)
	i.AddCallback("CTCP_ACTION", b.handlePrivMsg)
	i.AddCallback(ircm.RPL_TOPICWHOTIME, b.handleTopicWhoTime)
	i.AddCallback(ircm.RPL_ENDOFNAMES, b.endNames)
	i.AddCallback(ircm.RPL_NAMREPLY, b.storeNames)
	i.AddCallback(ircm.NOTICE, b.handleNotice)
	i.AddCallback(ircm.RPL_MYINFO, func(e *irc.Event) { flog.irc.Infof("%s: %s", e.Code, strings.Join(e.Arguments[1:], " ")) })
	i.AddCallback("PING", func(e *irc.Event) {
		i.SendRaw("PONG :" + e.Message())
		flog.irc.Debugf("PING/PONG")
	})
	if b.Config.Mattermost.ShowJoinPart {
		i.AddCallback("JOIN", b.handleJoinPart)
		i.AddCallback("PART", b.handleJoinPart)
	}
	i.AddCallback("*", b.handleOther)
	b.setupChannels()
}

func (b *Birc) handleJoinPart(event *irc.Event) {
	//b.Send(b.ircNick, b.ircNickFormat(event.Nick)+" "+strings.ToLower(event.Code)+"s "+event.Message(), b.getMMChannel(event.Arguments[0]))
}

func (b *Birc) handleNotice(event *irc.Event) {
	if strings.Contains(event.Message(), "This nickname is registered") {
		b.i.Privmsg(b.Config.IRC.NickServNick, "IDENTIFY "+b.Config.IRC.NickServPassword)
	}
}

func (b *Birc) handleOther(event *irc.Event) {
	flog.irc.Debugf("%#v", event)
}

func (b *Birc) handlePrivMsg(event *irc.Event) {
	flog.irc.Debugf("handlePrivMsg() %s %s", event.Nick, event.Message())
	msg := ""
	if event.Code == "CTCP_ACTION" {
		msg = event.Nick + " "
	}
	msg += event.Message()
	b.Remote <- config.Message{Username: event.Nick, Text: msg, Channel: event.Arguments[0], Origin: "irc"}
}

func (b *Birc) handleTopicWhoTime(event *irc.Event) {
	parts := strings.Split(event.Arguments[2], "!")
	t, err := strconv.ParseInt(event.Arguments[3], 10, 64)
	if err != nil {
		flog.irc.Errorf("Invalid time stamp: %s", event.Arguments[3])
	}
	user := parts[0]
	if len(parts) > 1 {
		user += " [" + parts[1] + "]"
	}
	flog.irc.Infof("%s: Topic set by %s [%s]", event.Code, user, time.Unix(t, 0))
}

func (b *Birc) ircNickFormat(nick string) string {
	flog.irc.Debug("ircnick", nick)
	if nick == b.ircNick {
		return nick
	}
	if b.Config.IRC.RemoteNickFormat == "" {
		return "irc-" + nick
	}
	return strings.Replace(b.Config.IRC.RemoteNickFormat, "{NICK}", nick, -1)
}

func (b *Birc) nicksPerRow() int {
	if b.Config.Mattermost.NicksPerRow < 1 {
		return 4
	}
	return b.Config.Mattermost.NicksPerRow
}

func (b *Birc) setupChannels() {
	for _, val := range b.Config.Channel {
		flog.irc.Infof("Joining %s as %s", val.IRC, b.ircNick)
		b.i.Join(val.IRC)
	}
}

func (b *Birc) storeNames(event *irc.Event) {
	channel := event.Arguments[2]
	b.names[channel] = append(
		b.names[channel],
		strings.Split(strings.TrimSpace(event.Message()), " ")...)
}

func (b *Birc) formatnicks(nicks []string, continued bool) string {
	switch b.Config.Mattermost.NickFormatter {
	case "table":
		return tableformatter(nicks, b.nicksPerRow(), continued)
	default:
		return plainformatter(nicks, b.nicksPerRow())
	}
}
