package object // import "github.com/SevereCloud/vksdk/v2/object"

import (
	"encoding/json"
	"fmt"
	"reflect"
)

// GroupsAddress WorkInfoStatus of information about timetable.
const (
	WorkStatusNoInformation     = "no_information"
	WorkStatusTemporarilyClosed = "temporarily_closed"
	WorkStatusAlwaysOpened      = "always_opened"
	WorkStatusTimetable         = "timetable"
	WorkStatusForeverClosed     = "forever_closed"
)

// GroupsAddress struct.
type GroupsAddress struct {
	// Additional address to the place (6 floor, left door).
	AdditionalAddress string                 `json:"additional_address"`
	Address           string                 `json:"address"`          // String address to the place (Nevsky, 28)
	CityID            int                    `json:"city_id"`          // City id of address
	CountryID         int                    `json:"country_id"`       // Country id of address
	Distance          int                    `json:"distance"`         // Distance from the point
	ID                int                    `json:"id"`               // Address id
	Latitude          float64                `json:"latitude"`         // Address latitude
	Longitude         float64                `json:"longitude"`        // Address longitude
	MetroStationID    int                    `json:"metro_station_id"` // Metro id of address
	Phone             string                 `json:"phone"`            // Address phone
	TimeOffset        int                    `json:"time_offset"`      // Time offset int minutes from utc time
	Timetable         GroupsAddressTimetable `json:"timetable"`        // Week timetable for the address
	Title             string                 `json:"title"`            // Title of the place (Zinger, etc)
	WorkInfoStatus    string                 `json:"work_info_status"` // Status of information about timetable
	PlaceID           int                    `json:"place_id"`
}

// GroupsAddressTimetable Timetable for a week.
type GroupsAddressTimetable struct {
	Fri GroupsAddressTimetableDay `json:"fri"` // Timetable for friday
	Mon GroupsAddressTimetableDay `json:"mon"` // Timetable for monday
	Sat GroupsAddressTimetableDay `json:"sat"` // Timetable for saturday
	Sun GroupsAddressTimetableDay `json:"sun"` // Timetable for sunday
	Thu GroupsAddressTimetableDay `json:"thu"` // Timetable for thursday
	Tue GroupsAddressTimetableDay `json:"tue"` // Timetable for tuesday
	Wed GroupsAddressTimetableDay `json:"wed"` // Timetable for wednesday
}

// GroupsAddressTimetableDay Timetable for one day.
type GroupsAddressTimetableDay struct {
	BreakCloseTime int `json:"break_close_time"` // Close time of the break in minutes
	BreakOpenTime  int `json:"break_open_time"`  // Start time of the break in minutes
	CloseTime      int `json:"close_time"`       // Close time in minutes
	OpenTime       int `json:"open_time"`        // Open time in minutes
}

// GroupsAddressesInfo struct.
type GroupsAddressesInfo struct {
	IsEnabled     BaseBoolInt `json:"is_enabled"`      // Information whether addresses is enabled
	MainAddressID int         `json:"main_address_id"` // Main address id for group
}

// GroupsGroup AdminLevel type.
const (
	GroupsAdminLevelModerator = iota
	GroupsAdminLevelEditor
	GroupsAdminLevelAdministrator
)

// GroupsGroup MainSection type.
const (
	GroupsMainSectionAbsent = iota
	GroupsMainSectionPhotos
	GroupsMainSectionTopics
	GroupsMainSectionAudio
	GroupsMainSectionVideo
	GroupsMainSectionMarket
)

// GroupsGroup MemberStatus(events_event_attach, newsfeed_event_activity).
const (
	GroupsMemberStatusNotMember = iota
	GroupsMemberStatusMember
	GroupsMemberStatusNotSure
	GroupsMemberStatusDeclined
	GroupsMemberStatusHasSentRequest
	GroupsMemberStatusInvited
)

// GroupsGroup Access or IsClosed type.
const (
	GroupsGroupOpen = iota
	GroupsGroupClosed
	GroupsGroupPrivate
)

// GroupsGroup AgeLimits.
const (
	GroupsAgeLimitsNo = iota
	GroupsAgeLimitsOver16
	GroupsAgeLimitsOver18
)

// GroupsGroup type.
const (
	GroupsTypeGroup = "group"
	GroupsTypePage  = "page"
	GroupsTypeEvent = "event"
)

