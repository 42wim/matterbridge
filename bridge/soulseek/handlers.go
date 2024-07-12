package bsoulseek

import (
	"fmt"
	"strings"

	"github.com/42wim/matterbridge/bridge/config"
)

func (b *Bsoulseek) handleMessage(msg soulseekMessageResponse) {
	if msg != nil {
		b.Log.Debugf("Handling message: %v", msg)
	}
	switch msg := msg.(type) {
	case loginMessageResponseSuccess, loginMessageResponseFailure:
		b.loginResponse <- msg
	case joinRoomMessageResponse:
		b.joinRoomResponse <- msg
	case kickedMessageResponse:
		b.fatalErrors <- fmt.Errorf("Logged in somewhere else")
	case privateMessageReceive:
		b.handleDM(msg)
	case sayChatroomMessageReceive:
		b.handleChatMessage(msg)
	case userJoinedRoomMessage:
		b.handleJoinMessage(msg)
	case userLeftRoomMessage:
		b.handleLeaveMessage(msg)
	default:
		// do nothing
	}
}

func (b *Bsoulseek) handleChatMessage(msg sayChatroomMessageReceive) {
	b.Log.Debugf("Handle chat message: %v", msg)
	if msg.Username == b.Config.GetString("Nick") {
		return
	}
	bridgeMessage := config.Message{
		Account:  b.Account,
		Text:     msg.Message,
		Channel:  msg.Room,
		Username: msg.Username,
	}
	b.local <- bridgeMessage
}

func (b *Bsoulseek) handleJoinMessage(msg userJoinedRoomMessage) {
	b.Log.Debugf("Handle join message: %v", msg)
	if msg.Username == b.Config.GetString("Nick") {
		return
	}
	bridgeMessage := config.Message{
		Account:  b.Account,
		Event:    config.EventJoinLeave,
		Text:     fmt.Sprintf("%s has joined the room", msg.Username),
		Channel:  msg.Room,
		Username: "system",
	}
	b.local <- bridgeMessage
}

func (b *Bsoulseek) handleLeaveMessage(msg userLeftRoomMessage) {
	b.Log.Debugf("Handle leave message: %v", msg)
	if msg.Username == b.Config.GetString("Nick") {
		return
	}
	bridgeMessage := config.Message{
		Account:  b.Account,
		Event:    config.EventJoinLeave,
		Text:     fmt.Sprintf("%s has left the room", msg.Username),
		Channel:  msg.Room,
		Username: "system",
	}
	b.local <- bridgeMessage
}

func (b *Bsoulseek) handleDM(msg privateMessageReceive) {
	b.Log.Debugf("Received private message: %+v", msg)
	if msg.Username == "server" {
		b.Log.Infof("Received system message: %s", msg.Message)
		if strings.HasPrefix(msg.Message, "System Message: You have been banned") {
			b.Log.Errorf("Banned from server. Message: %s", msg.Message)
			b.doDisconnect()
		}
	}
}
