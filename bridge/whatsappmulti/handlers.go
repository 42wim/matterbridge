// +build whatsappmulti

package bwhatsapp

import (
	"fmt"
	"mime"
	"strings"

	"github.com/42wim/matterbridge/bridge/config"
	"github.com/42wim/matterbridge/bridge/helper"

	"go.mau.fi/whatsmeow/binary/proto"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
)

// nolint:gocritic
func (b *Bwhatsapp) eventHandler(evt interface{}) {

	switch e := evt.(type) {
	case *events.Message:
		b.handleMessage(e)
	}
}

func (b *Bwhatsapp) handleMessage(message *events.Message) {
	msg := message.Message

	switch {
	case msg == nil, message.Info.IsFromMe, message.Info.Timestamp.Before(b.startedAt):
		return
	}

	b.Log.Infof("Receiving message %#v", msg)

	mEvent :=""

	if msg.GetProtocolMessage() != nil {
		pMsg := msg.GetProtocolMessage()
		b.Log.Debugf("ProtocolMessage is %v", pMsg.Type)

		if pMsg.GetType() == proto.ProtocolMessage_REVOKE {
		b.Log.Debug("Succesful trigger")
		mEvent = "msg_delete"
		}



		} //Delete info
  //b.Log.Debugf("Event is %#v", msg.GetProtocolMessage().String()) //Delete info

  //igIndex := 0
  //b.Log.Debugf("Ig Vid is %#v", b.IgVid[igIndex]) //filter test


	switch {
	case mEvent == "msg_delete":
		b.handleDelete(message)
	case (msg.Conversation != nil || msg.ExtendedTextMessage != nil):
		b.handleTextMessage(message.Info, msg)
	case msg.VideoMessage != nil:
		b.handleVideoMessage(message)
	case msg.AudioMessage != nil:
		b.handleAudioMessage(message)
	case msg.DocumentMessage != nil:
		b.handleDocumentMessage(message)
	case msg.ImageMessage != nil:
		b.handleImageMessage(message)
	}
}

// nolint:funlen
func (b *Bwhatsapp) handleTextMessage(messageInfo types.MessageInfo, msg *proto.Message) {
	senderJID := messageInfo.Sender
	channel := messageInfo.Chat
	mPushName := messageInfo.PushName

	senderName := b.getSenderName(messageInfo.Sender)
	if senderName == "" {
		senderName = "Someone" // don't expose telephone number
	}

	if msg.GetExtendedTextMessage() == nil && msg.GetConversation() == "" {
		b.Log.Debugf("message without text content? %#v", msg)
		return
	}


	var text string
	ci := msg.GetExtendedTextMessage().GetContextInfo()
	b.Log.Debugf("<= ContextInfo is %+v", ci)
	// nolint:nestif
	if msg.GetExtendedTextMessage() == nil {
		text = msg.GetConversation()
	} else {
		text = msg.GetExtendedTextMessage().GetText()
		//if ci != nil {ci := msg.GetExtendedTextMessage().GetContextInfo()}

		if senderJID == (types.JID{}) && ci.Participant != nil {
			senderJID = types.NewJID(ci.GetParticipant(), types.DefaultUserServer)
		}

		if ci != nil && ci.MentionedJid != nil {
			// handle user mentions
			for _, mentionedJID := range ci.MentionedJid {
				numberAndSuffix := strings.SplitN(mentionedJID, "@", 2)

				// mentions comes as telephone numbers and we don't want to expose it to other bridges
				// replace it with something more meaninful to others
				mention := b.getSenderNotify(types.NewJID(numberAndSuffix[0], types.DefaultUserServer))
				if mention == "" {
					mention = "someone"
				}

				text = strings.Replace(text, "@"+numberAndSuffix[0], "@"+mention, 1)
			}
		}
	}


if ci != nil{
	if len(ci.QuotedMessage.GetConversation()) >0 {
		//handleQuote , only for text msgs now
		text = strings.Join([]string{text, "(re" , ci.GetParticipant() ,ci.QuotedMessage.GetConversation(),")"}, " ")
	}
}


	rmsg := config.Message{
		UserID:   senderJID.String(),
		Username: mPushName, //changed to pushname
		Text:     text,
		Channel:  channel.String(),
		Account:  b.Account,
		Protocol: b.Protocol,
		Extra:    make(map[string][]interface{}),
		//      ParentID: TODO, // TODO handle thread replies  // map from Info.QuotedMessageID string
		ID: messageInfo.ID,
	}

	if avatarURL, exists := b.userAvatars[senderJID.String()]; exists {
		rmsg.Avatar = avatarURL
	}

	b.Log.Debugf("<= Sending message from %s on %s to gateway", senderJID, b.Account)
	b.Log.Debugf("<= Message is %#v", rmsg)

	b.Log.Debugf("<= PushName is %s", mPushName)

	b.Remote <- rmsg
}

