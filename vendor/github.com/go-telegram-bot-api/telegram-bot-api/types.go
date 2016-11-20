package tgbotapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"
)

// APIResponse is a response from the Telegram API with the result
// stored raw.
type APIResponse struct {
	Ok          bool            `json:"ok"`
	Result      json.RawMessage `json:"result"`
	ErrorCode   int             `json:"error_code"`
	Description string          `json:"description"`
}

// Update is an update response, from GetUpdates.
type Update struct {
	UpdateID           int                 `json:"update_id"`
	Message            *Message            `json:"message"`
	EditedMessage      *Message            `json:"edited_message"`
	InlineQuery        *InlineQuery        `json:"inline_query"`
	ChosenInlineResult *ChosenInlineResult `json:"chosen_inline_result"`
	CallbackQuery      *CallbackQuery      `json:"callback_query"`
}

// User is a user on Telegram.
type User struct {
	ID        int    `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"` // optional
	UserName  string `json:"username"`  // optional
}

// String displays a simple text version of a user.
//
// It is normally a user's username, but falls back to a first/last
// name as available.
func (u *User) String() string {
	if u.UserName != "" {
		return u.UserName
	}

	name := u.FirstName
	if u.LastName != "" {
		name += " " + u.LastName
	}

	return name
}

// GroupChat is a group chat.
type GroupChat struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
}

// Chat contains information about the place a message was sent.
type Chat struct {
	ID        int64  `json:"id"`
	Type      string `json:"type"`
	Title     string `json:"title"`      // optional
	UserName  string `json:"username"`   // optional
	FirstName string `json:"first_name"` // optional
	LastName  string `json:"last_name"`  // optional
}

// IsPrivate returns if the Chat is a private conversation.
func (c Chat) IsPrivate() bool {
	return c.Type == "private"
}

// IsGroup returns if the Chat is a group.
func (c Chat) IsGroup() bool {
	return c.Type == "group"
}

// IsSuperGroup returns if the Chat is a supergroup.
func (c Chat) IsSuperGroup() bool {
	return c.Type == "supergroup"
}

// IsChannel returns if the Chat is a channel.
func (c Chat) IsChannel() bool {
	return c.Type == "channel"
}

// ChatConfig returns a ChatConfig struct for chat related methods.
func (c Chat) ChatConfig() ChatConfig {
	return ChatConfig{ChatID: c.ID}
}

// Message is returned by almost every request, and contains data about
// almost anything.
type Message struct {
	MessageID             int              `json:"message_id"`
	From                  *User            `json:"from"` // optional
	Date                  int              `json:"date"`
	Chat                  *Chat            `json:"chat"`
	ForwardFrom           *User            `json:"forward_from"`            // optional
	ForwardFromChat       *Chat            `json:"forward_from_chat"`       // optional
	ForwardDate           int              `json:"forward_date"`            // optional
	ReplyToMessage        *Message         `json:"reply_to_message"`        // optional
	EditDate              int              `json:"edit_date"`               // optional
	Text                  string           `json:"text"`                    // optional
	Entities              *[]MessageEntity `json:"entities"`                // optional
	Audio                 *Audio           `json:"audio"`                   // optional
	Document              *Document        `json:"document"`                // optional
	Photo                 *[]PhotoSize     `json:"photo"`                   // optional
	Sticker               *Sticker         `json:"sticker"`                 // optional
	Video                 *Video           `json:"video"`                   // optional
	Voice                 *Voice           `json:"voice"`                   // optional
	Caption               string           `json:"caption"`                 // optional
	Contact               *Contact         `json:"contact"`                 // optional
	Location              *Location        `json:"location"`                // optional
	Venue                 *Venue           `json:"venue"`                   // optional
	NewChatMember         *User            `json:"new_chat_member"`         // optional
	LeftChatMember        *User            `json:"left_chat_member"`        // optional
	NewChatTitle          string           `json:"new_chat_title"`          // optional
	NewChatPhoto          *[]PhotoSize     `json:"new_chat_photo"`          // optional
	DeleteChatPhoto       bool             `json:"delete_chat_photo"`       // optional
	GroupChatCreated      bool             `json:"group_chat_created"`      // optional
	SuperGroupChatCreated bool             `json:"supergroup_chat_created"` // optional
	ChannelChatCreated    bool             `json:"channel_chat_created"`    // optional
	MigrateToChatID       int64            `json:"migrate_to_chat_id"`      // optional
	MigrateFromChatID     int64            `json:"migrate_from_chat_id"`    // optional
	PinnedMessage         *Message         `json:"pinned_message"`          // optional
}

