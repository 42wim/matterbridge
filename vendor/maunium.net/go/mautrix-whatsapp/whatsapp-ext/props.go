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

type ProtocolProps struct {
	WebPresence            bool   `json:"webPresence"`
	NotificationQuery      bool   `json:"notificationQuery"`
	FacebookCrashLog       bool   `json:"fbCrashlog"`
	Bucket                 string `json:"bucket"`
	GIFSearch              string `json:"gifSearch"`
	Spam                   bool   `json:"SPAM"`
	SetBlock               bool   `json:"SET_BLOCK"`
	MessageInfo            bool   `json:"MESSAGE_INFO"`
	MaxFileSize            int    `json:"maxFileSize"`
	Media                  int    `json:"media"`
	GroupNameLength        int    `json:"maxSubject"`
	GroupDescriptionLength int    `json:"groupDescLength"`
	MaxParticipants        int    `json:"maxParticipants"`
	VideoMaxEdge           int    `json:"videoMaxEdge"`
	ImageMaxEdge           int    `json:"imageMaxEdge"`
	ImageMaxKilobytes      int    `json:"imageMaxKBytes"`
	Edit                   int    `json:"edit"`
	FwdUIStartTimestamp    int    `json:"fwdUiStartTs"`
	GroupsV3               int    `json:"groupsV3"`
	RestrictGroups         int    `json:"restrictGroups"`
	AnnounceGroups         int    `json:"announceGroups"`
}

type ProtocolPropsHandler interface {
	whatsapp.Handler
	HandleProtocolProps(ProtocolProps)
}

func (ext *ExtendedConn) handleMessageProps(message []byte) {
	var event ProtocolProps
	err := json.Unmarshal(message, &event)
	if err != nil {
		ext.jsonParseError(err)
		return
	}
	for _, handler := range ext.handlers {
		protocolPropsHandler, ok := handler.(ProtocolPropsHandler)
		if !ok {
			continue
		}

		if ext.shouldCallSynchronously(protocolPropsHandler) {
			protocolPropsHandler.HandleProtocolProps(event)
		} else {
			go protocolPropsHandler.HandleProtocolProps(event)
		}
	}
}
