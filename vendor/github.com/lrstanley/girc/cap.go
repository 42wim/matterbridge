// Copyright (c) Liam Stanley <me@liamstanley.io>. All rights reserved. Use
// of this source code is governed by the MIT license that can be found in
// the LICENSE file.

package girc

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"sort"
	"strings"
)

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
	"multi-prefix":      nil,
	"userhost-in-names": nil,
}

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

	for k := range c.Config.SupportedCaps {
		out[k] = c.Config.SupportedCaps[k]
	}

	for k := range possibleCap {
		out[k] = possibleCap[k]
	}

	return out
}

func parseCap(raw string) map[string][]string {
	out := make(map[string][]string)
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

		out[parts[i][:val]] = strings.Split(parts[i][val+1:], ",")
	}

	return out
}

// handleCAP attempts to find out what IRCv3 capabilities the server supports.
// This will lock further registration until we have acknowledged the
// capabilities.
func handleCAP(c *Client, e Event) {
	if len(e.Params) >= 2 && (e.Params[1] == CAP_NEW || e.Params[1] == CAP_DEL) {
		c.listCAP()
		return
	}

	// We can assume there was a failure attempting to enable a capability.
	if len(e.Params) == 2 && e.Params[1] == CAP_NAK {
		// Let the server know that we're done.
		c.write(&Event{Command: CAP, Params: []string{CAP_END}})
		return
	}

	possible := possibleCapList(c)

	if len(e.Params) >= 2 && len(e.Trailing) > 1 && e.Params[1] == CAP_LS {
		c.state.Lock()

		caps := parseCap(e.Trailing)

		for k := range caps {
			if _, ok := possible[k]; !ok {
				continue
			}

			if len(possible[k]) == 0 || len(caps[k]) == 0 {
				c.state.tmpCap = append(c.state.tmpCap, k)
				continue
			}

			var contains bool
			for i := 0; i < len(caps[k]); i++ {
				for j := 0; j < len(possible[k]); j++ {
					if caps[k][i] == possible[k][j] {
						// Assume we have a matching split value.
						contains = true
						goto checkcontains
					}
				}
			}

		checkcontains:
			if !contains {
				continue
			}

			c.state.tmpCap = append(c.state.tmpCap, k)
		}
		c.state.Unlock()

		// Indicates if this is a multi-line LS. (2 args means it's the
		// last LS).
		if len(e.Params) == 2 {
			// If we support no caps, just ack the CAP message and END.
			if len(c.state.tmpCap) == 0 {
				c.write(&Event{Command: CAP, Params: []string{CAP_END}})
				return
			}

			// Let them know which ones we'd like to enable.
			c.write(&Event{Command: CAP, Params: []string{CAP_REQ}, Trailing: strings.Join(c.state.tmpCap, " ")})

			// Re-initialize the tmpCap, so if we get multiple 'CAP LS' requests
			// due to cap-notify, we can re-evaluate what we can support.
			c.state.Lock()
			c.state.tmpCap = []string{}
			c.state.Unlock()
		}
	}

	if len(e.Params) == 2 && len(e.Trailing) > 1 && e.Params[1] == CAP_ACK {
		c.state.Lock()
		c.state.enabledCap = strings.Split(e.Trailing, " ")

		// Do we need to do sasl auth?
		wantsSASL := false
		for i := 0; i < len(c.state.enabledCap); i++ {
			if c.state.enabledCap[i] == "sasl" {
				wantsSASL = true
				break
			}
		}
		c.state.Unlock()

		if wantsSASL {
			c.write(&Event{Command: AUTHENTICATE, Params: []string{c.Config.SASL.Method()}})
			// Don't "CAP END", since we want to authenticate.
			return
		}

		// Let the server know that we're done.
		c.write(&Event{Command: CAP, Params: []string{CAP_END}})
		return
	}
}

