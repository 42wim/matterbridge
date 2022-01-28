/*
Package object contains objects for VK.

See more https://vk.com/dev/objects
*/
package object // import "github.com/SevereCloud/vksdk/v2/object"

import (
	"bytes"
	"encoding/json"
	"reflect"

	"github.com/vmihailenco/msgpack/v5"
)

// Attachment interface.
type Attachment interface {
	ToAttachment() string
}

// JSONObject interface.
type JSONObject interface {
	ToJSON() string
}

// BaseBoolInt type.
type BaseBoolInt bool

// UnmarshalJSON func.
func (b *BaseBoolInt) UnmarshalJSON(data []byte) (err error) {
	switch {
	case bytes.Equal(data, []byte("1")), bytes.Equal(data, []byte("true")):
		*b = true
	case bytes.Equal(data, []byte("0")), bytes.Equal(data, []byte("false")):
		*b = false
	default:
		// return json error
		err = &json.UnmarshalTypeError{
			Value: string(data),
			Type:  reflect.TypeOf((*BaseBoolInt)(nil)),
		}
	}

	return
}

// DecodeMsgpack func.
func (b *BaseBoolInt) DecodeMsgpack(dec *msgpack.Decoder) (err error) {
	data, err := dec.DecodeRaw()
	if err != nil {
		return err
	}

	var (
		valueInt  int
		valueBool bool
	)

	switch {
	case msgpack.Unmarshal(data, &valueBool) == nil:
		*b = BaseBoolInt(valueBool)
	case msgpack.Unmarshal(data, &valueInt) == nil:
		if valueInt == 1 {
			*b = true
			break
		}

		if valueInt == 0 {
			*b = false
			break
		}

		fallthrough
	default:
		// return msgpack error
		err = &json.UnmarshalTypeError{
			Value: string(data),
			Type:  reflect.TypeOf((*BaseBoolInt)(nil)),
		}
	}

	return err
}

// BaseCountry struct.
type BaseCountry struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
}

// BaseObject struct.
type BaseObject struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
}

// BaseObjectCount struct.
type BaseObjectCount struct {
	Count int `json:"count"`
}

// BaseObjectWithName struct.
type BaseObjectWithName struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// BaseRequestParam struct.
type BaseRequestParam struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// BaseSex const.
const (
	SexUnknown = iota
	SexFemale
	SexMale
)

// LongPollResponse struct.
type LongPollResponse struct {
	Ts      int             `json:"ts"`
	Updates [][]interface{} `json:"updates"`
	Failed  int             `json:"failed"`
}

// BaseCommentsInfo struct.
type BaseCommentsInfo struct {
	Count         int         `json:"count"`
	CanPost       BaseBoolInt `json:"can_post"`
	GroupsCanPost BaseBoolInt `json:"groups_can_post"`
	CanClose      BaseBoolInt `json:"can_close"`
	CanOpen       BaseBoolInt `json:"can_open"`
}

// BaseGeo struct.
type BaseGeo struct {
	Coordinates string    `json:"coordinates"`
	Place       BasePlace `json:"place"`
	Showmap     int       `json:"showmap"`
	Type        string    `json:"type"`
}

// BaseMessageGeo struct.
type BaseMessageGeo struct {
	Coordinates BaseGeoCoordinates `json:"coordinates"`
	Place       BasePlace          `json:"place"`
	Showmap     int                `json:"showmap"`
	Type        string             `json:"type"`
}

// BaseGeoCoordinates struct.
type BaseGeoCoordinates struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

// BaseImage struct.
type BaseImage struct {
	Height float64 `json:"height"`
	URL    string  `json:"url"`
	Width  float64 `json:"width"`
	Type   string  `json:"type"`
}

