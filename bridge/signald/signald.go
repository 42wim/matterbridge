package bsignald

import (
	"bufio"
	"net"
	"encoding/json"
	"github.com/42wim/matterbridge/bridge"
	"github.com/42wim/matterbridge/bridge/config"
)

type JSONCMD map[string]interface{}

const (
	cfgNumber = "Number"
	cfgSocket = "UnixSocket"
	cfgGroupID = "GroupID"
)

type signaldMessage struct {
	ID    string
	Type  string
	Error json.RawMessage
	Data  json.RawMessage
}

type signaldUnexpectedError struct {
	Message string
}

type signaldMessageData struct {
	ID   string       `json:",omitempty"`
	Data signaldData  `json:",omitempty"`
	Type string       `json:",omitempty"`
}

type signaldData struct {
	CallMessage              json.RawMessage        `json:"callMessage,omitempty"`
	DataMessage              *signaldDataMessage    `json:"dataMessage,omitempty"`
	HasContent               bool                   `json:"hasContent,omitempty"`
	HasLegacyMessage         bool                   `json:"hasLegacyMessage,omitempty"`
	IsUnidentifiedSender     bool                   `json:"isUnidentifiedSender,omitempty"`
	Receipt                  json.RawMessage        `json:"receipt,omitempty"`
	Relay                    string                 `json:"relay,omitempty"`
	ServerDeliveredTimestamp int64                  `json:"serverDeliveredTimestamp,omitempty"`
	ServerTimestamp          int64                  `json:"serverTimestamp,omitempty"`
	Source                   *signaldAccount         `json:"source,omitempty"`
	SourceDevice             int32                  `json:"sourceDevice,omitempty"`
	SyncMessage              json.RawMessage        `json:"syncMessage,omitempty"`
	Timestamp                int64                  `json:"timestamp,omitempty"`
	TimestampISO             string                 `json:"timestampISO,omitempty"`
	Type                     string                 `json:"type,omitempty"`
	Typing                   json.RawMessage        `json:"typing,omitempty"`
	Username                 string                 `json:"username,omitempty"`
	UUID                     string                 `json:"uuid,omitempty"`
}

type signaldAccount struct {
	Number string `json:"number,omitempty"`
	Relay  string `json:"relay,omitempty"`
	UUID   string `json:"uuid,omitempty"`
}

type signaldDataMessage struct {
	Attachments      json.RawMessage     `json:"attachments,omitempty"`
	Body             string              `json:"body,omitempty"`
	Contacts         json.RawMessage     `json:"contacts,omitempty"`
	EndSession       bool                `json:"endSession,omitempty"`
	ExpiresInSeconds int32               `json:"expiresInSeconds,omitempty"`
	Group            *signaldGroupInfo   `json:"group,omitempty"`
	GroupV2          *signaldGroupV2Info `json:"groupV2,omitempty"`
	Mentions         json.RawMessage     `json:"mentions,omitempty"`
	Previews         json.RawMessage     `json:"previews,omitempty"`
	ProfileKeyUpdate bool                `json:"profileKeyUpdate,omitempty"`
	Quote            json.RawMessage     `json:"quote,omitempty"`
	Reaction         json.RawMessage     `json:"reaction,omitempty"`
	RemoteDelete     json.RawMessage     `json:"remoteDelete,omitempty"`
	Sticker          json.RawMessage     `json:"sticker,omitempty"`
	Timestamp        int64               `json:"timestamp,omitempty"`
	ViewOnce         bool                `json:"viewOnce,omitempty"`
}

type signaldGroupInfo struct {
	AvatarId int64           `json:"avatarId,omitempty"`
	ID       string          `json:"groupId,omitempty"`
	Members  json.RawMessage `json:"members,omitempty"`
	Name     string          `json:"name,omitempty"`
	Type     string          `json:"type,omitempty"`
}

type signaldGroupV2Info struct {
	AccessControl       json.RawMessage  `json:"accessControl,omitempty"`
	Avatar              string           `json:"avatar,omitempty"`
	ID                  string           `json:"id,omitempty"`
	InviteLink          string           `json:"inviteLink,omitempty"`
	MemberDetail        json.RawMessage  `json:"memberDetail,omitempty"`
	Members             json.RawMessage  `json:"members,omitempty"`
	PendingMemberDetail json.RawMessage  `json:"pendingMemberDetail,omitempty"`
	PendingMembers      json.RawMessage  `json:"pendingMembers,omitempty"`
	RequestingMembers   json.RawMessage  `json:"requestingMembers,omitempty"`
	Revision            int32            `json:"revision,omitempty"`
	Timer               int32            `json:"timer,omitempty"`
	Title               string           `son:"title,omitempty"`
}

type signaldSendMessage struct {
	Type             string `json:"type,omitempty"`
	Username         string `json:"username,omitempty"`
	RecipientGroupId string `json:"recipientGroupId,omitempty"`
	MessageBody      string `json:"messageBody,omitempty"`
}

type signaldContact struct {
	Name                  string          `json:"name,omitempty"`
	ProfileName           string          `json:"profile_name,omitempty"`
	Account               *signaldAccount `json:"address,omitempty"`
	Avatar                string          `json:"avatar,omitempty"`
	Color                 string          `json:"color,omitempty"`
	ProfileKey            string          `json:"profileKey,omitempty"`
	MessageExpirationTime int32           `json:"messageExpirationTime,omitempty"`
	InboxPosition         int32           `json:"inboxPosition,omitempty"`
}

