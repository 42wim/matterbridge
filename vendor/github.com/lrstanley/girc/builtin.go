// Copyright (c) Liam Stanley <me@liamstanley.io>. All rights reserved. Use
// of this source code is governed by the MIT license that can be found in
// the LICENSE file.

package girc

import (
	"strings"
	"time"
)

// registerBuiltin sets up built-in handlers, based on client
// configuration.
func (c *Client) registerBuiltins() {
	c.debug.Print("registering built-in handlers")
	c.Handlers.mu.Lock()

	// Built-in things that should always be supported.
	c.Handlers.register(true, true, RPL_WELCOME, HandlerFunc(handleConnect))
	c.Handlers.register(true, false, PING, HandlerFunc(handlePING))
	c.Handlers.register(true, false, PONG, HandlerFunc(handlePONG))

	if !c.Config.disableTracking {
		// Joins/parts/anything that may add/remove/rename users.
		c.Handlers.register(true, false, JOIN, HandlerFunc(handleJOIN))
		c.Handlers.register(true, false, PART, HandlerFunc(handlePART))
		c.Handlers.register(true, false, KICK, HandlerFunc(handleKICK))
		c.Handlers.register(true, false, QUIT, HandlerFunc(handleQUIT))
		c.Handlers.register(true, false, NICK, HandlerFunc(handleNICK))
		c.Handlers.register(true, false, RPL_NAMREPLY, HandlerFunc(handleNAMES))

		// Modes.
		c.Handlers.register(true, false, MODE, HandlerFunc(handleMODE))
		c.Handlers.register(true, false, RPL_CHANNELMODEIS, HandlerFunc(handleMODE))

		// WHO/WHOX responses.
		c.Handlers.register(true, false, RPL_WHOREPLY, HandlerFunc(handleWHO))
		c.Handlers.register(true, false, RPL_WHOSPCRPL, HandlerFunc(handleWHO))

		// Other misc. useful stuff.
		c.Handlers.register(true, false, TOPIC, HandlerFunc(handleTOPIC))
		c.Handlers.register(true, false, RPL_TOPIC, HandlerFunc(handleTOPIC))
		c.Handlers.register(true, false, RPL_MYINFO, HandlerFunc(handleMYINFO))
		c.Handlers.register(true, false, RPL_ISUPPORT, HandlerFunc(handleISUPPORT))
		c.Handlers.register(true, false, RPL_MOTDSTART, HandlerFunc(handleMOTD))
		c.Handlers.register(true, false, RPL_MOTD, HandlerFunc(handleMOTD))

		// Keep users lastactive times up to date.
		c.Handlers.register(true, false, PRIVMSG, HandlerFunc(updateLastActive))
		c.Handlers.register(true, false, NOTICE, HandlerFunc(updateLastActive))
		c.Handlers.register(true, false, TOPIC, HandlerFunc(updateLastActive))
		c.Handlers.register(true, false, KICK, HandlerFunc(updateLastActive))

		// CAP IRCv3-specific tracking and functionality.
		c.Handlers.register(true, false, CAP, HandlerFunc(handleCAP))
		c.Handlers.register(true, false, CAP_CHGHOST, HandlerFunc(handleCHGHOST))
		c.Handlers.register(true, false, CAP_AWAY, HandlerFunc(handleAWAY))
		c.Handlers.register(true, false, CAP_ACCOUNT, HandlerFunc(handleACCOUNT))
		c.Handlers.register(true, false, ALL_EVENTS, HandlerFunc(handleTags))

		// SASL IRCv3 support.
		c.Handlers.register(true, false, AUTHENTICATE, HandlerFunc(handleSASL))
		c.Handlers.register(true, false, RPL_SASLSUCCESS, HandlerFunc(handleSASL))
		c.Handlers.register(true, false, RPL_NICKLOCKED, HandlerFunc(handleSASLError))
		c.Handlers.register(true, false, ERR_SASLFAIL, HandlerFunc(handleSASLError))
		c.Handlers.register(true, false, ERR_SASLTOOLONG, HandlerFunc(handleSASLError))
		c.Handlers.register(true, false, ERR_SASLABORTED, HandlerFunc(handleSASLError))
		c.Handlers.register(true, false, RPL_SASLMECHS, HandlerFunc(handleSASLError))
	}

	// Nickname collisions.
	c.Handlers.register(true, false, ERR_NICKNAMEINUSE, HandlerFunc(nickCollisionHandler))
	c.Handlers.register(true, false, ERR_NICKCOLLISION, HandlerFunc(nickCollisionHandler))
	c.Handlers.register(true, false, ERR_UNAVAILRESOURCE, HandlerFunc(nickCollisionHandler))

	c.Handlers.mu.Unlock()
}