// UnmarshalJSON is required to support images with `src` field.
func (obj *BaseImage) UnmarshalJSON(data []byte) (err error) {
	type renamedBaseImage struct {
		Height float64 `json:"height"`
		URL    string  `json:"url"`
		Src    string  `json:"src"`
		Width  float64 `json:"width"`
		Type   string  `json:"type"`
	}

	var renamedObj renamedBaseImage

	err = json.Unmarshal(data, &renamedObj)

	obj.Height = renamedObj.Height
	obj.Width = renamedObj.Width
	obj.Type = renamedObj.Type

	if renamedObj.Src == "" {
		obj.URL = renamedObj.URL
	} else {
		obj.URL = renamedObj.Src
	}

	return err
}

// DecodeMsgpack is required to support images with `src` field.
func (obj *BaseImage) DecodeMsgpack(dec *msgpack.Decoder) (err error) {
	type renamedBaseImage struct {
		Height float64 `msgpack:"height"`
		URL    string  `msgpack:"url"`
		Src    string  `msgpack:"src"`
		Width  float64 `msgpack:"width"`
		Type   string  `msgpack:"type"`
	}

	var renamedObj renamedBaseImage

	err = dec.Decode(&renamedObj)

	obj.Height = renamedObj.Height
	obj.Width = renamedObj.Width
	obj.Type = renamedObj.Type

	if renamedObj.Src == "" {
		obj.URL = renamedObj.URL
	} else {
		obj.URL = renamedObj.Src
	}

	return err
}

// BaseLikes struct.
type BaseLikes struct {
	UserLikes BaseBoolInt `json:"user_likes"` // Information whether current user likes
	Count     int         `json:"count"`      // Likes number
}

// BaseLikesInfo struct.
type BaseLikesInfo struct {
	CanLike    BaseBoolInt `json:"can_like"`    // Information whether current user can like the post
	CanPublish BaseBoolInt `json:"can_publish"` // Information whether current user can repost
	UserLikes  BaseBoolInt `json:"user_likes"`  // Information whether current uer has liked the post
	Count      int         `json:"count"`       // Likes number
}

// BaseLink struct.
type BaseLink struct {
	Application  BaseLinkApplication `json:"application"`
	Button       BaseLinkButton      `json:"button"`
	ButtonText   string              `json:"button_text"`
	ButtonAction string              `json:"button_action"`
	Caption      string              `json:"caption"`
	Description  string              `json:"description"`
	Photo        PhotosPhoto         `json:"photo"`
	Video        VideoVideo          `json:"video"`
	PreviewPage  string              `json:"preview_page"`
	PreviewURL   string              `json:"preview_url"`
	Product      BaseLinkProduct     `json:"product"`
	Rating       BaseLinkRating      `json:"rating"`
	Title        string              `json:"title"`
	Target       string              `json:"target"`
	URL          string              `json:"url"`
	IsFavorite   BaseBoolInt         `json:"is_favorite"`
}

// BaseLinkApplication struct.
type BaseLinkApplication struct {
	AppID float64                  `json:"app_id"`
	Store BaseLinkApplicationStore `json:"store"`
}

// BaseLinkApplicationStore struct.
type BaseLinkApplicationStore struct {
	ID   float64 `json:"id"`
	Name string  `json:"name"`
}

// BaseLinkButton struct.
type BaseLinkButton struct {
	Action BaseLinkButtonAction `json:"action"`
	Title  string               `json:"title"`
}

// BaseLinkButtonAction struct.
type BaseLinkButtonAction struct {
	Type string `json:"type"`
	URL  string `json:"url"`
}

// BaseLinkProduct struct.
type BaseLinkProduct struct {
	Price       MarketPrice `json:"price"`
	Merchant    string      `json:"merchant"`
	OrdersCount int         `json:"orders_count"`
}

// BaseLinkRating struct.
type BaseLinkRating struct {
	ReviewsCount int     `json:"reviews_count"`
	Stars        float64 `json:"stars"`
}

