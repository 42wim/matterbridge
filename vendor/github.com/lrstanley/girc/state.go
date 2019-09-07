// Copyright (c) Liam Stanley <me@liamstanley.io>. All rights reserved. Use
// of this source code is governed by the MIT license that can be found in
// the LICENSE file.

package girc

import (
	"fmt"
	"sort"
	"sync"
	"time"
)

// state represents the actively-changing variables within the client
// runtime. Note that everything within the state should be guarded by the
// embedded sync.RWMutex.
type state struct {
	sync.RWMutex
	// nick, ident, and host are the internal trackers for our user.
	nick, ident, host string
	// channels represents all channels we're active in.
	channels map[string]*Channel
	// users represents all of users that we're tracking.
	users map[string]*User
	// enabledCap are the capabilities which are enabled for this connection.
	enabledCap map[string]map[string]string
	// tmpCap are the capabilties which we share with the server during the
	// last capability check. These will get sent once we have received the
	// last capability list command from the server.
	tmpCap map[string]map[string]string
	// serverOptions are the standard capabilities and configurations
	// supported by the server at connection time. This also includes
	// RPL_ISUPPORT entries.
	serverOptions map[string]string
	// motd is the servers message of the day.
	motd string

	// sts are strict transport security configurations, if specified by the
	// server.
	//
	// TODO: ideally, this would be a configurable policy store that the user could
	// optionally override (to store STS information on disk, memory, etc).
	sts strictTransport
}

// reset resets the state back to it's original form.
func (s *state) reset(initial bool) {
	s.Lock()
	s.nick = ""
	s.ident = ""
	s.host = ""
	s.channels = make(map[string]*Channel)
	s.users = make(map[string]*User)
	s.serverOptions = make(map[string]string)
	s.enabledCap = make(map[string]map[string]string)
	s.tmpCap = make(map[string]map[string]string)
	s.motd = ""

	if initial {
		s.sts.reset()
	}
	s.Unlock()
}

// User represents an IRC user and the state attached to them.
type User struct {
	// Nick is the users current nickname. rfc1459 compliant.
	Nick string `json:"nick"`
	// Ident is the users username/ident. Ident is commonly prefixed with a
	// "~", which indicates that they do not have a identd server setup for
	// authentication.
	Ident string `json:"ident"`
	// Host is the visible host of the users connection that the server has
	// provided to us for their connection. May not always be accurate due to
	// many networks spoofing/hiding parts of the hostname for privacy
	// reasons.
	Host string `json:"host"`

	// ChannelList is a sorted list of all channels that we are currently
	// tracking the user in. Each channel name is rfc1459 compliant. See
	// User.Channels() for a shorthand if you're looking for the *Channel
	// version of the channel list.
	ChannelList []string `json:"channels"`

	// FirstSeen represents the first time that the user was seen by the
	// client for the given channel. Only usable if from state, not in past.
	FirstSeen time.Time `json:"first_seen"`
	// LastActive represents the last time that we saw the user active,
	// which could be during nickname change, message, channel join, etc.
	// Only usable if from state, not in past.
	LastActive time.Time `json:"last_active"`

	// Perms are the user permissions applied to this user that affect the given
	// channel. This supports non-rfc style modes like Admin, Owner, and HalfOp.
	Perms *UserPerms `json:"perms"`

	// Extras are things added on by additional tracking methods, which may
	// or may not work on the IRC server in mention.
	Extras struct {
		// Name is the users "realname" or full name. Commonly contains links
		// to the IRC client being used, or something of non-importance. May
		// also be empty if unsupported by the server/tracking is disabled.
		Name string `json:"name"`
		// Account refers to the account which the user is authenticated as.
		// This differs between each network (e.g. usually Nickserv, but
		// could also be something like Undernet). May also be empty if
		// unsupported by the server/tracking is disabled.
		Account string `json:"account"`
		// Away refers to the away status of the user. An empty string
		// indicates that they are active, otherwise the string is what they
		// set as their away message. May also be empty if unsupported by the
		// server/tracking is disabled.
		Away string `json:"away"`
	} `json:"extras"`
}

// Channels returns a reference of *Channels that the client knows the user
// is in. If you're just looking for the namme of the channels, use
// User.ChannelList.
func (u User) Channels(c *Client) []*Channel {
	if c == nil {
		panic("nil Client provided")
	}

	channels := []*Channel{}

	c.state.RLock()
	for i := 0; i < len(u.ChannelList); i++ {
		ch := c.state.lookupChannel(u.ChannelList[i])
		if ch != nil {
			channels = append(channels, ch)
		}
	}
	c.state.RUnlock()

	return channels
}

// Copy returns a deep copy of the user which can be modified without making
// changes to the actual state.
func (u *User) Copy() *User {
	if u == nil {
		return nil
	}

	nu := &User{}
	*nu = *u

	nu.Perms = u.Perms.Copy()
	_ = copy(nu.ChannelList, u.ChannelList)

	return nu
}

