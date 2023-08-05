// Copyright (c) Liam Stanley <me@liamstanley.io>. All rights reserved. Use
// of this source code is governed by the MIT license that can be found in
// the LICENSE file.

package girc

import (
	"encoding/json"
	"strings"
	"sync"
)

// CMode represents a single step of a given mode change.
type CMode struct {
	add     bool   // if it's a +, or -.
	name    byte   // character representation of the given mode.
	setting bool   // if it's a setting (should be stored) or temporary (op/voice/etc).
	args    string // arguments to the mode, if arguments are supported.
}

// Short returns a short representation of a mode without arguments. E.g. "+a",
// or "-b".
func (c *CMode) Short() string {
	var status string
	if c.add {
		status = "+"
	} else {
		status = "-"
	}

	return status + string(c.name)
}

// String returns a string representation of a mode, including optional
// arguments. E.g. "+b user*!ident@host.*.com"
func (c *CMode) String() string {
	if c.args == "" {
		return c.Short()
	}

	return c.Short() + " " + c.args
}

// CModes is a representation of a set of modes. This may be the given state
// of a channel/user, or the given state changes to a given channel/user.
type CModes struct {
	raw           string // raw supported modes.
	modesListArgs string // modes that add/remove users from lists and support args.
	modesArgs     string // modes that support args.
	modesSetArgs  string // modes that support args ONLY when set.
	modesNoArgs   string // modes that do not support args.

	prefixes string  // user permission prefixes. these aren't a CMode.setting.
	modes    []CMode // the list of modes for this given state.
}

// Copy returns a deep copy of CModes.
func (c *CModes) Copy() (nc CModes) {
	nc = CModes{}
	nc = *c

	nc.modes = make([]CMode, len(c.modes))

	// Copy modes.
	for i := 0; i < len(c.modes); i++ {
		nc.modes[i] = c.modes[i]
	}

	return nc
}

// String returns a complete set of modes for this given state (change?). For
// example, "+a-b+cde some-arg".
func (c *CModes) String() string {
	var out string
	var args string

	if len(c.modes) > 0 {
		out += "+"
	}

	for i := 0; i < len(c.modes); i++ {
		out += string(c.modes[i].name)

		if len(c.modes[i].args) > 0 {
			args += " " + c.modes[i].args
		}
	}

	return out + args
}

// HasMode checks if the CModes state has a given mode. E.g. "m", or "I".
func (c *CModes) HasMode(mode string) bool {
	for i := 0; i < len(c.modes); i++ {
		if string(c.modes[i].name) == mode {
			return true
		}
	}

	return false
}

// Get returns the arguments for a given mode within this session, if it
// supports args.
func (c *CModes) Get(mode string) (args string, ok bool) {
	for i := 0; i < len(c.modes); i++ {
		if string(c.modes[i].name) == mode {
			if c.modes[i].args == "" {
				return "", false
			}

			return c.modes[i].args, true
		}
	}

	return "", false
}

// hasArg checks to see if the mode supports arguments. What ones support this?:
//
//	A = Mode that adds or removes a nick or address to a list. Always has a parameter.
//	B = Mode that changes a setting and always has a parameter.
//	C = Mode that changes a setting and only has a parameter when set.
//	D = Mode that changes a setting and never has a parameter.
//	Note: Modes of type A return the list when there is no parameter present.
//	Note: Some clients assumes that any mode not listed is of type D.
//	Note: Modes in PREFIX are not listed but could be considered type B.
func (c *CModes) hasArg(set bool, mode byte) (hasArgs, isSetting bool) {
	if len(c.raw) < 1 {
		return false, true
	}

	if strings.IndexByte(c.modesListArgs, mode) > -1 {
		return true, false
	}

	if strings.IndexByte(c.modesArgs, mode) > -1 {
		return true, true
	}

	if strings.IndexByte(c.modesSetArgs, mode) > -1 {
		if set {
			return true, true
		}

		return false, true
	}

	if strings.IndexByte(c.prefixes, mode) > -1 {
		return true, false
	}

	return false, true
}

// Apply merges two state changes, or one state change into a state of modes.
// For example, the latter would mean applying an incoming MODE with the modes
// stored for a channel.
func (c *CModes) Apply(modes []CMode) {
	var newModes []CMode

	for j := 0; j < len(c.modes); j++ {
		isin := false
		for i := 0; i < len(modes); i++ {
			if !modes[i].setting {
				continue
			}
			if c.modes[j].name == modes[i].name && modes[i].add {
				newModes = append(newModes, modes[i])
				isin = true
				break
			}
		}

		if !isin {
			newModes = append(newModes, c.modes[j])
		}
	}

	for i := 0; i < len(modes); i++ {
		if !modes[i].setting || !modes[i].add {
			continue
		}

		isin := false
		for j := 0; j < len(newModes); j++ {
			if modes[i].name == newModes[j].name {
				isin = true
				break
			}
		}

		if !isin {
			newModes = append(newModes, modes[i])
		}
	}

	c.modes = newModes
}

