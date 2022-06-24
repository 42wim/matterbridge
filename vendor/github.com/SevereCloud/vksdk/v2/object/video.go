package object // import "github.com/SevereCloud/vksdk/v2/object"

import (
	"fmt"
)

// VideoVideo struct.
type VideoVideo struct {
	// Video access key.
	AccessKey string `json:"access_key"`

	// Date when the video has been added in Unixtime.
	AddingDate int `json:"adding_date"`

	// Date when the video has been released in Unixtime.
	ReleaseDate int `json:"release_date"`

	// Information whether current user can add the video.
	CanAdd BaseBoolInt `json:"can_add"`

	// Information whether current user can add the video to faves.
	CanAddToFaves BaseBoolInt `json:"can_add_to_faves"`

	// Information whether current user can comment the video.
	CanComment BaseBoolInt `json:"can_comment"`

	// Information whether current user can edit the video.
	CanEdit BaseBoolInt `json:"can_edit"`

	// Information whether current user can like the video.
	CanLike BaseBoolInt `json:"can_like"`

	// Information whether current user can download the video.
	CanDownload int `json:"can_download"`

	// Information whether current user can repost this video.
	CanRepost         BaseBoolInt       `json:"can_repost"`
	CanSubscribe      BaseBoolInt       `json:"can_subscribe"`
	CanAttachLink     BaseBoolInt       `json:"can_attach_link"`
	IsFavorite        BaseBoolInt       `json:"is_favorite"`
	IsPrivate         BaseBoolInt       `json:"is_private"`
	IsExplicit        BaseBoolInt       `json:"is_explicit"`
	IsSubscribed      BaseBoolInt       `json:"is_subscribed"`
	Added             BaseBoolInt       `json:"added"`
	Repeat            BaseBoolInt       `json:"repeat"` // Information whether the video is repeated
	ContentRestricted int               `json:"content_restricted"`
	Live              BaseBoolInt       `json:"live"` // Returns if the video is a live stream
	Upcoming          BaseBoolInt       `json:"upcoming"`
	Comments          int               `json:"comments"`    // Number of comments
	Date              int               `json:"date"`        // Date when video has been uploaded in Unixtime
	Description       string            `json:"description"` // Video description
	Duration          int               `json:"duration"`    // Video duration in seconds
	Files             VideoVideoFiles   `json:"files"`
	Trailer           VideoVideoFiles   `json:"trailer,omitempty"`
	FirstFrame        []VideoVideoImage `json:"first_frame"`
	Image             []VideoVideoImage `json:"image"`
	Height            int               `json:"height"`   // Video height
	ID                int               `json:"id"`       // Video ID
	OwnerID           int               `json:"owner_id"` // Video owner ID
	UserID            int               `json:"user_id"`
	Photo130          string            `json:"photo_130"`  // URL of the preview image with 130 px in width
	Photo320          string            `json:"photo_320"`  // URL of the preview image with 320 px in width
	Photo640          string            `json:"photo_640"`  // URL of the preview image with 640 px in width
	Photo800          string            `json:"photo_800"`  // URL of the preview image with 800 px in width
	Photo1280         string            `json:"photo_1280"` // URL of the preview image with 1280 px in width

	// URL of the page with a player that can be used to play the video in the browser.
	Player                   string               `json:"player"`
	Processing               int                  `json:"processing"` // Returns if the video is processing
	Title                    string               `json:"title"`      // Video title
	Subtitle                 string               `json:"subtitle"`   // Video subtitle
	Type                     string               `json:"type"`
	Views                    int                  `json:"views"` // Number of views
	Width                    int                  `json:"width"` // Video width
	Platform                 string               `json:"platform"`
	LocalViews               int                  `json:"local_views"`
	Likes                    BaseLikesInfo        `json:"likes"`   // Count of likes
	Reposts                  BaseRepostsInfo      `json:"reposts"` // Count of views
	TrackCode                string               `json:"track_code"`
	PrivacyView              Privacy              `json:"privacy_view"`
	PrivacyComment           Privacy              `json:"privacy_comment"`
	ActionButton             VideoActionButton    `json:"action_button"`
	Restriction              VideoRestriction     `json:"restriction"`
	ContentRestrictedMessage string               `json:"content_restricted_message"`
	MainArtists              []AudioAudioArtist   `json:"main_artists"`
	FeaturedArtists          []AudioAudioArtist   `json:"featured_artists"`
	Genres                   []BaseObjectWithName `json:"genres"`
	OvID                     string               `json:"ov_id,omitempty"`
}

// ToAttachment return attachment format.
func (video VideoVideo) ToAttachment() string {
	return fmt.Sprintf("video%d_%d", video.OwnerID, video.ID)
}

