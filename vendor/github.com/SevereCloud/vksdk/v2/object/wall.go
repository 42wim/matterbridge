package object // import "github.com/SevereCloud/vksdk/v2/object"

// WallAppPost struct.
type WallAppPost struct {
	ID       int    `json:"id"`        // Application ID
	Name     string `json:"name"`      // Application name
	Photo130 string `json:"photo_130"` // URL of the preview image with 130 px in width
	Photo604 string `json:"photo_604"` // URL of the preview image with 604 px in width
}

// WallAttachedNote struct.
type WallAttachedNote struct {
	Comments     int    `json:"comments"`      // Comments number
	Date         int    `json:"date"`          // Date when the note has been created in Unixtime
	ID           int    `json:"id"`            // Note ID
	OwnerID      int    `json:"owner_id"`      // Note owner's ID
	ReadComments int    `json:"read_comments"` // Read comments number
	Title        string `json:"title"`         // Note title
	ViewURL      string `json:"view_url"`      // URL of the page with note preview
}

// WallCommentAttachment struct.
type WallCommentAttachment struct {
	Audio             AudioAudio        `json:"audio"`
	Doc               DocsDoc           `json:"doc"`
	Link              BaseLink          `json:"link"`
	Market            MarketMarketItem  `json:"market"`
	MarketMarketAlbum MarketMarketAlbum `json:"market_market_album"`
	Note              WallAttachedNote  `json:"note"`
	Page              PagesWikipageFull `json:"page"`
	Photo             PhotosPhoto       `json:"photo"`
	Sticker           BaseSticker       `json:"sticker"`
	Type              string            `json:"type"`
	Video             VideoVideo        `json:"video"`
	Graffiti          WallGraffiti      `json:"graffiti"`
}

// WallGraffiti struct.
type WallGraffiti struct {
	ID        int    `json:"id"`        // Graffiti ID
	OwnerID   int    `json:"owner_id"`  // Graffiti owner's ID
	Photo200  string `json:"photo_200"` // URL of the preview image with 200 px in width
	Photo586  string `json:"photo_586"` // URL of the preview image with 586 px in width
	URL       string `json:"url"`
	Width     int    `json:"width"`
	Height    int    `json:"height"`
	AccessKey string `json:"access_key"`
}

// Type of post source.
const (
	WallPostSourceTypeVk     = "vk"
	WallPostSourceTypeWidget = "widget"
	WallPostSourceTypeAPI    = "api"
	WallPostSourceTypeRss    = "rss"
	WallPostSourceTypeSms    = "sms"
)

// WallPostSource struct.
type WallPostSource struct {
	Link     BaseLink `json:"link"`
	Data     string   `json:"data"`     // Additional data
	Platform string   `json:"platform"` // Platform name
	Type     string   `json:"type"`
	URL      string   `json:"url"` // URL to an external site used to publish the post
}

// WallPostedPhoto struct.
type WallPostedPhoto struct {
	ID       int    `json:"id"`        // Photo ID
	OwnerID  int    `json:"owner_id"`  // Photo owner's ID
	Photo130 string `json:"photo_130"` // URL of the preview image with 130 px in width
	Photo604 string `json:"photo_604"` // URL of the preview image with 604 px in width
}

// WallViews struct.
type WallViews struct {
	Count int `json:"count"` // Count
}

// WallWallCommentThread struct.
type WallWallCommentThread struct {
	Count           int               `json:"count"` // Comments number
	Items           []WallWallComment `json:"items"`
	CanPost         BaseBoolInt       `json:"can_post"`        // Information whether current user can comment the post
	GroupsCanPost   BaseBoolInt       `json:"groups_can_post"` // Information whether groups can comment the post
	ShowReplyButton BaseBoolInt       `json:"show_reply_button"`
}

