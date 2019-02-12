package bwhatsapp

import (
	"encoding/gob"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/42wim/matterbridge/bridge"
	"github.com/42wim/matterbridge/bridge/config"

	"github.com/Baozisoftware/qrcode-terminal-go"
	"github.com/Rhymen/go-whatsapp"
	"maunium.net/go/mautrix-whatsapp/whatsapp-ext"
)

const (

	// Account config parameters
	cfgNumber = "Number"
)

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

// TODO do we want that? to allow login with QR code from a bridged channel? https://github.com/tulir/mautrix-whatsapp/blob/513eb18e2d59bada0dd515ee1abaaf38a3bfe3d5/commands.go#L76
//func (b *Bwhatsapp) Command(cmd string) string {
//	return ""
//}

// TODO learning GO: What is "(b *Bwhatsapp)" in this function's signature? Not argument and not a return value, so what? Does it add method on struct?
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

func (b *Bwhatsapp) Login() error {
	b.Log.Debugln("Logging in..")

	// TODO qrCode, err := qrcode.Encode(code, qrcode.Low, 256) to encode as image/png
	// and possibly send it to connected channels (to admin) to authorize the app
	// TODO invert configured in settings
	qrChan := qrFromTerminal(true)

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
	sessionFile := b.Config.GetString("SessionFile")

	if sessionFile == "" {
		return session, errors.New("If you won't set SessionFile then you will need to scan QR code on every restart")
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
	sessionFile := b.Config.GetString("SessionFile")

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

func (b *Bwhatsapp) Disconnect() error {
	return nil
}

func isGroupJid(identifier string) bool {
	return strings.HasSuffix(identifier, "@g.us") || strings.HasSuffix(identifier, "@temp")
}

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

func (b *Bwhatsapp) Send(msg config.Message) (string, error) {
	b.Log.Debugf("=> Receiving %#v", msg)

	// msg.Channel target group name
	// msg.Username empty
	// msg.UserID a weird string , probably slack user id
	// msg.Avatar has a nice image
	// msg.Timestamp has a nice timestamp with loc(ation) / timezone
	// msg.ID empty, // TODO why empty?!

	text := whatsapp.TextMessage{
		Info: whatsapp.MessageInfo{
			// Id: "", // TODO id
			// TODO Timestamp
			RemoteJid: msg.Channel, // which equals to group id

		},
		Text: msg.Username + msg.Text,
	}

	// TODO adapt gitter code
	//roomID := b.getRoomID(msg.Channel)
	//if roomID == "" {
	//	b.Log.Errorf("Could not find roomID for %v", msg.Channel)
	//	return "", nil
	//}
	//
	//// Delete message
	//if msg.Event == config.EventMsgDelete {
	//	if msg.ID == "" {
	//		return "", nil
	//	}
	//	// gitter has no delete message api so we edit message to ""
	//	_, err := b.c.UpdateMessage(roomID, msg.ID, "")
	//	if err != nil {
	//		return "", err
	//	}
	//	return "", nil
	//}
	//
	//// Upload a file (in gitter case send the upload URL because gitter has no native upload support)
	//if msg.Extra != nil {
	//	for _, rmsg := range helper.HandleExtra(&msg, b.General) {
	//		b.c.SendMessage(roomID, rmsg.Username+rmsg.Text)
	//	}
	//	if len(msg.Extra["file"]) > 0 {
	//		return b.handleUploadFile(&msg, roomID)
	//	}
	//}
	//
	//// Edit message
	//if msg.ID != "" {
	//	b.Log.Debugf("updating message with id %s", msg.ID)
	//	_, err := b.c.UpdateMessage(roomID, msg.ID, msg.Username+msg.Text)
	//	if err != nil {
	//		return "", err
	//	}
	//	return "", nil
	//}
	//
	//// Post normal message
	//resp, err := b.c.SendMessage(roomID, msg.Username+msg.Text)
	//if err != nil {
	//	return "", err
	//}
	//return resp.ID, nil

	b.Log.Debugf("=> Sending %#v", msg)

	err := b.conn.Send(text)

	// TODO return message id
	return "", err
}

// ================================================================
// handlers https://github.com/Rhymen/go-whatsapp#add-message-handlers & https://github.com/Rhymen/go-whatsapp/blob/master/handler.go

func (b *Bwhatsapp) HandleError(err error) {
	b.Log.Errorf("%v", err) // TODO implement proper handling? at least respond to different error types
}

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
	senderName := senderJid
	if sender, exists := b.users[senderJid]; exists {
		if sender.Name != "" {
			senderName = sender.Name

		} else {
			// if user is not in phone contacts
			// it is the most obvious scenario unless you sync your phone contacts with some remote updated source
			// users can change it in their WhatsApp settings -> profile -> click on Avatar
			senderName = sender.Notify
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

	if avatarUrl, exists := b.userAvatars[senderJid]; exists {
		rmsg.Avatar = avatarUrl
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

// TODO questions to Tulir
// Why are you locking on message processing? https://github.com/tulir/mautrix-whatsapp/blob/513eb18e2d59bada0dd515ee1abaaf38a3bfe3d5/portal.go#L212
// How are you showing nicks in WhatsApp? Inside message?
