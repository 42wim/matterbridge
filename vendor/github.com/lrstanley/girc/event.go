// Copyright (c) Liam Stanley <me@liamstanley.io>. All rights reserved. Use
// of this source code is governed by the MIT license that can be found in
// the LICENSE file.

package girc

import (
	"bytes"
	"fmt"
	"strings"
	"time"
)

const (
	eventSpace byte = ' ' // Separator.
	maxLength  int  = 510 // Maximum length is 510 (2 for line endings).
)

// cutCRFunc is used to trim CR characters from prefixes/messages.
func cutCRFunc(r rune) bool {
	return r == '\r' || r == '\n'
}

// ParseEvent takes a string and attempts to create a Event struct. Returns
// nil if the Event is invalid.
func ParseEvent(raw string) (e *Event) {
	// Ignore empty events.
	if raw = strings.TrimFunc(raw, cutCRFunc); len(raw) < 2 {
		return nil
	}

	var i, j int
	e = &Event{Timestamp: time.Now()}

	if raw[0] == prefixTag {
		// Tags end with a space.
		i = strings.IndexByte(raw, eventSpace)

		if i < 2 {
			return nil
		}

		e.Tags = ParseTags(raw[1:i])
		if rawServerTime, ok := e.Tags.Get("time"); ok {
			// Attempt to parse server-time. If we can't parse it, we just
			// fall back to the time we received the message (locally.)
			if stime, err := time.Parse(capServerTimeFormat, rawServerTime); err == nil {
				e.Timestamp = stime.Local()
			}
		}
		raw = raw[i+1:]
		i = 0
	}

	if raw != "" && raw[0] == messagePrefix {
		// Prefix ends with a space.
		i = strings.IndexByte(raw, eventSpace)

		// Prefix string must not be empty if the indicator is present.
		if i < 2 {
			return nil
		}

		e.Source = ParseSource(raw[1:i])

		// Skip space at the end of the prefix.
		i++
	}

	// Find end of command.
	j = i + strings.IndexByte(raw[i:], eventSpace)

	if j < i {
		// If there are no proceeding spaces, it's the only thing specified.
		e.Command = strings.ToUpper(raw[i:])
		return e
	}

	e.Command = strings.ToUpper(raw[i:j])

	// Skip the space after the command.
	j++

	// Check if and where the trailing text is within the incoming line.
	var lastIndex, trailerIndex int
	for {
		// We must loop through, as it's possible that the first message
		// prefix is not actually what we want. (e.g, colons are commonly
		// used within ISUPPORT to delegate things like CHANLIMIT or TARGMAX.)
		lastIndex = trailerIndex
		trailerIndex = strings.IndexByte(raw[j+lastIndex:], messagePrefix)

		if trailerIndex == -1 {
			// No trailing argument found, assume the rest is just params.
			e.Params = strings.Fields(raw[j:])
			return e
		}

		// This means we found a prefix that was proceeded by a space, and
		// it's good to assume this is the start of trailing text to the line.
		if raw[j+lastIndex+trailerIndex-1] == eventSpace {
			i = lastIndex + trailerIndex
			break
		}

		// Keep looping through until we either can't find any more prefixes,
		// or we find the one we want.
		trailerIndex += lastIndex + 1
	}

	// Set i to that of the substring we were using before, and where the
	// trailing prefix is.
	i = j + i

	// Check if we need to parse arguments. If so, take everything after the
	// command, and right before the trailing prefix, and cut it up.
	if i > j {
		e.Params = strings.Fields(raw[j : i-1])
	}

	e.Params = append(e.Params, raw[i+1:])

	return e
}

