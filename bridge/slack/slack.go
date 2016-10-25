package bslack

import (
	"fmt"
	"github.com/42wim/matterbridge/bridge/config"
	"github.com/42wim/matterbridge/matterhook"
	log "github.com/Sirupsen/logrus"
	"github.com/nlopes/slack"
	"strings"
	"time"
)

type MMMessage struct {
	Text     string
	Channel  string
	Username string
}

type Bslack struct {
	mh       *matterhook.Client
	sc       *slack.Client
	Config   *config.Protocol
	rtm      *slack.RTM
	Plus     bool
	Remote   chan config.Message
	protocol string
	origin   string
	channels []slack.Channel
}

var flog *log.Entry
var protocol = "slack"

func init() {
	flog = log.WithFields(log.Fields{"module": protocol})
}

func New(config config.Protocol, origin string, c chan config.Message) *Bslack {
	b := &Bslack{}
	b.Config = &config
	b.Remote = c
	b.protocol = protocol
	b.origin = origin
	b.Config.UseAPI = config.UseAPI
	return b
}

func (b *Bslack) Command(cmd string) string {
	return ""
}

func (b *Bslack) Connect() error {
	flog.Info("Connecting")
	if !b.Config.UseAPI {
		b.mh = matterhook.New(b.Config.URL,
			matterhook.Config{BindAddress: b.Config.BindAddress})
	} else {
		b.sc = slack.New(b.Config.Token)
		b.rtm = b.sc.NewRTM()
		go b.rtm.ManageConnection()
	}
	flog.Info("Connection succeeded")
	go b.handleSlack()
	return nil
}

func (b *Bslack) FullOrigin() string {
	return b.protocol + "." + b.origin
}

func (b *Bslack) JoinChannel(channel string) error {
	// we can only join channels using the API
	if b.Config.UseAPI {
		_, err := b.sc.JoinChannel(channel)
		if err != nil {
			return err
		}
	}
	return nil
}

func (b *Bslack) Name() string {
	return b.protocol + "." + b.origin
}

func (b *Bslack) Protocol() string {
	return b.protocol
}

func (b *Bslack) Origin() string {
	return b.origin
}

func (b *Bslack) Send(msg config.Message) error {
	flog.Debugf("Receiving %#v", msg)
	if msg.FullOrigin != b.FullOrigin() {
		return b.SendType(msg.Username, msg.Text, msg.Channel, "")
	}
	return nil
}

func (b *Bslack) SendType(nick string, message string, channel string, mtype string) error {
	if b.Config.PrefixMessagesWithNick {
		message = nick + " " + message
	}
	if !b.Config.UseAPI {
		matterMessage := matterhook.OMessage{IconURL: b.Config.IconURL}
		matterMessage.Channel = channel
		matterMessage.UserName = nick
		matterMessage.Type = mtype
		matterMessage.Text = message
		err := b.mh.Send(matterMessage)
		if err != nil {
			flog.Info(err)
			return err
		}
		return nil
	}
	schannel, err := b.getChannelByName(channel)
	if err != nil {
		return err
	}
	newmsg := b.rtm.NewOutgoingMessage(message, schannel.ID)
	b.rtm.SendMessage(newmsg)
	return nil
}

func (b *Bslack) getChannelByName(name string) (*slack.Channel, error) {
	if b.channels == nil {
		return nil, fmt.Errorf("%s: channel %s not found (no channels found)", b.FullOrigin(), name)
	}
	for _, channel := range b.channels {
		if channel.Name == name {
			return &channel, nil
		}
	}
	return nil, fmt.Errorf("%s: channel %s not found", b.FullOrigin(), name)
}

func (b *Bslack) handleSlack() {
	flog.Debugf("Choosing API based slack connection: %t", b.Config.UseAPI)
	mchan := make(chan *MMMessage)
	if b.Config.UseAPI {
		go b.handleSlackClient(mchan)
	} else {
		go b.handleMatterHook(mchan)
	}
	time.Sleep(time.Second)
	flog.Debug("Start listening for Slack messages")
	for message := range mchan {
		texts := strings.Split(message.Text, "\n")
		for _, text := range texts {
			flog.Debugf("Sending message from %s on %s to gateway", message.Username, b.FullOrigin())
			b.Remote <- config.Message{Text: text, Username: message.Username, Channel: message.Channel, Origin: b.origin, Protocol: b.protocol, FullOrigin: b.FullOrigin()}
		}
	}
}

func (b *Bslack) handleSlackClient(mchan chan *MMMessage) {
	for msg := range b.rtm.IncomingEvents {
		switch ev := msg.Data.(type) {
		case *slack.MessageEvent:
			flog.Debugf("Receiving from slackclient %#v", ev)
			channel, err := b.rtm.GetChannelInfo(ev.Channel)
			if err != nil {
				continue
			}
			user, err := b.rtm.GetUserInfo(ev.User)
			if err != nil {
				continue
			}
			m := &MMMessage{}
			m.Username = user.Name
			m.Channel = channel.Name
			m.Text = ev.Text
			mchan <- m
		case *slack.OutgoingErrorEvent:
			flog.Debugf("%#v", ev.Error())
		case *slack.ConnectedEvent:
			b.channels = ev.Info.Channels
		case *slack.InvalidAuthEvent:
			flog.Fatalf("Invalid Token %#v", ev)
		default:
		}
	}
}

func (b *Bslack) handleMatterHook(mchan chan *MMMessage) {
	for {
		message := b.mh.Receive()
		flog.Debugf("receiving from matterhook (slack) %#v", message)
		m := &MMMessage{}
		m.Username = message.UserName
		m.Text = message.Text
		m.Channel = message.ChannelName
		mchan <- m
	}
}
