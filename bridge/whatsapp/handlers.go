package bwhatsapp

import (
	"strings"
	"time"

	"github.com/42wim/matterbridge/bridge/config"
	"github.com/Rhymen/go-whatsapp"
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
	if strings.Contains(err.Error(), "error processing data: received invalid data") {
		return
	}
	b.Log.Errorf("%v", err) // TODO implement proper handling? at least respond to different error types
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
	groupJid := message.Info.RemoteJid

	senderJid := message.Info.SenderJid
	if len(senderJid) == 0 {
		// TODO workaround till https://github.com/Rhymen/go-whatsapp/issues/86 resolved
		senderJid = *message.Info.Source.Participant
	}

	// translate sender's Jid to the nicest username we can get
	senderName := b.getSenderName(senderJid)
	if senderName == "" {
		senderName = "Someone" // don't expose telephone number
	}

	extText := message.Info.Source.Message.ExtendedTextMessage
	if extText != nil && extText.ContextInfo != nil && extText.ContextInfo.MentionedJid != nil {
		// handle user mentions
		for _, mentionedJid := range extText.ContextInfo.MentionedJid {
			numberAndSuffix := strings.SplitN(mentionedJid, "@", 2)

			// mentions comes as telephone numbers and we don't want to expose it to other bridges
			// replace it with something more meaninful to others
			mention := b.getSenderNotify(numberAndSuffix[0] + "@s.whatsapp.net")
			if mention == "" {
				mention = "someone"
			}
			message.Text = strings.Replace(message.Text, "@"+numberAndSuffix[0], "@"+mention, 1)
		}
	}

	b.Log.Debugf("<= Sending message from %s on %s to gateway", senderJid, b.Account)
	rmsg := config.Message{
		UserID:    senderJid,
		Username:  senderName,
		Text:      message.Text,
		Timestamp: messageTime,
		Channel:   groupJid,
		Account:   b.Account,
		Protocol:  b.Protocol,
		Extra:     make(map[string][]interface{}),
		//		ParentID: TODO, // TODO handle thread replies  // map from Info.QuotedMessageID string
		//	Event     string    `json:"event"`
		//	Gateway   string  // will be added during message processing
		ID: message.Info.Id}

	if avatarURL, exists := b.userAvatars[senderJid]; exists {
		rmsg.Avatar = avatarURL
	}

	b.Log.Debugf("<= Message is %#v", rmsg)
	b.Remote <- rmsg
}

//
//func (b *Bwhatsapp) HandleImageMessage(message whatsapp.ImageMessage) {
//	fmt.Println(message) // TODO implement
//}
//
//func (b *Bwhatsapp) HandleVideoMessage(message whatsapp.VideoMessage) {
//	fmt.Println(message) // TODO implement
//}
//
//func (b *Bwhatsapp) HandleJsonMessage(message string) {
//	fmt.Println(message) // TODO implement
//}
// TODO HandleRawMessage
// TODO HandleAudioMessage