// HandleImageMessage sent from WhatsApp, relay it to the brige
func (b *Bwhatsapp) handleImageMessage(msg *events.Message) {
	imsg := msg.Message.GetImageMessage()

	senderJID := msg.Info.Sender
	senderName := b.getSenderName(senderJID)
	ci := imsg.GetContextInfo()

	if senderJID == (types.JID{}) && ci.Participant != nil {
		senderJID = types.NewJID(ci.GetParticipant(), types.DefaultUserServer)
	}

	rmsg := config.Message{
		UserID:   senderJID.String(),
		Username: senderName,
		Channel:  msg.Info.Chat.String(),
		Account:  b.Account,
		Protocol: b.Protocol,
		Extra:    make(map[string][]interface{}),
		ID:       msg.Info.ID,
	}

	if avatarURL, exists := b.userAvatars[senderJID.String()]; exists {
		rmsg.Avatar = avatarURL
	}

	fileExt, err := mime.ExtensionsByType(imsg.GetMimetype())
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

	filename := fmt.Sprintf("%v%v", msg.Info.ID, fileExt[0])

	b.Log.Debugf("Trying to download %s with type %s", filename, imsg.GetMimetype())

	data, err := b.wc.Download(imsg)
	if err != nil {
		b.Log.Errorf("Download image failed: %s", err)

		return
	}

	// Move file to bridge storage
	helper.HandleDownloadData(b.Log, &rmsg, filename, imsg.GetCaption(), "", &data, b.General)

	b.Log.Debugf("<= Sending message from %s on %s to gateway", senderJID, b.Account)
	b.Log.Debugf("<= Message is %#v", rmsg)

	b.Remote <- rmsg
}

// HandleVideoMessage downloads video messages
func (b *Bwhatsapp) handleVideoMessage(msg *events.Message) {
	imsg := msg.Message.GetVideoMessage()

	senderJID := msg.Info.Sender
	senderName := b.getSenderName(senderJID)
	ci := imsg.GetContextInfo()

	if senderJID == (types.JID{}) && ci.Participant != nil {
		senderJID = types.NewJID(ci.GetParticipant(), types.DefaultUserServer)
	}

	rmsg := config.Message{
		UserID:   senderJID.String(),
		Username: senderName,
		Channel:  msg.Info.Chat.String(),
		Account:  b.Account,
		Protocol: b.Protocol,
		Extra:    make(map[string][]interface{}),
		ID:       msg.Info.ID,
	}

	if avatarURL, exists := b.userAvatars[senderJID.String()]; exists {
		rmsg.Avatar = avatarURL
	}

	fileExt, err := mime.ExtensionsByType(imsg.GetMimetype())
	if err != nil {
		b.Log.Errorf("Mimetype detection error: %s", err)

		return
	}

	if len(fileExt) == 0 {
		fileExt = append(fileExt, ".mp4")
	}

	filename := fmt.Sprintf("%v%v", msg.Info.ID, fileExt[0])

	b.Log.Debugf("Trying to download %s with size %#v and type %s", filename, imsg.GetFileLength(), imsg.GetMimetype())

	data, err := b.wc.Download(imsg)
	if err != nil {
		b.Log.Errorf("Download video failed: %s", err)

		return
	}

	// Move file to bridge storage
	helper.HandleDownloadData(b.Log, &rmsg, filename, imsg.GetCaption(), "", &data, b.General)

	b.Log.Debugf("<= Sending message from %s on %s to gateway", senderJID, b.Account)
	b.Log.Debugf("<= Message is %#v", rmsg)

	b.Remote <- rmsg
}

