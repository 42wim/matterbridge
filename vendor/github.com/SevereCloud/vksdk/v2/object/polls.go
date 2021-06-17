package object // import "github.com/SevereCloud/vksdk/v2/object"

import (
	"fmt"
)

// PollsAnswer struct.
type PollsAnswer struct {
	ID    int     `json:"id"`
	Rate  float64 `json:"rate"`
	Text  string  `json:"text"`
	Votes int     `json:"votes"`
}

// PollsPoll struct.
type PollsPoll struct {
	AnswerID      int             `json:"answer_id"` // Current user's answer ID
	Answers       []PollsAnswer   `json:"answers"`
	Created       int             `json:"created"`  // Date when poll has been created in Unixtime
	ID            int             `json:"id"`       // Poll ID
	OwnerID       int             `json:"owner_id"` // Poll owner's ID
	Question      string          `json:"question"` // Poll question
	Votes         int             `json:"votes"`    // Votes number
	AnswerIDs     []int           `json:"answer_ids"`
	EndDate       int             `json:"end_date"`
	Anonymous     BaseBoolInt     `json:"anonymous"` // Information whether the pole is anonymous
	Closed        BaseBoolInt     `json:"closed"`
	IsBoard       BaseBoolInt     `json:"is_board"`
	CanEdit       BaseBoolInt     `json:"can_edit"`
	CanVote       BaseBoolInt     `json:"can_vote"`
	CanReport     BaseBoolInt     `json:"can_report"`
	CanShare      BaseBoolInt     `json:"can_share"`
	Multiple      BaseBoolInt     `json:"multiple"`
	DisableUnvote BaseBoolInt     `json:"disable_unvote"`
	Photo         PhotosPhoto     `json:"photo"`
	AuthorID      int             `json:"author_id"`
	Background    PollsBackground `json:"background"`
	Friends       []PollsFriend   `json:"friends"`
	Profiles      []UsersUser     `json:"profiles"`
	Groups        []GroupsGroup   `json:"groups"`
	EmbedHash     string          `json:"embed_hash"`
}

// ToAttachment return attachment format.
func (poll PollsPoll) ToAttachment() string {
	return fmt.Sprintf("poll%d_%d", poll.OwnerID, poll.ID)
}

// PollsFriend struct.
type PollsFriend struct {
	ID int `json:"id"`
}

// PollsVoters struct.
type PollsVoters struct {
	AnswerID int              `json:"answer_id"` // Answer ID
	Users    PollsVotersUsers `json:"users"`
}

// PollsVotersUsers struct.
type PollsVotersUsers struct {
	Count int   `json:"count"` // Votes number
	Items []int `json:"items"`
}

// PollsVotersFields struct.
type PollsVotersFields struct {
	AnswerID int                    `json:"answer_id"` // Answer ID
	Users    PollsVotersUsersFields `json:"users"`
}

// PollsVotersUsersFields struct.
type PollsVotersUsersFields struct {
	Count int         `json:"count"` // Votes number
	Items []UsersUser `json:"items"`
}

// PollsBackground struct.
type PollsBackground struct {
	Type   string `json:"type"`
	Angle  int    `json:"angle"`
	Color  string `json:"color"`
	Points []struct {
		Position float64 `json:"position"`
		Color    string  `json:"color"`
	} `json:"points"`
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// PollsPhoto struct.
type PollsPhoto struct {
	ID     int           `json:"id"`
	Color  string        `json:"color"`
	Images []PhotosImage `json:"images"`
}

// PollsPhotoUploadResponse struct.
type PollsPhotoUploadResponse struct {
	Photo string `json:"photo"` // Uploaded photo data
	Hash  string `json:"hash"`  // Uploaded hash
}