// handleConnect is a helper function which lets the client know that enough
// time has passed and now they can send commands.
//
// Should always run in separate thread due to blocking delay.
func handleConnect(c *Client, e Event) {
	// This should be the nick that the server gives us. 99% of the time, it's
	// the one we supplied during connection, but some networks will rename
	// users on connect.
	if len(e.Params) > 0 {
		c.state.Lock()
		c.state.nick = e.Params[0]
		c.state.Unlock()

		c.state.notify(c, UPDATE_GENERAL)
	}

	time.Sleep(2 * time.Second)

	c.mu.RLock()
	server := c.server()
	c.mu.RUnlock()
	c.RunHandlers(&Event{Command: CONNECTED, Params: []string{server}})
}

// nickCollisionHandler helps prevent the client from having conflicting
// nicknames with another bot, user, etc.
func nickCollisionHandler(c *Client, e Event) {
	if c.Config.HandleNickCollide == nil {
		c.Cmd.Nick(c.GetNick() + "_")
		return
	}

	newNick := c.Config.HandleNickCollide(c.GetNick())
	if newNick != "" {
		c.Cmd.Nick(newNick)
	}
}

// handlePING helps respond to ping requests from the server.
func handlePING(c *Client, e Event) {
	c.Cmd.Pong(e.Last())
}

func handlePONG(c *Client, e Event) {
	c.conn.mu.Lock()
	c.conn.lastPong = time.Now()
	c.conn.mu.Unlock()
}

// handleJOIN ensures that the state has updated users and channels.
func handleJOIN(c *Client, e Event) {
	if e.Source == nil || len(e.Params) == 0 {
		return
	}

	channelName := e.Params[0]

	c.state.Lock()

	channel := c.state.lookupChannel(channelName)
	if channel == nil {
		if ok := c.state.createChannel(channelName); !ok {
			c.state.Unlock()
			return
		}

		channel = c.state.lookupChannel(channelName)
	}

	user := c.state.lookupUser(e.Source.Name)
	if user == nil {
		if ok := c.state.createUser(e.Source); !ok {
			c.state.Unlock()
			return
		}
		user = c.state.lookupUser(e.Source.Name)
	}

	defer c.state.notify(c, UPDATE_STATE)

	channel.addUser(user.Nick)
	user.addChannel(channel.Name)

	// Assume extended-join (ircv3).
	if len(e.Params) >= 2 {
		if e.Params[1] != "*" {
			user.Extras.Account = e.Params[1]
		}

		if len(e.Params) > 2 {
			user.Extras.Name = e.Params[2]
		}
	}
	c.state.Unlock()

	if e.Source.ID() == c.GetID() {
		// If it's us, don't just add our user to the list. Run a WHO which
		// will tell us who exactly is in the entire channel.
		c.Send(&Event{Command: WHO, Params: []string{channelName, "%tacuhnr,1"}})

		// Also send a MODE to obtain the list of channel modes.
		c.Send(&Event{Command: MODE, Params: []string{channelName}})

		// Update our ident and host too, in state -- since there is no
		// cleaner method to do this.
		c.state.Lock()
		c.state.ident = e.Source.Ident
		c.state.host = e.Source.Host
		c.state.Unlock()
		return
	}

	// Only WHO the user, which is more efficient.
	c.Send(&Event{Command: WHO, Params: []string{e.Source.Name, "%tacuhnr,1"}})
}

// handlePART ensures that the state is clean of old user and channel entries.
func handlePART(c *Client, e Event) {
	if e.Source == nil || len(e.Params) < 1 {
		return
	}

	// TODO: does this work if it's not the bot?

	channel := e.Params[0]

	if channel == "" {
		return
	}

	defer c.state.notify(c, UPDATE_STATE)

	if e.Source.ID() == c.GetID() {
		c.state.Lock()
		c.state.deleteChannel(channel)
		c.state.Unlock()
		return
	}

	c.state.Lock()
	c.state.deleteUser(channel, e.Source.ID())
	c.state.Unlock()
}

// handleTOPIC handles incoming TOPIC events and keeps channel tracking info
// updated with the latest channel topic.
func handleTOPIC(c *Client, e Event) {
	var name string
	switch len(e.Params) {
	case 0:
		return
	case 1:
		name = e.Params[0]
	default:
		name = e.Params[1]
	}

	c.state.Lock()
	channel := c.state.lookupChannel(name)
	if channel == nil {
		c.state.Unlock()
		return
	}

	channel.Topic = e.Last()
	c.state.Unlock()
	c.state.notify(c, UPDATE_STATE)
}

