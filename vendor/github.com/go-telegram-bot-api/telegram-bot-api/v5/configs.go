package tgbotapi

import (
	"bytes"
	"fmt"
	"io"
	"net/url"
	"os"
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
	ChatTyping          = "typing"
	ChatUploadPhoto     = "upload_photo"
	ChatRecordVideo     = "record_video"
	ChatUploadVideo     = "upload_video"
	ChatRecordVoice     = "record_voice"
	ChatUploadVoice     = "upload_voice"
	ChatUploadDocument  = "upload_document"
	ChatChooseSticker   = "choose_sticker"
	ChatFindLocation    = "find_location"
	ChatRecordVideoNote = "record_video_note"
	ChatUploadVideoNote = "upload_video_note"
)

// API errors
const (
	// ErrAPIForbidden happens when a token is bad
	ErrAPIForbidden = "forbidden"
)

// Constant values for ParseMode in MessageConfig
const (
	ModeMarkdown   = "Markdown"
	ModeMarkdownV2 = "MarkdownV2"
	ModeHTML       = "HTML"
)

// Constant values for update types
const (
	// UpdateTypeMessage is new incoming message of any kind — text, photo, sticker, etc.
	UpdateTypeMessage = "message"

	// UpdateTypeEditedMessage is new version of a message that is known to the bot and was edited
	UpdateTypeEditedMessage = "edited_message"

	// UpdateTypeChannelPost is new incoming channel post of any kind — text, photo, sticker, etc.
	UpdateTypeChannelPost = "channel_post"

	// UpdateTypeEditedChannelPost is new version of a channel post that is known to the bot and was edited
	UpdateTypeEditedChannelPost = "edited_channel_post"

	// UpdateTypeInlineQuery is new incoming inline query
	UpdateTypeInlineQuery = "inline_query"

	// UpdateTypeChosenInlineResult i the result of an inline query that was chosen by a user and sent to their
	// chat partner. Please see the documentation on the feedback collecting for
	// details on how to enable these updates for your bot.
	UpdateTypeChosenInlineResult = "chosen_inline_result"

	// UpdateTypeCallbackQuery is new incoming callback query
	UpdateTypeCallbackQuery = "callback_query"

	// UpdateTypeShippingQuery is new incoming shipping query. Only for invoices with flexible price
	UpdateTypeShippingQuery = "shipping_query"

	// UpdateTypePreCheckoutQuery is new incoming pre-checkout query. Contains full information about checkout
	UpdateTypePreCheckoutQuery = "pre_checkout_query"

	// UpdateTypePoll is new poll state. Bots receive only updates about stopped polls and polls
	// which are sent by the bot
	UpdateTypePoll = "poll"

	// UpdateTypePollAnswer is when user changed their answer in a non-anonymous poll. Bots receive new votes
	// only in polls that were sent by the bot itself.
	UpdateTypePollAnswer = "poll_answer"

	// UpdateTypeMyChatMember is when the bot's chat member status was updated in a chat. For private chats, this
	// update is received only when the bot is blocked or unblocked by the user.
	UpdateTypeMyChatMember = "my_chat_member"

	// UpdateTypeChatMember is when the bot must be an administrator in the chat and must explicitly specify
	// this update in the list of allowed_updates to receive these updates.
	UpdateTypeChatMember = "chat_member"
)

// Library errors
const (
	ErrBadURL = "bad or empty url"
)

// Chattable is any config type that can be sent.
type Chattable interface {
	params() (Params, error)
	method() string
}

// Fileable is any config type that can be sent that includes a file.
type Fileable interface {
	Chattable
	files() []RequestFile
}

// RequestFile represents a file associated with a field name.
type RequestFile struct {
	// The file field name.
	Name string
	// The file data to include.
	Data RequestFileData
}

// RequestFileData represents the data to be used for a file.
type RequestFileData interface {
	// NeedsUpload shows if the file needs to be uploaded.
	NeedsUpload() bool

	// UploadData gets the file name and an `io.Reader` for the file to be uploaded. This
	// must only be called when the file needs to be uploaded.
	UploadData() (string, io.Reader, error)
	// SendData gets the file data to send when a file does not need to be uploaded. This
	// must only be called when the file does not need to be uploaded.
	SendData() string
}

// FileBytes contains information about a set of bytes to upload
// as a File.
type FileBytes struct {
	Name  string
	Bytes []byte
}

func (fb FileBytes) NeedsUpload() bool {
	return true
}

func (fb FileBytes) UploadData() (string, io.Reader, error) {
	return fb.Name, bytes.NewReader(fb.Bytes), nil
}

func (fb FileBytes) SendData() string {
	panic("FileBytes must be uploaded")
}

// FileReader contains information about a reader to upload as a File.
type FileReader struct {
	Name   string
	Reader io.Reader
}

func (fr FileReader) NeedsUpload() bool {
	return true
}

func (fr FileReader) UploadData() (string, io.Reader, error) {
	return fr.Name, fr.Reader, nil
}

func (fr FileReader) SendData() string {
	panic("FileReader must be uploaded")
}

// FilePath is a path to a local file.
type FilePath string

func (fp FilePath) NeedsUpload() bool {
	return true
}

func (fp FilePath) UploadData() (string, io.Reader, error) {
	fileHandle, err := os.Open(string(fp))
	if err != nil {
		return "", nil, err
	}

	name := fileHandle.Name()
	return name, fileHandle, err
}

func (fp FilePath) SendData() string {
	panic("FilePath must be uploaded")
}

// FileURL is a URL to use as a file for a request.
type FileURL string

func (fu FileURL) NeedsUpload() bool {
	return false
}

func (fu FileURL) UploadData() (string, io.Reader, error) {
	panic("FileURL cannot be uploaded")
}

func (fu FileURL) SendData() string {
	return string(fu)
}

// FileID is an ID of a file already uploaded to Telegram.
type FileID string

func (fi FileID) NeedsUpload() bool {
	return false
}

func (fi FileID) UploadData() (string, io.Reader, error) {
	panic("FileID cannot be uploaded")
}

func (fi FileID) SendData() string {
	return string(fi)
}

// fileAttach is an internal file type used for processed media groups.
type fileAttach string

func (fa fileAttach) NeedsUpload() bool {
	return false
}

func (fa fileAttach) UploadData() (string, io.Reader, error) {
	panic("fileAttach cannot be uploaded")
}

func (fa fileAttach) SendData() string {
	return string(fa)
}

// LogOutConfig is a request to log out of the cloud Bot API server.
//
// Note that you may not log back in for at least 10 minutes.
type LogOutConfig struct{}

func (LogOutConfig) method() string {
	return "logOut"
}

func (LogOutConfig) params() (Params, error) {
	return nil, nil
}

// CloseConfig is a request to close the bot instance on a local server.
//
// Note that you may not close an instance for the first 10 minutes after the
// bot has started.
type CloseConfig struct{}

func (CloseConfig) method() string {
	return "close"
}

func (CloseConfig) params() (Params, error) {
	return nil, nil
}

// BaseChat is base type for all chat config types.
type BaseChat struct {
	ChatID                   int64 // required
	ChannelUsername          string
	ReplyToMessageID         int
	ReplyMarkup              interface{}
	DisableNotification      bool
	AllowSendingWithoutReply bool
}

func (chat *BaseChat) params() (Params, error) {
	params := make(Params)

	params.AddFirstValid("chat_id", chat.ChatID, chat.ChannelUsername)
	params.AddNonZero("reply_to_message_id", chat.ReplyToMessageID)
	params.AddBool("disable_notification", chat.DisableNotification)
	params.AddBool("allow_sending_without_reply", chat.AllowSendingWithoutReply)

	err := params.AddInterface("reply_markup", chat.ReplyMarkup)

	return params, err
}

// BaseFile is a base type for all file config types.
type BaseFile struct {
	BaseChat
	File RequestFileData
}

func (file BaseFile) params() (Params, error) {
	return file.BaseChat.params()
}

// BaseEdit is base type of all chat edits.
type BaseEdit struct {
	ChatID          int64
	ChannelUsername string
	MessageID       int
	InlineMessageID string
	ReplyMarkup     *InlineKeyboardMarkup
}