// BasePlace struct.
type BasePlace struct {
	Address        string             `json:"address"`
	Checkins       int                `json:"checkins"`
	City           interface{}        `json:"city"` // BUG(VK): https://github.com/VKCOM/vk-api-schema/issues/143
	Country        interface{}        `json:"country"`
	Created        int                `json:"created"`
	ID             int                `json:"id"`
	Icon           string             `json:"icon"`
	Latitude       float64            `json:"latitude"`
	Longitude      float64            `json:"longitude"`
	Title          string             `json:"title"`
	Type           string             `json:"type"`
	IsDeleted      BaseBoolInt        `json:"is_deleted"`
	TotalCheckins  int                `json:"total_checkins"`
	Updated        int                `json:"updated"`
	CategoryObject BaseCategoryObject `json:"category_object"`
}

// BaseCategoryObject struct.
type BaseCategoryObject struct {
	ID    int         `json:"id"`
	Title string      `json:"title"`
	Icons []BaseImage `json:"icons"`
}

// BaseRepostsInfo struct.
type BaseRepostsInfo struct {
	Count        int `json:"count"`
	WallCount    int `json:"wall_count"`
	MailCount    int `json:"mail_count"`
	UserReposted int `json:"user_reposted"`
}

// BaseSticker struct.
type BaseSticker struct {
	Images               []BaseImage `json:"images"`
	ImagesWithBackground []BaseImage `json:"images_with_background"`
	ProductID            int         `json:"product_id"`
	StickerID            int         `json:"sticker_id"`
	IsAllowed            bool        `json:"is_allowed"`
	AnimationURL         string      `json:"animation_url"`
}

// MaxSize return the largest BaseSticker.
func (sticker BaseSticker) MaxSize() (maxImageSize BaseImage) {
	var max float64

	for _, imageSize := range sticker.Images {
		size := imageSize.Height * imageSize.Width
		if size > max {
			max = size
			maxImageSize = imageSize
		}
	}

	return
}

// MinSize return the smallest BaseSticker.
func (sticker BaseSticker) MinSize() (minImageSize BaseImage) {
	var min float64

	for _, imageSize := range sticker.Images {
		size := imageSize.Height * imageSize.Width
		if size < min || min == 0 {
			min = size
			minImageSize = imageSize
		}
	}

	return
}

// MaxSizeBackground return the largest BaseSticker with background.
func (sticker BaseSticker) MaxSizeBackground() (maxImageSize BaseImage) {
	var max float64

	for _, imageSize := range sticker.ImagesWithBackground {
		size := imageSize.Height * imageSize.Width
		if size > max {
			max = size
			maxImageSize = imageSize
		}
	}

	return
}

// MinSizeBackground return the smallest BaseSticker with background.
func (sticker BaseSticker) MinSizeBackground() (minImageSize BaseImage) {
	var min float64

	for _, imageSize := range sticker.ImagesWithBackground {
		size := imageSize.Height * imageSize.Width
		if size < min || min == 0 {
			min = size
			minImageSize = imageSize
		}
	}

	return
}

// BaseUserID struct.
type BaseUserID struct {
	UserID int `json:"user_id"`
}

// PrivacyCategory type.
type PrivacyCategory string

// Possible values.
const (
	PrivacyAll              PrivacyCategory = "all"
	PrivacyOnlyMe           PrivacyCategory = "only_me"
	PrivacyFriends          PrivacyCategory = "friends"
	PrivacyFriendsOfFriends PrivacyCategory = "friends_of_friends"
)

// Privacy struct.
type Privacy struct {
	Category PrivacyCategory `json:"category,omitempty"`
	Lists    struct {
		Allowed  []int `json:"allowed"`
		Excluded []int `json:"excluded"`
	} `json:"lists,omitempty"`
	Owners struct {
		Allowed  []int `json:"allowed"`
		Excluded []int `json:"excluded"`
	} `json:"owners,omitempty"`
}

