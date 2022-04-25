package discordgo

import (
	"bytes"
	"crypto/ed25519"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"time"
)

// InteractionDeadline is the time allowed to respond to an interaction.
const InteractionDeadline = time.Second * 3

// ApplicationCommandType represents the type of application command.
type ApplicationCommandType uint8

// Application command types
const (
	// ChatApplicationCommand is default command type. They are slash commands (i.e. called directly from the chat).
	ChatApplicationCommand ApplicationCommandType = 1
	// UserApplicationCommand adds command to user context menu.
	UserApplicationCommand ApplicationCommandType = 2
	// MessageApplicationCommand adds command to message context menu.
	MessageApplicationCommand ApplicationCommandType = 3
)

// ApplicationCommand represents an application's slash command.
type ApplicationCommand struct {
	ID                string                 `json:"id,omitempty"`
	ApplicationID     string                 `json:"application_id,omitempty"`
	Version           string                 `json:"version,omitempty"`
	Type              ApplicationCommandType `json:"type,omitempty"`
	Name              string                 `json:"name"`
	NameLocalizations *map[Locale]string     `json:"name_localizations,omitempty"`
	DefaultPermission *bool                  `json:"default_permission,omitempty"`

	// NOTE: Chat commands only. Otherwise it mustn't be set.

	Description              string                      `json:"description,omitempty"`
	DescriptionLocalizations *map[Locale]string          `json:"description_localizations,omitempty"`
	Options                  []*ApplicationCommandOption `json:"options"`
}

// ApplicationCommandOptionType indicates the type of a slash command's option.
type ApplicationCommandOptionType uint8

// Application command option types.
const (
	ApplicationCommandOptionSubCommand      ApplicationCommandOptionType = 1
	ApplicationCommandOptionSubCommandGroup ApplicationCommandOptionType = 2
	ApplicationCommandOptionString          ApplicationCommandOptionType = 3
	ApplicationCommandOptionInteger         ApplicationCommandOptionType = 4
	ApplicationCommandOptionBoolean         ApplicationCommandOptionType = 5
	ApplicationCommandOptionUser            ApplicationCommandOptionType = 6
	ApplicationCommandOptionChannel         ApplicationCommandOptionType = 7
	ApplicationCommandOptionRole            ApplicationCommandOptionType = 8
	ApplicationCommandOptionMentionable     ApplicationCommandOptionType = 9
	ApplicationCommandOptionNumber          ApplicationCommandOptionType = 10
	ApplicationCommandOptionAttachment      ApplicationCommandOptionType = 11
)

func (t ApplicationCommandOptionType) String() string {
	switch t {
	case ApplicationCommandOptionSubCommand:
		return "SubCommand"
	case ApplicationCommandOptionSubCommandGroup:
		return "SubCommandGroup"
	case ApplicationCommandOptionString:
		return "String"
	case ApplicationCommandOptionInteger:
		return "Integer"
	case ApplicationCommandOptionBoolean:
		return "Boolean"
	case ApplicationCommandOptionUser:
		return "User"
	case ApplicationCommandOptionChannel:
		return "Channel"
	case ApplicationCommandOptionRole:
		return "Role"
	case ApplicationCommandOptionMentionable:
		return "Mentionable"
	case ApplicationCommandOptionNumber:
		return "Number"
	case ApplicationCommandOptionAttachment:
		return "Attachment"
	}
	return fmt.Sprintf("ApplicationCommandOptionType(%d)", t)
}

// ApplicationCommandOption represents an option/subcommand/subcommands group.
type ApplicationCommandOption struct {
	Type                     ApplicationCommandOptionType `json:"type"`
	Name                     string                       `json:"name"`
	NameLocalizations        map[Locale]string            `json:"name_localizations,omitempty"`
	Description              string                       `json:"description,omitempty"`
	DescriptionLocalizations map[Locale]string            `json:"description_localizations,omitempty"`
	// NOTE: This feature was on the API, but at some point developers decided to remove it.
	// So I commented it, until it will be officially on the docs.
	// Default     bool                              `json:"default"`

	ChannelTypes []ChannelType               `json:"channel_types"`
	Required     bool                        `json:"required"`
	Options      []*ApplicationCommandOption `json:"options"`

	// NOTE: mutually exclusive with Choices.
	Autocomplete bool                              `json:"autocomplete"`
	Choices      []*ApplicationCommandOptionChoice `json:"choices"`
	// Minimal value of number/integer option.
	MinValue *float64 `json:"min_value,omitempty"`
	// Maximum value of number/integer option.
	MaxValue float64 `json:"max_value,omitempty"`
}

