package bslack

import (
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
	mh *matterhook.Client
	sc *slack.Client
	//	MMapi
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
	b.Plus = config.UseAPI
	return b
}

func (b *Bslack) Command(cmd string) string {
	return ""
}

func (b *Bslack) Connect() error {
	if !b.Plus {
		b.mh = matterhook.New(b.Config.URL,
			matterhook.Config{BindAddress: b.Config.BindAddress})
	} else {
		b.sc = slack.New(b.Config.Token)
		flog.Infof("Trying login on slack with Token")
		/*
			if err != nil {
				return err
			}
		*/
		flog.Info("Login ok")
	}
	b.rtm = b.sc.NewRTM()
	go b.rtm.ManageConnection()
	go b.handleSlack()
	return nil
}

func (b *Bslack) FullOrigin() string {
	return b.protocol + "." + b.origin
}

func (b *Bslack) JoinChannel(channel string) error {
	schannel := b.getChannelByName(channel)
	if schannel != nil && !schannel.IsMember {
		flog.Infof("Joining %s", channel)
		b.sc.JoinChannel(schannel.ID)
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
	flog.Infof("slack send %#v", msg)
	if msg.Origin != "slack" {
		return b.SendType(msg.Username, msg.Text, msg.Channel, "")
	}
	return nil
}

func (b *Bslack) SendType(nick string, message string, channel string, mtype string) error {
	if b.Config.PrefixMessagesWithNick {
		message = nick + " " + message
	}
	if !b.Plus {
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
		flog.Debug("->slack channel: ", channel, " ", message)
		return nil
	}
	flog.Debugf("sent to slack channel API: %s %s", channel, message)
	newmsg := b.rtm.NewOutgoingMessage(message, b.getChannelByName(channel).ID)
	b.rtm.SendMessage(newmsg)
	return nil
}

func (b *Bslack) getChannelByName(name string) *slack.Channel {
	if b.channels == nil {
		return nil
	}
	for _, channel := range b.channels {
		if channel.Name == name {
			return &channel
		}
	}
	return nil
}

func (b *Bslack) handleSlack() {
	flog.Infof("Choosing API based slack connection: %t", b.Plus)
	mchan := make(chan *MMMessage)
	if b.Plus {
		go b.handleSlackClient(mchan)
	} else {
		go b.handleMatterHook(mchan)
	}
	time.Sleep(time.Second)
	flog.Info("Start listening for Slack messages")
	for message := range mchan {
		texts := strings.Split(message.Text, "\n")
		for _, text := range texts {
			flog.Debug("Sending message from " + message.Username + " to " + message.Channel)
			b.Remote <- config.Message{Text: text, Username: message.Username, Channel: message.Channel, Origin: b.origin, Protocol: b.protocol, FullOrigin: b.FullOrigin()}
		}
	}
}

func (b *Bslack) handleSlackClient(mchan chan *MMMessage) {
	for msg := range b.rtm.IncomingEvents {
		switch ev := msg.Data.(type) {
		case *slack.MessageEvent:
			flog.Debugf("%#v", ev)
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
		flog.Debugf("receiving from slack %#v", message)
		m := &MMMessage{}
		m.Username = message.UserName
		m.Text = message.Text
		m.Channel = message.ChannelName
		mchan <- m
	}
}
