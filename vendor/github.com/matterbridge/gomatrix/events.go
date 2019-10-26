package gomatrix

import (
	"html"
	"regexp"
)

// Event represents a single Matrix event.
type Event struct {
	StateKey  *string                `json:"state_key,omitempty"` // The state key for the event. Only present on State Events.
	Sender    string                 `json:"sender"`              // The user ID of the sender of the event
	Type      string                 `json:"type"`                // The event type
	Timestamp int64                  `json:"origin_server_ts"`    // The unix timestamp when this message was sent by the origin server
	ID        string                 `json:"event_id"`            // The unique ID of this event
	RoomID    string                 `json:"room_id"`             // The room the event was sent to. May be nil (e.g. for presence)
	Content   map[string]interface{} `json:"content"`             // The JSON content of the event.
	Redacts   string                 `json:"redacts,omitempty"`   // The event ID that was redacted if a m.room.redaction event
	Unsigned  map[string]interface{} `json:"unsigned"`            // The unsigned portions of the event, such as age and prev_content
}

// Body returns the value of the "body" key in the event content if it is
// present and is a string.
func (event *Event) Body() (body string, ok bool) {
	value, exists := event.Content["body"]
	if !exists {
		return
	}
	body, ok = value.(string)
	return
}

// MessageType returns the value of the "msgtype" key in the event content if
// it is present and is a string.
func (event *Event) MessageType() (msgtype string, ok bool) {
	value, exists := event.Content["msgtype"]
	if !exists {
		return
	}
	msgtype, ok = value.(string)
	return
}

// TextMessage is the contents of a Matrix formated message event.
type TextMessage struct {
	MsgType       string `json:"msgtype"`
	Body          string `json:"body"`
	Format        string `json:"format,omitempty"`
	FormattedBody string `json:"formatted_body,omitempty"`
}

// ThumbnailInfo contains info about an thumbnail image - http://matrix.org/docs/spec/client_server/r0.2.0.html#m-image
type ThumbnailInfo struct {
	Height   uint   `json:"h,omitempty"`
	Width    uint   `json:"w,omitempty"`
	Mimetype string `json:"mimetype,omitempty"`
	Size     uint   `json:"size,omitempty"`
}

// ImageInfo contains info about an image - http://matrix.org/docs/spec/client_server/r0.2.0.html#m-image
type ImageInfo struct {
	Height        uint          `json:"h,omitempty"`
	Width         uint          `json:"w,omitempty"`
	Mimetype      string        `json:"mimetype,omitempty"`
	Size          uint          `json:"size,omitempty"`
	ThumbnailInfo ThumbnailInfo `json:"thumbnail_info,omitempty"`
	ThumbnailURL  string        `json:"thumbnail_url,omitempty"`
}

// VideoInfo contains info about a video - http://matrix.org/docs/spec/client_server/r0.2.0.html#m-video
type VideoInfo struct {
	Mimetype      string        `json:"mimetype,omitempty"`
	ThumbnailInfo ThumbnailInfo `json:"thumbnail_info"`
	ThumbnailURL  string        `json:"thumbnail_url,omitempty"`
	Height        uint          `json:"h,omitempty"`
	Width         uint          `json:"w,omitempty"`
	Duration      uint          `json:"duration,omitempty"`
	Size          uint          `json:"size,omitempty"`
}

// AudioInfo contains info about a file - http://matrix.org/docs/spec/client_server/r0.2.0.html#m-audio
type AudioInfo struct {
	Mimetype string `json:"mimetype,omitempty"`
	Size     uint   `json:"size,omitempty"`
	Duration uint   `json:"duration,omitempty"`
}

// FileInfo contains info about a file - http://matrix.org/docs/spec/client_server/r0.2.0.html#m-file
type FileInfo struct {
	Mimetype      string    `json:"mimetype,omitempty"`
	ThumbnailInfo ImageInfo `json:"thumbnail_info"`
	ThumbnailURL  string    `json:"thumbnail_url,omitempty"`
	Size          uint      `json:"size,omitempty"`
}

// VideoMessage is an m.video  - http://matrix.org/docs/spec/client_server/r0.2.0.html#m-video
type VideoMessage struct {
	MsgType string    `json:"msgtype"`
	Body    string    `json:"body"`
	URL     string    `json:"url"`
	Info    VideoInfo `json:"info"`
}

// ImageMessage is an m.image event
type ImageMessage struct {
	MsgType string    `json:"msgtype"`
	Body    string    `json:"body"`
	URL     string    `json:"url"`
	Info    ImageInfo `json:"info"`
}

// AudioMessage is an m.audio event
type AudioMessage struct {
	MsgType string    `json:"msgtype"`
	Body    string    `json:"body"`
	URL     string    `json:"url"`
	Info    AudioInfo `json:"info"`
}

// FileMessage is a m.file event
type FileMessage struct {
	MsgType string   `json:"msgtype"`
	Body    string   `json:"body"`
	URL     string   `json:"url"`
	Info    FileInfo `json:"info"`
}

// An HTMLMessage is the contents of a Matrix HTML formated message event.
type HTMLMessage struct {
	Body          string `json:"body"`
	MsgType       string `json:"msgtype"`
	Format        string `json:"format"`
	FormattedBody string `json:"formatted_body"`
}

var htmlRegex = regexp.MustCompile("<[^<]+?>")

// GetHTMLMessage returns an HTMLMessage with the body set to a stripped version of the provided HTML, in addition
// to the provided HTML.
func GetHTMLMessage(msgtype, htmlText string) HTMLMessage {
	return HTMLMessage{
		Body:          html.UnescapeString(htmlRegex.ReplaceAllLiteralString(htmlText, "")),
		MsgType:       msgtype,
		Format:        "org.matrix.custom.html",
		FormattedBody: htmlText,
	}
}