// Time converts the message timestamp into a Time.
func (m *Message) Time() time.Time {
	return time.Unix(int64(m.Date), 0)
}

// IsCommand returns true if message starts with '/'.
func (m *Message) IsCommand() bool {
	return m.Text != "" && m.Text[0] == '/'
}

// Command checks if the message was a command and if it was, returns the
// command. If the Message was not a command, it returns an empty string.
//
// If the command contains the at bot syntax, it removes the bot name.
func (m *Message) Command() string {
	if !m.IsCommand() {
		return ""
	}

	command := strings.SplitN(m.Text, " ", 2)[0][1:]

	if i := strings.Index(command, "@"); i != -1 {
		command = command[:i]
	}

	return command
}

// CommandArguments checks if the message was a command and if it was,
// returns all text after the command name. If the Message was not a
// command, it returns an empty string.
func (m *Message) CommandArguments() string {
	if !m.IsCommand() {
		return ""
	}

	split := strings.SplitN(m.Text, " ", 2)
	if len(split) != 2 {
		return ""
	}

	return strings.SplitN(m.Text, " ", 2)[1]
}

// MessageEntity contains information about data in a Message.
type MessageEntity struct {
	Type   string `json:"type"`
	Offset int    `json:"offset"`
	Length int    `json:"length"`
	URL    string `json:"url"`  // optional
	User   *User  `json:"user"` // optional
}

// ParseURL attempts to parse a URL contained within a MessageEntity.
func (entity MessageEntity) ParseURL() (*url.URL, error) {
	if entity.URL == "" {
		return nil, errors.New(ErrBadURL)
	}

	return url.Parse(entity.URL)
}

// PhotoSize contains information about photos.
type PhotoSize struct {
	FileID   string `json:"file_id"`
	Width    int    `json:"width"`
	Height   int    `json:"height"`
	FileSize int    `json:"file_size"` // optional
}

// Audio contains information about audio.
type Audio struct {
	FileID    string `json:"file_id"`
	Duration  int    `json:"duration"`
	Performer string `json:"performer"` // optional
	Title     string `json:"title"`     // optional
	MimeType  string `json:"mime_type"` // optional
	FileSize  int    `json:"file_size"` // optional
}

// Document contains information about a document.
type Document struct {
	FileID    string     `json:"file_id"`
	Thumbnail *PhotoSize `json:"thumb"`     // optional
	FileName  string     `json:"file_name"` // optional
	MimeType  string     `json:"mime_type"` // optional
	FileSize  int        `json:"file_size"` // optional
}

// Sticker contains information about a sticker.
type Sticker struct {
	FileID    string     `json:"file_id"`
	Width     int        `json:"width"`
	Height    int        `json:"height"`
	Thumbnail *PhotoSize `json:"thumb"`     // optional
	Emoji     string     `json:"emoji"`     // optional
	FileSize  int        `json:"file_size"` // optional
}

// Video contains information about a video.
type Video struct {
	FileID    string     `json:"file_id"`
	Width     int        `json:"width"`
	Height    int        `json:"height"`
	Duration  int        `json:"duration"`
	Thumbnail *PhotoSize `json:"thumb"`     // optional
	MimeType  string     `json:"mime_type"` // optional
	FileSize  int        `json:"file_size"` // optional
}

