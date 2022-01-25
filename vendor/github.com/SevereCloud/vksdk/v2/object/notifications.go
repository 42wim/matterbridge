package object // import "github.com/SevereCloud/vksdk/v2/object"

// NotificationsFeedback struct.
type NotificationsFeedback struct {
	Attachments []WallWallpostAttachment `json:"attachments"`
	FromID      int                      `json:"from_id"` // Reply author's ID
	Geo         BaseGeo                  `json:"geo"`
	ID          int                      `json:"id"` // Item ID
	Likes       BaseLikesInfo            `json:"likes"`
	Text        string                   `json:"text"`  // Reply text
	ToID        int                      `json:"to_id"` // Wall owner's ID
}

// NotificationsNotification struct.
type NotificationsNotification struct {
	Date     int                `json:"date"` // Date when the event has been occurred
	Feedback RawMessage         `json:"feedback"`
	Parent   RawMessage         `json:"parent"`
	Reply    NotificationsReply `json:"reply"`
	Type     string             `json:"type"` // Notification type
}

// NotificationsNotificationsComment struct.
type NotificationsNotificationsComment struct {
	Date    int          `json:"date"`     // Date when the comment has been added in Unixtime
	ID      int          `json:"id"`       // Comment ID
	OwnerID int          `json:"owner_id"` // Author ID
	Photo   PhotosPhoto  `json:"photo"`
	Post    WallWallpost `json:"post"`
	Text    string       `json:"text"` // Comment text
	Topic   BoardTopic   `json:"topic"`
	Video   VideoVideo   `json:"video"`
}

// NotificationsReply struct.
type NotificationsReply struct {
	Date string `json:"date"` // Date when the reply has been created in Unixtime
	ID   int    `json:"id"`   // Reply ID
	Text string `json:"text"` // Reply text
}
