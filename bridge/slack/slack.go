package bslack

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/42wim/matterbridge/bridge/config"
	"github.com/42wim/matterbridge/matterhook"
	log "github.com/Sirupsen/logrus"
	"github.com/matterbridge/slack"
	"html"
	"io"
	"net/http"
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
	rtm      *slack.RTM
	Plus     bool
	Users    []slack.User
	si       *slack.Info
	channels []slack.Channel
	*config.BridgeConfig
}

var flog *log.Entry
var protocol = "slack"

func init() {
	flog = log.WithFields(log.Fields{"module": protocol})
}

func New(cfg *config.BridgeConfig) *Bslack {
	return &Bslack{BridgeConfig: cfg}
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

func (b *Bslack) JoinChannel(channel config.ChannelInfo) error {
	// we can only join channels using the API
	if b.Config.WebhookURL == "" && b.Config.WebhookBindAddress == "" {
		if strings.HasPrefix(b.Config.Token, "xoxb") {
			// TODO check if bot has already joined channel
			return nil
		}
		_, err := b.sc.JoinChannel(channel.Name)
		if err != nil {
			if err.Error() != "name_taken" {
				return err
			}
		}
	}
	return nil
}

func (b *Bslack) Send(msg config.Message) (string, error) {
	flog.Debugf("Receiving %#v", msg)
	if msg.Event == config.EVENT_USER_ACTION {
		msg.Text = "_" + msg.Text + "_"
	}
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
			return "", err
		}
		return "", nil
	}
	schannel, err := b.getChannelByName(channel)
	if err != nil {
		return "", err
	}
	np := slack.NewPostMessageParameters()
	if b.Config.PrefixMessagesWithNick {
		np.AsUser = true
	}
	np.Username = nick
	np.IconURL = config.GetIconURL(&msg, &b.Config)
	if msg.Avatar != "" {
		np.IconURL = msg.Avatar
	}
	np.Attachments = append(np.Attachments, slack.Attachment{CallbackID: "matterbridge"})
	np.Attachments = append(np.Attachments, b.createAttach(msg.Extra)...)

	// replace mentions
	np.LinkNames = 1

	if msg.Event == config.EVENT_MSG_DELETE {
		// some protocols echo deletes, but with empty ID
		if msg.ID == "" {
			return "", nil
		}
		// we get a "slack <ID>", split it
		ts := strings.Fields(msg.ID)
		b.sc.DeleteMessage(schannel.ID, ts[1])
		return "", nil
	}
	// if we have no ID it means we're creating a new message, not updating an existing one
	if msg.ID != "" {
		ts := strings.Fields(msg.ID)
		b.sc.UpdateMessage(schannel.ID, ts[1], message)
		return "", nil
	}

	if msg.Extra != nil {
		// check if we have files to upload (from slack, telegram or mattermost)
		if len(msg.Extra["file"]) > 0 {
			var err error
			for _, f := range msg.Extra["file"] {
				fi := f.(config.FileInfo)
				_, err = b.sc.UploadFile(slack.FileUploadParameters{
					Reader:         bytes.NewReader(*fi.Data),
					Filename:       fi.Name,
					Channels:       []string{schannel.ID},
					InitialComment: fi.Comment,
				})
				if err != nil {
					flog.Errorf("uploadfile %#v", err)
				}
			}
		}
	}

	_, id, err := b.sc.PostMessage(schannel.ID, message, np)
	if err != nil {
		return "", err
	}
	return "slack " + id, nil
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
		if b.Config.WebhookURL == "" && b.Config.WebhookBindAddress == "" && message.Username == b.si.User.Name {
			continue
		}
		if (message.Text == "" || message.Username == "") && message.Raw.SubType != "message_deleted" {
			continue
		}
		text := message.Text
		text = b.replaceURL(text)
		text = html.UnescapeString(text)
		flog.Debugf("Sending message from %s on %s to gateway", message.Username, b.Account)
		msg := config.Message{Text: text, Username: message.Username, Channel: message.Channel, Account: b.Account, Avatar: b.getAvatar(message.Username), UserID: message.UserID, ID: "slack " + message.Raw.Timestamp, Extra: make(map[string][]interface{})}
		if message.Raw.SubType == "me_message" {
			msg.Event = config.EVENT_USER_ACTION
		}
		if message.Raw.SubType == "channel_leave" || message.Raw.SubType == "channel_join" {
			msg.Username = "system"
			msg.Event = config.EVENT_JOIN_LEAVE
		}
		// edited messages have a submessage, use this timestamp
		if message.Raw.SubMessage != nil {
			msg.ID = "slack " + message.Raw.SubMessage.Timestamp
		}
		if message.Raw.SubType == "message_deleted" {
			msg.Text = config.EVENT_MSG_DELETE
			msg.Event = config.EVENT_MSG_DELETE
			msg.ID = "slack " + message.Raw.DeletedTimestamp
		}

		// if we have a file attached, download it (in memory) and put a pointer to it in msg.Extra
		if message.Raw.File != nil {
			// limit to 1MB for now
			if message.Raw.File.Size <= b.General.MediaDownloadSize {
				comment := ""
				data, err := b.downloadFile(message.Raw.File.URLPrivateDownload)
				if err != nil {
					flog.Errorf("download %s failed %#v", message.Raw.File.URLPrivateDownload, err)
				} else {
					results := regexp.MustCompile(`.*?commented: (.*)`).FindAllStringSubmatch(msg.Text, -1)
					if len(results) > 0 {
						comment = results[0][1]
					}
					msg.Extra["file"] = append(msg.Extra["file"], config.FileInfo{Name: message.Raw.File.Name, Data: data, Comment: comment})
				}
			}
		}
		flog.Debugf("Message is %#v", msg)
		b.Remote <- msg
	}
}