// EventsEventAttach struct.
type EventsEventAttach struct {
	Address      string      `json:"address,omitempty"`       // address of event
	ButtonText   string      `json:"button_text"`             // text of attach
	Friends      []int       `json:"friends"`                 // array of friends ids
	ID           int         `json:"id"`                      // event ID
	IsFavorite   BaseBoolInt `json:"is_favorite"`             // is favorite
	MemberStatus int         `json:"member_status,omitempty"` // Current user's member status
	Text         string      `json:"text"`                    // text of attach
	Time         int         `json:"time,omitempty"`          // event start time
}

// OauthError struct.
type OauthError struct {
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
	RedirectURI      string `json:"redirect_uri"`
}

// Article struct.
type Article struct {
	ID            int         `json:"id"`
	OwnerID       int         `json:"owner_id"`
	OwnerName     string      `json:"owner_name"`
	OwnerPhoto    string      `json:"owner_photo"`
	State         string      `json:"state"`
	CanReport     BaseBoolInt `json:"can_report"`
	IsFavorite    BaseBoolInt `json:"is_favorite"`
	NoFooter      BaseBoolInt `json:"no_footer"`
	Title         string      `json:"title"`
	Subtitle      string      `json:"subtitle"`
	Views         int         `json:"views"`
	Shares        int         `json:"shares"`
	URL           string      `json:"url"`
	ViewURL       string      `json:"view_url"`
	AccessKey     string      `json:"access_key"`
	PublishedDate int         `json:"published_date"`
	Photo         PhotosPhoto `json:"photo"`
}

// ExtendedResponse struct.
type ExtendedResponse struct {
	Profiles []UsersUser   `json:"profiles,omitempty"`
	Groups   []GroupsGroup `json:"groups,omitempty"`
}

// ClientInfo struct.
type ClientInfo struct {
	ButtonActions  []string    `json:"button_actions"`
	Keyboard       BaseBoolInt `json:"keyboard"`
	InlineKeyboard BaseBoolInt `json:"inline_keyboard"`
	Carousel       BaseBoolInt `json:"carousel"`
	LangID         int         `json:"lang_id"`
}

