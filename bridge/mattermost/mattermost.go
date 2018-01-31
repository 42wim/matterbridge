package bmattermost

import (
	"errors"
	"fmt"
	"github.com/42wim/matterbridge/bridge/config"
	"github.com/42wim/matterbridge/matterclient"
	"github.com/42wim/matterbridge/matterhook"
	log "github.com/Sirupsen/logrus"
	"strings"
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
	ID       string
	Event    string
	Extra    map[string][]interface{}
}

type Bmattermost struct {
	MMhook
	MMapi
	TeamId string
	*config.BridgeConfig
}

var flog *log.Entry
var protocol = "mattermost"

func init() {
	flog = log.WithFields(log.Fields{"module": protocol})
}

func New(cfg *config.BridgeConfig) *Bmattermost {
	b := &Bmattermost{BridgeConfig: cfg}
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
		} else if b.Config.Token != "" {
			flog.Info("Connecting using token (sending)")
			err := b.apiLogin()
			if err != nil {
				return err
			}
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
		if b.Config.Token != "" {
			flog.Info("Connecting using token (receiving)")
			err := b.apiLogin()
			if err != nil {
				return err
			}
			go b.handleMatter()
		} else if b.Config.Login != "" {
			flog.Info("Connecting using login/password (receiving)")
			err := b.apiLogin()
			if err != nil {
				return err
			}
			go b.handleMatter()
		}
		return nil
	} else if b.Config.Token != "" {
		flog.Info("Connecting using token (sending and receiving)")
		err := b.apiLogin()
		if err != nil {
			return err
		}
		go b.handleMatter()
	} else if b.Config.Login != "" {
		flog.Info("Connecting using login/password (sending and receiving)")
		err := b.apiLogin()
		if err != nil {
			return err
		}
		go b.handleMatter()
	}
	if b.Config.WebhookBindAddress == "" && b.Config.WebhookURL == "" && b.Config.Login == "" && b.Config.Token == "" {
		return errors.New("No connection method found. See that you have WebhookBindAddress, WebhookURL or Token/Login/Password/Server/Team configured.")
	}
	return nil
}

func (b *Bmattermost) Disconnect() error {
	return nil
}

func (b *Bmattermost) JoinChannel(channel config.ChannelInfo) error {
	// we can only join channels using the API
	if b.Config.WebhookURL == "" && b.Config.WebhookBindAddress == "" {
		id := b.mc.GetChannelId(channel.Name, "")
		if id == "" {
			return fmt.Errorf("Could not find channel ID for channel %s", channel.Name)
		}
		return b.mc.JoinChannel(id)
	}
	return nil
}

func (b *Bmattermost) Send(msg config.Message) (string, error) {
	flog.Debugf("Receiving %#v", msg)
	if msg.Event == config.EVENT_USER_ACTION {
		msg.Text = "*" + msg.Text + "*"
	}
	nick := msg.Username
	message := msg.Text
	channel := msg.Channel

	if b.Config.PrefixMessagesWithNick {
		message = nick + message
	}
	if b.Config.WebhookURL != "" {

		if msg.Extra != nil {
			if len(msg.Extra["file"]) > 0 {
				for _, f := range msg.Extra["file"] {
					fi := f.(config.FileInfo)
					if fi.URL != "" {
						message += fi.URL
					}
				}
			}
		}

		matterMessage := matterhook.OMessage{IconURL: b.Config.IconURL}
		matterMessage.IconURL = msg.Avatar
		matterMessage.Channel = channel
		matterMessage.UserName = nick
		matterMessage.Type = ""
		matterMessage.Text = message
		matterMessage.Props = make(map[string]interface{})
		matterMessage.Props["matterbridge"] = true
		err := b.mh.Send(matterMessage)
		if err != nil {
			flog.Info(err)
			return "", err
		}
		return "", nil
	}
	if msg.Event == config.EVENT_MSG_DELETE {
		if msg.ID == "" {
			return "", nil
		}
		return msg.ID, b.mc.DeleteMessage(msg.ID)
	}
	if msg.Extra != nil {
		if len(msg.Extra["file"]) > 0 {
			var err error
			var res, id string
			for _, f := range msg.Extra["file"] {
				fi := f.(config.FileInfo)
				id, err = b.mc.UploadFile(*fi.Data, b.mc.GetChannelId(channel, ""), fi.Name)
				if err != nil {
					flog.Debugf("ERROR %#v", err)
					return "", err
				}
				message = fi.Comment
				if b.Config.PrefixMessagesWithNick {
					message = nick + fi.Comment
				}
				res, err = b.mc.PostMessageWithFiles(b.mc.GetChannelId(channel, ""), message, []string{id})
			}
			return res, err
		}
	}
	if msg.ID != "" {
		return b.mc.EditMessage(msg.ID, message)
	}
	return b.mc.PostMessage(b.mc.GetChannelId(channel, ""), message)
}