// ApplicationCommandOptionChoice represents a slash command option choice.
type ApplicationCommandOptionChoice struct {
	Name              string            `json:"name"`
	NameLocalizations map[Locale]string `json:"name_localizations,omitempty"`
	Value             interface{}       `json:"value"`
}

// ApplicationCommandPermissions represents a single user or role permission for a command.
type ApplicationCommandPermissions struct {
	ID         string                           `json:"id"`
	Type       ApplicationCommandPermissionType `json:"type"`
	Permission bool                             `json:"permission"`
}

// ApplicationCommandPermissionsList represents a list of ApplicationCommandPermissions, needed for serializing to JSON.
type ApplicationCommandPermissionsList struct {
	Permissions []*ApplicationCommandPermissions `json:"permissions"`
}

// GuildApplicationCommandPermissions represents all permissions for a single guild command.
type GuildApplicationCommandPermissions struct {
	ID            string                           `json:"id"`
	ApplicationID string                           `json:"application_id"`
	GuildID       string                           `json:"guild_id"`
	Permissions   []*ApplicationCommandPermissions `json:"permissions"`
}

// ApplicationCommandPermissionType indicates whether a permission is user or role based.
type ApplicationCommandPermissionType uint8

// Application command permission types.
const (
	ApplicationCommandPermissionTypeRole ApplicationCommandPermissionType = 1
	ApplicationCommandPermissionTypeUser ApplicationCommandPermissionType = 2
)

// InteractionType indicates the type of an interaction event.
type InteractionType uint8

// Interaction types
const (
	InteractionPing                           InteractionType = 1
	InteractionApplicationCommand             InteractionType = 2
	InteractionMessageComponent               InteractionType = 3
	InteractionApplicationCommandAutocomplete InteractionType = 4
	InteractionModalSubmit                    InteractionType = 5
)

func (t InteractionType) String() string {
	switch t {
	case InteractionPing:
		return "Ping"
	case InteractionApplicationCommand:
		return "ApplicationCommand"
	case InteractionMessageComponent:
		return "MessageComponent"
	case InteractionModalSubmit:
		return "ModalSubmit"
	}
	return fmt.Sprintf("InteractionType(%d)", t)
}

// Interaction represents data of an interaction.
type Interaction struct {
	ID        string          `json:"id"`
	AppID     string          `json:"application_id"`
	Type      InteractionType `json:"type"`
	Data      InteractionData `json:"data"`
	GuildID   string          `json:"guild_id"`
	ChannelID string          `json:"channel_id"`

	// The message on which interaction was used.
	// NOTE: this field is only filled when a button click triggered the interaction. Otherwise it will be nil.
	Message *Message `json:"message"`

	// The member who invoked this interaction.
	// NOTE: this field is only filled when the slash command was invoked in a guild;
	// if it was invoked in a DM, the `User` field will be filled instead.
	// Make sure to check for `nil` before using this field.
	Member *Member `json:"member"`
	// The user who invoked this interaction.
	// NOTE: this field is only filled when the slash command was invoked in a DM;
	// if it was invoked in a guild, the `Member` field will be filled instead.
	// Make sure to check for `nil` before using this field.
	User *User `json:"user"`

	// The user's discord client locale.
	Locale Locale `json:"locale"`
	// The guild's locale. This defaults to EnglishUS
	// NOTE: this field is only filled when the interaction was invoked in a guild.
	GuildLocale *Locale `json:"guild_locale"`

	Token   string `json:"token"`
	Version int    `json:"version"`
}

type interaction Interaction

type rawInteraction struct {
	interaction
	Data json.RawMessage `json:"data"`
}

