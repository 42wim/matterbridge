package bwhatsapp

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/42wim/matterbridge/bridge"
	"github.com/42wim/matterbridge/bridge/config"

	"github.com/Rhymen/go-whatsapp"
	"maunium.net/go/mautrix-whatsapp/whatsapp-ext"
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

	// TODO do we need cache?
	//newCache, err := lru.New(5000)
	//if err != nil {
	//	cfg.Log.Fatalf("Could not create LRU cache for Slack bridge: %v", err)
	//}
	b := &Bwhatsapp{
		Config: cfg,

		users:       make(map[string]whatsapp.Contact),
		userAvatars: make(map[string]string),

		//uuid:                   xid.New().String(),
		//users:                  map[string]*slack.User{},
		//channelsByID:           map[string]*slack.Channel{},
		//channelsByName:         map[string]*slack.Channel{},
		//earliestChannelRefresh: time.Now(),
		//earliestUserRefresh:    time.Now(),
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
		return errors.New("Failed to connect to WhatsApp: " + err.Error())
	}

	b.conn = conn
	b.connExt = whatsappExt.ExtendConn(b.conn)
	// TODO do we want to use it? b.connExt.SetClientName("Matterbridge WhatsApp bridge", "mb-wa")

	b.conn.AddHandler(b)
	b.Log.Debugln("WhatsApp connection successful")

	// load existing session in order to keep it between restarts
	// TODO try to load session from env vars or otherwise for Azure and other clouds
	// now implemented: load session from file
	if b.session == nil {
		session, err := b.readSession()

		if err == nil {
			b.Log.Debugln("Restoring WhatsApp session..")
			sess, err := b.conn.RestoreSession(session) // https://github.com/Rhymen/go-whatsapp#restore
			if err != nil {                             // restore session connection timed out
				// TODO return or continue to normal login?
				return errors.New("Failed to restore session: " + err.Error())
			}

			b.session = &sess
			b.Log.Debugln("Session restored successfully!")
		} else {
			b.Log.Warn(err.Error())
		}
	}

	// login to a new session
	if b.session == nil {
		if err := b.Login(); err != nil {
			return err
		}
	}
	b.startedAt = uint64(time.Now().Unix())

	_, err = b.conn.Contacts()
	if err != nil {
		b.Log.Errorln("Error on update of contacts: %v", err)
		return nil
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

	// TODO qrCode, err := qrcode.Encode(code, qrcode.Low, 256) to encode as image/png
	// and possibly send it to connected channels (to admin) to authorize the app
	invert := b.GetBoolOrDefault(qrOnWhiteTerminal, false)
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

	// session.Wid
	// conn.Info: Wid, Pushname, Connected, Battery, Plugged (TODO notification about unplugged and dead battery)
	// jid = strings.Replace(b.conn.Info.Wid, whatsappExt.OldUserSuffix, whatsappExt.NewUserSuffix, 1)

	return nil
}

// Disconnect TODO What does it mean
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
			return fmt.Errorf("Account doesn't belong to group with jid %s", channel.Name)
		}
	} else {
		// channel.Name specifies group name that might change, warn about it
		var jids []string
		for id, contact := range b.conn.Store.Contacts {
			if isGroupJid(id) && contact.Name == channel.Name {
				jids = append(jids, id)
			}
		}

		if len(jids) == 0 {
			// didn't match any group - print out possibilites
			// TODO sort
			// copy b;
			//sort.Slice(people, func(i, j int) bool {
			//	return people[i].Age > people[j].Age
			//})
			for id, contact := range b.conn.Store.Contacts {
				if isGroupJid(id) {
					// TODO b.Log.Info
					fmt.Printf("%s %s\n", contact.Jid, contact.Name)
				}
			}
			return fmt.Errorf("Please specify group's JID from the below list instead of the name '%s'", channel.Name)

		} else if len(jids) > 1 {
			return fmt.Errorf("There is more than one group with name '%s'. Please specify one of JIDs as channel name: %v", channel.Name, jids)

		} else {
			return fmt.Errorf("Group name might change. Please configure gateway with channel=\"%v\" instead of channel=\"%v\"", jids[0], channel.Name)
		}
	}

	return nil
}