func (edit BaseEdit) params() (Params, error) {
	params := make(Params)

	if edit.InlineMessageID != "" {
		params["inline_message_id"] = edit.InlineMessageID
	} else {
		params.AddFirstValid("chat_id", edit.ChatID, edit.ChannelUsername)
		params.AddNonZero("message_id", edit.MessageID)
	}

	err := params.AddInterface("reply_markup", edit.ReplyMarkup)

	return params, err
}

// MessageConfig contains information about a SendMessage request.
type MessageConfig struct {
	BaseChat
	Text                  string
	ParseMode             string
	Entities              []MessageEntity
	DisableWebPagePreview bool
}

func (config MessageConfig) params() (Params, error) {
	params, err := config.BaseChat.params()
	if err != nil {
		return params, err
	}

	params.AddNonEmpty("text", config.Text)
	params.AddBool("disable_web_page_preview", config.DisableWebPagePreview)
	params.AddNonEmpty("parse_mode", config.ParseMode)
	err = params.AddInterface("entities", config.Entities)

	return params, err
}

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

func (config ForwardConfig) params() (Params, error) {
	params, err := config.BaseChat.params()
	if err != nil {
		return params, err
	}

	params.AddNonZero64("from_chat_id", config.FromChatID)
	params.AddNonZero("message_id", config.MessageID)

	return params, nil
}

func (config ForwardConfig) method() string {
	return "forwardMessage"
}

// CopyMessageConfig contains information about a copyMessage request.
type CopyMessageConfig struct {
	BaseChat
	FromChatID          int64
	FromChannelUsername string
	MessageID           int
	Caption             string
	ParseMode           string
	CaptionEntities     []MessageEntity
}

func (config CopyMessageConfig) params() (Params, error) {
	params, err := config.BaseChat.params()
	if err != nil {
		return params, err
	}

	params.AddFirstValid("from_chat_id", config.FromChatID, config.FromChannelUsername)
	params.AddNonZero("message_id", config.MessageID)
	params.AddNonEmpty("caption", config.Caption)
	params.AddNonEmpty("parse_mode", config.ParseMode)
	err = params.AddInterface("caption_entities", config.CaptionEntities)

	return params, err
}

func (config CopyMessageConfig) method() string {
	return "copyMessage"
}

// PhotoConfig contains information about a SendPhoto request.
type PhotoConfig struct {
	BaseFile
	Thumb           RequestFileData
	Caption         string
	ParseMode       string
	CaptionEntities []MessageEntity
}

func (config PhotoConfig) params() (Params, error) {
	params, err := config.BaseFile.params()
	if err != nil {
		return params, err
	}

	params.AddNonEmpty("caption", config.Caption)
	params.AddNonEmpty("parse_mode", config.ParseMode)
	err = params.AddInterface("caption_entities", config.CaptionEntities)

	return params, err
}

func (config PhotoConfig) method() string {
	return "sendPhoto"
}

func (config PhotoConfig) files() []RequestFile {
	files := []RequestFile{{
		Name: "photo",
		Data: config.File,
	}}

	if config.Thumb != nil {
		files = append(files, RequestFile{
			Name: "thumb",
			Data: config.Thumb,
		})
	}

	return files
}

// AudioConfig contains information about a SendAudio request.
type AudioConfig struct {
	BaseFile
	Thumb           RequestFileData
	Caption         string
	ParseMode       string
	CaptionEntities []MessageEntity
	Duration        int
	Performer       string
	Title           string
}

func (config AudioConfig) params() (Params, error) {
	params, err := config.BaseChat.params()
	if err != nil {
		return params, err
	}

	params.AddNonZero("duration", config.Duration)
	params.AddNonEmpty("performer", config.Performer)
	params.AddNonEmpty("title", config.Title)
	params.AddNonEmpty("caption", config.Caption)
	params.AddNonEmpty("parse_mode", config.ParseMode)
	err = params.AddInterface("caption_entities", config.CaptionEntities)

	return params, err
}

func (config AudioConfig) method() string {
	return "sendAudio"
}

func (config AudioConfig) files() []RequestFile {
	files := []RequestFile{{
		Name: "audio",
		Data: config.File,
	}}

	if config.Thumb != nil {
		files = append(files, RequestFile{
			Name: "thumb",
			Data: config.Thumb,
		})
	}

	return files
}

// DocumentConfig contains information about a SendDocument request.
type DocumentConfig struct {
	BaseFile
	Thumb                       RequestFileData
	Caption                     string
	ParseMode                   string
	CaptionEntities             []MessageEntity
	DisableContentTypeDetection bool
}

func (config DocumentConfig) params() (Params, error) {
	params, err := config.BaseFile.params()

	params.AddNonEmpty("caption", config.Caption)
	params.AddNonEmpty("parse_mode", config.ParseMode)
	params.AddBool("disable_content_type_detection", config.DisableContentTypeDetection)

	return params, err
}

func (config DocumentConfig) method() string {
	return "sendDocument"
}

func (config DocumentConfig) files() []RequestFile {
	files := []RequestFile{{
		Name: "document",
		Data: config.File,
	}}

	if config.Thumb != nil {
		files = append(files, RequestFile{
			Name: "thumb",
			Data: config.Thumb,
		})
	}

	return files
}

// StickerConfig contains information about a SendSticker request.
type StickerConfig struct {
	BaseFile
}

func (config StickerConfig) params() (Params, error) {
	return config.BaseChat.params()
}

func (config StickerConfig) method() string {
	return "sendSticker"
}

func (config StickerConfig) files() []RequestFile {
	return []RequestFile{{
		Name: "sticker",
		Data: config.File,
	}}
}

// VideoConfig contains information about a SendVideo request.
type VideoConfig struct {
	BaseFile
	Thumb             RequestFileData
	Duration          int
	Caption           string
	ParseMode         string
	CaptionEntities   []MessageEntity
	SupportsStreaming bool
}

func (config VideoConfig) params() (Params, error) {
	params, err := config.BaseChat.params()
	if err != nil {
		return params, err
	}

	params.AddNonZero("duration", config.Duration)
	params.AddNonEmpty("caption", config.Caption)
	params.AddNonEmpty("parse_mode", config.ParseMode)
	params.AddBool("supports_streaming", config.SupportsStreaming)
	err = params.AddInterface("caption_entities", config.CaptionEntities)

	return params, err
}

func (config VideoConfig) method() string {
	return "sendVideo"
}

func (config VideoConfig) files() []RequestFile {
	files := []RequestFile{{
		Name: "video",
		Data: config.File,
	}}

	if config.Thumb != nil {
		files = append(files, RequestFile{
			Name: "thumb",
			Data: config.Thumb,
		})
	}

	return files
}

// AnimationConfig contains information about a SendAnimation request.
type AnimationConfig struct {
	BaseFile
	Duration        int
	Thumb           RequestFileData
	Caption         string
	ParseMode       string
	CaptionEntities []MessageEntity
}

func (config AnimationConfig) params() (Params, error) {
	params, err := config.BaseChat.params()
	if err != nil {
		return params, err
	}

	params.AddNonZero("duration", config.Duration)
	params.AddNonEmpty("caption", config.Caption)
	params.AddNonEmpty("parse_mode", config.ParseMode)
	err = params.AddInterface("caption_entities", config.CaptionEntities)

	return params, err
}

func (config AnimationConfig) method() string {
	return "sendAnimation"
}

func (config AnimationConfig) files() []RequestFile {
	files := []RequestFile{{
		Name: "animation",
		Data: config.File,
	}}

	if config.Thumb != nil {
		files = append(files, RequestFile{
			Name: "thumb",
			Data: config.Thumb,
		})
	}

	return files
}

// VideoNoteConfig contains information about a SendVideoNote request.
type VideoNoteConfig struct {
	BaseFile
	Thumb    RequestFileData
	Duration int
	Length   int
}

func (config VideoNoteConfig) params() (Params, error) {
	params, err := config.BaseChat.params()

	params.AddNonZero("duration", config.Duration)
	params.AddNonZero("length", config.Length)

	return params, err
}

func (config VideoNoteConfig) method() string {
	return "sendVideoNote"
}