// VideoRestriction struct.
type VideoRestriction struct {
	Title          string      `json:"title"`
	Text           string      `json:"text"`
	AlwaysShown    BaseBoolInt `json:"always_shown"`
	Blur           BaseBoolInt `json:"blur"`
	CanPlay        BaseBoolInt `json:"can_play"`
	CanPreview     BaseBoolInt `json:"can_preview"`
	CardIcon       []BaseImage `json:"card_icon"`
	ListIcon       []BaseImage `json:"list_icon"`
	DisclaimerType int         `json:"disclaimer_type"`
}

// VideoActionButton struct.
type VideoActionButton struct {
	ID      string       `json:"id"`
	Type    string       `json:"type"`
	URL     string       `json:"url"`
	Snippet VideoSnippet `json:"snippet"`
}

// VideoSnippet struct.
type VideoSnippet struct {
	Description string      `json:"description"`
	OpenTitle   string      `json:"open_title"`
	Title       string      `json:"title"`
	TypeName    string      `json:"type_name"`
	Date        int         `json:"date"`
	Image       []BaseImage `json:"image"`
}

// VideoVideoFiles struct.
type VideoVideoFiles struct {
	External     string `json:"external,omitempty"` // URL of the external player
	Mp4_1080     string `json:"mp4_1080,omitempty"` // URL of the mpeg4 file with 1080p quality
	Mp4_1440     string `json:"mp4_1440,omitempty"` // URL of the mpeg4 file with 2k quality
	Mp4_2160     string `json:"mp4_2160,omitempty"` // URL of the mpeg4 file with 4k quality
	Mp4_240      string `json:"mp4_240,omitempty"`  // URL of the mpeg4 file with 240p quality
	Mp4_360      string `json:"mp4_360,omitempty"`  // URL of the mpeg4 file with 360p quality
	Mp4_480      string `json:"mp4_480,omitempty"`  // URL of the mpeg4 file with 480p quality
	Mp4_720      string `json:"mp4_720,omitempty"`  // URL of the mpeg4 file with 720p quality
	Live         string `json:"live,omitempty"`
	HLS          string `json:"hls,omitempty"`
	DashUni      string `json:"dash_uni,omitempty"`
	DashSep      string `json:"dash_sep,omitempty"`
	DashWebm     string `json:"dash_webm,omitempty"`
	FailoverHost string `json:"failover_host,omitempty"`
}

// VideoCatBlock struct.
type VideoCatBlock struct {
	CanHide BaseBoolInt       `json:"can_hide"`
	ID      int               `json:"id"`
	Items   []VideoCatElement `json:"items"`
	Name    string            `json:"name"`
	Next    string            `json:"next"`
	Type    string            `json:"type"`
	View    string            `json:"view"`
}

// VideoCatElement struct.
type VideoCatElement struct {
	CanAdd      BaseBoolInt `json:"can_add"`
	CanEdit     BaseBoolInt `json:"can_edit"`
	IsPrivate   BaseBoolInt `json:"is_private"`
	Comments    int         `json:"comments"`
	Count       int         `json:"count"`
	Date        int         `json:"date"`
	Description string      `json:"description"`
	Duration    int         `json:"duration"`
	ID          int         `json:"id"`
	OwnerID     int         `json:"owner_id"`
	Photo130    string      `json:"photo_130"`
	Photo160    string      `json:"photo_160"`
	Photo320    string      `json:"photo_320"`
	Photo640    string      `json:"photo_640"`
	Photo800    string      `json:"photo_800"`
	Title       string      `json:"title"`
	Type        string      `json:"type"`
	UpdatedTime int         `json:"updated_time"`
	Views       int         `json:"views"`
}

// VideoSaveResult struct.
type VideoSaveResult struct {
	Description string `json:"description"` // Video description
	OwnerID     int    `json:"owner_id"`    // Video owner ID
	Title       string `json:"title"`       // Video title
	UploadURL   string `json:"upload_url"`  // URL for the video uploading
	VideoID     int    `json:"video_id"`    // Video ID
	AccessKey   string `json:"access_key"`  // Video access key
}

// VideoUploadResponse struct.
type VideoUploadResponse struct {
	Size    int `json:"size"`
	VideoID int `json:"video_id"`
}

// VideoVideoAlbum struct.
type VideoVideoAlbum struct {
	ID      int    `json:"id"`
	OwnerID int    `json:"owner_id"`
	Title   string `json:"title"`
}

// VideoVideoAlbumFull struct.
type VideoVideoAlbumFull struct {
	Count       int               `json:"count"`        // Total number of videos in album
	ID          int               `json:"id"`           // Album ID
	Image       []VideoVideoImage `json:"image"`        // Album cover image in different sizes
	IsSystem    BaseBoolInt       `json:"is_system"`    // Information whether album is system
	OwnerID     int               `json:"owner_id"`     // Album owner's ID
	Photo160    string            `json:"photo_160"`    // URL of the preview image with 160px in width
	Photo320    string            `json:"photo_320"`    // URL of the preview image with 320px in width
	Title       string            `json:"title"`        // Album title
	UpdatedTime int               `json:"updated_time"` // Date when the album has been updated last time in Unixtime
	ImageBlur   int               `json:"image_blur"`
	Privacy     Privacy           `json:"privacy"`
}

