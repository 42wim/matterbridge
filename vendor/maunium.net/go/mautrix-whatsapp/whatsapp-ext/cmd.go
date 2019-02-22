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

type CommandType string

const (
	CommandPicture CommandType = "picture"
)

type Command struct {
	Type CommandType `json:"type"`
	JID  string      `json:"jid"`

	*ProfilePicInfo
}

type CommandHandler interface {
	whatsapp.Handler
	HandleCommand(Command)
}

func (ext *ExtendedConn) handleMessageCommand(message []byte) {
	var event Command
	err := json.Unmarshal(message, &event)
	if err != nil {
		ext.jsonParseError(err)
		return
	}
	event.JID = strings.Replace(event.JID, OldUserSuffix, NewUserSuffix, 1)
	for _, handler := range ext.handlers {
		commandHandler, ok := handler.(CommandHandler)
		if !ok {
			continue
		}
		go commandHandler.HandleCommand(event)
	}
}
