// Copyright (c) 2020 Tulir Asokan
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package pushrules

import (
	"regexp"
	"strconv"
	"strings"
	"unicode"

	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/pushrules/glob"
)

// Room is an interface with the functions that are needed for processing room-specific push conditions
type Room interface {
	GetOwnDisplayname() string
	GetMemberCount() int
}

// PushCondKind is the type of a push condition.
type PushCondKind string

// The allowed push condition kinds as specified in section 11.12.1.4.3 of r0.3.0 of the Client-Server API.
const (
	KindEventMatch          PushCondKind = "event_match"
	KindContainsDisplayName PushCondKind = "contains_display_name"
	KindRoomMemberCount     PushCondKind = "room_member_count"
)

// PushCondition wraps a condition that is required for a specific PushRule to be used.
type PushCondition struct {
	// The type of the condition.
	Kind PushCondKind `json:"kind"`
	// The dot-separated field of the event to match. Only applicable if kind is EventMatch.
	Key string `json:"key,omitempty"`
	// The glob-style pattern to match the field against. Only applicable if kind is EventMatch.
	Pattern string `json:"pattern,omitempty"`
	// The condition that needs to be fulfilled for RoomMemberCount-type conditions.
	// A decimal integer optionally prefixed by ==, <, >, >= or <=. Prefix "==" is assumed if no prefix found.
	MemberCountCondition string `json:"is,omitempty"`
}

// MemberCountFilterRegex is the regular expression to parse the MemberCountCondition of PushConditions.
var MemberCountFilterRegex = regexp.MustCompile("^(==|[<>]=?)?([0-9]+)$")

// Match checks if this condition is fulfilled for the given event in the given room.
func (cond *PushCondition) Match(room Room, evt *event.Event) bool {
	switch cond.Kind {
	case KindEventMatch:
		return cond.matchValue(room, evt)
	case KindContainsDisplayName:
		return cond.matchDisplayName(room, evt)
	case KindRoomMemberCount:
		return cond.matchMemberCount(room)
	default:
		return false
	}
}

func (cond *PushCondition) matchValue(room Room, evt *event.Event) bool {
	index := strings.IndexRune(cond.Key, '.')
	key := cond.Key
	subkey := ""
	if index > 0 {
		subkey = key[index+1:]
		key = key[0:index]
	}

	pattern, err := glob.Compile(cond.Pattern)
	if err != nil {
		return false
	}

	switch key {
	case "type":
		return pattern.MatchString(evt.Type.String())
	case "sender":
		return pattern.MatchString(string(evt.Sender))
	case "room_id":
		return pattern.MatchString(string(evt.RoomID))
	case "state_key":
		if evt.StateKey == nil {
			return cond.Pattern == ""
		}
		return pattern.MatchString(*evt.StateKey)
	case "content":
		val, _ := evt.Content.Raw[subkey].(string)
		return pattern.MatchString(val)
	default:
		return false
	}
}

func (cond *PushCondition) matchDisplayName(room Room, evt *event.Event) bool {
	displayname := room.GetOwnDisplayname()
	if len(displayname) == 0 {
		return false
	}

	msg, ok := evt.Content.Raw["body"].(string)
	if !ok {
		return false
	}

	isAcceptable := func(r uint8) bool {
		return unicode.IsSpace(rune(r)) || unicode.IsPunct(rune(r))
	}
	length := len(displayname)
	for index := strings.Index(msg, displayname); index != -1; index = strings.Index(msg, displayname) {
		if (index <= 0 || isAcceptable(msg[index-1])) && (index+length >= len(msg) || isAcceptable(msg[index+length])) {
			return true
		}
		msg = msg[index+len(displayname):]
	}
	return false
}

func (cond *PushCondition) matchMemberCount(room Room) bool {
	group := MemberCountFilterRegex.FindStringSubmatch(cond.MemberCountCondition)
	if len(group) != 3 {
		return false
	}

	operator := group[1]
	wantedMemberCount, _ := strconv.Atoi(group[2])

	memberCount := room.GetMemberCount()

	switch operator {
	case "==", "":
		return memberCount == wantedMemberCount
	case ">":
		return memberCount > wantedMemberCount
	case ">=":
		return memberCount >= wantedMemberCount
	case "<":
		return memberCount < wantedMemberCount
	case "<=":
		return memberCount <= wantedMemberCount
	default:
		// Should be impossible due to regex.
		return false
	}
}