// addChannel adds the channel to the users channel list.
func (u *User) addChannel(name string) {
	if u.InChannel(name) {
		return
	}

	u.ChannelList = append(u.ChannelList, ToRFC1459(name))
	sort.Strings(u.ChannelList)

	u.Perms.set(name, Perms{})
}

// deleteChannel removes an existing channel from the users channel list.
func (u *User) deleteChannel(name string) {
	name = ToRFC1459(name)

	j := -1
	for i := 0; i < len(u.ChannelList); i++ {
		if u.ChannelList[i] == name {
			j = i
			break
		}
	}

	if j != -1 {
		u.ChannelList = append(u.ChannelList[:j], u.ChannelList[j+1:]...)
	}

	u.Perms.remove(name)
}

// InChannel checks to see if a user is in the given channel.
func (u *User) InChannel(name string) bool {
	name = ToRFC1459(name)

	for i := 0; i < len(u.ChannelList); i++ {
		if u.ChannelList[i] == name {
			return true
		}
	}

	return false
}

// Lifetime represents the amount of time that has passed since we have first
// seen the user.
func (u *User) Lifetime() time.Duration {
	return time.Since(u.FirstSeen)
}

// Active represents the the amount of time that has passed since we have
// last seen the user.
func (u *User) Active() time.Duration {
	return time.Since(u.LastActive)
}

// IsActive returns true if they were active within the last 30 minutes.
func (u *User) IsActive() bool {
	return u.Active() < (time.Minute * 30)
}

// Channel represents an IRC channel and the state attached to it.
type Channel struct {
	// Name of the channel. Must be rfc1459 compliant.
	Name string `json:"name"`
	// Topic of the channel.
	Topic string `json:"topic"`

	// UserList is a sorted list of all users we are currently tracking within
	// the channel. Each is the nickname, and is rfc1459 compliant.
	UserList []string `json:"user_list"`
	// Joined represents the first time that the client joined the channel.
	Joined time.Time `json:"joined"`
	// Modes are the known channel modes that the bot has captured.
	Modes CModes `json:"modes"`
}

// Users returns a reference of *Users that the client knows the channel has
// If you're just looking for just the name of the users, use Channnel.UserList.
func (ch Channel) Users(c *Client) []*User {
	if c == nil {
		panic("nil Client provided")
	}

	users := []*User{}

	c.state.RLock()
	for i := 0; i < len(ch.UserList); i++ {
		user := c.state.lookupUser(ch.UserList[i])
		if user != nil {
			users = append(users, user)
		}
	}
	c.state.RUnlock()

	return users
}

// Trusted returns a list of users which have voice or greater in the given
// channel. See Perms.IsTrusted() for more information.
func (ch Channel) Trusted(c *Client) []*User {
	if c == nil {
		panic("nil Client provided")
	}

	users := []*User{}

	c.state.RLock()
	for i := 0; i < len(ch.UserList); i++ {
		user := c.state.lookupUser(ch.UserList[i])
		if user == nil {
			continue
		}

		perms, ok := user.Perms.Lookup(ch.Name)
		if ok && perms.IsTrusted() {
			users = append(users, user)
		}
	}
	c.state.RUnlock()

	return users
}

// Admins returns a list of users which have half-op (if supported), or
// greater permissions (op, admin, owner, etc) in the given channel. See
// Perms.IsAdmin() for more information.
func (ch Channel) Admins(c *Client) []*User {
	if c == nil {
		panic("nil Client provided")
	}

	users := []*User{}

	c.state.RLock()
	for i := 0; i < len(ch.UserList); i++ {
		user := c.state.lookupUser(ch.UserList[i])
		if user == nil {
			continue
		}

		perms, ok := user.Perms.Lookup(ch.Name)
		if ok && perms.IsAdmin() {
			users = append(users, user)
		}
	}
	c.state.RUnlock()

	return users
}

// addUser adds a user to the users list.
func (ch *Channel) addUser(nick string) {
	if ch.UserIn(nick) {
		return
	}

	ch.UserList = append(ch.UserList, ToRFC1459(nick))
	sort.Strings(ch.UserList)
}

// deleteUser removes an existing user from the users list.
func (ch *Channel) deleteUser(nick string) {
	nick = ToRFC1459(nick)

	j := -1
	for i := 0; i < len(ch.UserList); i++ {
		if ch.UserList[i] == nick {
			j = i
			break
		}
	}

	if j != -1 {
		ch.UserList = append(ch.UserList[:j], ch.UserList[j+1:]...)
	}
}

// Copy returns a deep copy of a given channel.
func (ch *Channel) Copy() *Channel {
	if ch == nil {
		return nil
	}

	nc := &Channel{}
	*nc = *ch

	_ = copy(nc.UserList, ch.UserList)

	// And modes.
	nc.Modes = ch.Modes.Copy()

	return nc
}

// Len returns the count of users in a given channel.
func (ch *Channel) Len() int {
	return len(ch.UserList)
}

