package bmattermost

import (
	"github.com/42wim/matterbridge/bridge/config"
	"github.com/42wim/matterbridge/bridge/helper"
	"github.com/42wim/matterbridge/matterclient"
	matterclient6 "github.com/matterbridge/matterclient"
	"github.com/mattermost/mattermost-server/v5/model"
	model6 "github.com/mattermost/mattermost-server/v6/model"
)

// handleDownloadAvatar downloads the avatar of userid from channel
// sends a EVENT_AVATAR_DOWNLOAD message to the gateway if successful.
// logs an error message if it fails
func (b *Bmattermost) handleDownloadAvatar(userid string, channel string) {
	rmsg := config.Message{
		Username: "system",
		Text:     "avatar",
		Channel:  channel,
		Account:  b.Account,
		UserID:   userid,
		Event:    config.EventAvatarDownload,
		Extra:    make(map[string][]interface{}),
	}
	if _, ok := b.avatarMap[userid]; !ok {
		var (
			data []byte
			err  error
			resp *model.Response
		)
		if b.mc6 != nil {
			data, _, err = b.mc6.Client.GetProfileImage(userid, "")
			if err != nil {
				b.Log.Errorf("ProfileImage download failed for %#v %s", userid, err)
				return
			}
		} else {
			data, resp = b.mc.Client.GetProfileImage(userid, "")
			if resp.Error != nil {
				b.Log.Errorf("ProfileImage download failed for %#v %s", userid, resp.Error)
				return
			}
		}

		err = helper.HandleDownloadSize(b.Log, &rmsg, userid+".png", int64(len(data)), b.General)
		if err != nil {
			b.Log.Error(err)
			return
		}
		helper.HandleDownloadData(b.Log, &rmsg, userid+".png", rmsg.Text, "", &data, b.General)
		b.Remote <- rmsg
	}
}

// handleDownloadFile handles file download
func (b *Bmattermost) handleDownloadFile(rmsg *config.Message, id string) error {
	if b.mc6 != nil {
		return b.handleDownloadFile6(rmsg, id)
	}

	url, _ := b.mc.Client.GetFileLink(id)
	finfo, resp := b.mc.Client.GetFileInfo(id)
	if resp.Error != nil {
		return resp.Error
	}
	err := helper.HandleDownloadSize(b.Log, rmsg, finfo.Name, finfo.Size, b.General)
	if err != nil {
		return err
	}
	data, resp := b.mc.Client.DownloadFile(id, true)
	if resp.Error != nil {
		return resp.Error
	}
	helper.HandleDownloadData(b.Log, rmsg, finfo.Name, rmsg.Text, url, &data, b.General)
	return nil
}

// nolint:wrapcheck
func (b *Bmattermost) handleDownloadFile6(rmsg *config.Message, id string) error {
	url, _, _ := b.mc6.Client.GetFileLink(id)
	finfo, _, err := b.mc6.Client.GetFileInfo(id)
	if err != nil {
		return err
	}
	err = helper.HandleDownloadSize(b.Log, rmsg, finfo.Name, finfo.Size, b.General)
	if err != nil {
		return err
	}
	data, _, err := b.mc6.Client.DownloadFile(id, true)
	if err != nil {
		return err
	}
	helper.HandleDownloadData(b.Log, rmsg, finfo.Name, rmsg.Text, url, &data, b.General)
	return nil
}

func (b *Bmattermost) handleMatter() {
	messages := make(chan *config.Message)
	if b.GetString("WebhookBindAddress") != "" {
		b.Log.Debugf("Choosing webhooks based receiving")
		go b.handleMatterHook(messages)
	} else {
		if b.GetString("Token") != "" {
			b.Log.Debugf("Choosing token based receiving")
		} else {
			b.Log.Debugf("Choosing login/password based receiving")
		}
		// if for some reason we only want to sent stuff to mattermost but not receive, return
		if b.GetString("WebhookBindAddress") == "" && b.GetString("WebhookURL") != "" && b.GetString("Token") == "" && b.GetString("Login") == "" {
			b.Log.Debugf("No WebhookBindAddress specified, only WebhookURL. You will not receive messages from mattermost, only sending is possible.")
		}
		go b.handleMatterClient(messages)
	}
	var ok bool
	for message := range messages {
		message.Avatar = helper.GetAvatar(b.avatarMap, message.UserID, b.General)
		message.Account = b.Account
		message.Text, ok = b.replaceAction(message.Text)
		if ok {
			message.Event = config.EventUserAction
		}
		b.Log.Debugf("<= Sending message from %s on %s to gateway", message.Username, b.Account)
		b.Log.Debugf("<= Message is %#v", message)
		b.Remote <- *message
	}
}