// handlWHO updates our internal tracking of users/channels with WHO/WHOX
// information.
func handleWHO(c *Client, e Event) {
	var ident, host, nick, account, realname string

	// Assume WHOX related.
	if e.Command == RPL_WHOSPCRPL {
		if len(e.Params) != 8 {
			// Assume there was some form of error or invalid WHOX response.
			return
		}

		if e.Params[1] != "1" {
			// We should always be sending 1, and we should receive 1. If this
			// is anything but, then we didn't send the request and we can
			// ignore it.
			return
		}

		ident, host, nick, account = e.Params[3], e.Params[4], e.Params[5], e.Params[6]
		realname = e.Last()
	} else {
		// Assume RPL_WHOREPLY.
		// format: "<client> <channel> <user> <host> <server> <nick> <H|G>[*][@|+] :<hopcount> <real_name>"
		ident, host, nick, realname = e.Params[2], e.Params[3], e.Params[5], e.Last()

		// Strip the numbers from "<hopcount> <realname>"
		for i := 0; i < len(realname); i++ {
			// Check if it's not 0-9.
			if realname[i] < 0x30 || i > 0x39 {
				realname = strings.TrimLeft(realname[i+1:], " ")
				break
			}

			if i == len(realname)-1 {
				// Assume it's only numbers?
				realname = ""
			}
		}
	}

	c.state.Lock()
	user := c.state.lookupUser(nick)
	if user == nil {
		c.state.Unlock()
		return
	}

	user.Host = host
	user.Ident = ident
	user.Extras.Name = realname

	if account != "0" {
		user.Extras.Account = account
	}

	c.state.Unlock()
	c.state.notify(c, UPDATE_STATE)
}

// handleKICK ensures that users are cleaned up after being kicked from the
// channel
func handleKICK(c *Client, e Event) {
	if len(e.Params) < 2 {
		// Needs at least channel and user.
		return
	}

	defer c.state.notify(c, UPDATE_STATE)

	if e.Params[1] == c.GetNick() {
		c.state.Lock()
		c.state.deleteChannel(e.Params[0])
		c.state.Unlock()
		return
	}

	// Assume it's just another user.
	c.state.Lock()
	c.state.deleteUser(e.Params[0], e.Params[1])
	c.state.Unlock()
}

// handleNICK ensures that users are renamed in state, or the client name is
// up to date.
func handleNICK(c *Client, e Event) {
	if e.Source == nil {
		return
	}

	c.state.Lock()
	// renameUser updates the LastActive time automatically.
	if len(e.Params) >= 1 {
		c.state.renameUser(e.Source.ID(), e.Last())
	}
	c.state.Unlock()
	c.state.notify(c, UPDATE_STATE)
}

// handleQUIT handles users that are quitting from the network.
func handleQUIT(c *Client, e Event) {
	if e.Source == nil {
		return
	}

	if e.Source.ID() == c.GetID() {
		return
	}

	c.state.Lock()
	c.state.deleteUser("", e.Source.ID())
	c.state.Unlock()
	c.state.notify(c, UPDATE_STATE)
}

// handleMYINFO handles incoming MYINFO events -- these are commonly used
// to tell us what the server name is, what version of software is being used
// as well as what channel and user modes are being used on the server.
func handleMYINFO(c *Client, e Event) {
	// Malformed or odd output. As this can differ strongly between networks,
	// just skip it.
	if len(e.Params) < 3 {
		return
	}

	c.state.Lock()
	c.state.serverOptions["SERVER"] = e.Params[1]
	c.state.serverOptions["VERSION"] = e.Params[2]
	c.state.Unlock()
	c.state.notify(c, UPDATE_GENERAL)
}

