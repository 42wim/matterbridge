package bslack

import (
	"bytes"
	"fmt"
	"html"
	"regexp"
	"time"

	"github.com/42wim/matterbridge/bridge/config"
	"github.com/42wim/matterbridge/bridge/helper"
	"github.com/nlopes/slack"
)

func (b *Bslack) handleSlack() {
	messages := make(chan *config.Message)
	if b.GetString("WebhookBindAddress") != "" {
		b.Log.Debugf("Choosing webhooks based receiving")
		go b.handleMatterHook(messages)
	} else {
		b.Log.Debugf("Choosing token based receiving")
		go b.handleSlackClient(messages)
	}
	time.Sleep(time.Second)
	b.Log.Debug("Start listening for Slack messages")
	for message := range messages {
		b.Log.Debugf("<= Sending message from %s on %s to gateway", message.Username, b.Account)

		// cleanup the message
		message.Text = b.replaceMention(message.Text)
		message.Text = b.replaceVariable(message.Text)
		message.Text = b.replaceChannel(message.Text)
		message.Text = b.replaceURL(message.Text)
		message.Text = html.UnescapeString(message.Text)

		// Add the avatar
		message.Avatar = b.getAvatar(message.UserID)

		b.Log.Debugf("<= Message is %#v", message)
		b.Remote <- *message
	}
}

func (b *Bslack) handleSlackClient(messages chan *config.Message) {
	for msg := range b.rtm.IncomingEvents {
		if msg.Type != "user_typing" && msg.Type != "latency_report" {
			b.Log.Debugf("== Receiving event %#v", msg.Data)
		}
		switch ev := msg.Data.(type) {
		case *slack.MessageEvent:
			if b.skipMessageEvent(ev) {
				b.Log.Debugf("Skipped message: %#v", ev)
				continue
			}
			rmsg, err := b.handleMessageEvent(ev)
			if err != nil {
				b.Log.Errorf("%#v", err)
				continue
			}
			messages <- rmsg
		case *slack.OutgoingErrorEvent:
			b.Log.Debugf("%#v", ev.Error())
		case *slack.ChannelJoinedEvent:
			b.Users, _ = b.sc.GetUsers()
			b.Usergroups, _ = b.sc.GetUserGroups()
		case *slack.ConnectedEvent:
			var err error
			b.channels, _, err = b.sc.GetConversations(&slack.GetConversationsParameters{Limit: 1000, Types: []string{"public_channel,private_channel,mpim,im"}})
			if err != nil {
				b.Log.Errorf("Channel list failed: %#v", err)
			}
			b.si = ev.Info
			b.Users, _ = b.sc.GetUsers()
			b.Usergroups, _ = b.sc.GetUserGroups()
		case *slack.InvalidAuthEvent:
			b.Log.Fatalf("Invalid Token %#v", ev)
		case *slack.ConnectionErrorEvent:
			b.Log.Errorf("Connection failed %#v %#v", ev.Error(), ev.ErrorObj)
		default:
		}
	}
}

func (b *Bslack) handleMatterHook(messages chan *config.Message) {
	for {
		message := b.mh.Receive()
		b.Log.Debugf("receiving from matterhook (slack) %#v", message)
		if message.UserName == "slackbot" {
			continue
		}
		messages <- &config.Message{Username: message.UserName, Text: message.Text, Channel: message.ChannelName}
	}
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
	err := helper.HandleDownloadSize(b.Log, rmsg, file.Name, int64(file.Size), b.General)
	if err != nil {
		return err
	}
	// actually download the file
	data, err := helper.DownloadFileAuth(file.URLPrivateDownload, "Bearer "+b.GetString("Token"))
	if err != nil {
		return fmt.Errorf("download %s failed %#v", file.URLPrivateDownload, err)
	}
	// add the downloaded data to the message
	helper.HandleDownloadData(b.Log, rmsg, file.Name, comment, file.URLPrivateDownload, data, b.General)
	return nil
}

