// Copyright (c) 2020 Tulir Asokan
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package mautrix

import (
	"fmt"
	"runtime/debug"
	"time"

	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/id"
)

// EventSource represents the part of the sync response that an event came from.
type EventSource int

const (
	EventSourcePresence EventSource = 1 << iota
	EventSourceJoin
	EventSourceInvite
	EventSourceLeave
	EventSourceAccountData
	EventSourceTimeline
	EventSourceState
	EventSourceEphemeral
	EventSourceToDevice
)

func (es EventSource) String() string {
	switch {
	case es == EventSourcePresence:
		return "presence"
	case es == EventSourceAccountData:
		return "user account data"
	case es == EventSourceToDevice:
		return "to-device"
	case es&EventSourceJoin != 0:
		es -= EventSourceJoin
		switch es {
		case EventSourceState:
			return "joined state"
		case EventSourceTimeline:
			return "joined timeline"
		case EventSourceEphemeral:
			return "room ephemeral (joined)"
		case EventSourceAccountData:
			return "room account data (joined)"
		}
	case es&EventSourceInvite != 0:
		es -= EventSourceInvite
		switch es {
		case EventSourceState:
			return "invited state"
		}
	case es&EventSourceLeave != 0:
		es -= EventSourceLeave
		switch es {
		case EventSourceState:
			return "left state"
		case EventSourceTimeline:
			return "left timeline"
		}
	}
	return fmt.Sprintf("unknown (%d)", es)
}

// EventHandler handles a single event from a sync response.
type EventHandler func(source EventSource, evt *event.Event)

// SyncHandler handles a whole sync response. If the return value is false, handling will be stopped completely.
type SyncHandler func(resp *RespSync, since string) bool

// Syncer is an interface that must be satisfied in order to do /sync requests on a client.
type Syncer interface {
	// Process the /sync response. The since parameter is the since= value that was used to produce the response.
	// This is useful for detecting the very first sync (since=""). If an error is return, Syncing will be stopped
	// permanently.
	ProcessResponse(resp *RespSync, since string) error
	// OnFailedSync returns either the time to wait before retrying or an error to stop syncing permanently.
	OnFailedSync(res *RespSync, err error) (time.Duration, error)
	// GetFilterJSON for the given user ID. NOT the filter ID.
	GetFilterJSON(userID id.UserID) *Filter
}

type ExtensibleSyncer interface {
	OnSync(callback SyncHandler)
	OnEvent(callback EventHandler)
	OnEventType(eventType event.Type, callback EventHandler)
}

// DefaultSyncer is the default syncing implementation. You can either write your own syncer, or selectively
// replace parts of this default syncer (e.g. the ProcessResponse method). The default syncer uses the observer
// pattern to notify callers about incoming events. See DefaultSyncer.OnEventType for more information.
type DefaultSyncer struct {
	// syncListeners want the whole sync response, e.g. the crypto machine
	syncListeners []SyncHandler
	// globalListeners want all events
	globalListeners []EventHandler
	// listeners want a specific event type
	listeners map[event.Type][]EventHandler
	// ParseEventContent determines whether or not event content should be parsed before passing to handlers.
	ParseEventContent bool
	// ParseErrorHandler is called when event.Content.ParseRaw returns an error.
	// If it returns false, the event will not be forwarded to listeners.
	ParseErrorHandler func(evt *event.Event, err error) bool
}

var _ Syncer = (*DefaultSyncer)(nil)
var _ ExtensibleSyncer = (*DefaultSyncer)(nil)

// NewDefaultSyncer returns an instantiated DefaultSyncer
func NewDefaultSyncer() *DefaultSyncer {
	return &DefaultSyncer{
		listeners:         make(map[event.Type][]EventHandler),
		syncListeners:     []SyncHandler{},
		globalListeners:   []EventHandler{},
		ParseEventContent: true,
		ParseErrorHandler: func(evt *event.Event, err error) bool {
			return false
		},
	}
}

// ProcessResponse processes the /sync response in a way suitable for bots. "Suitable for bots" means a stream of
// unrepeating events. Returns a fatal error if a listener panics.
func (s *DefaultSyncer) ProcessResponse(res *RespSync, since string) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("ProcessResponse panicked! since=%s panic=%s\n%s", since, r, debug.Stack())
		}
	}()

	for _, listener := range s.syncListeners {
		if !listener(res, since) {
			return
		}
	}

	s.processSyncEvents("", res.Presence.Events, EventSourcePresence)
	s.processSyncEvents("", res.AccountData.Events, EventSourceAccountData)

	for roomID, roomData := range res.Rooms.Join {
		s.processSyncEvents(roomID, roomData.State.Events, EventSourceJoin|EventSourceState)
		s.processSyncEvents(roomID, roomData.Timeline.Events, EventSourceJoin|EventSourceTimeline)
		s.processSyncEvents(roomID, roomData.Ephemeral.Events, EventSourceJoin|EventSourceEphemeral)
		s.processSyncEvents(roomID, roomData.AccountData.Events, EventSourceJoin|EventSourceAccountData)
	}
	for roomID, roomData := range res.Rooms.Invite {
		s.processSyncEvents(roomID, roomData.State.Events, EventSourceInvite|EventSourceState)
	}
	for roomID, roomData := range res.Rooms.Leave {
		s.processSyncEvents(roomID, roomData.State.Events, EventSourceLeave|EventSourceState)
		s.processSyncEvents(roomID, roomData.Timeline.Events, EventSourceLeave|EventSourceTimeline)
	}
	return
}

