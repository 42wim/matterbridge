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
	"errors"
	"fmt"
	"net/http"
	"time"
)

// VERSION of DiscordGo, follows Semantic Versioning. (http://semver.org/)
const VERSION = "0.18.0"

// ErrMFA will be risen by New when the user has 2FA.
var ErrMFA = errors.New("account has 2FA enabled")

// New creates a new Discord session and will automate some startup
// tasks if given enough information to do so.  Currently you can pass zero
// arguments and it will return an empty Discord session.
// There are 3 ways to call New:
//     With a single auth token - All requests will use the token blindly,
//         no verification of the token will be done and requests may fail.
//         IF THE TOKEN IS FOR A BOT, IT MUST BE PREFIXED WITH `BOT `
//         eg: `"Bot <token>"`
//     With an email and password - Discord will sign in with the provided
//         credentials.
//     With an email, password and auth token - Discord will verify the auth
//         token, if it is invalid it will sign in with the provided
//         credentials. This is the Discord recommended way to sign in.
//
// NOTE: While email/pass authentication is supported by DiscordGo it is
// HIGHLY DISCOURAGED by Discord. Please only use email/pass to obtain a token
// and then use that authentication token for all future connections.
// Also, doing any form of automation with a user (non Bot) account may result
// in that account being permanently banned from Discord.
func New(args ...interface{}) (s *Session, err error) {

	// Create an empty Session interface.
	s = &Session{
		State:                  NewState(),
		Ratelimiter:            NewRatelimiter(),
		StateEnabled:           true,
		Compress:               true,
		ShouldReconnectOnError: true,
		ShardID:                0,
		ShardCount:             1,
		MaxRestRetries:         3,
		Client:                 &http.Client{Timeout: (20 * time.Second)},
		sequence:               new(int64),
		LastHeartbeatAck:       time.Now().UTC(),
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
				err = fmt.Errorf("too many string parameters provided")
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
				err = fmt.Errorf("too many string parameters provided")
				return
			}

			//		case Config:
			// TODO: Parse configuration struct

		default:
			err = fmt.Errorf("unsupported parameter type provided")
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
			if s.MFA {
				err = ErrMFA
			} else {
				err = fmt.Errorf("Unable to fetch discord authentication token. %v", err)
			}
			return
		}
	}

	// The Session is now able to have RestAPI methods called on it.
	// It is recommended that you now call Open() so that events will trigger.

	return
}