// GroupsGroup struct.
type GroupsGroup struct {
	AdminLevel   int              `json:"admin_level"`
	Deactivated  string           `json:"deactivated"` // Information whether community is banned
	FinishDate   int              `json:"finish_date"` // Finish date in Unixtime format
	ID           int              `json:"id"`          // Community ID
	Name         string           `json:"name"`        // Community name
	Photo100     string           `json:"photo_100"`   // URL of square photo of the community with 100 pixels in width
	Photo200     string           `json:"photo_200"`   // URL of square photo of the community with 200 pixels in width
	Photo50      string           `json:"photo_50"`    // URL of square photo of the community with 50 pixels in width
	ScreenName   string           `json:"screen_name"` // Domain of the community page
	StartDate    int              `json:"start_date"`  // Start date in Unixtime format
	Type         string           `json:"type"`
	Market       GroupsMarketInfo `json:"market"`
	MemberStatus int              `json:"member_status"` // Current user's member status
	IsClosed     int              `json:"is_closed"`
	City         BaseObject       `json:"city"`
	Country      BaseCountry      `json:"country"`

	// Information whether current user is administrator.
	IsAdmin BaseBoolInt `json:"is_admin"`

	// Information whether current user is advertiser.
	IsAdvertiser BaseBoolInt `json:"is_advertiser"`

	// Information whether current user is member.
	IsMember BaseBoolInt `json:"is_member"`

	// Information whether community is in faves.
	IsFavorite BaseBoolInt `json:"is_favorite"`

	// Information whether community is adult.
	IsAdult BaseBoolInt `json:"is_adult"`

	// Information whether current user is subscribed.
	IsSubscribed BaseBoolInt `json:"is_subscribed"`

	// Information whether current user can post on community's wall.
	CanPost BaseBoolInt `json:"can_post"`

	// Information whether current user can see all posts on community's wall.
	CanSeeAllPosts BaseBoolInt `json:"can_see_all_posts"`

	// Information whether current user can create topic.
	CanCreateTopic BaseBoolInt `json:"can_create_topic"`

	// Information whether current user can upload video.
	CanUploadVideo BaseBoolInt `json:"can_upload_video"`

	// Information whether current user can upload doc.
	CanUploadDoc BaseBoolInt `json:"can_upload_doc"`

	// Information whether community has photo.
	HasPhoto BaseBoolInt `json:"has_photo"`

	// Information whether current user can send a message to community.
	CanMessage BaseBoolInt `json:"can_message"`

	// Information whether community can send a message to current user.
	IsMessagesBlocked BaseBoolInt `json:"is_messages_blocked"`

	// Information whether community can send notifications by phone number to current user.
	CanSendNotify BaseBoolInt `json:"can_send_notify"`

	// Information whether current user is subscribed to podcasts.
	IsSubscribedPodcasts BaseBoolInt `json:"is_subscribed_podcasts"`

	// Owner in whitelist or not.
	CanSubscribePodcasts BaseBoolInt `json:"can_subscribe_podcasts"`

	// Can subscribe to wall.
	CanSubscribePosts BaseBoolInt `json:"can_subscribe_posts"`

	// Information whether community has market app.
	HasMarketApp        BaseBoolInt `json:"has_market_app"`
	IsHiddenFromFeed    BaseBoolInt `json:"is_hidden_from_feed"`
	IsMarketCartEnabled BaseBoolInt `json:"is_market_cart_enabled"`
	Verified            BaseBoolInt `json:"verified"` // Information whether community is verified

	// Information whether the community has a fire pictogram.
	Trending     BaseBoolInt         `json:"trending"`
	Description  string              `json:"description"`   // Community description
	WikiPage     string              `json:"wiki_page"`     // Community's main wiki page title
	MembersCount int                 `json:"members_count"` // Community members number
	Counters     GroupsCountersGroup `json:"counters"`
	Cover        GroupsCover         `json:"cover"`

	// Type of group, start date of event or category of public page.
	Activity        string               `json:"activity"`
	FixedPost       int                  `json:"fixed_post"`    // Fixed post ID
	Status          string               `json:"status"`        // Community status
	MainAlbumID     int                  `json:"main_album_id"` // Community's main photo album ID
	Links           []GroupsLinksItem    `json:"links"`
	Contacts        []GroupsContactsItem `json:"contacts"`
	Site            string               `json:"site"` // Community's website
	MainSection     int                  `json:"main_section"`
	OnlineStatus    GroupsOnlineStatus   `json:"online_status"` // Status of replies in community messages
	AgeLimits       int                  `json:"age_limits"`    // Information whether age limit
	BanInfo         GroupsGroupBanInfo   `json:"ban_info"`      // User ban info
	Addresses       GroupsAddressesInfo  `json:"addresses"`     // Info about addresses in Groups
	LiveCovers      GroupsLiveCovers     `json:"live_covers"`
	CropPhoto       UsersCropPhoto       `json:"crop_photo"`
	Wall            int                  `json:"wall"`
	ActionButton    GroupsActionButton   `json:"action_button"`
	TrackCode       string               `json:"track_code"`
	PublicDateLabel string               `json:"public_date_label"`
	AuthorID        int                  `json:"author_id"`
	Phone           string               `json:"phone"`
}

