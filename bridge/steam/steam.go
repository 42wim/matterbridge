package bsteam

import (
	"fmt"
	"github.com/42wim/matterbridge/bridge"
	"github.com/42wim/matterbridge/bridge/config"
	"github.com/Philipp15b/go-steam"
	"github.com/Philipp15b/go-steam/protocol/steamlang"
	"github.com/Philipp15b/go-steam/steamid"
	//"io/ioutil"
	"strconv"
	"sync"
	"time"
)

type Bsteam struct {
	c         *steam.Client
	connected chan struct{}
	userMap   map[steamid.SteamId]string
	sync.RWMutex
	*config.BridgeConfig
}

func New(cfg *config.BridgeConfig) bridge.Bridger {
	b := &Bsteam{BridgeConfig: cfg}
	b.userMap = make(map[steamid.SteamId]string)
	b.connected = make(chan struct{})
	return b
}

func (b *Bsteam) Connect() error {
	b.Log.Info("Connecting")
	b.c = steam.NewClient()
	go b.handleEvents()
	go b.c.Connect()
	select {
	case <-b.connected:
		b.Log.Info("Connection succeeded")
	case <-time.After(time.Second * 30):
		return fmt.Errorf("connection timed out")
	}
	return nil
}

func (b *Bsteam) Disconnect() error {
	b.c.Disconnect()
	return nil

}

func (b *Bsteam) JoinChannel(channel config.ChannelInfo) error {
	id, err := steamid.NewId(channel.Name)
	if err != nil {
		return err
	}
	b.c.Social.JoinChat(id)
	return nil
}

func (b *Bsteam) Send(msg config.Message) (string, error) {
	// ignore delete messages
	if msg.Event == config.EVENT_MSG_DELETE {
		return "", nil
	}
	id, err := steamid.NewId(msg.Channel)
	if err != nil {
		return "", err
	}
	b.c.Social.SendMessage(id, steamlang.EChatEntryType_ChatMsg, msg.Username+msg.Text)
	return "", nil
}

func (b *Bsteam) getNick(id steamid.SteamId) string {
	b.RLock()
	defer b.RUnlock()
	if name, ok := b.userMap[id]; ok {
		return name
	}
	return "unknown"
}

func (b *Bsteam) handleEvents() {
	myLoginInfo := new(steam.LogOnDetails)
	myLoginInfo.Username = b.Config.Login
	myLoginInfo.Password = b.Config.Password
	myLoginInfo.AuthCode = b.Config.AuthCode
	// Attempt to read existing auth hash to avoid steam guard.
	// Maybe works
	//myLoginInfo.SentryFileHash, _ = ioutil.ReadFile("sentry")
	for event := range b.c.Events() {
		//b.Log.Info(event)
		switch e := event.(type) {
		case *steam.ChatMsgEvent:
			b.Log.Debugf("Receiving ChatMsgEvent: %#v", e)
			b.Log.Debugf("Sending message from %s on %s to gateway", b.getNick(e.ChatterId), b.Account)
			var channel int64
			if e.ChatRoomId == 0 {
				channel = int64(e.ChatterId)
			} else {
				// for some reason we have to remove 0x18000000000000
				channel = int64(e.ChatRoomId) - 0x18000000000000
			}
			msg := config.Message{Username: b.getNick(e.ChatterId), Text: e.Message, Channel: strconv.FormatInt(channel, 10), Account: b.Account, UserID: strconv.FormatInt(int64(e.ChatterId), 10)}
			b.Remote <- msg
		case *steam.PersonaStateEvent:
			b.Log.Debugf("PersonaStateEvent: %#v\n", e)
			b.Lock()
			b.userMap[e.FriendId] = e.Name
			b.Unlock()
		case *steam.ConnectedEvent:
			b.c.Auth.LogOn(myLoginInfo)
		case *steam.MachineAuthUpdateEvent:
			/*
				b.Log.Info("authupdate", e)
				b.Log.Info("hash", e.Hash)
				ioutil.WriteFile("sentry", e.Hash, 0666)
			*/
		case *steam.LogOnFailedEvent:
			b.Log.Info("Logon failed", e)
			switch e.Result {
			case steamlang.EResult_AccountLogonDeniedNeedTwoFactorCode:
				{
					b.Log.Info("Steam guard isn't letting me in! Enter 2FA code:")
					var code string
					fmt.Scanf("%s", &code)
					myLoginInfo.TwoFactorCode = code
				}
			case steamlang.EResult_AccountLogonDenied:
				{
					b.Log.Info("Steam guard isn't letting me in! Enter auth code:")
					var code string
					fmt.Scanf("%s", &code)
					myLoginInfo.AuthCode = code
				}
			default:
				b.Log.Errorf("LogOnFailedEvent: %#v ", e.Result)
				// TODO: Handle EResult_InvalidLoginAuthCode
				return
			}
		case *steam.LoggedOnEvent:
			b.Log.Debugf("LoggedOnEvent: %#v", e)
			b.connected <- struct{}{}
			b.Log.Debugf("setting online")
			b.c.Social.SetPersonaState(steamlang.EPersonaState_Online)
		case *steam.DisconnectedEvent:
			b.Log.Info("Disconnected")
			b.Log.Info("Attempting to reconnect...")
			b.c.Connect()
		case steam.FatalErrorEvent:
			b.Log.Error(e)
		case error:
			b.Log.Error(e)
		default:
			b.Log.Debugf("unknown event %#v", e)
		}
	}
}