func (b *Bslack) handleSlackClient(mchan chan *MMMessage) {
	for msg := range b.rtm.IncomingEvents {
		switch ev := msg.Data.(type) {
		case *slack.MessageEvent:
			flog.Debugf("Receiving from slackclient %#v", ev)
			if len(ev.Attachments) > 0 {
				// skip messages we made ourselves
				if ev.Attachments[0].CallbackID == "matterbridge" {
					continue
				}
			}
			if !b.Config.EditDisable && ev.SubMessage != nil && ev.SubMessage.ThreadTimestamp != ev.SubMessage.Timestamp {
				flog.Debugf("SubMessage %#v", ev.SubMessage)
				ev.User = ev.SubMessage.User
				ev.Text = ev.SubMessage.Text + b.Config.EditSuffix

				// it seems ev.SubMessage.Edited == nil when slack unfurls
				// do not forward these messages #266
				if ev.SubMessage.Edited == nil {
					continue
				}
			}
			// use our own func because rtm.GetChannelInfo doesn't work for private channels
			channel, err := b.getChannelByID(ev.Channel)
			if err != nil {
				continue
			}
			m := &MMMessage{}
			if ev.BotID == "" && ev.SubType != "message_deleted" {
				user, err := b.rtm.GetUserInfo(ev.User)
				if err != nil {
					continue
				}
				m.UserID = user.ID
				m.Username = user.Name
				if user.Profile.DisplayName != "" {
					m.Username = user.Profile.DisplayName
				}
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
			m.Text = b.replaceVariable(m.Text)
			m.Text = b.replaceChannel(m.Text)
			// when using webhookURL we can't check if it's our webhook or not for now
			if ev.BotID != "" && b.Config.WebhookURL == "" {
				bot, err := b.rtm.GetBotInfo(ev.BotID)
				if err != nil {
					continue
				}
				if bot.Name != "" {
					m.Username = bot.Name
					if ev.Username != "" {
						m.Username = ev.Username
					}
					m.UserID = bot.ID
				}
			}
			mchan <- m
		case *slack.OutgoingErrorEvent:
			flog.Debugf("%#v", ev.Error())
		case *slack.ChannelJoinedEvent:
			b.Users, _ = b.sc.GetUsers()
		case *slack.ConnectedEvent:
			b.channels = ev.Info.Channels
			b.si = ev.Info
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
		m.Text = b.replaceVariable(m.Text)
		m.Text = b.replaceChannel(m.Text)
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
			if u.Profile.DisplayName != "" {
				return u.Profile.DisplayName
			}
			return u.Name
		}
	}
	return ""
}

// @see https://api.slack.com/docs/message-formatting#linking_to_channels_and_users
func (b *Bslack) replaceMention(text string) string {
	results := regexp.MustCompile(`<@([a-zA-z0-9]+)>`).FindAllStringSubmatch(text, -1)
	for _, r := range results {
		text = strings.Replace(text, "<@"+r[1]+">", "@"+b.userName(r[1]), -1)
	}
	return text
}

// @see https://api.slack.com/docs/message-formatting#linking_to_channels_and_users
func (b *Bslack) replaceChannel(text string) string {
	results := regexp.MustCompile(`<#[a-zA-Z0-9]+\|(.+?)>`).FindAllStringSubmatch(text, -1)
	for _, r := range results {
		text = strings.Replace(text, r[0], "#"+r[1], -1)
	}
	return text
}

// @see https://api.slack.com/docs/message-formatting#variables
func (b *Bslack) replaceVariable(text string) string {
	results := regexp.MustCompile(`<!([a-zA-Z0-9]+)(\|.+?)?>`).FindAllStringSubmatch(text, -1)
	for _, r := range results {
		text = strings.Replace(text, r[0], "@"+r[1], -1)
	}
	return text
}

// @see https://api.slack.com/docs/message-formatting#linking_to_urls
func (b *Bslack) replaceURL(text string) string {
	results := regexp.MustCompile(`<(.*?)(\|.*?)?>`).FindAllStringSubmatch(text, -1)
	for _, r := range results {
		text = strings.Replace(text, r[0], r[1], -1)
	}
	return text
}

func (b *Bslack) createAttach(extra map[string][]interface{}) []slack.Attachment {
	var attachs []slack.Attachment
	for _, v := range extra["attachments"] {
		entry := v.(map[string]interface{})
		s := slack.Attachment{}
		s.Fallback = entry["fallback"].(string)
		s.Color = entry["color"].(string)
		s.Pretext = entry["pretext"].(string)
		s.AuthorName = entry["author_name"].(string)
		s.AuthorLink = entry["author_link"].(string)
		s.AuthorIcon = entry["author_icon"].(string)
		s.Title = entry["title"].(string)
		s.TitleLink = entry["title_link"].(string)
		s.Text = entry["text"].(string)
		s.ImageURL = entry["image_url"].(string)
		s.ThumbURL = entry["thumb_url"].(string)
		s.Footer = entry["footer"].(string)
		s.FooterIcon = entry["footer_icon"].(string)
		attachs = append(attachs, s)
	}
	return attachs
}

func (b *Bslack) downloadFile(url string) (*[]byte, error) {
	var buf bytes.Buffer
	client := &http.Client{
		Timeout: time.Second * 5,
	}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", "Bearer "+b.Config.Token)
	resp, err := client.Do(req)
	if err != nil {
		resp.Body.Close()
		return nil, err
	}
	io.Copy(&buf, resp.Body)
	data := buf.Bytes()
	resp.Body.Close()
	return &data, nil
}