// ToMention return mention.
func (group GroupsGroup) ToMention() string {
	return fmt.Sprintf("[club%d|%s]", group.ID, group.Name)
}

// GroupsLiveCovers struct.
type GroupsLiveCovers struct {
	IsEnabled  BaseBoolInt `json:"is_enabled"`
	IsScalable BaseBoolInt `json:"is_scalable"`
	StoryIds   []string    `json:"story_ids"`
}

// GroupsBanInfo reason type.
const (
	GroupsBanReasonOther = iota
	GroupsBanReasonSpam
	GroupsBanReasonVerbalAbuse
	GroupsBanReasonStrongLanguage
	GroupsBanReasonFlood
)

// GroupsBanInfo struct.
type GroupsBanInfo struct {
	AdminID        int         `json:"admin_id"` // Administrator ID
	Comment        string      `json:"comment"`  // Comment for a ban
	Date           int         `json:"date"`     // Date when user has been added to blacklist in Unixtime
	EndDate        int         `json:"end_date"` // Date when user will be removed from blacklist in Unixtime
	Reason         int         `json:"reason"`
	CommentVisible BaseBoolInt `json:"comment_visible"`
}

// GroupsCallbackServer struct.
type GroupsCallbackServer struct {
	CreatorID int    `json:"creator_id"`
	ID        int    `json:"id"`
	SecretKey string `json:"secret_key"`
	Status    string `json:"status"`
	Title     string `json:"title"`
	URL       string `json:"url"`
}

// GroupsCallbackSettings struct.
type GroupsCallbackSettings struct {
	APIVersion string               `json:"api_version"` // API version used for the events
	Events     GroupsLongPollEvents `json:"events"`
}

// GroupsContactsItem struct.
type GroupsContactsItem struct {
	Desc   string `json:"desc"`    // Contact description
	Email  string `json:"email"`   // Contact email
	Phone  string `json:"phone"`   // Contact phone
	UserID int    `json:"user_id"` // User ID
}

// GroupsCountersGroup struct.
type GroupsCountersGroup struct {
	Addresses  int `json:"addresses"`  // Addresses number
	Albums     int `json:"albums"`     // Photo albums number
	Articles   int `json:"articles"`   // Articles number
	Audios     int `json:"audios"`     // Audios number
	Docs       int `json:"docs"`       // Docs number
	Market     int `json:"market"`     // Market items number
	Photos     int `json:"photos"`     // Photos number
	Topics     int `json:"topics"`     // Topics number
	Videos     int `json:"videos"`     // Videos number
	Narratives int `json:"narratives"` // Narratives number
}

// GroupsCover struct.
type GroupsCover struct {
	Enabled BaseBoolInt `json:"enabled"` // Information whether cover is enabled
	Images  []BaseImage `json:"images"`
}

// GroupsGroupBanInfo struct.
type GroupsGroupBanInfo struct {
	Comment string `json:"comment"`  // Ban comment
	EndDate int    `json:"end_date"` // End date of ban in Unixtime
}

// GroupsGroupCategory struct.
type GroupsGroupCategory struct {
	ID            int                  `json:"id"`   // Category ID
	Name          string               `json:"name"` // Category name
	Subcategories []BaseObjectWithName `json:"subcategories"`
}