func (config VideoNoteConfig) files() []RequestFile {
	files := []RequestFile{{
		Name: "video_note",
		Data: config.File,
	}}

	if config.Thumb != nil {
		files = append(files, RequestFile{
			Name: "thumb",
			Data: config.Thumb,
		})
	}

	return files
}

// VoiceConfig contains information about a SendVoice request.
type VoiceConfig struct {
	BaseFile
	Thumb           RequestFileData
	Caption         string
	ParseMode       string
	CaptionEntities []MessageEntity
	Duration        int
}

func (config VoiceConfig) params() (Params, error) {
	params, err := config.BaseChat.params()
	if err != nil {
		return params, err
	}

	params.AddNonZero("duration", config.Duration)
	params.AddNonEmpty("caption", config.Caption)
	params.AddNonEmpty("parse_mode", config.ParseMode)
	err = params.AddInterface("caption_entities", config.CaptionEntities)

	return params, err
}

func (config VoiceConfig) method() string {
	return "sendVoice"
}

func (config VoiceConfig) files() []RequestFile {
	files := []RequestFile{{
		Name: "voice",
		Data: config.File,
	}}

	if config.Thumb != nil {
		files = append(files, RequestFile{
			Name: "thumb",
			Data: config.Thumb,
		})
	}

	return files
}

// LocationConfig contains information about a SendLocation request.
type LocationConfig struct {
	BaseChat
	Latitude             float64 // required
	Longitude            float64 // required
	HorizontalAccuracy   float64 // optional
	LivePeriod           int     // optional
	Heading              int     // optional
	ProximityAlertRadius int     // optional
}

func (config LocationConfig) params() (Params, error) {
	params, err := config.BaseChat.params()

	params.AddNonZeroFloat("latitude", config.Latitude)
	params.AddNonZeroFloat("longitude", config.Longitude)
	params.AddNonZeroFloat("horizontal_accuracy", config.HorizontalAccuracy)
	params.AddNonZero("live_period", config.LivePeriod)
	params.AddNonZero("heading", config.Heading)
	params.AddNonZero("proximity_alert_radius", config.ProximityAlertRadius)

	return params, err
}

func (config LocationConfig) method() string {
	return "sendLocation"
}

// EditMessageLiveLocationConfig allows you to update a live location.
type EditMessageLiveLocationConfig struct {
	BaseEdit
	Latitude             float64 // required
	Longitude            float64 // required
	HorizontalAccuracy   float64 // optional
	Heading              int     // optional
	ProximityAlertRadius int     // optional
}

func (config EditMessageLiveLocationConfig) params() (Params, error) {
	params, err := config.BaseEdit.params()

	params.AddNonZeroFloat("latitude", config.Latitude)
	params.AddNonZeroFloat("longitude", config.Longitude)
	params.AddNonZeroFloat("horizontal_accuracy", config.HorizontalAccuracy)
	params.AddNonZero("heading", config.Heading)
	params.AddNonZero("proximity_alert_radius", config.ProximityAlertRadius)

	return params, err
}

func (config EditMessageLiveLocationConfig) method() string {
	return "editMessageLiveLocation"
}

// StopMessageLiveLocationConfig stops updating a live location.
type StopMessageLiveLocationConfig struct {
	BaseEdit
}

func (config StopMessageLiveLocationConfig) params() (Params, error) {
	return config.BaseEdit.params()
}

func (config StopMessageLiveLocationConfig) method() string {
	return "stopMessageLiveLocation"
}

// VenueConfig contains information about a SendVenue request.
type VenueConfig struct {
	BaseChat
	Latitude        float64 // required
	Longitude       float64 // required
	Title           string  // required
	Address         string  // required
	FoursquareID    string
	FoursquareType  string
	GooglePlaceID   string
	GooglePlaceType string
}

func (config VenueConfig) params() (Params, error) {
	params, err := config.BaseChat.params()

	params.AddNonZeroFloat("latitude", config.Latitude)
	params.AddNonZeroFloat("longitude", config.Longitude)
	params["title"] = config.Title
	params["address"] = config.Address
	params.AddNonEmpty("foursquare_id", config.FoursquareID)
	params.AddNonEmpty("foursquare_type", config.FoursquareType)
	params.AddNonEmpty("google_place_id", config.GooglePlaceID)
	params.AddNonEmpty("google_place_type", config.GooglePlaceType)

	return params, err
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
	VCard       string
}

func (config ContactConfig) params() (Params, error) {
	params, err := config.BaseChat.params()

	params["phone_number"] = config.PhoneNumber
	params["first_name"] = config.FirstName

	params.AddNonEmpty("last_name", config.LastName)
	params.AddNonEmpty("vcard", config.VCard)

	return params, err
}

func (config ContactConfig) method() string {
	return "sendContact"
}

// SendPollConfig allows you to send a poll.
type SendPollConfig struct {
	BaseChat
	Question              string
	Options               []string
	IsAnonymous           bool
	Type                  string
	AllowsMultipleAnswers bool
	CorrectOptionID       int64
	Explanation           string
	ExplanationParseMode  string
	ExplanationEntities   []MessageEntity
	OpenPeriod            int
	CloseDate             int
	IsClosed              bool
}

func (config SendPollConfig) params() (Params, error) {
	params, err := config.BaseChat.params()
	if err != nil {
		return params, err
	}

	params["question"] = config.Question
	if err = params.AddInterface("options", config.Options); err != nil {
		return params, err
	}
	params["is_anonymous"] = strconv.FormatBool(config.IsAnonymous)
	params.AddNonEmpty("type", config.Type)
	params["allows_multiple_answers"] = strconv.FormatBool(config.AllowsMultipleAnswers)
	params["correct_option_id"] = strconv.FormatInt(config.CorrectOptionID, 10)
	params.AddBool("is_closed", config.IsClosed)
	params.AddNonEmpty("explanation", config.Explanation)
	params.AddNonEmpty("explanation_parse_mode", config.ExplanationParseMode)
	params.AddNonZero("open_period", config.OpenPeriod)
	params.AddNonZero("close_date", config.CloseDate)
	err = params.AddInterface("explanation_entities", config.ExplanationEntities)

	return params, err
}

func (SendPollConfig) method() string {
	return "sendPoll"
}

// GameConfig allows you to send a game.
type GameConfig struct {
	BaseChat
	GameShortName string
}

func (config GameConfig) params() (Params, error) {
	params, err := config.BaseChat.params()

	params["game_short_name"] = config.GameShortName

	return params, err
}

func (config GameConfig) method() string {
	return "sendGame"
}

// SetGameScoreConfig allows you to update the game score in a chat.
type SetGameScoreConfig struct {
	UserID             int64
	Score              int
	Force              bool
	DisableEditMessage bool
	ChatID             int64
	ChannelUsername    string
	MessageID          int
	InlineMessageID    string
}

func (config SetGameScoreConfig) params() (Params, error) {
	params := make(Params)

	params.AddNonZero64("user_id", config.UserID)
	params.AddNonZero("scrore", config.Score)
	params.AddBool("disable_edit_message", config.DisableEditMessage)

	if config.InlineMessageID != "" {
		params["inline_message_id"] = config.InlineMessageID
	} else {
		params.AddFirstValid("chat_id", config.ChatID, config.ChannelUsername)
		params.AddNonZero("message_id", config.MessageID)
	}

	return params, nil
}

func (config SetGameScoreConfig) method() string {
	return "setGameScore"
}

// GetGameHighScoresConfig allows you to fetch the high scores for a game.
type GetGameHighScoresConfig struct {
	UserID          int64
	ChatID          int64
	ChannelUsername string
	MessageID       int
	InlineMessageID string
}

func (config GetGameHighScoresConfig) params() (Params, error) {
	params := make(Params)

	params.AddNonZero64("user_id", config.UserID)

	if config.InlineMessageID != "" {
		params["inline_message_id"] = config.InlineMessageID
	} else {
		params.AddFirstValid("chat_id", config.ChatID, config.ChannelUsername)
		params.AddNonZero("message_id", config.MessageID)
	}

	return params, nil
}

func (config GetGameHighScoresConfig) method() string {
	return "getGameHighScores"
}

