// Copyright (c) 2020 Tulir Asokan
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package event

import (
	"encoding/json"
	"strconv"

	"maunium.net/go/mautrix/crypto/attachment"
	"maunium.net/go/mautrix/id"
)

// MessageType is the sub-type of a m.room.message event.
// https://matrix.org/docs/spec/client_server/r0.6.0#m-room-message-msgtypes
type MessageType string

// Msgtypes
const (
	MsgText     MessageType = "m.text"
	MsgEmote    MessageType = "m.emote"
	MsgNotice   MessageType = "m.notice"
	MsgImage    MessageType = "m.image"
	MsgLocation MessageType = "m.location"
	MsgVideo    MessageType = "m.video"
	MsgAudio    MessageType = "m.audio"
	MsgFile     MessageType = "m.file"

	MsgVerificationRequest MessageType = "m.key.verification.request"
)

// Format specifies the format of the formatted_body in m.room.message events.
// https://matrix.org/docs/spec/client_server/r0.6.0#m-room-message-msgtypes
type Format string

// Message formats
const (
	FormatHTML Format = "org.matrix.custom.html"
)

// RedactionEventContent represents the content of a m.room.redaction message event.
//
// The redacted event ID is still at the top level, but will move in a future room version.
// See https://github.com/matrix-org/matrix-doc/pull/2244 and https://github.com/matrix-org/matrix-doc/pull/2174
//
// https://matrix.org/docs/spec/client_server/r0.6.0#m-room-redaction
type RedactionEventContent struct {
	Reason string `json:"reason,omitempty"`
}

// ReactionEventContent represents the content of a m.reaction message event.
// This is not yet in a spec release, see https://github.com/matrix-org/matrix-doc/pull/1849
type ReactionEventContent struct {
	RelatesTo RelatesTo `json:"m.relates_to"`
}

func (content *ReactionEventContent) GetRelatesTo() *RelatesTo {
	return &content.RelatesTo
}

func (content *ReactionEventContent) OptionalGetRelatesTo() *RelatesTo {
	return &content.RelatesTo
}

func (content *ReactionEventContent) SetRelatesTo(rel *RelatesTo) {
	content.RelatesTo = *rel
}

// MssageEventContent represents the content of a m.room.message event.
//
// It is also used to represent m.sticker events, as they are equivalent to m.room.message
// with the exception of the msgtype field.
//
// https://matrix.org/docs/spec/client_server/r0.6.0#m-room-message
type MessageEventContent struct {
	// Base m.room.message fields
	MsgType MessageType `json:"msgtype"`
	Body    string      `json:"body"`

	// Extra fields for text types
	Format        Format `json:"format,omitempty"`
	FormattedBody string `json:"formatted_body,omitempty"`

	// Extra field for m.location
	GeoURI string `json:"geo_uri,omitempty"`

	// Extra fields for media types
	URL  id.ContentURIString `json:"url,omitempty"`
	Info *FileInfo           `json:"info,omitempty"`
	File *EncryptedFileInfo  `json:"file,omitempty"`

	// Edits and relations
	NewContent *MessageEventContent `json:"m.new_content,omitempty"`
	RelatesTo  *RelatesTo           `json:"m.relates_to,omitempty"`

	// In-room verification
	To         id.UserID            `json:"to,omitempty"`
	FromDevice id.DeviceID          `json:"from_device,omitempty"`
	Methods    []VerificationMethod `json:"methods,omitempty"`

	replyFallbackRemoved bool
}

func (content *MessageEventContent) GetRelatesTo() *RelatesTo {
	if content.RelatesTo == nil {
		content.RelatesTo = &RelatesTo{}
	}
	return content.RelatesTo
}

func (content *MessageEventContent) OptionalGetRelatesTo() *RelatesTo {
	return content.RelatesTo
}

func (content *MessageEventContent) SetRelatesTo(rel *RelatesTo) {
	content.RelatesTo = rel
}

func (content *MessageEventContent) GetFile() *EncryptedFileInfo {
	if content.File == nil {
		content.File = &EncryptedFileInfo{}
	}
	return content.File
}