// GroupsGroupCategoryFull struct.
type GroupsGroupCategoryFull struct {
	ID            int                       `json:"id"`         // Category ID
	Name          string                    `json:"name"`       // Category name
	PageCount     int                       `json:"page_count"` // Pages number
	PagePreviews  []GroupsGroup             `json:"page_previews"`
	Subcategories []GroupsGroupCategoryFull `json:"subcategories"`
}

// GroupsGroupCategoryType struct.
type GroupsGroupCategoryType struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// GroupsGroupLink struct.
type GroupsGroupLink struct {
	Desc            string      `json:"desc"`             // Link description
	EditTitle       BaseBoolInt `json:"edit_title"`       // Information whether the title can be edited
	ImageProcessing BaseBoolInt `json:"image_processing"` // Information whether the image on processing
	Name            string      `json:"name"`
	ID              int         `json:"id"`  // Link ID
	URL             string      `json:"url"` // Link URL
}

// GroupsGroupPublicCategoryList struct.
type GroupsGroupPublicCategoryList struct {
	ID            int                       `json:"id"`
	Name          string                    `json:"name"`
	Subcategories []GroupsGroupCategoryType `json:"subcategories"`
}

// GroupsGroupSettings Photos type.
const (
	GroupsGroupPhotosDisabled = iota
	GroupsGroupPhotosOpen
	GroupsGroupPhotosLimited
)

// GroupsGroupSettings Subject type.
const (
	_ = iota
	GroupsGroupSubjectAuto
	GroupsGroupSubjectActivityHolidays
	GroupsGroupSubjectBusiness
	GroupsGroupSubjectPets
	GroupsGroupSubjectHealth
	GroupsGroupSubjectDatingAndCommunication
	GroupsGroupSubjectGames
	GroupsGroupSubjectIt
	GroupsGroupSubjectCinema
	GroupsGroupSubjectBeautyAndFashion
	GroupsGroupSubjectCooking
	GroupsGroupSubjectArtAndCulture
	GroupsGroupSubjectLiterature
	GroupsGroupSubjectMobileServicesAndInternet
	GroupsGroupSubjectMusic
	GroupsGroupSubjectScienceAndTechnology
	GroupsGroupSubjectRealEstate
	GroupsGroupSubjectNewsAndMedia
	GroupsGroupSubjectSecurity
	GroupsGroupSubjectEducation
	GroupsGroupSubjectHomeAndRenovations
	GroupsGroupSubjectPolitics
	GroupsGroupSubjectFood
	GroupsGroupSubjectIndustry
	GroupsGroupSubjectTravel
	GroupsGroupSubjectWork
	GroupsGroupSubjectEntertainment
	GroupsGroupSubjectReligion
	GroupsGroupSubjectFamily
	GroupsGroupSubjectSports
	GroupsGroupSubjectInsurance
	GroupsGroupSubjectTelevision
	GroupsGroupSubjectGoodsAndServices
	GroupsGroupSubjectHobbies
	GroupsGroupSubjectFinance
	GroupsGroupSubjectPhoto
	GroupsGroupSubjectEsoterics
	GroupsGroupSubjectElectronicsAndAppliances
	GroupsGroupSubjectErotic
	GroupsGroupSubjectHumor
	GroupsGroupSubjectSocietyHumanities
	GroupsGroupSubjectDesignAndGraphics
)

// GroupsGroupSettings Topics type.
const (
	GroupsGroupTopicsDisabled = iota
	GroupsGroupTopicsOpen
	GroupsGroupTopicsLimited
)

// GroupsGroupSettings Docs type.
const (
	GroupsGroupDocsDisabled = iota
	GroupsGroupDocsOpen
	GroupsGroupDocsLimited
)

// GroupsGroupSettings Audio type.
const (
	GroupsGroupAudioDisabled = iota
	GroupsGroupAudioOpen
	GroupsGroupAudioLimited
)

// GroupsGroupSettings Video type.
const (
	GroupsGroupVideoDisabled = iota
	GroupsGroupVideoOpen
	GroupsGroupVideoLimited
)

// GroupsGroupSettings Wall type.
const (
	GroupsGroupWallDisabled = iota
	GroupsGroupWallOpen
	GroupsGroupWallLimited
	GroupsGroupWallClosed
)

