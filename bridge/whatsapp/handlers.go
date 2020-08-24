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
	if strings.Contains(err.Error(), "error processing data: received invalid data") || strings.Contains(err.Error(), "invalid string with tag 174") {
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
	if message.Info.FromMe { // || !strings.Contains(strings.ToLower(message.Text), "@echo") {
		return
	}
	// whatsapp sends last messages to show context , cut them
	if message.Info.Timestamp < b.startedAt {
		return
	}

	messageTime := time.Unix(int64(message.Info.Timestamp), 0) // TODO check how behaves between timezones
	groupJID := message.Info.RemoteJid

	senderJID := message.Info.SenderJid
	if len(senderJID) == 0 {
		// TODO workaround till https://github.com/Rhymen/go-whatsapp/issues/86 resolved
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

	b.Log.Debugf("<= Sending message from %s on %s to gateway", senderJID, b.Account)
	rmsg := config.Message{
		UserID:    senderJID,
		Username:  senderName,
		Text:      message.Text,
		Timestamp: messageTime,
		Channel:   groupJID,
		Account:   b.Account,
		Protocol:  b.Protocol,
		Extra:     make(map[string][]interface{}),
		//	ParentID: TODO, // TODO handle thread replies  // map from Info.QuotedMessageID string
		//	Event     string    `json:"event"`
		//	Gateway   string  // will be added during message processing
		ID: message.Info.Id}

	if avatarURL, exists := b.userAvatars[senderJID]; exists {
		rmsg.Avatar = avatarURL
	}

	b.Log.Debugf("<= Message is %#v", rmsg)
	b.Remote <- rmsg
}

// HandleImageMessage sent from WhatsApp, relay it to the brige
func (b *Bwhatsapp) HandleImageMessage(message whatsapp.ImageMessage) {
	if message.Info.FromMe { // || !strings.Contains(strings.ToLower(message.Text), "@echo") {
		return
	}

	// whatsapp sends last messages to show context , cut them
	if message.Info.Timestamp < b.startedAt {
		return
	}

	messageTime := time.Unix(int64(message.Info.Timestamp), 0) // TODO check how behaves between timezones
	groupJID := message.Info.RemoteJid

	senderJID := message.Info.SenderJid
	// if len(senderJid) == 0 {
	//   // TODO workaround till https://github.com/Rhymen/go-whatsapp/issues/86 resolved
	//   senderJid = *message.Info.Source.Participant
	// }

	// translate sender's Jid to the nicest username we can get
	senderName := b.getSenderName(senderJID)
	if senderName == "" {
		senderName = "Someone" // don't expose telephone number
	}

	b.Log.Debugf("<= Sending message from %s on %s to gateway", senderJID, b.Account)
	rmsg := config.Message{
		UserID:    senderJID,
		Username:  senderName,
		Timestamp: messageTime,
		Channel:   groupJID,
		Account:   b.Account,
		Protocol:  b.Protocol,
		Extra:     make(map[string][]interface{}),
		//  ParentID: TODO,      // TODO handle thread replies  // map from Info.QuotedMessageID string
		//  Event     string    `json:"event"`
		//  Gateway   string     // will be added during message processing
		ID: message.Info.Id}

	if avatarURL, exists := b.userAvatars[senderJID]; exists {
		rmsg.Avatar = avatarURL
	}

	// Download and unencrypt content
	data, err := message.Download()
	if err != nil {
		b.Log.Errorf("%v", err)
		return
	}

	// Get file extension by mimetype
	fileExt, err := mime.ExtensionsByType(message.Type)
	if err != nil {
		b.Log.Errorf("%v", err)
		return
	}

	filename := fmt.Sprintf("%v%v", message.Info.Id, fileExt[0])

	b.Log.Debugf("<= Image downloaded and unencrypted")

	// Move file to bridge storage
	helper.HandleDownloadData(b.Log, &rmsg, filename, message.Caption, "", &data, b.General)

	b.Log.Debugf("<= Image Message is %#v", rmsg)
	b.Remote <- rmsg
}

//func (b *Bwhatsapp) HandleVideoMessage(message whatsapp.VideoMessage) {
//	fmt.Println(message) // TODO implement
//}
//
//func (b *Bwhatsapp) HandleJsonMessage(message string) {
//	fmt.Println(message) // TODO implement
//}
// TODO HandleRawMessage
// TODO HandleAudioMessage
