package bwhatsapp

import (
	"encoding/gob"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	qrcodeTerminal "github.com/Baozisoftware/qrcode-terminal-go"
	"github.com/Rhymen/go-whatsapp"
)

type ProfilePicInfo struct {
	URL    string `json:"eurl"`
	Tag    string `json:"tag"`
	Status int16  `json:"status"`
}

func qrFromTerminal(invert bool) chan string {
	qr := make(chan string)

	go func() {
		terminal := qrcodeTerminal.New()

		if invert {
			terminal = qrcodeTerminal.New2(qrcodeTerminal.ConsoleColors.BrightWhite, qrcodeTerminal.ConsoleColors.BrightBlack, qrcodeTerminal.QRCodeRecoveryLevels.Medium)
		}

		terminal.Get(<-qr).Print()
	}()

	return qr
}

func (b *Bwhatsapp) readSession() (whatsapp.Session, error) {
	session := whatsapp.Session{}
	sessionFile := b.Config.GetString(sessionFile)

	if sessionFile == "" {
		return session, errors.New("if you won't set SessionFile then you will need to scan QR code on every restart")
	}

	file, err := os.Open(sessionFile)
	if err != nil {
		return session, err
	}

	defer file.Close()

	decoder := gob.NewDecoder(file)

	return session, decoder.Decode(&session)
}

func (b *Bwhatsapp) writeSession(session whatsapp.Session) error {
	sessionFile := b.Config.GetString(sessionFile)

	if sessionFile == "" {
		// we already sent a warning while starting the bridge, so let's be quiet here
		return nil
	}

	file, err := os.Create(sessionFile)
	if err != nil {
		return err
	}

	defer file.Close()

	encoder := gob.NewEncoder(file)

	return encoder.Encode(session)
}

func (b *Bwhatsapp) restoreSession() (*whatsapp.Session, error) {
	session, err := b.readSession()
	if err != nil {
		b.Log.Warn(err.Error())
	}

	b.Log.Debugln("Restoring WhatsApp session..")

	session, err = b.conn.RestoreWithSession(session)
	if err != nil {
		// restore session connection timed out (I couldn't get over it without logging in again)
		return nil, errors.New("failed to restore session: " + err.Error())
	}

	b.Log.Debugln("Session restored successfully!")

	return &session, nil
}

func (b *Bwhatsapp) getSenderName(senderJid string) string {
	if sender, exists := b.users[senderJid]; exists {
		if sender.Name != "" {
			return sender.Name
		}
		// if user is not in phone contacts
		// it is the most obvious scenario unless you sync your phone contacts with some remote updated source
		// users can change it in their WhatsApp settings -> profile -> click on Avatar
		if sender.Notify != "" {
			return sender.Notify
		}

		if sender.Short != "" {
			return sender.Short
		}
	}

	// try to reload this contact
	if _, err := b.conn.Contacts(); err != nil {
		b.Log.Errorf("error on update of contacts: %v", err)
	}

	if contact, exists := b.conn.Store.Contacts[senderJid]; exists {
		// Add it to the user map
		b.users[senderJid] = contact

		if contact.Name != "" {
			return contact.Name
		}
		// if user is not in phone contacts
		// same as above
		return contact.Notify
	}

	return ""
}

func (b *Bwhatsapp) getSenderNotify(senderJid string) string {
	if sender, exists := b.users[senderJid]; exists {
		return sender.Notify
	}

	return ""
}

func (b *Bwhatsapp) GetProfilePicThumb(jid string) (*ProfilePicInfo, error) {
	data, err := b.conn.GetProfilePicThumb(jid)
	if err != nil {
		return nil, fmt.Errorf("failed to get avatar: %v", err)
	}

	content := <-data
	info := &ProfilePicInfo{}

	err = json.Unmarshal([]byte(content), info)
	if err != nil {
		return info, fmt.Errorf("failed to unmarshal avatar info: %v", err)
	}

	return info, nil
}

func isGroupJid(identifier string) bool {
	return strings.HasSuffix(identifier, "@g.us") ||
		strings.HasSuffix(identifier, "@temp") ||
		strings.HasSuffix(identifier, "@broadcast")
}