// GroupsGroupSettings Wiki type.
const (
	GroupsGroupWikiDisabled = iota
	GroupsGroupWikiOpen
	GroupsGroupWikiLimited
)

// GroupsGroupSettings struct.
type GroupsGroupSettings struct {
	Access             int                             `json:"access"`          // Community access settings
	Address            string                          `json:"address"`         // Community's page domain
	Audio              int                             `json:"audio"`           // Audio settings
	Description        string                          `json:"description"`     // Community description
	Docs               int                             `json:"docs"`            // Docs settings
	ObsceneWords       []string                        `json:"obscene_words"`   // The list of stop words
	Photos             int                             `json:"photos"`          // Photos settings
	PublicCategory     int                             `json:"public_category"` // Information about the group category
	PublicCategoryList []GroupsGroupPublicCategoryList `json:"public_category_list"`

	// Information about the group subcategory.
	PublicSubcategory int                 `json:"public_subcategory"`
	Rss               string              `json:"rss"`     // URL of the RSS feed
	Subject           int                 `json:"subject"` // Community subject ID
	SubjectList       []GroupsSubjectItem `json:"subject_list"`
	Title             string              `json:"title"`   // Community title
	Topics            int                 `json:"topics"`  // Topics settings
	Video             int                 `json:"video"`   // Video settings
	Wall              int                 `json:"wall"`    // Wall settings
	Website           string              `json:"website"` // Community website
	Wiki              int                 `json:"wiki"`    // Wiki settings
	CountryID         int                 `json:"country_id"`
	CityID            int                 `json:"city_id"`
	Messages          int                 `json:"messages"`
	Articles          int                 `json:"articles"`
	Events            int                 `json:"events"`
	AgeLimits         int                 `json:"age_limits"`

	// Information whether the obscene filter is enabled.
	ObsceneFilter BaseBoolInt `json:"obscene_filter"`

	// Information whether the stopwords filter is enabled.
	ObsceneStopwords BaseBoolInt `json:"obscene_stopwords"`
	LiveCovers       struct {
		IsEnabled BaseBoolInt `json:"is_enabled"`
	} `json:"live_covers"`
	Market           GroupsMarketInfo     `json:"market"`
	SectionsList     []GroupsSectionsList `json:"sections_list"`
	MainSection      int                  `json:"main_section"`
	SecondarySection int                  `json:"secondary_section"`
	ActionButton     GroupsActionButton   `json:"action_button"`
	Phone            string               `json:"phone"`
}

// GroupsSectionsList struct.
type GroupsSectionsList struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// UnmarshalJSON need for unmarshal dynamic array (Example: [1, "Фотографии"]) to struct.
//
// To unmarshal JSON into a value implementing the Unmarshaler interface,
// Unmarshal calls that value's UnmarshalJSON method.
// See more https://golang.org/pkg/encoding/json/#Unmarshal
func (g *GroupsSectionsList) UnmarshalJSON(data []byte) error {
	var alias []interface{}
	if err := json.Unmarshal(data, &alias); err != nil {
		return err
	}

	if len(alias) != 2 {
		return &json.UnmarshalTypeError{
			Value: string(data),
			Type:  reflect.TypeOf((*GroupsSectionsList)(nil)),
		}
	}

	// default concrete Go type float64 for JSON numbers
	id, ok := alias[0].(float64)
	if !ok {
		return &json.UnmarshalTypeError{
			Value:  string(data),
			Type:   reflect.TypeOf((*GroupsSectionsList)(nil)),
			Struct: "GroupsSectionsList",
			Field:  "ID",
		}
	}

	name, ok := alias[1].(string)
	if !ok {
		return &json.UnmarshalTypeError{
			Value:  string(data),
			Type:   reflect.TypeOf((*GroupsSectionsList)(nil)),
			Struct: "GroupsSectionsList",
			Field:  "Name",
		}
	}

	g.ID = int(id)
	g.Name = name

	return nil
}

// GroupsActionType for action_button in groups.
type GroupsActionType string

// GroupsActionType enums.
const (
	GroupsActionTypeOpenURL      GroupsActionType = "open_url"
	GroupsActionTypeSendEmail    GroupsActionType = "send_email"
	GroupsActionTypeCallPhone    GroupsActionType = "call_phone"
	GroupsActionTypeCallVK       GroupsActionType = "call_vk"
	GroupsActionTypeOpenGroupApp GroupsActionType = "open_group_app"
	GroupsActionTypeOpenApp      GroupsActionType = "open_app"
)