func (b *Bmattermost) handleMatter() {
	mchan := make(chan *MMMessage)
	if b.Config.WebhookBindAddress != "" {
		flog.Debugf("Choosing webhooks based receiving")
		go b.handleMatterHook(mchan)
	} else {
		if b.Config.Token != "" {
			flog.Debugf("Choosing token based receiving")
		} else {
			flog.Debugf("Choosing login/password based receiving")
		}
		go b.handleMatterClient(mchan)
	}
	for message := range mchan {
		rmsg := config.Message{Username: message.Username, Channel: message.Channel, Account: b.Account, UserID: message.UserID, ID: message.ID, Event: message.Event, Extra: message.Extra}
		text, ok := b.replaceAction(message.Text)
		if ok {
			rmsg.Event = config.EVENT_USER_ACTION
		}
		rmsg.Text = text
		flog.Debugf("Sending message from %s on %s to gateway", message.Username, b.Account)
		flog.Debugf("Message is %#v", rmsg)
		b.Remote <- rmsg
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

		m := &MMMessage{Extra: make(map[string][]interface{})}

		props := message.Post.Props
		if props != nil {
			if _, ok := props["matterbridge"].(bool); ok {
				flog.Debugf("sent by matterbridge, ignoring")
				continue
			}
			if _, ok := props["override_username"].(string); ok {
				message.Username = props["override_username"].(string)
			}
			if _, ok := props["attachments"].([]interface{}); ok {
				m.Extra["attachments"] = props["attachments"].([]interface{})
			}
		}
		// do not post our own messages back to irc
		// only listen to message from our team
		if (message.Raw.Event == "posted" || message.Raw.Event == "post_edited" || message.Raw.Event == "post_deleted") &&
			b.mc.User.Username != message.Username && message.Raw.Data["team_id"].(string) == b.TeamId {
			// if the message has reactions don't repost it (for now, until we can correlate reaction with message)
			if message.Post.HasReactions {
				continue
			}
			flog.Debugf("Receiving from matterclient %#v", message)
			m.UserID = message.UserID
			m.Username = message.Username
			m.Channel = message.Channel
			m.Text = message.Text
			m.ID = message.Post.Id
			if message.Raw.Event == "post_edited" && !b.Config.EditDisable {
				m.Text = message.Text + b.Config.EditSuffix
			}
			if message.Raw.Event == "post_deleted" {
				m.Event = config.EVENT_MSG_DELETE
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
	password := b.Config.Password
	if b.Config.Token != "" {
		password = "MMAUTHTOKEN=" + b.Config.Token
	}

	b.mc = matterclient.New(b.Config.Login, password,
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

func (b *Bmattermost) replaceAction(text string) (string, bool) {
	if strings.HasPrefix(text, "*") && strings.HasSuffix(text, "*") {
		return strings.Replace(text, "*", "", -1), true
	}
	return text, false
}
