// Discordgo - Discord bindings for Go
// Available at https://github.com/bwmarrin/discordgo

// Copyright 2015-2016 Bruce Marriner <bruce@sqls.net>.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This file contains high level helper functions and easy entry points for the
// entire discordgo package.  These functions are beling developed and are very
// experimental at this point.  They will most likley change so please use the
// low level functions if that's a problem.

// Package discordgo provides Discord binding for Go
package discordgo

import (
	"fmt"
	"reflect"
)

// VERSION of Discordgo, follows Symantic Versioning. (http://semver.org/)
const VERSION = "0.13.0"

// New creates a new Discord session and will automate some startup
// tasks if given enough information to do so.  Currently you can pass zero
// arguments and it will return an empty Discord session.
// There are 3 ways to call New:
//     With a single auth token - All requests will use the token blindly,
//         no verification of the token will be done and requests may fail.
//     With an email and password - Discord will sign in with the provided
//         credentials.
//     With an email, password and auth token - Discord will verify the auth
//         token, if it is invalid it will sign in with the provided
//         credentials. This is the Discord recommended way to sign in.
func New(args ...interface{}) (s *Session, err error) {

	// Create an empty Session interface.
	s = &Session{
		State:                  NewState(),
		StateEnabled:           true,
		Compress:               true,
		ShouldReconnectOnError: true,
		ShardID:                0,
		ShardCount:             1,
	}

	// If no arguments are passed return the empty Session interface.
	if args == nil {
		return
	}

	// Variables used below when parsing func arguments
	var auth, pass string

	// Parse passed arguments
	for _, arg := range args {

		switch v := arg.(type) {

		case []string:
			if len(v) > 3 {
				err = fmt.Errorf("Too many string parameters provided.")
				return
			}

			// First string is either token or username
			if len(v) > 0 {
				auth = v[0]
			}

			// If second string exists, it must be a password.
			if len(v) > 1 {
				pass = v[1]
			}

			// If third string exists, it must be an auth token.
			if len(v) > 2 {
				s.Token = v[2]
			}

		case string:
			// First string must be either auth token or username.
			// Second string must be a password.
			// Only 2 input strings are supported.

			if auth == "" {
				auth = v
			} else if pass == "" {
				pass = v
			} else if s.Token == "" {
				s.Token = v
			} else {
				err = fmt.Errorf("Too many string parameters provided.")
				return
			}

			//		case Config:
			// TODO: Parse configuration struct

		default:
			err = fmt.Errorf("Unsupported parameter type provided.")
			return
		}
	}

	// If only one string was provided, assume it is an auth token.
	// Otherwise get auth token from Discord, if a token was specified
	// Discord will verify it for free, or log the user in if it is
	// invalid.
	if pass == "" {
		s.Token = auth
	} else {
		err = s.Login(auth, pass)
		if err != nil || s.Token == "" {
			err = fmt.Errorf("Unable to fetch discord authentication token. %v", err)
			return
		}
	}

	// The Session is now able to have RestAPI methods called on it.
	// It is recommended that you now call Open() so that events will trigger.

	return
}

// validateHandler takes an event handler func, and returns the type of event.
// eg.
//     Session.validateHandler(func (s *discordgo.Session, m *discordgo.MessageCreate))
//     will return the reflect.Type of *discordgo.MessageCreate
func (s *Session) validateHandler(handler interface{}) reflect.Type {

	handlerType := reflect.TypeOf(handler)

	if handlerType.NumIn() != 2 {
		panic("Unable to add event handler, handler must be of the type func(*discordgo.Session, *discordgo.EventType).")
	}

	if handlerType.In(0) != reflect.TypeOf(s) {
		panic("Unable to add event handler, first argument must be of type *discordgo.Session.")
	}

	eventType := handlerType.In(1)

	// Support handlers of type interface{}, this is a special handler, which is triggered on every event.
	if eventType.Kind() == reflect.Interface {
		eventType = nil
	}

	return eventType
}

// AddHandler allows you to add an event handler that will be fired anytime
// the Discord WSAPI event that matches the interface fires.
// eventToInterface in events.go has a list of all the Discord WSAPI events
// and their respective interface.
// eg:
//     Session.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
//     })
//
// or:
//     Session.AddHandler(func(s *discordgo.Session, m *discordgo.PresenceUpdate) {
//     })
// The return value of this method is a function, that when called will remove the
// event handler.
func (s *Session) AddHandler(handler interface{}) func() {

	s.initialize()

	eventType := s.validateHandler(handler)

	s.handlersMu.Lock()
	defer s.handlersMu.Unlock()

	h := reflect.ValueOf(handler)

	s.handlers[eventType] = append(s.handlers[eventType], h)

	// This must be done as we need a consistent reference to the
	// reflected value, otherwise a RemoveHandler method would have
	// been nice.
	return func() {
		s.handlersMu.Lock()
		defer s.handlersMu.Unlock()

		handlers := s.handlers[eventType]
		for i, v := range handlers {
			if h == v {
				s.handlers[eventType] = append(handlers[:i], handlers[i+1:]...)
				return
			}
		}
	}
}

// handle calls any handlers that match the event type and any handlers of
// interface{}.
func (s *Session) handle(event interface{}) {

	s.handlersMu.RLock()
	defer s.handlersMu.RUnlock()

	if s.handlers == nil {
		return
	}

	handlerParameters := []reflect.Value{reflect.ValueOf(s), reflect.ValueOf(event)}

	if handlers, ok := s.handlers[nil]; ok {
		for _, handler := range handlers {
			go handler.Call(handlerParameters)
		}
	}

	if handlers, ok := s.handlers[reflect.TypeOf(event)]; ok {
		for _, handler := range handlers {
			go handler.Call(handlerParameters)
		}
	}
}

// initialize adds all internal handlers and state tracking handlers.
func (s *Session) initialize() {

	s.log(LogInformational, "called")

	s.handlersMu.Lock()
	if s.handlers != nil {
		s.handlersMu.Unlock()
		return
	}

	s.handlers = map[interface{}][]reflect.Value{}
	s.handlersMu.Unlock()

	s.AddHandler(s.onReady)
	s.AddHandler(s.onResumed)
	s.AddHandler(s.onVoiceServerUpdate)
	s.AddHandler(s.onVoiceStateUpdate)
	s.AddHandler(s.State.onInterface)
}

// onReady handles the ready event.
func (s *Session) onReady(se *Session, r *Ready) {

	// Store the SessionID within the Session struct.
	s.sessionID = r.SessionID

	// Start the heartbeat to keep the connection alive.
	go s.heartbeat(s.wsConn, s.listening, r.HeartbeatInterval)
}

// onResumed handles the resumed event.
func (s *Session) onResumed(se *Session, r *Resumed) {

	// Start the heartbeat to keep the connection alive.
	go s.heartbeat(s.wsConn, s.listening, r.HeartbeatInterval)
}
