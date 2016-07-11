package bridge

import (
	"crypto/tls"
	"github.com/42wim/matterbridge/matterclient"
	"github.com/42wim/matterbridge/matterhook"
	log "github.com/Sirupsen/logrus"
	"github.com/peterhellberg/giphy"
	ircm "github.com/sorcix/irc"
	"github.com/thoj/go-ircevent"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

//type Bridge struct {
type MMhook struct {
	mh *matterhook.Client
}

type MMapi struct {
	mc            *matterclient.MMClient
	mmMap         map[string]string
	mmIgnoreNicks []string
}

type MMirc struct {
	i              *irc.Connection
	ircNick        string
	ircMap         map[string]string
	names          map[string][]string
	ircIgnoreNicks []string
}

type MMMessage struct {
	Text     string
	Channel  string
	Username string
}

type Bridge struct {
	MMhook
	MMapi
	MMirc
	*Config
	kind string
}

type FancyLog struct {
	irc *log.Entry
	mm  *log.Entry
}

var flog FancyLog

const Legacy = "legacy"

func initFLog() {
	flog.irc = log.WithFields(log.Fields{"module": "irc"})
	flog.mm = log.WithFields(log.Fields{"module": "mattermost"})
}

func NewBridge(name string, config *Config, kind string) *Bridge {
	initFLog()
	b := &Bridge{}
	b.Config = config
	b.kind = kind
	b.ircNick = b.Config.IRC.Nick
	b.ircMap = make(map[string]string)
	b.mmMap = make(map[string]string)
	b.MMirc.names = make(map[string][]string)
	b.ircIgnoreNicks = strings.Fields(b.Config.IRC.IgnoreNicks)
	b.mmIgnoreNicks = strings.Fields(b.Config.Mattermost.IgnoreNicks)
	for _, val := range b.Config.Channel {
		b.ircMap[val.IRC] = val.Mattermost
		b.mmMap[val.Mattermost] = val.IRC
	}
	if kind == Legacy {
		b.mh = matterhook.New(b.Config.Mattermost.URL,
			matterhook.Config{Port: b.Config.Mattermost.Port, Token: b.Config.Mattermost.Token,
				InsecureSkipVerify: b.Config.Mattermost.SkipTLSVerify,
				BindAddress:        b.Config.Mattermost.BindAddress})
	} else {
		b.mc = matterclient.New(b.Config.Mattermost.Login, b.Config.Mattermost.Password,
			b.Config.Mattermost.Team, b.Config.Mattermost.Server)
		b.mc.SkipTLSVerify = b.Config.Mattermost.SkipTLSVerify
		b.mc.NoTLS = b.Config.Mattermost.NoTLS
		flog.mm.Infof("Trying login %s (team: %s) on %s", b.Config.Mattermost.Login, b.Config.Mattermost.Team, b.Config.Mattermost.Server)
		err := b.mc.Login()
		if err != nil {
			flog.mm.Fatal("Can not connect", err)
		}
		flog.mm.Info("Login ok")
		b.mc.JoinChannel(b.Config.Mattermost.Channel)
		for _, val := range b.Config.Channel {
			b.mc.JoinChannel(val.Mattermost)
		}
		go b.mc.WsReceiver()
	}
	flog.irc.Info("Trying IRC connection")
	b.i = b.createIRC(name)
	flog.irc.Info("Connection succeeded")
	go b.handleMatter()
	return b
}

func (b *Bridge) createIRC(name string) *irc.Connection {
	i := irc.IRC(b.Config.IRC.Nick, b.Config.IRC.Nick)
	i.UseTLS = b.Config.IRC.UseTLS
	i.TLSConfig = &tls.Config{InsecureSkipVerify: b.Config.IRC.SkipTLSVerify}
	if b.Config.IRC.Password != "" {
		i.Password = b.Config.IRC.Password
	}
	i.AddCallback(ircm.RPL_WELCOME, b.handleNewConnection)
	i.Connect(b.Config.IRC.Server + ":" + strconv.Itoa(b.Config.IRC.Port))
	return i
}

func (b *Bridge) handleNewConnection(event *irc.Event) {
	flog.irc.Info("Registering callbacks")
	i := b.i
	b.ircNick = event.Arguments[0]
	i.AddCallback("PRIVMSG", b.handlePrivMsg)
	i.AddCallback("CTCP_ACTION", b.handlePrivMsg)
	i.AddCallback(ircm.RPL_ENDOFNAMES, b.endNames)
	i.AddCallback(ircm.RPL_NAMREPLY, b.storeNames)
	i.AddCallback(ircm.RPL_TOPICWHOTIME, b.handleTopicWhoTime)
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

func (b *Bridge) setupChannels() {
	i := b.i
	for _, val := range b.Config.Channel {
		flog.irc.Infof("Joining %s as %s", val.IRC, b.ircNick)
		i.Join(val.IRC)
	}
}

func (b *Bridge) handleIrcBotCommand(event *irc.Event) bool {
	parts := strings.Fields(event.Message())
	exp, _ := regexp.Compile("[:,]+$")
	channel := event.Arguments[0]
	command := ""
	if len(parts) == 2 {
		command = parts[1]
	}
	if exp.ReplaceAllString(parts[0], "") == b.ircNick {
		switch command {
		case "users":
			usernames := b.mc.UsernamesInChannel(b.getMMChannel(channel))
			sort.Strings(usernames)
			b.i.Privmsg(channel, "Users on Mattermost: "+strings.Join(usernames, ", "))
		default:
			b.i.Privmsg(channel, "Valid commands are: [users, help]")
		}
		return true
	}
	return false
}

func (b *Bridge) ircNickFormat(nick string) string {
	if nick == b.ircNick {
		return nick
	}
	if b.Config.Mattermost.RemoteNickFormat == nil {
		return "irc-" + nick
	}
	return strings.Replace(*b.Config.Mattermost.RemoteNickFormat, "{NICK}", nick, -1)
}

func (b *Bridge) handlePrivMsg(event *irc.Event) {
	flog.irc.Debugf("handlePrivMsg() %s %s", event.Nick, event.Message)
	if b.ignoreMessage(event.Nick, event.Message(), "irc") {
		return
	}
	if b.handleIrcBotCommand(event) {
		return
	}
	msg := ""
	if event.Code == "CTCP_ACTION" {
		msg = event.Nick + " "
	}
	msg += event.Message()
	b.Send(b.ircNickFormat(event.Nick), msg, b.getMMChannel(event.Arguments[0]))
}

func (b *Bridge) handleJoinPart(event *irc.Event) {
	b.Send(b.ircNick, b.ircNickFormat(event.Nick)+" "+strings.ToLower(event.Code)+"s "+event.Message(), b.getMMChannel(event.Arguments[0]))
}

func (b *Bridge) handleNotice(event *irc.Event) {
	if strings.Contains(event.Message(), "This nickname is registered") {
		b.i.Privmsg(b.Config.IRC.NickServNick, "IDENTIFY "+b.Config.IRC.NickServPassword)
	}
}

func (b *Bridge) nicksPerRow() int {
	if b.Config.Mattermost.NicksPerRow < 1 {
		return 4
	}
	return b.Config.Mattermost.NicksPerRow
}

func (b *Bridge) formatnicks(nicks []string, continued bool) string {
	switch b.Config.Mattermost.NickFormatter {
	case "table":
		return tableformatter(nicks, b.nicksPerRow(), continued)
	default:
		return plainformatter(nicks, b.nicksPerRow())
	}
}

func (b *Bridge) storeNames(event *irc.Event) {
	channel := event.Arguments[2]
	b.MMirc.names[channel] = append(
		b.MMirc.names[channel],
		strings.Split(strings.TrimSpace(event.Message()), " ")...)
}

func (b *Bridge) endNames(event *irc.Event) {
	channel := event.Arguments[1]
	sort.Strings(b.MMirc.names[channel])
	maxNamesPerPost := (300 / b.nicksPerRow()) * b.nicksPerRow()
	continued := false
	for len(b.MMirc.names[channel]) > maxNamesPerPost {
		b.Send(
			b.ircNick,
			b.formatnicks(b.MMirc.names[channel][0:maxNamesPerPost], continued),
			b.getMMChannel(channel))
		b.MMirc.names[channel] = b.MMirc.names[channel][maxNamesPerPost:]
		continued = true
	}
	b.Send(b.ircNick, b.formatnicks(b.MMirc.names[channel], continued), b.getMMChannel(channel))
	b.MMirc.names[channel] = nil
}

func (b *Bridge) handleTopicWhoTime(event *irc.Event) {
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

func (b *Bridge) handleOther(event *irc.Event) {
	flog.irc.Debugf("%#v", event)
}

func (b *Bridge) Send(nick string, message string, channel string) error {
	return b.SendType(nick, message, channel, "")
}

func (b *Bridge) SendType(nick string, message string, channel string, mtype string) error {
	if b.Config.Mattermost.PrefixMessagesWithNick {
		if IsMarkup(message) {
			message = nick + "\n\n" + message
		} else {
			message = nick + " " + message
		}
	}
	if b.kind == Legacy {
		matterMessage := matterhook.OMessage{IconURL: b.Config.Mattermost.IconURL}
		matterMessage.Channel = channel
		matterMessage.UserName = nick
		matterMessage.Type = mtype
		matterMessage.Text = message
		err := b.mh.Send(matterMessage)
		if err != nil {
			flog.mm.Info(err)
			return err
		}
		flog.mm.Debug("->mattermost channel: ", channel, " ", message)
		return nil
	}
	flog.mm.Debug("->mattermost channel: ", channel, " ", message)
	b.mc.PostMessage(channel, message)
	return nil
}

func (b *Bridge) handleMatterHook(mchan chan *MMMessage) {
	for {
		message := b.mh.Receive()
		flog.mm.Debugf("receiving from matterhook %#v", message)
		m := &MMMessage{}
		m.Username = message.UserName
		m.Text = message.Text
		m.Channel = message.ChannelName
		mchan <- m
	}
}

func (b *Bridge) handleMatterClient(mchan chan *MMMessage) {
	for message := range b.mc.MessageChan {
		// do not post our own messages back to irc
		if message.Raw.Action == "posted" && b.mc.User.Username != message.Username {
			flog.mm.Debugf("receiving from matterclient %#v", message)
			m := &MMMessage{}
			m.Username = message.Username
			m.Channel = message.Channel
			m.Text = message.Text
			mchan <- m
		}
	}
}

func (b *Bridge) handleMatter() {
	flog.mm.Infof("Choosing Mattermost connection type %s", b.kind)
	mchan := make(chan *MMMessage)
	if b.kind == Legacy {
		go b.handleMatterHook(mchan)
	} else {
		go b.handleMatterClient(mchan)
	}
	flog.mm.Info("Start listening for Mattermost messages")
	for message := range mchan {
		var username string
		if b.ignoreMessage(message.Username, message.Text, "mattermost") {
			continue
		}
		username = message.Username + ": "
		if b.Config.IRC.RemoteNickFormat != "" {
			username = strings.Replace(b.Config.IRC.RemoteNickFormat, "{NICK}", message.Username, -1)
		}
		cmds := strings.Fields(message.Text)
		// empty message
		if len(cmds) == 0 {
			continue
		}
		cmd := cmds[0]
		switch cmd {
		case "!users":
			flog.mm.Info("Received !users from ", message.Username)
			b.i.SendRaw("NAMES " + b.getIRCChannel(message.Channel))
			continue
		case "!gif":
			message.Text = b.giphyRandom(strings.Fields(strings.Replace(message.Text, "!gif ", "", 1)))
			b.Send(b.ircNick, message.Text, b.getIRCChannel(message.Channel))
			continue
		}
		texts := strings.Split(message.Text, "\n")
		for _, text := range texts {
			flog.mm.Debug("Sending message from " + message.Username + " to " + message.Channel)
			b.i.Privmsg(b.getIRCChannel(message.Channel), username+text)
		}
	}
}

func (b *Bridge) giphyRandom(query []string) string {
	g := giphy.DefaultClient
	if b.Config.General.GiphyAPIKey != "" {
		g.APIKey = b.Config.General.GiphyAPIKey
	}
	res, err := g.Random(query)
	if err != nil {
		return "error"
	}
	return res.Data.FixedHeightDownsampledURL
}

func (b *Bridge) getMMChannel(ircChannel string) string {
	mmChannel := b.ircMap[ircChannel]
	if b.kind == Legacy {
		return mmChannel
	}
	return b.mc.GetChannelId(mmChannel, "")
}

func (b *Bridge) getIRCChannel(mmChannel string) string {
	return b.mmMap[mmChannel]
}

func (b *Bridge) ignoreMessage(nick string, message string, protocol string) bool {
	var ignoreNicks = b.mmIgnoreNicks
	if protocol == "irc" {
		ignoreNicks = b.ircIgnoreNicks
	}
	// should we discard messages ?
	for _, entry := range ignoreNicks {
		if nick == entry {
			return true
		}
	}
	return false
}
