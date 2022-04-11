package object // import "github.com/SevereCloud/vksdk/v2/object"

// AccountNameRequest struct.
type AccountNameRequest struct {
	FirstName string `json:"first_name"` // First name in request
	ID        int    `json:"id"`         // Request ID needed to cancel the request
	LastName  string `json:"last_name"`  // Last name in request
	Status    string `json:"status"`
}

// AccountPushConversations struct.
type AccountPushConversations struct {
	Count int                             `json:"count"` // Items count
	Items []*AccountPushConversationsItem `json:"items"`
}

// AccountPushConversationsItem struct.
type AccountPushConversationsItem struct {
	DisabledUntil        int         `json:"disabled_until"` // Time until that notifications are disabled in seconds
	PeerID               int         `json:"peer_id"`        // Peer ID
	Sound                int         `json:"sound"`          // Information whether the sound are enabled
	DisabledMentions     BaseBoolInt `json:"disabled_mentions"`
	DisabledMassMentions BaseBoolInt `json:"disabled_mass_mentions"`
}

// AccountPushParams struct.
type AccountPushParams struct {
	AppRequest     []string `json:"app_request"`
	Birthday       []string `json:"birthday"`
	Chat           []string `json:"chat"`
	Comment        []string `json:"comment"`
	EventSoon      []string `json:"event_soon"`
	Friend         []string `json:"friend"`
	FriendAccepted []string `json:"friend_accepted"`
	FriendFound    []string `json:"friend_found"`
	GroupAccepted  []string `json:"group_accepted"`
	GroupInvite    []string `json:"group_invite"`
	Like           []string `json:"like"`
	Mention        []string `json:"mention"`
	Msg            []string `json:"msg"`
	NewPost        []string `json:"new_post"`
	PhotosTag      []string `json:"photos_tag"`
	Reply          []string `json:"reply"`
	Repost         []string `json:"repost"`
	SdkOpen        []string `json:"sdk_open"`
	WallPost       []string `json:"wall_post"`
	WallPublish    []string `json:"wall_publish"`
}

// AccountOffer struct.
type AccountOffer struct {
	Description      string `json:"description"`       // Offer description
	ID               int    `json:"id"`                // Offer ID
	Img              string `json:"img"`               // URL of the preview image
	Instruction      string `json:"instruction"`       // Instruction how to process the offer
	InstructionHTML  string `json:"instruction_html"`  // Instruction how to process the offer (HTML format)
	Price            int    `json:"price"`             // Offer price
	ShortDescription string `json:"short_description"` // Offer short description
	Tag              string `json:"tag"`               // Offer tag
	Title            string `json:"title"`             // Offer title
}

// AccountAccountCounters struct.
type AccountAccountCounters struct {
	AppRequests              int `json:"app_requests"`            // New app requests number
	Events                   int `json:"events"`                  // New events number
	Friends                  int `json:"friends"`                 // New friends requests number
	FriendsRecommendations   int `json:"friends_recommendations"` // New friends recommendations number
	FriendsSuggestions       int `json:"friends_suggestions"`     // New friends suggestions number
	Gifts                    int `json:"gifts"`                   // New gifts number
	Groups                   int `json:"groups"`                  // New groups number
	Messages                 int `json:"messages"`                // New messages number
	Notifications            int `json:"notifications"`           // New notifications number
	Photos                   int `json:"photos"`                  // New photo tags number
	SDK                      int `json:"sdk"`                     // New SDK number
	MenuDiscoverBadge        int `json:"menu_discover_badge"`     // New menu discover badge number
	MenuClipsBadge           int `json:"menu_clips_badge"`        // New menu clips badge number
	Videos                   int `json:"videos"`                  // New video tags number
	Faves                    int `json:"faves"`                   // New faves number
	Calls                    int `json:"calls"`                   // New calls number
	MenuSuperappFriendsBadge int `json:"menu_superapp_friends_badge"`
	MenuNewClipsBadge        int `json:"menu_new_clips_badge"`
}

// AccountInfo struct.
type AccountInfo struct {
	// Country code.
	Country string `json:"country"`

	// Language ID.
	Lang int `json:"lang"`

	// Information whether HTTPS-only is enabled.
	HTTPSRequired BaseBoolInt `json:"https_required"`

	// Information whether user has been processed intro.
	Intro BaseBoolInt `json:"intro"`

	// Information whether wall comments should be hidden.
	NoWallReplies BaseBoolInt `json:"no_wall_replies"`

	// Information whether only owners posts should be shown.
	OwnPostsDefault BaseBoolInt `json:"own_posts_default"`

	// Two factor authentication is enabled.
	TwoFactorRequired         BaseBoolInt       `json:"2fa_required"`
	EuUser                    BaseBoolInt       `json:"eu_user"`
	CommunityComments         BaseBoolInt       `json:"community_comments"`
	IsLiveStreamingEnabled    BaseBoolInt       `json:"is_live_streaming_enabled"`
	IsNewLiveStreamingEnabled BaseBoolInt       `json:"is_new_live_streaming_enabled"`
	LinkRedirects             map[string]string `json:"link_redirects"`
	VkPayEndpointV2           string            `json:"vk_pay_endpoint_v2"`
}

// AccountPushSettings struct.
type AccountPushSettings struct {
	Conversations AccountPushConversations `json:"conversations"`

	// Information whether notifications are disabled.
	Disabled BaseBoolInt `json:"disabled"`

	// Time until that notifications are disabled in Unixtime.
	DisabledUntil int               `json:"disabled_until"`
	Settings      AccountPushParams `json:"settings"`
}

// AccountUserSettings struct.
type AccountUserSettings struct {
	Bdate            string             `json:"bdate"`            // User's date of birth
	BdateVisibility  int                `json:"bdate_visibility"` // Information whether user's birthdate are hidden
	City             BaseObject         `json:"city"`
	Country          BaseCountry        `json:"country"`
	FirstName        string             `json:"first_name"`  // User first name
	HomeTown         string             `json:"home_town"`   // User's hometown
	LastName         string             `json:"last_name"`   // User last name
	MaidenName       string             `json:"maiden_name"` // User maiden name
	NameRequest      AccountNameRequest `json:"name_request"`
	Phone            string             `json:"phone"`    // User phone number with some hidden digits
	Relation         int                `json:"relation"` // User relationship status
	RelationPartner  UsersUserMin       `json:"relation_partner"`
	RelationPending  BaseBoolInt        `json:"relation_pending"` // Information whether relation status is pending
	RelationRequests []UsersUserMin     `json:"relation_requests"`
	ScreenName       string             `json:"screen_name"` // Domain name of the user's page
	Sex              int                `json:"sex"`         // User sex
	Status           string             `json:"status"`      // User status
	ID               int                `json:"id"`          // TODO: Check it https://vk.com/bug230405 (always return 0)
}
