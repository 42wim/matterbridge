package main

import (
	"crypto/tls"
	"flag"
	"github.com/42wim/matterbridge/matterhook"
	"github.com/peterhellberg/giphy"
	"github.com/thoj/go-ircevent"
	"log"
	"strconv"
	"strings"
	"time"
)

type Bridge struct {
	i    *irc.Connection
	m    *matterhook.Client
	cmap map[string]string
	*Config
}

func NewBridge(name string, config *Config) *Bridge {
	b := &Bridge{}
	b.Config = config
	b.cmap = make(map[string]string)
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
	i.Connect(b.Config.IRC.Server + ":" + strconv.Itoa(b.Config.IRC.Port))
	time.Sleep(time.Second)
	log.Println("Joining", b.Config.IRC.Channel, "as", b.Config.IRC.Nick)
	i.Join(b.Config.IRC.Channel)
	for _, val := range b.Config.Token {
		log.Println("Joining", val.IRCChannel, "as", b.Config.IRC.Nick)
		i.Join(val.IRCChannel)
	}
	i.AddCallback("PRIVMSG", b.handlePrivMsg)
	i.AddCallback("CTCP_ACTION", b.handlePrivMsg)
	if b.Config.Mattermost.ShowJoinPart {
		i.AddCallback("JOIN", b.handleJoinPart)
		i.AddCallback("PART", b.handleJoinPart)
	}
	//i.AddCallback("353", b.handleOther)
	return i
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
	b.Send(b.Config.IRC.Nick, "irc-"+event.Nick+" "+strings.ToLower(event.Code)+"s "+event.Message(), b.getMMChannel(event.Arguments[0]))
	//b.SendType(b.Config.IRC.Nick, "irc-"+event.Nick+" "+strings.ToLower(event.Code)+"s "+event.Message(), b.getMMChannel(event.Arguments[0]), "join_leave")
}

func (b *Bridge) handleOther(event *irc.Event) {
	switch event.Code {
	case "353":
		log.Println("handleOther", b.getMMChannel(event.Arguments[0]))
		b.Send(b.Config.IRC.Nick, event.Message()+" currently on IRC", b.getMMChannel(event.Arguments[0]))
	}
}

func (b *Bridge) Send(nick string, message string, channel string) error {
	return b.SendType(nick, message, channel, "")
}

func (b *Bridge) SendType(nick string, message string, channel string, mtype string) error {
	matterMessage := matterhook.OMessage{IconURL: b.Config.Mattermost.IconURL}
	matterMessage.Channel = channel
	matterMessage.UserName = nick
	matterMessage.Text = message
	matterMessage.Type = mtype
	err := b.m.Send(matterMessage)
	if err != nil {
		log.Println(err)
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
			log.Println("received !users from", message.UserName)
			b.i.SendRaw("NAMES " + b.getIRCChannel(message.Token))
		case "!gif":
			message.Text = b.giphyRandom(strings.Fields(strings.Replace(message.Text, "!gif ", "", 1)))
			b.Send(b.Config.IRC.Nick, message.Text, b.getIRCChannel(message.Token))
		}
		texts := strings.Split(message.Text, "\n")
		for _, text := range texts {
			b.i.Privmsg(b.getIRCChannel(message.Token), username+text)
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

func main() {
	flagConfig := flag.String("conf", "matterbridge.conf", "config file")
	flag.Parse()
	NewBridge("matterbot", NewConfig(*flagConfig))
	select {}
}