// Event represents an IRC protocol message, see RFC1459 section 2.3.1
//
//    <message>  :: [':' <prefix> <SPACE>] <command> <params> <crlf>
//    <prefix>   :: <servername> | <nick> ['!' <user>] ['@' <host>]
//    <command>  :: <letter>{<letter>} | <number> <number> <number>
//    <SPACE>    :: ' '{' '}
//    <params>   :: <SPACE> [':' <trailing> | <middle> <params>]
//    <middle>   :: <Any *non-empty* sequence of octets not including SPACE or NUL
//                   or CR or LF, the first of which may not be ':'>
//    <trailing> :: <Any, possibly empty, sequence of octets not including NUL or
//                   CR or LF>
//    <crlf>     :: CR LF
type Event struct {
	// Source is the origin of the event.
	Source *Source `json:"source"`
	// Tags are the IRCv3 style message tags for the given event. Only use
	// if network supported.
	Tags Tags `json:"tags"`
	// Timestamp is the time the event was received. This could optionally be
	// used for client-stored sent messages too. If the server supports the
	// "server-time" capability, this is synced to the UTC time that the server
	// specifies.
	Timestamp time.Time `json:"timestamp"`
	// Command that represents the event, e.g. JOIN, PRIVMSG, KILL.
	Command string `json:"command"`
	// Params (parameters/args) to the command. Commonly nickname, channel, etc.
	// The last item in the slice could potentially contain spaces (commonly
	// referred to as the "trailing" parameter).
	Params []string `json:"params"`
	// Sensitive should be true if the message is sensitive (e.g. and should
	// not be logged/shown in debugging output).
	Sensitive bool `json:"sensitive"`
	// If the event is an echo-message response.
	Echo bool `json:"echo"`
}

// Last returns the last parameter in Event.Params if it exists.
func (e *Event) Last() string {
	if len(e.Params) >= 1 {
		return e.Params[len(e.Params)-1]
	}
	return ""
}

// Copy makes a deep copy of a given event, for use with allowing untrusted
// functions/handlers edit the event without causing potential issues with
// other handlers.
func (e *Event) Copy() *Event {
	if e == nil {
		return nil
	}

	newEvent := &Event{
		Timestamp: e.Timestamp,
		Command:   e.Command,
		Sensitive: e.Sensitive,
		Echo:      e.Echo,
	}

	// Copy Source field, as it's a pointer and needs to be dereferenced.
	if e.Source != nil {
		newEvent.Source = e.Source.Copy()
	}

	// Copy Params in order to dereference as well.
	if e.Params != nil {
		newEvent.Params = make([]string, len(e.Params))
		copy(newEvent.Params, e.Params)
	}

	// Copy tags as necessary.
	if e.Tags != nil {
		newEvent.Tags = Tags{}
		for k, v := range e.Tags {
			newEvent.Tags[k] = v
		}
	}

	return newEvent
}

// Equals compares two Events for equality.
func (e *Event) Equals(ev *Event) bool {
	if e.Command != ev.Command || len(e.Params) != len(ev.Params) {
		return false
	}

	for i := 0; i < len(e.Params); i++ {
		if e.Params[i] != ev.Params[i] {
			return false
		}
	}

	if !e.Source.Equals(ev.Source) || !e.Tags.Equals(ev.Tags) {
		return false
	}

	return true
}

// Len calculates the length of the string representation of event. Note that
// this will return the true length (even if longer than what IRC supports),
// which may be useful if you are trying to check and see if a message is
// too long, to trim it down yourself.
func (e *Event) Len() (length int) {
	if e.Tags != nil {
		// Include tags and trailing space.
		length = e.Tags.Len() + 1
	}
	if e.Source != nil {
		// Include prefix and trailing space.
		length += e.Source.Len() + 2
	}

	length += len(e.Command)

	if len(e.Params) > 0 {
		// Spaces before each param.
		length += len(e.Params)

		for i := 0; i < len(e.Params); i++ {
			length += len(e.Params[i])

			// If param contains a space or it's empty, it's trailing, so it should be
			// prefixed with a colon (:).
			if i == len(e.Params)-1 && (strings.Contains(e.Params[i], " ") || strings.HasPrefix(e.Params[i], ":") || e.Params[i] == "") {
				length++
			}
		}
	}

	return
}

