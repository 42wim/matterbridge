// Copyright (c) 2021 Tulir Asokan
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package event

import (
	"maunium.net/go/mautrix/id"
)

// CanonicalAliasEventContent represents the content of a m.room.canonical_alias state event.
// https://spec.matrix.org/v1.2/client-server-api/#mroomcanonical_alias
type CanonicalAliasEventContent struct {
	Alias      id.RoomAlias   `json:"alias"`
	AltAliases []id.RoomAlias `json:"alt_aliases,omitempty"`
}

// RoomNameEventContent represents the content of a m.room.name state event.
// https://spec.matrix.org/v1.2/client-server-api/#mroomname
type RoomNameEventContent struct {
	Name string `json:"name"`
}

// RoomAvatarEventContent represents the content of a m.room.avatar state event.
// https://spec.matrix.org/v1.2/client-server-api/#mroomavatar
type RoomAvatarEventContent struct {
	URL  id.ContentURI `json:"url"`
	Info *FileInfo     `json:"info,omitempty"`
}

// ServerACLEventContent represents the content of a m.room.server_acl state event.
// https://spec.matrix.org/v1.2/client-server-api/#server-access-control-lists-acls-for-rooms
type ServerACLEventContent struct {
	Allow           []string `json:"allow,omitempty"`
	AllowIPLiterals bool     `json:"allow_ip_literals"`
	Deny            []string `json:"deny,omitempty"`
}

// TopicEventContent represents the content of a m.room.topic state event.
// https://spec.matrix.org/v1.2/client-server-api/#mroomtopic
type TopicEventContent struct {
	Topic string `json:"topic"`
}

// TombstoneEventContent represents the content of a m.room.tombstone state event.
// https://spec.matrix.org/v1.2/client-server-api/#mroomtombstone
type TombstoneEventContent struct {
	Body            string    `json:"body"`
	ReplacementRoom id.RoomID `json:"replacement_room"`
}

type Predecessor struct {
	RoomID  id.RoomID  `json:"room_id"`
	EventID id.EventID `json:"event_id"`
}

// CreateEventContent represents the content of a m.room.create state event.
// https://spec.matrix.org/v1.2/client-server-api/#mroomcreate
type CreateEventContent struct {
	Type        RoomType     `json:"type,omitempty"`
	Creator     id.UserID    `json:"creator,omitempty"`
	Federate    bool         `json:"m.federate,omitempty"`
	RoomVersion string       `json:"room_version,omitempty"`
	Predecessor *Predecessor `json:"predecessor,omitempty"`
}

// JoinRule specifies how open a room is to new members.
// https://spec.matrix.org/v1.2/client-server-api/#mroomjoin_rules
type JoinRule string

const (
	JoinRulePublic     JoinRule = "public"
	JoinRuleKnock      JoinRule = "knock"
	JoinRuleInvite     JoinRule = "invite"
	JoinRuleRestricted JoinRule = "restricted"
	JoinRulePrivate    JoinRule = "private"
)

// JoinRulesEventContent represents the content of a m.room.join_rules state event.
// https://spec.matrix.org/v1.2/client-server-api/#mroomjoin_rules
type JoinRulesEventContent struct {
	JoinRule JoinRule        `json:"join_rule"`
	Allow    []JoinRuleAllow `json:"allow,omitempty"`
}

type JoinRuleAllowType string

const (
	JoinRuleAllowRoomMembership JoinRuleAllowType = "m.room_membership"
)

type JoinRuleAllow struct {
	RoomID id.RoomID         `json:"room_id"`
	Type   JoinRuleAllowType `json:"type"`
}

// PinnedEventsEventContent represents the content of a m.room.pinned_events state event.
// https://spec.matrix.org/v1.2/client-server-api/#mroompinned_events
type PinnedEventsEventContent struct {
	Pinned []id.EventID `json:"pinned"`
}

// HistoryVisibility specifies who can see new messages.
// https://spec.matrix.org/v1.2/client-server-api/#mroomhistory_visibility
type HistoryVisibility string

const (
	HistoryVisibilityInvited       HistoryVisibility = "invited"
	HistoryVisibilityJoined        HistoryVisibility = "joined"
	HistoryVisibilityShared        HistoryVisibility = "shared"
	HistoryVisibilityWorldReadable HistoryVisibility = "world_readable"
)

// HistoryVisibilityEventContent represents the content of a m.room.history_visibility state event.
// https://spec.matrix.org/v1.2/client-server-api/#mroomhistory_visibility
type HistoryVisibilityEventContent struct {
	HistoryVisibility HistoryVisibility `json:"history_visibility"`
}

// GuestAccess specifies whether or not guest accounts can join.
// https://spec.matrix.org/v1.2/client-server-api/#mroomguest_access
type GuestAccess string

const (
	GuestAccessCanJoin   GuestAccess = "can_join"
	GuestAccessForbidden GuestAccess = "forbidden"
)

// GuestAccessEventContent represents the content of a m.room.guest_access state event.
// https://spec.matrix.org/v1.2/client-server-api/#mroomguest_access
type GuestAccessEventContent struct {
	GuestAccess GuestAccess `json:"guest_access"`
}

type BridgeInfoSection struct {
	ID          string              `json:"id"`
	DisplayName string              `json:"displayname,omitempty"`
	AvatarURL   id.ContentURIString `json:"avatar_url,omitempty"`
	ExternalURL string              `json:"external_url,omitempty"`
}

// BridgeEventContent represents the content of a m.bridge state event.
// https://github.com/matrix-org/matrix-doc/pull/2346
type BridgeEventContent struct {
	BridgeBot id.UserID          `json:"bridgebot"`
	Creator   id.UserID          `json:"creator,omitempty"`
	Protocol  BridgeInfoSection  `json:"protocol"`
	Network   *BridgeInfoSection `json:"network,omitempty"`
	Channel   BridgeInfoSection  `json:"channel"`
}

type SpaceChildEventContent struct {
	Via       []string `json:"via,omitempty"`
	Order     string   `json:"order,omitempty"`
	Suggested bool     `json:"suggested,omitempty"`
}

type SpaceParentEventContent struct {
	Via       []string `json:"via,omitempty"`
	Canonical bool     `json:"canonical,omitempty"`
}

// ModPolicyContent represents the content of a m.room.rule.user, m.room.rule.room, and m.room.rule.server state event.
// https://spec.matrix.org/v1.2/client-server-api/#moderation-policy-lists
type ModPolicyContent struct {
	Entity         string `json:"entity"`
	Reason         string `json:"reason"`
	Recommendation string `json:"recommendation"`
}

type InsertionMarkerContent struct {
	InsertionID id.EventID `json:"org.matrix.msc2716.marker.insertion"`
	Timestamp   int64      `json:"com.beeper.timestamp,omitempty"`
}
