package bmattermost

import (
	"errors"
	"github.com/42wim/matterbridge/bridge/config"
	"github.com/42wim/matterbridge/matterclient"
	"github.com/42wim/matterbridge/matterhook"
	log "github.com/Sirupsen/logrus"
)

type MMhook struct {
	mh *matterhook.Client
}

type MMapi struct {
	mc    *matterclient.MMClient
	mmMap map[string]string
}

type MMMessage struct {
	Text     string
	Channel  string
	Username string
	UserID   string
}

type Bmattermost struct {
	MMhook
	MMapi
	Config  *config.Protocol
	Remote  chan config.Message
	TeamId  string
	Account string
}

var flog *log.Entry
var protocol = "mattermost"

func init() {
	flog = log.WithFields(log.Fields{"module": protocol})
}

func New(cfg config.Protocol, account string, c chan config.Message) *Bmattermost {
	b := &Bmattermost{}
	b.Config = &cfg
	b.Remote = c
	b.Account = account
	b.mmMap = make(map[string]string)
	return b
}

func (b *Bmattermost) Command(cmd string) string {
	return ""
}

func (b *Bmattermost) Connect() error {
	if b.Config.WebhookBindAddress != "" {
		if b.Config.WebhookURL != "" {
			flog.Info("Connecting using webhookurl (sending) and webhookbindaddress (receiving)")
			b.mh = matterhook.New(b.Config.WebhookURL,
				matterhook.Config{InsecureSkipVerify: b.Config.SkipTLSVerify,
					BindAddress: b.Config.WebhookBindAddress})
		} else if b.Config.Login != "" {
			flog.Info("Connecting using login/password (sending)")
			err := b.apiLogin()
			if err != nil {
				return err
			}
		} else {
			flog.Info("Connecting using webhookbindaddress (receiving)")
			b.mh = matterhook.New(b.Config.WebhookURL,
				matterhook.Config{InsecureSkipVerify: b.Config.SkipTLSVerify,
					BindAddress: b.Config.WebhookBindAddress})
		}
		go b.handleMatter()
		return nil
	}
	if b.Config.WebhookURL != "" {
		flog.Info("Connecting using webhookurl (sending)")
		b.mh = matterhook.New(b.Config.WebhookURL,
			matterhook.Config{InsecureSkipVerify: b.Config.SkipTLSVerify,
				DisableServer: true})
		if b.Config.Login != "" {
			flog.Info("Connecting using login/password (receiving)")
			err := b.apiLogin()
			if err != nil {
				return err
			}
			go b.handleMatter()
		}
		return nil
	} else if b.Config.Login != "" {
		flog.Info("Connecting using login/password (sending and receiving)")
		err := b.apiLogin()
		if err != nil {
			return err
		}
		go b.handleMatter()
	}
	if b.Config.WebhookBindAddress == "" && b.Config.WebhookURL == "" && b.Config.Login == "" {
		return errors.New("No connection method found. See that you have WebhookBindAddress, WebhookURL or Login/Password/Server/Team configured.")
	}
	return nil
}

func (b *Bmattermost) Disconnect() error {
	return nil
}

func (b *Bmattermost) JoinChannel(channel string) error {
	// we can only join channels using the API
	if b.Config.WebhookURL == "" && b.Config.WebhookBindAddress == "" {
		return b.mc.JoinChannel(b.mc.GetChannelId(channel, ""))
	}
	return nil
}

func (b *Bmattermost) Send(msg config.Message) error {
	flog.Debugf("Receiving %#v", msg)
	nick := msg.Username
	message := msg.Text
	channel := msg.Channel

	if b.Config.PrefixMessagesWithNick {
		message = nick + message
	}
	if b.Config.WebhookURL != "" {
		matterMessage := matterhook.OMessage{IconURL: b.Config.IconURL}
		matterMessage.IconURL = msg.Avatar
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
	b.mc.PostMessage(b.mc.GetChannelId(channel, ""), message)
	return nil
}

func (b *Bmattermost) handleMatter() {
	mchan := make(chan *MMMessage)
	if b.Config.WebhookBindAddress != "" {
		flog.Debugf("Choosing webhooks based receiving")
		go b.handleMatterHook(mchan)
	} else {
		flog.Debugf("Choosing login/password based receiving")
		go b.handleMatterClient(mchan)
	}
	for message := range mchan {
		flog.Debugf("Sending message from %s on %s to gateway", message.Username, b.Account)
		b.Remote <- config.Message{Text: message.Text, Username: message.Username, Channel: message.Channel, Account: b.Account, UserID: message.UserID}
	}
}

func (b *Bmattermost) handleMatterClient(mchan chan *MMMessage) {
	for message := range b.mc.MessageChan {
		flog.Debugf("%#v", message.Raw.Data)
		if message.Type == "system_join_leave" ||
			message.Type == "system_join_channel" ||
			message.Type == "system_leave_channel" {
			flog.Debugf("Sending JOIN_LEAVE event from %s to gateway", b.Account)
			b.Remote <- config.Message{Username: "system", Text: message.Text, Channel: message.Channel, Account: b.Account, Event: config.EVENT_JOIN_LEAVE}
			continue
		}
		if (message.Raw.Event == "post_edited") && b.Config.EditDisable {
			continue
		}
		// do not post our own messages back to irc
		// only listen to message from our team
		if (message.Raw.Event == "posted" || message.Raw.Event == "post_edited") &&
			b.mc.User.Username != message.Username && message.Raw.Data["team_id"].(string) == b.TeamId {
			flog.Debugf("Receiving from matterclient %#v", message)
			m := &MMMessage{}
			m.UserID = message.UserID
			m.Username = message.Username
			m.Channel = message.Channel
			m.Text = message.Text
			if message.Raw.Event == "post_edited" && !b.Config.EditDisable {
				m.Text = message.Text + b.Config.EditSuffix
			}
			if len(message.Post.FileIds) > 0 {
				for _, link := range b.mc.GetFileLinks(message.Post.FileIds) {
					m.Text = m.Text + "\n" + link
				}
			}
			mchan <- m
		}
	}
}

func (b *Bmattermost) handleMatterHook(mchan chan *MMMessage) {
	for {
		message := b.mh.Receive()
		flog.Debugf("Receiving from matterhook %#v", message)
		m := &MMMessage{}
		m.UserID = message.UserID
		m.Username = message.UserName
		m.Text = message.Text
		m.Channel = message.ChannelName
		mchan <- m
	}
}

func (b *Bmattermost) apiLogin() error {
	b.mc = matterclient.New(b.Config.Login, b.Config.Password,
		b.Config.Team, b.Config.Server)
	b.mc.SkipTLSVerify = b.Config.SkipTLSVerify
	b.mc.NoTLS = b.Config.NoTLS
	flog.Infof("Connecting %s (team: %s) on %s", b.Config.Login, b.Config.Team, b.Config.Server)
	err := b.mc.Login()
	if err != nil {
		return err
	}
	flog.Info("Connection succeeded")
	b.TeamId = b.mc.GetTeamId()
	go b.mc.WsReceiver()
	go b.mc.StatusLoop()
	return nil
}
