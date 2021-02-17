package object // import "github.com/SevereCloud/vksdk/v2/object"

import (
	"fmt"
)

// DocsDoc struct.
type DocsDoc struct {
	AccessKey  string         `json:"access_key"` // Access key for the document
	Date       int            `json:"date"`       // Date when file has been uploaded in Unixtime
	Ext        string         `json:"ext"`        // File extension
	ID         int            `json:"id"`         // Document ID
	IsLicensed BaseBoolInt    `json:"is_licensed"`
	OwnerID    int            `json:"owner_id"` // Document owner ID
	Preview    DocsDocPreview `json:"preview"`
	Size       int            `json:"size"`  // File size in bites
	Title      string         `json:"title"` // Document title
	Type       int            `json:"type"`  // Document type
	URL        string         `json:"url"`   // File URL
	DocsDocPreviewAudioMessage
	DocsDocPreviewGraffiti
}

// ToAttachment return attachment format.
func (doc DocsDoc) ToAttachment() string {
	return fmt.Sprintf("doc%d_%d", doc.OwnerID, doc.ID)
}

// DocsDocPreview struct.
type DocsDocPreview struct {
	Photo        DocsDocPreviewPhoto        `json:"photo"`
	Graffiti     DocsDocPreviewGraffiti     `json:"graffiti"`
	Video        DocsDocPreviewVideo        `json:"video"`
	AudioMessage DocsDocPreviewAudioMessage `json:"audio_message"`
}

// DocsDocPreviewPhoto struct.
type DocsDocPreviewPhoto struct {
	Sizes []DocsDocPreviewPhotoSizes `json:"sizes"`
}

// MaxSize return the largest DocsDocPreviewPhotoSizes.
func (photo DocsDocPreviewPhoto) MaxSize() (maxPhotoSize DocsDocPreviewPhotoSizes) {
	var max float64

	for _, photoSize := range photo.Sizes {
		size := photoSize.Height * photoSize.Width
		if size > max {
			max = size
			maxPhotoSize = photoSize
		}
	}

	return
}

// MinSize return the smallest DocsDocPreviewPhotoSizes.
func (photo DocsDocPreviewPhoto) MinSize() (minPhotoSize DocsDocPreviewPhotoSizes) {
	var min float64

	for _, photoSize := range photo.Sizes {
		size := photoSize.Height * photoSize.Width
		if size < min || min == 0 {
			min = size
			minPhotoSize = photoSize
		}
	}

	return
}

// DocsDocPreviewPhotoSizes struct.
type DocsDocPreviewPhotoSizes struct {
	// BUG(VK): json: cannot unmarshal number 162.000000 into Go struct field
	// DocsDocPreviewPhotoSizes.doc.preview.photo.sizes.height of type Int
	Height float64 `json:"height"` // Height in px
	Src    string  `json:"src"`    // URL of the image
	Type   string  `json:"type"`
	Width  float64 `json:"width"` // Width in px
}

// DocsDocPreviewGraffiti struct.
type DocsDocPreviewGraffiti struct {
	Src    string `json:"src"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}

// DocsDocPreviewVideo struct.
type DocsDocPreviewVideo struct {
	FileSize int    `json:"file_size"` // Video file size in bites
	Height   int    `json:"height"`    // Video's height in pixels
	Src      string `json:"src"`       // Video URL
	Width    int    `json:"width"`     // Video's width in pixels
}

// DocsDocPreviewAudioMessage struct.
type DocsDocPreviewAudioMessage struct {
	Duration        int    `json:"duration"`
	Waveform        []int  `json:"waveform"`
	LinkOgg         string `json:"link_ogg"`
	LinkMp3         string `json:"link_mp3"`
	Transcript      string `json:"transcript"`
	TranscriptState string `json:"transcript_state"`
}

// DocsDocTypes struct.
type DocsDocTypes struct {
	Count int    `json:"count"` // Number of docs
	ID    int    `json:"id"`    // Doc type ID
	Name  string `json:"name"`  // Doc type Title
}

// DocsDocUploadResponse struct.
type DocsDocUploadResponse struct {
	File string `json:"file"` // Uploaded file data
}
