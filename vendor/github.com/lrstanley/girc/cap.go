// Copyright (c) Liam Stanley <me@liamstanley.io>. All rights reserved. Use
// of this source code is governed by the MIT license that can be found in
// the LICENSE file.

package girc

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Something not in the list? Depending on the type of capability, you can
// enable it using Config.SupportedCaps.
var possibleCap = map[string][]string{
	"account-notify":    nil,
	"account-tag":       nil,
	"away-notify":       nil,
	"batch":             nil,
	"cap-notify":        nil,
	"chghost":           nil,
	"extended-join":     nil,
	"invite-notify":     nil,
	"message-tags":      nil,
	"msgid":             nil,
	"multi-prefix":      nil,
	"server-time":       nil,
	"userhost-in-names": nil,

	// Supported draft versions, some may be duplicated above, this is for backwards
	// compatibility.
	"draft/message-tags-0.2": nil,
	"draft/msgid":            nil,

	// sts, sasl, etc are enabled dynamically/depending on client configuration,
	// so aren't included on this list.

	// "echo-message" is supported, but it's not enabled by default. This is
	// to prevent unwanted confusion and utilize less traffic if it's not needed.
	// echo messages aren't sent to girc.PRIVMSG and girc.NOTICE handlers,
	// rather they are only sent to girc.ALL_EVENTS handlers (this is to prevent
	// each handler to have to check these types of things for each message).
	// You can compare events using Event.Equals() to see if they are the same.
}

// https://ircv3.net/specs/extensions/server-time-3.2.html
// <value> ::= YYYY-MM-DDThh:mm:ss.sssZ
const capServerTimeFormat = "2006-01-02T15:04:05.999Z"

func (c *Client) listCAP() {
	if !c.Config.disableTracking {
		c.write(&Event{Command: CAP, Params: []string{CAP_LS, "302"}})
	}
}

func possibleCapList(c *Client) map[string][]string {
	out := make(map[string][]string)

	if c.Config.SASL != nil {
		out["sasl"] = nil
	}

	if !c.Config.DisableSTS && !c.Config.SSL {
		// If fallback supported, and we failed recently, don't try negotiating STS.
		// ONLY do this fallback if we're expired (primarily useful during the first
		// sts negotiation).
		if time.Since(c.state.sts.lastFailed) < 5*time.Minute && !c.Config.DisableSTSFallback {
			c.debug.Println("skipping strict transport policy negotiation; failed within the last 5 minutes")
		} else {
			out["sts"] = nil
		}
	}

	for k := range c.Config.SupportedCaps {
		out[k] = c.Config.SupportedCaps[k]
	}

	for k := range possibleCap {
		out[k] = possibleCap[k]
	}

	return out
}

func parseCap(raw string) map[string]map[string]string {
	out := make(map[string]map[string]string)
	parts := strings.Split(raw, " ")

	var val int

	for i := 0; i < len(parts); i++ {
		val = strings.IndexByte(parts[i], prefixTagValue) // =

		// No value splitter, or has splitter but no trailing value.
		if val < 1 || len(parts[i]) < val+1 {
			// The capability doesn't contain a value.
			out[parts[i]] = nil
			continue
		}

		out[parts[i][:val]] = make(map[string]string)
		for _, option := range strings.Split(parts[i][val+1:], ",") {
			j := strings.Index(option, "=")

			if j < 0 {
				out[parts[i][:val]][option] = ""
			} else {
				out[parts[i][:val]][option[:j]] = option[j+1:]
			}
		}
	}

	return out
}

