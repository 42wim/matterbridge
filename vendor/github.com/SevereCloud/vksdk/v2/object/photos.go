package object // import "github.com/SevereCloud/vksdk/v2/object"

import (
	"fmt"
)

// PhotosPhoto struct.
type PhotosPhoto struct {
	AccessKey          string             `json:"access_key"` // Access key for the photo
	AlbumID            int                `json:"album_id"`   // Album ID
	Date               int                `json:"date"`       // Date when uploaded
	Height             int                `json:"height"`     // Original photo height
	ID                 int                `json:"id"`         // Photo ID
	Images             []PhotosImage      `json:"images"`
	Lat                float64            `json:"lat"`      // Latitude
	Long               float64            `json:"long"`     // Longitude
	OwnerID            int                `json:"owner_id"` // Photo owner's ID
	PostID             int                `json:"post_id"`  // Post ID
	Text               string             `json:"text"`     // Photo caption
	UserID             int                `json:"user_id"`  // ID of the user who have uploaded the photo
	Width              int                `json:"width"`    // Original photo width
	CanUpload          BaseBoolInt        `json:"can_upload"`
	CommentsDisabled   BaseBoolInt        `json:"comments_disabled"`
	ThumbIsLast        BaseBoolInt        `json:"thumb_is_last"`
	UploadByAdminsOnly BaseBoolInt        `json:"upload_by_admins_only"`
	HasTags            BaseBoolInt        `json:"has_tags"`
	Created            int                `json:"created"`
	Description        string             `json:"description"`
	PrivacyComment     []string           `json:"privacy_comment"`
	PrivacyView        []string           `json:"privacy_view"`
	Size               int                `json:"size"`
	Sizes              []PhotosPhotoSizes `json:"sizes"`
	ThumbID            int                `json:"thumb_id"`
	ThumbSrc           string             `json:"thumb_src"`
	Title              string             `json:"title"`
	Updated            int                `json:"updated"`
	Color              string             `json:"color"`
}

// ToAttachment return attachment format.
func (photo PhotosPhoto) ToAttachment() string {
	return fmt.Sprintf("photo%d_%d", photo.OwnerID, photo.ID)
}

