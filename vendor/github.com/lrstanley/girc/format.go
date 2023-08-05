// Copyright (c) Liam Stanley <me@liamstanley.io>. All rights reserved. Use
// of this source code is governed by the MIT license that can be found in
// the LICENSE file.

package girc

import (
	"bytes"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"unicode/utf8"
)

const (
	fmtOpenChar        = '{'
	fmtCloseChar       = '}'
	maxWordSplitLength = 30
)

var (
	reCode  = regexp.MustCompile(`(\x02|\x1d|\x0f|\x03|\x16|\x1f|\x01)`)
	reColor = regexp.MustCompile(`\x03([019]?\d(,[019]?\d)?)`)
)

var fmtColors = map[string]int{
	"white":       0,
	"black":       1,
	"blue":        2,
	"navy":        2,
	"green":       3,
	"red":         4,
	"brown":       5,
	"maroon":      5,
	"purple":      6,
	"gold":        7,
	"olive":       7,
	"orange":      7,
	"yellow":      8,
	"lightgreen":  9,
	"lime":        9,
	"teal":        10,
	"cyan":        11,
	"lightblue":   12,
	"royal":       12,
	"fuchsia":     13,
	"lightpurple": 13,
	"pink":        13,
	"gray":        14,
	"grey":        14,
	"lightgrey":   15,
	"silver":      15,
}

var fmtCodes = map[string]string{
	"bold":      "\x02",
	"b":         "\x02",
	"italic":    "\x1d",
	"i":         "\x1d",
	"reset":     "\x0f",
	"r":         "\x0f",
	"clear":     "\x03",
	"c":         "\x03", // Clears formatting.
	"reverse":   "\x16",
	"underline": "\x1f",
	"ul":        "\x1f",
	"ctcp":      "\x01", // CTCP/ACTION delimiter.
}

// Fmt takes format strings like "{red}" or "{red,blue}" (for background
// colors) and turns them into the resulting ASCII format/color codes for IRC.
// See format.go for the list of supported format codes allowed.
//
// For example:
//
//	client.Message("#channel", Fmt("{red}{b}Hello {red,blue}World{c}"))
func Fmt(text string) string {
	last := -1
	for i := 0; i < len(text); i++ {
		if text[i] == fmtOpenChar {
			last = i
			continue
		}

		if text[i] == fmtCloseChar && last > -1 {
			code := strings.ToLower(text[last+1 : i])

			// Check to see if they're passing in a second (background) color
			// as {fgcolor,bgcolor}.
			var secondary string
			if com := strings.Index(code, ","); com > -1 {
				secondary = code[com+1:]
				code = code[:com]
			}

			var repl string

			if color, ok := fmtColors[code]; ok {
				repl = fmt.Sprintf("\x03%02d", color)
			}

			if repl != "" && secondary != "" {
				if color, ok := fmtColors[secondary]; ok {
					repl += fmt.Sprintf(",%02d", color)
				}
			}

			if repl == "" {
				if fmtCode, ok := fmtCodes[code]; ok {
					repl = fmtCode
				}
			}

			next := len(text[:last]+repl) - 1
			text = text[:last] + repl + text[i+1:]
			last = -1
			i = next
			continue
		}

		if last > -1 {
			// A-Z, a-z, and ","
			if text[i] != ',' && (text[i] < 'A' || text[i] > 'Z') && (text[i] < 'a' || text[i] > 'z') {
				last = -1
				continue
			}
		}
	}

	return text
}

// TrimFmt strips all "{fmt}" formatting strings from the input text.
// See Fmt() for more information.
func TrimFmt(text string) string {
	for color := range fmtColors {
		text = strings.ReplaceAll(text, string(fmtOpenChar)+color+string(fmtCloseChar), "")
	}
	for code := range fmtCodes {
		text = strings.ReplaceAll(text, string(fmtOpenChar)+code+string(fmtCloseChar), "")
	}

	return text
}

