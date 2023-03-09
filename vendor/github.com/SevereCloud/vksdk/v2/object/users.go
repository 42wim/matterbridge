package object // import "github.com/SevereCloud/vksdk/v2/object"

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/vmihailenco/msgpack/v5"
	"github.com/vmihailenco/msgpack/v5/msgpcode"
)

// User relationship status.
const (
	UserRelationNotSpecified      = iota // not specified
	UserRelationSingle                   // single
	UserRelationInRelationship           // in a relationship
	UserRelationEngaged                  // engaged
	UserRelationMarried                  // married
	UserRelationComplicated              // complicated
	UserRelationActivelySearching        // actively searching
	UserRelationInLove                   // in love
	UserRelationCivilUnion               // in a civil union
)

// UsersUser struct.
type UsersUser struct {
	ID                     int         `json:"id"`
	FirstName              string      `json:"first_name"`
	LastName               string      `json:"last_name"`
	FirstNameNom           string      `json:"first_name_nom"`
	FirstNameGen           string      `json:"first_name_gen"`
	FirstNameDat           string      `json:"first_name_dat"`
	FirstNameAcc           string      `json:"first_name_acc"`
	FirstNameIns           string      `json:"first_name_ins"`
	FirstNameAbl           string      `json:"first_name_abl"`
	LastNameNom            string      `json:"last_name_nom"`
	LastNameGen            string      `json:"last_name_gen"`
	LastNameDat            string      `json:"last_name_dat"`
	LastNameAcc            string      `json:"last_name_acc"`
	LastNameIns            string      `json:"last_name_ins"`
	LastNameAbl            string      `json:"last_name_abl"`
	MaidenName             string      `json:"maiden_name"`
	Sex                    int         `json:"sex"`
	Nickname               string      `json:"nickname"`
	Domain                 string      `json:"domain"`
	ScreenName             string      `json:"screen_name"`
	Bdate                  string      `json:"bdate"`
	City                   BaseObject  `json:"city"`
	Country                BaseObject  `json:"country"`
	Photo50                string      `json:"photo_50"`
	Photo100               string      `json:"photo_100"`
	Photo200               string      `json:"photo_200"`
	PhotoMax               string      `json:"photo_max"`
	Photo200Orig           string      `json:"photo_200_orig"`
	Photo400Orig           string      `json:"photo_400_orig"`
	PhotoMaxOrig           string      `json:"photo_max_orig"`
	PhotoID                string      `json:"photo_id"`
	FriendStatus           int         `json:"friend_status"` // see FriendStatus const
	OnlineApp              int         `json:"online_app"`
	Online                 BaseBoolInt `json:"online"`
	OnlineMobile           BaseBoolInt `json:"online_mobile"`
	HasPhoto               BaseBoolInt `json:"has_photo"`
	HasMobile              BaseBoolInt `json:"has_mobile"`
	IsClosed               BaseBoolInt `json:"is_closed"`
	IsFriend               BaseBoolInt `json:"is_friend"`
	IsFavorite             BaseBoolInt `json:"is_favorite"`
	IsHiddenFromFeed       BaseBoolInt `json:"is_hidden_from_feed"`
	CanAccessClosed        BaseBoolInt `json:"can_access_closed"`
	CanBeInvitedGroup      BaseBoolInt `json:"can_be_invited_group"`
	CanPost                BaseBoolInt `json:"can_post"`
	CanSeeAllPosts         BaseBoolInt `json:"can_see_all_posts"`
	CanSeeAudio            BaseBoolInt `json:"can_see_audio"`
	CanWritePrivateMessage BaseBoolInt `json:"can_write_private_message"`
	CanSendFriendRequest   BaseBoolInt `json:"can_send_friend_request"`
	CanCallFromGroup       BaseBoolInt `json:"can_call_from_group"`
	Verified               BaseBoolInt `json:"verified"`
	Trending               BaseBoolInt `json:"trending"`
	Blacklisted            BaseBoolInt `json:"blacklisted"`
	BlacklistedByMe        BaseBoolInt `json:"blacklisted_by_me"`
	// Deprecated: Facebook и Instagram запрещены в России, Meta признана экстремистской организацией...
	Facebook string `json:"facebook"`
	// Deprecated: Facebook и Instagram запрещены в России, Meta признана экстремистской организацией...
	FacebookName string `json:"facebook_name"`
	// Deprecated: Facebook и Instagram запрещены в России, Meta признана экстремистской организацией...
	Instagram       string                `json:"instagram"`
	Twitter         string                `json:"twitter"`
	Site            string                `json:"site"`
	Status          string                `json:"status"`
	StatusAudio     AudioAudio            `json:"status_audio"`
	LastSeen        UsersLastSeen         `json:"last_seen"`
	CropPhoto       UsersCropPhoto        `json:"crop_photo"`
	FollowersCount  int                   `json:"followers_count"`
	CommonCount     int                   `json:"common_count"`
	Occupation      UsersOccupation       `json:"occupation"`
	Career          []UsersCareer         `json:"career"`
	Military        []UsersMilitary       `json:"military"`
	University      int                   `json:"university"`
	UniversityName  string                `json:"university_name"`
	Faculty         int                   `json:"faculty"`
	FacultyName     string                `json:"faculty_name"`
	Graduation      int                   `json:"graduation"`
	EducationForm   string                `json:"education_form"`
	EducationStatus string                `json:"education_status"`
	HomeTown        string                `json:"home_town"`
	Relation        int                   `json:"relation"`
	Personal        UsersPersonal         `json:"personal"`
	Interests       string                `json:"interests"`
	Music           string                `json:"music"`
	Activities      string                `json:"activities"`
	Movies          string                `json:"movies"`
	Tv              string                `json:"tv"`
	Books           string                `json:"books"`
	Games           string                `json:"games"`
	Universities    []UsersUniversity     `json:"universities"`
	Schools         []UsersSchool         `json:"schools"`
	About           string                `json:"about"`
	Relatives       []UsersRelative       `json:"relatives"`
	Quotes          string                `json:"quotes"`
	Lists           []int                 `json:"lists"`
	Deactivated     string                `json:"deactivated"`
	WallDefault     string                `json:"wall_default"`
	Timezone        int                   `json:"timezone"`
	Exports         UsersExports          `json:"exports"`
	Counters        UsersUserCounters     `json:"counters"`
	MobilePhone     string                `json:"mobile_phone"`
	HomePhone       string                `json:"home_phone"`
	FoundWith       int                   `json:"found_with"` // TODO: check it
	ImageStatus     ImageStatusInfo       `json:"image_status"`
	OnlineInfo      UsersOnlineInfo       `json:"online_info"`
	Mutual          FriendsRequestsMutual `json:"mutual"`
	TrackCode       string                `json:"track_code"`
	RelationPartner UsersUserMin          `json:"relation_partner"`
	Type            string                `json:"type"`
	Skype           string                `json:"skype"`
}

