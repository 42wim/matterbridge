package tgbotapi

import (
	"encoding/json"
	"io"
	"net/url"
	"strconv"
)

// Telegram constants
const (
	// APIEndpoint is the endpoint for all API methods,
	// with formatting for Sprintf.
	APIEndpoint = "https://api.telegram.org/bot%s/%s"
	// FileEndpoint is the endpoint for downloading a file from Telegram.
	FileEndpoint = "https://api.telegram.org/file/bot%s/%s"
)

// Constant values for ChatActions
const (
	ChatTyping         = "typing"
	ChatUploadPhoto    = "upload_photo"
	ChatRecordVideo    = "record_video"
	ChatUploadVideo    = "upload_video"
	ChatRecordAudio    = "record_audio"
	ChatUploadAudio    = "upload_audio"
	ChatUploadDocument = "upload_document"
	ChatFindLocation   = "find_location"
)

// API errors
const (
	// ErrAPIForbidden happens when a token is bad
	ErrAPIForbidden = "forbidden"
)

// Constant values for ParseMode in MessageConfig
const (
	ModeMarkdown = "Markdown"
	ModeHTML     = "HTML"
)

// Library errors
const (
	// ErrBadFileType happens when you pass an unknown type
	ErrBadFileType = "bad file type"
	ErrBadURL      = "bad or empty url"
)

// Chattable is any config type that can be sent.
type Chattable interface {
	values() (url.Values, error)
	method() string
}

// Fileable is any config type that can be sent that includes a file.
type Fileable interface {
	Chattable
	params() (map[string]string, error)
	name() string
	getFile() interface{}
	useExistingFile() bool
}

// BaseChat is base type for all chat config types.
type BaseChat struct {
	ChatID              int64 // required
	ChannelUsername     string
	ReplyToMessageID    int
	ReplyMarkup         interface{}
	DisableNotification bool
}

// values returns url.Values representation of BaseChat
func (chat *BaseChat) values() (url.Values, error) {
	v := url.Values{}
	if chat.ChannelUsername != "" {
		v.Add("chat_id", chat.ChannelUsername)
	} else {
		v.Add("chat_id", strconv.FormatInt(chat.ChatID, 10))
	}

	if chat.ReplyToMessageID != 0 {
		v.Add("reply_to_message_id", strconv.Itoa(chat.ReplyToMessageID))
	}

	if chat.ReplyMarkup != nil {
		data, err := json.Marshal(chat.ReplyMarkup)
		if err != nil {
			return v, err
		}

		v.Add("reply_markup", string(data))
	}

	v.Add("disable_notification", strconv.FormatBool(chat.DisableNotification))

	return v, nil
}

// BaseFile is a base type for all file config types.
type BaseFile struct {
	BaseChat
	File        interface{}
	FileID      string
	UseExisting bool
	MimeType    string
	FileSize    int
}

// params returns a map[string]string representation of BaseFile.
func (file BaseFile) params() (map[string]string, error) {
	params := make(map[string]string)

	if file.ChannelUsername != "" {
		params["chat_id"] = file.ChannelUsername
	} else {
		params["chat_id"] = strconv.FormatInt(file.ChatID, 10)
	}

	if file.ReplyToMessageID != 0 {
		params["reply_to_message_id"] = strconv.Itoa(file.ReplyToMessageID)
	}

	if file.ReplyMarkup != nil {
		data, err := json.Marshal(file.ReplyMarkup)
		if err != nil {
			return params, err
		}

		params["reply_markup"] = string(data)
	}

	if file.MimeType != "" {
		params["mime_type"] = file.MimeType
	}

	if file.FileSize > 0 {
		params["file_size"] = strconv.Itoa(file.FileSize)
	}

	params["disable_notification"] = strconv.FormatBool(file.DisableNotification)

	return params, nil
}

// getFile returns the file.
func (file BaseFile) getFile() interface{} {
	return file.File
}

// useExistingFile returns if the BaseFile has already been uploaded.
func (file BaseFile) useExistingFile() bool {
	return file.UseExisting
}

// BaseEdit is base type of all chat edits.
type BaseEdit struct {
	ChatID          int64
	ChannelUsername string
	MessageID       int
	InlineMessageID string
	ReplyMarkup     *InlineKeyboardMarkup
}

