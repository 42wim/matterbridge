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
	c.Handlers.register(true, RPL_WELCOME, HandlerFunc(func(c *Client, e Event) {
		go handleConnect(c, e)
	}))
	c.Handlers.register(true, PING, HandlerFunc(handlePING))
	c.Handlers.register(true, PONG, HandlerFunc(handlePONG))

	if !c.Config.disableTracking {
		// Joins/parts/anything that may add/remove/rename users.
		c.Handlers.register(true, JOIN, HandlerFunc(handleJOIN))
		c.Handlers.register(true, PART, HandlerFunc(handlePART))
		c.Handlers.register(true, KICK, HandlerFunc(handleKICK))
		c.Handlers.register(true, QUIT, HandlerFunc(handleQUIT))
		c.Handlers.register(true, NICK, HandlerFunc(handleNICK))
		c.Handlers.register(true, RPL_NAMREPLY, HandlerFunc(handleNAMES))

		// Modes.
		c.Handlers.register(true, MODE, HandlerFunc(handleMODE))
		c.Handlers.register(true, RPL_CHANNELMODEIS, HandlerFunc(handleMODE))

		// WHO/WHOX responses.
		c.Handlers.register(true, RPL_WHOREPLY, HandlerFunc(handleWHO))
		c.Handlers.register(true, RPL_WHOSPCRPL, HandlerFunc(handleWHO))

		// Other misc. useful stuff.
		c.Handlers.register(true, TOPIC, HandlerFunc(handleTOPIC))
		c.Handlers.register(true, RPL_TOPIC, HandlerFunc(handleTOPIC))
		c.Handlers.register(true, RPL_MYINFO, HandlerFunc(handleMYINFO))
		c.Handlers.register(true, RPL_ISUPPORT, HandlerFunc(handleISUPPORT))
		c.Handlers.register(true, RPL_MOTDSTART, HandlerFunc(handleMOTD))
		c.Handlers.register(true, RPL_MOTD, HandlerFunc(handleMOTD))

		// Keep users lastactive times up to date.
		c.Handlers.register(true, PRIVMSG, HandlerFunc(updateLastActive))
		c.Handlers.register(true, NOTICE, HandlerFunc(updateLastActive))
		c.Handlers.register(true, TOPIC, HandlerFunc(updateLastActive))
		c.Handlers.register(true, KICK, HandlerFunc(updateLastActive))

		// CAP IRCv3-specific tracking and functionality.
		c.Handlers.register(true, CAP, HandlerFunc(handleCAP))
		c.Handlers.register(true, CAP_CHGHOST, HandlerFunc(handleCHGHOST))
		c.Handlers.register(true, CAP_AWAY, HandlerFunc(handleAWAY))
		c.Handlers.register(true, CAP_ACCOUNT, HandlerFunc(handleACCOUNT))
		c.Handlers.register(true, ALL_EVENTS, HandlerFunc(handleTags))

		// SASL IRCv3 support.
		c.Handlers.register(true, AUTHENTICATE, HandlerFunc(handleSASL))
		c.Handlers.register(true, RPL_SASLSUCCESS, HandlerFunc(handleSASL))
		c.Handlers.register(true, RPL_NICKLOCKED, HandlerFunc(handleSASLError))
		c.Handlers.register(true, ERR_SASLFAIL, HandlerFunc(handleSASLError))
		c.Handlers.register(true, ERR_SASLTOOLONG, HandlerFunc(handleSASLError))
		c.Handlers.register(true, ERR_SASLABORTED, HandlerFunc(handleSASLError))
		c.Handlers.register(true, RPL_SASLMECHS, HandlerFunc(handleSASLError))
	}

	// Nickname collisions.
	c.Handlers.register(true, ERR_NICKNAMEINUSE, HandlerFunc(nickCollisionHandler))
	c.Handlers.register(true, ERR_NICKCOLLISION, HandlerFunc(nickCollisionHandler))
	c.Handlers.register(true, ERR_UNAVAILRESOURCE, HandlerFunc(nickCollisionHandler))

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
	c.RunHandlers(&Event{Command: CONNECTED, Trailing: c.Server()})
}

// nickCollisionHandler helps prevent the client from having conflicting
// nicknames with another bot, user, etc.
func nickCollisionHandler(c *Client, e Event) {
	if c.Config.HandleNickCollide == nil {
		c.Cmd.Nick(c.GetNick() + "_")
		return
	}

	c.Cmd.Nick(c.Config.HandleNickCollide(c.GetNick()))
}

// handlePING helps respond to ping requests from the server.
func handlePING(c *Client, e Event) {
	c.Cmd.Pong(e.Trailing)
}

func handlePONG(c *Client, e Event) {
	c.conn.lastPong = time.Now()
}

