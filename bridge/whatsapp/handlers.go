// nolint:goconst
package bwhatsapp

import (
	"fmt"
	"mime"
	"strings"
	"time"

	"github.com/42wim/matterbridge/bridge/config"
	"github.com/42wim/matterbridge/bridge/helper"
	"github.com/Rhymen/go-whatsapp"
	"github.com/jpillora/backoff"
)

/*
Implement handling messages coming from WhatsApp
Check:
- https://github.com/Rhymen/go-whatsapp#add-message-handlers
- https://github.com/Rhymen/go-whatsapp/blob/master/handler.go
- https://github.com/tulir/mautrix-whatsapp/tree/master/whatsapp-ext for more advanced command handling
*/

// HandleError received from WhatsApp
func (b *Bwhatsapp) HandleError(err error) {
	// ignore received invalid data errors. https://github.com/42wim/matterbridge/issues/843
	// ignore tag 174 errors. https://github.com/42wim/matterbridge/issues/1094
	if strings.Contains(err.Error(), "error processing data: received invalid data") ||
		strings.Contains(err.Error(), "invalid string with tag 174") {
		return
	}

	switch err.(type) {
	case *whatsapp.ErrConnectionClosed, *whatsapp.ErrConnectionFailed:
		b.reconnect(err)
	default:
		switch err {
		case whatsapp.ErrConnectionTimeout:
			b.reconnect(err)
		default:
			b.Log.Errorf("%v", err)
		}
	}
}

func (b *Bwhatsapp) reconnect(err error) {
	bf := &backoff.Backoff{
		Min:    time.Second,
		Max:    5 * time.Minute,
		Jitter: true,
	}

	for {
		d := bf.Duration()

		b.Log.Errorf("Connection failed, underlying error: %v", err)
		b.Log.Infof("Waiting %s...", d)

		time.Sleep(d)

		b.Log.Info("Reconnecting...")

		err := b.conn.Restore()
		if err == nil {
			bf.Reset()
			b.startedAt = uint64(time.Now().Unix())

			return
		}
	}
}

// HandleTextMessage sent from WhatsApp, relay it to the brige
func (b *Bwhatsapp) HandleTextMessage(message whatsapp.TextMessage) {
	if message.Info.FromMe {
		return
	}
	// whatsapp sends last messages to show context , cut them
	if message.Info.Timestamp < b.startedAt {
		return
	}

	groupJID := message.Info.RemoteJid
	senderJID := message.Info.SenderJid

	if len(senderJID) == 0 {
		if message.Info.Source != nil && message.Info.Source.Participant != nil {
			senderJID = *message.Info.Source.Participant
		}
	}

	// translate sender's JID to the nicest username we can get
	senderName := b.getSenderName(senderJID)
	if senderName == "" {
		senderName = "Someone" // don't expose telephone number
	}

	extText := message.Info.Source.Message.ExtendedTextMessage
	if extText != nil && extText.ContextInfo != nil && extText.ContextInfo.MentionedJid != nil {
		// handle user mentions
		for _, mentionedJID := range extText.ContextInfo.MentionedJid {
			numberAndSuffix := strings.SplitN(mentionedJID, "@", 2)

			// mentions comes as telephone numbers and we don't want to expose it to other bridges
			// replace it with something more meaninful to others
			mention := b.getSenderNotify(numberAndSuffix[0] + "@s.whatsapp.net")
			if mention == "" {
				mention = "someone"
			}

			message.Text = strings.Replace(message.Text, "@"+numberAndSuffix[0], "@"+mention, 1)
		}
	}

	rmsg := config.Message{
		UserID:   senderJID,
		Username: senderName,
		Text:     message.Text,
		Channel:  groupJID,
		Account:  b.Account,
		Protocol: b.Protocol,
		Extra:    make(map[string][]interface{}),
		//	ParentID: TODO, // TODO handle thread replies  // map from Info.QuotedMessageID string
		ID: message.Info.Id,
	}

	if avatarURL, exists := b.userAvatars[senderJID]; exists {
		rmsg.Avatar = avatarURL
	}

	b.Log.Debugf("<= Sending message from %s on %s to gateway", senderJID, b.Account)
	b.Log.Debugf("<= Message is %#v", rmsg)

	b.Remote <- rmsg
}