// SASLMech is an representation of what a SASL mechanism should support.
// See SASLExternal and SASLPlain for implementations of this.
type SASLMech interface {
	// Method returns the uppercase version of the SASL mechanism name.
	Method() string
	// Encode returns the response that the SASL mechanism wants to use. If
	// the returned string is empty (e.g. the mechanism gives up), the handler
	// will attempt to panic, as expectation is that if SASL authentication
	// fails, the client will disconnect.
	Encode(params []string) (output string)
}

// SASLExternal implements the "EXTERNAL" SASL type.
type SASLExternal struct {
	// Identity is an optional field which allows the client to specify
	// pre-authentication identification. This means that EXTERNAL will
	// supply this in the initial response. This usually isn't needed (e.g.
	// CertFP).
	Identity string `json:"identity"`
}

// Method identifies what type of SASL this implements.
func (sasl *SASLExternal) Method() string {
	return "EXTERNAL"
}

// Encode for external SALS authentication should really only return a "+",
// unless the user has specified pre-authentication or identification data.
// See https://tools.ietf.org/html/rfc4422#appendix-A for more info.
func (sasl *SASLExternal) Encode(params []string) string {
	if len(params) != 1 || params[0] != "+" {
		return ""
	}

	if sasl.Identity != "" {
		return sasl.Identity
	}

	return "+"
}

// SASLPlain contains the user and password needed for PLAIN SASL authentication.
type SASLPlain struct {
	User string `json:"user"` // User is the username for SASL.
	Pass string `json:"pass"` // Pass is the password for SASL.
}

// Method identifies what type of SASL this implements.
func (sasl *SASLPlain) Method() string {
	return "PLAIN"
}

// Encode encodes the plain user+password into a SASL PLAIN implementation.
// See https://tools.ietf.org/rfc/rfc4422.txt for more info.
func (sasl *SASLPlain) Encode(params []string) string {
	if len(params) != 1 || params[0] != "+" {
		return ""
	}

	in := []byte(sasl.User)

	in = append(in, 0x0)
	in = append(in, []byte(sasl.User)...)
	in = append(in, 0x0)
	in = append(in, []byte(sasl.Pass)...)

	return base64.StdEncoding.EncodeToString(in)
}

const saslChunkSize = 400

func handleSASL(c *Client, e Event) {
	if e.Command == RPL_SASLSUCCESS || e.Command == ERR_SASLALREADY {
		// Let the server know that we're done.
		c.write(&Event{Command: CAP, Params: []string{CAP_END}})
		return
	}

	// Assume they want us to handle sending auth.
	auth := c.Config.SASL.Encode(e.Params)

	if auth == "" {
		// Assume the SASL authentication method doesn't want to respond for
		// some reason. The SASL spec and IRCv3 spec do not define a clear
		// way to abort a SASL exchange, other than to disconnect, or proceed
		// with CAP END.
		c.rx <- &Event{Command: ERROR, Trailing: fmt.Sprintf(
			"closing connection: invalid %s SASL configuration provided: %s",
			c.Config.SASL.Method(), e.Trailing,
		)}
		return
	}

	// Send in "saslChunkSize"-length byte chunks. If the last chuck is
	// exactly "saslChunkSize" bytes, send a "AUTHENTICATE +" 0-byte
	// acknowledgement response to let the server know that we're done.
	for {
		if len(auth) > saslChunkSize {
			c.write(&Event{Command: AUTHENTICATE, Params: []string{auth[0 : saslChunkSize-1]}, Sensitive: true})
			auth = auth[saslChunkSize:]
			continue
		}

		if len(auth) <= saslChunkSize {
			c.write(&Event{Command: AUTHENTICATE, Params: []string{auth}, Sensitive: true})

			if len(auth) == 400 {
				c.write(&Event{Command: AUTHENTICATE, Params: []string{"+"}})
			}
			break
		}
	}
	return
}