// handleUploadFile handles native upload of files
func (b *Bslack) handleUploadFile(msg *config.Message, channelID string) (string, error) {
	for _, f := range msg.Extra["file"] {
		fi := f.(config.FileInfo)
		if msg.Text == fi.Comment {
			msg.Text = ""
		}
		/* because the result of the UploadFile is slower than the MessageEvent from slack
		we can't match on the file ID yet, so we have to match on the filename too
		*/
		b.Log.Debugf("Adding file %s to cache %s", fi.Name, time.Now().String())
		b.cache.Add("filename"+fi.Name, time.Now())
		res, err := b.sc.UploadFile(slack.FileUploadParameters{
			Reader:         bytes.NewReader(*fi.Data),
			Filename:       fi.Name,
			Channels:       []string{channelID},
			InitialComment: fi.Comment,
		})
		if res.ID != "" {
			b.Log.Debugf("Adding fileid %s to cache %s", res.ID, time.Now().String())
			b.cache.Add("file"+res.ID, time.Now())
		}
		if err != nil {
			b.Log.Errorf("uploadfile %#v", err)
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
	if !b.GetBool("EditDisable") && ev.SubMessage != nil && ev.SubMessage.ThreadTimestamp != ev.SubMessage.Timestamp {
		b.Log.Debugf("SubMessage %#v", ev.SubMessage)
		ev.User = ev.SubMessage.User
		ev.Text = ev.SubMessage.Text + b.GetString("EditSuffix")
	}

	// use our own func because rtm.GetChannelInfo doesn't work for private channels
	channel, err := b.getChannelByID(ev.Channel)
	if err != nil {
		return nil, err
	}

	rmsg := config.Message{Text: ev.Text, Channel: channel.Name, Account: b.Account, ID: "slack " + ev.Timestamp, Extra: make(map[string][]interface{})}

	if b.UseChannelID {
		rmsg.Channel = "ID:" + channel.ID
	}

	// find the user id and name
	if ev.User != "" && ev.SubType != messageDeleted && ev.SubType != "file_comment" {
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
				if attach.Title != "" {
					rmsg.Text = attach.Title + "\n"
				}
				rmsg.Text += attach.Text
			} else {
				rmsg.Text = attach.Fallback
			}
		}
	}

	// when using webhookURL we can't check if it's our webhook or not for now
	if rmsg.Username == "" && ev.BotID != "" && b.GetString("WebhookURL") == "" {
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

		// fixes issues with matterircd users
		if bot.Name == "Slack API Tester" {
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
	if (rmsg.Text == "" || rmsg.Username == "") && ev.SubType != messageDeleted && len(ev.Files) == 0 {
		// this is probably a webhook we couldn't resolve
		if ev.BotID != "" {
			return nil, fmt.Errorf("probably an incoming webhook we couldn't resolve (maybe ourselves)")
		}
		return nil, fmt.Errorf("empty message and not a deleted message")
	}

	// save the attachments, so that we can send them to other slack (compatible) bridges
	if len(ev.Attachments) > 0 {
		rmsg.Extra["slack_attachment"] = append(rmsg.Extra["slack_attachment"], ev.Attachments)
	}

	// if we have a file attached, download it (in memory) and put a pointer to it in msg.Extra
	if len(ev.Files) > 0 {
		for _, f := range ev.Files {
			err := b.handleDownloadFile(&rmsg, &f)
			if err != nil {
				b.Log.Errorf("download failed: %s", err)
			}
		}
	}

	return &rmsg, nil
}

// skipMessageEvent skips event that need to be skipped :-)
func (b *Bslack) skipMessageEvent(ev *slack.MessageEvent) bool {
	if ev.SubType == "channel_leave" || ev.SubType == "channel_join" {
		return b.GetBool("nosendjoinpart")
	}

	// ignore pinned items
	if ev.SubType == "pinned_item" || ev.SubType == "unpinned_item" {
		return true
	}

	// do not send messages from ourself
	if b.GetString("WebhookURL") == "" && b.GetString("WebhookBindAddress") == "" && ev.Username == b.si.User.Name {
		return true
	}

	// skip messages we made ourselves
	if len(ev.Attachments) > 0 {
		if ev.Attachments[0].CallbackID == "matterbridge_"+b.uuid {
			return true
		}
	}

	if !b.GetBool("EditDisable") && ev.SubMessage != nil && ev.SubMessage.ThreadTimestamp != ev.SubMessage.Timestamp {
		// it seems ev.SubMessage.Edited == nil when slack unfurls
		// do not forward these messages #266
		if ev.SubMessage.Edited == nil {
			return true
		}
	}

	if len(ev.Files) > 0 {
		for _, f := range ev.Files {
			// if the file is in the cache and isn't older then a minute, skip it
			if ts, ok := b.cache.Get("file" + f.ID); ok && time.Since(ts.(time.Time)) < time.Minute {
				b.Log.Debugf("Not downloading file id %s which we uploaded", f.ID)
				return true
			} else {
				if ts, ok := b.cache.Get("filename" + f.Name); ok && time.Since(ts.(time.Time)) < time.Second*10 {
					b.Log.Debugf("Not downloading file name %s which we uploaded", f.Name)
					return true
				} else {
					b.Log.Debugf("Not skipping %s %s", f.Name, time.Now().String())
				}
			}
		}
	}

	return false
}