// UnmarshalJSON is a method for unmarshalling JSON object to Interaction.
func (i *Interaction) UnmarshalJSON(raw []byte) error {
	var tmp rawInteraction
	err := json.Unmarshal(raw, &tmp)
	if err != nil {
		return err
	}

	*i = Interaction(tmp.interaction)

	switch tmp.Type {
	case InteractionApplicationCommand, InteractionApplicationCommandAutocomplete:
		v := ApplicationCommandInteractionData{}
		err = json.Unmarshal(tmp.Data, &v)
		if err != nil {
			return err
		}
		i.Data = v
	case InteractionMessageComponent:
		v := MessageComponentInteractionData{}
		err = json.Unmarshal(tmp.Data, &v)
		if err != nil {
			return err
		}
		i.Data = v
	case InteractionModalSubmit:
		v := ModalSubmitInteractionData{}
		err = json.Unmarshal(tmp.Data, &v)
		if err != nil {
			return err
		}
		i.Data = v
	}
	return nil
}

// MessageComponentData is helper function to assert the inner InteractionData to MessageComponentInteractionData.
// Make sure to check that the Type of the interaction is InteractionMessageComponent before calling.
func (i Interaction) MessageComponentData() (data MessageComponentInteractionData) {
	if i.Type != InteractionMessageComponent {
		panic("MessageComponentData called on interaction of type " + i.Type.String())
	}
	return i.Data.(MessageComponentInteractionData)
}

// ApplicationCommandData is helper function to assert the inner InteractionData to ApplicationCommandInteractionData.
// Make sure to check that the Type of the interaction is InteractionApplicationCommand before calling.
func (i Interaction) ApplicationCommandData() (data ApplicationCommandInteractionData) {
	if i.Type != InteractionApplicationCommand && i.Type != InteractionApplicationCommandAutocomplete {
		panic("ApplicationCommandData called on interaction of type " + i.Type.String())
	}
	return i.Data.(ApplicationCommandInteractionData)
}

// ModalSubmitData is helper function to assert the inner InteractionData to ModalSubmitInteractionData.
// Make sure to check that the Type of the interaction is InteractionModalSubmit before calling.
func (i Interaction) ModalSubmitData() (data ModalSubmitInteractionData) {
	if i.Type != InteractionModalSubmit {
		panic("ModalSubmitData called on interaction of type " + i.Type.String())
	}
	return i.Data.(ModalSubmitInteractionData)
}

// InteractionData is a common interface for all types of interaction data.
type InteractionData interface {
	Type() InteractionType
}

// ApplicationCommandInteractionData contains the data of application command interaction.
type ApplicationCommandInteractionData struct {
	ID       string                                     `json:"id"`
	Name     string                                     `json:"name"`
	Resolved *ApplicationCommandInteractionDataResolved `json:"resolved"`

	// Slash command options
	Options []*ApplicationCommandInteractionDataOption `json:"options"`
	// Target (user/message) id on which context menu command was called.
	// The details are stored in Resolved according to command type.
	TargetID string `json:"target_id"`
}

// ApplicationCommandInteractionDataResolved contains resolved data of command execution.
// Partial Member objects are missing user, deaf and mute fields.
// Partial Channel objects only have id, name, type and permissions fields.
type ApplicationCommandInteractionDataResolved struct {
	Users       map[string]*User              `json:"users"`
	Members     map[string]*Member            `json:"members"`
	Roles       map[string]*Role              `json:"roles"`
	Channels    map[string]*Channel           `json:"channels"`
	Messages    map[string]*Message           `json:"messages"`
	Attachments map[string]*MessageAttachment `json:"attachments"`
}

// Type returns the type of interaction data.
func (ApplicationCommandInteractionData) Type() InteractionType {
	return InteractionApplicationCommand
}

// MessageComponentInteractionData contains the data of message component interaction.
type MessageComponentInteractionData struct {
	CustomID      string        `json:"custom_id"`
	ComponentType ComponentType `json:"component_type"`

	// NOTE: Only filled when ComponentType is SelectMenuComponent (3). Otherwise is nil.
	Values []string `json:"values"`
}

// Type returns the type of interaction data.
func (MessageComponentInteractionData) Type() InteractionType {
	return InteractionMessageComponent
}

