package object // import "github.com/SevereCloud/vksdk/v2/object"

import (
	"encoding/json"
)

// StoriesViewer struct.
type StoriesViewer struct {
	IsLiked bool `json:"is_liked"`
	UserID  int  `json:"user_id"`

	// For extended
	User struct {
		Type            string `json:"type"`
		ID              int    `json:"id"`
		FirstName       string `json:"first_name"`
		LastName        string `json:"last_name"`
		IsClosed        bool   `json:"is_closed"`
		CanAccessClosed bool   `json:"can_access_closed"`
	} `json:"user,omitempty"`
}

// StoriesNarrativeInfo type.
type StoriesNarrativeInfo struct {
	Author string `json:"author"`
	Title  string `json:"title"`
	Views  int    `json:"views"`
}

// StoriesPromoData additional data for promo stories.
//
// TODO: v3 rename StoriesPromoBlock.
type StoriesPromoData struct {
	Name        string      `json:"name"`
	Photo50     string      `json:"photo_50"`
	Photo100    string      `json:"photo_100"`
	NotAnimated BaseBoolInt `json:"not_animated"`
}

// StoriesStoryLink struct.
type StoriesStoryLink struct {
	Text string `json:"text"` // Link text
	URL  string `json:"url"`  // Link URL
}

// StoriesReplies struct.
type StoriesReplies struct {
	Count int `json:"count"` // Replies number.
	New   int `json:"new"`   // New replies number.
}

// StoriesQuestions struct.
type StoriesQuestions struct {
	Count int `json:"count"` // Replies number.
	New   int `json:"new"`   // New replies number.
}

// StoriesStoryStats struct.
type StoriesStoryStats struct {
	Answer      StoriesStoryStatsStat `json:"answer"`
	Bans        StoriesStoryStatsStat `json:"bans"`
	OpenLink    StoriesStoryStatsStat `json:"open_link"`
	Replies     StoriesStoryStatsStat `json:"replies"`
	Shares      StoriesStoryStatsStat `json:"shares"`
	Subscribers StoriesStoryStatsStat `json:"subscribers"`
	Views       StoriesStoryStatsStat `json:"views"`
	Likes       StoriesStoryStatsStat `json:"likes"`
}

// StoriesStoryStatsStat struct.
type StoriesStoryStatsStat struct {
	Count int    `json:"count"` // Stat value
	State string `json:"state"`
}

// StoriesStoryType story type.
type StoriesStoryType string

// Possible values.
const (
	StoriesStoryPhoto          StoriesStoryType = "photo"
	StoriesStoryVideo          StoriesStoryType = "video"
	StoriesStoryLiveActive     StoriesStoryType = "live_active"
	StoriesStoryLiveFinished   StoriesStoryType = "live_finished"
	StoriesStoryBirthdayInvite StoriesStoryType = "birthday_invite"
)

// StoriesStory struct.
type StoriesStory struct {
	AccessKey string      `json:"access_key"` // Access key for private object.
	ExpiresAt int         `json:"expires_at"` // Story expiration time. Unixtime.
	CanHide   BaseBoolInt `json:"can_hide"`
	// Information whether story has question sticker and current user can send question to the author
	CanAsk BaseBoolInt `json:"can_ask"`
	// Information whether story has question sticker and current user can send anonymous question to the author
	CanAskAnonymous BaseBoolInt `json:"can_ask_anonymous"`

	// Information whether current user can comment the story (0 - no, 1 - yes).
	CanComment BaseBoolInt `json:"can_comment"`

	// Information whether current user can reply to the story
	// (0 - no, 1 - yes).
	CanReply BaseBoolInt `json:"can_reply"`

	// Information whether current user can see the story (0 - no, 1 - yes).
	CanSee BaseBoolInt `json:"can_see"`

	// Information whether current user can share the story (0 - no, 1 - yes).
	CanShare BaseBoolInt `json:"can_share"`

	// Information whether the story is deleted (false - no, true - yes).
	IsDeleted BaseBoolInt `json:"is_deleted"`

	// Information whether the story is expired (false - no, true - yes).
	IsExpired BaseBoolInt `json:"is_expired"`

	// Is video without sound
	NoSound BaseBoolInt `json:"no_sound"`

	// Does author have stories privacy restrictions
	IsRestricted BaseBoolInt `json:"is_restricted"`

	CanUseInNarrative BaseBoolInt `json:"can_use_in_narrative"`

	// Information whether current user has seen the story or not
	// (0 - no, 1 - yes).
	Seen                 BaseBoolInt              `json:"seen"`
	IsOwnerPinned        BaseBoolInt              `json:"is_owner_pinned"`
	IsOneTime            BaseBoolInt              `json:"is_one_time"`
	IsAdvice             BaseBoolInt              `json:"is_advice,omitempty"`
	NeedMute             BaseBoolInt              `json:"need_mute"`
	MuteReply            BaseBoolInt              `json:"mute_reply"`
	CanLike              BaseBoolInt              `json:"can_like"`
	Date                 int                      `json:"date"` // Date when story has been added in Unixtime.
	ID                   int                      `json:"id"`   // Story ID.
	Link                 StoriesStoryLink         `json:"link"`
	OwnerID              int                      `json:"owner_id"` // Story owner's ID.
	ParentStory          *StoriesStory            `json:"parent_story"`
	ParentStoryAccessKey string                   `json:"parent_story_access_key"` // Access key for private object.
	ParentStoryID        int                      `json:"parent_story_id"`         // Parent story ID.
	ParentStoryOwnerID   int                      `json:"parent_story_owner_id"`   // Parent story owner's ID.
	Photo                PhotosPhoto              `json:"photo"`
	Replies              StoriesReplies           `json:"replies"` // Replies to current story.
	Type                 string                   `json:"type"`
	Video                VideoVideo               `json:"video"`
	Views                int                      `json:"views"` // Views number.
	ClickableStickers    StoriesClickableStickers `json:"clickable_stickers"`
	TrackCode            string                   `json:"track_code"`
	LikesCount           int                      `json:"likes_count"`
	NarrativeID          int                      `json:"narrative_id"`
	NarrativeOwnerID     int                      `json:"narrative_owner_id"`
	NarrativeInfo        StoriesNarrativeInfo     `json:"narrative_info"`
	NarrativesCount      int                      `json:"narratives_count"`
	FirstNarrativeTitle  string                   `json:"first_narrative_title"`
	Questions            StoriesQuestions         `json:"questions"`
	ReactionSetID        string                   `json:"reaction_set_id"`
}

