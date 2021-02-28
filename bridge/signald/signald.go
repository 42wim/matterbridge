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
	GroupId  string          `json:"groupId,omitempty"`
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
	Username         string          `json:"username,omitempty"`
	//RecipientAddress signaldAccount  `json:"recipientAddress,omitempty"`
	RecipientGroupId string          `json:"recipientGroupId,omitempty"`
	MessageBody      string          `json:"messageBody,omitempty"`
	//Attachments      json.RawMessage `json:"attachments,omitempty"`
	//Quote            json.RawMessage `json:"quote,omitempty"`
	//Timestamp        int64           `json:"timestamp,omitempty"`
	//Mentions         json.RawMessage `json:"mentions,omitempty"`
}

type Bsignald struct {
	*bridge.Config
	socketpath string
	socket     net.Conn
	subscribed bool
	reader     *bufio.Scanner
	//listeners  map[string]chan signald.BasicResponse
	groupid   string
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
	}
}

func (b *Bsignald) Connect() error {
	b.Log.Infof("Connecting %s", b.socketpath)

	s, err := net.Dial("unix", b.socketpath)
	if err != nil {
		b.Log.Fatalf(err.Error())
	}
	//defer s.Close()
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
			if err := b.reader.Err(); err != nil {
				b.Log.Errorf(err.Error())
				continue
			}

			raw := b.reader.Text()

			var msg signaldMessage
			if err := json.Unmarshal([]byte(raw), &msg); err != nil {
				b.Log.Errorln("Error unmarshaling raw response:", err.Error())
				continue
			}

			if msg.Type == "unexpected_error" {
				var errorResponse signaldUnexpectedError
				if err := json.Unmarshal(msg.Data, &errorResponse); err != nil {
					b.Log.Errorln("Error unmarshaling error response:", err.Error())
					continue
				}
				b.Log.Errorln("Unexpected error", errorResponse.Message)
				continue
			}

			if msg.Type != "message" {
				b.Log.Debugln("skipping: not 'message'");
				continue
			} else {
				b.Log.Debugln("FOUND A MESSAGE!", raw);

			}

			response := signaldMessageData{ID: msg.ID, Type: msg.Type}
			if err := json.Unmarshal(msg.Data, &response.Data); err != nil {
				b.Log.Errorln("receive error: ", err)
				continue
			}

			//b.Log.Debugf("%#v", response);

			if response.Data.DataMessage != nil {
				if response.Data.DataMessage.GroupV2 != nil {
					if b.groupid == response.Data.DataMessage.GroupV2.ID {
						rmsg := config.Message{
							UserID:   response.Data.Source.UUID,
							Username: response.Data.Source.Number,
							Text:     response.Data.DataMessage.Body,
							Channel:  response.Data.DataMessage.GroupV2.ID,
							Account:  b.Account,
							Protocol: b.Protocol,
						}
						b.Log.Debugf("<= Sending message from %s on %s to gateway", rmsg.Username, b.Account)
						b.Log.Debugf("<= Message is %#v", rmsg)
						b.Remote <- rmsg
					}
				}
			}

			//if response.Data.SyncMessage != nil {
				//if response.Data.SyncMessage.Sent != nil {
					//if response.Data.SyncMessage.Sent.Message != nil {
						//if response.Data.SyncMessage.Sent.Message != nil {
							//if response.Data.SyncMessage.Sent.Message.GroupV2 != nil {
								//if b.groupid == response.Data.SyncMessage.Sent.Message.GroupV2.id {

								//}
							//}
						//}
					//}
				//}
			//}
		}
	}
}


func (b *Bsignald) Login() error {
	if ! b.subscribed {
		subscribe := JSONCMD{
			"type": "subscribe",
			"username": b.GetString(cfgNumber),
		}
		err := json.NewEncoder(b.socket).Encode(subscribe)
		if err != nil {
			b.Log.Fatalf(err.Error())
		}
		// TODO: this should be done from the listener after the response
		// was checked
		b.subscribed = true
	}
	return nil
}

func (b *Bsignald) Disconnect() error {
	b.Log.Debugln("Disconnecting..")
	b.socket.Close()
	return nil
}

func (b *Bsignald) Send(msg config.Message) (string, error) {
	b.Log.Debugf("message to forward into signal: %#v", msg)

	msgJSON := JSONCMD{
		"type": "send",
		"username": b.GetString(cfgNumber),
		"recipientGroupId": b.groupid,
		"messageBody": msg.Text,
	}
	err := json.NewEncoder(b.socket).Encode(msgJSON)
	if err != nil {
		b.Log.Errorln(err.Error())
	}

    return "", err
}