// Language code.
const (
	LangRU               = 0   // Русский
	LangUK               = 1   // Українська
	LangBE               = 2   // Беларуская (тарашкевiца)
	LangEN               = 3   // English
	LangES               = 4   // Español
	LangFI               = 5   // Suomi
	LangDE               = 6   // Deutsch
	LangIT               = 7   // Italiano
	LangBG               = 8   // Български
	LangHR               = 9   // Hrvatski
	LangHU               = 10  // Magyar
	LangSR               = 11  // Српски
	LangPT               = 12  // Português
	LangEL               = 14  // Ελληνικά
	LangPL               = 15  // Polski
	LangFR               = 16  // Français
	LangKO               = 17  // 한국어
	LangZH               = 18  // 汉语
	LangLT               = 19  // Lietuvių
	LangJA               = 20  // 日本語
	LangCS               = 21  // Čeština
	LangET               = 22  // Eesti
	LangTT               = 50  // Татарча
	LangBA               = 51  // Башҡортса
	LangCV               = 52  // Чăвашла
	LangSK               = 53  // Slovenčina
	LangRO               = 54  // Română
	LangNO               = 55  // Norsk
	LangLV               = 56  // Latviešu
	LangAZ               = 57  // Azərbaycan dili
	LangHY               = 58  // Հայերեն
	LangSQ               = 59  // Shqip
	LangSV               = 60  // Svenska
	LangNL               = 61  // Nederlands
	LangTK               = 62  // Türkmen
	LangKA               = 63  // ქართული
	LangDA               = 64  // Dansk
	LangUZ               = 65  // O‘zbek
	LangMO               = 66  // Moldovenească
	LangBUA              = 67  // Буряад
	LangTH               = 68  // ภาษาไทย
	LangID               = 69  // Bahasa Indonesia
	LangTG               = 70  // Тоҷикӣ
	LangSL               = 71  // Slovenščina
	LangBS               = 72  // Bosanski
	LangPTBR             = 73  // Português brasileiro
	LangFA               = 74  // فارسی
	LangVI               = 75  // Tiếng Việt
	LangHI               = 76  // हिन्दी
	LangSI               = 77  // සිංහල
	LangBN               = 78  // বাংলা
	LangTL               = 79  // Tagalog
	LangMN               = 80  // Монгол
	LangMY               = 81  // ဗမာစာ
	LangTR               = 82  // Türkçe
	LangNE               = 83  // नेपाली
	LangUR               = 85  // اردو
	LangKY               = 87  // Кыргыз тили
	LangPA               = 90  // پنجابی
	LangOS               = 91  // Ирон
	LangKN               = 94  // ಕನ್ನಡ
	LangSW               = 95  // Kiswahili
	LangKK               = 97  // Қазақша
	LangAR               = 98  // العربية
	LangHE               = 99  // עברית
	LangPreRevolutionary = 100 // Дореволюцiонный
	LangMYV              = 101 // Эрзянь кель
	LangKDB              = 102 // Адыгэбзэ
	LangSAH              = 105 // Саха тыла
	LangADY              = 106 // Адыгабзэ
	LangUDM              = 107 // Удмурт
	LangCHM              = 108 // Марий йылме
	LangBE2              = 114 // Беларуская
	LangLEZ              = 118 // Лезги чІал
	LangTW               = 119 // 臺灣話
	LangKUM              = 236 // Къумукъ тил
	LangMVL              = 270 // Mirandés
	LangSLA              = 298 // Русинськый
	LangKRL              = 379 // Karjalan kieli
	LangTYV              = 344 // Тыва дыл
	LangXAL              = 357 // Хальмг келн
	LangTLY              = 373 // Tolışə zıvon
	LangKV               = 375 // Коми кыв
	LangUKClassic        = 452 // Українська (клясична)
	LangUKGalitska       = 454 // Українська (Галицка)
	LangKAB              = 457 // Taqbaylit
	LangEO               = 555 // Esperanto
	LangLA               = 666 // Lingua Latina
	LangSoviet           = 777 // Советский
)

// Button action type.
const (
	// A button that sends a message with text specified in the label.
	ButtonText = "text"

	// Opens the VK Pay window with predefined parameters. The button is called
	// “Pay with VK Pay” (VK Pay is displayed as a logo). This button always
	// stretches to the whole keyboard width.
	ButtonVKPay = "vkpay"

	// Opens a specified VK Apps app. This button always stretches to the whole
	// keyboard width.
	ButtonVKApp = "open_app"

	// Sends the location to the chat. This button always stretches to the
	// whole keyboard width.
	ButtonLocation = "location"

	// Opens the specified link.
	ButtonOpenLink = "open_link"

	// Allows, without sending a message from the user, to receive a
	// notification about pressing the button and perform the necessary action.
	ButtonCallback = "callback"
)

// Button color. This parameter is used only for buttons with the text and callback types.
const (
	Primary = "primary" // Blue button, indicates the main action. #5181B8
	ButtonBlue

	Secondary = "secondary" // Default white button. #FFFFFF
	ButtonWhite

	Negative = "negative" // Dangerous or negative action (cancel, delete etc.) #E64646
	ButtonRed

	Positive = "positive" // Accept, agree. #4BB34B
	ButtonGreen
)

// Platform content creation platform.
type Platform int

// Possible values.
const (
	_                    Platform = iota
	PlatformMobile                // mobile web version
	PlatformIPhone                // iPhone
	PlatformIPad                  // iPad
	PlatformAndroid               // Android
	PlatformWindowsPhone          // Windows Phone
	PlatformWindows               // Windows 8
	PlatformFull                  // full web version
	PlatformOther                 // other apps
)

// Conversations types.
const (
	PeerUser  = "user"
	PeerChat  = "chat"
	PeerGroup = "group"
	PeerEmail = "email"
)
