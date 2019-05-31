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

type CallInfoType string

const (
	CallOffer        CallInfoType = "offer"
	CallOfferVideo   CallInfoType = "offer_video"
	CallTransport    CallInfoType = "transport"
	CallRelayLatency CallInfoType = "relaylatency"
	CallTerminate    CallInfoType = "terminate"
)

type CallInfo struct {
	ID   string       `json:"id"`
	Type CallInfoType `json:"type"`
	From string       `json:"from"`

	Platform string `json:"platform"`
	Version  []int  `json:"version"`

	Data [][]interface{} `json:"data"`
}

type CallInfoHandler interface {
	whatsapp.Handler
	HandleCallInfo(CallInfo)
}

func (ext *ExtendedConn) handleMessageCall(message []byte) {
	var event CallInfo
	err := json.Unmarshal(message, &event)
	if err != nil {
		ext.jsonParseError(err)
		return
	}
	event.From = strings.Replace(event.From, OldUserSuffix, NewUserSuffix, 1)
	for _, handler := range ext.handlers {
		callInfoHandler, ok := handler.(CallInfoHandler)
		if !ok {
			continue
		}

		if ext.shouldCallSynchronously(callInfoHandler) {
			callInfoHandler.HandleCallInfo(event)
		} else {
			go callInfoHandler.HandleCallInfo(event)
		}
	}
}