// ToMention return mention.
func (user UsersUser) ToMention() string {
	return fmt.Sprintf("[id%d|%s %s]", user.ID, user.FirstName, user.LastName)
}

// ImageStatusInfo struct.
type ImageStatusInfo struct {
	ID     int         `json:"id"`
	Name   string      `json:"name"`
	Images []BaseImage `json:"images"`
}

// UsersOnlineInfo struct.
type UsersOnlineInfo struct {
	AppID    int         `json:"app_id"`
	LastSeen int         `json:"last_seen"`
	Status   string      `json:"status"`
	Visible  BaseBoolInt `json:"visible"`
	IsOnline BaseBoolInt `json:"is_online"`
	IsMobile BaseBoolInt `json:"is_mobile"`
}

// UsersUserMin struct.
type UsersUserMin struct {
	Deactivated string `json:"deactivated"` // Returns if a profile is deleted or blocked
	FirstName   string `json:"first_name"`  // User first name
	Hidden      int    `json:"hidden"`      // Returns if a profile is hidden.
	ID          int    `json:"id"`          // User ID
	LastName    string `json:"last_name"`   // User last name
}

// ToMention return mention.
func (user UsersUserMin) ToMention() string {
	return fmt.Sprintf("[id%d|%s %s]", user.ID, user.FirstName, user.LastName)
}