// Parse parses a set of flags and args, returning the necessary list of
// mappings for the mode flags.
func (c *CModes) Parse(flags string, args []string) (out []CMode) {
	// add is the mode state we're currently in. Adding, or removing modes.
	add := true
	var argCount int

	for i := 0; i < len(flags); i++ {
		if flags[i] == '+' {
			add = true
			continue
		}
		if flags[i] == '-' {
			add = false
			continue
		}

		mode := CMode{
			name: flags[i],
			add:  add,
		}

		hasArgs, isSetting := c.hasArg(add, flags[i])
		if hasArgs && len(args) > argCount {
			mode.args = args[argCount]
			argCount++
		}
		mode.setting = isSetting

		out = append(out, mode)
	}

	return out
}

// NewCModes returns a new CModes reference. channelModes and userPrefixes
// would be something you see from the server's "CHANMODES" and "PREFIX"
// ISUPPORT capability messages (alternatively, fall back to the standard)
// DefaultPrefixes and ModeDefaults.
func NewCModes(channelModes, userPrefixes string) CModes {
	split := strings.SplitN(channelModes, ",", 4)
	if len(split) != 4 {
		for i := len(split); i < 4; i++ {
			split = append(split, "")
		}
	}

	return CModes{
		raw:           channelModes,
		modesListArgs: split[0],
		modesArgs:     split[1],
		modesSetArgs:  split[2],
		modesNoArgs:   split[3],

		prefixes: userPrefixes,
		modes:    []CMode{},
	}
}

// IsValidChannelMode validates a channel mode (CHANMODES).
func IsValidChannelMode(raw string) bool {
	if len(raw) < 1 {
		return false
	}

	for i := 0; i < len(raw); i++ {
		// Allowed are: ",", A-Z and a-z.
		if raw[i] != ',' && (raw[i] < 'A' || raw[i] > 'Z') && (raw[i] < 'a' || raw[i] > 'z') {
			return false
		}
	}

	return true
}

// isValidUserPrefix validates a list of ISUPPORT-style user prefixes (PREFIX).
func isValidUserPrefix(raw string) bool {
	if len(raw) < 1 {
		return false
	}

	if raw[0] != '(' {
		return false
	}

	var keys, rep int
	var passedKeys bool

	// Skip the first one as we know it's (.
	for i := 1; i < len(raw); i++ {
		if raw[i] == ')' {
			passedKeys = true
			continue
		}

		if passedKeys {
			rep++
		} else {
			keys++
		}
	}

	return keys == rep
}

// parsePrefixes parses the mode character mappings from the symbols of a
// ISUPPORT-style user prefixes list (PREFIX).
func parsePrefixes(raw string) (modes, prefixes string) {
	if !isValidUserPrefix(raw) {
		return modes, prefixes
	}

	i := strings.Index(raw, ")")
	if i < 1 {
		return modes, prefixes
	}

	return raw[1:i], raw[i+1:]
}

// handleMODE handles incoming MODE messages, and updates the tracking
// information for each channel, as well as if any of the modes affect user
// permissions.
func handleMODE(c *Client, e Event) {
	// Check if it's a RPL_CHANNELMODEIS.
	if e.Command == RPL_CHANNELMODEIS && len(e.Params) > 2 {
		// RPL_CHANNELMODEIS sends the user as the first param, skip it.
		e.Params = e.Params[1:]
	}
	// Should be at least MODE <target> <flags>, to be useful. As well, only
	// tracking channel modes at the moment.
	if len(e.Params) < 2 || !IsValidChannel(e.Params[0]) {
		return
	}

	c.state.RLock()
	channel := c.state.lookupChannel(e.Params[0])
	if channel == nil {
		c.state.RUnlock()
		return
	}

	flags := e.Params[1]
	var args []string
	if len(e.Params) > 2 {
		args = append(args, e.Params[2:]...)
	}

	modes := channel.Modes.Parse(flags, args)
	channel.Modes.Apply(modes)

	// Loop through and update users modes as necessary.
	for i := 0; i < len(modes); i++ {
		if modes[i].setting || modes[i].args == "" {
			continue
		}

		user := c.state.lookupUser(modes[i].args)
		if user != nil {
			perms, _ := user.Perms.Lookup(channel.Name)
			perms.setFromMode(modes[i])
			user.Perms.set(channel.Name, perms)
		}
	}

	c.state.RUnlock()
	c.state.notify(c, UPDATE_STATE)
}

// chanModes returns the ISUPPORT list of server-supported channel modes,
// alternatively falling back to ModeDefaults.
func (s *state) chanModes() string {
	if modes, ok := s.serverOptions["CHANMODES"]; ok && IsValidChannelMode(modes) {
		return modes
	}

	return ModeDefaults
}