// StripRaw tries to strip all ASCII format codes that are used for IRC.
// Primarily, foreground/background colors, and other control bytes like
// reset, bold, italic, reverse, etc. This also is done in a specific way
// in order to ensure no truncation of other non-irc formatting.
func StripRaw(text string) string {
	text = reColor.ReplaceAllString(text, "")

	for _, code := range fmtCodes {
		text = strings.ReplaceAll(text, code, "")
	}

	return text
}

// IsValidChannel validates if channel is an RFC compliant channel or not.
//
// NOTE: If you are using this to validate a channel that contains a channel
// ID, (!<channelid>NAME), this only supports the standard 5 character length.
//
// NOTE: If you do not need to validate against servers that support unicode,
// you may want to ensure that all channel chars are within the range of
// all ASCII printable chars. This function will NOT do that for
// compatibility reasons.
//
//	channel    =  ( "#" / "+" / ( "!" channelid ) / "&" ) chanstring
//	              [ ":" chanstring ]
//	chanstring =  0x01-0x07 / 0x08-0x09 / 0x0B-0x0C / 0x0E-0x1F / 0x21-0x2B
//	chanstring =  / 0x2D-0x39 / 0x3B-0xFF
//	                ; any octet except NUL, BELL, CR, LF, " ", "," and ":"
//	channelid  = 5( 0x41-0x5A / digit )   ; 5( A-Z / 0-9 )
func IsValidChannel(channel string) bool {
	if len(channel) <= 1 || len(channel) > 50 {
		return false
	}

	// #, +, !<channelid>, ~, or &
	// Including "*" and "~" in the prefix list, as these are commonly used
	// (e.g. ZNC.)
	if bytes.IndexByte([]byte{'!', '#', '&', '*', '~', '+'}, channel[0]) == -1 {
		return false
	}

	// !<channelid> -- not very commonly supported, but we'll check it anyway.
	// The ID must be 5 chars. This means min-channel size should be:
	//   1 (prefix) + 5 (id) + 1 (+, channel name)
	// On some networks, this may be extended with ISUPPORT capabilities,
	// however this is extremely uncommon.
	if channel[0] == '!' {
		if len(channel) < 7 {
			return false
		}

		// check for valid ID
		for i := 1; i < 6; i++ {
			if (channel[i] < '0' || channel[i] > '9') && (channel[i] < 'A' || channel[i] > 'Z') {
				return false
			}
		}
	}

	// Check for invalid octets here.
	bad := []byte{0x00, 0x07, 0x0D, 0x0A, 0x20, 0x2C, 0x3A}
	for i := 1; i < len(channel); i++ {
		if bytes.IndexByte(bad, channel[i]) != -1 {
			return false
		}
	}

	return true
}

// IsValidNick validates an IRC nickname. Note that this does not validate
// IRC nickname length.
//
//	nickname =  ( letter / special ) *8( letter / digit / special / "-" )
//	letter   =  0x41-0x5A / 0x61-0x7A
//	digit    =  0x30-0x39
//	special  =  0x5B-0x60 / 0x7B-0x7D
func IsValidNick(nick string) bool {
	if nick == "" {
		return false
	}

	// Check the first index. Some characters aren't allowed for the first
	// index of an IRC nickname.
	if (nick[0] < 'A' || nick[0] > '}') && nick[0] != '?' {
		// a-z, A-Z, '_\[]{}^|', and '?' in the case of znc.
		return false
	}

	for i := 1; i < len(nick); i++ {
		if (nick[i] < 'A' || nick[i] > '}') && (nick[i] < '0' || nick[i] > '9') && nick[i] != '-' {
			// a-z, A-Z, 0-9, -, and _\[]{}^|
			return false
		}
	}

	return true
}