// GroupsActionButton struct.
type GroupsActionButton struct {
	ActionType GroupsActionType         `json:"action_type"`
	Target     GroupsActionButtonTarget `json:"target"`
	Title      string                   `json:"title"`

	// IsEnabled for GroupsGroupSettings
	IsEnabled BaseBoolInt `json:"is_enabled,omitempty"`
}

// GroupsActionButtonTarget struct.
type GroupsActionButtonTarget struct {
	// ActionType == ActionTypeSendEmail
	Email string `json:"email"`

	// ActionType == ActionTypeCallPhone
	Phone string `json:"phone"`

	// ActionType == ActionTypeCallVK
	UserID int `json:"user_id"`

	// ActionType == ActionTypeOpenURL
	URL string `json:"url"`

	// ActionType == ActionTypeOpenApp
	GoogleStoreURL string `json:"google_store_url"`
	ItunesURL      string `json:"itunes_url"`
	// URL string `json:"url"`

	// ActionType == ActionTypeOpenGroupApp
	AppID int `json:"app_id"`

	IsInternal BaseBoolInt `json:"is_internal"`
}

// GroupsGroupXtrInvitedBy struct.
type GroupsGroupXtrInvitedBy struct {
	AdminLevel   int         `json:"admin_level"`
	ID           int         `json:"id"`          // Community ID
	InvitedBy    int         `json:"invited_by"`  // Inviter ID
	Name         string      `json:"name"`        // Community name
	Photo100     string      `json:"photo_100"`   // URL of square photo of the community with 100 pixels in width
	Photo200     string      `json:"photo_200"`   // URL of square photo of the community with 200 pixels in width
	Photo50      string      `json:"photo_50"`    // URL of square photo of the community with 50 pixels in width
	ScreenName   string      `json:"screen_name"` // Domain of the community page
	Type         string      `json:"type"`
	IsClosed     int         `json:"is_closed"`     // Information whether community is closed
	IsAdmin      BaseBoolInt `json:"is_admin"`      // Information whether current user is manager
	IsMember     BaseBoolInt `json:"is_member"`     // Information whether current user is member
	IsAdvertiser BaseBoolInt `json:"is_advertiser"` // Information whether current user is advertiser
}

// ToMention return mention.
func (group GroupsGroupXtrInvitedBy) ToMention() string {
	return fmt.Sprintf("[club%d|%s]", group.ID, group.Name)
}

// GroupsLinksItem struct.
type GroupsLinksItem struct {
	Desc      string      `json:"desc"`       // Link description
	EditTitle BaseBoolInt `json:"edit_title"` // Information whether the link title can be edited
	ID        int         `json:"id"`         // Link ID
	Name      string      `json:"name"`       // Link title
	Photo100  string      `json:"photo_100"`  // URL of square image of the link with 100 pixels in width
	Photo50   string      `json:"photo_50"`   // URL of square image of the link with 50 pixels in width
	URL       string      `json:"url"`        // Link URL
}