func (b *Bmattermost) handleMatterClient(messages chan *config.Message) {
	if b.mc6 != nil {
		b.Log.Debug("starting matterclient6")
		b.handleMatterClient6(messages)
		return
	}

	for message := range b.mc.MessageChan {
		b.Log.Debugf("%#v", message.Raw.Data)

		if b.skipMessage(message) {
			b.Log.Debugf("Skipped message: %#v", message)
			continue
		}

		channelName := b.getChannelName(message.Post.ChannelId)
		if channelName == "" {
			channelName = message.Channel
		}

		// only download avatars if we have a place to upload them (configured mediaserver)
		if b.General.MediaServerUpload != "" || b.General.MediaDownloadPath != "" {
			b.handleDownloadAvatar(message.UserID, channelName)
		}

		b.Log.Debugf("== Receiving event %#v", message)

		rmsg := &config.Message{
			Username: message.Username,
			UserID:   message.UserID,
			Channel:  channelName,
			Text:     message.Text,
			ID:       message.Post.Id,
			ParentID: message.Post.RootId, // ParentID is obsolete with mattermost
			Extra:    make(map[string][]interface{}),
		}

		// handle mattermost post properties (override username and attachments)
		b.handleProps(rmsg, message)

		// create a text for bridges that don't support native editing
		if message.Raw.Event == model.WEBSOCKET_EVENT_POST_EDITED && !b.GetBool("EditDisable") {
			rmsg.Text = message.Text + b.GetString("EditSuffix")
		}

		if message.Raw.Event == model.WEBSOCKET_EVENT_POST_DELETED {
			rmsg.Event = config.EventMsgDelete
		}

		for _, id := range message.Post.FileIds {
			err := b.handleDownloadFile(rmsg, id)
			if err != nil {
				b.Log.Errorf("download failed: %s", err)
			}
		}

		// Use nickname instead of username if defined
		if !b.GetBool("useusername") {
			if nick := b.mc.GetNickName(rmsg.UserID); nick != "" {
				rmsg.Username = nick
			}
		}

		messages <- rmsg
	}
}

// nolint:cyclop
func (b *Bmattermost) handleMatterClient6(messages chan *config.Message) {
	for message := range b.mc6.MessageChan {
		b.Log.Debugf("%#v %#v", message.Raw.GetData(), message.Raw.EventType())

		if b.skipMessage6(message) {
			b.Log.Debugf("Skipped message: %#v", message)
			continue
		}

		channelName := b.getChannelName(message.Post.ChannelId)
		if channelName == "" {
			channelName = message.Channel
		}

		// only download avatars if we have a place to upload them (configured mediaserver)
		if b.General.MediaServerUpload != "" || b.General.MediaDownloadPath != "" {
			b.handleDownloadAvatar(message.UserID, channelName)
		}

		b.Log.Debugf("== Receiving event %#v", message)

		rmsg := &config.Message{
			Username: message.Username,
			UserID:   message.UserID,
			Channel:  channelName,
			Text:     message.Text,
			ID:       message.Post.Id,
			ParentID: message.Post.RootId, // ParentID is obsolete with mattermost
			Extra:    make(map[string][]interface{}),
		}

		// handle mattermost post properties (override username and attachments)
		b.handleProps6(rmsg, message)

		// create a text for bridges that don't support native editing
		if message.Raw.EventType() == model6.WebsocketEventPostEdited && !b.GetBool("EditDisable") {
			rmsg.Text = message.Text + b.GetString("EditSuffix")
		}

		if message.Raw.EventType() == model6.WebsocketEventPostDeleted {
			rmsg.Event = config.EventMsgDelete
		}

		for _, id := range message.Post.FileIds {
			err := b.handleDownloadFile(rmsg, id)
			if err != nil {
				b.Log.Errorf("download failed: %s", err)
			}
		}

		// Use nickname instead of username if defined
		if !b.GetBool("useusername") {
			if nick := b.mc6.GetNickName(rmsg.UserID); nick != "" {
				rmsg.Username = nick
			}
		}

		messages <- rmsg
	}
}