func handleSASLError(c *Client, e Event) {
	if c.Config.SASL == nil {
		c.write(&Event{Command: CAP, Params: []string{CAP_END}})
		return
	}

	// Authentication failed. The SASL spec and IRCv3 spec do not define a
	// clear way to abort a SASL exchange, other than to disconnect, or
	// proceed with CAP END.
	c.rx <- &Event{Command: ERROR, Trailing: "closing connection: " + e.Trailing}
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
		user.Extras.Away = e.Trailing
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

// handleTags handles any messages that have tags that will affect state. (e.g.
// 'account' tags.)
func handleTags(c *Client, e Event) {
	if len(e.Tags) == 0 {
		return
	}

	account, ok := e.Tags.Get("account")
	if !ok {
		return
	}

	c.state.Lock()
	user := c.state.lookupUser(e.Source.Name)
	if user != nil {
		user.Extras.Account = account
	}
	c.state.Unlock()
	c.state.notify(c, UPDATE_STATE)
}

const (
	prefixTag      byte = 0x40 // @
	prefixTagValue byte = 0x3D // =
	prefixUserTag  byte = 0x2B // +
	tagSeparator   byte = 0x3B // ;
	maxTagLength   int  = 511  // 510 + @ and " " (space), though space usually not included.
)

// Tags represents the key-value pairs in IRCv3 message tags. The map contains
// the encoded message-tag values. If the tag is present, it may still be
// empty. See Tags.Get() and Tags.Set() for use with getting/setting
// information within the tags.
//
// Note that retrieving and setting tags are not concurrent safe. If this is
// necessary, you will need to implement it yourself.
type Tags map[string]string

// ParseTags parses out the key-value map of tags. raw should only be the tag
// data, not a full message. For example:
//   @aaa=bbb;ccc;example.com/ddd=eee
// NOT:
//   @aaa=bbb;ccc;example.com/ddd=eee :nick!ident@host.com PRIVMSG me :Hello
func ParseTags(raw string) (t Tags) {
	t = make(Tags)

	if len(raw) > 0 && raw[0] == prefixTag {
		raw = raw[1:]
	}

	parts := strings.Split(raw, string(tagSeparator))
	var hasValue int

	for i := 0; i < len(parts); i++ {
		hasValue = strings.IndexByte(parts[i], prefixTagValue)

		// The tag doesn't contain a value or has a splitter with no value.
		if hasValue < 1 || len(parts[i]) < hasValue+1 {
			if !validTag(parts[i]) {
				continue
			}

			t[parts[i]] = ""
			continue
		}

		// Check if tag key or decoded value are invalid.
		if !validTag(parts[i][:hasValue]) || !validTagValue(tagDecoder.Replace(parts[i][hasValue+1:])) {
			continue
		}

		t[parts[i][:hasValue]] = parts[i][hasValue+1:]
	}

	return t
}

// Len determines the length of the bytes representation of this tag map. This
// does not include the trailing space required when creating an event, but
// does include the tag prefix ("@").
func (t Tags) Len() (length int) {
	if t == nil {
		return 0
	}

	return len(t.Bytes())
}

// Count finds how many total tags that there are.
func (t Tags) Count() int {
	if t == nil {
		return 0
	}

	return len(t)
}

// Bytes returns a []byte representation of this tag map, including the tag
// prefix ("@"). Note that this will return the tags sorted, regardless of
// the order of how they were originally parsed.
func (t Tags) Bytes() []byte {
	if t == nil {
		return []byte{}
	}

	max := len(t)
	if max == 0 {
		return nil
	}

	buffer := new(bytes.Buffer)
	buffer.WriteByte(prefixTag)

	var current int

	// Sort the writing of tags so we can at least guarantee that they will
	// be in order, and testable.
	var names []string
	for tagName := range t {
		names = append(names, tagName)
	}
	sort.Strings(names)

	for i := 0; i < len(names); i++ {
		// Trim at max allowed chars.
		if (buffer.Len() + len(names[i]) + len(t[names[i]]) + 2) > maxTagLength {
			return buffer.Bytes()
		}

		buffer.WriteString(names[i])

		// Write the value as necessary.
		if len(t[names[i]]) > 0 {
			buffer.WriteByte(prefixTagValue)
			buffer.WriteString(t[names[i]])
		}

		// add the separator ";" between tags.
		if current < max-1 {
			buffer.WriteByte(tagSeparator)
		}

		current++
	}

	return buffer.Bytes()
}

// String returns a string representation of this tag map.
func (t Tags) String() string {
	if t == nil {
		return ""
	}

	return string(t.Bytes())
}

// writeTo writes the necessary tag bytes to an io.Writer, including a trailing
// space-separator.
func (t Tags) writeTo(w io.Writer) (n int, err error) {
	b := t.Bytes()
	if len(b) == 0 {
		return n, err
	}

	n, err = w.Write(b)
	if err != nil {
		return n, err
	}

	var j int
	j, err = w.Write([]byte{eventSpace})
	n += j

	return n, err
}

// tagDecode are encoded -> decoded pairs for replacement to decode.
var tagDecode = []string{
	"\\:", ";",
	"\\s", " ",
	"\\\\", "\\",
	"\\r", "\r",
	"\\n", "\n",
}
var tagDecoder = strings.NewReplacer(tagDecode...)

// tagEncode are decoded -> encoded pairs for replacement to decode.
var tagEncode = []string{
	";", "\\:",
	" ", "\\s",
	"\\", "\\\\",
	"\r", "\\r",
	"\n", "\\n",
}
var tagEncoder = strings.NewReplacer(tagEncode...)

// Get returns the unescaped value of given tag key. Note that this is not
// concurrent safe.
func (t Tags) Get(key string) (tag string, success bool) {
	if t == nil {
		return "", false
	}

	if _, ok := t[key]; ok {
		tag = tagDecoder.Replace(t[key])
		success = true
	}

	return tag, success
}

// Set escapes given value and saves it as the value for given key. Note that
// this is not concurrent safe.
func (t Tags) Set(key, value string) error {
	if t == nil {
		t = make(Tags)
	}

	if !validTag(key) {
		return fmt.Errorf("tag key %q is invalid", key)
	}

	value = tagEncoder.Replace(value)

	if len(value) > 0 && !validTagValue(value) {
		return fmt.Errorf("tag value %q of key %q is invalid", value, key)
	}

	// Check to make sure it's not too long here.
	if (t.Len() + len(key) + len(value) + 2) > maxTagLength {
		return fmt.Errorf("unable to set tag %q [value %q]: tags too long for message", key, value)
	}

	t[key] = value

	return nil
}

// Remove deletes the tag frwom the tag map.
func (t Tags) Remove(key string) (success bool) {
	if t == nil {
		return false
	}

	if _, success = t[key]; success {
		delete(t, key)
	}

	return success
}

// validTag validates an IRC tag.
func validTag(name string) bool {
	if len(name) < 1 {
		return false
	}

	// Allow user tags to be passed to validTag.
	if len(name) >= 2 && name[0] == prefixUserTag {
		name = name[1:]
	}

	for i := 0; i < len(name); i++ {
		// A-Z, a-z, 0-9, -/._
		if (name[i] < 0x41 || name[i] > 0x5A) && (name[i] < 0x61 || name[i] > 0x7A) && (name[i] < 0x2D || name[i] > 0x39) && name[i] != 0x5F {
			return false
		}
	}

	return true
}

// validTagValue valids a decoded IRC tag value. If the value is not decoded
// with tagDecoder first, it may be seen as invalid.
func validTagValue(value string) bool {
	for i := 0; i < len(value); i++ {
		// Don't allow any invisible chars within the tag, or semicolons.
		if value[i] < 0x21 || value[i] > 0x7E || value[i] == 0x3B {
			return false
		}
	}
	return true
}
