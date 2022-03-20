// +build whatsappmulti

package bwhatsapp

import (
	"fmt"
	"strings"

	"go.mau.fi/whatsmeow/store"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types"
)

type ProfilePicInfo struct {
	URL    string `json:"eurl"`
	Tag    string `json:"tag"`
	Status int16  `json:"status"`
}

func (b *Bwhatsapp) getSenderName(senderJid types.JID) string {
	if sender, exists := b.contacts[senderJid]; exists {
		if sender.FullName != "" {
			return sender.FullName
		}
		// if user is not in phone contacts
		// it is the most obvious scenario unless you sync your phone contacts with some remote updated source
		// users can change it in their WhatsApp settings -> profile -> click on Avatar
		if sender.PushName != "" {
			return sender.PushName
		}

		if sender.FirstName != "" {
			return sender.FirstName
		}
	}

	// try to reload this contact
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

	if sender, exists := b.contacts[senderJid]; exists {
		if sender.FullName != "" {
			return sender.FullName
		}
		// if user is not in phone contacts
		// it is the most obvious scenario unless you sync your phone contacts with some remote updated source
		// users can change it in their WhatsApp settings -> profile -> click on Avatar
		if sender.PushName != "" {
			return sender.PushName
		}

		if sender.FirstName != "" {
			return sender.FirstName
		}
	}

	return "Someone"
}

func (b *Bwhatsapp) getSenderNotify(senderJid types.JID) string {
	if sender, exists := b.contacts[senderJid]; exists {
		return sender.PushName
	}

	return ""
}

func (b *Bwhatsapp) GetProfilePicThumb(jid string) (*types.ProfilePictureInfo, error) {
	pjid, _ := types.ParseJID(jid)
	info, err := b.wc.GetProfilePictureInfo(pjid, true)
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

	storeContainer, err := sqlstore.New("sqlite", "file:"+b.Config.GetString("sessionfile")+".db?_foreign_keys=on&_pragma=busy_timeout=10000", nil)
	if err != nil {
		return device, fmt.Errorf("failed to connect to database: %v", err)
	}

	device, err = storeContainer.GetFirstDevice()
	if err != nil {
		return device, fmt.Errorf("failed to get device: %v", err)
	}

	return device, nil
}
