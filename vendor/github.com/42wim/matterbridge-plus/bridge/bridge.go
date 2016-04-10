package bridge

import (
	"crypto/tls"
	"github.com/42wim/matterbridge-plus/matterclient"
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
	mc    *matterclient.MMClient
	mmMap map[string]string
}

type MMirc struct {
	i       *irc.Connection
	ircNick string
	ircMap  map[string]string
	names   map[string][]string
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
	b.MMirc.names = make(map[string][]string)
	if kind == "legacy" {
		if len(b.Config.Token) > 0 {
			for _, val := range b.Config.Token {
				b.ircMap[val.IRCChannel] = val.MMChannel
			}
		}

		b.mh = matterhook.New(b.Config.Mattermost.URL,
			matterhook.Config{Port: b.Config.Mattermost.Port, Token: b.Config.Mattermost.Token,
				InsecureSkipVerify: b.Config.Mattermost.SkipTLSVerify,
				BindAddress:        b.Config.Mattermost.BindAddress})
	} else {
		b.mmMap = make(map[string]string)
		if len(b.Config.Channel) > 0 {
			for _, val := range b.Config.Channel {
				b.ircMap[val.IRC] = val.Mattermost
				b.mmMap[val.Mattermost] = val.IRC
			}
		}
		b.mc = matterclient.New(b.Config.Mattermost.Login, b.Config.Mattermost.Password,
			b.Config.Mattermost.Team, b.Config.Mattermost.Server)
		err := b.mc.Login()
		if err != nil {
			flog.mm.Fatal("can not connect", err)
		}
		b.mc.JoinChannel(b.Config.Mattermost.Channel)
		if len(b.Config.Channel) > 0 {
			for _, val := range b.Config.Channel {
				b.mc.JoinChannel(val.Mattermost)
			}
		}
		go b.mc.WsReceiver()
	}
	b.i = b.createIRC(name)
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
	i.AddCallback("*", b.handleOther)
	i.Connect(b.Config.IRC.Server + ":" + strconv.Itoa(b.Config.IRC.Port))
	return i
}

func (b *Bridge) handleNewConnection(event *irc.Event) {
	b.ircNick = event.Arguments[0]
	b.setupChannels()
}

