// Copyright (c) 2020 Tulir Asokan
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package appservice

import (
	"sync"
	"time"

	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/id"
)

type StateStore interface {
	IsRegistered(userID id.UserID) bool
	MarkRegistered(userID id.UserID)

	IsTyping(roomID id.RoomID, userID id.UserID) bool
	SetTyping(roomID id.RoomID, userID id.UserID, timeout int64)

	IsInRoom(roomID id.RoomID, userID id.UserID) bool
	IsInvited(roomID id.RoomID, userID id.UserID) bool
	IsMembership(roomID id.RoomID, userID id.UserID, allowedMemberships ...event.Membership) bool
	GetMember(roomID id.RoomID, userID id.UserID) *event.MemberEventContent
	TryGetMember(roomID id.RoomID, userID id.UserID) (*event.MemberEventContent, bool)
	SetMembership(roomID id.RoomID, userID id.UserID, membership event.Membership)
	SetMember(roomID id.RoomID, userID id.UserID, member *event.MemberEventContent)

	SetPowerLevels(roomID id.RoomID, levels *event.PowerLevelsEventContent)
	GetPowerLevels(roomID id.RoomID) *event.PowerLevelsEventContent
	GetPowerLevel(roomID id.RoomID, userID id.UserID) int
	GetPowerLevelRequirement(roomID id.RoomID, eventType event.Type) int
	HasPowerLevel(roomID id.RoomID, userID id.UserID, eventType event.Type) bool
}

func (as *AppService) UpdateState(evt *event.Event) {
	switch content := evt.Content.Parsed.(type) {
	case *event.MemberEventContent:
		as.StateStore.SetMember(evt.RoomID, id.UserID(evt.GetStateKey()), content)
	case *event.PowerLevelsEventContent:
		as.StateStore.SetPowerLevels(evt.RoomID, content)
	}
}

type TypingStateStore struct {
	typing     map[id.RoomID]map[id.UserID]int64
	typingLock sync.RWMutex
}

func NewTypingStateStore() *TypingStateStore {
	return &TypingStateStore{
		typing: make(map[id.RoomID]map[id.UserID]int64),
	}
}

func (store *TypingStateStore) IsTyping(roomID id.RoomID, userID id.UserID) bool {
	store.typingLock.RLock()
	defer store.typingLock.RUnlock()
	roomTyping, ok := store.typing[roomID]
	if !ok {
		return false
	}
	typingEndsAt, _ := roomTyping[userID]
	return typingEndsAt >= time.Now().Unix()
}

func (store *TypingStateStore) SetTyping(roomID id.RoomID, userID id.UserID, timeout int64) {
	store.typingLock.Lock()
	defer store.typingLock.Unlock()
	roomTyping, ok := store.typing[roomID]
	if !ok {
		if timeout >= 0 {
			roomTyping = map[id.UserID]int64{
				userID: time.Now().Unix() + timeout,
			}
		} else {
			return
		}
	} else {
		if timeout >= 0 {
			roomTyping[userID] = time.Now().Unix() + timeout
		} else {
			delete(roomTyping, userID)
		}
	}
	store.typing[roomID] = roomTyping
}

type BasicStateStore struct {
	registrationsLock sync.RWMutex                                          `json:"-"`
	Registrations     map[id.UserID]bool                                    `json:"registrations"`
	membersLock       sync.RWMutex                                          `json:"-"`
	Members           map[id.RoomID]map[id.UserID]*event.MemberEventContent `json:"memberships"`
	powerLevelsLock   sync.RWMutex                                          `json:"-"`
	PowerLevels       map[id.RoomID]*event.PowerLevelsEventContent          `json:"power_levels"`

	*TypingStateStore
}

func NewBasicStateStore() StateStore {
	return &BasicStateStore{
		Registrations:    make(map[id.UserID]bool),
		Members:          make(map[id.RoomID]map[id.UserID]*event.MemberEventContent),
		PowerLevels:      make(map[id.RoomID]*event.PowerLevelsEventContent),
		TypingStateStore: NewTypingStateStore(),
	}
}

func (store *BasicStateStore) IsRegistered(userID id.UserID) bool {
	store.registrationsLock.RLock()
	defer store.registrationsLock.RUnlock()
	registered, ok := store.Registrations[userID]
	return ok && registered
}

