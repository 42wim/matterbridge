// Copyright (c) 2023 Tulir Asokan
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package mautrix

import (
	"sync"

	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/id"
)

// StateStore is an interface for storing basic room state information.
type StateStore interface {
	IsInRoom(roomID id.RoomID, userID id.UserID) bool
	IsInvited(roomID id.RoomID, userID id.UserID) bool
	IsMembership(roomID id.RoomID, userID id.UserID, allowedMemberships ...event.Membership) bool
	GetMember(roomID id.RoomID, userID id.UserID) *event.MemberEventContent
	TryGetMember(roomID id.RoomID, userID id.UserID) (*event.MemberEventContent, bool)
	SetMembership(roomID id.RoomID, userID id.UserID, membership event.Membership)
	SetMember(roomID id.RoomID, userID id.UserID, member *event.MemberEventContent)

	SetPowerLevels(roomID id.RoomID, levels *event.PowerLevelsEventContent)
	GetPowerLevels(roomID id.RoomID) *event.PowerLevelsEventContent

	SetEncryptionEvent(roomID id.RoomID, content *event.EncryptionEventContent)
	IsEncrypted(roomID id.RoomID) bool
}

func UpdateStateStore(store StateStore, evt *event.Event) {
	if store == nil || evt == nil || evt.StateKey == nil {
		return
	}
	// We only care about events without a state key (power levels, encryption) or member events with state key
	if evt.Type != event.StateMember && evt.GetStateKey() != "" {
		return
	}
	switch content := evt.Content.Parsed.(type) {
	case *event.MemberEventContent:
		store.SetMember(evt.RoomID, id.UserID(evt.GetStateKey()), content)
	case *event.PowerLevelsEventContent:
		store.SetPowerLevels(evt.RoomID, content)
	case *event.EncryptionEventContent:
		store.SetEncryptionEvent(evt.RoomID, content)
	}
}

// StateStoreSyncHandler can be added as an event handler in the syncer to update the state store automatically.
//
//	client.Syncer.(mautrix.ExtensibleSyncer).OnEvent(client.StateStoreSyncHandler)
//
// DefaultSyncer.ParseEventContent must also be true for this to work (which it is by default).
func (cli *Client) StateStoreSyncHandler(_ EventSource, evt *event.Event) {
	UpdateStateStore(cli.StateStore, evt)
}

type MemoryStateStore struct {
	Registrations map[id.UserID]bool                                    `json:"registrations"`
	Members       map[id.RoomID]map[id.UserID]*event.MemberEventContent `json:"memberships"`
	PowerLevels   map[id.RoomID]*event.PowerLevelsEventContent          `json:"power_levels"`
	Encryption    map[id.RoomID]*event.EncryptionEventContent           `json:"encryption"`

	registrationsLock sync.RWMutex
	membersLock       sync.RWMutex
	powerLevelsLock   sync.RWMutex
	encryptionLock    sync.RWMutex
}

func NewMemoryStateStore() StateStore {
	return &MemoryStateStore{
		Registrations: make(map[id.UserID]bool),
		Members:       make(map[id.RoomID]map[id.UserID]*event.MemberEventContent),
		PowerLevels:   make(map[id.RoomID]*event.PowerLevelsEventContent),
	}
}

func (store *MemoryStateStore) IsRegistered(userID id.UserID) bool {
	store.registrationsLock.RLock()
	defer store.registrationsLock.RUnlock()
	registered, ok := store.Registrations[userID]
	return ok && registered
}

func (store *MemoryStateStore) MarkRegistered(userID id.UserID) {
	store.registrationsLock.Lock()
	defer store.registrationsLock.Unlock()
	store.Registrations[userID] = true
}

func (store *MemoryStateStore) GetRoomMembers(roomID id.RoomID) map[id.UserID]*event.MemberEventContent {
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

func (store *MemoryStateStore) GetMembership(roomID id.RoomID, userID id.UserID) event.Membership {
	return store.GetMember(roomID, userID).Membership
}

func (store *MemoryStateStore) GetMember(roomID id.RoomID, userID id.UserID) *event.MemberEventContent {
	member, ok := store.TryGetMember(roomID, userID)
	if !ok {
		member = &event.MemberEventContent{Membership: event.MembershipLeave}
	}
	return member
}

func (store *MemoryStateStore) TryGetMember(roomID id.RoomID, userID id.UserID) (member *event.MemberEventContent, ok bool) {
	store.membersLock.RLock()
	defer store.membersLock.RUnlock()
	members, membersOk := store.Members[roomID]
	if !membersOk {
		return
	}
	member, ok = members[userID]
	return
}

func (store *MemoryStateStore) IsInRoom(roomID id.RoomID, userID id.UserID) bool {
	return store.IsMembership(roomID, userID, "join")
}

func (store *MemoryStateStore) IsInvited(roomID id.RoomID, userID id.UserID) bool {
	return store.IsMembership(roomID, userID, "join", "invite")
}

func (store *MemoryStateStore) IsMembership(roomID id.RoomID, userID id.UserID, allowedMemberships ...event.Membership) bool {
	membership := store.GetMembership(roomID, userID)
	for _, allowedMembership := range allowedMemberships {
		if allowedMembership == membership {
			return true
		}
	}
	return false
}

func (store *MemoryStateStore) SetMembership(roomID id.RoomID, userID id.UserID, membership event.Membership) {
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

func (store *MemoryStateStore) SetMember(roomID id.RoomID, userID id.UserID, member *event.MemberEventContent) {
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

func (store *MemoryStateStore) SetPowerLevels(roomID id.RoomID, levels *event.PowerLevelsEventContent) {
	store.powerLevelsLock.Lock()
	store.PowerLevels[roomID] = levels
	store.powerLevelsLock.Unlock()
}

func (store *MemoryStateStore) GetPowerLevels(roomID id.RoomID) (levels *event.PowerLevelsEventContent) {
	store.powerLevelsLock.RLock()
	levels = store.PowerLevels[roomID]
	store.powerLevelsLock.RUnlock()
	return
}

func (store *MemoryStateStore) GetPowerLevel(roomID id.RoomID, userID id.UserID) int {
	return store.GetPowerLevels(roomID).GetUserLevel(userID)
}

func (store *MemoryStateStore) GetPowerLevelRequirement(roomID id.RoomID, eventType event.Type) int {
	return store.GetPowerLevels(roomID).GetEventLevel(eventType)
}

func (store *MemoryStateStore) HasPowerLevel(roomID id.RoomID, userID id.UserID, eventType event.Type) bool {
	return store.GetPowerLevel(roomID, userID) >= store.GetPowerLevelRequirement(roomID, eventType)
}

func (store *MemoryStateStore) SetEncryptionEvent(roomID id.RoomID, content *event.EncryptionEventContent) {
	store.encryptionLock.Lock()
	store.Encryption[roomID] = content
	store.encryptionLock.Unlock()
}

func (store *MemoryStateStore) GetEncryptionEvent(roomID id.RoomID) *event.EncryptionEventContent {
	store.encryptionLock.RLock()
	defer store.encryptionLock.RUnlock()
	return store.Encryption[roomID]
}

func (store *MemoryStateStore) IsEncrypted(roomID id.RoomID) bool {
	cfg := store.GetEncryptionEvent(roomID)
	return cfg != nil && cfg.Algorithm == id.AlgorithmMegolmV1
}