func (edit BaseEdit) values() (url.Values, error) {
	v := url.Values{}

	if edit.InlineMessageID == "" {
		if edit.ChannelUsername != "" {
			v.Add("chat_id", edit.ChannelUsername)
		} else {
			v.Add("chat_id", strconv.FormatInt(edit.ChatID, 10))
		}
		v.Add("message_id", strconv.Itoa(edit.MessageID))
	} else {
		v.Add("inline_message_id", edit.InlineMessageID)
	}

	if edit.ReplyMarkup != nil {
		data, err := json.Marshal(edit.ReplyMarkup)
		if err != nil {
			return v, err
		}
		v.Add("reply_markup", string(data))
	}

	return v, nil
}

// MessageConfig contains information about a SendMessage request.
type MessageConfig struct {
	BaseChat
	Text                  string
	ParseMode             string
	DisableWebPagePreview bool
}

// values returns a url.Values representation of MessageConfig.
func (config MessageConfig) values() (url.Values, error) {
	v, _ := config.BaseChat.values()
	v.Add("text", config.Text)
	v.Add("disable_web_page_preview", strconv.FormatBool(config.DisableWebPagePreview))
	if config.ParseMode != "" {
		v.Add("parse_mode", config.ParseMode)
	}

	return v, nil
}

// method returns Telegram API method name for sending Message.
func (config MessageConfig) method() string {
	return "sendMessage"
}

// ForwardConfig contains information about a ForwardMessage request.
type ForwardConfig struct {
	BaseChat
	FromChatID          int64 // required
	FromChannelUsername string
	MessageID           int // required
}

// values returns a url.Values representation of ForwardConfig.
func (config ForwardConfig) values() (url.Values, error) {
	v, _ := config.BaseChat.values()
	v.Add("from_chat_id", strconv.FormatInt(config.FromChatID, 10))
	v.Add("message_id", strconv.Itoa(config.MessageID))
	return v, nil
}

// method returns Telegram API method name for sending Forward.
func (config ForwardConfig) method() string {
	return "forwardMessage"
}

// PhotoConfig contains information about a SendPhoto request.
type PhotoConfig struct {
	BaseFile
	Caption string
}

// Params returns a map[string]string representation of PhotoConfig.
func (config PhotoConfig) params() (map[string]string, error) {
	params, _ := config.BaseFile.params()

	if config.Caption != "" {
		params["caption"] = config.Caption
	}

	return params, nil
}

// Values returns a url.Values representation of PhotoConfig.
func (config PhotoConfig) values() (url.Values, error) {
	v, _ := config.BaseChat.values()

	v.Add(config.name(), config.FileID)
	if config.Caption != "" {
		v.Add("caption", config.Caption)
	}
	return v, nil
}

// name returns the field name for the Photo.
func (config PhotoConfig) name() string {
	return "photo"
}

// method returns Telegram API method name for sending Photo.
func (config PhotoConfig) method() string {
	return "sendPhoto"
}

// AudioConfig contains information about a SendAudio request.
type AudioConfig struct {
	BaseFile
	Duration  int
	Performer string
	Title     string
}

// values returns a url.Values representation of AudioConfig.
func (config AudioConfig) values() (url.Values, error) {
	v, _ := config.BaseChat.values()

	v.Add(config.name(), config.FileID)
	if config.Duration != 0 {
		v.Add("duration", strconv.Itoa(config.Duration))
	}

	if config.Performer != "" {
		v.Add("performer", config.Performer)
	}
	if config.Title != "" {
		v.Add("title", config.Title)
	}

	return v, nil
}

// params returns a map[string]string representation of AudioConfig.
func (config AudioConfig) params() (map[string]string, error) {
	params, _ := config.BaseFile.params()

	if config.Duration != 0 {
		params["duration"] = strconv.Itoa(config.Duration)
	}

	if config.Performer != "" {
		params["performer"] = config.Performer
	}
	if config.Title != "" {
		params["title"] = config.Title
	}

	return params, nil
}

// name returns the field name for the Audio.
func (config AudioConfig) name() string {
	return "audio"
}

// method returns Telegram API method name for sending Audio.
func (config AudioConfig) method() string {
	return "sendAudio"
}

// DocumentConfig contains information about a SendDocument request.
type DocumentConfig struct {
	BaseFile
}

// values returns a url.Values representation of DocumentConfig.
func (config DocumentConfig) values() (url.Values, error) {
	v, _ := config.BaseChat.values()

	v.Add(config.name(), config.FileID)

	return v, nil
}