// Voice contains information about a voice.
type Voice struct {
	FileID   string `json:"file_id"`
	Duration int    `json:"duration"`
	MimeType string `json:"mime_type"` // optional
	FileSize int    `json:"file_size"` // optional
}

// Contact contains information about a contact.
//
// Note that LastName and UserID may be empty.
type Contact struct {
	PhoneNumber string `json:"phone_number"`
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"` // optional
	UserID      int    `json:"user_id"`   // optional
}

// Location contains information about a place.
type Location struct {
	Longitude float64 `json:"longitude"`
	Latitude  float64 `json:"latitude"`
}

// Venue contains information about a venue, including its Location.
type Venue struct {
	Location     Location `json:"location"`
	Title        string   `json:"title"`
	Address      string   `json:"address"`
	FoursquareID string   `json:"foursquare_id"` // optional
}

// UserProfilePhotos contains a set of user profile photos.
type UserProfilePhotos struct {
	TotalCount int           `json:"total_count"`
	Photos     [][]PhotoSize `json:"photos"`
}

// File contains information about a file to download from Telegram.
type File struct {
	FileID   string `json:"file_id"`
	FileSize int    `json:"file_size"` // optional
	FilePath string `json:"file_path"` // optional
}

// Link returns a full path to the download URL for a File.
//
// It requires the Bot Token to create the link.
func (f *File) Link(token string) string {
	return fmt.Sprintf(FileEndpoint, token, f.FilePath)
}

// ReplyKeyboardMarkup allows the Bot to set a custom keyboard.
type ReplyKeyboardMarkup struct {
	Keyboard        [][]KeyboardButton `json:"keyboard"`
	ResizeKeyboard  bool               `json:"resize_keyboard"`   // optional
	OneTimeKeyboard bool               `json:"one_time_keyboard"` // optional
	Selective       bool               `json:"selective"`         // optional
}

// KeyboardButton is a button within a custom keyboard.
type KeyboardButton struct {
	Text            string `json:"text"`
	RequestContact  bool   `json:"request_contact"`
	RequestLocation bool   `json:"request_location"`
}

// ReplyKeyboardHide allows the Bot to hide a custom keyboard.
type ReplyKeyboardHide struct {
	HideKeyboard bool `json:"hide_keyboard"`
	Selective    bool `json:"selective"` // optional
}

// InlineKeyboardMarkup is a custom keyboard presented for an inline bot.
type InlineKeyboardMarkup struct {
	InlineKeyboard [][]InlineKeyboardButton `json:"inline_keyboard"`
}

// InlineKeyboardButton is a button within a custom keyboard for
// inline query responses.
//
// Note that some values are references as even an empty string
// will change behavior.
type InlineKeyboardButton struct {
	Text              string  `json:"text"`
	URL               *string `json:"url,omitempty"`                 // optional
	CallbackData      *string `json:"callback_data,omitempty"`       // optional
	SwitchInlineQuery *string `json:"switch_inline_query,omitempty"` // optional
}

// CallbackQuery is data sent when a keyboard button with callback data
// is clicked.
type CallbackQuery struct {
	ID              string   `json:"id"`
	From            *User    `json:"from"`
	Message         *Message `json:"message"`           // optional
	InlineMessageID string   `json:"inline_message_id"` // optional
	Data            string   `json:"data"`              // optional
}

// ForceReply allows the Bot to have users directly reply to it without
// additional interaction.
type ForceReply struct {
	ForceReply bool `json:"force_reply"`
	Selective  bool `json:"selective"` // optional
}

// ChatMember is information about a member in a chat.
type ChatMember struct {
	User   *User  `json:"user"`
	Status string `json:"status"`
}

// IsCreator returns if the ChatMember was the creator of the chat.
func (chat ChatMember) IsCreator() bool { return chat.Status == "creator" }

// IsAdministrator returns if the ChatMember is a chat administrator.
func (chat ChatMember) IsAdministrator() bool { return chat.Status == "administrator" }

