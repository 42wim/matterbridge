// Discordgo - Discord bindings for Go
// Available at https://github.com/bwmarrin/discordgo

// Copyright 2015-2016 Bruce Marriner <bruce@sqls.net>.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This file contains code related to state tracking.  If enabled, state
// tracking will capture the initial READY packet and many other websocket
// events and maintain an in-memory state of of guilds, channels, users, and
// so forth.  This information can be accessed through the Session.State struct.

package discordgo

import (
	"errors"
	"sync"
)

// ErrNilState is returned when the state is nil.
var ErrNilState = errors.New("State not instantiated, please use discordgo.New() or assign Session.State.")

// A State contains the current known state.
// As discord sends this in a READY blob, it seems reasonable to simply
// use that struct as the data store.
type State struct {
	sync.RWMutex
	Ready

	MaxMessageCount int
	TrackChannels   bool
	TrackEmojis     bool
	TrackMembers    bool
	TrackRoles      bool
	TrackVoice      bool

	guildMap   map[string]*Guild
	channelMap map[string]*Channel
}

// NewState creates an empty state.
func NewState() *State {
	return &State{
		Ready: Ready{
			PrivateChannels: []*Channel{},
			Guilds:          []*Guild{},
		},
		TrackChannels: true,
		TrackEmojis:   true,
		TrackMembers:  true,
		TrackRoles:    true,
		TrackVoice:    true,
		guildMap:      make(map[string]*Guild),
		channelMap:    make(map[string]*Channel),
	}
}

// OnReady takes a Ready event and updates all internal state.
func (s *State) OnReady(r *Ready) error {
	if s == nil {
		return ErrNilState
	}

	s.Lock()
	defer s.Unlock()

	s.Ready = *r

	for _, g := range s.Guilds {
		s.guildMap[g.ID] = g

		for _, c := range g.Channels {
			c.GuildID = g.ID
			s.channelMap[c.ID] = c
		}
	}

	for _, c := range s.PrivateChannels {
		s.channelMap[c.ID] = c
	}

	return nil
}

// GuildAdd adds a guild to the current world state, or
// updates it if it already exists.
func (s *State) GuildAdd(guild *Guild) error {
	if s == nil {
		return ErrNilState
	}

	s.Lock()
	defer s.Unlock()

	// Update the channels to point to the right guild, adding them to the channelMap as we go
	for _, c := range guild.Channels {
		c.GuildID = guild.ID
		s.channelMap[c.ID] = c
	}

	// If the guild exists, replace it.
	if g, ok := s.guildMap[guild.ID]; ok {
		// If this guild already exists with data, don't stomp on props.
		if g.Unavailable != nil && !*g.Unavailable {
			guild.Members = g.Members
			guild.Presences = g.Presences
			guild.Channels = g.Channels
			guild.VoiceStates = g.VoiceStates
		}

		*g = *guild
		return nil
	}

	s.Guilds = append(s.Guilds, guild)
	s.guildMap[guild.ID] = guild

	return nil
}

// GuildRemove removes a guild from current world state.
func (s *State) GuildRemove(guild *Guild) error {
	if s == nil {
		return ErrNilState
	}

	_, err := s.Guild(guild.ID)
	if err != nil {
		return err
	}

	s.Lock()
	defer s.Unlock()

	delete(s.guildMap, guild.ID)

	for i, g := range s.Guilds {
		if g.ID == guild.ID {
			s.Guilds = append(s.Guilds[:i], s.Guilds[i+1:]...)
			return nil
		}
	}

	return nil
}

// Guild gets a guild by ID.
// Useful for querying if @me is in a guild:
//     _, err := discordgo.Session.State.Guild(guildID)
//     isInGuild := err == nil
func (s *State) Guild(guildID string) (*Guild, error) {
	if s == nil {
		return nil, ErrNilState
	}

	s.RLock()
	defer s.RUnlock()

	if g, ok := s.guildMap[guildID]; ok {
		return g, nil
	}

	return nil, errors.New("Guild not found.")
}

// TODO: Consider moving Guild state update methods onto *Guild.

// MemberAdd adds a member to the current world state, or
// updates it if it already exists.
func (s *State) MemberAdd(member *Member) error {
	if s == nil {
		return ErrNilState
	}

	guild, err := s.Guild(member.GuildID)
	if err != nil {
		return err
	}

	s.Lock()
	defer s.Unlock()

	for i, m := range guild.Members {
		if m.User.ID == member.User.ID {
			guild.Members[i] = member
			return nil
		}
	}

	guild.Members = append(guild.Members, member)
	return nil
}