// userPrefixes returns the ISUPPORT list of server-supported user prefixes.
// This includes mode characters, as well as user prefix symbols. Falls back
// to DefaultPrefixes if not server-supported.
func (s *state) userPrefixes() string {
	if prefix, ok := s.serverOptions["PREFIX"]; ok && isValidUserPrefix(prefix) {
		return prefix
	}

	return DefaultPrefixes
}

// UserPerms contains all of the permissions for each channel the user is
// in.
type UserPerms struct {
	mu       sync.RWMutex
	channels map[string]Perms
}

// Copy returns a deep copy of the channel permissions.
func (p *UserPerms) Copy() (perms *UserPerms) {
	np := &UserPerms{
		channels: make(map[string]Perms),
	}

	p.mu.RLock()
	for key := range p.channels {
		np.channels[key] = p.channels[key]
	}
	p.mu.RUnlock()

	return np
}

// MarshalJSON implements json.Marshaler.
func (p *UserPerms) MarshalJSON() ([]byte, error) {
	p.mu.Lock()
	out, err := json.Marshal(&p.channels)
	p.mu.Unlock()

	return out, err
}

// Lookup looks up the users permissions for a given channel. ok is false
// if the user is not in the given channel.
func (p *UserPerms) Lookup(channel string) (perms Perms, ok bool) {
	p.mu.RLock()
	perms, ok = p.channels[ToRFC1459(channel)]
	p.mu.RUnlock()

	return perms, ok
}

func (p *UserPerms) set(channel string, perms Perms) {
	p.mu.Lock()
	p.channels[ToRFC1459(channel)] = perms
	p.mu.Unlock()
}

func (p *UserPerms) remove(channel string) {
	p.mu.Lock()
	delete(p.channels, ToRFC1459(channel))
	p.mu.Unlock()
}

// Perms contains all channel-based user permissions. The minimum op, and
// voice should be supported on all networks. This also supports non-rfc
// Owner, Admin, and HalfOp, if the network has support for it.
type Perms struct {
	// Owner (non-rfc) indicates that the user has full permissions to the
	// channel. More than one user can have owner permission.
	Owner bool `json:"owner"`
	// Admin (non-rfc) is commonly given to users that are trusted enough
	// to manage channel permissions, as well as higher level service settings.
	Admin bool `json:"admin"`
	// Op is commonly given to trusted users who can manage a given channel
	// by kicking, and banning users.
	Op bool `json:"op"`
	// HalfOp (non-rfc) is commonly used to give users permissions like the
	// ability to kick, without giving them greater abilities to ban all users.
	HalfOp bool `json:"half_op"`
	// Voice indicates the user has voice permissions, commonly given to known
	// users, with very light trust, or to indicate a user is active.
	Voice bool `json:"voice"`
}

// IsAdmin indicates that the user has banning abilities, and are likely a
// very trustable user (e.g. op+).
func (m Perms) IsAdmin() bool {
	if m.Owner || m.Admin || m.Op {
		return true
	}

	return false
}

// IsTrusted indicates that the user at least has modes set upon them, higher
// than a regular joining user.
func (m Perms) IsTrusted() bool {
	if m.IsAdmin() || m.HalfOp || m.Voice {
		return true
	}

	return false
}

// reset resets the modes of a user.
func (m *Perms) reset() {
	m.Owner = false
	m.Admin = false
	m.Op = false
	m.HalfOp = false
	m.Voice = false
}

// set translates raw prefix characters into proper permissions. Only
// use this function when you have a session lock.
func (m *Perms) set(prefix string, add bool) {
	if !add {
		m.reset()
	}

	for i := 0; i < len(prefix); i++ {
		switch string(prefix[i]) {
		case OwnerPrefix:
			m.Owner = true
		case AdminPrefix:
			m.Admin = true
		case OperatorPrefix:
			m.Op = true
		case HalfOperatorPrefix:
			m.HalfOp = true
		case VoicePrefix:
			m.Voice = true
		}
	}
}

// setFromMode sets user-permissions based on channel user mode chars. E.g.
// "o" being oper, "v" being voice, etc.
func (m *Perms) setFromMode(mode CMode) {
	switch string(mode.name) {
	case ModeOwner:
		m.Owner = mode.add
	case ModeAdmin:
		m.Admin = mode.add
	case ModeOperator:
		m.Op = mode.add
	case ModeHalfOperator:
		m.HalfOp = mode.add
	case ModeVoice:
		m.Voice = mode.add
	}
}

// parseUserPrefix parses a raw mode line, like "@user" or "@+user".
func parseUserPrefix(raw string) (modes, nick string, success bool) {
	for i := 0; i < len(raw); i++ {
		char := string(raw[i])

		if char == OwnerPrefix || char == AdminPrefix || char == HalfOperatorPrefix ||
			char == OperatorPrefix || char == VoicePrefix {
			modes += char
			continue
		}

		// Assume we've gotten to the nickname part.
		return modes, raw[i:], true
	}

	return
}