// VideoVideoFull struct.
type VideoVideoFull struct {
	AccessKey     string          `json:"access_key"`  // Video access key
	AddingDate    int             `json:"adding_date"` // Date when the video has been added in Unixtime
	IsFavorite    BaseBoolInt     `json:"is_favorite"`
	CanAdd        BaseBoolInt     `json:"can_add"`     // Information whether current user can add the video
	CanComment    BaseBoolInt     `json:"can_comment"` // Information whether current user can comment the video
	CanEdit       BaseBoolInt     `json:"can_edit"`    // Information whether current user can edit the video
	CanRepost     BaseBoolInt     `json:"can_repost"`  // Information whether current user can comment the video
	CanLike       BaseBoolInt     `json:"can_like"`
	CanAddToFaves BaseBoolInt     `json:"can_add_to_faves"`
	Repeat        BaseBoolInt     `json:"repeat"`      // Information whether the video is repeated
	Comments      int             `json:"comments"`    // Number of comments
	Date          int             `json:"date"`        // Date when video has been uploaded in Unixtime
	Description   string          `json:"description"` // Video description
	Duration      int             `json:"duration"`    // Video duration in seconds
	Files         VideoVideoFiles `json:"files"`
	Trailer       VideoVideoFiles `json:"trailer"`
	ID            int             `json:"id"` // Video ID
	Likes         BaseLikes       `json:"likes"`
	Live          int             `json:"live"`     // Returns if the video is live translation
	OwnerID       int             `json:"owner_id"` // Video owner ID

	// URL of the page with a player that can be used to play the video in the browser.
	Player     string            `json:"player"`
	Processing int               `json:"processing"` // Returns if the video is processing
	Title      string            `json:"title"`      // Video title
	Views      int               `json:"views"`      // Number of views
	Width      int               `json:"width"`
	Height     int               `json:"height"`
	Image      []VideoVideoImage `json:"image"`
	FirstFrame []VideoVideoImage `json:"first_frame"`
	Added      int               `json:"added"`
	Type       string            `json:"type"`
	Reposts    BaseRepostsInfo   `json:"reposts"`
}

// ToAttachment return attachment format.
func (video VideoVideoFull) ToAttachment() string {
	return fmt.Sprintf("video%d_%d", video.OwnerID, video.ID)
}

// VideoVideoTag struct.
type VideoVideoTag struct {
	Date       int         `json:"date"`
	ID         int         `json:"id"`
	PlacerID   int         `json:"placer_id"`
	TaggedName string      `json:"tagged_name"`
	UserID     int         `json:"user_id"`
	Viewed     BaseBoolInt `json:"viewed"`
}

// VideoVideoTagInfo struct.
type VideoVideoTagInfo struct {
	AccessKey   string          `json:"access_key"`
	AddingDate  int             `json:"adding_date"`
	CanAdd      BaseBoolInt     `json:"can_add"`
	CanEdit     BaseBoolInt     `json:"can_edit"`
	Comments    int             `json:"comments"`
	Date        int             `json:"date"`
	Description string          `json:"description"`
	Duration    int             `json:"duration"`
	Files       VideoVideoFiles `json:"files"`
	ID          int             `json:"id"`
	Live        int             `json:"live"`
	OwnerID     int             `json:"owner_id"`
	Photo130    string          `json:"photo_130"`
	Photo320    string          `json:"photo_320"`
	Photo800    string          `json:"photo_800"`
	PlacerID    int             `json:"placer_id"`
	Player      string          `json:"player"`
	Processing  int             `json:"processing"`
	TagCreated  int             `json:"tag_created"`
	TagID       int             `json:"tag_id"`
	Title       string          `json:"title"`
	Views       int             `json:"views"`
}

// VideoVideoImage struct.
type VideoVideoImage struct {
	BaseImage
	WithPadding BaseBoolInt `json:"with_padding"`
}

// VideoLive struct.
type VideoLive struct {
	OwnerID     int             `json:"owner_id"`
	VideoID     int             `json:"video_id"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	AccessKey   string          `json:"access_key"`
	Stream      VideoLiveStream `json:"stream"`
}

// VideoLiveStream struct.
type VideoLiveStream struct {
	URL     string `json:"url"`
	Key     string `json:"key"`
	OKMPURL string `json:"okmp_url"`
}

// VideoLiveCategory struct.
type VideoLiveCategory struct {
	ID      int                 `json:"id"`
	Label   string              `json:"label"`
	Sublist []VideoLiveCategory `json:"sublist,omitempty"`
}
