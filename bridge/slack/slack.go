package bslack

import (
	"errors"
	"fmt"
	"github.com/42wim/matterbridge/bridge/config"
	"github.com/42wim/matterbridge/matterhook"
	log "github.com/Sirupsen/logrus"
	"github.com/nlopes/slack"
	"html"
	"regexp"
	"strings"
	"time"
)

type MMMessage struct {
	Text     string
	Channel  string
	Username string
	UserID   string
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
	BotID    string
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
	if b.Config.WebhookBindAddress != "" {
		if b.Config.WebhookURL != "" {
			flog.Info("Connecting using webhookurl (sending) and webhookbindaddress (receiving)")
			b.mh = matterhook.New(b.Config.WebhookURL,
				matterhook.Config{InsecureSkipVerify: b.Config.SkipTLSVerify,
					BindAddress: b.Config.WebhookBindAddress})
		} else if b.Config.Token != "" {
			flog.Info("Connecting using token (sending)")
			b.sc = slack.New(b.Config.Token)
			b.rtm = b.sc.NewRTM()
			go b.rtm.ManageConnection()
			flog.Info("Connecting using webhookbindaddress (receiving)")
			b.mh = matterhook.New(b.Config.WebhookURL,
				matterhook.Config{InsecureSkipVerify: b.Config.SkipTLSVerify,
					BindAddress: b.Config.WebhookBindAddress})
		} else {
			flog.Info("Connecting using webhookbindaddress (receiving)")
			b.mh = matterhook.New(b.Config.WebhookURL,
				matterhook.Config{InsecureSkipVerify: b.Config.SkipTLSVerify,
					BindAddress: b.Config.WebhookBindAddress})
		}
		go b.handleSlack()
		return nil
	}
	if b.Config.WebhookURL != "" {
		flog.Info("Connecting using webhookurl (sending)")
		b.mh = matterhook.New(b.Config.WebhookURL,
			matterhook.Config{InsecureSkipVerify: b.Config.SkipTLSVerify,
				DisableServer: true})
		if b.Config.Token != "" {
			flog.Info("Connecting using token (receiving)")
			b.sc = slack.New(b.Config.Token)
			b.rtm = b.sc.NewRTM()
			go b.rtm.ManageConnection()
			go b.handleSlack()
		}
	} else if b.Config.Token != "" {
		flog.Info("Connecting using token (sending and receiving)")
		b.sc = slack.New(b.Config.Token)
		b.rtm = b.sc.NewRTM()
		go b.rtm.ManageConnection()
		go b.handleSlack()
	}
	if b.Config.WebhookBindAddress == "" && b.Config.WebhookURL == "" && b.Config.Token == "" {
		return errors.New("No connection method found. See that you have WebhookBindAddress, WebhookURL or Token configured.")
	}
	return nil
}

func (b *Bslack) Disconnect() error {
	return nil

}

func (b *Bslack) JoinChannel(channel string) error {
	// we can only join channels using the API
	if b.Config.WebhookURL == "" && b.Config.WebhookBindAddress == "" {
		if strings.HasPrefix(b.Config.Token, "xoxb") {
			// TODO check if bot has already joined channel
			return nil
		}
		_, err := b.sc.JoinChannel(channel)
		if err != nil {
			if err.Error() != "name_taken" {
				return err
			}
		}
	}
	return nil
}

