package main

import (
	"crypto/tls"
	"github.com/42wim/matterbridge/matterhook"
	"github.com/thoj/go-ircevent"
	"log"
	"strconv"
	"strings"
	"time"
)

type Bridge struct {
	i *irc.Connection
	m *matterhook.Client
	*Config
}

func NewBridge(name string, config *Config) *Bridge {
	b := &Bridge{}
	b.Config = config
	b.m = matterhook.New(b.Config.Mattermost.URL,
		matterhook.Config{Port: b.Config.Mattermost.Port, Token: b.Config.Mattermost.Token})
	b.i = b.createIRC(name)
	go b.handleMatter()
	return b
}

func (b *Bridge) createIRC(name string) *irc.Connection {
	i := irc.IRC(b.Config.IRC.Nick, b.Config.IRC.Nick)
	i.UseTLS = b.Config.IRC.UseTLS
	i.TLSConfig = &tls.Config{InsecureSkipVerify: b.Config.IRC.SkipTLSVerify}
	i.Connect(b.Config.IRC.Server + ":" + strconv.Itoa(b.Config.IRC.Port))
	time.Sleep(time.Second)
	log.Println("Joining", b.Config.IRC.Channel, "as", b.Config.IRC.Nick)
	i.Join(b.Config.IRC.Channel)
	i.AddCallback("PRIVMSG", b.handlePrivMsg)
	i.AddCallback("CTCP_ACTION", b.handlePrivMsg)
	if b.Config.Mattermost.ShowJoinPart {
		i.AddCallback("JOIN", b.handleJoinPart)
		i.AddCallback("PART", b.handleJoinPart)
	}
	i.AddCallback("353", b.handleOther)
	return i
}

func (b *Bridge) handlePrivMsg(event *irc.Event) {
	matterMessage := matterhook.OMessage{}
	if event.Code == "CTCP_ACTION" {
		matterMessage.Text = event.Nick + " "
	}
	matterMessage.Text += event.Message()
	matterMessage.UserName = "irc-" + event.Nick
	b.m.Send(matterMessage)
}

func (b *Bridge) handleJoinPart(event *irc.Event) {
	matterMessage := matterhook.OMessage{}
	matterMessage.Text = "irc-" + event.Nick + " " + strings.ToLower(event.Code) + "s " + event.Message()
	matterMessage.UserName = b.Config.IRC.Nick
	b.m.Send(matterMessage)
}

func (b *Bridge) handleOther(event *irc.Event) {
	matterMessage := matterhook.OMessage{}
	switch event.Code {
	case "353":
		matterMessage.UserName = b.Config.IRC.Nick
		matterMessage.Text = event.Message() + " currently on IRC"
	}
	b.m.Send(matterMessage)
}

func (b *Bridge) handleMatter() {
	for {
		message := b.m.Receive()
		switch message.Text {
		case "!users":
			log.Println("received !users from", message.UserName)
			b.i.SendRaw("NAMES " + b.Config.IRC.Channel)
		}
		b.i.Privmsg(b.Config.IRC.Channel, message.UserName+": "+message.Text)
	}
}

func main() {
	NewBridge("matterbot", NewConfig("matterbridge.conf"))
	select {}
}