// UsersCareer struct.
type UsersCareer struct {
	CityID    int    `json:"city_id"`    // City ID
	CityName  string `json:"city_name"`  // City name
	Company   string `json:"company"`    // Company name
	CountryID int    `json:"country_id"` // Country ID
	From      int    `json:"from"`       // From year
	GroupID   int    `json:"group_id"`   // Community ID
	ID        int    `json:"id"`         // Career ID
	Position  string `json:"position"`   // Position
	Until     int    `json:"until"`      // Till year
}

// UsersCropPhoto struct.
type UsersCropPhoto struct {
	Crop  UsersCropPhotoCrop `json:"crop"`
	Photo PhotosPhoto        `json:"photo"`
	Rect  UsersCropPhotoRect `json:"rect"`
}

// UsersCropPhotoCrop struct.
type UsersCropPhotoCrop struct {
	X  float64 `json:"x"`  // Coordinate X of the left upper corner
	X2 float64 `json:"x2"` // Coordinate X of the right lower corner
	Y  float64 `json:"y"`  // Coordinate Y of the left upper corner
	Y2 float64 `json:"y2"` // Coordinate Y of the right lower corner
}

// UsersCropPhotoRect struct.
type UsersCropPhotoRect struct {
	X  float64 `json:"x"`  // Coordinate X of the left upper corner
	X2 float64 `json:"x2"` // Coordinate X of the right lower corner
	Y  float64 `json:"y"`  // Coordinate Y of the left upper corner
	Y2 float64 `json:"y2"` // Coordinate Y of the right lower corner
}

// UsersExports struct.
type UsersExports struct {
	Facebook    int `json:"facebook"`
	Livejournal int `json:"livejournal"`
	Twitter     int `json:"twitter"`
}

// UsersLastSeen struct.
type UsersLastSeen struct {
	Platform int `json:"platform"` // Type of the platform that used for the last authorization
	Time     int `json:"time"`     // Last visit date (in Unix time)
}

// UsersMilitary struct.
type UsersMilitary struct {
	CountryID int    `json:"country_id"` // Country ID
	From      int    `json:"from"`       // From year
	ID        int    `json:"id"`         // Military ID
	Unit      string `json:"unit"`       // Unit name
	UnitID    int    `json:"unit_id"`    // Unit ID
	Until     int    `json:"until"`      // Till year
}

// UsersOccupation struct.
type UsersOccupation struct {
	// BUG(VK): UsersOccupation.ID is float https://vk.com/bug136108
	ID   float64 `json:"id"`   // ID of school, university, company group
	Name string  `json:"name"` // Name of occupation
	Type string  `json:"type"` // Type of occupation
}

// UsersPersonal struct.
type UsersPersonal struct {
	Alcohol    int      `json:"alcohol"`     // User's views on alcohol
	InspiredBy string   `json:"inspired_by"` // User's inspired by
	Langs      []string `json:"langs"`
	LifeMain   int      `json:"life_main"`   // User's personal priority in life
	PeopleMain int      `json:"people_main"` // User's personal priority in people
	Political  int      `json:"political"`   // User's political views
	Religion   string   `json:"religion"`    // User's religion
	Smoking    int      `json:"smoking"`     // User's views on smoking
	ReligionID int      `json:"religion_id"`
}

