// Copyright (c) 2020 Tulir Asokan
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package id

import (
	"fmt"
)

// A RoomID is a string starting with ! that references a specific room.
// https://matrix.org/docs/spec/appendices#room-ids-and-event-ids
type RoomID string

// A RoomAlias is a string starting with # that can be resolved into.
// https://matrix.org/docs/spec/appendices#room-aliases
type RoomAlias string

func NewRoomAlias(localpart, server string) RoomAlias {
	return RoomAlias(fmt.Sprintf("#%s:%s", localpart, server))
}

// An EventID is a string starting with $ that references a specific event.
//
// https://matrix.org/docs/spec/appendices#room-ids-and-event-ids
// https://matrix.org/docs/spec/rooms/v4#event-ids
type EventID string

func (roomID RoomID) String() string {
	return string(roomID)
}

func (roomAlias RoomAlias) String() string {
	return string(roomAlias)
}

func (eventID EventID) String() string {
	return string(eventID)
}
