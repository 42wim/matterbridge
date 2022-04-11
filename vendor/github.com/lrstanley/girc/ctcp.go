// Copyright (c) Liam Stanley <me@liamstanley.io>. All rights reserved. Use
// of this source code is governed by the MIT license that can be found in
// the LICENSE file.

package girc

import (
	"fmt"
	"runtime"
	"strings"
	"sync"
	"time"
)

// ctcpDelim if the delimiter used for CTCP formatted events/messages.
const ctcpDelim byte = 0x01 // Prefix and suffix for CTCP messages.

// CTCPEvent is the necessary information from an IRC message.
type CTCPEvent struct {
	// Origin is the original event that the CTCP event was decoded from.
	Origin *Event `json:"origin"`
	// Source is the author of the CTCP event.
	Source *Source `json:"source"`
	// Command is the type of CTCP event. E.g. PING, TIME, VERSION.
	Command string `json:"command"`
	// Text is the raw arguments following the command.
	Text string `json:"text"`
	// Reply is true if the CTCP event is intended to be a reply to a
	// previous CTCP (e.g, if we sent one).
	Reply bool `json:"reply"`
}

// DecodeCTCP decodes an incoming CTCP event, if it is CTCP. nil is returned
// if the incoming event does not have valid CTCP encoding.
func DecodeCTCP(e *Event) *CTCPEvent {
	// http://www.irchelp.org/protocol/ctcpspec.html

	if e == nil {
		return nil
	}

	// Must be targeting a user/channel, AND trailing must have
	// DELIM+TAG+DELIM minimum (at least 3 chars).
	if len(e.Params) != 2 || len(e.Params[1]) < 3 {
		return nil
	}

	if e.Command != PRIVMSG && e.Command != NOTICE {
		return nil
	}

	if e.Params[1][0] != ctcpDelim || e.Params[1][len(e.Params[1])-1] != ctcpDelim {
		return nil
	}

	// Strip delimiters.
	text := e.Params[1][1 : len(e.Params[1])-1]

	s := strings.IndexByte(text, eventSpace)

	// Check to see if it only contains a tag.
	if s < 0 {
		for i := 0; i < len(text); i++ {
			// Check for A-Z, 0-9.
			if (text[i] < 'A' || text[i] > 'Z') && (text[i] < '0' || text[i] > '9') {
				return nil
			}
		}

		return &CTCPEvent{
			Origin:  e,
			Source:  e.Source,
			Command: text,
			Reply:   e.Command == NOTICE,
		}
	}

	// Loop through checking the tag first.
	for i := 0; i < s; i++ {
		// Check for A-Z, 0-9.
		if (text[i] < 'A' || text[i] > 'Z') && (text[i] < '0' || text[i] > '9') {
			return nil
		}
	}

	return &CTCPEvent{
		Origin:  e,
		Source:  e.Source,
		Command: text[0:s],
		Text:    text[s+1:],
		Reply:   e.Command == NOTICE,
	}
}

// EncodeCTCP encodes a CTCP event into a string, including delimiters.
func EncodeCTCP(ctcp *CTCPEvent) (out string) {
	if ctcp == nil {
		return ""
	}

	return EncodeCTCPRaw(ctcp.Command, ctcp.Text)
}

// EncodeCTCPRaw is much like EncodeCTCP, however accepts a raw command and
// string as input.
func EncodeCTCPRaw(cmd, text string) (out string) {
	if cmd == "" {
		return ""
	}

	out = string(ctcpDelim) + cmd

	if len(text) > 0 {
		out += string(eventSpace) + text
	}

	return out + string(ctcpDelim)
}

// CTCP handles the storage and execution of CTCP handlers against incoming
// CTCP events.
type CTCP struct {
	// mu is the mutex that should be used when accessing any ctcp handlers.
	mu sync.RWMutex
	// handlers is a map of CTCP message -> functions.
	handlers map[string]CTCPHandler
}

// newCTCP returns a new clean CTCP handler.
func newCTCP() *CTCP {
	return &CTCP{handlers: map[string]CTCPHandler{}}
}

// call executes the necessary CTCP handler for the incoming event/CTCP
// command.
func (c *CTCP) call(client *Client, event *CTCPEvent) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// If they want to catch any panics, add to defer stack.
	if client.Config.RecoverFunc != nil && event.Origin != nil {
		defer recoverHandlerPanic(client, event.Origin, "ctcp-"+strings.ToLower(event.Command), 3)
	}

	// Support wildcard CTCP event handling. Gets executed first before
	// regular event handlers.
	if _, ok := c.handlers["*"]; ok {
		c.handlers["*"](client, *event)
	}

	if _, ok := c.handlers[event.Command]; !ok {
		// If ACTION, don't do anything.
		if event.Command == CTCP_ACTION {
			return
		}

		// Send a ERRMSG reply, if we know who sent it.
		if event.Source != nil && IsValidNick(event.Source.ID()) {
			client.Cmd.SendCTCPReply(event.Source.ID(), CTCP_ERRMSG, "that is an unknown CTCP query")
		}
		return
	}

	c.handlers[event.Command](client, *event)
}

