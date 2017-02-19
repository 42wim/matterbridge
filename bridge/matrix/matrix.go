package bmatrix

import (
	"github.com/42wim/matterbridge/bridge/config"
	log "github.com/Sirupsen/logrus"
	matrix "github.com/matrix-org/gomatrix"
)

type Bmatrix struct {
	mc      *matrix.Client
	Config  *config.Protocol
	Remote  chan config.Message
	Account string
	UserID  string
}

var flog *log.Entry
var protocol = "matrix"

func init() {
	flog = log.WithFields(log.Fields{"module": protocol})
}

func New(cfg config.Protocol, account string, c chan config.Message) *Bmatrix {
	b := &Bmatrix{}
	b.Config = &cfg
	b.Account = account
	b.Remote = c
	return b
}

func (b *Bmatrix) Connect() error {
	var err error
	flog.Infof("Connecting %s", b.Config.Server)
	b.mc, err = matrix.NewClient(b.Config.Server, "", "")
	if err != nil {
		flog.Debugf("%#v", err)
		return err
	}
	resp, err := b.mc.Login(&matrix.ReqLogin{
		Type:     "m.login.password",
		User:     b.Config.Login,
		Password: b.Config.Password,
	})
	if err != nil {
		flog.Debugf("%#v", err)
		return err
	}
	b.mc.SetCredentials(resp.UserID, resp.AccessToken)
	b.UserID = resp.UserID
	flog.Info("Connection succeeded")
	go b.handlematrix()
	return nil
}

func (b *Bmatrix) Disconnect() error {
	return nil
}

func (b *Bmatrix) JoinChannel(channel string) error {
	_, err := b.mc.JoinRoom(channel, "", nil)
	return err
}

func (b *Bmatrix) Send(msg config.Message) error {
	flog.Debugf("Receiving %#v", msg)
	b.mc.SendText(msg.Channel, msg.Username+msg.Text)
	return nil
}

func (b *Bmatrix) handlematrix() error {
	warning := "Not relaying this message, please setup a dedicated bot user"
	syncer := b.mc.Syncer.(*matrix.DefaultSyncer)
	syncer.OnEventType("m.room.message", func(ev *matrix.Event) {
		if ev.Content["msgtype"].(string) == "m.text" && ev.Sender != b.UserID {
			flog.Debugf("Sending message from %s on %s to gateway", ev.Sender, b.Account)
			b.Remote <- config.Message{Username: ev.Sender, Text: ev.Content["body"].(string), Channel: ev.RoomID, Account: b.Account}
		}
		if ev.Sender == b.UserID && ev.Content["body"].(string) != warning {
			b.mc.SendText(ev.RoomID, warning)
		}
		flog.Debugf("Received: %#v", ev)
	})
	go func() {
		for {
			if err := b.mc.Sync(); err != nil {
				flog.Println("Sync() returned ", err)
			}
		}
	}()
	return nil
}
