// Copyright (c) 2020 Tulir Asokan
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package event

import (
	"maunium.net/go/mautrix/id"
)

// Event represents a single Matrix event.
type Event struct {
	StateKey  *string    `json:"state_key,omitempty"` // The state key for the event. Only present on State Events.
	Sender    id.UserID  `json:"sender"`              // The user ID of the sender of the event
	Type      Type       `json:"type"`                // The event type
	Timestamp int64      `json:"origin_server_ts"`    // The unix timestamp when this message was sent by the origin server
	ID        id.EventID `json:"event_id"`            // The unique ID of this event
	RoomID    id.RoomID  `json:"room_id"`             // The room the event was sent to. May be nil (e.g. for presence)
	Content   Content    `json:"content"`             // The JSON content of the event.
	Redacts   id.EventID `json:"redacts,omitempty"`   // The event ID that was redacted if a m.room.redaction event
	Unsigned  Unsigned   `json:"unsigned,omitempty"`  // Unsigned content set by own homeserver.

	Mautrix MautrixInfo `json:"-"`
}

type MautrixInfo struct {
	Verified bool
}

func (evt *Event) GetStateKey() string {
	if evt.StateKey != nil {
		return *evt.StateKey
	}
	return ""
}

type StrippedState struct {
	Content  Content `json:"content"`
	Type     Type    `json:"type"`
	StateKey string  `json:"state_key"`
}

type Unsigned struct {
	PrevContent     *Content        `json:"prev_content,omitempty"`
	PrevSender      id.UserID       `json:"prev_sender,omitempty"`
	ReplacesState   id.EventID      `json:"replaces_state,omitempty"`
	Age             int64           `json:"age,omitempty"`
	TransactionID   string          `json:"transaction_id,omitempty"`
	Relations       Relations       `json:"m.relations,omitempty"`
	RedactedBecause *Event          `json:"redacted_because,omitempty"`
	InviteRoomState []StrippedState `json:"invite_room_state"`
}
