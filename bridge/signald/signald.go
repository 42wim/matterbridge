package bsignald

import (
	"bufio"
	"net"
	"encoding/json"
	"github.com/42wim/matterbridge/bridge"
	"github.com/42wim/matterbridge/bridge/config"
	"gitlab.com/signald/signald-go/signald"
	//"gitlab.com/signald/signald-go/signald/client-protocol/v0"
	"gitlab.com/signald/signald-go/signald/client-protocol/v1"
)

type JSONCMD map[string]interface{}

const (
	cfgNumber = "Number"
	cfgSocket = "UnixSocket"
	cfgGroupID = "GroupID"
)

type envelopeResponse struct {
	ID   string                 `json:",omitempty"`
	Data v1.JsonMessageEnvelope `json:",omitempty"`
	Type string                 `json:",omitempty"`
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
			b.Log.Debugln(raw);

			var msg signald.BasicResponse
			if err := json.Unmarshal([]byte(raw), &msg); err != nil {
				b.Log.Errorln("Error unmarshaling raw response:", err.Error())
				continue
			}

			if msg.Type == "unexpected_error" {
				var errorResponse signald.UnexpectedError
				if err := json.Unmarshal(msg.Data, &errorResponse); err != nil {
					b.Log.Errorln("signald-go: Error unmarshaling error response:", err.Error())
					continue
				}
				b.Log.Errorln("signald-go: Unexpected error", errorResponse.Message)
				continue
			}

			if msg.Type != "message" {
				b.Log.Debugln("not 'message' from signald: ", raw);
				continue
			}

			response := envelopeResponse{ID: msg.ID, Type: msg.Type}
			if err := json.Unmarshal(msg.Data, &response.Data); err != nil {
				b.Log.Errorln("signald-go receive error: ", err)
				continue
			}

			b.Log.Debugf("%#v", response);

			if response.Data.DataMessage != nil {
				if response.Data.DataMessage.GroupV2 != nil {
					if b.groupid == response.Data.DataMessage.GroupV2.ID {
						rmsg := config.Message{
							UserID:   response.Data.Username,
							Username: response.Data.Username,
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

			//req := v1.SendRequest{
				//Username:    account,
				//MessageBody: strings.Join(args[1:], " "),
			//}

			//if strings.HasPrefix(args[0], "+") {
				//req.RecipientAddress = &v1.JsonAddress{Number: args[0]}
			//} else {
				//req.RecipientGroupID = args[0]
			//}

			//resp, err := req.Submit(common.Signald)
			//if err != nil {
				//log.Fatal("error sending request to signald: ", err)
			//}

    return "", nil
}