// StoriesFeedItemType type.
type StoriesFeedItemType string

// Possible values.
const (
	StoriesFeedItemStories   StoriesFeedItemType = "stories"
	StoriesFeedItemCommunity StoriesFeedItemType = "community_grouped_stories"
	StoriesFeedItemApp       StoriesFeedItemType = "app_grouped_stories"
)

// StoriesFeedItem struct.
type StoriesFeedItem struct {
	Type           StoriesFeedItemType `json:"type"`
	ID             string              `json:"id"`
	Stories        []StoriesStory      `json:"stories"`
	Grouped        StoriesFeedItemType `json:"grouped"`
	App            AppsApp             `json:"app"`
	BirthdayUserID int                 `json:"birthday_user_id"`
	TrackCode      string              `json:"track_code"`
	HasUnseen      BaseBoolInt         `json:"has_unseen"`
	Name           string              `json:"name"`
	PromoData      StoriesPromoData    `json:"promo_data"`
}

// StoriesClickableStickers struct.
//
// The field clickable_stickers is available in the history object.
// The sticker object is pasted by the developer on the client himself, only
// coordinates are transmitted to the server.
//
// https://vk.com/dev/objects/clickable_stickers
type StoriesClickableStickers struct {
	OriginalWidth     int                       `json:"original_width"`
	OriginalHeight    int                       `json:"original_height"`
	ClickableStickers []StoriesClickableSticker `json:"clickable_stickers"`
}

// NewClickableStickers return new StoriesClickableStickers.
//
// Requires the width and height of the original photo or video.
func NewClickableStickers(width, height int) *StoriesClickableStickers {
	return &StoriesClickableStickers{
		OriginalWidth:     width,
		OriginalHeight:    height,
		ClickableStickers: []StoriesClickableSticker{},
	}
}

// AddMention add mention sticker.
//
// Mention should be in the format of a VK mentioning, for example: [id1|name] or [club1|name].
func (cs *StoriesClickableStickers) AddMention(mention string, area []StoriesClickablePoint) *StoriesClickableStickers {
	cs.ClickableStickers = append(cs.ClickableStickers, StoriesClickableSticker{
		Type:          ClickableStickerMention,
		ClickableArea: area,
		Mention:       mention,
	})

	return cs
}

// AddHashtag add hashtag sticker.
//
// Hashtag must necessarily begin with the symbol #.
func (cs *StoriesClickableStickers) AddHashtag(hashtag string, area []StoriesClickablePoint) *StoriesClickableStickers {
	cs.ClickableStickers = append(cs.ClickableStickers, StoriesClickableSticker{
		Type:          ClickableStickerHashtag,
		ClickableArea: area,
		Hashtag:       hashtag,
	})

	return cs
}

// TODO: Add more clickable stickers func

// ToJSON returns the JSON encoding of StoriesClickableStickers.
func (cs StoriesClickableStickers) ToJSON() string {
	b, _ := json.Marshal(cs)
	return string(b)
}

