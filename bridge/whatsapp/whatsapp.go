package bwhatsapp

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"mime"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/42wim/matterbridge/bridge"
	"github.com/42wim/matterbridge/bridge/config"
	"github.com/Rhymen/go-whatsapp"
)

const (
	// Account config parameters
	cfgNumber         = "Number"
	qrOnWhiteTerminal = "QrOnWhiteTerminal"
	sessionFile       = "SessionFile"
)

// Bwhatsapp Bridge structure keeping all the information needed for relying
type Bwhatsapp struct {
	*bridge.Config

	session   *whatsapp.Session
	conn      *whatsapp.Conn
	startedAt uint64

	users       map[string]whatsapp.Contact
	userAvatars map[string]string
}

// New Create a new WhatsApp bridge. This will be called for each [whatsapp.<server>] entry you have in the config file
func New(cfg *bridge.Config) bridge.Bridger {
	number := cfg.GetString(cfgNumber)

	cfg.Log.Warn("!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!")
	cfg.Log.Warn("This bridge is deprecated and not supported anymore. Use the new multidevice whatsapp bridge")
	cfg.Log.Warn("See https://github.com/42wim/matterbridge#building-with-whatsapp-beta-multidevice-support for more info")
	cfg.Log.Warn("!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!")

	if number == "" {
		cfg.Log.Fatalf("Missing configuration for WhatsApp bridge: Number")
	}

	b := &Bwhatsapp{
		Config: cfg,

		users:       make(map[string]whatsapp.Contact),
		userAvatars: make(map[string]string),
	}

	return b
}

// Connect to WhatsApp. Required implementation of the Bridger interface
func (b *Bwhatsapp) Connect() error {
	number := b.GetString(cfgNumber)
	if number == "" {
		return errors.New("whatsapp's telephone number need to be configured")
	}

	b.Log.Debugln("Connecting to WhatsApp..")
	conn, err := whatsapp.NewConn(20 * time.Second)
	if err != nil {
		return errors.New("failed to connect to WhatsApp: " + err.Error())
	}

	b.conn = conn

	b.conn.AddHandler(b)
	b.Log.Debugln("WhatsApp connection successful")

	// load existing session in order to keep it between restarts
	b.session, err = b.restoreSession()
	if err != nil {
		b.Log.Warn(err.Error())
	}

	// login to a new session
	if b.session == nil {
		if err = b.Login(); err != nil {
			return err
		}
	}

	b.startedAt = uint64(time.Now().Unix())

	_, err = b.conn.Contacts()
	if err != nil {
		return fmt.Errorf("error on update of contacts: %v", err)
	}

	// see https://github.com/Rhymen/go-whatsapp/issues/137#issuecomment-480316013
	for len(b.conn.Store.Contacts) == 0 {
		b.conn.Contacts() // nolint:errcheck

		<-time.After(1 * time.Second)
	}

	// map all the users
	for id, contact := range b.conn.Store.Contacts {
		if !isGroupJid(id) && id != "status@broadcast" {
			// it is user
			b.users[id] = contact
		}
	}

	// get user avatar asynchronously
	go func() {
		b.Log.Debug("Getting user avatars..")

		for jid := range b.users {
			info, err := b.GetProfilePicThumb(jid)
			if err != nil {
				b.Log.Warnf("Could not get profile photo of %s: %v", jid, err)
			} else {
				b.Lock()
				b.userAvatars[jid] = info.URL
				b.Unlock()
			}
		}

		b.Log.Debug("Finished getting avatars..")
	}()

	return nil
}

// Login to WhatsApp creating a new session. This will require to scan a QR code on your mobile device
func (b *Bwhatsapp) Login() error {
	b.Log.Debugln("Logging in..")

	invert := b.GetBool(qrOnWhiteTerminal) // false is the default
	qrChan := qrFromTerminal(invert)

	session, err := b.conn.Login(qrChan)
	if err != nil {
		b.Log.Warnln("Failed to log in:", err)

		return err
	}

	b.session = &session

	b.Log.Infof("Logged into session: %#v", session)
	b.Log.Infof("Connection: %#v", b.conn)

	err = b.writeSession(session)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error saving session: %v\n", err)
	}

	return nil
}

// Disconnect is called while reconnecting to the bridge
// Required implementation of the Bridger interface
func (b *Bwhatsapp) Disconnect() error {
	// We could Logout, but that would close the session completely and would require a new QR code scan
	// https://github.com/Rhymen/go-whatsapp/blob/c31092027237441cffba1b9cb148eadf7c83c3d2/session.go#L377-L381
	return nil
}