// GroupsLongPollEvents struct.
type GroupsLongPollEvents struct {
	MessageNew                    BaseBoolInt `json:"message_new"`
	MessageReply                  BaseBoolInt `json:"message_reply"`
	PhotoNew                      BaseBoolInt `json:"photo_new"`
	AudioNew                      BaseBoolInt `json:"audio_new"`
	VideoNew                      BaseBoolInt `json:"video_new"`
	WallReplyNew                  BaseBoolInt `json:"wall_reply_new"`
	WallReplyEdit                 BaseBoolInt `json:"wall_reply_edit"`
	WallReplyDelete               BaseBoolInt `json:"wall_reply_delete"`
	WallReplyRestore              BaseBoolInt `json:"wall_reply_restore"`
	WallPostNew                   BaseBoolInt `json:"wall_post_new"`
	BoardPostNew                  BaseBoolInt `json:"board_post_new"`
	BoardPostEdit                 BaseBoolInt `json:"board_post_edit"`
	BoardPostRestore              BaseBoolInt `json:"board_post_restore"`
	BoardPostDelete               BaseBoolInt `json:"board_post_delete"`
	PhotoCommentNew               BaseBoolInt `json:"photo_comment_new"`
	PhotoCommentEdit              BaseBoolInt `json:"photo_comment_edit"`
	PhotoCommentDelete            BaseBoolInt `json:"photo_comment_delete"`
	PhotoCommentRestore           BaseBoolInt `json:"photo_comment_restore"`
	VideoCommentNew               BaseBoolInt `json:"video_comment_new"`
	VideoCommentEdit              BaseBoolInt `json:"video_comment_edit"`
	VideoCommentDelete            BaseBoolInt `json:"video_comment_delete"`
	VideoCommentRestore           BaseBoolInt `json:"video_comment_restore"`
	MarketCommentNew              BaseBoolInt `json:"market_comment_new"`
	MarketCommentEdit             BaseBoolInt `json:"market_comment_edit"`
	MarketCommentDelete           BaseBoolInt `json:"market_comment_delete"`
	MarketCommentRestore          BaseBoolInt `json:"market_comment_restore"`
	MarketOrderNew                BaseBoolInt `json:"market_order_new"`
	MarketOrderEdit               BaseBoolInt `json:"market_order_edit"`
	PollVoteNew                   BaseBoolInt `json:"poll_vote_new"`
	GroupJoin                     BaseBoolInt `json:"group_join"`
	GroupLeave                    BaseBoolInt `json:"group_leave"`
	GroupChangeSettings           BaseBoolInt `json:"group_change_settings"`
	GroupChangePhoto              BaseBoolInt `json:"group_change_photo"`
	GroupOfficersEdit             BaseBoolInt `json:"group_officers_edit"`
	MessageAllow                  BaseBoolInt `json:"message_allow"`
	MessageDeny                   BaseBoolInt `json:"message_deny"`
	WallRepost                    BaseBoolInt `json:"wall_repost"`
	UserBlock                     BaseBoolInt `json:"user_block"`
	UserUnblock                   BaseBoolInt `json:"user_unblock"`
	MessageEdit                   BaseBoolInt `json:"message_edit"`
	MessageTypingState            BaseBoolInt `json:"message_typing_state"`
	LeadFormsNew                  BaseBoolInt `json:"lead_forms_new"`
	LikeAdd                       BaseBoolInt `json:"like_add"`
	LikeRemove                    BaseBoolInt `json:"like_remove"`
	VkpayTransaction              BaseBoolInt `json:"vkpay_transaction"`
	AppPayload                    BaseBoolInt `json:"app_payload"`
	MessageRead                   BaseBoolInt `json:"message_read"`
	MessageEvent                  BaseBoolInt `json:"message_event"`
	DonutSubscriptionCreate       BaseBoolInt `json:"donut_subscription_create"`
	DonutSubscriptionProlonged    BaseBoolInt `json:"donut_subscription_prolonged"`
	DonutSubscriptionExpired      BaseBoolInt `json:"donut_subscription_expired"`
	DonutSubscriptionCancelled    BaseBoolInt `json:"donut_subscription_cancelled"`
	DonutSubscriptionPriceChanged BaseBoolInt `json:"donut_subscription_price_changed"`
	DonutMoneyWithdraw            BaseBoolInt `json:"donut_money_withdraw"`
	DonutMoneyWithdrawError       BaseBoolInt `json:"donut_money_withdraw_error"`

	// Bugs
	// MessagesEdit  BaseBoolInt `json:"messages_edit"`
	// WallNew       BaseBoolInt `json:"wall_new"`
	// WallNewReply  BaseBoolInt `json:"wall_new_reply"`
	// WallEditReply BaseBoolInt `json:"wall_edit_reply"`
}

// GroupsLongPollServer struct.
type GroupsLongPollServer struct {
	Key    string `json:"key"`    // Long Poll key
	Server string `json:"server"` // Long Poll server address
	Ts     string `json:"ts"`     // Number of the last event
}

// TODO: func (g GroupsLongPollServer) GetURL() string {

// GroupsLongPollSettings struct.
type GroupsLongPollSettings struct {
	APIVersion string               `json:"api_version"` // API version used for the events
	Events     GroupsLongPollEvents `json:"events"`
	IsEnabled  BaseBoolInt          `json:"is_enabled"` // Shows whether Long Poll is enabled
}