// ModalSubmitInteractionData contains the data of modal submit interaction.
type ModalSubmitInteractionData struct {
	CustomID   string             `json:"custom_id"`
	Components []MessageComponent `json:"-"`
}

// Type returns the type of interaction data.
func (ModalSubmitInteractionData) Type() InteractionType {
	return InteractionModalSubmit
}

// UnmarshalJSON is a helper function to correctly unmarshal Components.
func (d *ModalSubmitInteractionData) UnmarshalJSON(data []byte) error {
	type modalSubmitInteractionData ModalSubmitInteractionData
	var v struct {
		modalSubmitInteractionData
		RawComponents []unmarshalableMessageComponent `json:"components"`
	}
	err := json.Unmarshal(data, &v)
	if err != nil {
		return err
	}
	*d = ModalSubmitInteractionData(v.modalSubmitInteractionData)
	d.Components = make([]MessageComponent, len(v.RawComponents))
	for i, v := range v.RawComponents {
		d.Components[i] = v.MessageComponent
	}
	return err
}

// ApplicationCommandInteractionDataOption represents an option of a slash command.
type ApplicationCommandInteractionDataOption struct {
	Name string                       `json:"name"`
	Type ApplicationCommandOptionType `json:"type"`
	// NOTE: Contains the value specified by Type.
	Value   interface{}                                `json:"value,omitempty"`
	Options []*ApplicationCommandInteractionDataOption `json:"options,omitempty"`

	// NOTE: autocomplete interaction only.
	Focused bool `json:"focused,omitempty"`
}

// IntValue is a utility function for casting option value to integer
func (o ApplicationCommandInteractionDataOption) IntValue() int64 {
	if o.Type != ApplicationCommandOptionInteger {
		panic("IntValue called on data option of type " + o.Type.String())
	}
	return int64(o.Value.(float64))
}

// UintValue is a utility function for casting option value to unsigned integer
func (o ApplicationCommandInteractionDataOption) UintValue() uint64 {
	if o.Type != ApplicationCommandOptionInteger {
		panic("UintValue called on data option of type " + o.Type.String())
	}
	return uint64(o.Value.(float64))
}

// FloatValue is a utility function for casting option value to float
func (o ApplicationCommandInteractionDataOption) FloatValue() float64 {
	if o.Type != ApplicationCommandOptionNumber {
		panic("FloatValue called on data option of type " + o.Type.String())
	}
	return o.Value.(float64)
}

// StringValue is a utility function for casting option value to string
func (o ApplicationCommandInteractionDataOption) StringValue() string {
	if o.Type != ApplicationCommandOptionString {
		panic("StringValue called on data option of type " + o.Type.String())
	}
	return o.Value.(string)
}

// BoolValue is a utility function for casting option value to bool
func (o ApplicationCommandInteractionDataOption) BoolValue() bool {
	if o.Type != ApplicationCommandOptionBoolean {
		panic("BoolValue called on data option of type " + o.Type.String())
	}
	return o.Value.(bool)
}

// ChannelValue is a utility function for casting option value to channel object.
// s : Session object, if not nil, function additionally fetches all channel's data
func (o ApplicationCommandInteractionDataOption) ChannelValue(s *Session) *Channel {
	if o.Type != ApplicationCommandOptionChannel {
		panic("ChannelValue called on data option of type " + o.Type.String())
	}
	chanID := o.Value.(string)

	if s == nil {
		return &Channel{ID: chanID}
	}

	ch, err := s.State.Channel(chanID)
	if err != nil {
		ch, err = s.Channel(chanID)
		if err != nil {
			return &Channel{ID: chanID}
		}
	}

	return ch
}

// RoleValue is a utility function for casting option value to role object.
// s : Session object, if not nil, function additionally fetches all role's data
func (o ApplicationCommandInteractionDataOption) RoleValue(s *Session, gID string) *Role {
	if o.Type != ApplicationCommandOptionRole && o.Type != ApplicationCommandOptionMentionable {
		panic("RoleValue called on data option of type " + o.Type.String())
	}
	roleID := o.Value.(string)

	if s == nil || gID == "" {
		return &Role{ID: roleID}
	}

	r, err := s.State.Role(roleID, gID)
	if err != nil {
		roles, err := s.GuildRoles(gID)
		if err == nil {
			for _, r = range roles {
				if r.ID == roleID {
					return r
				}
			}
		}
		return &Role{ID: roleID}
	}

	return r
}