// params returns a map[string]string representation of DocumentConfig.
func (config DocumentConfig) params() (map[string]string, error) {
	params, _ := config.BaseFile.params()

	return params, nil
}

// name returns the field name for the Document.
func (config DocumentConfig) name() string {
	return "document"
}

// method returns Telegram API method name for sending Document.
func (config DocumentConfig) method() string {
	return "sendDocument"
}

// StickerConfig contains information about a SendSticker request.
type StickerConfig struct {
	BaseFile
}

// values returns a url.Values representation of StickerConfig.
func (config StickerConfig) values() (url.Values, error) {
	v, _ := config.BaseChat.values()

	v.Add(config.name(), config.FileID)

	return v, nil
}

// params returns a map[string]string representation of StickerConfig.
func (config StickerConfig) params() (map[string]string, error) {
	params, _ := config.BaseFile.params()

	return params, nil
}

// name returns the field name for the Sticker.
func (config StickerConfig) name() string {
	return "sticker"
}

// method returns Telegram API method name for sending Sticker.
func (config StickerConfig) method() string {
	return "sendSticker"
}

// VideoConfig contains information about a SendVideo request.
type VideoConfig struct {
	BaseFile
	Duration int
	Caption  string
}

// values returns a url.Values representation of VideoConfig.
func (config VideoConfig) values() (url.Values, error) {
	v, _ := config.BaseChat.values()

	v.Add(config.name(), config.FileID)
	if config.Duration != 0 {
		v.Add("duration", strconv.Itoa(config.Duration))
	}
	if config.Caption != "" {
		v.Add("caption", config.Caption)
	}

	return v, nil
}

// params returns a map[string]string representation of VideoConfig.
func (config VideoConfig) params() (map[string]string, error) {
	params, _ := config.BaseFile.params()

	return params, nil
}

// name returns the field name for the Video.
func (config VideoConfig) name() string {
	return "video"
}

// method returns Telegram API method name for sending Video.
func (config VideoConfig) method() string {
	return "sendVideo"
}

// VoiceConfig contains information about a SendVoice request.
type VoiceConfig struct {
	BaseFile
	Duration int
}

// values returns a url.Values representation of VoiceConfig.
func (config VoiceConfig) values() (url.Values, error) {
	v, _ := config.BaseChat.values()

	v.Add(config.name(), config.FileID)
	if config.Duration != 0 {
		v.Add("duration", strconv.Itoa(config.Duration))
	}

	return v, nil
}

// params returns a map[string]string representation of VoiceConfig.
func (config VoiceConfig) params() (map[string]string, error) {
	params, _ := config.BaseFile.params()

	if config.Duration != 0 {
		params["duration"] = strconv.Itoa(config.Duration)
	}

	return params, nil
}

// name returns the field name for the Voice.
func (config VoiceConfig) name() string {
	return "voice"
}

// method returns Telegram API method name for sending Voice.
func (config VoiceConfig) method() string {
	return "sendVoice"
}

// LocationConfig contains information about a SendLocation request.
type LocationConfig struct {
	BaseChat
	Latitude  float64 // required
	Longitude float64 // required
}

// values returns a url.Values representation of LocationConfig.
func (config LocationConfig) values() (url.Values, error) {
	v, _ := config.BaseChat.values()

	v.Add("latitude", strconv.FormatFloat(config.Latitude, 'f', 6, 64))
	v.Add("longitude", strconv.FormatFloat(config.Longitude, 'f', 6, 64))

	return v, nil
}

// method returns Telegram API method name for sending Location.
func (config LocationConfig) method() string {
	return "sendLocation"
}

// VenueConfig contains information about a SendVenue request.
type VenueConfig struct {
	BaseChat
	Latitude     float64 // required
	Longitude    float64 // required
	Title        string  // required
	Address      string  // required
	FoursquareID string
}

func (config VenueConfig) values() (url.Values, error) {
	v, _ := config.BaseChat.values()

	v.Add("latitude", strconv.FormatFloat(config.Latitude, 'f', 6, 64))
	v.Add("longitude", strconv.FormatFloat(config.Longitude, 'f', 6, 64))
	v.Add("title", config.Title)
	v.Add("address", config.Address)
	if config.FoursquareID != "" {
		v.Add("foursquare_id", config.FoursquareID)
	}

	return v, nil
}

func (config VenueConfig) method() string {
	return "sendVenue"
}

// ContactConfig allows you to send a contact.
type ContactConfig struct {
	BaseChat
	PhoneNumber string
	FirstName   string
	LastName    string
}

