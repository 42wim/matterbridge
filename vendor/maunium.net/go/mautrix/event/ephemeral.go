// Copyright (c) 2020 Tulir Asokan
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package event

import (
	"encoding/json"

	"maunium.net/go/mautrix/id"
)

// TagEventContent represents the content of a m.typing ephemeral event.
// https://matrix.org/docs/spec/client_server/r0.6.0#m-typing
type TypingEventContent struct {
	UserIDs []id.UserID `json:"user_ids"`
}

// ReceiptEventContent represents the content of a m.receipt ephemeral event.
// https://matrix.org/docs/spec/client_server/r0.6.0#m-receipt
type ReceiptEventContent map[id.EventID]Receipts

type Receipts struct {
	Read map[id.UserID]ReadReceipt `json:"m.read"`
}

type ReadReceipt struct {
	Timestamp int64 `json:"ts"`

	// Extra contains any unknown fields in the read receipt event.
	// Most servers don't allow clients to set them, so this will be empty in most cases.
	Extra map[string]interface{} `json:"-"`
}

func (rr *ReadReceipt) UnmarshalJSON(data []byte) error {
	// Hacky compatibility hack against crappy clients that send double-encoded read receipts.
	// TODO is this actually needed? clients can't currently set custom content in receipts 🤔
	if data[0] == '"' && data[len(data)-1] == '"' {
		var strData string
		err := json.Unmarshal(data, &strData)
		if err != nil {
			return err
		}
		data = []byte(strData)
	}

	var parsed map[string]interface{}
	err := json.Unmarshal(data, &parsed)
	if err != nil {
		return err
	}
	ts, _ := parsed["ts"].(float64)
	delete(parsed, "ts")
	*rr = ReadReceipt{
		Timestamp: int64(ts),
		Extra:     parsed,
	}
	return nil
}

type Presence string

const (
	PresenceOnline      Presence = "online"
	PresenceOffline     Presence = "offline"
	PresenceUnavailable Presence = "unavailable"
)

// PresenceEventContent represents the content of a m.presence ephemeral event.
// https://matrix.org/docs/spec/client_server/r0.6.0#m-presence
type PresenceEventContent struct {
	Presence        Presence            `json:"presence"`
	Displayname     string              `json:"displayname,omitempty"`
	AvatarURL       id.ContentURIString `json:"avatar_url,omitempty"`
	LastActiveAgo   int64               `json:"last_active_ago,omitempty"`
	CurrentlyActive bool                `json:"currently_active,omitempty"`
	StatusMessage   string              `json:"status_msg,omitempty"`
}
