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
	"github.com/Rhymen/go-whatsapp"
	"github.com/Rhymen/go-whatsapp/binary/proto"
)

type MessageRevokeHandler interface {
	whatsapp.Handler
	HandleMessageRevoke(key MessageRevocation)
}

type MessageRevocation struct {
	Id          string
	RemoteJid   string
	FromMe      bool
	Participant string
}

func (ext *ExtendedConn) HandleRawMessage(message *proto.WebMessageInfo) {
	protoMsg := message.GetMessage().GetProtocolMessage()
	if protoMsg != nil && protoMsg.GetType() == proto.ProtocolMessage_REVOKE {
		key := protoMsg.GetKey()
		deletedMessage := MessageRevocation{
			Id:          key.GetId(),
			RemoteJid:   key.GetRemoteJid(),
			FromMe:      key.GetFromMe(),
			Participant: key.GetParticipant(),
		}
		for _, handler := range ext.handlers {
			mrHandler, ok := handler.(MessageRevokeHandler)
			if !ok {
				continue
			}

			if ext.shouldCallSynchronously(mrHandler) {
				mrHandler.HandleMessageRevoke(deletedMessage)
			} else {
				go mrHandler.HandleMessageRevoke(deletedMessage)
			}
		}
	}
}