func (b *Bridge) setupChannels() {
	i := b.i
	flog.irc.Info("Joining ", b.Config.IRC.Channel, " as ", b.ircNick)
	i.Join(b.Config.IRC.Channel)
	if b.kind == "legacy" {
		for _, val := range b.Config.Token {
			flog.irc.Info("Joining ", val.IRCChannel, " as ", b.ircNick)
			i.Join(val.IRCChannel)
		}
	} else {
		for _, val := range b.Config.Channel {
			flog.irc.Info("Joining ", val.IRC, " as ", b.ircNick)
			i.Join(val.IRC)
		}
	}
	i.AddCallback("PRIVMSG", b.handlePrivMsg)
	i.AddCallback("CTCP_ACTION", b.handlePrivMsg)
	if b.Config.Mattermost.ShowJoinPart {
		i.AddCallback("JOIN", b.handleJoinPart)
		i.AddCallback("PART", b.handleJoinPart)
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

func (b *Bridge) handleTopicWhoTime(event *irc.Event) bool {
	parts := strings.Split(event.Arguments[2], "!")
	t_i, err := strconv.ParseInt(event.Arguments[3], 10, 64)
	if err != nil {
		flog.irc.Errorf("Invalid time stamp: %s", event.Arguments[3])
		return false
	}
	user := parts[0]
	if len(parts) > 1 {
		user += " [" + parts[1] + "]"
	}
	flog.irc.Infof("%s: Topic set by %s [%s]", event.Code, user, time.Unix(t_i, 0))
	return true
}

func (b *Bridge) handleOther(event *irc.Event) {
	flog.irc.Debugf("%#v", event)
	switch event.Code {
	case ircm.RPL_WELCOME:
		b.handleNewConnection(event)
	case ircm.RPL_ENDOFNAMES:
		b.endNames(event)
	case ircm.RPL_NAMREPLY:
		b.storeNames(event)
	case ircm.RPL_ISUPPORT:
		fallthrough
	case ircm.RPL_LUSEROP:
		fallthrough
	case ircm.RPL_LUSERUNKNOWN:
		fallthrough
	case ircm.RPL_LUSERCHANNELS:
		fallthrough
	case ircm.RPL_MYINFO:
		flog.irc.Infof("%s: %s", event.Code, strings.Join(event.Arguments[1:], " "))
	case ircm.RPL_YOURHOST:
		fallthrough
	case ircm.RPL_CREATED:
		fallthrough
	case ircm.RPL_STATSDLINE:
		fallthrough
	case ircm.RPL_LUSERCLIENT:
		fallthrough
	case ircm.RPL_LUSERME:
		fallthrough
	case ircm.RPL_LOCALUSERS:
		fallthrough
	case ircm.RPL_GLOBALUSERS:
		fallthrough
	case ircm.RPL_MOTD:
		flog.irc.Infof("%s: %s", event.Code, event.Message())
		// flog.irc.Info(event.Message())
	case ircm.RPL_TOPIC:
		flog.irc.Infof("%s: Topic for %s: %s", event.Code, event.Arguments[1], event.Message())
	case ircm.RPL_TOPICWHOTIME:
		if !b.handleTopicWhoTime(event) {
			break
		}
	case ircm.MODE:
		flog.irc.Infof("%s: %s %s", event.Code, event.Arguments[1], event.Arguments[0])
	case ircm.JOIN:
		fallthrough
	case ircm.PING:
		fallthrough
	case ircm.PONG:
		flog.irc.Infof("%s: %s", event.Code, event.Message())
	case ircm.RPL_ENDOFMOTD:
	case ircm.RPL_MOTDSTART:
	case ircm.ERR_NICKNAMEINUSE:
		flog.irc.Warn(event.Message())
	case ircm.NOTICE:
		b.handleNotice(event)
	default:
		flog.irc.Infof("UNKNOWN EVENT: %#v", event)
		return
	}
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
	if b.kind == "legacy" {
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
		return nil
	}
	flog.mm.Debug("->mattermost channel: ", channel, " ", message)
	b.mc.PostMessage(channel, message)
	return nil
}

func (b *Bridge) handleMatterHook(mchan chan *MMMessage) {
	for {
		message := b.mh.Receive()
		m := &MMMessage{}
		m.Username = message.UserName
		m.Text = message.Text
		m.Channel = message.Token
		mchan <- m
	}
}

func (b *Bridge) handleMatterClient(mchan chan *MMMessage) {
	for message := range b.mc.MessageChan {
		// do not post our own messages back to irc
		if message.Raw.Action == "posted" && b.mc.User.Username != message.Username {
			m := &MMMessage{}
			m.Username = message.Username
			m.Channel = message.Channel
			m.Text = message.Text
			flog.mm.Debugf("<-mattermost channel: %s %#v %#v", message.Channel, message.Post, message.Raw)
			mchan <- m
		}
	}
}

func (b *Bridge) handleMatter() {
	mchan := make(chan *MMMessage)
	if b.kind == "legacy" {
		go b.handleMatterHook(mchan)
	} else {
		go b.handleMatterClient(mchan)
	}
	for message := range mchan {
		var username string
		username = message.Username + ": "
		if b.Config.IRC.RemoteNickFormat != "" {
			username = strings.Replace(b.Config.IRC.RemoteNickFormat, "{NICK}", message.Username, -1)
		} else if b.Config.IRC.UseSlackCircumfix {
			username = "<" + message.Username + "> "
		}
		cmd := strings.Fields(message.Text)[0]
		switch cmd {
		case "!users":
			flog.mm.Info("received !users from ", message.Username)
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
	mmchannel, ok := b.ircMap[ircChannel]
	if !ok {
		mmchannel = b.Config.Mattermost.Channel
	}
	return mmchannel
}

func (b *Bridge) getIRCChannel(channel string) string {
	if b.kind == "legacy" {
		ircchannel := b.Config.IRC.Channel
		_, ok := b.Config.Token[channel]
		if ok {
			ircchannel = b.Config.Token[channel].IRCChannel
		}
		return ircchannel
	}
	ircchannel, ok := b.mmMap[channel]
	if !ok {
		ircchannel = b.Config.IRC.Channel
	}
	return ircchannel
}