// handleISUPPORT handles incoming RPL_ISUPPORT (also known as RPL_PROTOCTL)
// events. These commonly contain the server capabilities and limitations.
// For example, things like max channel name length, or nickname length.
func handleISUPPORT(c *Client, e Event) {
	// Must be a ISUPPORT-based message.

	// Also known as RPL_PROTOCTL.
	if !strings.HasSuffix(e.Last(), "this server") {
		return
	}

	// Must have at least one configuration.
	if len(e.Params) < 2 {
		return
	}

	c.state.Lock()
	// Skip the first parameter, as it's our nickname, and the last, as it's the doc.
	for i := 1; i < len(e.Params)-1; i++ {
		j := strings.IndexByte(e.Params[i], '=')

		if j < 1 || (j+1) == len(e.Params[i]) {
			c.state.serverOptions[e.Params[i]] = ""
			continue
		}

		name := e.Params[i][0:j]
		val := e.Params[i][j+1:]
		c.state.serverOptions[name] = val
	}
	c.state.Unlock()

	// Check for max line/nick/user/host lengths here.
	c.state.RLock()
	maxLineLength := c.state.maxLineLength
	c.state.RUnlock()
	maxNickLength := defaultNickLength
	maxUserLength := defaultUserLength
	maxHostLength := defaultHostLength

	var ok bool
	var tmp int

	if tmp, ok = c.GetServerOptionInt("LINELEN"); ok {
		maxLineLength = tmp
		c.state.Lock()
		c.state.maxLineLength = maxTagLength - 2 // -2 for CR-LF.
		c.state.Unlock()
	}

	if tmp, ok = c.GetServerOptionInt("NICKLEN"); ok {
		maxNickLength = tmp
	}
	if tmp, ok = c.GetServerOptionInt("MAXNICKLEN"); ok && tmp > maxNickLength {
		maxNickLength = tmp
	}
	if tmp, ok = c.GetServerOptionInt("USERLEN"); ok && tmp > maxUserLength {
		maxUserLength = tmp
	}
	if tmp, ok = c.GetServerOptionInt("HOSTLEN"); ok && tmp > maxHostLength {
		maxHostLength = tmp
	}

	prefixLen := defaultPrefixPadding + maxNickLength + maxUserLength + maxHostLength
	if prefixLen >= maxLineLength {
		// Give up and go with defaults.
		c.state.notify(c, UPDATE_GENERAL)
		return
	}
	c.state.Lock()
	c.state.maxPrefixLength = prefixLen
	c.state.Unlock()

	c.state.notify(c, UPDATE_GENERAL)
}

// handleMOTD handles incoming MOTD messages and buffers them up for use with
// Client.ServerMOTD().
func handleMOTD(c *Client, e Event) {
	c.state.Lock()

	defer c.state.notify(c, UPDATE_GENERAL)

	// Beginning of the MOTD.
	if e.Command == RPL_MOTDSTART {
		c.state.motd = ""

		c.state.Unlock()
		return
	}

	// Otherwise, assume we're getting sent the MOTD line-by-line.
	if c.state.motd != "" {
		c.state.motd += "\n"
	}
	c.state.motd += e.Last()
	c.state.Unlock()
}

// handleNAMES handles incoming NAMES queries, of which lists all users in
// a given channel. Optionally also obtains ident/host values, as well as
// permissions for each user, depending on what capabilities are enabled.
func handleNAMES(c *Client, e Event) {
	if len(e.Params) < 1 {
		return
	}

	channel := c.state.lookupChannel(e.Params[2])
	if channel == nil {
		return
	}

	parts := strings.Split(e.Last(), " ")

	var modes, nick string
	var ok bool
	var s *Source

	c.state.Lock()
	for i := 0; i < len(parts); i++ {
		modes, nick, ok = parseUserPrefix(parts[i])
		if !ok {
			continue
		}

		// If userhost-in-names.
		if strings.Contains(nick, "@") {
			s = ParseSource(nick)
			if s == nil {
				continue
			}

		} else {
			s = &Source{
				Name: nick,
			}

			if !IsValidNick(s.Name) {
				continue
			}
		}

		c.state.createUser(s)
		user := c.state.lookupUser(s.Name)
		if user == nil {
			continue
		}

		user.addChannel(channel.Name)
		channel.addUser(s.ID())

		// Don't append modes, overwrite them.
		perms, _ := user.Perms.Lookup(channel.Name)
		perms.set(modes, false)
		user.Perms.set(channel.Name, perms)
	}
	c.state.Unlock()
	c.state.notify(c, UPDATE_STATE)
}

// updateLastActive is a wrapper for any event which the source author
// should have it's LastActive time updated. This is useful for things like
// a KICK where we know they are active, as they just kicked another user,
// even though they may not be talking.
func updateLastActive(c *Client, e Event) {
	if e.Source == nil {
		return
	}

	c.state.Lock()

	// Update the users last active time, if they exist.
	user := c.state.lookupUser(e.Source.Name)
	if user == nil {
		c.state.Unlock()
		return
	}

	user.LastActive = time.Now()
	c.state.Unlock()
}
