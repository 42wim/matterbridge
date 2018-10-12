package bslack

import (
	"errors"
	"strings"
	"sync"

	"github.com/42wim/matterbridge/bridge"
	"github.com/42wim/matterbridge/bridge/config"
	"github.com/42wim/matterbridge/bridge/helper"
	"github.com/42wim/matterbridge/matterhook"
	"github.com/hashicorp/golang-lru"
	"github.com/nlopes/slack"
	"github.com/rs/xid"
)

type Bslack struct {
	mh           *matterhook.Client
	sc           *slack.Client
	rtm          *slack.RTM
	Users        []slack.User
	Usergroups   []slack.UserGroup
	si           *slack.Info
	channels     []slack.Channel
	cache        *lru.Cache
	UseChannelID bool
	uuid         string
	*bridge.Config
	sync.RWMutex
}

const messageDeleted = "message_deleted"

func New(cfg *bridge.Config) bridge.Bridger {
	b := &Bslack{Config: cfg, uuid: xid.New().String()}
	b.cache, _ = lru.New(5000)
	return b
}

func (b *Bslack) Command(cmd string) string {
	return ""
}

func (b *Bslack) Connect() error {
	b.RLock()
	defer b.RUnlock()
	if b.GetString("WebhookBindAddress") != "" {
		if b.GetString("WebhookURL") != "" {
			b.Log.Info("Connecting using webhookurl (sending) and webhookbindaddress (receiving)")
			b.mh = matterhook.New(b.GetString("WebhookURL"),
				matterhook.Config{InsecureSkipVerify: b.GetBool("SkipTLSVerify"),
					BindAddress: b.GetString("WebhookBindAddress")})
		} else if b.GetString("Token") != "" {
			b.Log.Info("Connecting using token (sending)")
			b.sc = slack.New(b.GetString("Token"))
			b.rtm = b.sc.NewRTM()
			go b.rtm.ManageConnection()
			b.Log.Info("Connecting using webhookbindaddress (receiving)")
			b.mh = matterhook.New(b.GetString("WebhookURL"),
				matterhook.Config{InsecureSkipVerify: b.GetBool("SkipTLSVerify"),
					BindAddress: b.GetString("WebhookBindAddress")})
		} else {
			b.Log.Info("Connecting using webhookbindaddress (receiving)")
			b.mh = matterhook.New(b.GetString("WebhookURL"),
				matterhook.Config{InsecureSkipVerify: b.GetBool("SkipTLSVerify"),
					BindAddress: b.GetString("WebhookBindAddress")})
		}
		go b.handleSlack()
		return nil
	}
	if b.GetString("WebhookURL") != "" {
		b.Log.Info("Connecting using webhookurl (sending)")
		b.mh = matterhook.New(b.GetString("WebhookURL"),
			matterhook.Config{InsecureSkipVerify: b.GetBool("SkipTLSVerify"),
				DisableServer: true})
		if b.GetString("Token") != "" {
			b.Log.Info("Connecting using token (receiving)")
			b.sc = slack.New(b.GetString("Token"))
			b.rtm = b.sc.NewRTM()
			go b.rtm.ManageConnection()
			go b.handleSlack()
		}
	} else if b.GetString("Token") != "" {
		b.Log.Info("Connecting using token (sending and receiving)")
		b.sc = slack.New(b.GetString("Token"))
		b.rtm = b.sc.NewRTM()
		go b.rtm.ManageConnection()
		go b.handleSlack()
	}
	if b.GetString("WebhookBindAddress") == "" && b.GetString("WebhookURL") == "" && b.GetString("Token") == "" {
		return errors.New("no connection method found. See that you have WebhookBindAddress, WebhookURL or Token configured")
	}
	return nil
}

func (b *Bslack) Disconnect() error {
	return b.rtm.Disconnect()
}

func (b *Bslack) JoinChannel(channel config.ChannelInfo) error {
	// use ID:channelid and resolve it to the actual name
	idcheck := strings.Split(channel.Name, "ID:")
	if len(idcheck) > 1 {
		b.UseChannelID = true
		ch, err := b.sc.GetChannelInfo(idcheck[1])
		if err != nil {
			return err
		}
		channel.Name = ch.Name
		if err != nil {
			return err
		}
	}

	// we can only join channels using the API
	if b.sc != nil {
		if strings.HasPrefix(b.GetString("Token"), "xoxb") {
			// TODO check if bot has already joined channel
			return nil
		}
		_, err := b.sc.JoinChannel(channel.Name)
		if err != nil {
			switch err.Error() {
			case "name_taken", "restricted_action":
			case "default":
				{
					return err
				}
			}
		}
	}
	return nil
}