// handleCAP attempts to find out what IRCv3 capabilities the server supports.
// This will lock further registration until we have acknowledged (or denied)
// the capabilities.
func handleCAP(c *Client, e Event) {
	c.state.Lock()
	defer c.state.Unlock()

	if len(e.Params) >= 2 && e.Params[1] == CAP_DEL {
		caps := parseCap(e.Last())
		for cap := range caps {
			// TODO: test the deletion.
			delete(c.state.enabledCap, cap)
		}
		return
	}

	// We can assume there was a failure attempting to enable a capability.
	if len(e.Params) >= 2 && e.Params[1] == CAP_NAK {
		// Let the server know that we're done.
		c.write(&Event{Command: CAP, Params: []string{CAP_END}})
		return
	}

	possible := possibleCapList(c)
	// TODO: test the addition.
	if len(e.Params) >= 3 && (e.Params[1] == CAP_LS || e.Params[1] == CAP_NEW) {
		caps := parseCap(e.Last())

		for capName := range caps {
			if _, ok := possible[capName]; !ok {
				continue
			}

			if len(possible[capName]) == 0 || len(caps[capName]) == 0 {
				c.state.tmpCap[capName] = caps[capName]
				continue
			}

			var contains bool

			for capAttr := range caps[capName] {
				for i := 0; i < len(possible[capName]); i++ {
					if _, ok := caps[capName][capAttr]; ok {
						// Assuming we have a matching attribute for the capability.
						contains = true
						goto checkcontains
					}
				}
			}

		checkcontains:
			if !contains {
				continue
			}

			c.state.tmpCap[capName] = caps[capName]
		}

		// Indicates if this is a multi-line LS. (3 args means it's the
		// last LS).
		if len(e.Params) == 3 {
			// If we support no caps, just ack the CAP message and END.
			if len(c.state.tmpCap) == 0 {
				c.write(&Event{Command: CAP, Params: []string{CAP_END}})
				return
			}

			// Let them know which ones we'd like to enable.
			reqKeys := make([]string, len(c.state.tmpCap))
			i := 0
			for k := range c.state.tmpCap {
				reqKeys[i] = k
				i++
			}
			c.write(&Event{Command: CAP, Params: []string{CAP_REQ, strings.Join(reqKeys, " ")}})
		}
	}

	if len(e.Params) == 3 && e.Params[1] == CAP_ACK {
		enabled := strings.Split(e.Last(), " ")
		for _, cap := range enabled {
			if val, ok := c.state.tmpCap[cap]; ok {
				c.state.enabledCap[cap] = val
			} else {
				c.state.enabledCap[cap] = nil
			}
		}

		// Anything client side that needs to be setup post-capability-acknowledgement,
		// should be done here.

		// Handle STS, and only if it's something specifically we enabled (client
		// may choose to disable girc automatic STS, and do it themselves).
		if sts, sok := c.state.enabledCap["sts"]; sok && !c.Config.DisableSTS {
			var isError bool

			// Some things are updated in the policy depending on if the current
			// connection is over tls or not.
			var hasTLSConnection bool
			if tlsState, _ := c.TLSConnectionState(); tlsState != nil {
				hasTLSConnection = true
			}

			// "This key indicates the port number for making a secure connection.
			// This key’s value MUST be a single port number. If the client is not
			// already connected securely to the server at the requested hostname,
			// it MUST close the insecure connection and reconnect securely on the
			// stated port.
			//
			// To enforce an STS upgrade policy, servers MUST send this key to
			// insecurely connected clients. Servers MAY send this key to securely
			// connected clients, but it will be ignored."
			//
			// See: https://ircv3.net/specs/extensions/sts#the-port-key
			if !hasTLSConnection {
				if port, ok := sts["port"]; ok {
					c.state.sts.upgradePort, _ = strconv.Atoi(port)
					if c.state.sts.upgradePort < 21 {
						isError = true
					}
				} else {
					isError = true
				}
			}

			// "This key is used on secure connections to indicate how long clients
			// MUST continue to use secure connections when connecting to the server
			// at the requested hostname. The value of this key MUST be given as a
			// single integer which represents the number of seconds until the persistence
			// policy expires.
			//
			// To enforce an STS persistence policy, servers MUST send this key to
			// securely connected clients. Servers MAY send this key to all clients,
			// but insecurely connected clients MUST ignore it."
			//
			// See: https://ircv3.net/specs/extensions/sts#the-duration-key
			if hasTLSConnection {
				if duration, ok := sts["duration"]; ok {
					c.state.sts.persistenceDuration, _ = strconv.Atoi(duration)
					c.state.sts.persistenceReceived = time.Now()
				} else {
					isError = true
				}
			}

			// See: https://ircv3.net/specs/extensions/sts#the-preload-key
			if hasTLSConnection {
				if preload, ok := sts["preload"]; ok {
					c.state.sts.preload, _ = strconv.ParseBool(preload)
				}
			}

			if isError {
				c.rx <- &Event{Command: ERROR, Params: []string{
					fmt.Sprintf("closing connection: strict transport policy provided by server is invalid; possible MITM? config: %#v", sts),
				}}
				return
			}

			// Only upgrade if not already upgraded.
			if !hasTLSConnection {
				c.state.sts.beginUpgrade = true

				c.RunHandlers(&Event{Command: STS_UPGRADE_INIT})
				c.debug.Println("strict transport security policy provided by server; closing connection to begin upgrade...")
				c.Close()
				return
			}
		}

		// Re-initialize the tmpCap, so if we get multiple 'CAP LS' requests
		// due to cap-notify, we can re-evaluate what we can support.
		c.state.tmpCap = make(map[string]map[string]string)

		if _, ok := c.state.enabledCap["sasl"]; ok && c.Config.SASL != nil {
			c.write(&Event{Command: AUTHENTICATE, Params: []string{c.Config.SASL.Method()}})
			// Don't "CAP END", since we want to authenticate.
			return
		}

		// Let the server know that we're done.
		c.write(&Event{Command: CAP, Params: []string{CAP_END}})
		return
	}
}

// handleCHGHOST handles incoming IRCv3 hostname change events. CHGHOST is
// what occurs (when enabled) when a servers services change the hostname of
// a user. Traditionally, this was simply resolved with a quick QUIT and JOIN,
// however CHGHOST resolves this in a much cleaner fashion.
func handleCHGHOST(c *Client, e Event) {
	if len(e.Params) != 2 {
		return
	}

	c.state.Lock()
	user := c.state.lookupUser(e.Source.Name)
	if user != nil {
		user.Ident = e.Params[0]
		user.Host = e.Params[1]
	}
	c.state.Unlock()
	c.state.notify(c, UPDATE_STATE)
}

// handleAWAY handles incoming IRCv3 AWAY events, for which are sent both
// when users are no longer away, or when they are away.
func handleAWAY(c *Client, e Event) {
	c.state.Lock()
	user := c.state.lookupUser(e.Source.Name)
	if user != nil {
		user.Extras.Away = e.Last()
	}
	c.state.Unlock()
	c.state.notify(c, UPDATE_STATE)
}

// handleACCOUNT handles incoming IRCv3 ACCOUNT events. ACCOUNT is sent when
// a user logs into an account, logs out of their account, or logs into a
// different account. The account backend is handled server-side, so this
// could be NickServ, X (undernet?), etc.
func handleACCOUNT(c *Client, e Event) {
	if len(e.Params) != 1 {
		return
	}

	account := e.Params[0]
	if account == "*" {
		account = ""
	}

	c.state.Lock()
	user := c.state.lookupUser(e.Source.Name)
	if user != nil {
		user.Extras.Account = account
	}
	c.state.Unlock()
	c.state.notify(c, UPDATE_STATE)
}