// MemberRemove removes a member from current world state.
func (s *State) MemberRemove(member *Member) error {
	if s == nil {
		return ErrNilState
	}

	guild, err := s.Guild(member.GuildID)
	if err != nil {
		return err
	}

	s.Lock()
	defer s.Unlock()

	for i, m := range guild.Members {
		if m.User.ID == member.User.ID {
			guild.Members = append(guild.Members[:i], guild.Members[i+1:]...)
			return nil
		}
	}

	return errors.New("Member not found.")
}

// Member gets a member by ID from a guild.
func (s *State) Member(guildID, userID string) (*Member, error) {
	if s == nil {
		return nil, ErrNilState
	}

	guild, err := s.Guild(guildID)
	if err != nil {
		return nil, err
	}

	s.RLock()
	defer s.RUnlock()

	for _, m := range guild.Members {
		if m.User.ID == userID {
			return m, nil
		}
	}

	return nil, errors.New("Member not found.")
}

// RoleAdd adds a role to the current world state, or
// updates it if it already exists.
func (s *State) RoleAdd(guildID string, role *Role) error {
	if s == nil {
		return ErrNilState
	}

	guild, err := s.Guild(guildID)
	if err != nil {
		return err
	}

	s.Lock()
	defer s.Unlock()

	for i, r := range guild.Roles {
		if r.ID == role.ID {
			guild.Roles[i] = role
			return nil
		}
	}

	guild.Roles = append(guild.Roles, role)
	return nil
}

// RoleRemove removes a role from current world state by ID.
func (s *State) RoleRemove(guildID, roleID string) error {
	if s == nil {
		return ErrNilState
	}

	guild, err := s.Guild(guildID)
	if err != nil {
		return err
	}

	s.Lock()
	defer s.Unlock()

	for i, r := range guild.Roles {
		if r.ID == roleID {
			guild.Roles = append(guild.Roles[:i], guild.Roles[i+1:]...)
			return nil
		}
	}

	return errors.New("Role not found.")
}

// Role gets a role by ID from a guild.
func (s *State) Role(guildID, roleID string) (*Role, error) {
	if s == nil {
		return nil, ErrNilState
	}

	guild, err := s.Guild(guildID)
	if err != nil {
		return nil, err
	}

	s.RLock()
	defer s.RUnlock()

	for _, r := range guild.Roles {
		if r.ID == roleID {
			return r, nil
		}
	}

	return nil, errors.New("Role not found.")
}

// ChannelAdd adds a guild to the current world state, or
// updates it if it already exists.
// Channels may exist either as PrivateChannels or inside
// a guild.
func (s *State) ChannelAdd(channel *Channel) error {
	if s == nil {
		return ErrNilState
	}

	s.Lock()
	defer s.Unlock()

	// If the channel exists, replace it
	if c, ok := s.channelMap[channel.ID]; ok {
		channel.Messages = c.Messages
		channel.PermissionOverwrites = c.PermissionOverwrites

		*c = *channel
		return nil
	}

	if channel.IsPrivate {
		s.PrivateChannels = append(s.PrivateChannels, channel)
	} else {
		guild, ok := s.guildMap[channel.GuildID]
		if !ok {
			return errors.New("Guild for channel not found.")
		}

		guild.Channels = append(guild.Channels, channel)
	}

	s.channelMap[channel.ID] = channel

	return nil
}

// ChannelRemove removes a channel from current world state.
func (s *State) ChannelRemove(channel *Channel) error {
	if s == nil {
		return ErrNilState
	}

	_, err := s.Channel(channel.ID)
	if err != nil {
		return err
	}

	if channel.IsPrivate {
		s.Lock()
		defer s.Unlock()

		for i, c := range s.PrivateChannels {
			if c.ID == channel.ID {
				s.PrivateChannels = append(s.PrivateChannels[:i], s.PrivateChannels[i+1:]...)
				break
			}
		}
	} else {
		guild, err := s.Guild(channel.GuildID)
		if err != nil {
			return err
		}

		s.Lock()
		defer s.Unlock()

		for i, c := range guild.Channels {
			if c.ID == channel.ID {
				guild.Channels = append(guild.Channels[:i], guild.Channels[i+1:]...)
				break
			}
		}
	}

	delete(s.channelMap, channel.ID)

	return nil
}