// WallWallComment struct.
type WallWallComment struct {
	Attachments    []WallCommentAttachment `json:"attachments"`
	Date           int                     `json:"date"` // Date when the comment has been added in Unixtime
	Deleted        BaseBoolInt             `json:"deleted"`
	FromID         int                     `json:"from_id"` // Author ID
	ID             int                     `json:"id"`      // Comment ID
	Likes          BaseLikesInfo           `json:"likes"`
	RealOffset     int                     `json:"real_offset"`      // Real position of the comment
	ReplyToComment int                     `json:"reply_to_comment"` // Replied comment ID
	ReplyToUser    int                     `json:"reply_to_user"`    // Replied user ID
	Text           string                  `json:"text"`             // Comment text
	PostID         int                     `json:"post_id"`
	PostOwnerID    int                     `json:"post_owner_id"`
	PhotoID        int                     `json:"photo_id"`
	PhotoOwnerID   int                     `json:"photo_owner_id"`
	VideoID        int                     `json:"video_id"`
	VideoOwnerID   int                     `json:"video_owner_id"`
	ItemID         int                     `json:"item_id"`
	MarketOwnerID  int                     `json:"market_owner_id"`
	ParentsStack   []int                   `json:"parents_stack"`
	OwnerID        int                     `json:"owner_id"`
	Thread         WallWallCommentThread   `json:"thread"`
	Donut          WallWallCommentDonut    `json:"donut"`
}

// WallWallCommentDonut info about VK Donut.
type WallWallCommentDonut struct {
	IsDonut     BaseBoolInt `json:"is_donut"`
	Placeholder string      `json:"placeholder"`
}

// WallPost type.
const (
	WallPostTypePost     = "post"
	WallPostTypeCopy     = "copy"
	WallPostTypeReply    = "reply"
	WallPostTypePostpone = "postpone"
	WallPostTypeSuggest  = "suggest"
)

// WallWallpost struct.
type WallWallpost struct {
	AccessKey      string                   `json:"access_key"` // Access key to private object
	ID             int                      `json:"id"`         // Post ID
	OwnerID        int                      `json:"owner_id"`   // Wall owner's ID
	FromID         int                      `json:"from_id"`    // Post author ID
	CreatedBy      int                      `json:"created_by"`
	Date           int                      `json:"date"` // Date of publishing in Unixtime
	Text           string                   `json:"text"` // Post text
	ReplyOwnerID   int                      `json:"reply_owner_id"`
	ReplyPostID    int                      `json:"reply_post_id"`
	FriendsOnly    int                      `json:"friends_only"`
	Comments       BaseCommentsInfo         `json:"comments"`
	Likes          BaseLikesInfo            `json:"likes"`   // Count of likes
	Reposts        BaseRepostsInfo          `json:"reposts"` // Count of reposts
	Views          WallViews                `json:"views"`   // Count of views
	PostType       string                   `json:"post_type"`
	PostSource     WallPostSource           `json:"post_source"`
	Attachments    []WallWallpostAttachment `json:"attachments"`
	Geo            BaseGeo                  `json:"geo"`
	SignerID       int                      `json:"signer_id"` // Post signer ID
	CopyHistory    []WallWallpost           `json:"copy_history"`
	CanPin         BaseBoolInt              `json:"can_pin"`
	CanDelete      BaseBoolInt              `json:"can_delete"`
	CanEdit        BaseBoolInt              `json:"can_edit"`
	IsPinned       BaseBoolInt              `json:"is_pinned"`
	IsFavorite     BaseBoolInt              `json:"is_favorite"` // Information whether the post in favorites list
	IsArchived     BaseBoolInt              `json:"is_archived"` // Is post archived, only for post owners
	IsDeleted      BaseBoolInt              `json:"is_deleted"`
	MarkedAsAds    BaseBoolInt              `json:"marked_as_ads"`
	Edited         int                      `json:"edited"` // Date of editing in Unixtime
	Copyright      WallPostCopyright        `json:"copyright"`
	PostID         int                      `json:"post_id"`
	PostponedID    int                      `json:"postponed_id"` // ID from scheduled posts
	ParentsStack   []int                    `json:"parents_stack"`
	Donut          WallWallpostDonut        `json:"donut"`
	ShortTextRate  float64                  `json:"short_text_rate"`
	CarouselOffset int                      `json:"carousel_offset"`
	Header         WallWallpostHeader       `json:"header"`
	Hash           string                   `json:"hash"`
}

