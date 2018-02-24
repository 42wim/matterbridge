package bslack

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/42wim/matterbridge/bridge/config"
	"github.com/42wim/matterbridge/bridge/helper"
	"github.com/42wim/matterbridge/matterhook"
	"github.com/nlopes/slack"
	log "github.com/sirupsen/logrus"
	"html"
	"regexp"
	"strings"
	"time"
)

type Bslack struct {
	mh       *matterhook.Client
	sc       *slack.Client
	rtm      *slack.RTM
	Users    []slack.User
	si       *slack.Info
	channels []slack.Channel
	*config.BridgeConfig
}

var flog *log.Entry
var protocol = "slack"

const messageDeleted = "messsage_deleted"

func init() {
	flog = log.WithFields(log.Fields{"prefix": protocol})
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
		return errors.New("no connection method found. See that you have WebhookBindAddress, WebhookURL or Token configured")
	}
	return nil
}

func (b *Bslack) Disconnect() error {
	return nil

}

func (b *Bslack) JoinChannel(channel config.ChannelInfo) error {
	// we can only join channels using the API
	if b.sc != nil {
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

	// Make a action /me of the message
	if msg.Event == config.EVENT_USER_ACTION {
		msg.Text = "_" + msg.Text + "_"
	}

	// Use webhook to send the message
	if b.Config.WebhookURL != "" {
		return b.sendWebhook(msg)
	}

	// get the slack channel
	schannel, err := b.getChannelByName(msg.Channel)
	if err != nil {
		return "", err
	}

	// Delete message
	if msg.Event == config.EVENT_MSG_DELETE {
		// some protocols echo deletes, but with empty ID
		if msg.ID == "" {
			return "", nil
		}
		// we get a "slack <ID>", split it
		ts := strings.Fields(msg.ID)
		_, _, err := b.sc.DeleteMessage(schannel.ID, ts[1])
		if err != nil {
			return msg.ID, err
		}
		return msg.ID, nil
	}

	// Prepend nick if configured
	if b.Config.PrefixMessagesWithNick {
		msg.Text = msg.Username + msg.Text
	}

	// Edit message if we have an ID
	if msg.ID != "" {
		ts := strings.Fields(msg.ID)
		_, _, _, err := b.sc.UpdateMessage(schannel.ID, ts[1], msg.Text)
		if err != nil {
			return msg.ID, err
		}
		return msg.ID, nil
	}

	// create slack new post parameters
	np := slack.NewPostMessageParameters()
	if b.Config.PrefixMessagesWithNick {
		np.AsUser = true
	}
	np.Username = msg.Username
	np.LinkNames = 1 // replace mentions
	np.IconURL = config.GetIconURL(&msg, &b.Config)
	if msg.Avatar != "" {
		np.IconURL = msg.Avatar
	}
	// add a callback ID so we can see we created it
	np.Attachments = append(np.Attachments, slack.Attachment{CallbackID: "matterbridge"})
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
			b.sc.PostMessage(schannel.ID, rmsg.Username+rmsg.Text, np)
		}
		// check if we have files to upload (from slack, telegram or mattermost)
		if len(msg.Extra["file"]) > 0 {
			b.handleUploadFile(&msg, schannel.ID)
		}
	}

	// Post normal message
	_, id, err := b.sc.PostMessage(schannel.ID, msg.Text, np)
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
	messages := make(chan *config.Message)
	if b.Config.WebhookBindAddress != "" {
		flog.Debugf("Choosing webhooks based receiving")
		go b.handleMatterHook(messages)
	} else {
		flog.Debugf("Choosing token based receiving")
		go b.handleSlackClient(messages)
	}
	time.Sleep(time.Second)
	flog.Debug("Start listening for Slack messages")
	for message := range messages {
		flog.Debugf("Sending message from %s on %s to gateway", message.Username, b.Account)

		// cleanup the message
		message.Text = b.replaceURL(message.Text)
		message.Text = html.UnescapeString(message.Text)
		message.Text = b.replaceMention(message.Text)
		message.Text = b.replaceVariable(message.Text)
		message.Text = b.replaceChannel(message.Text)

		// Add the avatar
		message.Avatar = b.getAvatar(message.Username)

		flog.Debugf("Message is %#v", message)
		b.Remote <- *message
	}
}