// UserIn checks to see if a given user is in a channel.
func (ch *Channel) UserIn(name string) bool {
	name = ToRFC1459(name)

	for i := 0; i < len(ch.UserList); i++ {
		if ch.UserList[i] == name {
			return true
		}
	}

	return false
}

// Lifetime represents the amount of time that has passed since we have first
// joined the channel.
func (ch *Channel) Lifetime() time.Duration {
	return time.Since(ch.Joined)
}

// createChannel creates the channel in state, if not already done.
func (s *state) createChannel(name string) (ok bool) {
	supported := s.chanModes()
	prefixes, _ := parsePrefixes(s.userPrefixes())

	if _, ok := s.channels[ToRFC1459(name)]; ok {
		return false
	}

	s.channels[ToRFC1459(name)] = &Channel{
		Name:     name,
		UserList: []string{},
		Joined:   time.Now(),
		Modes:    NewCModes(supported, prefixes),
	}

	return true
}

// deleteChannel removes the channel from state, if not already done.
func (s *state) deleteChannel(name string) {
	name = ToRFC1459(name)

	_, ok := s.channels[name]
	if !ok {
		return
	}

	for _, user := range s.channels[name].UserList {
		s.users[user].deleteChannel(name)

		if len(s.users[user].ChannelList) == 0 {
			// Assume we were only tracking them in this channel, and they
			// should be removed from state.

			delete(s.users, user)
		}
	}

	delete(s.channels, name)
}

// lookupChannel returns a reference to a channel, nil returned if no results
// found.
func (s *state) lookupChannel(name string) *Channel {
	return s.channels[ToRFC1459(name)]
}

// lookupUser returns a reference to a user, nil returned if no results
// found.
func (s *state) lookupUser(name string) *User {
	return s.users[ToRFC1459(name)]
}

// createUser creates the user in state, if not already done.
func (s *state) createUser(src *Source) (ok bool) {
	if _, ok := s.users[src.ID()]; ok {
		// User already exists.
		return false
	}

	s.users[src.ID()] = &User{
		Nick:       src.Name,
		Host:       src.Host,
		Ident:      src.Ident,
		FirstSeen:  time.Now(),
		LastActive: time.Now(),
		Perms:      &UserPerms{channels: make(map[string]Perms)},
	}

	return true
}

// deleteUser removes the user from channel state.
func (s *state) deleteUser(channelName, nick string) {
	user := s.lookupUser(nick)
	if user == nil {
		return
	}

	if channelName == "" {
		for i := 0; i < len(user.ChannelList); i++ {
			s.channels[user.ChannelList[i]].deleteUser(nick)
		}

		delete(s.users, ToRFC1459(nick))
		return
	}

	channel := s.lookupChannel(channelName)
	if channel == nil {
		return
	}

	user.deleteChannel(channelName)
	channel.deleteUser(nick)

	if len(user.ChannelList) == 0 {
		// This means they are no longer in any channels we track, delete
		// them from state.

		delete(s.users, ToRFC1459(nick))
	}
}

// renameUser renames the user in state, in all locations where relevant.
func (s *state) renameUser(from, to string) {
	from = ToRFC1459(from)

	// Update our nickname.
	if from == ToRFC1459(s.nick) {
		s.nick = to
	}

	user := s.lookupUser(from)
	if user == nil {
		return
	}

	delete(s.users, from)

	user.Nick = to
	user.LastActive = time.Now()
	s.users[ToRFC1459(to)] = user

	for i := 0; i < len(user.ChannelList); i++ {
		for j := 0; j < len(s.channels[user.ChannelList[i]].UserList); j++ {
			if s.channels[user.ChannelList[i]].UserList[j] == from {
				s.channels[user.ChannelList[i]].UserList[j] = ToRFC1459(to)

				sort.Strings(s.channels[user.ChannelList[i]].UserList)
				break
			}
		}
	}
}

type strictTransport struct {
	beginUpgrade        bool
	upgradePort         int
	persistenceDuration int
	persistenceReceived time.Time
	preload             bool
	lastFailed          time.Time
}

func (s *strictTransport) reset() {
	s.upgradePort = -1
	s.persistenceDuration = -1
	s.preload = false
}

func (s *strictTransport) expired() bool {
	return int(time.Since(s.persistenceReceived).Seconds()) > s.persistenceDuration
}

func (s *strictTransport) enabled() bool {
	return s.upgradePort > 0
}

// ErrSTSUpgradeFailed is an error that occurs when a connection that was attempted
// to be upgraded via a strict transport policy, failed. This does not necessarily
// indicate that STS was to blame, but the underlying connection failed for some
// reason.
type ErrSTSUpgradeFailed struct {
	Err error
}

func (e ErrSTSUpgradeFailed) Error() string {
	return fmt.Sprintf("fail to upgrade to secure (sts) connection: %v", e.Err)
}

// notify sends state change notifications so users can update their refs
// when state changes.
func (s *state) notify(c *Client, ntype string) {
	c.RunHandlers(&Event{Command: ntype})
}