// JoinChannel Join a WhatsApp group specified in gateway config as channel='number-id@g.us' or channel='Channel name'
// Required implementation of the Bridger interface
// https://github.com/42wim/matterbridge/blob/2cfd880cdb0df29771bf8f31df8d990ab897889d/bridge/bridge.go#L11-L16
func (b *Bwhatsapp) JoinChannel(channel config.ChannelInfo) error {
	byJid := isGroupJid(channel.Name)

	// see https://github.com/Rhymen/go-whatsapp/issues/137#issuecomment-480316013
	for len(b.conn.Store.Contacts) == 0 {
		b.conn.Contacts() // nolint:errcheck
		<-time.After(1 * time.Second)
	}

	// verify if we are member of the given group
	if byJid {
		// channel.Name specifies static group jID, not the name
		if _, exists := b.conn.Store.Contacts[channel.Name]; !exists {
			return fmt.Errorf("account doesn't belong to group with jid %s", channel.Name)
		}

		return nil
	}

	// channel.Name specifies group name that might change, warn about it
	var jids []string
	for id, contact := range b.conn.Store.Contacts {
		if isGroupJid(id) && contact.Name == channel.Name {
			jids = append(jids, id)
		}
	}

	switch len(jids) {
	case 0:
		// didn't match any group - print out possibilites
		for id, contact := range b.conn.Store.Contacts {
			if isGroupJid(id) {
				b.Log.Infof("%s %s", contact.Jid, contact.Name)
			}
		}

		return fmt.Errorf("please specify group's JID from the list above instead of the name '%s'", channel.Name)
	case 1:
		return fmt.Errorf("group name might change. Please configure gateway with channel=\"%v\" instead of channel=\"%v\"", jids[0], channel.Name)
	default:
		return fmt.Errorf("there is more than one group with name '%s'. Please specify one of JIDs as channel name: %v", channel.Name, jids)
	}
}

// Post a document message from the bridge to WhatsApp
func (b *Bwhatsapp) PostDocumentMessage(msg config.Message, filetype string) (string, error) {
	fi := msg.Extra["file"][0].(config.FileInfo)

	// Post document message
	message := whatsapp.DocumentMessage{
		Info: whatsapp.MessageInfo{
			RemoteJid: msg.Channel,
		},
		Title:    fi.Name,
		FileName: fi.Name,
		Type:     filetype,
		Content:  bytes.NewReader(*fi.Data),
	}

	b.Log.Debugf("=> Sending %#v", msg)

	// create message ID
	// TODO follow and act if https://github.com/Rhymen/go-whatsapp/issues/101 implemented
	idBytes := make([]byte, 10)
	if _, err := rand.Read(idBytes); err != nil {
		b.Log.Warn(err.Error())
	}

	message.Info.Id = strings.ToUpper(hex.EncodeToString(idBytes))
	_, err := b.conn.Send(message)

	return message.Info.Id, err
}

// Post an image message from the bridge to WhatsApp
// Handle, for sure image/jpeg, image/png and image/gif MIME types
func (b *Bwhatsapp) PostImageMessage(msg config.Message, filetype string) (string, error) {
	fi := msg.Extra["file"][0].(config.FileInfo)

	// Post image message
	message := whatsapp.ImageMessage{
		Info: whatsapp.MessageInfo{
			RemoteJid: msg.Channel,
		},
		Type:    filetype,
		Caption: msg.Username + fi.Comment,
		Content: bytes.NewReader(*fi.Data),
	}

	b.Log.Debugf("=> Sending %#v", msg)

	// create message ID
	// TODO follow and act if https://github.com/Rhymen/go-whatsapp/issues/101 implemented
	idBytes := make([]byte, 10)
	if _, err := rand.Read(idBytes); err != nil {
		b.Log.Warn(err.Error())
	}

	message.Info.Id = strings.ToUpper(hex.EncodeToString(idBytes))
	_, err := b.conn.Send(message)

	return message.Info.Id, err
}

// Send a message from the bridge to WhatsApp
// Required implementation of the Bridger interface
// https://github.com/42wim/matterbridge/blob/2cfd880cdb0df29771bf8f31df8d990ab897889d/bridge/bridge.go#L11-L16
func (b *Bwhatsapp) Send(msg config.Message) (string, error) {
	b.Log.Debugf("=> Receiving %#v", msg)

	// Delete message
	if msg.Event == config.EventMsgDelete {
		if msg.ID == "" {
			// No message ID in case action is executed on a message sent before the bridge was started
			// and then the bridge cache doesn't have this message ID mapped
			return "", nil
		}

		_, err := b.conn.RevokeMessage(msg.Channel, msg.ID, true)

		return "", err
	}

	// Edit message
	if msg.ID != "" {
		b.Log.Debugf("updating message with id %s", msg.ID)

		if b.GetString("editsuffix") != "" {
			msg.Text += b.GetString("EditSuffix")
		} else {
			msg.Text += " (edited)"
		}
	}

	// Handle Upload a file
	if msg.Extra["file"] != nil {
		fi := msg.Extra["file"][0].(config.FileInfo)
		filetype := mime.TypeByExtension(filepath.Ext(fi.Name))

		b.Log.Debugf("Extra file is %#v", filetype)

		// TODO: add different types
		// TODO: add webp conversion
		switch filetype {
		case "image/jpeg", "image/png", "image/gif":
			return b.PostImageMessage(msg, filetype)
		default:
			return b.PostDocumentMessage(msg, filetype)
		}
	}

	// Post text message
	message := whatsapp.TextMessage{
		Info: whatsapp.MessageInfo{
			RemoteJid: msg.Channel, // which equals to group id
		},
		Text: msg.Username + msg.Text,
	}

	b.Log.Debugf("=> Sending %#v", msg)

	return b.conn.Send(message)
}

// TODO do we want that? to allow login with QR code from a bridged channel? https://github.com/tulir/mautrix-whatsapp/blob/513eb18e2d59bada0dd515ee1abaaf38a3bfe3d5/commands.go#L76
//func (b *Bwhatsapp) Command(cmd string) string {
//	return ""
//}