// UnmarshalJSON UsersPersonal.
//
// BUG(VK): UsersPersonal return [].
func (personal *UsersPersonal) UnmarshalJSON(data []byte) error {
	if bytes.Equal(data, []byte("[]")) {
		return nil
	}

	type renamedUsersPersonal UsersPersonal

	var r renamedUsersPersonal

	err := json.Unmarshal(data, &r)
	if err != nil {
		return err
	}

	*personal = UsersPersonal(r)

	return nil
}

// DecodeMsgpack UsersPersonal.
//
// BUG(VK): UsersPersonal return [].
func (personal *UsersPersonal) DecodeMsgpack(dec *msgpack.Decoder) error {
	data, err := dec.DecodeRaw()
	if err != nil {
		return err
	}

	if bytes.Equal(data, []byte{msgpcode.FixedArrayLow}) {
		return nil
	}

	type renamedUsersPersonal UsersPersonal

	var r renamedUsersPersonal

	d := msgpack.NewDecoder(bytes.NewReader(data))
	d.SetCustomStructTag("json")

	err = d.Decode(&r)
	if err != nil {
		return err
	}

	*personal = UsersPersonal(r)

	return nil
}

// UsersRelative struct.
type UsersRelative struct {
	BirthDate string `json:"birth_date"` // Date of child birthday (format dd.mm.yyyy)
	ID        int    `json:"id"`         // Relative ID
	Name      string `json:"name"`       // Name of relative
	Type      string `json:"type"`       // Relative type
}

// UsersSchool struct.
type UsersSchool struct {
	City          int    `json:"city"`           // City ID
	Class         string `json:"class"`          // School class letter
	Country       int    `json:"country"`        // Country ID
	ID            string `json:"id"`             // School ID
	Name          string `json:"name"`           // School name
	Type          int    `json:"type"`           // School type ID
	TypeStr       string `json:"type_str"`       // School type name
	YearFrom      int    `json:"year_from"`      // Year the user started to study
	YearGraduated int    `json:"year_graduated"` // Graduation year
	YearTo        int    `json:"year_to"`        // Year the user finished to study
	Speciality    string `json:"speciality,omitempty"`
}

// UsersUniversity struct.
type UsersUniversity struct {
	Chair           int    `json:"chair"`            // Chair ID
	ChairName       string `json:"chair_name"`       // Chair name
	City            int    `json:"city"`             // City ID
	Country         int    `json:"country"`          // Country ID
	EducationForm   string `json:"education_form"`   // Education form
	EducationStatus string `json:"education_status"` // Education status
	Faculty         int    `json:"faculty"`          // Faculty ID
	FacultyName     string `json:"faculty_name"`     // Faculty name
	Graduation      int    `json:"graduation"`       // Graduation year
	ID              int    `json:"id"`               // University ID
	Name            string `json:"name"`             // University name
}

// UsersUserCounters struct.
type UsersUserCounters struct {
	Albums        int `json:"albums"`         // Albums number
	Audios        int `json:"audios"`         // Audios number
	Followers     int `json:"followers"`      // Followers number
	Friends       int `json:"friends"`        // Friends number
	Gifts         int `json:"gifts"`          // Gifts number
	Groups        int `json:"groups"`         // Communities number
	Notes         int `json:"notes"`          // Notes number
	OnlineFriends int `json:"online_friends"` // Online friends number
	Pages         int `json:"pages"`          // Public pages number
	Photos        int `json:"photos"`         // Photos number
	Subscriptions int `json:"subscriptions"`  // Subscriptions number
	UserPhotos    int `json:"user_photos"`    // Number of photos with user
	UserVideos    int `json:"user_videos"`    // Number of videos with user
	Videos        int `json:"videos"`         // Videos number
	MutualFriends int `json:"mutual_friends"`
}

// UsersUserLim struct.
type UsersUserLim struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	NameGen string `json:"name_gen"`
	Photo   string `json:"photo"`
}