// GroupsMarketType ...
type GroupsMarketType string

// Possible values.
const (
	GroupsMarketBasic    GroupsMarketType = "basic"
	GroupsMarketAdvanced GroupsMarketType = "advanced"
)

// GroupsMarketInfo struct.
type GroupsMarketInfo struct {
	// information about the type of store. Returned if the group includes
	// the "Products" section.
	Type            GroupsMarketType  `json:"type,omitempty"`
	ContactID       int               `json:"contact_id,omitempty"` // Contact person ID
	Currency        MarketCurrency    `json:"currency,omitempty"`
	CurrencyText    string            `json:"currency_text,omitempty"` // Currency name
	Enabled         BaseBoolInt       `json:"enabled"`                 // Information whether the market is enabled
	CommentsEnabled BaseBoolInt       `json:"comments_enabled,omitempty"`
	CanMessage      BaseBoolInt       `json:"can_message,omitempty"`
	MainAlbumID     int               `json:"main_album_id,omitempty"` // Main market album ID
	PriceMax        string            `json:"price_max,omitempty"`     // Maximum price
	PriceMin        string            `json:"price_min,omitempty"`     // Minimum price
	Wiki            PagesWikipageFull `json:"wiki,omitempty"`
	CityIDs         []int             `json:"city_ids"`
	CountryIDs      []int             `json:"country_ids,omitempty"`
}

// GroupsGroupRole Role type.
const (
	GroupsGroupRoleModerator     = "moderator"
	GroupsGroupRoleEditor        = "editor"
	GroupsGroupRoleAdministrator = "administrator"
	GroupsGroupRoleCreator       = "creator"
)

// GroupsMemberRole struct.
type GroupsMemberRole struct {
	ID          int      `json:"id"` // User ID
	Role        string   `json:"role"`
	Permissions []string `json:"permissions"`
}

// GroupsMemberRoleXtrUsersUser struct.
type GroupsMemberRoleXtrUsersUser struct {
	UsersUser
	Role        string   `json:"role"`
	Permissions []string `json:"permissions"`
}

// GroupsMemberStatus struct.
type GroupsMemberStatus struct {
	Member      BaseBoolInt `json:"member"`  // Information whether user is a member of the group
	UserID      int         `json:"user_id"` // User ID
	Permissions []string    `json:"permissions"`
}

// GroupsMemberStatusFull struct.
type GroupsMemberStatusFull struct {
	Invitation BaseBoolInt `json:"invitation"` // Information whether user has been invited to the group
	Member     BaseBoolInt `json:"member"`     // Information whether user is a member of the group
	Request    BaseBoolInt `json:"request"`    // Information whether user has send request to the group
	CanInvite  BaseBoolInt `json:"can_invite"` // Information whether user can be invite
	CanRecall  BaseBoolInt `json:"can_recall"` // Information whether user's invite to the group can be recalled
	UserID     int         `json:"user_id"`    // User ID
}

// GroupsOnlineStatus Status type.
const (
	GroupsOnlineStatusTypeNone       = "none"
	GroupsOnlineStatusTypeOnline     = "online"
	GroupsOnlineStatusTypeAnswerMark = "answer_mark"
)

// GroupsOnlineStatus struct.
type GroupsOnlineStatus struct {
	Minutes int    `json:"minutes"` // Estimated time of answer (for status = answer_mark)
	Status  string `json:"status"`
}

// GroupsOwnerXtrBanInfo struct.
type GroupsOwnerXtrBanInfo struct {
	BanInfo GroupsBanInfo `json:"ban_info"`
	Group   GroupsGroup   `json:"group"`
	Profile UsersUser     `json:"profile"`
	Type    string        `json:"type"`
}

// GroupsSubjectItem struct.
type GroupsSubjectItem struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// GroupsTokenPermissionSetting struct.
type GroupsTokenPermissionSetting struct {
	Name    string `json:"name"`
	Setting int    `json:"setting"`
}

// GroupsTokenPermissions struct.
type GroupsTokenPermissions struct {
	Mask        int                            `json:"mask"`
	Permissions []GroupsTokenPermissionSetting `json:"permissions"`
}

// GroupsTag struct.
type GroupsTag struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Color string `json:"color"`
}
