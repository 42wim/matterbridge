package object // import "github.com/SevereCloud/vksdk/v2/object"

// AdsAccesses struct.
type AdsAccesses struct {
	ClientID string `json:"client_id"`
	Role     string `json:"role"`
}

// AdsAccount struct.
type AdsAccount struct {
	AccessRole                  string      `json:"access_role"`
	AccountID                   int         `json:"account_id"` // Account ID
	AccountName                 string      `json:"account_name"`
	AccountStatus               BaseBoolInt `json:"account_status"` // Information whether account is active
	CanViewBudget               BaseBoolInt `json:"can_view_budget"`
	AdNetworkAllowedPotentially BaseBoolInt `json:"ad_network_allowed_potentially"`
	AccountType                 string      `json:"account_type"`
}

// AdsAdLayout struct.
type AdsAdLayout struct {
	AdFormat       interface{} `json:"ad_format"`    // Ad format
	Description    string      `json:"description"`  // Ad description
	ImageSrc       string      `json:"image_src"`    // Image URL
	ImageSrc2x     string      `json:"image_src_2x"` // URL of the preview image in double size
	LinkDomain     string      `json:"link_domain"`  // Domain of advertised object
	LinkURL        string      `json:"link_url"`     // URL of advertised object
	PreviewLink    string      `json:"preview_link"` // preview an ad as it is shown on the website
	Title          string      `json:"title"`        // Ad title
	Video          BaseBoolInt `json:"video"`        // Information whether the ad is a video
	ID             string      `json:"id"`
	CampaignID     int         `json:"campaign_id"`
	GoalType       int         `json:"goal_type"`
	CostType       int         `json:"cost_type"`
	AgeRestriction string      `json:"age_restriction"`
	LinkType       string      `json:"link_type"`
}

// AdsCampaign struct.
type AdsCampaign struct {
	AllLimit  string `json:"all_limit"`  // Campaign's total limit, rubles
	DayLimit  string `json:"day_limit"`  // Campaign's day limit, rubles
	ID        int    `json:"id"`         // Campaign ID
	Name      string `json:"name"`       // Campaign title
	StartTime int    `json:"start_time"` // Campaign start time, as Unixtime
	Status    int    `json:"status"`
	StopTime  int    `json:"stop_time"` // Campaign stop time, as Unixtime
	Type      string `json:"type"`
}

// AdsCategory struct.
type AdsCategory struct {
	ID            int                  `json:"id"`   // Category ID
	Name          string               `json:"name"` // Category name
	Subcategories []BaseObjectWithName `json:"subcategories"`
}

// AdsClient struct.
type AdsClient struct {
	AllLimit string `json:"all_limit"` // Client's total limit, rubles
	DayLimit string `json:"day_limit"` // Client's day limit, rubles
	ID       int    `json:"id"`        // Client ID
	Name     string `json:"name"`      // Client name
}

// AdsCriteria struct.
type AdsCriteria struct {
	AgeFrom            int    `json:"age_from"`            // Age from
	AgeTo              int    `json:"age_to"`              // Age to
	Apps               string `json:"apps"`                // Apps IDs
	AppsNot            string `json:"apps_not"`            // Apps IDs to except
	Birthday           int    `json:"birthday"`            // Days to birthday
	Cities             string `json:"cities"`              // Cities IDs
	CitiesNot          string `json:"cities_not"`          // Cities IDs to except
	Country            int    `json:"country"`             // Country ID
	Districts          string `json:"districts"`           // Districts IDs
	Groups             string `json:"groups"`              // Communities IDs
	InterestCategories string `json:"interest_categories"` // Interests categories IDs
	Interests          string `json:"interests"`           // Interests

	// Information whether the user has proceeded VK payments before.
	Paying               BaseBoolInt `json:"paying"`
	Positions            string      `json:"positions"`              // Positions IDs
	Religions            string      `json:"religions"`              // Religions IDs
	RetargetingGroups    string      `json:"retargeting_groups"`     // Retargeting groups IDs
	RetargetingGroupsNot string      `json:"retargeting_groups_not"` // Retargeting groups IDs to except
	SchoolFrom           int         `json:"school_from"`            // School graduation year from
	SchoolTo             int         `json:"school_to"`              // School graduation year to
	Schools              string      `json:"schools"`                // Schools IDs
	Sex                  int         `json:"sex"`
	Stations             string      `json:"stations"`      // Stations IDs
	Statuses             string      `json:"statuses"`      // Relationship statuses
	Streets              string      `json:"streets"`       // Streets IDs
	Travellers           int         `json:"travellers"`    // Travellers only
	UniFrom              int         `json:"uni_from"`      // University graduation year from
	UniTo                int         `json:"uni_to"`        // University graduation year to
	UserBrowsers         string      `json:"user_browsers"` // Browsers
	UserDevices          string      `json:"user_devices"`  // Devices
	UserOs               string      `json:"user_os"`       // Operating systems
}

