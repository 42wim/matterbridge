// Copyright 2014 Vic Demuzere
//
// Use of this source code is governed by the MIT license.

package ctcp

// Sources:
// http://www.irchelp.org/irchelp/rfc/ctcpspec.html
// http://www.kvirc.net/doc/doc_ctcp_handling.html

import (
	"fmt"
	"runtime"
	"strings"
	"time"
)

// Various constants used for formatting CTCP messages.
const (
	delimiter byte = 0x01 // Prefix and suffix for CTCP tagged messages.
	space     byte = 0x20 // Token separator

	empty = "" // The empty string

	timeFormat    = time.RFC1123Z
	versionFormat = "Go v%s (" + runtime.GOOS + ", " + runtime.GOARCH + ")"
)

// Tags extracted from the CTCP spec.
const (
	ACTION     = "ACTION"
	PING       = "PING"
	PONG       = "PONG"
	VERSION    = "VERSION"
	USERINFO   = "USERINFO"
	CLIENTINFO = "CLIENTINFO"
	FINGER     = "FINGER"
	SOURCE     = "SOURCE"
	TIME       = "TIME"
)

// Decode attempts to decode CTCP tagged data inside given message text.
//
// If the message text does not contain tagged data, ok will be false.
//
//    <text>  ::= <delim> <tag> [<SPACE> <message>] <delim>
//    <delim> ::= 0x01
//
func Decode(text string) (tag, message string, ok bool) {

	// Fast path, return if this text does not contain a CTCP message.
	if len(text) < 3 || text[0] != delimiter || text[len(text)-1] != delimiter {
		return empty, empty, false
	}

	s := strings.IndexByte(text, space)

	if s < 0 {

		// Messages may contain only a tag.
		return text[1 : len(text)-1], empty, true
	}

	return text[1:s], text[s+1 : len(text)-1], true
}

// Encode returns the IRC message text for CTCP tagged data.
//
//    <text>  ::= <delim> <tag> [<SPACE> <message>] <delim>
//    <delim> ::= 0x01
//
func Encode(tag, message string) (text string) {

	switch {

	// We can't build a valid CTCP tagged message without at least a tag.
	case len(tag) <= 0:
		return empty

	// Tagged data with a message
	case len(message) > 0:
		return string(delimiter) + tag + string(space) + message + string(delimiter)

	// Tagged data without a message
	default:
		return string(delimiter) + tag + string(delimiter)

	}
}

// Action is a shortcut for Encode(ctcp.ACTION, message).
func Action(message string) string {
	return Encode(ACTION, message)
}

// Ping is a shortcut for Encode(ctcp.PING, message).
func Ping(message string) string {
	return Encode(PING, message)
}

// Pong is a shortcut for Encode(ctcp.PONG, message).
func Pong(message string) string {
	return Encode(PONG, message)
}

// Version is a shortcut for Encode(ctcp.VERSION, message).
func Version(message string) string {
	return Encode(VERSION, message)
}

// VersionReply is a shortcut for ENCODE(ctcp.VERSION, go version info).
func VersionReply() string {
	return Encode(VERSION, fmt.Sprintf(versionFormat, runtime.Version()))
}

// UserInfo is a shortcut for Encode(ctcp.USERINFO, message).
func UserInfo(message string) string {
	return Encode(USERINFO, message)
}

// ClientInfo is a shortcut for Encode(ctcp.CLIENTINFO, message).
func ClientInfo(message string) string {
	return Encode(CLIENTINFO, message)
}

// Finger is a shortcut for Encode(ctcp.FINGER, message).
func Finger(message string) string {
	return Encode(FINGER, message)
}

// Source is a shortcut for Encode(ctcp.SOURCE, message).
func Source(message string) string {
	return Encode(SOURCE, message)
}

// Time is a shortcut for Encode(ctcp.TIME, message).
func Time(message string) string {
	return Encode(TIME, message)
}

// TimeReply is a shortcut for Encode(ctcp.TIME, currenttime).
func TimeReply() string {
	return Encode(TIME, time.Now().Format(timeFormat))
}