func (b *Bslack) Send(msg config.Message) error {
	flog.Debugf("Receiving %#v", msg)
	nick := msg.Username
	message := msg.Text
	channel := msg.Channel
	if b.Config.PrefixMessagesWithNick {
		message = nick + " " + message
	}
	if b.Config.WebhookURL != "" {
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
	if b.Config.PrefixMessagesWithNick {
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

func (b *Bslack) getChannelByID(ID string) (*slack.Channel, error) {
	if b.channels == nil {
		return nil, fmt.Errorf("%s: channel %s not found (no channels found)", b.Account, ID)
	}
	for _, channel := range b.channels {
		if channel.ID == ID {
			return &channel, nil
		}
	}
	return nil, fmt.Errorf("%s: channel %s not found", b.Account, ID)
}

func (b *Bslack) handleSlack() {
	mchan := make(chan *MMMessage)
	if b.Config.WebhookBindAddress != "" {
		flog.Debugf("Choosing webhooks based receiving")
		go b.handleMatterHook(mchan)
	} else {
		flog.Debugf("Choosing token based receiving")
		go b.handleSlackClient(mchan)
	}
	time.Sleep(time.Second)
	flog.Debug("Start listening for Slack messages")
	for message := range mchan {
		// do not send messages from ourself
		if b.Config.WebhookURL == "" && b.Config.WebhookBindAddress == "" && (message.Username == b.si.User.Name || message.UserID == b.BotID) {
			continue
		}
		texts := strings.Split(message.Text, "\n")
		for _, text := range texts {
			text = b.replaceURL(text)
			text = html.UnescapeString(text)
			flog.Debugf("Sending message from %s on %s to gateway", message.Username, b.Account)
			msg := config.Message{Text: text, Username: message.Username, Channel: message.Channel, Account: b.Account, Avatar: b.getAvatar(message.Username), UserID: message.UserID}
			b.Remote <- msg
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
				if !b.Config.EditDisable && ev.SubMessage != nil {
					flog.Debugf("SubMessage %#v", ev.SubMessage)
					ev.User = ev.SubMessage.User
					ev.Text = ev.SubMessage.Text + b.Config.EditSuffix
				}
				// use our own func because rtm.GetChannelInfo doesn't work for private channels
				channel, err := b.getChannelByID(ev.Channel)
				if err != nil {
					continue
				}
				m := &MMMessage{}
				if ev.BotID == "" {
					user, err := b.rtm.GetUserInfo(ev.User)
					if err != nil {
						continue
					}
					m.UserID = user.ID
					m.Username = user.Name
				}
				m.Channel = channel.Name
				m.Text = ev.Text
				if m.Text == "" {
					for _, attach := range ev.Attachments {
						if attach.Text != "" {
							m.Text = attach.Text
						} else {
							m.Text = attach.Fallback
						}
					}
				}
				m.Raw = ev
				m.Text = b.replaceMention(m.Text)
				if ev.BotID != "" {
					bot, err := b.rtm.GetBotInfo(ev.BotID)
					if err != nil {
						continue
					}
					if bot.Name != "" {
						m.Username = bot.Name
						m.UserID = bot.ID
					}
				}
				mchan <- m
			}
			count++
		case *slack.OutgoingErrorEvent:
			flog.Debugf("%#v", ev.Error())
		case *slack.ChannelJoinedEvent:
			b.Users, _ = b.sc.GetUsers()
		case *slack.ConnectedEvent:
			b.channels = ev.Info.Channels
			b.si = ev.Info
			for _, bot := range b.si.Bots {
				if bot.Name == "Slack API Tester" {
					b.BotID = bot.ID
					flog.Debugf("my bot ID is %#v", bot.ID)
				}
			}
			b.Users, _ = b.sc.GetUsers()
			// add private channels
			groups, _ := b.sc.GetGroups(true)
			for _, g := range groups {
				channel := new(slack.Channel)
				channel.ID = g.ID
				channel.Name = g.Name
				b.channels = append(b.channels, *channel)
			}
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
		m.Text = b.replaceMention(m.Text)
		m.Channel = message.ChannelName
		if m.Username == "slackbot" {
			continue
		}
		mchan <- m
	}
}

func (b *Bslack) userName(id string) string {
	for _, u := range b.Users {
		if u.ID == id {
			return u.Name
		}
	}
	return ""
}

func (b *Bslack) replaceMention(text string) string {
	results := regexp.MustCompile(`<@([a-zA-z0-9]+)>`).FindAllStringSubmatch(text, -1)
	for _, r := range results {
		text = strings.Replace(text, "<@"+r[1]+">", "@"+b.userName(r[1]), -1)

	}
	return text
}

func (b *Bslack) replaceURL(text string) string {
	results := regexp.MustCompile(`<(.*?)\|.*?>`).FindAllStringSubmatch(text, -1)
	for _, r := range results {
		text = strings.Replace(text, r[0], r[1], -1)
	}
	return text
}
