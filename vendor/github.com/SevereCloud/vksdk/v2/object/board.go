package object // import "github.com/SevereCloud/vksdk/v2/object"

// BoardTopic struct.
type BoardTopic struct {
	Comments  int         `json:"comments"`   // Comments number
	Created   int         `json:"created"`    // Date when the topic has been created in Unixtime
	CreatedBy int         `json:"created_by"` // Creator ID
	ID        int         `json:"id"`         // Topic ID
	IsClosed  BaseBoolInt `json:"is_closed"`  // Information whether the topic is closed
	IsFixed   BaseBoolInt `json:"is_fixed"`   // Information whether the topic is fixed
	Title     string      `json:"title"`      // Topic title
	Updated   int         `json:"updated"`    // Date when the topic has been updated in Unixtime
	UpdatedBy int         `json:"updated_by"` // ID of user who updated the topic
}

// BoardTopicComment struct.
type BoardTopicComment struct {
	Attachments []WallCommentAttachment `json:"attachments"`
	Date        int                     `json:"date"`    // Date when the comment has been added in Unixtime
	FromID      int                     `json:"from_id"` // Author ID
	ID          int                     `json:"id"`      // Comment ID
	// RealOffset   int                     `json:"real_offset"` // Real position of the comment
	Text string `json:"text"` // Comment text
	// TopicID      int                     `json:"topic_id"`
	// TopicOwnerID int                     `json:"topic_owner_id"`
	Likes   BaseLikesInfo `json:"likes"`
	CanEdit BaseBoolInt   `json:"can_edit"` // Information whether current user can edit the comment
}

// BoardTopicPoll struct.
type BoardTopicPoll struct {
	AnswerID int           `json:"answer_id"` // Current user's answer ID
	Answers  []PollsAnswer `json:"answers"`
	Created  int           `json:"created"`   // Date when poll has been created in Unixtime
	IsClosed BaseBoolInt   `json:"is_closed"` // Information whether the poll is closed
	OwnerID  int           `json:"owner_id"`  // Poll owner's ID
	PollID   int           `json:"poll_id"`   // Poll ID
	Question string        `json:"question"`  // Poll question
	Votes    string        `json:"votes"`     // Votes number
}