func (b *Bmattermost) handleMatterHook(messages chan *config.Message) {
	for {
		message := b.mh.Receive()
		b.Log.Debugf("Receiving from matterhook %#v", message)

		messages <- &config.Message{
			UserID:   message.UserID,
			Username: message.UserName,
			Text:     message.Text,
			Channel:  message.ChannelName,
		}
	}
}

// handleUploadFile handles native upload of files
func (b *Bmattermost) handleUploadFile(msg *config.Message) (string, error) {
	if b.mc6 != nil {
		return b.handleUploadFile6(msg)
	}

	var err error
	var res, id string
	channelID := b.getChannelID(msg.Channel)
	for _, f := range msg.Extra["file"] {
		fi := f.(config.FileInfo)
		id, err = b.mc.UploadFile(*fi.Data, channelID, fi.Name)
		if err != nil {
			return "", err
		}
		msg.Text = fi.Comment
		if b.GetBool("PrefixMessagesWithNick") {
			msg.Text = msg.Username + msg.Text
		}
		res, err = b.mc.PostMessageWithFiles(channelID, msg.Text, msg.ParentID, []string{id})
	}
	return res, err
}

// nolint:forcetypeassert,wrapcheck
func (b *Bmattermost) handleUploadFile6(msg *config.Message) (string, error) {
	var err error
	var res, id string
	channelID := b.getChannelID(msg.Channel)
	for _, f := range msg.Extra["file"] {
		fi := f.(config.FileInfo)
		id, err = b.mc6.UploadFile(*fi.Data, channelID, fi.Name)
		if err != nil {
			return "", err
		}
		msg.Text = fi.Comment
		if b.GetBool("PrefixMessagesWithNick") {
			msg.Text = msg.Username + msg.Text
		}
		res, err = b.mc6.PostMessageWithFiles(channelID, msg.Text, msg.ParentID, []string{id})
	}
	return res, err
}

func (b *Bmattermost) handleProps(rmsg *config.Message, message *matterclient.Message) {
	props := message.Post.Props
	if props == nil {
		return
	}
	if _, ok := props["override_username"].(string); ok {
		rmsg.Username = props["override_username"].(string)
	}
	if _, ok := props["attachments"].([]interface{}); ok {
		rmsg.Extra["attachments"] = props["attachments"].([]interface{})
		if rmsg.Text == "" {
			for _, attachment := range rmsg.Extra["attachments"] {
				attach := attachment.(map[string]interface{})
				if attach["text"].(string) != "" {
					rmsg.Text += attach["text"].(string)
					continue
				}
				if attach["fallback"].(string) != "" {
					rmsg.Text += attach["fallback"].(string)
				}
			}
		}
	}
}

// nolint:forcetypeassert
func (b *Bmattermost) handleProps6(rmsg *config.Message, message *matterclient6.Message) {
	props := message.Post.Props
	if props == nil {
		return
	}
	if _, ok := props["override_username"].(string); ok {
		rmsg.Username = props["override_username"].(string)
	}
	if _, ok := props["attachments"].([]interface{}); ok {
		rmsg.Extra["attachments"] = props["attachments"].([]interface{})
		if rmsg.Text != "" {
			return
		}

		for _, attachment := range rmsg.Extra["attachments"] {
			attach := attachment.(map[string]interface{})
			if attach["text"].(string) != "" {
				rmsg.Text += attach["text"].(string)
				continue
			}
			if attach["fallback"].(string) != "" {
				rmsg.Text += attach["fallback"].(string)
			}
		}
	}
}