type Bsignald struct {
	*bridge.Config
	socketpath string
	socket     net.Conn
	subscribed bool
	reader     *bufio.Scanner
	groupid   string
	contacts  map[string]signaldContact
}

func New(cfg *bridge.Config) bridge.Bridger {
	number := cfg.GetString(cfgNumber)
	if number == "" {
		cfg.Log.Fatalf("Missing configuration for Signald bridge: Number")
	}

	socketpath := cfg.GetString(cfgSocket)
	if socketpath == "" {
		socketpath = "/var/run/signald/signald.sock"
	}

	return &Bsignald{
		Config: cfg,
		socketpath: socketpath,
		subscribed: false,
		contacts: make(map[string]signaldContact),
	}
}

func (b *Bsignald) Connect() error {
	b.Log.Infof("Connecting %s", b.socketpath)

	s, err := net.Dial("unix", b.socketpath)
	if err != nil {
		b.Log.Fatalf(err.Error())
	}
	b.socket = s
	r := bufio.NewScanner(s)
	b.reader = r
	go b.Listen()
	go b.Login()
	return nil
}

func (b *Bsignald) JoinChannel(channel config.ChannelInfo) error {
	b.groupid = channel.Name
	return nil
}

func (b *Bsignald) Listen() {
	for {
		for b.reader.Scan() {
			var err error
			if err = b.reader.Err(); err != nil {
				b.Log.Errorf(err.Error())
				continue
			}

			raw := b.reader.Text()

			var msg signaldMessage
			if err = json.Unmarshal([]byte(raw), &msg); err != nil {
				b.Log.Errorln("Error unmarshaling raw response:", err.Error())
				continue
			}

			if msg.Type == "subscribed" {
				b.Log.Debugln("subscribe successful", b.GetString(cfgNumber))
				b.subscribed = true
				go b.GetContacts()
				continue
			}

			if msg.Type == "listen_stopped" {
				b.Log.Errorln("got listen stopped, trying to re-subscribe")
				b.subscribed = false
				go b.Login()
				continue
			}

			if msg.Type == "unexpected_error" {
				var errorResponse signaldUnexpectedError
				if err = json.Unmarshal(msg.Data, &errorResponse); err != nil {
					b.Log.Errorln("Error unmarshaling error response:", err.Error())
					continue
				}
				b.Log.Errorln("Unexpected error", errorResponse.Message)
				continue
			}

			if msg.Type == "contact_list" {
				var contacts []signaldContact
				if err = json.Unmarshal(msg.Data, &contacts); err != nil {
					b.Log.Errorln("failed to parse contact_list: ", err)
				} else {
					for _, contact := range contacts {
						b.contacts[contact.Account.UUID] = contact
					}
					b.Log.Debugf("%#v", b.contacts)
				}
				continue
			}

			if msg.Type != "message" {
				b.Log.Debugln("skipping: not 'message'");
				continue
			}

			response := signaldMessageData{ID: msg.ID, Type: msg.Type}
			if err = json.Unmarshal(msg.Data, &response.Data); err != nil {
				b.Log.Errorln("receive error: ", err)
				continue
			}

			if response.Data.DataMessage != nil {
				groupMatched := false
				if response.Data.DataMessage.GroupV2 != nil {
					if b.groupid == response.Data.DataMessage.GroupV2.ID {
						groupMatched = true
					}
				}
				if response.Data.DataMessage.Group != nil {
					if b.groupid == response.Data.DataMessage.Group.ID {
						groupMatched = true
					}
				}

				if false == groupMatched {
					b.Log.Debugln("skipping non-group message")
					continue
				}

				username := response.Data.Source.Number
				if v, found := b.contacts[response.Data.Source.UUID]; found {
					if "" != v.ProfileName {
						username = v.ProfileName
					} else if "" != v.Name {
						username = v.Name
					}
				}
				rmsg := config.Message{
					UserID:   response.Data.Source.UUID,
					Username: username,
					Text:     response.Data.DataMessage.Body,
					Channel:  b.groupid,
					Account:  b.Account,
					Protocol: b.Protocol,
				}

				b.Log.Debugf("<= Sending message from %s on %s to gateway", rmsg.Username, b.Account)
				b.Log.Debugf("<= Message is %#v", rmsg)
				b.Remote <- rmsg

				// TODO: send read receipt
			}
		}
	}
}

func (b *Bsignald) GetContacts() error {
	cmd := JSONCMD{
		"type": "list_contacts",
		"username": b.GetString(cfgNumber),
	}
	return b.SendRawJSON(cmd)
}

func (b *Bsignald) Login() error {
	var err error
	if ! b.subscribed {
		cmd := JSONCMD{
			"type": "subscribe",
			"username": b.GetString(cfgNumber),
		}
		err = b.SendRawJSON(cmd)
	}
	return err
}

func (b *Bsignald) SendRawJSON(cmd JSONCMD) (error) {
	err := json.NewEncoder(b.socket).Encode(cmd)
	if err != nil {
		b.Log.Errorln(err.Error())
	}
	return err
}

func (b *Bsignald) Disconnect() error {
	b.Log.Debugln("Disconnecting..")
	b.socket.Close()
	return nil
}

func (b *Bsignald) Send(msg config.Message) (string, error) {
	b.Log.Debugf("message to forward into signal: %#v", msg)

	msgJSON := signaldSendMessage {
		Type: "send",
		Username: b.GetString(cfgNumber),
		RecipientGroupId: b.groupid,
		MessageBody: msg.Username + msg.Text,
	}

	err := json.NewEncoder(b.socket).Encode(msgJSON)
	if err != nil {
		b.Log.Errorln(err.Error())
	}
	return "", err
}