// IsValidUser validates an IRC ident/username. Note that this does not
// validate IRC ident length.
//
// The validation checks are much like what characters are allowed with an
// IRC nickname (see IsValidNick()), however an ident/username can:
//
// 1. Must either start with alphanumberic char, or "~" then alphanumberic
// char.
//
// 2. Contain a "." (period), for use with "first.last". Though, this may
// not be supported on all networks. Some limit this to only a single period.
//
// Per RFC:
//
//	user =  1*( %x01-09 / %x0B-0C / %x0E-1F / %x21-3F / %x41-FF )
//	        ; any octet except NUL, CR, LF, " " and "@"
func IsValidUser(name string) bool {
	if name == "" {
		return false
	}

	// "~" is prepended (commonly) if there was no ident server response.
	if name[0] == '~' {
		// Means name only contained "~".
		if len(name) < 2 {
			return false
		}

		name = name[1:]
	}

	// Check to see if the first index is alphanumeric.
	if (name[0] < 'A' || name[0] > 'Z') && (name[0] < 'a' || name[0] > 'z') && (name[0] < '0' || name[0] > '9') {
		return false
	}

	for i := 1; i < len(name); i++ {
		if (name[i] < 'A' || name[i] > '}') && (name[i] < '0' || name[i] > '9') && name[i] != '-' && name[i] != '.' {
			// a-z, A-Z, 0-9, -, and _\[]{}^|
			return false
		}
	}

	return true
}

// ToRFC1459 converts a string to the stripped down conversion within RFC
// 1459. This will do things like replace an "A" with an "a", "[]" with "{}",
// and so forth. Useful to compare two nicknames or channels. Note that this
// should not be used to normalize nicknames or similar, as this may convert
// valid input characters to non-rfc-valid characters. As such, it's main use
// is for comparing two nicks.
func ToRFC1459(input string) string {
	var out string

	for i := 0; i < len(input); i++ {
		if input[i] >= 65 && input[i] <= 94 {
			out += string(rune(input[i]) + 32)
		} else {
			out += string(input[i])
		}
	}

	return out
}

const globChar = "*"

// Glob will test a string pattern, potentially containing globs, against a
// string. The glob character is *.
func Glob(input, match string) bool {
	// Empty pattern.
	if match == "" {
		return input == match
	}

	// If a glob, match all.
	if match == globChar {
		return true
	}

	parts := strings.Split(match, globChar)

	if len(parts) == 1 {
		// No globs, test for equality.
		return input == match
	}

	leadingGlob, trailingGlob := strings.HasPrefix(match, globChar), strings.HasSuffix(match, globChar)
	last := len(parts) - 1

	// Check prefix first.
	if !leadingGlob && !strings.HasPrefix(input, parts[0]) {
		return false
	}

	// Check middle section.
	for i := 1; i < last; i++ {
		if !strings.Contains(input, parts[i]) {
			return false
		}

		// Trim already-evaluated text from input during loop over match
		// text.
		idx := strings.Index(input, parts[i]) + len(parts[i])
		input = input[idx:]
	}

	// Check suffix last.
	return trailingGlob || strings.HasSuffix(input, parts[last])
}

// sliceInsert inserts a string into a slice at a specific index, while trying
// to avoid as many allocations as possible.
func sliceInsert(input []string, i int, v ...string) []string {
	total := len(input) + len(v)
	if total <= cap(input) {
		output := input[:total]
		copy(output[i+len(v):], input[i:])
		copy(output[i:], v)
		return output
	}
	output := make([]string, total)
	copy(output, input[:i])
	copy(output[i:], v)
	copy(output[i+len(v):], input[i:])
	return output
}

