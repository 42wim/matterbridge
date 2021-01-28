package object // import "github.com/SevereCloud/vksdk/v2/object"

import (
	"fmt"
)

// NotesNote struct.
type NotesNote struct {
	CanComment     BaseBoolInt   `json:"can_comment"` // Information whether current user can comment the note
	Comments       int           `json:"comments"`    // Comments number
	Date           int           `json:"date"`        // Date when the note has been created in Unixtime
	ID             int           `json:"id"`          // Note ID
	OwnerID        int           `json:"owner_id"`    // Note owner's ID
	Text           string        `json:"text"`        // Note text
	TextWiki       string        `json:"text_wiki"`   // Note text in wiki format
	Title          string        `json:"title"`       // Note title
	ViewURL        string        `json:"view_url"`    // URL of the page with note preview
	ReadComments   int           `json:"read_comments"`
	PrivacyView    []interface{} `json:"privacy_view"`    // NOTE: old type privacy
	PrivacyComment []interface{} `json:"privacy_comment"` // NOTE: old type privacy
}

// ToAttachment return attachment format.
func (note NotesNote) ToAttachment() string {
	return fmt.Sprintf("note%d_%d", note.OwnerID, note.ID)
}

// NotesNoteComment struct.
type NotesNoteComment struct {
	Date    int    `json:"date"`     // Date when the comment has been added in Unixtime
	ID      int    `json:"id"`       // Comment ID
	Message string `json:"message"`  // Comment text
	NID     int    `json:"nid"`      // Note ID
	OID     int    `json:"oid"`      // Note ID
	ReplyTo int    `json:"reply_to"` // ID of replied comment
	UID     int    `json:"uid"`      // Comment author's ID
}
