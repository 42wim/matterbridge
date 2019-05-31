package bwhatsapp

import (
	"encoding/gob"
	"encoding/json"
	"errors"
	"fmt"
	"os"

	qrcodeTerminal "github.com/Baozisoftware/qrcode-terminal-go"
	"github.com/Rhymen/go-whatsapp"
)

type ProfilePicInfo struct {
	URL string `json:"eurl"`
	Tag string `json:"tag"`

	Status int16 `json:"status"`
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
	err = decoder.Decode(&session)
	if err != nil {
		return session, err
	}
	return session, nil
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
	err = encoder.Encode(session)

	return err
}

func (b *Bwhatsapp) getSenderName(senderJid string) string {
	if sender, exists := b.users[senderJid]; exists {
		if sender.Name != "" {
			return sender.Name
		}
		// if user is not in phone contacts
		// it is the most obvious scenario unless you sync your phone contacts with some remote updated source
		// users can change it in their WhatsApp settings -> profile -> click on Avatar
		return sender.Notify
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