// splitMessage is a text splitter that takes into consideration a few things:
//   - Ensuring the returned text is no longer than maxWidth.
//   - Attempting to split at the closest word boundary, while still staying inside
//     of the specific maxWidth.
//   - if there is no good word boundary for longer words (or e.g. links, raw data, etc)
//     that are above maxWordSplitLength characters, split the word into chunks to fit the
//
// maximum width.
func splitMessage(input string, maxWidth int) (output []string) {
	input = strings.ToValidUTF8(input, "?")

	words := strings.FieldsFunc(strings.TrimSpace(input), func(r rune) bool {
		switch r { // Same as unicode.IsSpace, but without ctrl/lf.
		case '\t', '\v', '\f', ' ', 0x85, 0xA0:
			return true
		}
		return false
	})

	output = []string{""}
	codes := []string{}

	var lastColor string
	var match []string

	for i := 0; i < len(words); i++ {
		j := strings.IndexAny(words[i], "\n\r")
		if j == -1 {
			continue
		}

		word := words[i]
		words[i] = word[:j]

		words = sliceInsert(words, i+1, "", strings.TrimLeft(word[j:], "\n\r"))
	}

	for _, word := range words {
		// Used in place of a single newline.
		if word == "" {
			// Last line was already empty or already only had control characters.
			if output[len(output)-1] == "" || output[len(output)-1] == lastColor+word {
				continue
			}

			output = append(output, strings.Join(codes, "")+lastColor+word)
			continue
		}

		// Keep track of the last used color codes.
		match = reColor.FindAllString(word, -1)
		if len(match) > 0 {
			lastColor = match[len(match)-1]
		}

		// Find all sequence codes -- this approach isn't perfect (ideally, a lexer
		// should be used to track each exact type of code), but it's good enough for
		// most cases.
		match = reCode.FindAllString(word, -1)
		if len(match) > 0 {
			for _, m := range match {
				// Reset was used, so clear all codes.
				if m == fmtCodes["reset"] {
					lastColor = ""
					codes = []string{}
					continue
				}

				// Check if we already have the code, and if so, remove it (closing).
				contains := false
				for i := 0; i < len(codes); i++ {
					if m == codes[i] {
						contains = true
						codes = append(codes[:i], codes[i+1:]...)

						// If it's a closing color code, reset the last used color
						// as well.
						if m == fmtCodes["clear"] {
							lastColor = ""
						}

						break
					}
				}

				// Track the new code, unless it's a color clear but we aren't
				// tracking a color right now.
				if !contains && (lastColor == "" || m != fmtCodes["clear"]) {
					codes = append(codes, m)
				}
			}
		}

	checkappend:

		// Check if we can append, otherwise we must split.
		if 1+utf8.RuneCountInString(word)+utf8.RuneCountInString(output[len(output)-1]) < maxWidth {
			if output[len(output)-1] != "" {
				output[len(output)-1] += " "
			}
			output[len(output)-1] += word
			continue
		}

		// If the word can fit on a line by itself, check if it's a url. If it is,
		// put it on it's own line.
		if utf8.RuneCountInString(word+strings.Join(codes, "")+lastColor) < maxWidth {
			if _, err := url.Parse(word); err == nil {
				output = append(output, strings.Join(codes, "")+lastColor+word)
				continue
			}
		}

		// Check to see if we can split by misc symbols, but must be at least a few
		// characters long to be split by it.
		if j := strings.IndexAny(word, "-+_=|/~:;,."); j > 3 && 1+utf8.RuneCountInString(word[0:j])+utf8.RuneCountInString(output[len(output)-1]) < maxWidth {
			if output[len(output)-1] != "" {
				output[len(output)-1] += " "
			}
			output[len(output)-1] += word[0:j]
			word = word[j+1:]
			goto checkappend
		}

		// If the word is longer than is acceptable to just put on the next line,
		// split it into chunks. Also don't split the word if only a few characters
		// left of the word would be on the next line.
		if 1+utf8.RuneCountInString(word) > maxWordSplitLength && maxWidth-utf8.RuneCountInString(output[len(output)-1]) > 5 {
			left := maxWidth - utf8.RuneCountInString(output[len(output)-1]) - 1 // -1 for the space

			if output[len(output)-1] != "" {
				output[len(output)-1] += " "
			}
			output[len(output)-1] += word[0:left]
			word = word[left:]
			goto checkappend
		}

		left := maxWidth - utf8.RuneCountInString(output[len(output)-1])
		output[len(output)-1] += word[0:left]

		output = append(output, strings.Join(codes, "")+lastColor)
		word = word[left:]
		goto checkappend
	}

	for i := 0; i < len(output); i++ {
		output[i] = strings.ToValidUTF8(output[i], "?")
	}
	return output
}