// StoriesClickableSticker struct.
type StoriesClickableSticker struct { //nolint: maligned
	ID            int                     `json:"id"`
	Type          string                  `json:"type"`
	ClickableArea []StoriesClickablePoint `json:"clickable_area"`
	Style         string                  `json:"style,omitempty"`

	// type=post
	PostOwnerID int `json:"post_owner_id,omitempty"`
	PostID      int `json:"post_id,omitempty"`

	// type=sticker
	StickerID     int `json:"sticker_id,omitempty"`
	StickerPackID int `json:"sticker_pack_id,omitempty"`

	// type=place or geo
	PlaceID int `json:"place_id,omitempty"`
	// Title
	CategoryID int `json:"category_id,omitempty"`

	// type=question
	Question               string      `json:"question,omitempty"`
	QuestionButton         string      `json:"question_button,omitempty"`
	QuestionDefaultPrivate BaseBoolInt `json:"question_default_private,omitempty"`
	Color                  string      `json:"color,omitempty"`

	// type=mention
	Mention string `json:"mention,omitempty"`

	// type=hashtag
	Hashtag string `json:"hashtag,omitempty"`

	// type=link
	LinkObject     BaseLink `json:"link_object,omitempty"`
	TooltipText    string   `json:"tooltip_text,omitempty"`
	TooltipTextKey string   `json:"tooltip_text_key,omitempty"`

	// type=time
	TimestampMs int64  `json:"timestamp_ms,omitempty"`
	Date        string `json:"date,omitempty"`
	Title       string `json:"title,omitempty"`

	// type=market_item
	Subtype string `json:"subtype,omitempty"`
	// LinkObject BaseLink         `json:"link_object,omitempty"` // subtype=aliexpress_product
	MarketItem MarketMarketItem `json:"market_item,omitempty"` // subtype=market_item

	// type=story_reply
	OwnerID int `json:"owner_id,omitempty"`
	StoryID int `json:"story_id,omitempty"`

	// type=owner
	// OwnerID int `json:"owner_id,omitempty"`

	// type=poll
	Poll PollsPoll `json:"poll,omitempty"`

	// type=music
	Audio          AudioAudio `json:"audio,omitempty"`
	AudioStartTime int        `json:"audio_start_time,omitempty"`

	// type=app
	App                      AppsApp     `json:"app,omitempty"`
	AppContext               string      `json:"app_context,omitempty"`
	HasNewInteractions       BaseBoolInt `json:"has_new_interactions,omitempty"`
	IsBroadcastNotifyAllowed BaseBoolInt `json:"is_broadcast_notify_allowed,omitempty"`

	// type=emoji
	Emoji string `json:"emoji,omitempty"`

	// type=text
	Text            string `json:"text,omitempty"`
	BackgroundStyle string `json:"background_style,omitempty"`
	Alignment       string `json:"alignment,omitempty"`
	SelectionColor  string `json:"selection_color,omitempty"`
}

// TODO: сделать несколько структур для кликабельного стикера

// Type of clickable sticker.
const (
	ClickableStickerPost       = "post"
	ClickableStickerSticker    = "sticker"
	ClickableStickerPlace      = "place"
	ClickableStickerQuestion   = "question"
	ClickableStickerMention    = "mention"
	ClickableStickerHashtag    = "hashtag"
	ClickableStickerMarketItem = "market_item"
	ClickableStickerLink       = "link"
	ClickableStickerStoryReply = "story_reply"
	ClickableStickerOwner      = "owner"
	ClickableStickerPoll       = "poll"
	ClickableStickerMusic      = "music"
	ClickableStickerApp        = "app"
	ClickableStickerTime       = "time"
	ClickableStickerEmoji      = "emoji"
	ClickableStickerGeo        = "geo"
	ClickableStickerText       = "text"
)

// Subtype of clickable sticker.
const (
	ClickableStickerSubtypeMarketItem        = "market_item"
	ClickableStickerSubtypeAliexpressProduct = "aliexpress_product"
)

// Clickable sticker style.
const (
	ClickableStickerTransparent   = "transparent"
	ClickableStickerBlueGradient  = "blue_gradient"
	ClickableStickerRedGradient   = "red_gradient"
	ClickableStickerUnderline     = "underline"
	ClickableStickerBlue          = "blue"
	ClickableStickerGreen         = "green"
	ClickableStickerWhite         = "white"
	ClickableStickerQuestionReply = "question_reply"
	ClickableStickerLight         = "light"
	ClickableStickerImpressive    = "impressive"
)

// StoriesClickablePoint struct.
type StoriesClickablePoint struct {
	X int `json:"x"`
	Y int `json:"y"`
}