func (s *DefaultSyncer) processSyncEvents(roomID id.RoomID, events []*event.Event, source EventSource) {
	for _, evt := range events {
		s.processSyncEvent(roomID, evt, source)
	}
}

func (s *DefaultSyncer) processSyncEvent(roomID id.RoomID, evt *event.Event, source EventSource) {
	evt.RoomID = roomID

	// Ensure the type class is correct. It's safe to mutate the class since the event type is not a pointer.
	// Listeners are keyed by type structs, which means only the correct class will pass.
	switch {
	case evt.StateKey != nil:
		evt.Type.Class = event.StateEventType
	case source == EventSourcePresence, source&EventSourceEphemeral != 0:
		evt.Type.Class = event.EphemeralEventType
	case source&EventSourceAccountData != 0:
		evt.Type.Class = event.AccountDataEventType
	case source == EventSourceToDevice:
		evt.Type.Class = event.ToDeviceEventType
	default:
		evt.Type.Class = event.MessageEventType
	}

	if s.ParseEventContent {
		err := evt.Content.ParseRaw(evt.Type)
		if err != nil && !s.ParseErrorHandler(evt, err) {
			return
		}
	}

	s.notifyListeners(source, evt)
}

func (s *DefaultSyncer) notifyListeners(source EventSource, evt *event.Event) {
	for _, fn := range s.globalListeners {
		fn(source, evt)
	}
	listeners, exists := s.listeners[evt.Type]
	if exists {
		for _, fn := range listeners {
			fn(source, evt)
		}
	}
}

// OnEventType allows callers to be notified when there are new events for the given event type.
// There are no duplicate checks.
func (s *DefaultSyncer) OnEventType(eventType event.Type, callback EventHandler) {
	_, exists := s.listeners[eventType]
	if !exists {
		s.listeners[eventType] = []EventHandler{}
	}
	s.listeners[eventType] = append(s.listeners[eventType], callback)
}

func (s *DefaultSyncer) OnSync(callback SyncHandler) {
	s.syncListeners = append(s.syncListeners, callback)
}

func (s *DefaultSyncer) OnEvent(callback EventHandler) {
	s.globalListeners = append(s.globalListeners, callback)
}

// OnFailedSync always returns a 10 second wait period between failed /syncs, never a fatal error.
func (s *DefaultSyncer) OnFailedSync(res *RespSync, err error) (time.Duration, error) {
	return 10 * time.Second, nil
}

// GetFilterJSON returns a filter with a timeline limit of 50.
func (s *DefaultSyncer) GetFilterJSON(userID id.UserID) *Filter {
	return &Filter{
		Room: RoomFilter{
			Timeline: FilterPart{
				Limit: 50,
			},
		},
	}
}

// OldEventIgnorer is an utility struct for bots to ignore events from before the bot joined the room.
// Create a struct and call Register with your DefaultSyncer to register the sync handler.
type OldEventIgnorer struct {
	UserID id.UserID
}

func (oei *OldEventIgnorer) Register(syncer ExtensibleSyncer) {
	syncer.OnSync(oei.DontProcessOldEvents)
}

// DontProcessOldEvents returns true if a sync response should be processed. May modify the response to remove
// stuff that shouldn't be processed.
func (oei *OldEventIgnorer) DontProcessOldEvents(resp *RespSync, since string) bool {
	if since == "" {
		return false
	}
	// This is a horrible hack because /sync will return the most recent messages for a room
	// as soon as you /join it. We do NOT want to process those events in that particular room
	// because they may have already been processed (if you toggle the bot in/out of the room).
	//
	// Work around this by inspecting each room's timeline and seeing if an m.room.member event for us
	// exists and is "join" and then discard processing that room entirely if so.
	// TODO: We probably want to process messages from after the last join event in the timeline.
	for roomID, roomData := range resp.Rooms.Join {
		for i := len(roomData.Timeline.Events) - 1; i >= 0; i-- {
			evt := roomData.Timeline.Events[i]
			if evt.Type == event.StateMember && evt.GetStateKey() == string(oei.UserID) {
				membership, _ := evt.Content.Raw["membership"].(string)
				if membership == "join" {
					_, ok := resp.Rooms.Join[roomID]
					if !ok {
						continue
					}
					delete(resp.Rooms.Join, roomID)   // don't re-process messages
					delete(resp.Rooms.Invite, roomID) // don't re-process invites
					break
				}
			}
		}
	}
	return true
}
