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

type ConnInfo struct {
	ProtocolVersion []int `json:"protoVersion"`
	BinaryVersion   int   `json:"binVersion"`
	Phone           struct {
		WhatsAppVersion    string `json:"wa_version"`
		MCC                string `json:"mcc"`
		MNC                string `json:"mnc"`
		OSVersion          string `json:"os_version"`
		DeviceManufacturer string `json:"device_manufacturer"`
		DeviceModel        string `json:"device_model"`
		OSBuildNumber      string `json:"os_build_number"`
	} `json:"phone"`
	Features map[string]interface{} `json:"features"`
	PushName string                 `json:"pushname"`
}

type ConnInfoHandler interface {
	whatsapp.Handler
	HandleConnInfo(ConnInfo)
}

func (ext *ExtendedConn) handleMessageConn(message []byte) {
	var event ConnInfo
	err := json.Unmarshal(message, &event)
	if err != nil {
		ext.jsonParseError(err)
		return
	}
	for _, handler := range ext.handlers {
		connInfoHandler, ok := handler.(ConnInfoHandler)
		if !ok {
			continue
		}

		if ext.shouldCallSynchronously(connInfoHandler) {
			connInfoHandler.HandleConnInfo(event)
		} else {
			go connInfoHandler.HandleConnInfo(event)
		}
	}
}