// GuildChannel gets a channel by ID from a guild.
// This method is Deprecated, use Channel(channelID)
func (s *State) GuildChannel(guildID, channelID string) (*Channel, error) {
	return s.Channel(channelID)
}

// PrivateChannel gets a private channel by ID.
// This method is Deprecated, use Channel(channelID)
func (s *State) PrivateChannel(channelID string) (*Channel, error) {
	return s.Channel(channelID)
}

// Channel gets a channel by ID, it will look in all guilds an private channels.
func (s *State) Channel(channelID string) (*Channel, error) {
	if s == nil {
		return nil, ErrNilState
	}

	s.RLock()
	defer s.RUnlock()

	if c, ok := s.channelMap[channelID]; ok {
		return c, nil
	}

	return nil, errors.New("Channel not found.")
}

// Emoji returns an emoji for a guild and emoji id.
func (s *State) Emoji(guildID, emojiID string) (*Emoji, error) {
	if s == nil {
		return nil, ErrNilState
	}

	guild, err := s.Guild(guildID)
	if err != nil {
		return nil, err
	}

	s.RLock()
	defer s.RUnlock()

	for _, e := range guild.Emojis {
		if e.ID == emojiID {
			return e, nil
		}
	}

	return nil, errors.New("Emoji not found.")
}

// EmojiAdd adds an emoji to the current world state.
func (s *State) EmojiAdd(guildID string, emoji *Emoji) error {
	if s == nil {
		return ErrNilState
	}

	guild, err := s.Guild(guildID)
	if err != nil {
		return err
	}

	s.Lock()
	defer s.Unlock()

	for i, e := range guild.Emojis {
		if e.ID == emoji.ID {
			guild.Emojis[i] = emoji
			return nil
		}
	}

	guild.Emojis = append(guild.Emojis, emoji)
	return nil
}

// EmojisAdd adds multiple emojis to the world state.
func (s *State) EmojisAdd(guildID string, emojis []*Emoji) error {
	for _, e := range emojis {
		if err := s.EmojiAdd(guildID, e); err != nil {
			return err
		}
	}
	return nil
}

// MessageAdd adds a message to the current world state, or updates it if it exists.
// If the channel cannot be found, the message is discarded.
// Messages are kept in state up to s.MaxMessageCount
func (s *State) MessageAdd(message *Message) error {
	if s == nil {
		return ErrNilState
	}

	c, err := s.Channel(message.ChannelID)
	if err != nil {
		return err
	}

	s.Lock()
	defer s.Unlock()

	// If the message exists, merge in the new message contents.
	for _, m := range c.Messages {
		if m.ID == message.ID {
			if message.Content != "" {
				m.Content = message.Content
			}
			if message.EditedTimestamp != "" {
				m.EditedTimestamp = message.EditedTimestamp
			}
			if message.Mentions != nil {
				m.Mentions = message.Mentions
			}
			if message.Embeds != nil {
				m.Embeds = message.Embeds
			}
			if message.Attachments != nil {
				m.Attachments = message.Attachments
			}

			return nil
		}
	}

	c.Messages = append(c.Messages, message)

	if len(c.Messages) > s.MaxMessageCount {
		c.Messages = c.Messages[len(c.Messages)-s.MaxMessageCount:]
	}
	return nil
}

// MessageRemove removes a message from the world state.
func (s *State) MessageRemove(message *Message) error {
	if s == nil {
		return ErrNilState
	}

	c, err := s.Channel(message.ChannelID)
	if err != nil {
		return err
	}

	s.Lock()
	defer s.Unlock()

	for i, m := range c.Messages {
		if m.ID == message.ID {
			c.Messages = append(c.Messages[:i], c.Messages[i+1:]...)
			return nil
		}
	}

	return errors.New("Message not found.")
}

func (s *State) voiceStateUpdate(update *VoiceStateUpdate) error {
	guild, err := s.Guild(update.GuildID)
	if err != nil {
		return err
	}

	s.Lock()
	defer s.Unlock()

	// Handle Leaving Channel
	if update.ChannelID == "" {
		for i, state := range guild.VoiceStates {
			if state.UserID == update.UserID {
				guild.VoiceStates = append(guild.VoiceStates[:i], guild.VoiceStates[i+1:]...)
				return nil
			}
		}
	} else {
		for i, state := range guild.VoiceStates {
			if state.UserID == update.UserID {
				guild.VoiceStates[i] = update.VoiceState
				return nil
			}
		}

		guild.VoiceStates = append(guild.VoiceStates, update.VoiceState)
	}

	return nil
}

