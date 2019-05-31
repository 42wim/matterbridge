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
	"strings"

	"github.com/Rhymen/go-whatsapp"
)

type MsgInfoCommand string

const (
	MsgInfoCommandAck  MsgInfoCommand = "ack"
	MsgInfoCommandAcks MsgInfoCommand = "acks"
)

type Acknowledgement int

const (
	AckMessageSent      Acknowledgement = 1
	AckMessageDelivered Acknowledgement = 2
	AckMessageRead      Acknowledgement = 3
)

type JSONStringOrArray []string

func (jsoa *JSONStringOrArray) UnmarshalJSON(data []byte) error {
	var str string
	if json.Unmarshal(data, &str) == nil {
		*jsoa = []string{str}
		return nil
	}
	var strs []string
	json.Unmarshal(data, &strs)
	*jsoa = strs
	return nil
}

type MsgInfo struct {
	Command         MsgInfoCommand    `json:"cmd"`
	IDs             JSONStringOrArray `json:"id"`
	Acknowledgement Acknowledgement   `json:"ack"`
	MessageFromJID  string            `json:"from"`
	SenderJID       string            `json:"participant"`
	ToJID           string            `json:"to"`
	Timestamp       int64             `json:"t"`
}

type MsgInfoHandler interface {
	whatsapp.Handler
	HandleMsgInfo(MsgInfo)
}

func (ext *ExtendedConn) handleMessageMsgInfo(msgType JSONMessageType, message []byte) {
	var event MsgInfo
	err := json.Unmarshal(message, &event)
	if err != nil {
		ext.jsonParseError(err)
		return
	}
	event.MessageFromJID = strings.Replace(event.MessageFromJID, OldUserSuffix, NewUserSuffix, 1)
	event.SenderJID = strings.Replace(event.SenderJID, OldUserSuffix, NewUserSuffix, 1)
	event.ToJID = strings.Replace(event.ToJID, OldUserSuffix, NewUserSuffix, 1)
	if msgType == MessageMsg {
		event.SenderJID = event.ToJID
	}
	for _, handler := range ext.handlers {
		msgInfoHandler, ok := handler.(MsgInfoHandler)
		if !ok {
			continue
		}

		if ext.shouldCallSynchronously(msgInfoHandler) {
			msgInfoHandler.HandleMsgInfo(event)
		} else {
			go msgInfoHandler.HandleMsgInfo(event)
		}
	}
}
