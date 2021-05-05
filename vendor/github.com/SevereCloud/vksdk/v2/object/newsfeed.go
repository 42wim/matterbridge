package object // import "github.com/SevereCloud/vksdk/v2/object"

// NewsfeedEventActivity struct.
type NewsfeedEventActivity struct {
	Address      string `json:"address"`       // address of event
	ButtonText   string `json:"button_text"`   // text of attach
	Friends      []int  `json:"friends"`       // array of friends ids
	MemberStatus int    `json:"member_status"` // Current user's member status
	Text         string `json:"text"`          // text of attach
	Time         int    `json:"time"`          // event start time
}

// NewsfeedItemAudio struct.
type NewsfeedItemAudio struct {
	Audio NewsfeedItemAudioAudio `json:"audio"`
}

// NewsfeedItemAudioAudio struct.
type NewsfeedItemAudioAudio struct {
	Count int          `json:"count"` // Audios number
	Items []AudioAudio `json:"items"`
}

// NewsfeedItemDigest struct.
type NewsfeedItemDigest struct {
	ButtonText  string         `json:"button_text"`
	FeedID      string         `json:"feed_id"` // id of feed in digest
	Items       []WallWallpost `json:"items"`
	MainPostIDs []string       `json:"main_post_ids"`
	Template    string         `json:"template"` // type of digest
	Title       string         `json:"title"`
	TrackCode   string         `json:"track_code"`
	// Type        string         `json:"type"`
}

// NewsfeedItemFriend struct.
type NewsfeedItemFriend struct {
	Friends NewsfeedItemFriendFriends `json:"friends"`
}

// NewsfeedItemFriendFriends struct.
type NewsfeedItemFriendFriends struct {
	Count int          `json:"count"` // Number of friends has been added
	Items []BaseUserID `json:"items"`
}

// NewsfeedItemNote struct.
type NewsfeedItemNote struct {
	Notes NewsfeedItemNoteNotes `json:"notes"`
}

// NewsfeedItemNoteNotes struct.
type NewsfeedItemNoteNotes struct {
	Count int                    `json:"count"` // Notes number
	Items []NewsfeedNewsfeedNote `json:"items"`
}

// NewsfeedItemPhoto struct.
type NewsfeedItemPhoto struct {
	Photos NewsfeedItemPhotoPhotos `json:"photos"`
}

// NewsfeedItemPhotoPhotos struct.
type NewsfeedItemPhotoPhotos struct {
	Count int               `json:"count"` // Photos number
	Items []PhotosPhotoFull `json:"items"`
}

// NewsfeedItemPhotoTag struct.
type NewsfeedItemPhotoTag struct {
	PhotoTags NewsfeedItemPhotoTagPhotoTags `json:"photo_tags"`
}

// NewsfeedItemPhotoTagPhotoTags struct.
type NewsfeedItemPhotoTagPhotoTags struct {
	Count int               `json:"count"` // Tags number
	Items []PhotosPhotoFull `json:"items"`
}

// NewsfeedItemStoriesBlock struct.
type NewsfeedItemStoriesBlock struct {
	BlockType string         `json:"block_type"`
	Stories   []StoriesStory `json:"stories"`
	// Title     string         `json:"title"`
	// TrackCode string `json:"track_code"`
	// Type      string         `json:"type"`
}

// NewsfeedItemTopic struct.
//
// Comments BaseCommentsInfo `json:"comments"`
// Likes BaseLikesInfo `json:"likes"`
// Text  string        `json:"text"` // Post text.
type NewsfeedItemTopic struct{}

// NewsfeedItemVideo struct.
type NewsfeedItemVideo struct {
	Video NewsfeedItemVideoVideo `json:"video"`
}

// NewsfeedItemVideoVideo struct.
type NewsfeedItemVideoVideo struct {
	Count int          `json:"count"` // Tags number
	Items []VideoVideo `json:"items"`
}

// NewsfeedItemWallpost struct.
type NewsfeedItemWallpost struct {
	Activity       NewsfeedEventActivity    `json:"activity"`
	Attachments    []WallWallpostAttachment `json:"attachments"`
	Comments       BaseCommentsInfo         `json:"comments"`
	FromID         int                      `json:"from_id"`
	CopyHistory    []WallWallpost           `json:"copy_history"`
	Geo            BaseGeo                  `json:"geo"`
	Likes          BaseLikesInfo            `json:"likes"`
	PostSource     WallPostSource           `json:"post_source"`
	PostType       string                   `json:"post_type"`
	Reposts        BaseRepostsInfo          `json:"reposts"`
	MarkedAsAds    int                      `json:"marked_as_ads,omitempty"`
	Views          interface{}              `json:"views,omitempty"` // BUG: Views int or wallViews
	IsFavorite     BaseBoolInt              `json:"is_favorite,omitempty"`
	CanDelete      BaseBoolInt              `json:"can_delete"`
	CanArchive     BaseBoolInt              `json:"can_archive"`
	IsArchived     BaseBoolInt              `json:"is_archived"`
	SignerID       int                      `json:"signer_id,omitempty"`
	Text           string                   `json:"text"` // Post text
	Copyright      WallPostCopyright        `json:"copyright"`
	CategoryAction NewsfeedCategoryAction   `json:"category_action"`
}

// NewsfeedCategoryAction struct.
type NewsfeedCategoryAction struct {
	Action struct {
		Target string `json:"target"`
		Type   string `json:"type"`
		URL    string `json:"url"`
	} `json:"action"`
	Name string `json:"name"`
}

// NewsfeedList struct.
type NewsfeedList struct {
	ID    int    `json:"id"`    // List ID
	Title string `json:"title"` // List title
}

// NewsfeedItemMarket struct.
type NewsfeedItemMarket struct {
	MarketMarketItem
}

// NewsfeedNewsfeedItem struct.
type NewsfeedNewsfeedItem struct {
	Type     string `json:"type"`
	SourceID int    `json:"source_id"`
	Date     int    `json:"date"`
	TopicID  int    `json:"topic_id"`

	PostID int `json:"post_id,omitempty"`

	NewsfeedItemWallpost
	NewsfeedItemPhoto
	NewsfeedItemPhotoTag
	NewsfeedItemFriend
	NewsfeedItemNote
	NewsfeedItemAudio
	NewsfeedItemTopic
	NewsfeedItemVideo
	NewsfeedItemDigest
	NewsfeedItemStoriesBlock
	NewsfeedItemMarket

	CreatedBy        int         `json:"created_by,omitempty"`
	CanEdit          BaseBoolInt `json:"can_edit,omitempty"`
	CanDelete        BaseBoolInt `json:"can_delete,omitempty"`
	CanDoubtCategory BaseBoolInt `json:"can_doubt_category"`
	CanSetCategory   BaseBoolInt `json:"can_set_category"`
	// TODO: Need more fields
}

// NewsfeedNewsfeedNote struct.
type NewsfeedNewsfeedNote struct {
	Comments int    `json:"comments"` // Comments Number
	ID       int    `json:"id"`       // Note ID
	OwnerID  int    `json:"owner_id"` // integer
	Title    string `json:"title"`    // Note title
}