// HandleAudioMessage downloads audio messages
func (b *Bwhatsapp) handleAudioMessage(msg *events.Message) {
	imsg := msg.Message.GetAudioMessage()

	senderJID := msg.Info.Sender
	senderName := b.getSenderName(senderJID)
	ci := imsg.GetContextInfo()

	if senderJID == (types.JID{}) && ci.Participant != nil {
		senderJID = types.NewJID(ci.GetParticipant(), types.DefaultUserServer)
	}

	rmsg := config.Message{
		UserID:   senderJID.String(),
		Username: senderName,
		Channel:  msg.Info.Chat.String(),
		Account:  b.Account,
		Protocol: b.Protocol,
		Extra:    make(map[string][]interface{}),
		ID:       msg.Info.ID,
	}

	if avatarURL, exists := b.userAvatars[senderJID.String()]; exists {
		rmsg.Avatar = avatarURL
	}

	fileExt, err := mime.ExtensionsByType(imsg.GetMimetype())
	if err != nil {
		b.Log.Errorf("Mimetype detection error: %s", err)

		return
	}

	if len(fileExt) == 0 {
		fileExt = append(fileExt, ".ogg")
	}

	filename := fmt.Sprintf("%v%v", msg.Info.ID, fileExt[0])

	b.Log.Debugf("Trying to download %s with size %#v and type %s", filename, imsg.GetFileLength(), imsg.GetMimetype())

	data, err := b.wc.Download(imsg)
	if err != nil {
		b.Log.Errorf("Download video failed: %s", err)

		return
	}

	// Move file to bridge storage
	helper.HandleDownloadData(b.Log, &rmsg, filename, "audio message", "", &data, b.General)

	b.Log.Debugf("<= Sending message from %s on %s to gateway", senderJID, b.Account)
	b.Log.Debugf("<= Message is %#v", rmsg)

	b.Remote <- rmsg
}

// HandleDocumentMessage downloads documents
func (b *Bwhatsapp) handleDocumentMessage(msg *events.Message) {
	imsg := msg.Message.GetDocumentMessage()

	senderJID := msg.Info.Sender
	senderName := b.getSenderName(senderJID)
	ci := imsg.GetContextInfo()

	if senderJID == (types.JID{}) && ci.Participant != nil {
		senderJID = types.NewJID(ci.GetParticipant(), types.DefaultUserServer)
	}

	rmsg := config.Message{
		UserID:   senderJID.String(),
		Username: senderName,
		Channel:  msg.Info.Chat.String(),
		Account:  b.Account,
		Protocol: b.Protocol,
		Extra:    make(map[string][]interface{}),
		ID:       msg.Info.ID,
	}

	if avatarURL, exists := b.userAvatars[senderJID.String()]; exists {
		rmsg.Avatar = avatarURL
	}

	fileExt, err := mime.ExtensionsByType(imsg.GetMimetype())
	if err != nil {
		b.Log.Errorf("Mimetype detection error: %s", err)

		return
	}

	filename := fmt.Sprintf("%v", imsg.GetFileName())

	b.Log.Debugf("Trying to download %s with extension %s and type %s", filename, fileExt, imsg.GetMimetype())

	data, err := b.wc.Download(imsg)
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


// Handle Delete
func (b *Bwhatsapp) handleDelete(msg *events.Message) {
	//imsg := msg.Message.GetDocumentMessage()

	senderJID := msg.Info.Sender
	senderName := b.getSenderName(senderJID)
	//ci := imsg.GetContextInfo()

	//if senderJID == (types.JID{}) && ci.Participant != nil {
	//	senderJID = types.NewJID(ci.GetParticipant(), types.DefaultUserServer)
//	}

	rmsg := config.Message{
		UserID:   senderJID.String(),
		Username: senderName,
		Channel:  msg.Info.Chat.String(),
		Account:  b.Account,
		Protocol: b.Protocol,
		//Extra:    make(map[string][]interface{}),
		ID:       msg.Info.ID,
		Event:		"msg_delete",
	}

	if avatarURL, exists := b.userAvatars[senderJID.String()]; exists {
		rmsg.Avatar = avatarURL
	}

	//fileExt, err := mime.ExtensionsByType(imsg.GetMimetype())
//	if err != nil {
	//	b.Log.Errorf("Mimetype detection error: %s", err)

//		return
//	}

//	filename := fmt.Sprintf("%v", imsg.GetFileName())

//	b.Log.Debugf("Trying to download %s with extension %s and type %s", filename, fileExt, imsg.GetMimetype())

//	data, err := b.wc.Download(imsg)
//	if err != nil {
//		b.Log.Errorf("Download document message failed: %s", err)

//		return
//	}

	// Move file to bridge storage
//	helper.HandleDownloadData(b.Log, &rmsg, filename, "document", "", &data, b.General)

	b.Log.Debugf("<= Sending delete message from %s on %s to gateway", senderJID, b.Account)
	b.Log.Debugf("<= Message is %#v", rmsg)

	b.Remote <- rmsg
}