// IsMember returns if the ChatMember is a current member of the chat.
func (chat ChatMember) IsMember() bool { return chat.Status == "member" }

// HasLeft returns if the ChatMember left the chat.
func (chat ChatMember) HasLeft() bool { return chat.Status == "left" }

// WasKicked returns if the ChatMember was kicked from the chat.
func (chat ChatMember) WasKicked() bool { return chat.Status == "kicked" }

// InlineQuery is a Query from Telegram for an inline request.
type InlineQuery struct {
	ID       string    `json:"id"`
	From     *User     `json:"from"`
	Location *Location `json:"location"` // optional
	Query    string    `json:"query"`
	Offset   string    `json:"offset"`
}

// InlineQueryResultArticle is an inline query response article.
type InlineQueryResultArticle struct {
	Type                string                `json:"type"`                            // required
	ID                  string                `json:"id"`                              // required
	Title               string                `json:"title"`                           // required
	InputMessageContent interface{}           `json:"input_message_content,omitempty"` // required
	ReplyMarkup         *InlineKeyboardMarkup `json:"reply_markup,omitempty"`
	URL                 string                `json:"url"`
	HideURL             bool                  `json:"hide_url"`
	Description         string                `json:"description"`
	ThumbURL            string                `json:"thumb_url"`
	ThumbWidth          int                   `json:"thumb_width"`
	ThumbHeight         int                   `json:"thumb_height"`
}

// InlineQueryResultPhoto is an inline query response photo.
type InlineQueryResultPhoto struct {
	Type                string                `json:"type"`      // required
	ID                  string                `json:"id"`        // required
	URL                 string                `json:"photo_url"` // required
	MimeType            string                `json:"mime_type"`
	Width               int                   `json:"photo_width"`
	Height              int                   `json:"photo_height"`
	ThumbURL            string                `json:"thumb_url"`
	Title               string                `json:"title"`
	Description         string                `json:"description"`
	Caption             string                `json:"caption"`
	ReplyMarkup         *InlineKeyboardMarkup `json:"reply_markup,omitempty"`
	InputMessageContent interface{}           `json:"input_message_content,omitempty"`
}

// InlineQueryResultGIF is an inline query response GIF.
type InlineQueryResultGIF struct {
	Type                string                `json:"type"`    // required
	ID                  string                `json:"id"`      // required
	URL                 string                `json:"gif_url"` // required
	Width               int                   `json:"gif_width"`
	Height              int                   `json:"gif_height"`
	ThumbURL            string                `json:"thumb_url"`
	Title               string                `json:"title"`
	Caption             string                `json:"caption"`
	ReplyMarkup         *InlineKeyboardMarkup `json:"reply_markup,omitempty"`
	InputMessageContent interface{}           `json:"input_message_content,omitempty"`
}

// InlineQueryResultMPEG4GIF is an inline query response MPEG4 GIF.
type InlineQueryResultMPEG4GIF struct {
	Type                string                `json:"type"`      // required
	ID                  string                `json:"id"`        // required
	URL                 string                `json:"mpeg4_url"` // required
	Width               int                   `json:"mpeg4_width"`
	Height              int                   `json:"mpeg4_height"`
	ThumbURL            string                `json:"thumb_url"`
	Title               string                `json:"title"`
	Caption             string                `json:"caption"`
	ReplyMarkup         *InlineKeyboardMarkup `json:"reply_markup,omitempty"`
	InputMessageContent interface{}           `json:"input_message_content,omitempty"`
}

// InlineQueryResultVideo is an inline query response video.
type InlineQueryResultVideo struct {
	Type                string                `json:"type"`      // required
	ID                  string                `json:"id"`        // required
	URL                 string                `json:"video_url"` // required
	MimeType            string                `json:"mime_type"` // required
	ThumbURL            string                `json:"thumb_url"`
	Title               string                `json:"title"`
	Caption             string                `json:"caption"`
	Width               int                   `json:"video_width"`
	Height              int                   `json:"video_height"`
	Duration            int                   `json:"video_duration"`
	Description         string                `json:"description"`
	ReplyMarkup         *InlineKeyboardMarkup `json:"reply_markup,omitempty"`
	InputMessageContent interface{}           `json:"input_message_content,omitempty"`
}