// handleJOIN ensures that the state has updated users and channels.
func handleJOIN(c *Client, e Event) {
	if e.Source == nil {
		return
	}

	var channelName string
	if len(e.Params) > 0 {
		channelName = e.Params[0]
	} else {
		channelName = e.Trailing
	}

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
		if ok := c.state.createUser(e.Source.Name); !ok {
			c.state.Unlock()
			return
		}
		user = c.state.lookupUser(e.Source.Name)
	}

	defer c.state.notify(c, UPDATE_STATE)

	channel.addUser(user.Nick)
	user.addChannel(channel.Name)

	// Assume extended-join (ircv3).
	if len(e.Params) == 2 {
		if e.Params[1] != "*" {
			user.Extras.Account = e.Params[1]
		}

		if len(e.Trailing) > 0 {
			user.Extras.Name = e.Trailing
		}
	}
	c.state.Unlock()

	if e.Source.Name == c.GetNick() {
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
	if e.Source == nil {
		return
	}

	var channel string
	if len(e.Params) > 0 {
		channel = e.Params[0]
	} else {
		channel = e.Trailing
	}

	if channel == "" {
		return
	}

	defer c.state.notify(c, UPDATE_STATE)

	if e.Source.Name == c.GetNick() {
		c.state.Lock()
		c.state.deleteChannel(channel)
		c.state.Unlock()
		return
	}

	c.state.Lock()
	c.state.deleteUser(channel, e.Source.Name)
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
		name = e.Params[len(e.Params)-1]
	}

	c.state.Lock()
	channel := c.state.lookupChannel(name)
	if channel == nil {
		c.state.Unlock()
		return
	}

	channel.Topic = e.Trailing
	c.state.Unlock()
	c.state.notify(c, UPDATE_STATE)
}

// handlWHO updates our internal tracking of users/channels with WHO/WHOX
// information.
func handleWHO(c *Client, e Event) {
	var ident, host, nick, account, realname string

	// Assume WHOX related.
	if e.Command == RPL_WHOSPCRPL {
		if len(e.Params) != 7 {
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
		realname = e.Trailing
	} else {
		// Assume RPL_WHOREPLY.
		ident, host, nick = e.Params[2], e.Params[3], e.Params[5]
		if len(e.Trailing) > 2 {
			realname = e.Trailing[2:]
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
	if len(e.Params) == 1 {
		c.state.renameUser(e.Source.Name, e.Params[0])
	} else if len(e.Trailing) > 0 {
		c.state.renameUser(e.Source.Name, e.Trailing)
	}
	c.state.Unlock()
	c.state.notify(c, UPDATE_STATE)
}

// handleQUIT handles users that are quitting from the network.
func handleQUIT(c *Client, e Event) {
	if e.Source == nil {
		return
	}

	if e.Source.Name == c.GetNick() {
		return
	}

	c.state.Lock()
	c.state.deleteUser("", e.Source.Name)
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
	// Must be a ISUPPORT-based message. 005 is also used for server bounce
	// related things, so this handler may be triggered during other
	// situations.

	// Also known as RPL_PROTOCTL.
	if !strings.HasSuffix(e.Trailing, "this server") {
		return
	}

	// Must have at least one configuration.
	if len(e.Params) < 2 {
		return
	}

	c.state.Lock()
	// Skip the first parameter, as it's our nickname.
	for i := 1; i < len(e.Params); i++ {
		j := strings.IndexByte(e.Params[i], 0x3D) // =

		if j < 1 || (j+1) == len(e.Params[i]) {
			c.state.serverOptions[e.Params[i]] = ""
			continue
		}

		name := e.Params[i][0:j]
		val := e.Params[i][j+1:]
		c.state.serverOptions[name] = val
	}
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
	if len(c.state.motd) != 0 {
		e.Trailing = "\n" + e.Trailing
	}

	c.state.motd += e.Trailing
	c.state.Unlock()
}

// handleNAMES handles incoming NAMES queries, of which lists all users in
// a given channel. Optionally also obtains ident/host values, as well as
// permissions for each user, depending on what capabilities are enabled.
func handleNAMES(c *Client, e Event) {
	if len(e.Params) < 1 {
		return
	}

	channel := c.state.lookupChannel(e.Params[len(e.Params)-1])
	if channel == nil {
		return
	}

	parts := strings.Split(e.Trailing, " ")

	var host, ident, modes, nick string
	var ok bool

	c.state.Lock()
	for i := 0; i < len(parts); i++ {
		modes, nick, ok = parseUserPrefix(parts[i])
		if !ok {
			continue
		}

		// If userhost-in-names.
		if strings.Contains(nick, "@") {
			s := ParseSource(nick)
			if s == nil {
				continue
			}

			host = s.Host
			nick = s.Name
			ident = s.Ident
		}

		if !IsValidNick(nick) {
			continue
		}

		c.state.createUser(nick)
		user := c.state.lookupUser(nick)
		if user == nil {
			continue
		}

		user.addChannel(channel.Name)
		channel.addUser(nick)

		// Add necessary userhost-in-names data into the user.
		if host != "" {
			user.Host = host
		}
		if ident != "" {
			user.Ident = ident
		}

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