// MaxSize return the largest PhotosPhotoSizes.
func (photo PhotosPhoto) MaxSize() (maxPhotoSize PhotosPhotoSizes) {
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

// MinSize return the smallest PhotosPhotoSizes.
func (photo PhotosPhoto) MinSize() (minPhotoSize PhotosPhotoSizes) {
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

// PhotosCommentXtrPid struct.
type PhotosCommentXtrPid struct {
	Attachments    []WallCommentAttachment `json:"attachments"`
	Date           int                     `json:"date"`    // Date when the comment has been added in Unixtime
	FromID         int                     `json:"from_id"` // Author ID
	ID             int                     `json:"id"`      // Comment ID
	Likes          BaseLikesInfo           `json:"likes"`
	ParentsStack   []int                   `json:"parents_stack"`
	Pid            int                     `json:"pid"`              // Photo ID
	ReplyToComment int                     `json:"reply_to_comment"` // Replied comment ID
	ReplyToUser    int                     `json:"reply_to_user"`    // Replied user ID
	Text           string                  `json:"text"`             // Comment text
	Thread         WallWallCommentThread   `json:"thread"`
}

// PhotosImage struct.
type PhotosImage struct {
	BaseImage
	Type string `json:"type"`
}

// PhotosChatUploadResponse struct.
type PhotosChatUploadResponse struct {
	Response string `json:"response"` // Uploaded photo data
}

// PhotosMarketAlbumUploadResponse struct.
type PhotosMarketAlbumUploadResponse struct {
	GID    int    `json:"gid"`    // Community ID
	Hash   string `json:"hash"`   // Uploading hash
	Photo  string `json:"photo"`  // Uploaded photo data
	Server int    `json:"server"` // Upload server number
}

// PhotosMarketUploadResponse struct.
type PhotosMarketUploadResponse struct {
	CropData string `json:"crop_data"` // Crop data
	CropHash string `json:"crop_hash"` // Crop hash
	GroupID  int    `json:"group_id"`  // Community ID
	Hash     string `json:"hash"`      // Uploading hash
	Photo    string `json:"photo"`     // Uploaded photo data
	Server   int    `json:"server"`    // Upload server number
}

// PhotosMessageUploadResponse struct.
type PhotosMessageUploadResponse struct {
	Hash   string `json:"hash"`   // Uploading hash
	Photo  string `json:"photo"`  // Uploaded photo data
	Server int    `json:"server"` // Upload server number
}

// PhotosOwnerUploadResponse struct.
type PhotosOwnerUploadResponse struct {
	Hash   string `json:"hash"`   // Uploading hash
	Photo  string `json:"photo"`  // Uploaded photo data
	Server int    `json:"server"` // Upload server number
}

// PhotosPhotoAlbum struct.
type PhotosPhotoAlbum struct {
	Created     int         `json:"created"`     // Date when the album has been created in Unixtime
	Description string      `json:"description"` // Photo album description
	ID          int         `json:"id"`          // Photo album ID
	OwnerID     int         `json:"owner_id"`    // Album owner's ID
	Size        int         `json:"size"`        // Photos number
	Thumb       PhotosPhoto `json:"thumb"`
	Title       string      `json:"title"`   // Photo album title
	Updated     int         `json:"updated"` // Date when the album has been updated last time in Unixtime
}

// ToAttachment return attachment format.
func (album PhotosPhotoAlbum) ToAttachment() string {
	return fmt.Sprintf("album%d_%d", album.OwnerID, album.ID)
}

// PhotosPhotoAlbumFull struct.
type PhotosPhotoAlbumFull struct {
	// Information whether current user can upload photo to the album.
	CanUpload        BaseBoolInt        `json:"can_upload"`
	CommentsDisabled BaseBoolInt        `json:"comments_disabled"` // Information whether album comments are disabled
	Created          int                `json:"created"`           // Date when the album has been created in Unixtime
	Description      string             `json:"description"`       // Photo album description
	ID               int                `json:"id"`                // Photo album ID
	OwnerID          int                `json:"owner_id"`          // Album owner's ID
	Size             int                `json:"size"`              // Photos number
	PrivacyComment   Privacy            `json:"privacy_comment"`
	PrivacyView      Privacy            `json:"privacy_view"`
	Sizes            []PhotosPhotoSizes `json:"sizes"`
	ThumbID          int                `json:"thumb_id"` // Thumb photo ID

	// Information whether the album thumb is last photo.
	ThumbIsLast int    `json:"thumb_is_last"`
	ThumbSrc    string `json:"thumb_src"` // URL of the thumb image
	Title       string `json:"title"`     // Photo album title

	// Date when the album has been updated last time in Unixtime.
	Updated int `json:"updated"`

	// Information whether only community administrators can upload photos.
	UploadByAdminsOnly int `json:"upload_by_admins_only"`
}

// ToAttachment return attachment format.
func (album PhotosPhotoAlbumFull) ToAttachment() string {
	return fmt.Sprintf("album%d_%d", album.OwnerID, album.ID)
}

// MaxSize return the largest PhotosPhotoSizes.
func (album PhotosPhotoAlbumFull) MaxSize() (maxPhotoSize PhotosPhotoSizes) {
	var max float64

	for _, photoSize := range album.Sizes {
		size := photoSize.Height * photoSize.Width
		if size > max {
			max = size
			maxPhotoSize = photoSize
		}
	}

	return
}

// MinSize return the smallest PhotosPhotoSizes.
func (album PhotosPhotoAlbumFull) MinSize() (minPhotoSize PhotosPhotoSizes) {
	var min float64

	for _, photoSize := range album.Sizes {
		size := photoSize.Height * photoSize.Width
		if size < min || min == 0 {
			min = size
			minPhotoSize = photoSize
		}
	}

	return
}

// PhotosPhotoFull struct.
type PhotosPhotoFull struct {
	AccessKey  string             `json:"access_key"`  // Access key for the photo
	AlbumID    int                `json:"album_id"`    // Album ID
	CanComment BaseBoolInt        `json:"can_comment"` // Information whether current user can comment the photo
	CanRepost  BaseBoolInt        `json:"can_repost"`  // Information whether current user can repost the photo
	HasTags    BaseBoolInt        `json:"has_tags"`
	Comments   BaseObjectCount    `json:"comments"`
	Date       int                `json:"date"`   // Date when uploaded
	Height     int                `json:"height"` // Original photo height
	ID         int                `json:"id"`     // Photo ID
	Images     []PhotosImage      `json:"images"`
	Lat        float64            `json:"lat"` // Latitude
	Likes      BaseLikes          `json:"likes"`
	Long       float64            `json:"long"`     // Longitude
	OwnerID    int                `json:"owner_id"` // Photo owner's ID
	PostID     int                `json:"post_id"`  // Post ID
	Reposts    BaseRepostsInfo    `json:"reposts"`
	Tags       BaseObjectCount    `json:"tags"`
	Text       string             `json:"text"`       // Photo caption
	UserID     int                `json:"user_id"`    // ID of the user who have uploaded the photo
	Width      int                `json:"width"`      // Original photo width
	Hidden     int                `json:"hidden"`     // Returns if the photo is hidden above the wall
	Photo75    string             `json:"photo_75"`   // URL of image with 75 px width
	Photo130   string             `json:"photo_130"`  // URL of image with 130 px width
	Photo604   string             `json:"photo_604"`  // URL of image with 604 px width
	Photo807   string             `json:"photo_807"`  // URL of image with 807 px width
	Photo1280  string             `json:"photo_1280"` // URL of image with 1280 px width
	Photo2560  string             `json:"photo_2560"` // URL of image with 2560 px width
	Sizes      []PhotosPhotoSizes `json:"sizes"`
	OrigPhoto  PhotosPhotoSizes   `json:"orig_photo"`
}

// ToAttachment return attachment format.
func (photo PhotosPhotoFull) ToAttachment() string {
	return fmt.Sprintf("photo%d_%d", photo.OwnerID, photo.ID)
}

// MaxSize return the largest PhotosPhotoSizes.
func (photo PhotosPhotoFull) MaxSize() (maxPhotoSize PhotosPhotoSizes) {
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

// MinSize return the smallest PhotosPhotoSizes.
func (photo PhotosPhotoFull) MinSize() (minPhotoSize PhotosPhotoSizes) {
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

// PhotosPhotoFullXtrRealOffset struct.
type PhotosPhotoFullXtrRealOffset struct {
	PhotosPhotoFull
	RealOffset int `json:"real_offset"` // Real position of the photo
}

// PhotosPhotoSizes struct.
type PhotosPhotoSizes struct {
	// BUG(VK): json: cannot unmarshal number 180.000000 into Go struct field PhotosPhotoSizes.height of type int
	BaseImage
}

// PhotosPhotoTag struct.
type PhotosPhotoTag struct {
	Date        int         `json:"date"`        // Date when tag has been added in Unixtime
	ID          int         `json:"id"`          // Tag ID
	PlacerID    int         `json:"placer_id"`   // ID of the tag creator
	TaggedName  string      `json:"tagged_name"` // Tag description
	Description string      `json:"description"` // Tagged description.
	UserID      int         `json:"user_id"`     // Tagged user ID
	Viewed      BaseBoolInt `json:"viewed"`      // Information whether the tag is reviewed
	X           float64     `json:"x"`           // Coordinate X of the left upper corner
	X2          float64     `json:"x2"`          // Coordinate X of the right lower corner
	Y           float64     `json:"y"`           // Coordinate Y of the left upper corner
	Y2          float64     `json:"y2"`          // Coordinate Y of the right lower corner
}

// PhotosPhotoUpload struct.
type PhotosPhotoUpload struct {
	AlbumID   int    `json:"album_id"`   // Album ID
	UploadURL string `json:"upload_url"` // URL to upload photo
	UserID    int    `json:"user_id"`    // User ID
}

// PhotosPhotoUploadResponse struct.
type PhotosPhotoUploadResponse struct {
	AID        int    `json:"aid"`         // Album ID
	Hash       string `json:"hash"`        // Uploading hash
	PhotosList string `json:"photos_list"` // Uploaded photos data
	Server     int    `json:"server"`      // Upload server number
}

// PhotosPhotoXtrRealOffset struct.
type PhotosPhotoXtrRealOffset struct {
	PhotosPhoto
	RealOffset int `json:"real_offset"` // Real position of the photo
}

// PhotosPhotoXtrTagInfo struct.
type PhotosPhotoXtrTagInfo struct {
	PhotosPhoto
	TagCreated int `json:"tag_created"` // Date when tag has been added in Unixtime
	TagID      int `json:"tag_id"`      // Tag ID
}

// PhotosWallUploadResponse struct.
type PhotosWallUploadResponse struct {
	Hash   string `json:"hash"`   // Uploading hash
	Photo  string `json:"photo"`  // Uploaded photo data
	Server int    `json:"server"` // Upload server number
}