// parseCMD parses a CTCP command/tag, ensuring it's valid. If not, an empty
// string is returned.
func (c *CTCP) parseCMD(cmd string) string {
	// TODO: Needs proper testing.
	// Check if wildcard.
	if cmd == "*" {
		return "*"
	}

	cmd = strings.ToUpper(cmd)

	for i := 0; i < len(cmd); i++ {
		// Check for A-Z, 0-9.
		if (cmd[i] < 'A' || cmd[i] > 'Z') && (cmd[i] < '0' || cmd[i] > '9') {
			return ""
		}
	}

	return cmd
}

// Set saves handler for execution upon a matching incoming CTCP event.
// Use SetBg if the handler may take an extended period of time to execute.
// If you would like to have a handler which will catch ALL CTCP requests,
// simply use "*" in place of the command.
func (c *CTCP) Set(cmd string, handler func(client *Client, ctcp CTCPEvent)) {
	if cmd = c.parseCMD(cmd); cmd == "" {
		return
	}

	c.mu.Lock()
	c.handlers[cmd] = CTCPHandler(handler)
	c.mu.Unlock()
}

// SetBg is much like Set, however the handler is executed in the background,
// ensuring that event handling isn't hung during long running tasks. See Set
// for more information.
func (c *CTCP) SetBg(cmd string, handler func(client *Client, ctcp CTCPEvent)) {
	c.Set(cmd, func(client *Client, ctcp CTCPEvent) {
		go handler(client, ctcp)
	})
}

// Clear removes currently setup handler for cmd, if one is set.
func (c *CTCP) Clear(cmd string) {
	if cmd = c.parseCMD(cmd); cmd == "" {
		return
	}

	c.mu.Lock()
	delete(c.handlers, cmd)
	c.mu.Unlock()
}

// ClearAll removes all currently setup and re-sets the default handlers.
func (c *CTCP) ClearAll() {
	c.mu.Lock()
	c.handlers = map[string]CTCPHandler{}
	c.mu.Unlock()

	// Register necessary handlers.
	c.addDefaultHandlers()
}

// CTCPHandler is a type that represents the function necessary to
// implement a CTCP handler.
type CTCPHandler func(client *Client, ctcp CTCPEvent)

// addDefaultHandlers adds some useful default CTCP response handlers.
func (c *CTCP) addDefaultHandlers() {
	c.SetBg(CTCP_PING, handleCTCPPing)
	c.SetBg(CTCP_PONG, handleCTCPPong)
	c.SetBg(CTCP_VERSION, handleCTCPVersion)
	c.SetBg(CTCP_SOURCE, handleCTCPSource)
	c.SetBg(CTCP_TIME, handleCTCPTime)
	c.SetBg(CTCP_FINGER, handleCTCPFinger)
}

// handleCTCPPing replies with a ping and whatever was originally requested.
func handleCTCPPing(client *Client, ctcp CTCPEvent) {
	if ctcp.Reply {
		return
	}
	client.Cmd.SendCTCPReply(ctcp.Source.ID(), CTCP_PING, ctcp.Text)
}

// handleCTCPPong replies with a pong.
func handleCTCPPong(client *Client, ctcp CTCPEvent) {
	if ctcp.Reply {
		return
	}
	client.Cmd.SendCTCPReply(ctcp.Source.ID(), CTCP_PONG, "")
}

// handleCTCPVersion replies with the name of the client, Go version, as well
// as the os type (darwin, linux, windows, etc) and architecture type (x86,
// arm, etc).
func handleCTCPVersion(client *Client, ctcp CTCPEvent) {
	if client.Config.Version != "" {
		client.Cmd.SendCTCPReply(ctcp.Source.ID(), CTCP_VERSION, client.Config.Version)
		return
	}

	client.Cmd.SendCTCPReplyf(
		ctcp.Source.ID(), CTCP_VERSION,
		"girc (github.com/lrstanley/girc) using %s (%s, %s)",
		runtime.Version(), runtime.GOOS, runtime.GOARCH,
	)
}

// handleCTCPSource replies with the public git location of this library.
func handleCTCPSource(client *Client, ctcp CTCPEvent) {
	client.Cmd.SendCTCPReply(ctcp.Source.ID(), CTCP_SOURCE, "https://github.com/lrstanley/girc")
}

// handleCTCPTime replies with a RFC 1123 (Z) formatted version of Go's
// local time.
func handleCTCPTime(client *Client, ctcp CTCPEvent) {
	client.Cmd.SendCTCPReply(ctcp.Source.ID(), CTCP_TIME, ":"+time.Now().Format(time.RFC1123Z))
}

// handleCTCPFinger replies with the realname and idle time of the user. This
// is obsoleted by improvements to the IRC protocol, however still supported.
func handleCTCPFinger(client *Client, ctcp CTCPEvent) {
	client.conn.mu.RLock()
	active := client.conn.lastActive
	client.conn.mu.RUnlock()

	client.Cmd.SendCTCPReply(ctcp.Source.ID(), CTCP_FINGER, fmt.Sprintf("%s -- idle %s", client.Config.Name, time.Since(active)))
}
