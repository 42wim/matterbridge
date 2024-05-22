//go:build whatsappmulti
// +build whatsappmulti

package bwhatsapp

import (
	"fmt"
	"strings"

	goproto "google.golang.org/protobuf/proto"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/binary/proto"
	"go.mau.fi/whatsmeow/store"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types"
)

type ProfilePicInfo struct {
	URL    string `json:"eurl"`
	Tag    string `json:"tag"`
	Status int16  `json:"status"`
}

func (b *Bwhatsapp) reloadContacts() {
	if _, err := b.wc.Store.Contacts.GetAllContacts(); err != nil {
		b.Log.Errorf("error on update of contacts: %v", err)
	}

	allcontacts, err := b.wc.Store.Contacts.GetAllContacts()
	if err != nil {
		b.Log.Errorf("error on update of contacts: %v", err)
	}

	if len(allcontacts) > 0 {
		b.contacts = allcontacts
	}
}

func (b *Bwhatsapp) getSenderName(info types.MessageInfo) string {
	// Parse AD JID
	var senderJid types.JID
	senderJid.User, senderJid.Server = info.Sender.User, info.Sender.Server

	sender, exists := b.contacts[senderJid]

	if !exists || (sender.FullName == "" && sender.FirstName == "") {
		b.reloadContacts() // Contacts may need to be reloaded
		sender, exists = b.contacts[senderJid]
	}

	if exists && sender.FullName != "" {
		return sender.FullName
	}

	if info.PushName != "" {
		return info.PushName
	}

	if exists && sender.FirstName != "" {
		return sender.FirstName
	}

	return "Someone"
}

func (b *Bwhatsapp) getSenderNameFromJID(senderJid types.JID) string {
	sender, exists := b.contacts[senderJid]

	if !exists || (sender.FullName == "" && sender.FirstName == "") {
		b.reloadContacts() // Contacts may need to be reloaded
		sender, exists = b.contacts[senderJid]
	}

	if exists && sender.FullName != "" {
		return sender.FullName
	}

	if exists && sender.FirstName != "" {
		return sender.FirstName
	}

	if sender.PushName != "" {
		return sender.PushName
	}

	return "Someone"
}

func (b *Bwhatsapp) getSenderNotify(senderJid types.JID) string {
	sender, exists := b.contacts[senderJid]

	if !exists || (sender.FullName == "" && sender.PushName == "" && sender.FirstName == "") {
		b.reloadContacts() // Contacts may need to be reloaded
		sender, exists = b.contacts[senderJid]
	}

	if !exists {
		return "someone"
	}

	if exists && sender.FullName != "" {
		return sender.FullName
	}

	if exists && sender.PushName != "" {
		return sender.PushName
	}

	if exists && sender.FirstName != "" {
		return sender.FirstName
	}

	return "someone"
}

func (b *Bwhatsapp) GetProfilePicThumb(jid string) (*types.ProfilePictureInfo, error) {
	pjid, _ := types.ParseJID(jid)

	info, err := b.wc.GetProfilePictureInfo(pjid, &whatsmeow.GetProfilePictureParams{
		Preview: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get avatar: %v", err)
	}

	return info, nil
}

func isGroupJid(identifier string) bool {
	return strings.HasSuffix(identifier, "@g.us") ||
		strings.HasSuffix(identifier, "@temp") ||
		strings.HasSuffix(identifier, "@broadcast")
}

func (b *Bwhatsapp) getDevice() (*store.Device, error) {
	device := &store.Device{}

	storeContainer, err := sqlstore.New("sqlite", "file:"+b.Config.GetString("sessionfile")+".db?_pragma=foreign_keys(1)&busy_timeout=10000", nil)
	if err != nil {
		return device, fmt.Errorf("failed to connect to database: %v", err)
	}

	device, err = storeContainer.GetFirstDevice()
	if err != nil {
		return device, fmt.Errorf("failed to get device: %v", err)
	}

	return device, nil
}

func (b *Bwhatsapp) getNewReplyContext(parentID string) (*proto.ContextInfo, error) {
	replyInfo, err := b.parseMessageID(parentID)

	if err != nil {
		return nil, err
	}

	sender := fmt.Sprintf("%s@%s", replyInfo.Sender.User, replyInfo.Sender.Server)
	ctx := &proto.ContextInfo{
		StanzaId:      &replyInfo.MessageID,
		Participant:   &sender,
		QuotedMessage: &proto.Message{Conversation: goproto.String("")},
	}

	return ctx, nil
}

func (b *Bwhatsapp) parseMessageID(id string) (*Replyable, error) {
	// No message ID in case action is executed on a message sent before the bridge was started
	// and then the bridge cache doesn't have this message ID mapped
	if id == "" {
		return &Replyable{MessageID: id}, nil
	}

	replyInfo := strings.Split(id, "/")

	if len(replyInfo) == 2 {
		sender, err := types.ParseJID(replyInfo[0])

		if err == nil {
			return &Replyable{
				MessageID: types.MessageID(replyInfo[1]),
				Sender:    sender,
			}, nil
		}
	}

	err := fmt.Errorf("MessageID does not match format of {senderJID}:{messageID} : \"%s\"", id)

	return &Replyable{MessageID: id}, err
}

func getParentIdFromCtx(ci *proto.ContextInfo) string {
	if ci != nil && ci.StanzaId != nil {
		senderJid, err := types.ParseJID(*ci.Participant)

		if err == nil {
			return getMessageIdFormat(senderJid, *ci.StanzaId)
		}
	}

	return ""
}

func getMessageIdFormat(jid types.JID, messageID string) string {
	// we're crafting our own JID str as AD JID format messes with how stuff looks on a webclient
	jidStr := fmt.Sprintf("%s@%s", jid.User, jid.Server)
	return fmt.Sprintf("%s/%s", jidStr, messageID)
}