func (config ContactConfig) values() (url.Values, error) {
	v, _ := config.BaseChat.values()

	v.Add("phone_number", config.PhoneNumber)
	v.Add("first_name", config.FirstName)
	v.Add("last_name", config.LastName)

	return v, nil
}

func (config ContactConfig) method() string {
	return "sendContact"
}

// ChatActionConfig contains information about a SendChatAction request.
type ChatActionConfig struct {
	BaseChat
	Action string // required
}

// values returns a url.Values representation of ChatActionConfig.
func (config ChatActionConfig) values() (url.Values, error) {
	v, _ := config.BaseChat.values()
	v.Add("action", config.Action)
	return v, nil
}

// method returns Telegram API method name for sending ChatAction.
func (config ChatActionConfig) method() string {
	return "sendChatAction"
}

// EditMessageTextConfig allows you to modify the text in a message.
type EditMessageTextConfig struct {
	BaseEdit
	Text                  string
	ParseMode             string
	DisableWebPagePreview bool
}

func (config EditMessageTextConfig) values() (url.Values, error) {
	v, _ := config.BaseEdit.values()

	v.Add("text", config.Text)
	v.Add("parse_mode", config.ParseMode)
	v.Add("disable_web_page_preview", strconv.FormatBool(config.DisableWebPagePreview))

	return v, nil
}

func (config EditMessageTextConfig) method() string {
	return "editMessageText"
}

// EditMessageCaptionConfig allows you to modify the caption of a message.
type EditMessageCaptionConfig struct {
	BaseEdit
	Caption string
}

func (config EditMessageCaptionConfig) values() (url.Values, error) {
	v, _ := config.BaseEdit.values()

	v.Add("caption", config.Caption)

	return v, nil
}

func (config EditMessageCaptionConfig) method() string {
	return "editMessageCaption"
}

// EditMessageReplyMarkupConfig allows you to modify the reply markup
// of a message.
type EditMessageReplyMarkupConfig struct {
	BaseEdit
}

func (config EditMessageReplyMarkupConfig) values() (url.Values, error) {
	return config.BaseEdit.values()
}

func (config EditMessageReplyMarkupConfig) method() string {
	return "editMessageReplyMarkup"
}

// UserProfilePhotosConfig contains information about a
// GetUserProfilePhotos request.
type UserProfilePhotosConfig struct {
	UserID int
	Offset int
	Limit  int
}

// FileConfig has information about a file hosted on Telegram.
type FileConfig struct {
	FileID string
}

// UpdateConfig contains information about a GetUpdates request.
type UpdateConfig struct {
	Offset  int
	Limit   int
	Timeout int
}

// WebhookConfig contains information about a SetWebhook request.
type WebhookConfig struct {
	URL         *url.URL
	Certificate interface{}
}

// FileBytes contains information about a set of bytes to upload
// as a File.
type FileBytes struct {
	Name  string
	Bytes []byte
}

// FileReader contains information about a reader to upload as a File.
// If Size is -1, it will read the entire Reader into memory to
// calculate a Size.
type FileReader struct {
	Name   string
	Reader io.Reader
	Size   int64
}

// InlineConfig contains information on making an InlineQuery response.
type InlineConfig struct {
	InlineQueryID     string        `json:"inline_query_id"`
	Results           []interface{} `json:"results"`
	CacheTime         int           `json:"cache_time"`
	IsPersonal        bool          `json:"is_personal"`
	NextOffset        string        `json:"next_offset"`
	SwitchPMText      string        `json:"switch_pm_text"`
	SwitchPMParameter string        `json:"switch_pm_parameter"`
}

// CallbackConfig contains information on making a CallbackQuery response.
type CallbackConfig struct {
	CallbackQueryID string `json:"callback_query_id"`
	Text            string `json:"text"`
	ShowAlert       bool   `json:"show_alert"`
}

// ChatMemberConfig contains information about a user in a chat for use
// with administrative functions such as kicking or unbanning a user.
type ChatMemberConfig struct {
	ChatID             int64
	SuperGroupUsername string
	UserID             int
}

// ChatConfig contains information about getting information on a chat.
type ChatConfig struct {
	ChatID             int64
	SuperGroupUsername string
}

// ChatConfigWithUser contains information about getting information on
// a specific user within a chat.
type ChatConfigWithUser struct {
	ChatID             int64
	SuperGroupUsername string
	UserID             int
}