// InlineQueryResultAudio is an inline query response audio.
type InlineQueryResultAudio struct {
	Type                string                `json:"type"`      // required
	ID                  string                `json:"id"`        // required
	URL                 string                `json:"audio_url"` // required
	Title               string                `json:"title"`     // required
	Performer           string                `json:"performer"`
	Duration            int                   `json:"audio_duration"`
	ReplyMarkup         *InlineKeyboardMarkup `json:"reply_markup,omitempty"`
	InputMessageContent interface{}           `json:"input_message_content,omitempty"`
}

// InlineQueryResultVoice is an inline query response voice.
type InlineQueryResultVoice struct {
	Type                string                `json:"type"`      // required
	ID                  string                `json:"id"`        // required
	URL                 string                `json:"voice_url"` // required
	Title               string                `json:"title"`     // required
	Duration            int                   `json:"voice_duration"`
	ReplyMarkup         *InlineKeyboardMarkup `json:"reply_markup,omitempty"`
	InputMessageContent interface{}           `json:"input_message_content,omitempty"`
}

// InlineQueryResultDocument is an inline query response document.
type InlineQueryResultDocument struct {
	Type                string                `json:"type"`  // required
	ID                  string                `json:"id"`    // required
	Title               string                `json:"title"` // required
	Caption             string                `json:"caption"`
	URL                 string                `json:"document_url"` // required
	MimeType            string                `json:"mime_type"`    // required
	Description         string                `json:"description"`
	ReplyMarkup         *InlineKeyboardMarkup `json:"reply_markup,omitempty"`
	InputMessageContent interface{}           `json:"input_message_content,omitempty"`
	ThumbURL            string                `json:"thumb_url"`
	ThumbWidth          int                   `json:"thumb_width"`
	ThumbHeight         int                   `json:"thumb_height"`
}

// InlineQueryResultLocation is an inline query response location.
type InlineQueryResultLocation struct {
	Type                string                `json:"type"`      // required
	ID                  string                `json:"id"`        // required
	Latitude            float64               `json:"latitude"`  // required
	Longitude           float64               `json:"longitude"` // required
	Title               string                `json:"title"`     // required
	ReplyMarkup         *InlineKeyboardMarkup `json:"reply_markup,omitempty"`
	InputMessageContent interface{}           `json:"input_message_content,omitempty"`
	ThumbURL            string                `json:"thumb_url"`
	ThumbWidth          int                   `json:"thumb_width"`
	ThumbHeight         int                   `json:"thumb_height"`
}

// ChosenInlineResult is an inline query result chosen by a User
type ChosenInlineResult struct {
	ResultID        string    `json:"result_id"`
	From            *User     `json:"from"`
	Location        *Location `json:"location"`
	InlineMessageID string    `json:"inline_message_id"`
	Query           string    `json:"query"`
}

// InputTextMessageContent contains text for displaying
// as an inline query result.
type InputTextMessageContent struct {
	Text                  string `json:"message_text"`
	ParseMode             string `json:"parse_mode"`
	DisableWebPagePreview bool   `json:"disable_web_page_preview"`
}

// InputLocationMessageContent contains a location for displaying
// as an inline query result.
type InputLocationMessageContent struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

// InputVenueMessageContent contains a venue for displaying
// as an inline query result.
type InputVenueMessageContent struct {
	Latitude     float64 `json:"latitude"`
	Longitude    float64 `json:"longitude"`
	Title        string  `json:"title"`
	Address      string  `json:"address"`
	FoursquareID string  `json:"foursquare_id"`
}

// InputContactMessageContent contains a contact for displaying
// as an inline query result.
type InputContactMessageContent struct {
	PhoneNumber string `json:"phone_number"`
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
}