// AdsDemoStats struct.
type AdsDemoStats struct {
	ID    int                `json:"id"` // Object ID
	Stats AdsDemostatsFormat `json:"stats"`
	Type  string             `json:"type"`
}

// AdsDemostatsFormat struct.
type AdsDemostatsFormat struct {
	Age     []AdsStatsAge    `json:"age"`
	Cities  []AdsStatsCities `json:"cities"`
	Day     string           `json:"day"`     // Day as YYYY-MM-DD
	Month   string           `json:"month"`   // Month as YYYY-MM
	Overall int              `json:"overall"` // 1 if period=overall
	Sex     []AdsStatsSex    `json:"sex"`
	SexAge  []AdsStatsSexAge `json:"sex_age"`
}

// AdsFloodStats struct.
type AdsFloodStats struct {
	Left    int `json:"left"`    // Requests left
	Refresh int `json:"refresh"` // Time to refresh in seconds
}

// AdsLinkStatus link status.
type AdsLinkStatus string

// Possible values.
const (
	// allowed to use in ads.
	AdsLinkAllowed AdsLinkStatus = "allowed"

	// prohibited to use for this type of the object.
	AdsLinkDisallowed AdsLinkStatus = "disallowed"

	// checking, wait please.
	AdsLinkInProgress AdsLinkStatus = "in_progress"
)

// AdsParagraphs struct.
type AdsParagraphs struct {
	Paragraph string `json:"paragraph"` // Rules paragraph
}

// AdsRejectReason struct.
type AdsRejectReason struct {
	Comment string     `json:"comment"` // Comment text
	Rules   []AdsRules `json:"rules"`
}

// AdsRules struct.
type AdsRules struct {
	Paragraphs []AdsParagraphs `json:"paragraphs"`
	Title      string          `json:"title"` // Comment
}

// AdsStats struct.
type AdsStats struct {
	ID    int            `json:"id"` // Object ID
	Stats AdsStatsFormat `json:"stats"`
	Type  string         `json:"type"`
}

// AdsStatsAge struct.
type AdsStatsAge struct {
	ClicksRate      float64 `json:"clicks_rate"`      // Clicks rate
	ImpressionsRate float64 `json:"impressions_rate"` // Impressions rate
	Value           string  `json:"value"`            // Age interval
}

// AdsStatsCities struct.
type AdsStatsCities struct {
	ClicksRate      float64 `json:"clicks_rate"`      // Clicks rate
	ImpressionsRate float64 `json:"impressions_rate"` // Impressions rate
	Name            string  `json:"name"`             // City name
	Value           int     `json:"value"`            // City ID
}

// AdsStatsFormat struct.
type AdsStatsFormat struct {
	Clicks          int    `json:"clicks"`            // Clicks number
	Day             string `json:"day"`               // Day as YYYY-MM-DD
	Impressions     int    `json:"impressions"`       // Impressions number
	JoinRate        int    `json:"join_rate"`         // Events number
	Month           string `json:"month"`             // Month as YYYY-MM
	Overall         int    `json:"overall"`           // 1 if period=overall
	Reach           int    `json:"reach"`             // Reach
	Spent           int    `json:"spent"`             // Spent funds
	VideoClicksSite int    `json:"video_clicks_site"` // Click-thoughts to the advertised site
	VideoViews      int    `json:"video_views"`       // Video views number
	VideoViewsFull  int    `json:"video_views_full"`  // Video views (full video)
	VideoViewsHalf  int    `json:"video_views_half"`  // Video views (half of video)
}

// AdsStatsSex struct.
type AdsStatsSex struct {
	ClicksRate      float64 `json:"clicks_rate"`      // Clicks rate
	ImpressionsRate float64 `json:"impressions_rate"` // Impressions rate
	Value           string  `json:"value"`
}

// AdsStatsSexAge struct.
type AdsStatsSexAge struct {
	ClicksRate      float64 `json:"clicks_rate"`      // Clicks rate
	ImpressionsRate float64 `json:"impressions_rate"` // Impressions rate
	Value           string  `json:"value"`            // Sex and age interval
}

// AdsTargSettings struct.
type AdsTargSettings struct{}

// AdsTargStats struct.
type AdsTargStats struct {
	AudienceCount  int     `json:"audience_count"`  // Audience
	RecommendedCpc float64 `json:"recommended_cpc"` // Recommended CPC value
	RecommendedCpm float64 `json:"recommended_cpm"` // Recommended CPM value
}

// AdsTargSuggestions struct.
type AdsTargSuggestions struct {
	ID   int    `json:"id"`   // Object ID
	Name string `json:"name"` // Object name
}

// AdsTargSuggestionsCities struct.
type AdsTargSuggestionsCities struct {
	ID     int    `json:"id"`     // Object ID
	Name   string `json:"name"`   // Object name
	Parent string `json:"parent"` // Parent object
}

