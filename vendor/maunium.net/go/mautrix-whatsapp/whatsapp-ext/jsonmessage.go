// mautrix-whatsapp - A Matrix-WhatsApp puppeting bridge.
// Copyright (C) 2019 Tulir Asokan
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package whatsappExt

import (
	"encoding/json"

	"github.com/Rhymen/go-whatsapp"
)

type JSONMessage []json.RawMessage

type JSONMessageType string

const (
	MessageMsgInfo  JSONMessageType = "MsgInfo"
	MessageMsg      JSONMessageType = "Msg"
	MessagePresence JSONMessageType = "Presence"
	MessageStream   JSONMessageType = "Stream"
	MessageConn     JSONMessageType = "Conn"
	MessageProps    JSONMessageType = "Props"
	MessageCmd      JSONMessageType = "Cmd"
	MessageChat     JSONMessageType = "Chat"
	MessageCall     JSONMessageType = "Call"
)

func (ext *ExtendedConn) HandleError(error) {}

type UnhandledJSONMessageHandler interface {
	whatsapp.Handler
	HandleUnhandledJSONMessage(string)
}

type JSONParseErrorHandler interface {
	whatsapp.Handler
	HandleJSONParseError(error)
}

func (ext *ExtendedConn) jsonParseError(err error) {
	for _, handler := range ext.handlers {
		errorHandler, ok := handler.(JSONParseErrorHandler)
		if !ok {
			continue
		}
		errorHandler.HandleJSONParseError(err)
	}
}

func (ext *ExtendedConn) HandleJsonMessage(message string) {
	msg := JSONMessage{}
	err := json.Unmarshal([]byte(message), &msg)
	if err != nil || len(msg) < 2 {
		ext.jsonParseError(err)
		return
	}

	var msgType JSONMessageType
	json.Unmarshal(msg[0], &msgType)

	switch msgType {
	case MessagePresence:
		ext.handleMessagePresence(msg[1])
	case MessageStream:
		ext.handleMessageStream(msg[1:])
	case MessageConn:
		ext.handleMessageConn(msg[1])
	case MessageProps:
		ext.handleMessageProps(msg[1])
	case MessageMsgInfo, MessageMsg:
		ext.handleMessageMsgInfo(msgType, msg[1])
	case MessageCmd:
		ext.handleMessageCommand(msg[1])
	case MessageChat:
		ext.handleMessageChatUpdate(msg[1])
	case MessageCall:
		ext.handleMessageCall(msg[1])
	default:
		for _, handler := range ext.handlers {
			ujmHandler, ok := handler.(UnhandledJSONMessageHandler)
			if !ok {
				continue
			}

			if ext.shouldCallSynchronously(ujmHandler) {
				ujmHandler.HandleUnhandledJSONMessage(message)
			} else {
				go ujmHandler.HandleUnhandledJSONMessage(message)
			}
		}
	}
}