// HandleImageMessage sent from WhatsApp, relay it to the brige
// nolint:funlen
func (b *Bwhatsapp) HandleImageMessage(message whatsapp.ImageMessage) {
	if message.Info.FromMe || message.Info.Timestamp < b.startedAt {
		return
	}

	senderJID := message.Info.SenderJid
	if len(message.Info.SenderJid) == 0 && message.Info.Source != nil && message.Info.Source.Participant != nil {
		senderJID = *message.Info.Source.Participant
	}

	senderName := b.getSenderName(message.Info.SenderJid)
	if senderName == "" {
		senderName = "Someone" // don't expose telephone number
	}

	rmsg := config.Message{
		UserID:   senderJID,
		Username: senderName,
		Channel:  message.Info.RemoteJid,
		Account:  b.Account,
		Protocol: b.Protocol,
		Extra:    make(map[string][]interface{}),
		ID:       message.Info.Id,
	}

	if avatarURL, exists := b.userAvatars[senderJID]; exists {
		rmsg.Avatar = avatarURL
	}

	fileExt, err := mime.ExtensionsByType(message.Type)
	if err != nil {
		b.Log.Errorf("Mimetype detection error: %s", err)

		return
	}

	// rename .jfif to .jpg https://github.com/42wim/matterbridge/issues/1292
	if fileExt[0] == ".jfif" {
		fileExt[0] = ".jpg"
	}

	// rename .jpe to .jpg https://github.com/42wim/matterbridge/issues/1463
	if fileExt[0] == ".jpe" {
		fileExt[0] = ".jpg"
	}

	filename := fmt.Sprintf("%v%v", message.Info.Id, fileExt[0])

	b.Log.Debugf("Trying to download %s with type %s", filename, message.Type)

	data, err := message.Download()
	if err != nil {
		b.Log.Errorf("Download image failed: %s", err)

		return
	}

	// Move file to bridge storage
	helper.HandleDownloadData(b.Log, &rmsg, filename, message.Caption, "", &data, b.General)

	b.Log.Debugf("<= Sending message from %s on %s to gateway", senderJID, b.Account)
	b.Log.Debugf("<= Message is %#v", rmsg)

	b.Remote <- rmsg
}

// HandleVideoMessage downloads video messages
func (b *Bwhatsapp) HandleVideoMessage(message whatsapp.VideoMessage) {
	if message.Info.FromMe || message.Info.Timestamp < b.startedAt {
		return
	}

	senderJID := message.Info.SenderJid
	if len(message.Info.SenderJid) == 0 && message.Info.Source != nil && message.Info.Source.Participant != nil {
		senderJID = *message.Info.Source.Participant
	}

	senderName := b.getSenderName(message.Info.SenderJid)
	if senderName == "" {
		senderName = "Someone" // don't expose telephone number
	}

	rmsg := config.Message{
		UserID:   senderJID,
		Username: senderName,
		Channel:  message.Info.RemoteJid,
		Account:  b.Account,
		Protocol: b.Protocol,
		Extra:    make(map[string][]interface{}),
		ID:       message.Info.Id,
	}

	if avatarURL, exists := b.userAvatars[senderJID]; exists {
		rmsg.Avatar = avatarURL
	}

	fileExt, err := mime.ExtensionsByType(message.Type)
	if err != nil {
		b.Log.Errorf("Mimetype detection error: %s", err)

		return
	}

	if len(fileExt) == 0 {
		fileExt = append(fileExt, ".mp4")
	}

	filename := fmt.Sprintf("%v%v", message.Info.Id, fileExt[0])

	b.Log.Debugf("Trying to download %s with size %#v and type %s", filename, message.Length, message.Type)

	data, err := message.Download()
	if err != nil {
		b.Log.Errorf("Download video failed: %s", err)

		return
	}

	// Move file to bridge storage
	helper.HandleDownloadData(b.Log, &rmsg, filename, message.Caption, "", &data, b.General)

	b.Log.Debugf("<= Sending message from %s on %s to gateway", senderJID, b.Account)
	b.Log.Debugf("<= Message is %#v", rmsg)

	b.Remote <- rmsg
}