// Message gets a message by channel and message ID.
func (s *State) Message(channelID, messageID string) (*Message, error) {
	if s == nil {
		return nil, ErrNilState
	}

	c, err := s.Channel(channelID)
	if err != nil {
		return nil, err
	}

	s.RLock()
	defer s.RUnlock()

	for _, m := range c.Messages {
		if m.ID == messageID {
			return m, nil
		}
	}

	return nil, errors.New("Message not found.")
}

// onInterface handles all events related to states.
func (s *State) onInterface(se *Session, i interface{}) (err error) {
	if s == nil {
		return ErrNilState
	}
	if !se.StateEnabled {
		return nil
	}

	switch t := i.(type) {
	case *Ready:
		err = s.OnReady(t)
	case *GuildCreate:
		err = s.GuildAdd(t.Guild)
	case *GuildUpdate:
		err = s.GuildAdd(t.Guild)
	case *GuildDelete:
		err = s.GuildRemove(t.Guild)
	case *GuildMemberAdd:
		if s.TrackMembers {
			err = s.MemberAdd(t.Member)
		}
	case *GuildMemberUpdate:
		if s.TrackMembers {
			err = s.MemberAdd(t.Member)
		}
	case *GuildMemberRemove:
		if s.TrackMembers {
			err = s.MemberRemove(t.Member)
		}
	case *GuildRoleCreate:
		if s.TrackRoles {
			err = s.RoleAdd(t.GuildID, t.Role)
		}
	case *GuildRoleUpdate:
		if s.TrackRoles {
			err = s.RoleAdd(t.GuildID, t.Role)
		}
	case *GuildRoleDelete:
		if s.TrackRoles {
			err = s.RoleRemove(t.GuildID, t.RoleID)
		}
	case *GuildEmojisUpdate:
		if s.TrackEmojis {
			err = s.EmojisAdd(t.GuildID, t.Emojis)
		}
	case *ChannelCreate:
		if s.TrackChannels {
			err = s.ChannelAdd(t.Channel)
		}
	case *ChannelUpdate:
		if s.TrackChannels {
			err = s.ChannelAdd(t.Channel)
		}
	case *ChannelDelete:
		if s.TrackChannels {
			err = s.ChannelRemove(t.Channel)
		}
	case *MessageCreate:
		if s.MaxMessageCount != 0 {
			err = s.MessageAdd(t.Message)
		}
	case *MessageUpdate:
		if s.MaxMessageCount != 0 {
			err = s.MessageAdd(t.Message)
		}
	case *MessageDelete:
		if s.MaxMessageCount != 0 {
			err = s.MessageRemove(t.Message)
		}
	case *VoiceStateUpdate:
		if s.TrackVoice {
			err = s.voiceStateUpdate(t)
		}
	}

	return
}

// UserChannelPermissions returns the permission of a user in a channel.
// userID    : The ID of the user to calculate permissions for.
// channelID : The ID of the channel to calculate permission for.
func (s *State) UserChannelPermissions(userID, channelID string) (apermissions int, err error) {

	channel, err := s.Channel(channelID)
	if err != nil {
		return
	}

	guild, err := s.Guild(channel.GuildID)
	if err != nil {
		return
	}

	if userID == guild.OwnerID {
		apermissions = PermissionAll
		return
	}

	member, err := s.Member(guild.ID, userID)
	if err != nil {
		return
	}

	for _, role := range guild.Roles {
		for _, roleID := range member.Roles {
			if role.ID == roleID {
				apermissions |= role.Permissions
				break
			}
		}
	}

	if apermissions&PermissionManageRoles > 0 {
		apermissions |= PermissionAll
	}

	// Member overwrites can override role overrides, so do two passes
	for _, overwrite := range channel.PermissionOverwrites {
		for _, roleID := range member.Roles {
			if overwrite.Type == "role" && roleID == overwrite.ID {
				apermissions &= ^overwrite.Deny
				apermissions |= overwrite.Allow
				break
			}
		}
	}

	for _, overwrite := range channel.PermissionOverwrites {
		if overwrite.Type == "member" && overwrite.ID == userID {
			apermissions &= ^overwrite.Deny
			apermissions |= overwrite.Allow
			break
		}
	}

	if apermissions&PermissionManageRoles > 0 {
		apermissions |= PermissionAllChannel
	}

	return
}