func (b *Bslack) handleSlackClient(messages chan *config.Message) {
	for msg := range b.rtm.IncomingEvents {
		if msg.Type != "user_typing" && msg.Type != "latency_report" {
			flog.Debugf("Receiving from slackclient %#v", msg.Data)
		}
		switch ev := msg.Data.(type) {
		case *slack.MessageEvent:
			if b.skipMessageEvent(ev) {
				flog.Debugf("Skipped message: %#v", ev)
				continue
			}
			rmsg, err := b.handleMessageEvent(ev)
			if err != nil {
				flog.Errorf("%#v", err)
				continue
			}
			messages <- rmsg
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
		case *slack.ConnectionErrorEvent:
			flog.Errorf("Connection failed %#v %#v", ev.Error(), ev.ErrorObj)
		default:
		}
	}
}

func (b *Bslack) handleMatterHook(messages chan *config.Message) {
	for {
		message := b.mh.Receive()
		flog.Debugf("receiving from matterhook (slack) %#v", message)
		if message.UserName == "slackbot" {
			continue
		}
		messages <- &config.Message{Username: message.UserName, Text: message.Text, Channel: message.ChannelName}
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

// handleDownloadFile handles file download
func (b *Bslack) handleDownloadFile(rmsg *config.Message, file *slack.File) error {
	// if we have a file attached, download it (in memory) and put a pointer to it in msg.Extra
	// limit to 1MB for now
	comment := ""
	results := regexp.MustCompile(`.*?commented: (.*)`).FindAllStringSubmatch(rmsg.Text, -1)
	if len(results) > 0 {
		comment = results[0][1]
	}

	err := helper.HandleDownloadSize(flog, rmsg, file.Name, int64(file.Size), b.General)
	if err != nil {
		return err
	}
	// actually download the file
	data, err := helper.DownloadFileAuth(file.URLPrivateDownload, "Bearer "+b.Config.Token)
	if err != nil {
		return fmt.Errorf("download %s failed %#v", file.URLPrivateDownload, err)
	}
	// add the downloaded data to the message
	helper.HandleDownloadData(flog, rmsg, file.Name, comment, file.URLPrivateDownload, data, b.General)
	return nil
}

// handleUploadFile handles native upload of files
func (b *Bslack) handleUploadFile(msg *config.Message, channelID string) (string, error) {
	var err error
	for _, f := range msg.Extra["file"] {
		fi := f.(config.FileInfo)
		_, err = b.sc.UploadFile(slack.FileUploadParameters{
			Reader:         bytes.NewReader(*fi.Data),
			Filename:       fi.Name,
			Channels:       []string{channelID},
			InitialComment: fi.Comment,
		})
		if err != nil {
			flog.Errorf("uploadfile %#v", err)
		}
	}
	return "", nil
}

// handleMessageEvent handles the message events
func (b *Bslack) handleMessageEvent(ev *slack.MessageEvent) (*config.Message, error) {
	// update the userlist on a channel_join
	if ev.SubType == "channel_join" {
		b.Users, _ = b.sc.GetUsers()
	}

	// Edit message
	if !b.Config.EditDisable && ev.SubMessage != nil && ev.SubMessage.ThreadTimestamp != ev.SubMessage.Timestamp {
		flog.Debugf("SubMessage %#v", ev.SubMessage)
		ev.User = ev.SubMessage.User
		ev.Text = ev.SubMessage.Text + b.Config.EditSuffix
	}

	// use our own func because rtm.GetChannelInfo doesn't work for private channels
	channel, err := b.getChannelByID(ev.Channel)
	if err != nil {
		return nil, err
	}

	rmsg := config.Message{Text: ev.Text, Channel: channel.Name, Account: b.Account, ID: "slack " + ev.Timestamp, Extra: make(map[string][]interface{})}

	// find the user id and name
	if ev.BotID == "" && ev.SubType != messageDeleted && ev.SubType != "file_comment" {
		user, err := b.rtm.GetUserInfo(ev.User)
		if err != nil {
			return nil, err
		}
		rmsg.UserID = user.ID
		rmsg.Username = user.Name
		if user.Profile.DisplayName != "" {
			rmsg.Username = user.Profile.DisplayName
		}
	}

	// See if we have some text in the attachments
	if rmsg.Text == "" {
		for _, attach := range ev.Attachments {
			if attach.Text != "" {
				rmsg.Text = attach.Text
			} else {
				rmsg.Text = attach.Fallback
			}
		}
	}

	// when using webhookURL we can't check if it's our webhook or not for now
	if ev.BotID != "" && b.Config.WebhookURL == "" {
		bot, err := b.rtm.GetBotInfo(ev.BotID)
		if err != nil {
			return nil, err
		}
		if bot.Name != "" {
			rmsg.Username = bot.Name
			if ev.Username != "" {
				rmsg.Username = ev.Username
			}
			rmsg.UserID = bot.ID
		}
	}

	// file comments are set by the system (because there is no username given)
	if ev.SubType == "file_comment" {
		rmsg.Username = "system"
	}

	// do we have a /me action
	if ev.SubType == "me_message" {
		rmsg.Event = config.EVENT_USER_ACTION
	}

	// Handle join/leave
	if ev.SubType == "channel_leave" || ev.SubType == "channel_join" {
		rmsg.Username = "system"
		rmsg.Event = config.EVENT_JOIN_LEAVE
	}

	// edited messages have a submessage, use this timestamp
	if ev.SubMessage != nil {
		rmsg.ID = "slack " + ev.SubMessage.Timestamp
	}

	// deleted message event
	if ev.SubType == messageDeleted {
		rmsg.Text = config.EVENT_MSG_DELETE
		rmsg.Event = config.EVENT_MSG_DELETE
		rmsg.ID = "slack " + ev.DeletedTimestamp
	}

	// topic change event
	if ev.SubType == "channel_topic" || ev.SubType == "channel_purpose" {
		rmsg.Event = config.EVENT_TOPIC_CHANGE
	}

	// Only deleted messages can have a empty username and text
	if (rmsg.Text == "" || rmsg.Username == "") && ev.SubType != messageDeleted {
		return nil, fmt.Errorf("empty message and not a deleted message")
	}

	// save the attachments, so that we can send them to other slack (compatible) bridges
	if len(ev.Attachments) > 0 {
		rmsg.Extra["slack_attachment"] = append(rmsg.Extra["slack_attachment"], ev.Attachments)
	}

	// if we have a file attached, download it (in memory) and put a pointer to it in msg.Extra
	if ev.File != nil {
		err := b.handleDownloadFile(&rmsg, ev.File)
		if err != nil {
			flog.Errorf("download failed: %s", err)
		}
	}

	return &rmsg, nil
}

// sendWebhook uses the configured WebhookURL to send the message
func (b *Bslack) sendWebhook(msg config.Message) (string, error) {
	// skip events
	if msg.Event != "" {
		return "", nil
	}

	if b.Config.PrefixMessagesWithNick {
		msg.Text = msg.Username + msg.Text
	}

	if msg.Extra != nil {
		// this sends a message only if we received a config.EVENT_FILE_FAILURE_SIZE
		for _, rmsg := range helper.HandleExtra(&msg, b.General) {
			matterMessage := matterhook.OMessage{IconURL: b.Config.IconURL, Channel: msg.Channel, UserName: rmsg.Username, Text: rmsg.Text}
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

	matterMessage := matterhook.OMessage{IconURL: b.Config.IconURL, Attachments: attachs, Channel: msg.Channel, UserName: msg.Username, Text: msg.Text}
	if msg.Avatar != "" {
		matterMessage.IconURL = msg.Avatar
	}
	err := b.mh.Send(matterMessage)
	if err != nil {
		flog.Error(err)
		return "", err
	}
	return "", nil
}

// skipMessageEvent skips event that need to be skipped :-)
func (b *Bslack) skipMessageEvent(ev *slack.MessageEvent) bool {
	// ignore pinned items
	if ev.SubType == "pinned_item" || ev.SubType == "unpinned_item" {
		return true
	}

	// do not send messages from ourself
	if b.Config.WebhookURL == "" && b.Config.WebhookBindAddress == "" && ev.Username == b.si.User.Name {
		return true
	}

	// skip messages we made ourselves
	if len(ev.Attachments) > 0 {
		if ev.Attachments[0].CallbackID == "matterbridge" {
			return true
		}
	}

	if !b.Config.EditDisable && ev.SubMessage != nil && ev.SubMessage.ThreadTimestamp != ev.SubMessage.Timestamp {
		// it seems ev.SubMessage.Edited == nil when slack unfurls
		// do not forward these messages #266
		if ev.SubMessage.Edited == nil {
			return true
		}
	}
	return false
}
