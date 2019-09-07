// Copyright (c) Liam Stanley <me@liamstanley.io>. All rights reserved. Use
// of this source code is governed by the MIT license that can be found in
// the LICENSE file.

package girc

import (
	"bytes"
	"fmt"
	"io"
	"sort"
	"strings"
)

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
	user := c.state.lookupUser(e.Source.ID())
	if user != nil {
		user.Extras.Account = account
	}
	c.state.Unlock()
	c.state.notify(c, UPDATE_STATE)
}

const (
	prefixTag      byte = '@'
	prefixTagValue byte = '='
	prefixUserTag  byte = '+'
	tagSeparator   byte = ';'
	maxTagLength   int  = 4094 // 4094 + @ and " " (space) = 4096, though space usually not included.
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
//
// Technically, there is a length limit of 4096, but the server should reject
// tag messages longer than this.
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
		// if !validTag(parts[i][:hasValue]) || !validTagValue(tagDecoder.Replace(parts[i][hasValue+1:])) {
		// 	continue
		// }

		t[parts[i][:hasValue]] = tagDecoder.Replace(parts[i][hasValue+1:])
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

// Equals compares two Tags for equality. With the msgid IRCv3 spec +\
// echo-message (amongst others), we may receive events that have msgid's,
// whereas our local events will not have the msgid. As such, don't compare
// all tags, only the necessary/important tags.
func (t Tags) Equals(tt Tags) bool {
	// The only tag which is important at this time.
	taccount, _ := t.Get("account")
	ttaccount, _ := tt.Get("account")
	return taccount == ttaccount
}

// Keys returns a slice of (unsorted) tag keys.
func (t Tags) Keys() (keys []string) {
	keys = make([]string, 0, t.Count())
	for key := range t {
		keys = append(keys, key)
	}
	return keys
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
		if (name[i] < 'A' || name[i] > 'Z') && (name[i] < 'a' || name[i] > 'z') && (name[i] < '-' || name[i] > '9') && name[i] != '_' {
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
		if value[i] < '!' || value[i] > '~' || value[i] == ';' {
			return false
		}
	}
	return true
}