// ChatActionConfig contains information about a SendChatAction request.
type ChatActionConfig struct {
	BaseChat
	Action string // required
}

func (config ChatActionConfig) params() (Params, error) {
	params, err := config.BaseChat.params()

	params["action"] = config.Action

	return params, err
}

func (config ChatActionConfig) method() string {
	return "sendChatAction"
}

// EditMessageTextConfig allows you to modify the text in a message.
type EditMessageTextConfig struct {
	BaseEdit
	Text                  string
	ParseMode             string
	Entities              []MessageEntity
	DisableWebPagePreview bool
}

func (config EditMessageTextConfig) params() (Params, error) {
	params, err := config.BaseEdit.params()
	if err != nil {
		return params, err
	}

	params["text"] = config.Text
	params.AddNonEmpty("parse_mode", config.ParseMode)
	params.AddBool("disable_web_page_preview", config.DisableWebPagePreview)
	err = params.AddInterface("entities", config.Entities)

	return params, err
}

func (config EditMessageTextConfig) method() string {
	return "editMessageText"
}

// EditMessageCaptionConfig allows you to modify the caption of a message.
type EditMessageCaptionConfig struct {
	BaseEdit
	Caption         string
	ParseMode       string
	CaptionEntities []MessageEntity
}

func (config EditMessageCaptionConfig) params() (Params, error) {
	params, err := config.BaseEdit.params()
	if err != nil {
		return params, err
	}

	params["caption"] = config.Caption
	params.AddNonEmpty("parse_mode", config.ParseMode)
	err = params.AddInterface("caption_entities", config.CaptionEntities)

	return params, err
}

func (config EditMessageCaptionConfig) method() string {
	return "editMessageCaption"
}

// EditMessageMediaConfig allows you to make an editMessageMedia request.
type EditMessageMediaConfig struct {
	BaseEdit

	Media interface{}
}

func (EditMessageMediaConfig) method() string {
	return "editMessageMedia"
}

func (config EditMessageMediaConfig) params() (Params, error) {
	params, err := config.BaseEdit.params()
	if err != nil {
		return params, err
	}

	err = params.AddInterface("media", prepareInputMediaParam(config.Media, 0))

	return params, err
}

func (config EditMessageMediaConfig) files() []RequestFile {
	return prepareInputMediaFile(config.Media, 0)
}

// EditMessageReplyMarkupConfig allows you to modify the reply markup
// of a message.
type EditMessageReplyMarkupConfig struct {
	BaseEdit
}

func (config EditMessageReplyMarkupConfig) params() (Params, error) {
	return config.BaseEdit.params()
}

func (config EditMessageReplyMarkupConfig) method() string {
	return "editMessageReplyMarkup"
}

// StopPollConfig allows you to stop a poll sent by the bot.
type StopPollConfig struct {
	BaseEdit
}

func (config StopPollConfig) params() (Params, error) {
	return config.BaseEdit.params()
}

func (StopPollConfig) method() string {
	return "stopPoll"
}

// UserProfilePhotosConfig contains information about a
// GetUserProfilePhotos request.
type UserProfilePhotosConfig struct {
	UserID int64
	Offset int
	Limit  int
}

func (UserProfilePhotosConfig) method() string {
	return "getUserProfilePhotos"
}

func (config UserProfilePhotosConfig) params() (Params, error) {
	params := make(Params)

	params.AddNonZero64("user_id", config.UserID)
	params.AddNonZero("offset", config.Offset)
	params.AddNonZero("limit", config.Limit)

	return params, nil
}

// FileConfig has information about a file hosted on Telegram.
type FileConfig struct {
	FileID string
}

func (FileConfig) method() string {
	return "getFile"
}

func (config FileConfig) params() (Params, error) {
	params := make(Params)

	params["file_id"] = config.FileID

	return params, nil
}

// UpdateConfig contains information about a GetUpdates request.
type UpdateConfig struct {
	Offset         int
	Limit          int
	Timeout        int
	AllowedUpdates []string
}

func (UpdateConfig) method() string {
	return "getUpdates"
}

func (config UpdateConfig) params() (Params, error) {
	params := make(Params)

	params.AddNonZero("offset", config.Offset)
	params.AddNonZero("limit", config.Limit)
	params.AddNonZero("timeout", config.Timeout)
	params.AddInterface("allowed_updates", config.AllowedUpdates)

	return params, nil
}

// WebhookConfig contains information about a SetWebhook request.
type WebhookConfig struct {
	URL                *url.URL
	Certificate        RequestFileData
	IPAddress          string
	MaxConnections     int
	AllowedUpdates     []string
	DropPendingUpdates bool
}

func (config WebhookConfig) method() string {
	return "setWebhook"
}

func (config WebhookConfig) params() (Params, error) {
	params := make(Params)

	if config.URL != nil {
		params["url"] = config.URL.String()
	}

	params.AddNonEmpty("ip_address", config.IPAddress)
	params.AddNonZero("max_connections", config.MaxConnections)
	err := params.AddInterface("allowed_updates", config.AllowedUpdates)
	params.AddBool("drop_pending_updates", config.DropPendingUpdates)

	return params, err
}

func (config WebhookConfig) files() []RequestFile {
	if config.Certificate != nil {
		return []RequestFile{{
			Name: "certificate",
			Data: config.Certificate,
		}}
	}

	return nil
}

// DeleteWebhookConfig is a helper to delete a webhook.
type DeleteWebhookConfig struct {
	DropPendingUpdates bool
}

func (config DeleteWebhookConfig) method() string {
	return "deleteWebhook"
}

