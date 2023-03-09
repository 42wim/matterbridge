package object // import "github.com/SevereCloud/vksdk/v2/object"

// StatsActivity struct.
type StatsActivity struct {
	Comments     int `json:"comments"`     // Comments number
	Copies       int `json:"copies"`       // Reposts number
	Hidden       int `json:"hidden"`       // Hidden from news count
	Likes        int `json:"likes"`        // Likes number
	Subscribed   int `json:"subscribed"`   // New subscribers count
	Unsubscribed int `json:"unsubscribed"` // Unsubscribed count
}

// StatsCity struct.
type StatsCity struct {
	Count int    `json:"count"` // Visitors number
	Name  string `json:"name"`  // City name
	Value int    `json:"value"` // City ID
}

// StatsCountry struct.
type StatsCountry struct {
	Code  string `json:"code"`  // Country code
	Count int    `json:"count"` // Visitors number
	Name  string `json:"name"`  // Country name
	Value int    `json:"value"` // Country ID
}

// StatsPeriod struct.
type StatsPeriod struct {
	Activity   StatsActivity `json:"activity"`
	PeriodFrom int           `json:"period_from"` // Unix timestamp
	PeriodTo   int           `json:"period_to"`   // Unix timestamp
	Reach      StatsReach    `json:"reach"`
	Visitors   StatsViews    `json:"visitors"`
}

// StatsReach struct.
type StatsReach struct {
	Age              []StatsSexAge  `json:"age"`
	Cities           []StatsCity    `json:"cities"`
	Countries        []StatsCountry `json:"countries"`
	MobileReach      int            `json:"mobile_reach"`      // Reach count from mobile devices
	Reach            int            `json:"reach"`             // Reach count
	ReachSubscribers int            `json:"reach_subscribers"` // Subscribers reach count
	Sex              []StatsSexAge  `json:"sex"`
	SexAge           []StatsSexAge  `json:"sex_age"`
}

// StatsSexAge struct.
type StatsSexAge struct {
	Count int    `json:"count"` // Visitors number
	Value string `json:"value"` // Sex/age value
}

// StatsViews struct.
type StatsViews struct {
	Age         []StatsSexAge  `json:"age"`
	Cities      []StatsCity    `json:"cities"`
	Countries   []StatsCountry `json:"countries"`
	MobileViews int            `json:"mobile_views"` // Number of views from mobile devices
	Sex         []StatsSexAge  `json:"sex"`
	SexAge      []StatsSexAge  `json:"sex_age"`
	Views       int            `json:"views"`    // Views number
	Visitors    int            `json:"visitors"` // Visitors number
}

// StatsWallpostStat struct.
type StatsWallpostStat struct {
	PostID           int `json:"post_id"`
	Hide             int `json:"hide"`              // Hidings number
	JoinGroup        int `json:"join_group"`        // People have joined the group
	Links            int `json:"links"`             // Link click-through
	ReachSubscribers int `json:"reach_subscribers"` // Subscribers reach
	ReachTotal       int `json:"reach_total"`       // Total reach
	ReachViral       int `json:"reach_viral"`       // Viral reach
	ReachAds         int `json:"reach_ads"`         // Advertising reach
	Report           int `json:"report"`            // Reports number
	ToGroup          int `json:"to_group"`          // Click-through to community
	Unsubscribe      int `json:"unsubscribe"`       // Unsubscribed members
	AdViews          int `json:"ad_views"`
	AdSubscribers    int `json:"ad_subscribers"`
	AdHide           int `json:"ad_hide"`
	AdUnsubscribe    int `json:"ad_unsubscribe"`
	AdLinks          int `json:"ad_links"`
	AdToGroup        int `json:"ad_to_group"`
	AdJoinGroup      int `json:"ad_join_group"`
	AdCoverage       int `json:"ad_coverage"`
	AdReport         int `json:"ad_report"`
}