// Bytes returns a []byte representation of event. Strips all newlines and
// carriage returns.
//
// Per RFC2812 section 2.3, messages should not exceed 512 characters in
// length. This method forces that limit by discarding any characters
// exceeding the length limit.
func (e *Event) Bytes() []byte {
	buffer := new(bytes.Buffer)

	// Tags.
	if e.Tags != nil {
		e.Tags.writeTo(buffer)
	}

	// Event prefix.
	if e.Source != nil {
		buffer.WriteByte(messagePrefix)
		e.Source.writeTo(buffer)
		buffer.WriteByte(eventSpace)
	}

	// Command is required.
	buffer.WriteString(e.Command)

	// Space separated list of arguments.
	if len(e.Params) > 0 {
		for i := 0; i < len(e.Params); i++ {
			if i == len(e.Params)-1 && (strings.Contains(e.Params[i], " ") || strings.HasPrefix(e.Params[i], ":") || e.Params[i] == "") {
				buffer.WriteString(string(eventSpace) + string(messagePrefix) + e.Params[i])
				continue
			}
			buffer.WriteString(string(eventSpace) + e.Params[i])
		}
	}

	// We need the limit the buffer length.
	if buffer.Len() > (maxLength) {
		buffer.Truncate(maxLength)
	}

	// If we truncated in the middle of a utf8 character, we need to remove
	// the other (now invalid) bytes.
	out := bytes.ToValidUTF8(buffer.Bytes(), nil)

	// Strip newlines and carriage returns.
	for i := 0; i < len(out); i++ {
		if out[i] == '\n' || out[i] == '\r' {
			out = append(out[:i], out[i+1:]...)
			i-- // Decrease the index so we can pick up where we left off.
		}
	}

	return out
}

// String returns a string representation of this event. Strips all newlines
// and carriage returns.
func (e *Event) String() string {
	return string(e.Bytes())
}