// UserValue is a utility function for casting option value to user object.
// s : Session object, if not nil, function additionally fetches all user's data
func (o ApplicationCommandInteractionDataOption) UserValue(s *Session) *User {
	if o.Type != ApplicationCommandOptionUser && o.Type != ApplicationCommandOptionMentionable {
		panic("UserValue called on data option of type " + o.Type.String())
	}
	userID := o.Value.(string)

	if s == nil {
		return &User{ID: userID}
	}

	u, err := s.User(userID)
	if err != nil {
		return &User{ID: userID}
	}

	return u
}

// InteractionResponseType is type of interaction response.
type InteractionResponseType uint8

// Interaction response types.
const (
	// InteractionResponsePong is for ACK ping event.
	InteractionResponsePong InteractionResponseType = 1
	// InteractionResponseChannelMessageWithSource is for responding with a message, showing the user's input.
	InteractionResponseChannelMessageWithSource InteractionResponseType = 4
	// InteractionResponseDeferredChannelMessageWithSource acknowledges that the event was received, and that a follow-up will come later.
	InteractionResponseDeferredChannelMessageWithSource InteractionResponseType = 5
	// InteractionResponseDeferredMessageUpdate acknowledges that the message component interaction event was received, and message will be updated later.
	InteractionResponseDeferredMessageUpdate InteractionResponseType = 6
	// InteractionResponseUpdateMessage is for updating the message to which message component was attached.
	InteractionResponseUpdateMessage InteractionResponseType = 7
	// InteractionApplicationCommandAutocompleteResult shows autocompletion results. Autocomplete interaction only.
	InteractionApplicationCommandAutocompleteResult InteractionResponseType = 8
	// InteractionResponseModal is for responding to an interaction with a modal window.
	InteractionResponseModal InteractionResponseType = 9
)

// InteractionResponse represents a response for an interaction event.
type InteractionResponse struct {
	Type InteractionResponseType  `json:"type,omitempty"`
	Data *InteractionResponseData `json:"data,omitempty"`
}

// InteractionResponseData is response data for an interaction.
type InteractionResponseData struct {
	TTS             bool                    `json:"tts"`
	Content         string                  `json:"content"`
	Components      []MessageComponent      `json:"components"`
	Embeds          []*MessageEmbed         `json:"embeds"`
	AllowedMentions *MessageAllowedMentions `json:"allowed_mentions,omitempty"`
	Flags           uint64                  `json:"flags,omitempty"`
	Files           []*File                 `json:"-"`

	// NOTE: autocomplete interaction only.
	Choices []*ApplicationCommandOptionChoice `json:"choices,omitempty"`

	// NOTE: modal interaction only.

	CustomID string `json:"custom_id,omitempty"`
	Title    string `json:"title,omitempty"`
}

// VerifyInteraction implements message verification of the discord interactions api
// signing algorithm, as documented here:
// https://discord.com/developers/docs/interactions/receiving-and-responding#security-and-authorization
func VerifyInteraction(r *http.Request, key ed25519.PublicKey) bool {
	var msg bytes.Buffer

	signature := r.Header.Get("X-Signature-Ed25519")
	if signature == "" {
		return false
	}

	sig, err := hex.DecodeString(signature)
	if err != nil {
		return false
	}

	if len(sig) != ed25519.SignatureSize {
		return false
	}

	timestamp := r.Header.Get("X-Signature-Timestamp")
	if timestamp == "" {
		return false
	}

	msg.WriteString(timestamp)

	defer r.Body.Close()
	var body bytes.Buffer

	// at the end of the function, copy the original body back into the request
	defer func() {
		r.Body = ioutil.NopCloser(&body)
	}()

	// copy body into buffers
	_, err = io.Copy(&msg, io.TeeReader(r.Body, &body))
	if err != nil {
		return false
	}

	return ed25519.Verify(key, msg.Bytes(), sig)
}