// AdsTargSuggestionsRegions struct.
type AdsTargSuggestionsRegions struct {
	ID   int    `json:"id"`   // Object ID
	Name string `json:"name"` // Object name
	Type string `json:"type"` // Object type
}

// AdsTargSuggestionsSchools struct.
type AdsTargSuggestionsSchools struct {
	Desc   string `json:"desc"`   // Full school title
	ID     int    `json:"id"`     // School ID
	Name   string `json:"name"`   // School title
	Parent string `json:"parent"` // City name
	Type   string `json:"type"`
}

// AdsTargetGroup struct.
type AdsTargetGroup struct {
	AudienceCount   int         `json:"audience_count"` // Audience
	ID              int         `json:"id"`             // Group ID
	Lifetime        int         `json:"lifetime"`       // Number of days for user to be in group
	Name            string      `json:"name"`           // Group name
	LastUpdated     int         `json:"last_updated"`
	IsAudience      BaseBoolInt `json:"is_audience"`
	IsShared        BaseBoolInt `json:"is_shared"`
	FileSource      BaseBoolInt `json:"file_source"`
	APISource       BaseBoolInt `json:"api_source"`
	LookalikeSource BaseBoolInt `json:"lookalike_source"`
	Domain          string      `json:"domain,omitempty"` // Site domain
	Pixel           string      `json:"pixel,omitempty"`  // Pixel code
}

// AdsUsers struct.
type AdsUsers struct {
	Accesses []AdsAccesses `json:"accesses"`
	UserID   int           `json:"user_id"` // User ID
}

// AdsAd struct.
type AdsAd struct {
	Approved                string      `json:"approved"`
	AllLimit                string      `json:"all_limit"`
	Category1ID             string      `json:"category1_id"`
	Category2ID             string      `json:"category2_id"`
	Cpm                     string      `json:"cpm"`
	AdFormat                int         `json:"ad_format"`   // Ad format
	AdPlatform              interface{} `json:"ad_platform"` // Ad platform
	CampaignID              int         `json:"campaign_id"` // Campaign ID
	CostType                int         `json:"cost_type"`
	Cpc                     int         `json:"cpc"`                    // Cost of a click, kopecks
	DisclaimerMedical       BaseBoolInt `json:"disclaimer_medical"`     // Information whether disclaimer is enabled
	DisclaimerSpecialist    BaseBoolInt `json:"disclaimer_specialist"`  // Information whether disclaimer is enabled
	DisclaimerSupplements   BaseBoolInt `json:"disclaimer_supplements"` // Information whether disclaimer is enabled
	Video                   BaseBoolInt `json:"video"`                  // Information whether the ad is a video
	ImpressionsLimited      BaseBoolInt `json:"impressions_limited"`    // Information whether impressions are limited
	Autobidding             BaseBoolInt `json:"autobidding"`
	ImpressionsLimit        int         `json:"impressions_limit"` // Impressions limit
	ID                      string      `json:"id"`                // Ad ID
	Name                    string      `json:"name"`              // Ad title
	Status                  int         `json:"status"`
	CreateTime              string      `json:"create_time"`
	UpdateTime              string      `json:"update_time"`
	GoalType                int         `json:"goal_type"`
	DayLimit                string      `json:"day_limit"`
	StartTime               string      `json:"start_time"`
	StopTime                string      `json:"stop_time"`
	AgeRestriction          string      `json:"age_restriction"`
	EventsRetargetingGroups interface{} `json:"events_retargeting_groups"`
	ImpressionsLimitPeriod  string      `json:"impressions_limit_period"`
}

// AdsPromotedPostReach struct.
type AdsPromotedPostReach struct {
	Hide             int `json:"hide"`              // Hides amount
	ID               int `json:"id"`                // Object ID from 'ids' parameter
	JoinGroup        int `json:"join_group"`        // Community joins
	Links            int `json:"links"`             // Link clicks
	ReachSubscribers int `json:"reach_subscribers"` // Subscribers reach
	ReachTotal       int `json:"reach_total"`       // Total reach
	Report           int `json:"report"`            // Reports amount
	ToGroup          int `json:"to_group"`          // Community clicks
	Unsubscribe      int `json:"unsubscribe"`       // 'Unsubscribe' events amount
	VideoViews100p   int `json:"video_views_100p"`  // Video views for 100 percent
	VideoViews25p    int `json:"video_views_25p"`   // Video views for 25 percent
	VideoViews3s     int `json:"video_views_3s"`    // Video views for 3 seconds
	VideoViews50p    int `json:"video_views_50p"`   // Video views for 50 percent
	VideoViews75p    int `json:"video_views_75p"`   // Video views for 75 percent
	VideoViewsStart  int `json:"video_views_start"` // Video starts
}

// AdsMusician struct.
type AdsMusician struct {
	ID     int    `json:"id"`               // Targeting music artist ID
	Name   string `json:"name"`             // Music artist name
	Avatar string `json:"avatar,omitempty"` // Music artist photo.
}
