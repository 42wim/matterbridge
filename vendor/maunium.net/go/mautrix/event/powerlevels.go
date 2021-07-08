// Copyright (c) 2020 Tulir Asokan
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package event

import (
	"sync"

	"maunium.net/go/mautrix/id"
)

// PowerLevelsEventContent represents the content of a m.room.power_levels state event content.
// https://matrix.org/docs/spec/client_server/r0.6.0#m-room-power-levels
type PowerLevelsEventContent struct {
	usersLock    sync.RWMutex      `json:"-"`
	Users        map[id.UserID]int `json:"users"`
	UsersDefault int               `json:"users_default"`

	eventsLock    sync.RWMutex   `json:"-"`
	Events        map[string]int `json:"events"`
	EventsDefault int            `json:"events_default"`

	StateDefaultPtr *int `json:"state_default,omitempty"`

	InvitePtr *int `json:"invite,omitempty"`
	KickPtr   *int `json:"kick,omitempty"`
	BanPtr    *int `json:"ban,omitempty"`
	RedactPtr *int `json:"redact,omitempty"`
}

func (pl *PowerLevelsEventContent) Invite() int {
	if pl.InvitePtr != nil {
		return *pl.InvitePtr
	}
	return 50
}

func (pl *PowerLevelsEventContent) Kick() int {
	if pl.KickPtr != nil {
		return *pl.KickPtr
	}
	return 50
}

func (pl *PowerLevelsEventContent) Ban() int {
	if pl.BanPtr != nil {
		return *pl.BanPtr
	}
	return 50
}

func (pl *PowerLevelsEventContent) Redact() int {
	if pl.RedactPtr != nil {
		return *pl.RedactPtr
	}
	return 50
}

func (pl *PowerLevelsEventContent) StateDefault() int {
	if pl.StateDefaultPtr != nil {
		return *pl.StateDefaultPtr
	}
	return 50
}

func (pl *PowerLevelsEventContent) GetUserLevel(userID id.UserID) int {
	pl.usersLock.RLock()
	defer pl.usersLock.RUnlock()
	level, ok := pl.Users[userID]
	if !ok {
		return pl.UsersDefault
	}
	return level
}

func (pl *PowerLevelsEventContent) SetUserLevel(userID id.UserID, level int) {
	pl.usersLock.Lock()
	defer pl.usersLock.Unlock()
	if level == pl.UsersDefault {
		delete(pl.Users, userID)
	} else {
		pl.Users[userID] = level
	}
}

func (pl *PowerLevelsEventContent) EnsureUserLevel(userID id.UserID, level int) bool {
	existingLevel := pl.GetUserLevel(userID)
	if existingLevel != level {
		pl.SetUserLevel(userID, level)
		return true
	}
	return false
}

func (pl *PowerLevelsEventContent) GetEventLevel(eventType Type) int {
	pl.eventsLock.RLock()
	defer pl.eventsLock.RUnlock()
	level, ok := pl.Events[eventType.String()]
	if !ok {
		if eventType.IsState() {
			return pl.StateDefault()
		}
		return pl.EventsDefault
	}
	return level
}

func (pl *PowerLevelsEventContent) SetEventLevel(eventType Type, level int) {
	pl.eventsLock.Lock()
	defer pl.eventsLock.Unlock()
	if (eventType.IsState() && level == pl.StateDefault()) || (!eventType.IsState() && level == pl.EventsDefault) {
		delete(pl.Events, eventType.String())
	} else {
		pl.Events[eventType.String()] = level
	}
}

func (pl *PowerLevelsEventContent) EnsureEventLevel(eventType Type, level int) bool {
	existingLevel := pl.GetEventLevel(eventType)
	if existingLevel != level {
		pl.SetEventLevel(eventType, level)
		return true
	}
	return false
}