func (config DeleteWebhookConfig) params() (Params, error) {
	params := make(Params)

	params.AddBool("drop_pending_updates", config.DropPendingUpdates)

	return params, nil
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

func (config InlineConfig) method() string {
	return "answerInlineQuery"
}

func (config InlineConfig) params() (Params, error) {
	params := make(Params)

	params["inline_query_id"] = config.InlineQueryID
	params.AddNonZero("cache_time", config.CacheTime)
	params.AddBool("is_personal", config.IsPersonal)
	params.AddNonEmpty("next_offset", config.NextOffset)
	params.AddNonEmpty("switch_pm_text", config.SwitchPMText)
	params.AddNonEmpty("switch_pm_parameter", config.SwitchPMParameter)
	err := params.AddInterface("results", config.Results)

	return params, err
}

// CallbackConfig contains information on making a CallbackQuery response.
type CallbackConfig struct {
	CallbackQueryID string `json:"callback_query_id"`
	Text            string `json:"text"`
	ShowAlert       bool   `json:"show_alert"`
	URL             string `json:"url"`
	CacheTime       int    `json:"cache_time"`
}

func (config CallbackConfig) method() string {
	return "answerCallbackQuery"
}

func (config CallbackConfig) params() (Params, error) {
	params := make(Params)

	params["callback_query_id"] = config.CallbackQueryID
	params.AddNonEmpty("text", config.Text)
	params.AddBool("show_alert", config.ShowAlert)
	params.AddNonEmpty("url", config.URL)
	params.AddNonZero("cache_time", config.CacheTime)

	return params, nil
}

// ChatMemberConfig contains information about a user in a chat for use
// with administrative functions such as kicking or unbanning a user.
type ChatMemberConfig struct {
	ChatID             int64
	SuperGroupUsername string
	ChannelUsername    string
	UserID             int64
}

// UnbanChatMemberConfig allows you to unban a user.
type UnbanChatMemberConfig struct {
	ChatMemberConfig
	OnlyIfBanned bool
}

func (config UnbanChatMemberConfig) method() string {
	return "unbanChatMember"
}

func (config UnbanChatMemberConfig) params() (Params, error) {
	params := make(Params)

	params.AddFirstValid("chat_id", config.ChatID, config.SuperGroupUsername, config.ChannelUsername)
	params.AddNonZero64("user_id", config.UserID)
	params.AddBool("only_if_banned", config.OnlyIfBanned)

	return params, nil
}

// BanChatMemberConfig contains extra fields to kick user.
type BanChatMemberConfig struct {
	ChatMemberConfig
	UntilDate      int64
	RevokeMessages bool
}

func (config BanChatMemberConfig) method() string {
	return "banChatMember"
}

func (config BanChatMemberConfig) params() (Params, error) {
	params := make(Params)

	params.AddFirstValid("chat_id", config.ChatID, config.SuperGroupUsername)
	params.AddNonZero64("user_id", config.UserID)
	params.AddNonZero64("until_date", config.UntilDate)
	params.AddBool("revoke_messages", config.RevokeMessages)

	return params, nil
}

// KickChatMemberConfig contains extra fields to ban user.
//
// This was renamed to BanChatMember in later versions of the Telegram Bot API.
type KickChatMemberConfig = BanChatMemberConfig

// RestrictChatMemberConfig contains fields to restrict members of chat
type RestrictChatMemberConfig struct {
	ChatMemberConfig
	UntilDate   int64
	Permissions *ChatPermissions
}

func (config RestrictChatMemberConfig) method() string {
	return "restrictChatMember"
}

func (config RestrictChatMemberConfig) params() (Params, error) {
	params := make(Params)

	params.AddFirstValid("chat_id", config.ChatID, config.SuperGroupUsername, config.ChannelUsername)
	params.AddNonZero64("user_id", config.UserID)

	err := params.AddInterface("permissions", config.Permissions)
	params.AddNonZero64("until_date", config.UntilDate)

	return params, err
}

// PromoteChatMemberConfig contains fields to promote members of chat
type PromoteChatMemberConfig struct {
	ChatMemberConfig
	IsAnonymous         bool
	CanManageChat       bool
	CanChangeInfo       bool
	CanPostMessages     bool
	CanEditMessages     bool
	CanDeleteMessages   bool
	CanManageVoiceChats bool
	CanInviteUsers      bool
	CanRestrictMembers  bool
	CanPinMessages      bool
	CanPromoteMembers   bool
}

func (config PromoteChatMemberConfig) method() string {
	return "promoteChatMember"
}

func (config PromoteChatMemberConfig) params() (Params, error) {
	params := make(Params)

	params.AddFirstValid("chat_id", config.ChatID, config.SuperGroupUsername, config.ChannelUsername)
	params.AddNonZero64("user_id", config.UserID)

	params.AddBool("is_anonymous", config.IsAnonymous)
	params.AddBool("can_manage_chat", config.CanManageChat)
	params.AddBool("can_change_info", config.CanChangeInfo)
	params.AddBool("can_post_messages", config.CanPostMessages)
	params.AddBool("can_edit_messages", config.CanEditMessages)
	params.AddBool("can_delete_messages", config.CanDeleteMessages)
	params.AddBool("can_manage_voice_chats", config.CanManageVoiceChats)
	params.AddBool("can_invite_users", config.CanInviteUsers)
	params.AddBool("can_restrict_members", config.CanRestrictMembers)
	params.AddBool("can_pin_messages", config.CanPinMessages)
	params.AddBool("can_promote_members", config.CanPromoteMembers)

	return params, nil
}

// SetChatAdministratorCustomTitle sets the title of an administrative user
// promoted by the bot for a chat.
type SetChatAdministratorCustomTitle struct {
	ChatMemberConfig
	CustomTitle string
}

func (SetChatAdministratorCustomTitle) method() string {
	return "setChatAdministratorCustomTitle"
}

func (config SetChatAdministratorCustomTitle) params() (Params, error) {
	params := make(Params)

	params.AddFirstValid("chat_id", config.ChatID, config.SuperGroupUsername, config.ChannelUsername)
	params.AddNonZero64("user_id", config.UserID)
	params.AddNonEmpty("custom_title", config.CustomTitle)

	return params, nil
}

// BanChatSenderChatConfig bans a channel chat in a supergroup or a channel. The
// owner of the chat will not be able to send messages and join live streams on
// behalf of the chat, unless it is unbanned first. The bot must be an
// administrator in the supergroup or channel for this to work and must have the
// appropriate administrator rights.
type BanChatSenderChatConfig struct {
	ChatID          int64
	ChannelUsername string
	SenderChatID    int64
	UntilDate       int
}

func (config BanChatSenderChatConfig) method() string {
	return "banChatSenderChat"
}

func (config BanChatSenderChatConfig) params() (Params, error) {
	params := make(Params)

	_ = params.AddFirstValid("chat_id", config.ChatID, config.ChannelUsername)
	params.AddNonZero64("sender_chat_id", config.SenderChatID)
	params.AddNonZero("until_date", config.UntilDate)

	return params, nil
}

// UnbanChatSenderChatConfig unbans a previously banned channel chat in a
// supergroup or channel. The bot must be an administrator for this to work and
// must have the appropriate administrator rights.
type UnbanChatSenderChatConfig struct {
	ChatID          int64
	ChannelUsername string
	SenderChatID    int64
}

func (config UnbanChatSenderChatConfig) method() string {
	return "unbanChatSenderChat"
}

func (config UnbanChatSenderChatConfig) params() (Params, error) {
	params := make(Params)

	_ = params.AddFirstValid("chat_id", config.ChatID, config.ChannelUsername)
	params.AddNonZero64("sender_chat_id", config.SenderChatID)

	return params, nil
}

// ChatConfig contains information about getting information on a chat.
type ChatConfig struct {
	ChatID             int64
	SuperGroupUsername string
}

func (config ChatConfig) params() (Params, error) {
	params := make(Params)

	params.AddFirstValid("chat_id", config.ChatID, config.SuperGroupUsername)

	return params, nil
}

// ChatInfoConfig contains information about getting chat information.
type ChatInfoConfig struct {
	ChatConfig
}

func (ChatInfoConfig) method() string {
	return "getChat"
}

// ChatMemberCountConfig contains information about getting the number of users in a chat.
type ChatMemberCountConfig struct {
	ChatConfig
}

func (ChatMemberCountConfig) method() string {
	return "getChatMembersCount"
}

// ChatAdministratorsConfig contains information about getting chat administrators.
type ChatAdministratorsConfig struct {
	ChatConfig
}

func (ChatAdministratorsConfig) method() string {
	return "getChatAdministrators"
}

// SetChatPermissionsConfig allows you to set default permissions for the
// members in a group. The bot must be an administrator and have rights to
// restrict members.
type SetChatPermissionsConfig struct {
	ChatConfig
	Permissions *ChatPermissions
}

func (SetChatPermissionsConfig) method() string {
	return "setChatPermissions"
}

func (config SetChatPermissionsConfig) params() (Params, error) {
	params := make(Params)

	params.AddFirstValid("chat_id", config.ChatID, config.SuperGroupUsername)
	err := params.AddInterface("permissions", config.Permissions)

	return params, err
}

// ChatInviteLinkConfig contains information about getting a chat link.
//
// Note that generating a new link will revoke any previous links.
type ChatInviteLinkConfig struct {
	ChatConfig
}

func (ChatInviteLinkConfig) method() string {
	return "exportChatInviteLink"
}

func (config ChatInviteLinkConfig) params() (Params, error) {
	params := make(Params)

	params.AddFirstValid("chat_id", config.ChatID, config.SuperGroupUsername)

	return params, nil
}

// CreateChatInviteLinkConfig allows you to create an additional invite link for
// a chat. The bot must be an administrator in the chat for this to work and
// must have the appropriate admin rights. The link can be revoked using the
// RevokeChatInviteLinkConfig.
type CreateChatInviteLinkConfig struct {
	ChatConfig
	Name               string
	ExpireDate         int
	MemberLimit        int
	CreatesJoinRequest bool
}

func (CreateChatInviteLinkConfig) method() string {
	return "createChatInviteLink"
}

func (config CreateChatInviteLinkConfig) params() (Params, error) {
	params := make(Params)

	params.AddNonEmpty("name", config.Name)
	params.AddFirstValid("chat_id", config.ChatID, config.SuperGroupUsername)
	params.AddNonZero("expire_date", config.ExpireDate)
	params.AddNonZero("member_limit", config.MemberLimit)
	params.AddBool("creates_join_request", config.CreatesJoinRequest)

	return params, nil
}

// EditChatInviteLinkConfig allows you to edit a non-primary invite link created
// by the bot. The bot must be an administrator in the chat for this to work and
// must have the appropriate admin rights.
type EditChatInviteLinkConfig struct {
	ChatConfig
	InviteLink         string
	Name               string
	ExpireDate         int
	MemberLimit        int
	CreatesJoinRequest bool
}

func (EditChatInviteLinkConfig) method() string {
	return "editChatInviteLink"
}

func (config EditChatInviteLinkConfig) params() (Params, error) {
	params := make(Params)

	params.AddFirstValid("chat_id", config.ChatID, config.SuperGroupUsername)
	params.AddNonEmpty("name", config.Name)
	params["invite_link"] = config.InviteLink
	params.AddNonZero("expire_date", config.ExpireDate)
	params.AddNonZero("member_limit", config.MemberLimit)
	params.AddBool("creates_join_request", config.CreatesJoinRequest)

	return params, nil
}

// RevokeChatInviteLinkConfig allows you to revoke an invite link created by the
// bot. If the primary link is revoked, a new link is automatically generated.
// The bot must be an administrator in the chat for this to work and must have
// the appropriate admin rights.
type RevokeChatInviteLinkConfig struct {
	ChatConfig
	InviteLink string
}

func (RevokeChatInviteLinkConfig) method() string {
	return "revokeChatInviteLink"
}

func (config RevokeChatInviteLinkConfig) params() (Params, error) {
	params := make(Params)

	params.AddFirstValid("chat_id", config.ChatID, config.SuperGroupUsername)
	params["invite_link"] = config.InviteLink

	return params, nil
}

// ApproveChatJoinRequestConfig allows you to approve a chat join request.
type ApproveChatJoinRequestConfig struct {
	ChatConfig
	UserID int64
}

func (ApproveChatJoinRequestConfig) method() string {
	return "approveChatJoinRequest"
}

func (config ApproveChatJoinRequestConfig) params() (Params, error) {
	params := make(Params)

	params.AddFirstValid("chat_id", config.ChatID, config.SuperGroupUsername)
	params.AddNonZero("user_id", int(config.UserID))

	return params, nil
}

// DeclineChatJoinRequest allows you to decline a chat join request.
type DeclineChatJoinRequest struct {
	ChatConfig
	UserID int64
}

func (DeclineChatJoinRequest) method() string {
	return "declineChatJoinRequest"
}

func (config DeclineChatJoinRequest) params() (Params, error) {
	params := make(Params)

	params.AddFirstValid("chat_id", config.ChatID, config.SuperGroupUsername)
	params.AddNonZero("user_id", int(config.UserID))

	return params, nil
}

// LeaveChatConfig allows you to leave a chat.
type LeaveChatConfig struct {
	ChatID          int64
	ChannelUsername string
}

func (config LeaveChatConfig) method() string {
	return "leaveChat"
}

func (config LeaveChatConfig) params() (Params, error) {
	params := make(Params)

	params.AddFirstValid("chat_id", config.ChatID, config.ChannelUsername)

	return params, nil
}

// ChatConfigWithUser contains information about a chat and a user.
type ChatConfigWithUser struct {
	ChatID             int64
	SuperGroupUsername string
	UserID             int64
}

func (config ChatConfigWithUser) params() (Params, error) {
	params := make(Params)

	params.AddFirstValid("chat_id", config.ChatID, config.SuperGroupUsername)
	params.AddNonZero64("user_id", config.UserID)

	return params, nil
}

// GetChatMemberConfig is information about getting a specific member in a chat.
type GetChatMemberConfig struct {
	ChatConfigWithUser
}

func (GetChatMemberConfig) method() string {
	return "getChatMember"
}

// InvoiceConfig contains information for sendInvoice request.
type InvoiceConfig struct {
	BaseChat
	Title                     string         // required
	Description               string         // required
	Payload                   string         // required
	ProviderToken             string         // required
	Currency                  string         // required
	Prices                    []LabeledPrice // required
	MaxTipAmount              int
	SuggestedTipAmounts       []int
	StartParameter            string
	ProviderData              string
	PhotoURL                  string
	PhotoSize                 int
	PhotoWidth                int
	PhotoHeight               int
	NeedName                  bool
	NeedPhoneNumber           bool
	NeedEmail                 bool
	NeedShippingAddress       bool
	SendPhoneNumberToProvider bool
	SendEmailToProvider       bool
	IsFlexible                bool
}

func (config InvoiceConfig) params() (Params, error) {
	params, err := config.BaseChat.params()
	if err != nil {
		return params, err
	}

	params["title"] = config.Title
	params["description"] = config.Description
	params["payload"] = config.Payload
	params["provider_token"] = config.ProviderToken
	params["currency"] = config.Currency
	if err = params.AddInterface("prices", config.Prices); err != nil {
		return params, err
	}

	params.AddNonZero("max_tip_amount", config.MaxTipAmount)
	err = params.AddInterface("suggested_tip_amounts", config.SuggestedTipAmounts)
	params.AddNonEmpty("start_parameter", config.StartParameter)
	params.AddNonEmpty("provider_data", config.ProviderData)
	params.AddNonEmpty("photo_url", config.PhotoURL)
	params.AddNonZero("photo_size", config.PhotoSize)
	params.AddNonZero("photo_width", config.PhotoWidth)
	params.AddNonZero("photo_height", config.PhotoHeight)
	params.AddBool("need_name", config.NeedName)
	params.AddBool("need_phone_number", config.NeedPhoneNumber)
	params.AddBool("need_email", config.NeedEmail)
	params.AddBool("need_shipping_address", config.NeedShippingAddress)
	params.AddBool("is_flexible", config.IsFlexible)
	params.AddBool("send_phone_number_to_provider", config.SendPhoneNumberToProvider)
	params.AddBool("send_email_to_provider", config.SendEmailToProvider)

	return params, err
}

func (config InvoiceConfig) method() string {
	return "sendInvoice"
}

// ShippingConfig contains information for answerShippingQuery request.
type ShippingConfig struct {
	ShippingQueryID string // required
	OK              bool   // required
	ShippingOptions []ShippingOption
	ErrorMessage    string
}

func (config ShippingConfig) method() string {
	return "answerShippingQuery"
}

func (config ShippingConfig) params() (Params, error) {
	params := make(Params)

	params["shipping_query_id"] = config.ShippingQueryID
	params.AddBool("ok", config.OK)
	err := params.AddInterface("shipping_options", config.ShippingOptions)
	params.AddNonEmpty("error_message", config.ErrorMessage)

	return params, err
}

// PreCheckoutConfig conatins information for answerPreCheckoutQuery request.
type PreCheckoutConfig struct {
	PreCheckoutQueryID string // required
	OK                 bool   // required
	ErrorMessage       string
}

func (config PreCheckoutConfig) method() string {
	return "answerPreCheckoutQuery"
}

func (config PreCheckoutConfig) params() (Params, error) {
	params := make(Params)

	params["pre_checkout_query_id"] = config.PreCheckoutQueryID
	params.AddBool("ok", config.OK)
	params.AddNonEmpty("error_message", config.ErrorMessage)

	return params, nil
}

// DeleteMessageConfig contains information of a message in a chat to delete.
type DeleteMessageConfig struct {
	ChannelUsername string
	ChatID          int64
	MessageID       int
}

func (config DeleteMessageConfig) method() string {
	return "deleteMessage"
}

func (config DeleteMessageConfig) params() (Params, error) {
	params := make(Params)

	params.AddFirstValid("chat_id", config.ChatID, config.ChannelUsername)
	params.AddNonZero("message_id", config.MessageID)

	return params, nil
}

// PinChatMessageConfig contains information of a message in a chat to pin.
type PinChatMessageConfig struct {
	ChatID              int64
	ChannelUsername     string
	MessageID           int
	DisableNotification bool
}

func (config PinChatMessageConfig) method() string {
	return "pinChatMessage"
}

func (config PinChatMessageConfig) params() (Params, error) {
	params := make(Params)

	params.AddFirstValid("chat_id", config.ChatID, config.ChannelUsername)
	params.AddNonZero("message_id", config.MessageID)
	params.AddBool("disable_notification", config.DisableNotification)

	return params, nil
}

// UnpinChatMessageConfig contains information of a chat message to unpin.
//
// If MessageID is not specified, it will unpin the most recent pin.
type UnpinChatMessageConfig struct {
	ChatID          int64
	ChannelUsername string
	MessageID       int
}

func (config UnpinChatMessageConfig) method() string {
	return "unpinChatMessage"
}

func (config UnpinChatMessageConfig) params() (Params, error) {
	params := make(Params)

	params.AddFirstValid("chat_id", config.ChatID, config.ChannelUsername)
	params.AddNonZero("message_id", config.MessageID)

	return params, nil
}

// UnpinAllChatMessagesConfig contains information of all messages to unpin in
// a chat.
type UnpinAllChatMessagesConfig struct {
	ChatID          int64
	ChannelUsername string
}

func (config UnpinAllChatMessagesConfig) method() string {
	return "unpinAllChatMessages"
}

func (config UnpinAllChatMessagesConfig) params() (Params, error) {
	params := make(Params)

	params.AddFirstValid("chat_id", config.ChatID, config.ChannelUsername)

	return params, nil
}

// SetChatPhotoConfig allows you to set a group, supergroup, or channel's photo.
type SetChatPhotoConfig struct {
	BaseFile
}

func (config SetChatPhotoConfig) method() string {
	return "setChatPhoto"
}

func (config SetChatPhotoConfig) files() []RequestFile {
	return []RequestFile{{
		Name: "photo",
		Data: config.File,
	}}
}

// DeleteChatPhotoConfig allows you to delete a group, supergroup, or channel's photo.
type DeleteChatPhotoConfig struct {
	ChatID          int64
	ChannelUsername string
}

func (config DeleteChatPhotoConfig) method() string {
	return "deleteChatPhoto"
}

func (config DeleteChatPhotoConfig) params() (Params, error) {
	params := make(Params)

	params.AddFirstValid("chat_id", config.ChatID, config.ChannelUsername)

	return params, nil
}

// SetChatTitleConfig allows you to set the title of something other than a private chat.
type SetChatTitleConfig struct {
	ChatID          int64
	ChannelUsername string

	Title string
}

func (config SetChatTitleConfig) method() string {
	return "setChatTitle"
}

func (config SetChatTitleConfig) params() (Params, error) {
	params := make(Params)

	params.AddFirstValid("chat_id", config.ChatID, config.ChannelUsername)
	params["title"] = config.Title

	return params, nil
}

// SetChatDescriptionConfig allows you to set the description of a supergroup or channel.
type SetChatDescriptionConfig struct {
	ChatID          int64
	ChannelUsername string

	Description string
}

func (config SetChatDescriptionConfig) method() string {
	return "setChatDescription"
}

func (config SetChatDescriptionConfig) params() (Params, error) {
	params := make(Params)

	params.AddFirstValid("chat_id", config.ChatID, config.ChannelUsername)
	params["description"] = config.Description

	return params, nil
}

// GetStickerSetConfig allows you to get the stickers in a set.
type GetStickerSetConfig struct {
	Name string
}

func (config GetStickerSetConfig) method() string {
	return "getStickerSet"
}

func (config GetStickerSetConfig) params() (Params, error) {
	params := make(Params)

	params["name"] = config.Name

	return params, nil
}

// UploadStickerConfig allows you to upload a sticker for use in a set later.
type UploadStickerConfig struct {
	UserID     int64
	PNGSticker RequestFileData
}

func (config UploadStickerConfig) method() string {
	return "uploadStickerFile"
}

func (config UploadStickerConfig) params() (Params, error) {
	params := make(Params)

	params.AddNonZero64("user_id", config.UserID)

	return params, nil
}

func (config UploadStickerConfig) files() []RequestFile {
	return []RequestFile{{
		Name: "png_sticker",
		Data: config.PNGSticker,
	}}
}

// NewStickerSetConfig allows creating a new sticker set.
//
// You must set either PNGSticker or TGSSticker.
type NewStickerSetConfig struct {
	UserID        int64
	Name          string
	Title         string
	PNGSticker    RequestFileData
	TGSSticker    RequestFileData
	Emojis        string
	ContainsMasks bool
	MaskPosition  *MaskPosition
}

func (config NewStickerSetConfig) method() string {
	return "createNewStickerSet"
}

func (config NewStickerSetConfig) params() (Params, error) {
	params := make(Params)

	params.AddNonZero64("user_id", config.UserID)
	params["name"] = config.Name
	params["title"] = config.Title

	params["emojis"] = config.Emojis

	params.AddBool("contains_masks", config.ContainsMasks)

	err := params.AddInterface("mask_position", config.MaskPosition)

	return params, err
}

func (config NewStickerSetConfig) files() []RequestFile {
	if config.PNGSticker != nil {
		return []RequestFile{{
			Name: "png_sticker",
			Data: config.PNGSticker,
		}}
	}

	return []RequestFile{{
		Name: "tgs_sticker",
		Data: config.TGSSticker,
	}}
}

// AddStickerConfig allows you to add a sticker to a set.
type AddStickerConfig struct {
	UserID       int64
	Name         string
	PNGSticker   RequestFileData
	TGSSticker   RequestFileData
	Emojis       string
	MaskPosition *MaskPosition
}

func (config AddStickerConfig) method() string {
	return "addStickerToSet"
}

func (config AddStickerConfig) params() (Params, error) {
	params := make(Params)

	params.AddNonZero64("user_id", config.UserID)
	params["name"] = config.Name
	params["emojis"] = config.Emojis

	err := params.AddInterface("mask_position", config.MaskPosition)

	return params, err
}

func (config AddStickerConfig) files() []RequestFile {
	if config.PNGSticker != nil {
		return []RequestFile{{
			Name: "png_sticker",
			Data: config.PNGSticker,
		}}
	}

	return []RequestFile{{
		Name: "tgs_sticker",
		Data: config.TGSSticker,
	}}

}

// SetStickerPositionConfig allows you to change the position of a sticker in a set.
type SetStickerPositionConfig struct {
	Sticker  string
	Position int
}

func (config SetStickerPositionConfig) method() string {
	return "setStickerPositionInSet"
}

func (config SetStickerPositionConfig) params() (Params, error) {
	params := make(Params)

	params["sticker"] = config.Sticker
	params.AddNonZero("position", config.Position)

	return params, nil
}

// DeleteStickerConfig allows you to delete a sticker from a set.
type DeleteStickerConfig struct {
	Sticker string
}

func (config DeleteStickerConfig) method() string {
	return "deleteStickerFromSet"
}

func (config DeleteStickerConfig) params() (Params, error) {
	params := make(Params)

	params["sticker"] = config.Sticker

	return params, nil
}

// SetStickerSetThumbConfig allows you to set the thumbnail for a sticker set.
type SetStickerSetThumbConfig struct {
	Name   string
	UserID int64
	Thumb  RequestFileData
}

func (config SetStickerSetThumbConfig) method() string {
	return "setStickerSetThumb"
}

func (config SetStickerSetThumbConfig) params() (Params, error) {
	params := make(Params)

	params["name"] = config.Name
	params.AddNonZero64("user_id", config.UserID)

	return params, nil
}

func (config SetStickerSetThumbConfig) files() []RequestFile {
	return []RequestFile{{
		Name: "thumb",
		Data: config.Thumb,
	}}
}

// SetChatStickerSetConfig allows you to set the sticker set for a supergroup.
type SetChatStickerSetConfig struct {
	ChatID             int64
	SuperGroupUsername string

	StickerSetName string
}

func (config SetChatStickerSetConfig) method() string {
	return "setChatStickerSet"
}

func (config SetChatStickerSetConfig) params() (Params, error) {
	params := make(Params)

	params.AddFirstValid("chat_id", config.ChatID, config.SuperGroupUsername)
	params["sticker_set_name"] = config.StickerSetName

	return params, nil
}

// DeleteChatStickerSetConfig allows you to remove a supergroup's sticker set.
type DeleteChatStickerSetConfig struct {
	ChatID             int64
	SuperGroupUsername string
}

func (config DeleteChatStickerSetConfig) method() string {
	return "deleteChatStickerSet"
}

func (config DeleteChatStickerSetConfig) params() (Params, error) {
	params := make(Params)

	params.AddFirstValid("chat_id", config.ChatID, config.SuperGroupUsername)

	return params, nil
}

// MediaGroupConfig allows you to send a group of media.
//
// Media consist of InputMedia items (InputMediaPhoto, InputMediaVideo).
type MediaGroupConfig struct {
	ChatID          int64
	ChannelUsername string

	Media               []interface{}
	DisableNotification bool
	ReplyToMessageID    int
}

func (config MediaGroupConfig) method() string {
	return "sendMediaGroup"
}

func (config MediaGroupConfig) params() (Params, error) {
	params := make(Params)

	params.AddFirstValid("chat_id", config.ChatID, config.ChannelUsername)
	params.AddBool("disable_notification", config.DisableNotification)
	params.AddNonZero("reply_to_message_id", config.ReplyToMessageID)

	err := params.AddInterface("media", prepareInputMediaForParams(config.Media))

	return params, err
}

func (config MediaGroupConfig) files() []RequestFile {
	return prepareInputMediaForFiles(config.Media)
}

// DiceConfig contains information about a sendDice request.
type DiceConfig struct {
	BaseChat
	// Emoji on which the dice throw animation is based.
	// Currently, must be one of 🎲, 🎯, 🏀, ⚽, 🎳, or 🎰.
	// Dice can have values 1-6 for 🎲, 🎯, and 🎳, values 1-5 for 🏀 and ⚽,
	// and values 1-64 for 🎰.
	// Defaults to “🎲”
	Emoji string
}

func (config DiceConfig) method() string {
	return "sendDice"
}

func (config DiceConfig) params() (Params, error) {
	params, err := config.BaseChat.params()
	if err != nil {
		return params, err
	}

	params.AddNonEmpty("emoji", config.Emoji)

	return params, err
}

// GetMyCommandsConfig gets a list of the currently registered commands.
type GetMyCommandsConfig struct {
	Scope        *BotCommandScope
	LanguageCode string
}

func (config GetMyCommandsConfig) method() string {
	return "getMyCommands"
}

func (config GetMyCommandsConfig) params() (Params, error) {
	params := make(Params)

	err := params.AddInterface("scope", config.Scope)
	params.AddNonEmpty("language_code", config.LanguageCode)

	return params, err
}

// SetMyCommandsConfig sets a list of commands the bot understands.
type SetMyCommandsConfig struct {
	Commands     []BotCommand
	Scope        *BotCommandScope
	LanguageCode string
}

func (config SetMyCommandsConfig) method() string {
	return "setMyCommands"
}

func (config SetMyCommandsConfig) params() (Params, error) {
	params := make(Params)

	if err := params.AddInterface("commands", config.Commands); err != nil {
		return params, err
	}
	err := params.AddInterface("scope", config.Scope)
	params.AddNonEmpty("language_code", config.LanguageCode)

	return params, err
}

type DeleteMyCommandsConfig struct {
	Scope        *BotCommandScope
	LanguageCode string
}

func (config DeleteMyCommandsConfig) method() string {
	return "deleteMyCommands"
}

func (config DeleteMyCommandsConfig) params() (Params, error) {
	params := make(Params)

	err := params.AddInterface("scope", config.Scope)
	params.AddNonEmpty("language_code", config.LanguageCode)

	return params, err
}

// prepareInputMediaParam evaluates a single InputMedia and determines if it
// needs to be modified for a successful upload. If it returns nil, then the
// value does not need to be included in the params. Otherwise, it will return
// the same type as was originally provided.
//
// The idx is used to calculate the file field name. If you only have a single
// file, 0 may be used. It is formatted into "attach://file-%d" for the primary
// media and "attach://file-%d-thumb" for thumbnails.
//
// It is expected to be used in conjunction with prepareInputMediaFile.
func prepareInputMediaParam(inputMedia interface{}, idx int) interface{} {
	switch m := inputMedia.(type) {
	case InputMediaPhoto:
		if m.Media.NeedsUpload() {
			m.Media = fileAttach(fmt.Sprintf("attach://file-%d", idx))
		}

		return m
	case InputMediaVideo:
		if m.Media.NeedsUpload() {
			m.Media = fileAttach(fmt.Sprintf("attach://file-%d", idx))
		}

		if m.Thumb != nil && m.Thumb.NeedsUpload() {
			m.Thumb = fileAttach(fmt.Sprintf("attach://file-%d-thumb", idx))
		}

		return m
	case InputMediaAudio:
		if m.Media.NeedsUpload() {
			m.Media = fileAttach(fmt.Sprintf("attach://file-%d", idx))
		}

		if m.Thumb != nil && m.Thumb.NeedsUpload() {
			m.Thumb = fileAttach(fmt.Sprintf("attach://file-%d-thumb", idx))
		}

		return m
	case InputMediaDocument:
		if m.Media.NeedsUpload() {
			m.Media = fileAttach(fmt.Sprintf("attach://file-%d", idx))
		}

		if m.Thumb != nil && m.Thumb.NeedsUpload() {
			m.Thumb = fileAttach(fmt.Sprintf("attach://file-%d-thumb", idx))
		}

		return m
	}

	return nil
}

// prepareInputMediaFile generates an array of RequestFile to provide for
// Fileable's files method. It returns an array as a single InputMedia may have
// multiple files, for the primary media and a thumbnail.
//
// The idx parameter is used to generate file field names. It uses the names
// "file-%d" for the main file and "file-%d-thumb" for the thumbnail.
//
// It is expected to be used in conjunction with prepareInputMediaParam.
func prepareInputMediaFile(inputMedia interface{}, idx int) []RequestFile {
	files := []RequestFile{}

	switch m := inputMedia.(type) {
	case InputMediaPhoto:
		if m.Media.NeedsUpload() {
			files = append(files, RequestFile{
				Name: fmt.Sprintf("file-%d", idx),
				Data: m.Media,
			})
		}
	case InputMediaVideo:
		if m.Media.NeedsUpload() {
			files = append(files, RequestFile{
				Name: fmt.Sprintf("file-%d", idx),
				Data: m.Media,
			})
		}

		if m.Thumb != nil && m.Thumb.NeedsUpload() {
			files = append(files, RequestFile{
				Name: fmt.Sprintf("file-%d", idx),
				Data: m.Thumb,
			})
		}
	case InputMediaDocument:
		if m.Media.NeedsUpload() {
			files = append(files, RequestFile{
				Name: fmt.Sprintf("file-%d", idx),
				Data: m.Media,
			})
		}

		if m.Thumb != nil && m.Thumb.NeedsUpload() {
			files = append(files, RequestFile{
				Name: fmt.Sprintf("file-%d", idx),
				Data: m.Thumb,
			})
		}
	case InputMediaAudio:
		if m.Media.NeedsUpload() {
			files = append(files, RequestFile{
				Name: fmt.Sprintf("file-%d", idx),
				Data: m.Media,
			})
		}

		if m.Thumb != nil && m.Thumb.NeedsUpload() {
			files = append(files, RequestFile{
				Name: fmt.Sprintf("file-%d", idx),
				Data: m.Thumb,
			})
		}
	}

	return files
}

// prepareInputMediaForParams calls prepareInputMediaParam for each item
// provided and returns a new array with the correct params for a request.
//
// It is expected that files will get data from the associated function,
// prepareInputMediaForFiles.
func prepareInputMediaForParams(inputMedia []interface{}) []interface{} {
	newMedia := make([]interface{}, len(inputMedia))
	copy(newMedia, inputMedia)

	for idx, media := range inputMedia {
		if param := prepareInputMediaParam(media, idx); param != nil {
			newMedia[idx] = param
		}
	}

	return newMedia
}

// prepareInputMediaForFiles calls prepareInputMediaFile for each item
// provided and returns a new array with the correct files for a request.
//
// It is expected that params will get data from the associated function,
// prepareInputMediaForParams.
func prepareInputMediaForFiles(inputMedia []interface{}) []RequestFile {
	files := []RequestFile{}

	for idx, media := range inputMedia {
		if file := prepareInputMediaFile(media, idx); file != nil {
			files = append(files, file...)
		}
	}

	return files
}