// Pretty returns a prettified string of the event. If the event doesn't
// support prettification, ok is false. Pretty is not just useful to make
// an event prettier, but also to filter out events that most don't visually
// see in normal IRC clients. e.g. most clients don't show WHO queries.
func (e *Event) Pretty() (out string, ok bool) {
	if e.Sensitive || e.Echo {
		return "", false
	}

	if e.Command == ERROR {
		return fmt.Sprintf("[*] an error occurred: %s", e.Last()), true
	}

	if e.Source == nil {
		if e.Command != PRIVMSG && e.Command != NOTICE {
			return "", false
		}

		if len(e.Params) > 0 {
			return fmt.Sprintf("[>] writing %s", e.String()), true
		}

		return "", false
	}

	if e.Command == INITIALIZED {
		return fmt.Sprintf("[*] connection to %s initialized", e.Last()), true
	}

	if e.Command == CONNECTED {
		return fmt.Sprintf("[*] successfully connected to %s", e.Last()), true
	}

	if (e.Command == PRIVMSG || e.Command == NOTICE) && len(e.Params) > 0 {
		if ctcp := DecodeCTCP(e); ctcp != nil {
			if ctcp.Reply {
				return
			}

			if ctcp.Command == CTCP_ACTION {
				return fmt.Sprintf("[%s] **%s** %s", strings.Join(e.Params[0:len(e.Params)-1], ","), ctcp.Source.Name, ctcp.Text), true
			}

			return fmt.Sprintf("[*] CTCP query from %s: %s%s", ctcp.Source.Name, ctcp.Command, " "+ctcp.Text), true
		}

		var source string
		if e.Command == PRIVMSG {
			source = fmt.Sprintf("(%s)", e.Source.Name)
		} else { // NOTICE
			source = fmt.Sprintf("--%s--", e.Source.Name)
		}

		return fmt.Sprintf("[%s] %s %s", strings.Join(e.Params[0:len(e.Params)-1], ","), source, e.Last()), true
	}

	if e.Command == RPL_MOTD || e.Command == RPL_MOTDSTART ||
		e.Command == RPL_WELCOME || e.Command == RPL_YOURHOST ||
		e.Command == RPL_CREATED || e.Command == RPL_LUSERCLIENT {
		return "[*] " + e.Last(), true
	}

	if e.Command == JOIN && len(e.Params) > 0 {
		return fmt.Sprintf("[*] %s (%s) has joined %s", e.Source.Name, e.Source.Host, e.Params[0]), true
	}

	if e.Command == PART && len(e.Params) > 0 {
		return fmt.Sprintf("[*] %s (%s) has left %s (%s)", e.Source.Name, e.Source.Host, e.Params[0], e.Last()), true
	}

	if e.Command == QUIT {
		return fmt.Sprintf("[*] %s has quit (%s)", e.Source.Name, e.Last()), true
	}

	if e.Command == INVITE && len(e.Params) == 1 {
		return fmt.Sprintf("[*] %s invited to %s by %s", e.Params[0], e.Last(), e.Source.Name), true
	}

	if e.Command == KICK && len(e.Params) >= 2 {
		return fmt.Sprintf("[%s] *** %s has kicked %s: %s", e.Params[0], e.Source.Name, e.Params[1], e.Last()), true
	}

	if e.Command == NICK {
		return fmt.Sprintf("[*] %s is now known as %s", e.Source.Name, e.Last()), true
	}

	if e.Command == TOPIC && len(e.Params) >= 2 {
		return fmt.Sprintf("[%s] *** %s has set the topic to: %s", e.Params[0], e.Source.Name, e.Last()), true
	}

	if e.Command == RPL_TOPIC && len(e.Params) > 0 {
		if len(e.Params) >= 2 {
			return fmt.Sprintf("[*] topic for %s is: %s", e.Params[1], e.Last()), true
		}
		return fmt.Sprintf("[*] topic for %s is: %s", e.Params[0], e.Last()), true
	}

	if e.Command == MODE && len(e.Params) > 2 {
		return fmt.Sprintf("[%s] *** %s set modes: %s", e.Params[0], e.Source.Name, strings.Join(e.Params[1:], " ")), true
	}

	if e.Command == CAP_AWAY {
		if len(e.Params) > 0 {
			return fmt.Sprintf("[*] %s is now away: %s", e.Source.Name, e.Last()), true
		}

		return fmt.Sprintf("[*] %s is no longer away", e.Source.Name), true
	}

	if e.Command == CAP_CHGHOST && len(e.Params) == 2 {
		return fmt.Sprintf("[*] %s has changed their host to %s (was %s)", e.Source.Name, e.Params[1], e.Source.Host), true
	}

	if e.Command == CAP_ACCOUNT && len(e.Params) == 1 {
		if e.Params[0] == "*" {
			return fmt.Sprintf("[*] %s has become un-authenticated", e.Source.Name), true
		}

		return fmt.Sprintf("[*] %s has authenticated for account: %s", e.Source.Name, e.Params[0]), true
	}

	if e.Command == CAP && len(e.Params) >= 2 && e.Params[1] == CAP_ACK {
		return "[*] enabling capabilities: " + e.Last(), true
	}

	return "", false
}

// IsAction checks to see if the event is an ACTION (/me).
func (e *Event) IsAction() bool {
	if e.Command != PRIVMSG {
		return false
	}

	ok, ctcp := e.IsCTCP()
	return ok && ctcp.Command == CTCP_ACTION
}

// IsCTCP checks to see if the event is a CTCP event, and if so, returns the
// converted CTCP event.
func (e *Event) IsCTCP() (ok bool, ctcp *CTCPEvent) {
	ctcp = DecodeCTCP(e)
	return ctcp != nil, ctcp
}

// IsFromChannel checks to see if a message was from a channel (rather than
// a private message).
func (e *Event) IsFromChannel() bool {
	if e.Source == nil || (e.Command != PRIVMSG && e.Command != NOTICE) || len(e.Params) < 1 {
		return false
	}

	if !IsValidChannel(e.Params[0]) {
		return false
	}

	return true
}

// IsFromUser checks to see if a message was from a user (rather than a
// channel).
func (e *Event) IsFromUser() bool {
	if e.Source == nil || (e.Command != PRIVMSG && e.Command != NOTICE) || len(e.Params) < 1 {
		return false
	}

	if !IsValidNick(e.Params[0]) {
		return false
	}

	return true
}