// HandleAudioMessage downloads audio messages
func (b *Bwhatsapp) HandleAudioMessage(message whatsapp.AudioMessage) {
	if message.Info.FromMe || message.Info.Timestamp < b.startedAt {
		return
	}

	senderJID := message.Info.SenderJid
	if len(message.Info.SenderJid) == 0 && message.Info.Source != nil && message.Info.Source.Participant != nil {
		senderJID = *message.Info.Source.Participant
	}

	senderName := b.getSenderName(message.Info.SenderJid)
	if senderName == "" {
		senderName = "Someone" // don't expose telephone number
	}

	rmsg := config.Message{
		UserID:   senderJID,
		Username: senderName,
		Channel:  message.Info.RemoteJid,
		Account:  b.Account,
		Protocol: b.Protocol,
		Extra:    make(map[string][]interface{}),
		ID:       message.Info.Id,
	}

	if avatarURL, exists := b.userAvatars[senderJID]; exists {
		rmsg.Avatar = avatarURL
	}

	fileExt, err := mime.ExtensionsByType(message.Type)
	if err != nil {
		b.Log.Errorf("Mimetype detection error: %s", err)

		return
	}

	if len(fileExt) == 0 {
		fileExt = append(fileExt, ".ogg")
	}

	filename := fmt.Sprintf("%v%v", message.Info.Id, fileExt[0])

	b.Log.Debugf("Trying to download %s with size %#v and type %s", filename, message.Length, message.Type)

	data, err := message.Download()
	if err != nil {
		b.Log.Errorf("Download audio failed: %s", err)

		return
	}

	// Move file to bridge storage
	helper.HandleDownloadData(b.Log, &rmsg, filename, "audio message", "", &data, b.General)

	b.Log.Debugf("<= Sending message from %s on %s to gateway", senderJID, b.Account)
	b.Log.Debugf("<= Message is %#v", rmsg)

	b.Remote <- rmsg
}

// HandleDocumentMessage downloads documents
func (b *Bwhatsapp) HandleDocumentMessage(message whatsapp.DocumentMessage) {
	if message.Info.FromMe || message.Info.Timestamp < b.startedAt {
		return
	}

	senderJID := message.Info.SenderJid
	if len(message.Info.SenderJid) == 0 && message.Info.Source != nil && message.Info.Source.Participant != nil {
		senderJID = *message.Info.Source.Participant
	}

	senderName := b.getSenderName(message.Info.SenderJid)
	if senderName == "" {
		senderName = "Someone" // don't expose telephone number
	}

	rmsg := config.Message{
		UserID:   senderJID,
		Username: senderName,
		Channel:  message.Info.RemoteJid,
		Account:  b.Account,
		Protocol: b.Protocol,
		Extra:    make(map[string][]interface{}),
		ID:       message.Info.Id,
	}

	if avatarURL, exists := b.userAvatars[senderJID]; exists {
		rmsg.Avatar = avatarURL
	}

	fileExt, err := mime.ExtensionsByType(message.Type)
	if err != nil {
		b.Log.Errorf("Mimetype detection error: %s", err)

		return
	}

	filename := fmt.Sprintf("%v", message.FileName)

	b.Log.Debugf("Trying to download %s with extension %s and type %s", filename, fileExt, message.Type)

	data, err := message.Download()
	if err != nil {
		b.Log.Errorf("Download document message failed: %s", err)

		return
	}

	// Move file to bridge storage
	helper.HandleDownloadData(b.Log, &rmsg, filename, "document", "", &data, b.General)

	b.Log.Debugf("<= Sending message from %s on %s to gateway", senderJID, b.Account)
	b.Log.Debugf("<= Message is %#v", rmsg)

	b.Remote <- rmsg
}