func (store *BasicStateStore) MarkRegistered(userID id.UserID) {
	store.registrationsLock.Lock()
	defer store.registrationsLock.Unlock()
	store.Registrations[userID] = true
}

func (store *BasicStateStore) GetRoomMembers(roomID id.RoomID) map[id.UserID]*event.MemberEventContent {
	store.membersLock.RLock()
	members, ok := store.Members[roomID]
	store.membersLock.RUnlock()
	if !ok {
		members = make(map[id.UserID]*event.MemberEventContent)
		store.membersLock.Lock()
		store.Members[roomID] = members
		store.membersLock.Unlock()
	}
	return members
}

func (store *BasicStateStore) GetMembership(roomID id.RoomID, userID id.UserID) event.Membership {
	return store.GetMember(roomID, userID).Membership
}

func (store *BasicStateStore) GetMember(roomID id.RoomID, userID id.UserID) *event.MemberEventContent {
	member, ok := store.TryGetMember(roomID, userID)
	if !ok {
		member = &event.MemberEventContent{Membership: event.MembershipLeave}
	}
	return member
}

func (store *BasicStateStore) TryGetMember(roomID id.RoomID, userID id.UserID) (member *event.MemberEventContent, ok bool) {
	store.membersLock.RLock()
	defer store.membersLock.RUnlock()
	members, membersOk := store.Members[roomID]
	if !membersOk {
		return
	}
	member, ok = members[userID]
	return
}

func (store *BasicStateStore) IsInRoom(roomID id.RoomID, userID id.UserID) bool {
	return store.IsMembership(roomID, userID, "join")
}

func (store *BasicStateStore) IsInvited(roomID id.RoomID, userID id.UserID) bool {
	return store.IsMembership(roomID, userID, "join", "invite")
}

func (store *BasicStateStore) IsMembership(roomID id.RoomID, userID id.UserID, allowedMemberships ...event.Membership) bool {
	membership := store.GetMembership(roomID, userID)
	for _, allowedMembership := range allowedMemberships {
		if allowedMembership == membership {
			return true
		}
	}
	return false
}

func (store *BasicStateStore) SetMembership(roomID id.RoomID, userID id.UserID, membership event.Membership) {
	store.membersLock.Lock()
	members, ok := store.Members[roomID]
	if !ok {
		members = map[id.UserID]*event.MemberEventContent{
			userID: {Membership: membership},
		}
	} else {
		member, ok := members[userID]
		if !ok {
			members[userID] = &event.MemberEventContent{Membership: membership}
		} else {
			member.Membership = membership
			members[userID] = member
		}
	}
	store.Members[roomID] = members
	store.membersLock.Unlock()
}

func (store *BasicStateStore) SetMember(roomID id.RoomID, userID id.UserID, member *event.MemberEventContent) {
	store.membersLock.Lock()
	members, ok := store.Members[roomID]
	if !ok {
		members = map[id.UserID]*event.MemberEventContent{
			userID: member,
		}
	} else {
		members[userID] = member
	}
	store.Members[roomID] = members
	store.membersLock.Unlock()
}

func (store *BasicStateStore) SetPowerLevels(roomID id.RoomID, levels *event.PowerLevelsEventContent) {
	store.powerLevelsLock.Lock()
	store.PowerLevels[roomID] = levels
	store.powerLevelsLock.Unlock()
}

func (store *BasicStateStore) GetPowerLevels(roomID id.RoomID) (levels *event.PowerLevelsEventContent) {
	store.powerLevelsLock.RLock()
	levels, _ = store.PowerLevels[roomID]
	store.powerLevelsLock.RUnlock()
	return
}

func (store *BasicStateStore) GetPowerLevel(roomID id.RoomID, userID id.UserID) int {
	return store.GetPowerLevels(roomID).GetUserLevel(userID)
}

func (store *BasicStateStore) GetPowerLevelRequirement(roomID id.RoomID, eventType event.Type) int {
	return store.GetPowerLevels(roomID).GetEventLevel(eventType)
}

func (store *BasicStateStore) HasPowerLevel(roomID id.RoomID, userID id.UserID, eventType event.Type) bool {
	return store.GetPowerLevel(roomID, userID) >= store.GetPowerLevelRequirement(roomID, eventType)
}
