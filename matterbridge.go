package main

import (
	"crypto/tls"
	"flag"
	"github.com/42wim/matterbridge/matterhook"
	log "github.com/Sirupsen/logrus"
	"github.com/peterhellberg/giphy"
	"github.com/thoj/go-ircevent"
	"strconv"
	"strings"
	"sort"
)

type Bridge struct {
	i           *irc.Connection
	m           *matterhook.Client
	cmap        map[string]string
	ircNick     string
	ircNickPass string
	names       []string
	*Config
}

func NewBridge(name string, config *Config) *Bridge {
	b := &Bridge{}
	b.Config = config
	b.cmap = make(map[string]string)
	b.ircNick = b.Config.IRC.Nick
	b.names = make([]string, 0, 10000)
	if len(b.Config.Token) > 0 {
		for _, val := range b.Config.Token {
			b.cmap[val.IRCChannel] = val.MMChannel
		}
	}
	b.m = matterhook.New(b.Config.Mattermost.URL,
		matterhook.Config{Port: b.Config.Mattermost.Port, Token: b.Config.Mattermost.Token,
			InsecureSkipVerify: b.Config.Mattermost.SkipTLSVerify,
			BindAddress:        b.Config.Mattermost.BindAddress})
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
	log.Info("Joining ", b.Config.IRC.Channel, " as ", b.ircNick)
	i.Join(b.Config.IRC.Channel)
	for _, val := range b.Config.Token {
		log.Info("Joining ", val.IRCChannel, " as ", b.ircNick)
		i.Join(val.IRCChannel)
	}
	i.AddCallback("PRIVMSG", b.handlePrivMsg)
	i.AddCallback("CTCP_ACTION", b.handlePrivMsg)
	if b.Config.Mattermost.ShowJoinPart {
		i.AddCallback("JOIN", b.handleJoinPart)
		i.AddCallback("PART", b.handleJoinPart)
	}
}

func (b *Bridge) handlePrivMsg(event *irc.Event) {
	msg := ""
	if event.Code == "CTCP_ACTION" {
		msg = event.Nick + " "
	}
	msg += event.Message()
	b.Send("irc-"+event.Nick, msg, b.getMMChannel(event.Arguments[0]))
}

func (b *Bridge) handleJoinPart(event *irc.Event) {
	b.Send(b.ircNick, "irc-"+event.Nick+" "+strings.ToLower(event.Code)+"s "+event.Message(), b.getMMChannel(event.Arguments[0]))
}

func (b *Bridge) handleNotice(event *irc.Event) {
	if strings.Contains(event.Message(), "This nickname is registered") {
		b.i.Privmsg(b.Config.IRC.NickServNick, "IDENTIFY "+b.Config.IRC.NickServPassword)
	}
}

func tableformatter(nicks []string, nicksPerRow int) string {
	result := "|IRC users"
	if nicksPerRow < 1 {
		nicksPerRow = 4
	}
	for i := 0; i < 2; i++ {
		for j := 1; j <= nicksPerRow && j <= len(nicks); j++ {
			if i == 0 {
				result += "|"
			} else {
				result += ":-|"
			}
		}
		result += "\r\n|"
	}
	result += nicks[0] + "|"
	for i := 1; i < len(nicks); i++ {
		if i%nicksPerRow == 0 {
			result += "\r\n|" + nicks[i] + "|"
		} else {
			result += nicks[i] + "|"
		}
	}
	return result
}

func plainformatter(nicks []string, nicksPerRow int) string {
	return strings.Join(nicks, ", ") + " currently on IRC"
}

func (b *Bridge) formatnicks(nicks []string) string {
	switch b.Config.Mattermost.NickFormatter {
	case "table":
		return tableformatter(nicks, b.Config.Mattermost.NicksPerRow)
	default:
		return plainformatter(nicks, b.Config.Mattermost.NicksPerRow)
	}
}

func (b *Bridge) storeNames(event *irc.Event) {
	b.names = append(b.names, strings.Split(event.Message(), " ")...)
}

func (b *Bridge) endNames(event *irc.Event) {
	sort.Strings(b.names)
	b.Send(b.ircNick, b.formatnicks(b.names), b.getMMChannel(event.Arguments[0]))
	b.names = make([]string, 0, 10000)
}

func (b *Bridge) handleOther(event *irc.Event) {
	switch event.Code {
	case "001":
		b.handleNewConnection(event)
	case "366":
		b.endNames(event)
	case "353":
		b.storeNames(event)
	case "NOTICE":
		b.handleNotice(event)
	default:
		log.Debugf("UNKNOWN EVENT: %+v", event)
		return
	}
	log.Debugf("%+v", event)
}

func (b *Bridge) Send(nick string, message string, channel string) error {
	return b.SendType(nick, message, channel, "")
}

func IsMarkup(message string) bool {
	switch message[0] {
	case '|':
		fallthrough
	case '#':
		fallthrough
	case '_':
		fallthrough
	case '*':
		fallthrough
	case '~':
		fallthrough
	case '-':
		fallthrough
	case ':':
		fallthrough
	case '>':
		fallthrough
	case '=':
		return true
	}
	return false
}

func (b *Bridge) SendType(nick string, message string, channel string, mtype string) error {
	matterMessage := matterhook.OMessage{IconURL: b.Config.Mattermost.IconURL}
	matterMessage.Channel = channel
	matterMessage.UserName = nick
	matterMessage.Type = mtype
	if b.Config.Mattermost.PrefixMessagesWithNick {
		if IsMarkup(message) {
			matterMessage.Text = nick + ":\n\n" + message
		} else {
			matterMessage.Text = nick + ": " + message
		}
	} else {
		matterMessage.Text = message
	}
	err := b.m.Send(matterMessage)
	if err != nil {
		log.Info(err)
		return err
	}
	return nil
}

func (b *Bridge) handleMatter() {
	var username string
	for {
		message := b.m.Receive()
		username = message.UserName + ": "
		if b.Config.IRC.UseSlackCircumfix {
			username = "<" + message.UserName + "> "
		}
		cmd := strings.Fields(message.Text)[0]
		switch cmd {
		case "!users":
			log.Info("received !users from ", message.UserName)
			b.i.SendRaw("NAMES " + b.getIRCChannel(message.Token))
			return
		case "!gif":
			message.Text = b.giphyRandom(strings.Fields(strings.Replace(message.Text, "!gif ", "", 1)))
			b.Send(b.ircNick, message.Text, b.getIRCChannel(message.Token))
			return
		}
		texts := strings.Split(message.Text, "\n")
		for _, text := range texts {
			channel := b.getIRCChannel(message.Token)
			log.Debug("Sending message from " + message.UserName + " to " + channel)
			b.i.Privmsg(channel, username+text)
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
	mmchannel, ok := b.cmap[ircChannel]
	if !ok {
		mmchannel = b.Config.Mattermost.Channel
	}
	return mmchannel
}

func (b *Bridge) getIRCChannel(token string) string {
	ircchannel := b.Config.IRC.Channel
	_, ok := b.Config.Token[token]
	if ok {
		ircchannel = b.Config.Token[token].IRCChannel
	}
	return ircchannel
}

func init() {
	log.SetFormatter(&log.TextFormatter{FullTimestamp: true})
}

func main() {
	flagConfig := flag.String("conf", "matterbridge.conf", "config file")
	flagDebug := flag.Bool("debug", false, "enable debug")
	flag.Parse()
	if *flagDebug {
		log.Info("enabling debug")
		log.SetLevel(log.DebugLevel)
	}
	NewBridge("matterbot", NewConfig(*flagConfig))
	select {}
}
