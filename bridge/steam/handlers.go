package bsteam

import (
	"fmt"
	"strconv"

	"github.com/42wim/matterbridge/bridge/config"
	"github.com/Philipp15b/go-steam"
	"github.com/Philipp15b/go-steam/protocol/steamlang"
)

func (b *Bsteam) handleChatMsg(e *steam.ChatMsgEvent) {
	b.Log.Debugf("Receiving ChatMsgEvent: %#v", e)
	b.Log.Debugf("<= Sending message from %s on %s to gateway", b.getNick(e.ChatterId), b.Account)
	var channel int64
	if e.ChatRoomId == 0 {
		channel = int64(e.ChatterId)
	} else {
		// for some reason we have to remove 0x18000000000000
		// TODO
		// https://github.com/42wim/matterbridge/pull/630#discussion_r238102751
		// channel = int64(e.ChatRoomId) & 0xfffffffffffff
		channel = int64(e.ChatRoomId) - 0x18000000000000
	}
	msg := config.Message{
		Username: b.getNick(e.ChatterId),
		Text:     e.Message,
		Channel:  strconv.FormatInt(channel, 10),
		Account:  b.Account,
		UserID:   strconv.FormatInt(int64(e.ChatterId), 10),
	}
	b.Remote <- msg
}

func (b *Bsteam) handleEvents() {
	myLoginInfo := &steam.LogOnDetails{
		Username: b.GetString("Login"),
		Password: b.GetString("Password"),
		AuthCode: b.GetString("AuthCode"),
	}
	// TODO Attempt to read existing auth hash to avoid steam guard.
	// Maybe works
	//myLoginInfo.SentryFileHash, _ = ioutil.ReadFile("sentry")
	for event := range b.c.Events() {
		switch e := event.(type) {
		case *steam.ChatMsgEvent:
			b.handleChatMsg(e)
		case *steam.PersonaStateEvent:
			b.Log.Debugf("PersonaStateEvent: %#v\n", e)
			b.Lock()
			b.userMap[e.FriendId] = e.Name
			b.Unlock()
		case *steam.ConnectedEvent:
			b.c.Auth.LogOn(myLoginInfo)
		case *steam.MachineAuthUpdateEvent:
		// TODO sentry files for 2 auth
		/*
			b.Log.Info("authupdate", e)
			b.Log.Info("hash", e.Hash)
			ioutil.WriteFile("sentry", e.Hash, 0666)
		*/
		case *steam.LogOnFailedEvent:
			b.Log.Info("Logon failed", e)
			err := b.handleLogOnFailed(e, myLoginInfo)
			if err != nil {
				b.Log.Error(err)
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
			b.Log.Errorf("steam FatalErrorEvent: %#v", e)
		default:
			b.Log.Debugf("unknown event %#v", e)
		}
	}
}

func (b *Bsteam) handleLogOnFailed(e *steam.LogOnFailedEvent, myLoginInfo *steam.LogOnDetails) error {
	switch e.Result {
	case steamlang.EResult_AccountLoginDeniedNeedTwoFactor:
		b.Log.Info("Steam guard isn't letting me in! Enter 2FA code:")
		var code string
		fmt.Scanf("%s", &code)
		// TODO https://github.com/42wim/matterbridge/pull/630#discussion_r238103978
		myLoginInfo.TwoFactorCode = code
	case steamlang.EResult_AccountLogonDenied:
		b.Log.Info("Steam guard isn't letting me in! Enter auth code:")
		var code string
		fmt.Scanf("%s", &code)
		// TODO https://github.com/42wim/matterbridge/pull/630#discussion_r238103978
		myLoginInfo.AuthCode = code
	case steamlang.EResult_InvalidLoginAuthCode:
		return fmt.Errorf("Steam guard: invalid login auth code: %#v ", e.Result)
	default:
		return fmt.Errorf("LogOnFailedEvent: %#v ", e.Result)
		// TODO: Handle EResult_InvalidLoginAuthCode
	}
	return nil
}

// handleFileInfo handles config.FileInfo and adds correct file comment or URL to msg.Text.
// Returns error if cast fails.
func (b *Bsteam) handleFileInfo(msg *config.Message, f interface{}) error {
	if _, ok := f.(config.FileInfo); !ok {
		return fmt.Errorf("handleFileInfo cast failed %#v", f)
	}
	fi := f.(config.FileInfo)
	if fi.Comment != "" {
		msg.Text += fi.Comment + ": "
	}
	if fi.URL != "" {
		msg.Text = fi.URL
		if fi.Comment != "" {
			msg.Text = fi.Comment + ": " + fi.URL
		}
	}
	return nil
}
