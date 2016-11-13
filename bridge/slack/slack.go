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
	Raw      *slack.MessageEvent
}

type Bslack struct {
	mh       *matterhook.Client
	sc       *slack.Client
	Config   *config.Protocol
	rtm      *slack.RTM
	Plus     bool
	Remote   chan config.Message
	Users    []slack.User
	Account  string
	si       *slack.Info
	channels []slack.Channel
}

var flog *log.Entry
var protocol = "slack"

func init() {
	flog = log.WithFields(log.Fields{"module": protocol})
}

func New(cfg config.Protocol, account string, c chan config.Message) *Bslack {
	b := &Bslack{}
	b.Config = &cfg
	b.Remote = c
	b.Account = account
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

func (b *Bslack) Send(msg config.Message) error {
	flog.Debugf("Receiving %#v", msg)
	if msg.Account == b.Account {
		return nil
	}
	nick := msg.Username
	message := msg.Text
	channel := msg.Channel
	if b.Config.PrefixMessagesWithNick {
		message = nick + " " + message
	}
	if !b.Config.UseAPI {
		matterMessage := matterhook.OMessage{IconURL: b.Config.IconURL}
		matterMessage.Channel = channel
		matterMessage.UserName = nick
		matterMessage.Type = ""
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
	np := slack.NewPostMessageParameters()
	if b.Config.PrefixMessagesWithNick == true {
		np.AsUser = true
	}
	np.Username = nick
	np.IconURL = config.GetIconURL(&msg, b.Config)
	if msg.Avatar != "" {
		np.IconURL = msg.Avatar
	}
	b.sc.PostMessage(schannel.ID, message, np)

	/*
	   newmsg := b.rtm.NewOutgoingMessage(message, schannel.ID)
	   b.rtm.SendMessage(newmsg)
	*/

	return nil
}

func (b *Bslack) getAvatar(user string) string {
	var avatar string
	if b.Users != nil {
		for _, u := range b.Users {
			if user == u.Name {
				return u.Profile.Image48
			}
		}
	}
	return avatar
}

func (b *Bslack) getChannelByName(name string) (*slack.Channel, error) {
	if b.channels == nil {
		return nil, fmt.Errorf("%s: channel %s not found (no channels found)", b.Account, name)
	}
	for _, channel := range b.channels {
		if channel.Name == name {
			return &channel, nil
		}
	}
	return nil, fmt.Errorf("%s: channel %s not found", b.Account, name)
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
		// do not send messages from ourself
		if message.Username == b.si.User.Name {
			continue
		}
		texts := strings.Split(message.Text, "\n")
		for _, text := range texts {
			flog.Debugf("Sending message from %s on %s to gateway", message.Username, b.Account)
			b.Remote <- config.Message{Text: text, Username: message.Username, Channel: message.Channel, Account: b.Account, Avatar: b.getAvatar(message.Username)}
		}
	}
}

func (b *Bslack) handleSlackClient(mchan chan *MMMessage) {
	count := 0
	for msg := range b.rtm.IncomingEvents {
		switch ev := msg.Data.(type) {
		case *slack.MessageEvent:
			// ignore first message
			if count > 0 {
				flog.Debugf("Receiving from slackclient %#v", ev)
				//ev.ReplyTo
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
				m.Raw = ev
				mchan <- m
			}
			count++
		case *slack.OutgoingErrorEvent:
			flog.Debugf("%#v", ev.Error())
		case *slack.ConnectedEvent:
			b.channels = ev.Info.Channels
			b.si = ev.Info
			b.Users, _ = b.sc.GetUsers()
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