func (b *Bslack) Send(msg config.Message) (string, error) {
	b.Log.Debugf("=> Receiving %#v", msg)

	// Make a action /me of the message
	if msg.Event == config.EVENT_USER_ACTION {
		msg.Text = "_" + msg.Text + "_"
	}

	// Use webhook to send the message
	if b.GetString("WebhookURL") != "" {
		return b.sendWebhook(msg)
	}

	channelID := b.getChannelID(msg.Channel)

	// Delete message
	if msg.Event == config.EVENT_MSG_DELETE {
		// some protocols echo deletes, but with empty ID
		if msg.ID == "" {
			return "", nil
		}
		// we get a "slack <ID>", split it
		ts := strings.Fields(msg.ID)
		_, _, err := b.sc.DeleteMessage(channelID, ts[1])
		if err != nil {
			return msg.ID, err
		}
		return msg.ID, nil
	}

	// Prepend nick if configured
	if b.GetBool("PrefixMessagesWithNick") {
		msg.Text = msg.Username + msg.Text
	}

	// Edit message if we have an ID
	if msg.ID != "" {
		ts := strings.Fields(msg.ID)
		_, _, _, err := b.sc.UpdateMessage(channelID, ts[1], msg.Text)
		if err != nil {
			return msg.ID, err
		}
		return msg.ID, nil
	}

	// create slack new post parameters
	np := slack.NewPostMessageParameters()
	if b.GetBool("PrefixMessagesWithNick") {
		np.AsUser = true
	}
	np.Username = msg.Username
	np.LinkNames = 1 // replace mentions
	np.IconURL = config.GetIconURL(&msg, b.GetString("iconurl"))
	if msg.Avatar != "" {
		np.IconURL = msg.Avatar
	}
	// add a callback ID so we can see we created it
	np.Attachments = append(np.Attachments, slack.Attachment{CallbackID: "matterbridge_" + b.uuid})
	// add file attachments
	np.Attachments = append(np.Attachments, b.createAttach(msg.Extra)...)
	// add slack attachments (from another slack bridge)
	if len(msg.Extra["slack_attachment"]) > 0 {
		for _, attach := range msg.Extra["slack_attachment"] {
			np.Attachments = append(np.Attachments, attach.([]slack.Attachment)...)
		}
	}

	// Upload a file if it exists
	if msg.Extra != nil {
		for _, rmsg := range helper.HandleExtra(&msg, b.General) {
			b.sc.PostMessage(channelID, rmsg.Username+rmsg.Text, np)
		}
		// check if we have files to upload (from slack, telegram or mattermost)
		if len(msg.Extra["file"]) > 0 {
			b.handleUploadFile(&msg, channelID)
		}
	}

	// Post normal message
	_, id, err := b.sc.PostMessage(channelID, msg.Text, np)
	if err != nil {
		return "", err
	}
	return "slack " + id, nil
}

func (b *Bslack) Reload(cfg *bridge.Config) (string, error) {
	return "", nil
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

// sendWebhook uses the configured WebhookURL to send the message
func (b *Bslack) sendWebhook(msg config.Message) (string, error) {
	// skip events
	if msg.Event != "" {
		return "", nil
	}

	if b.GetBool("PrefixMessagesWithNick") {
		msg.Text = msg.Username + msg.Text
	}

	if msg.Extra != nil {
		// this sends a message only if we received a config.EVENT_FILE_FAILURE_SIZE
		for _, rmsg := range helper.HandleExtra(&msg, b.General) {
			iconURL := config.GetIconURL(&rmsg, b.GetString("iconurl"))
			matterMessage := matterhook.OMessage{IconURL: iconURL, Channel: msg.Channel, UserName: rmsg.Username, Text: rmsg.Text}
			b.mh.Send(matterMessage)
		}

		// webhook doesn't support file uploads, so we add the url manually
		if len(msg.Extra["file"]) > 0 {
			for _, f := range msg.Extra["file"] {
				fi := f.(config.FileInfo)
				if fi.URL != "" {
					msg.Text += " " + fi.URL
				}
			}
		}
	}

	// if we have native slack_attachments add them
	var attachs []slack.Attachment
	if len(msg.Extra["slack_attachment"]) > 0 {
		for _, attach := range msg.Extra["slack_attachment"] {
			attachs = append(attachs, attach.([]slack.Attachment)...)
		}
	}

	iconURL := config.GetIconURL(&msg, b.GetString("iconurl"))
	matterMessage := matterhook.OMessage{IconURL: iconURL, Attachments: attachs, Channel: msg.Channel, UserName: msg.Username, Text: msg.Text}
	if msg.Avatar != "" {
		matterMessage.IconURL = msg.Avatar
	}
	err := b.mh.Send(matterMessage)
	if err != nil {
		b.Log.Error(err)
		return "", err
	}
	return "", nil
}