// StripAction returns the stripped version of the action encoding from a
// PRIVMSG ACTION (/me).
func (e *Event) StripAction() string {
	if !e.IsAction() {
		return e.Last()
	}

	msg := e.Last()
	return msg[8 : len(msg)-1]
}

const (
	messagePrefix byte = ':' // Prefix or last argument.
	prefixIdent   byte = '!' // Username.
	prefixHost    byte = '@' // Hostname.
)

// Source represents the sender of an IRC event, see RFC1459 section 2.3.1.
// <servername> | <nick> [ '!' <user> ] [ '@' <host> ]
type Source struct {
	// Name is the nickname, server name, or service name, in its original
	// non-rfc1459 form.
	Name string `json:"name"`
	// Ident is commonly known as the "user".
	Ident string `json:"ident"`
	// Host is the hostname or IP address of the user/service. Is not accurate
	// due to how IRC servers can spoof hostnames.
	Host string `json:"host"`
}

// ID is the nickname, server name, or service name, in it's converted
// and comparable) form.
func (s *Source) ID() string {
	return ToRFC1459(s.Name)
}

// Equals compares two Sources for equality.
func (s *Source) Equals(ss *Source) bool {
	if s == nil && ss == nil {
		return true
	}
	if s != nil && ss == nil || s == nil && ss != nil {
		return false
	}
	if s.ID() != ss.ID() || s.Ident != ss.Ident || s.Host != ss.Host {
		return false
	}
	return true
}

// Copy returns a deep copy of Source.
func (s *Source) Copy() *Source {
	if s == nil {
		return nil
	}

	newSource := &Source{
		Name:  s.Name,
		Ident: s.Ident,
		Host:  s.Host,
	}

	return newSource
}

// ParseSource takes a string and attempts to create a Source struct.
func ParseSource(raw string) (src *Source) {
	src = new(Source)

	user := strings.IndexByte(raw, prefixIdent)
	host := strings.IndexByte(raw, prefixHost)

	switch {
	case user > 0 && host > user:
		src.Name = raw[:user]
		src.Ident = raw[user+1 : host]
		src.Host = raw[host+1:]
	case user > 0:
		src.Name = raw[:user]
		src.Ident = raw[user+1:]
	case host > 0:
		src.Name = raw[:host]
		src.Host = raw[host+1:]
	default:
		src.Name = raw
	}

	return src
}

// Len calculates the length of the string representation of prefix
func (s *Source) Len() (length int) {
	length = len(s.Name)
	if len(s.Ident) > 0 {
		length = 1 + length + len(s.Ident)
	}
	if len(s.Host) > 0 {
		length = 1 + length + len(s.Host)
	}

	return
}

// Bytes returns a []byte representation of source.
func (s *Source) Bytes() []byte {
	buffer := new(bytes.Buffer)
	s.writeTo(buffer)

	return buffer.Bytes()
}

// String returns a string representation of source.
func (s *Source) String() (out string) {
	out = s.Name
	if len(s.Ident) > 0 {
		out = out + string(prefixIdent) + s.Ident
	}
	if len(s.Host) > 0 {
		out = out + string(prefixHost) + s.Host
	}

	return
}

// IsHostmask returns true if source looks like a user hostmask.
func (s *Source) IsHostmask() bool {
	return len(s.Ident) > 0 && len(s.Host) > 0
}

// IsServer returns true if this source looks like a server name.
func (s *Source) IsServer() bool {
	return s.Ident == "" && s.Host == ""
}

// writeTo is an utility function to write the source to the bytes.Buffer
// in Event.String().
func (s *Source) writeTo(buffer *bytes.Buffer) {
	buffer.WriteString(s.Name)
	if len(s.Ident) > 0 {
		buffer.WriteByte(prefixIdent)
		buffer.WriteString(s.Ident)
	}
	if len(s.Host) > 0 {
		buffer.WriteByte(prefixHost)
		buffer.WriteString(s.Host)
	}
}