// Attachment type.
//
// TODO: check this.
const (
	AttachmentTypePhoto       = "photo"
	AttachmentTypePostedPhoto = "posted_photo"
	AttachmentTypeAudio       = "audio"
	AttachmentTypeVideo       = "video"
	AttachmentTypeDoc         = "doc"
	AttachmentTypeLink        = "link"
	AttachmentTypeGraffiti    = "graffiti"
	AttachmentTypeNote        = "note"
	AttachmentTypeApp         = "app"
	AttachmentTypePoll        = "poll"
	AttachmentTypePage        = "page"
	AttachmentTypeAlbum       = "album"
	AttachmentTypePhotosList  = "photos_list"
	AttachmentTypeMarketAlbum = "market_album"
	AttachmentTypeMarket      = "market"
	AttachmentTypeEvent       = "event"
	AttachmentTypeWall        = "wall"
	AttachmentTypeStory       = "story"
	AttachmentTypePodcast     = "podcast"
)

// WallWallpostAttachment struct.
type WallWallpostAttachment struct {
	AccessKey         string            `json:"access_key"` // Access key for the audio
	Album             PhotosPhotoAlbum  `json:"album"`
	App               WallAppPost       `json:"app"`
	Audio             AudioAudio        `json:"audio"`
	Doc               DocsDoc           `json:"doc"`
	Event             EventsEventAttach `json:"event"`
	Graffiti          WallGraffiti      `json:"graffiti"`
	Link              BaseLink          `json:"link"`
	Market            MarketMarketItem  `json:"market"`
	MarketMarketAlbum MarketMarketAlbum `json:"market_market_album"`
	Note              WallAttachedNote  `json:"note"`
	Page              PagesWikipageFull `json:"page"`
	Photo             PhotosPhoto       `json:"photo"`
	PhotosList        []string          `json:"photos_list"`
	Poll              PollsPoll         `json:"poll"`
	PostedPhoto       WallPostedPhoto   `json:"posted_photo"`
	Type              string            `json:"type"`
	Video             VideoVideo        `json:"video"`
	Podcast           PodcastsEpisode   `json:"podcast"`
}

// WallWallpostToID struct.
type WallWallpostToID struct {
	Attachments   []WallWallpostAttachment `json:"attachments"`
	Comments      BaseCommentsInfo         `json:"comments"`
	CopyOwnerID   int                      `json:"copy_owner_id"` // ID of the source post owner
	CopyPostID    int                      `json:"copy_post_id"`  // ID of the source post
	Date          int                      `json:"date"`          // Date of publishing in Unixtime
	FromID        int                      `json:"from_id"`       // Post author ID
	Geo           BaseGeo                  `json:"geo"`
	ID            int                      `json:"id"` // Post ID
	Likes         BaseLikesInfo            `json:"likes"`
	PostID        int                      `json:"post_id"` // wall post ID (if comment)
	PostSource    WallPostSource           `json:"post_source"`
	PostType      string                   `json:"post_type"`
	Reposts       BaseRepostsInfo          `json:"reposts"`
	SignerID      int                      `json:"signer_id"`   // Post signer ID
	Text          string                   `json:"text"`        // Post text
	ToID          int                      `json:"to_id"`       // Wall owner's ID
	IsFavorite    BaseBoolInt              `json:"is_favorite"` // Information whether the post in favorites list
	MarkedAsAds   BaseBoolInt              `json:"marked_as_ads"`
	ParentsStack  []int                    `json:"parents_stack"`
	Donut         WallWallpostDonut        `json:"donut"`
	ShortTextRate float64                  `json:"short_text_rate"`
	Views         WallViews                `json:"views"` // Count of views
	Header        WallWallpostHeader       `json:"header"`
}

// WallWallpostDonut info about VK Donut.
type WallWallpostDonut struct {
	IsDonut            BaseBoolInt          `json:"is_donut"`
	CanPublishFreeCopy BaseBoolInt          `json:"can_publish_free_copy"`
	PaidDuration       int                  `json:"paid_duration"`
	EditMode           string               `json:"edit_mode"`
	Durations          []BaseObjectWithName `json:"durations"`
}

// WallPostCopyright information about the source of the post.
type WallPostCopyright struct {
	ID   int    `json:"id,omitempty"`
	Link string `json:"link"`
	Type string `json:"type"`
	Name string `json:"name"`
}

// WallWallpostHeader struct.
type WallWallpostHeader struct {
	Type              string                              `json:"type"`
	CustomDescription WallWallpostHeaderCustomDescription `json:"custom_description"`
}

// WallWallpostHeaderCustomDescription struct.
type WallWallpostHeaderCustomDescription struct {
	SourceID int `json:"source_id"`
	Date     int `json:"date"`
}
