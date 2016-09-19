package discordgo

// eventToInterface is a mapping of Discord WSAPI events to their
// DiscordGo event container.
// Each Discord WSAPI event maps to a unique interface.
// Use Session.AddHandler with one of these types to handle that
// type of event.
// eg:
//     Session.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
//     })
//
// or:
//     Session.AddHandler(func(s *discordgo.Session, m *discordgo.PresenceUpdate) {
//     })
var eventToInterface = map[string]interface{}{
	"CHANNEL_CREATE":             ChannelCreate{},
	"CHANNEL_UPDATE":             ChannelUpdate{},
	"CHANNEL_DELETE":             ChannelDelete{},
	"GUILD_CREATE":               GuildCreate{},
	"GUILD_UPDATE":               GuildUpdate{},
	"GUILD_DELETE":               GuildDelete{},
	"GUILD_BAN_ADD":              GuildBanAdd{},
	"GUILD_BAN_REMOVE":           GuildBanRemove{},
	"GUILD_MEMBER_ADD":           GuildMemberAdd{},
	"GUILD_MEMBER_UPDATE":        GuildMemberUpdate{},
	"GUILD_MEMBER_REMOVE":        GuildMemberRemove{},
	"GUILD_ROLE_CREATE":          GuildRoleCreate{},
	"GUILD_ROLE_UPDATE":          GuildRoleUpdate{},
	"GUILD_ROLE_DELETE":          GuildRoleDelete{},
	"GUILD_INTEGRATIONS_UPDATE":  GuildIntegrationsUpdate{},
	"GUILD_EMOJIS_UPDATE":        GuildEmojisUpdate{},
	"MESSAGE_ACK":                MessageAck{},
	"MESSAGE_CREATE":             MessageCreate{},
	"MESSAGE_UPDATE":             MessageUpdate{},
	"MESSAGE_DELETE":             MessageDelete{},
	"PRESENCE_UPDATE":            PresenceUpdate{},
	"PRESENCES_REPLACE":          PresencesReplace{},
	"READY":                      Ready{},
	"USER_UPDATE":                UserUpdate{},
	"USER_SETTINGS_UPDATE":       UserSettingsUpdate{},
	"USER_GUILD_SETTINGS_UPDATE": UserGuildSettingsUpdate{},
	"TYPING_START":               TypingStart{},
	"VOICE_SERVER_UPDATE":        VoiceServerUpdate{},
	"VOICE_STATE_UPDATE":         VoiceStateUpdate{},
	"RESUMED":                    Resumed{},
}

// Connect is an empty struct for an event.
type Connect struct{}

// Disconnect is an empty struct for an event.
type Disconnect struct{}

// RateLimit is a struct for the RateLimited event
type RateLimit struct {
	*TooManyRequests
	URL string
}

// MessageCreate is a wrapper struct for an event.
type MessageCreate struct {
	*Message
}

// MessageUpdate is a wrapper struct for an event.
type MessageUpdate struct {
	*Message
}

// MessageDelete is a wrapper struct for an event.
type MessageDelete struct {
	*Message
}

// ChannelCreate is a wrapper struct for an event.
type ChannelCreate struct {
	*Channel
}

// ChannelUpdate is a wrapper struct for an event.
type ChannelUpdate struct {
	*Channel
}

// ChannelDelete is a wrapper struct for an event.
type ChannelDelete struct {
	*Channel
}

// GuildCreate is a wrapper struct for an event.
type GuildCreate struct {
	*Guild
}

// GuildUpdate is a wrapper struct for an event.
type GuildUpdate struct {
	*Guild
}

// GuildDelete is a wrapper struct for an event.
type GuildDelete struct {
	*Guild
}

// GuildBanAdd is a wrapper struct for an event.
type GuildBanAdd struct {
	*GuildBan
}

// GuildBanRemove is a wrapper struct for an event.
type GuildBanRemove struct {
	*GuildBan
}

// GuildMemberAdd is a wrapper struct for an event.
type GuildMemberAdd struct {
	*Member
}

// GuildMemberUpdate is a wrapper struct for an event.
type GuildMemberUpdate struct {
	*Member
}

// GuildMemberRemove is a wrapper struct for an event.
type GuildMemberRemove struct {
	*Member
}

// GuildRoleCreate is a wrapper struct for an event.
type GuildRoleCreate struct {
	*GuildRole
}

// GuildRoleUpdate is a wrapper struct for an event.
type GuildRoleUpdate struct {
	*GuildRole
}

// PresencesReplace is an array of Presences for an event.
type PresencesReplace []*Presence

// VoiceStateUpdate is a wrapper struct for an event.
type VoiceStateUpdate struct {
	*VoiceState
}

// UserUpdate is a wrapper struct for an event.
type UserUpdate struct {
	*User
}

// UserSettingsUpdate is a map for an event.
type UserSettingsUpdate map[string]interface{}

// UserGuildSettingsUpdate is a map for an event.
type UserGuildSettingsUpdate struct {
	*UserGuildSettings
}
