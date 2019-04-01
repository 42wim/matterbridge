package brocketchat

import (
	"fmt"
	"github.com/42wim/matterbridge/bridge/config"
	"github.com/42wim/matterbridge/bridge/helper"
	"github.com/matterbridge/Rocket.Chat.Go.SDK/models"
)

func (b *Brocketchat) handleRocket() {
	messages := make(chan *config.Message)
	if b.GetString("WebhookBindAddress") != "" {
		b.Log.Debugf("Choosing webhooks based receiving")
		go b.handleRocketHook(messages)
	} else {
		b.Log.Debugf("Choosing login/password based receiving")
		go b.handleRocketClient(messages)
	}
	for message := range messages {
		message.Account = b.Account
		b.Log.Debugf("<= Sending message from %s on %s to gateway", message.Username, b.Account)
		b.Log.Debugf("<= Message is %#v", message)
		b.Remote <- *message
	}
}

func (b *Brocketchat) handleRocketHook(messages chan *config.Message) {
	for {
		message := b.rh.Receive()
		b.Log.Debugf("Receiving from rockethook %#v", message)
		// do not loop
		if message.UserName == b.GetString("Nick") {
			continue
		}
		messages <- &config.Message{
			UserID:   message.UserID,
			Username: message.UserName,
			Text:     message.Text,
			Channel:  message.ChannelName,
		}
	}
}

func (b *Brocketchat) handleRocketClient(messages chan *config.Message) {
	for message := range b.messageChan {
		// skip messages with same ID, apparently messages get duplicated for an unknown reason
		if _, ok := b.cache.Get(message.ID); ok {
			continue
		}
		b.cache.Add(message.ID, true)
		b.Log.Debugf("message %#v", message)
		m := message
		if b.skipMessage(&m) {
			b.Log.Debugf("Skipped message: %#v", message)
			continue
		}

		extra := make(map[string][]interface{})

		rmsg := &config.Message{Text: message.Msg,
			Username: message.User.UserName,
			Channel:  b.getChannelName(message.RoomID),
			Account:  b.Account,
			UserID:   message.User.ID,
			ID:       message.ID,
			Extra:    extra,
		}

		b.handleAttachments(&message, rmsg)

		messages <- rmsg
	}
}

func (b *Brocketchat) handleAttachments(message *models.Message, rmsg *config.Message) {

	// See if we have some text in the attachments.
	if rmsg.Text == "" {
		for _, attach := range message.Attachments {
			if attach.Text != "" {
				if attach.Title != "" {
					rmsg.Text = attach.Title + "\n"
				}
				rmsg.Text += attach.Text
			}
		}
	}

	// Save the attachments, so that we can send them to other slack (compatible) bridges.
	if len(message.Attachments) > 0 {
		rmsg.Extra["rocketchat_attachments"] = append(rmsg.Extra["rocketchat_attachments"], message.Attachments)
	}

	// If we have files attached, download them (in memory) and put a pointer to it in msg.Extra.
	for i := range message.Attachments {
		if err := b.handleDownloadFile(rmsg, &message.Attachments[i], false); err != nil {
			b.Log.Errorf("Could not download incoming file: %#v", err)
		}
	}
}

func (b *Brocketchat) handleDownloadFile(rmsg *config.Message, file *models.Attachment, retry bool) error {
	// TODO: Check that the file is neither too large nor blacklisted.
	/*	if err := helper.HandleDownloadSize(b.Log, rmsg, file.Title, int64(file.Size), b.General); err != nil {
			b.Log.WithError(err).Infof("Skipping download of incoming file.")
			return nil
		}*/
	if !file.TitleLinkDownload {
		b.Log.Infof("File is not intended to be downloaded.")
		return nil
	}

	downloadUrl := b.GetString("server") + file.TitleLink
	// Actually download the file.
	data, err := helper.DownloadFileAuthRocket(downloadUrl, b.user.Token, b.user.ID)
	if err != nil {
		return fmt.Errorf("download %s failed %#v", downloadUrl, err)
	}

	// If a comment is attached to the file(s) it is in the 'Text' field of the Slack messge event
	// and should be added as comment to only one of the files. We reset the 'Text' field to ensure
	// that the comment is not duplicated.
	comment := rmsg.Text
	rmsg.Text = ""
	helper.HandleDownloadData(b.Log, rmsg, file.Title, comment, downloadUrl, data, b.General)
	return nil
}

func (b *Brocketchat) handleUploadFile(msg *config.Message) error {
	for _, f := range msg.Extra["file"] {
		fi := f.(config.FileInfo)
		if err := b.uploadFile(&fi, b.getChannelID(msg.Channel)); err != nil {
			return err
		}
	}
	return nil
}