func (content *MessageEventContent) GetInfo() *FileInfo {
	if content.Info == nil {
		content.Info = &FileInfo{}
	}
	return content.Info
}

type EncryptedFileInfo struct {
	attachment.EncryptedFile
	URL id.ContentURIString `json:"url"`
}

type FileInfo struct {
	MimeType      string              `json:"mimetype,omitempty"`
	ThumbnailInfo *FileInfo           `json:"thumbnail_info,omitempty"`
	ThumbnailURL  id.ContentURIString `json:"thumbnail_url,omitempty"`
	ThumbnailFile *EncryptedFileInfo  `json:"thumbnail_file,omitempty"`
	Width         int                 `json:"-"`
	Height        int                 `json:"-"`
	Duration      int                 `json:"-"`
	Size          int                 `json:"-"`
}

type serializableFileInfo struct {
	MimeType      string                `json:"mimetype,omitempty"`
	ThumbnailInfo *serializableFileInfo `json:"thumbnail_info,omitempty"`
	ThumbnailURL  id.ContentURIString   `json:"thumbnail_url,omitempty"`
	ThumbnailFile *EncryptedFileInfo    `json:"thumbnail_file,omitempty"`

	Width    json.Number `json:"w,omitempty"`
	Height   json.Number `json:"h,omitempty"`
	Duration json.Number `json:"duration,omitempty"`
	Size     json.Number `json:"size,omitempty"`
}

func (sfi *serializableFileInfo) CopyFrom(fileInfo *FileInfo) *serializableFileInfo {
	if fileInfo == nil {
		return nil
	}
	*sfi = serializableFileInfo{
		MimeType:      fileInfo.MimeType,
		ThumbnailURL:  fileInfo.ThumbnailURL,
		ThumbnailInfo: (&serializableFileInfo{}).CopyFrom(fileInfo.ThumbnailInfo),
		ThumbnailFile: fileInfo.ThumbnailFile,
	}
	if fileInfo.Width > 0 {
		sfi.Width = json.Number(strconv.Itoa(fileInfo.Width))
	}
	if fileInfo.Height > 0 {
		sfi.Height = json.Number(strconv.Itoa(fileInfo.Height))
	}
	if fileInfo.Size > 0 {
		sfi.Size = json.Number(strconv.Itoa(fileInfo.Size))

	}
	if fileInfo.Duration > 0 {
		sfi.Duration = json.Number(strconv.Itoa(int(fileInfo.Duration)))
	}
	return sfi
}

func (sfi *serializableFileInfo) CopyTo(fileInfo *FileInfo) {
	*fileInfo = FileInfo{
		Width:        numberToInt(sfi.Width),
		Height:       numberToInt(sfi.Height),
		Size:         numberToInt(sfi.Size),
		Duration:     numberToInt(sfi.Duration),
		MimeType:     sfi.MimeType,
		ThumbnailURL: sfi.ThumbnailURL,
	}
	if sfi.ThumbnailInfo != nil {
		fileInfo.ThumbnailInfo = &FileInfo{}
		sfi.ThumbnailInfo.CopyTo(fileInfo.ThumbnailInfo)
	}
}

func (fileInfo *FileInfo) UnmarshalJSON(data []byte) error {
	sfi := &serializableFileInfo{}
	if err := json.Unmarshal(data, sfi); err != nil {
		return err
	}
	sfi.CopyTo(fileInfo)
	return nil
}

func (fileInfo *FileInfo) MarshalJSON() ([]byte, error) {
	return json.Marshal((&serializableFileInfo{}).CopyFrom(fileInfo))
}

func numberToInt(val json.Number) int {
	f64, _ := val.Float64()
	if f64 > 0 {
		return int(f64)
	}
	return 0
}

func (fileInfo *FileInfo) GetThumbnailInfo() *FileInfo {
	if fileInfo.ThumbnailInfo == nil {
		fileInfo.ThumbnailInfo = &FileInfo{}
	}
	return fileInfo.ThumbnailInfo
}
