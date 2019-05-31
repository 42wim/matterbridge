package bwhatsapp

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/42wim/matterbridge/bridge"
	"github.com/42wim/matterbridge/bridge/config"
	"github.com/Rhymen/go-whatsapp"

	whatsappExt "maunium.net/go/mautrix-whatsapp/whatsapp-ext"
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

	// https://github.com/Rhymen/go-whatsapp/blob/c31092027237441cffba1b9cb148eadf7c83c3d2/session.go#L18-L21
	session *whatsapp.Session
	conn    *whatsapp.Conn
	// https://github.com/tulir/mautrix-whatsapp/blob/master/whatsapp-ext/whatsapp.go
	connExt   *whatsappExt.ExtendedConn
	startedAt uint64

	users       map[string]whatsapp.Contact
	userAvatars map[string]string
}

// New Create a new WhatsApp bridge. This will be called for each [whatsapp.<server>] entry you have in the config file
func New(cfg *bridge.Config) bridge.Bridger {
	number := cfg.GetString(cfgNumber)
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
// https://github.com/42wim/matterbridge/blob/2cfd880cdb0df29771bf8f31df8d990ab897889d/bridge/bridge.go#L11-L16
func (b *Bwhatsapp) Connect() error {
	b.RLock() // TODO do we need locking for Whatsapp?
	defer b.RUnlock()

	number := b.GetString(cfgNumber)
	if number == "" {
		return errors.New("WhatsApp's telephone Number need to be configured")
	}

	// https://github.com/Rhymen/go-whatsapp#creating-a-connection
	b.Log.Debugln("Connecting to WhatsApp..")
	conn, err := whatsapp.NewConn(20 * time.Second)
	if err != nil {
		return errors.New("failed to connect to WhatsApp: " + err.Error())
	}

	b.conn = conn
	b.connExt = whatsappExt.ExtendConn(b.conn)
	// TODO do we want to use it? b.connExt.SetClientName("Matterbridge WhatsApp bridge", "mb-wa")

	b.conn.AddHandler(b)
	b.Log.Debugln("WhatsApp connection successful")

	// load existing session in order to keep it between restarts
	if b.session == nil {
		var session whatsapp.Session
		session, err = b.readSession()

		if err == nil {
			b.Log.Debugln("Restoring WhatsApp session..")

			// https://github.com/Rhymen/go-whatsapp#restore
			session, err = b.conn.RestoreWithSession(session)
			if err != nil {
				// TODO return or continue to normal login?
				// restore session connection timed out (I couldn't get over it without logging in again)
				return errors.New("failed to restore session: " + err.Error())
			}

			b.session = &session
			b.Log.Debugln("Session restored successfully!")
		} else {
			b.Log.Warn(err.Error())
		}
	}

	// login to a new session
	if b.session == nil {
		err = b.Login()
		if err != nil {
			return err
		}
	}
	b.startedAt = uint64(time.Now().Unix())

	_, err = b.conn.Contacts()
	if err != nil {
		return fmt.Errorf("error on update of contacts: %v", err)
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
			info, err := b.connExt.GetProfilePicThumb(jid)
			if err != nil {
				b.Log.Warnf("Could not get profile photo of %s: %v", jid, err)

			} else {
				// TODO any race conditions here?
				b.userAvatars[jid] = info.URL
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

	// TODO change connection strings to configured ones longClientName:"github.com/rhymen/go-whatsapp", shortClientName:"go-whatsapp"}" prefix=whatsapp
	// TODO get also a nice logo

	// TODO notification about unplugged and dead battery
	// conn.Info: Wid, Pushname, Connected, Battery, Plugged

	return nil
}

// Disconnect is called while reconnecting to the bridge
// TODO 42wim Documentation would be helpful on when reconnects happen and what should be done in this function
// Required implementation of the Bridger interface
// https://github.com/42wim/matterbridge/blob/2cfd880cdb0df29771bf8f31df8d990ab897889d/bridge/bridge.go#L11-L16
func (b *Bwhatsapp) Disconnect() error {
	// We could Logout, but that would close the session completely and would require a new QR code scan
	// https://github.com/Rhymen/go-whatsapp/blob/c31092027237441cffba1b9cb148eadf7c83c3d2/session.go#L377-L381
	return nil
}

func isGroupJid(identifier string) bool {
	return strings.HasSuffix(identifier, "@g.us") || strings.HasSuffix(identifier, "@temp")
}

// JoinChannel Join a WhatsApp group specified in gateway config as channel='number-id@g.us' or channel='Channel name'
// Required implementation of the Bridger interface
// https://github.com/42wim/matterbridge/blob/2cfd880cdb0df29771bf8f31df8d990ab897889d/bridge/bridge.go#L11-L16
func (b *Bwhatsapp) JoinChannel(channel config.ChannelInfo) error {
	byJid := isGroupJid(channel.Name)

	// verify if we are member of the given group
	if byJid {
		// channel.Name specifies static group jID, not the name
		if _, exists := b.conn.Store.Contacts[channel.Name]; !exists {
			return fmt.Errorf("account doesn't belong to group with jid %s", channel.Name)
		}
	} else {
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
			// TODO sort
			// copy b;
			//sort.Slice(people, func(i, j int) bool {
			//	return people[i].Age > people[j].Age
			//})
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

	return nil
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

			// TODO 42wim Doesn't the app get clogged with a ton of IDs after some time of running?
			// WhatsApp allows to set any ID so in that case we could use external IDs and don't do mapping
			// but external IDs are not set
			return "", nil
		}
		// TODO delete message on WhatsApp https://github.com/Rhymen/go-whatsapp/issues/100
		return "", nil
	}

	// Edit message
	if msg.ID != "" {
		b.Log.Debugf("updating message with id %s", msg.ID)

		msg.Text += " (edited)"
		// TODO handle edit as a message reply with updated text
	}

	//// TODO Handle Upload a file
	//if msg.Extra != nil {
	//	for _, rmsg := range helper.HandleExtra(&msg, b.General) {
	//		b.c.SendMessage(roomID, rmsg.Username+rmsg.Text)
	//	}
	//	if len(msg.Extra["file"]) > 0 {
	//		return b.handleUploadFile(&msg, roomID)
	//	}
	//}

	// Post text message
	text := whatsapp.TextMessage{
		Info: whatsapp.MessageInfo{
			RemoteJid: msg.Channel, // which equals to group id
		},
		Text: msg.Username + msg.Text,
	}

	b.Log.Debugf("=> Sending %#v", msg)

	// create message ID
	// TODO follow and act if https://github.com/Rhymen/go-whatsapp/issues/101 implemented
	bytes := make([]byte, 10)
	if _, err := rand.Read(bytes); err != nil {
		b.Log.Warn(err.Error())
	}
	text.Info.Id = strings.ToUpper(hex.EncodeToString(bytes))

	_, err := b.conn.Send(text)

	return text.Info.Id, err
}

// TODO do we want that? to allow login with QR code from a bridged channel? https://github.com/tulir/mautrix-whatsapp/blob/513eb18e2d59bada0dd515ee1abaaf38a3bfe3d5/commands.go#L76
//func (b *Bwhatsapp) Command(cmd string) string {
//	return ""
//}
