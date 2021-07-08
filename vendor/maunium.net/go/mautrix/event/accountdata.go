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

// TagEventContent represents the content of a m.tag room account data event.
// https://matrix.org/docs/spec/client_server/r0.6.0#m-tag
type TagEventContent struct {
	Tags Tags `json:"tags"`
}

type Tags map[string]Tag

type Tag struct {
	Order json.Number `json:"order,omitempty"`
}

// DirectChatsEventContent represents the content of a m.direct account data event.
// https://matrix.org/docs/spec/client_server/r0.6.0#m-direct
type DirectChatsEventContent map[id.UserID][]id.RoomID

// FullyReadEventContent represents the content of a m.fully_read account data event.
// https://matrix.org/docs/spec/client_server/r0.6.0#m-fully-read
type FullyReadEventContent struct {
	EventID id.EventID `json:"event_id"`
}

// IgnoredUserListEventContent represents the content of a m.ignored_user_list account data event.
// https://matrix.org/docs/spec/client_server/r0.6.0#m-ignored-user-list
type IgnoredUserListEventContent struct {
	IgnoredUsers map[id.UserID]IgnoredUser `json:"ignored_users"`
}

type IgnoredUser struct {
	// This is an empty object
}