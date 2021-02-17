package object // import "github.com/SevereCloud/vksdk/v2/object"

// UtilsDomainResolvedType object type.
const (
	UtilsDomainResolvedTypeUser        = "user"
	UtilsDomainResolvedTypeGroup       = "group"
	UtilsDomainResolvedTypeApplication = "application"
	UtilsDomainResolvedTypePage        = "page"
	UtilsDomainResolvedTypeVkApp       = "vk_app"
)

// UtilsDomainResolved struct.
type UtilsDomainResolved struct {
	ObjectID int    `json:"object_id"` // Object ID
	Type     string `json:"type"`
}

// UtilsLastShortenedLink struct.
type UtilsLastShortenedLink struct {
	AccessKey string `json:"access_key"` // Access key for private stats
	Key       string `json:"key"`        // Link key (characters after vk.cc/)
	ShortURL  string `json:"short_url"`  // Short link URL
	Timestamp int    `json:"timestamp"`  // Creation time in Unixtime
	URL       string `json:"url"`        // Full URL
	Views     int    `json:"views"`      // Total views number
}

// Link status.
const (
	UtilsLinkCheckedStatusNotBanned  = "not_banned"
	UtilsLinkCheckedStatusBanned     = "banned"
	UtilsLinkCheckedStatusProcessing = "processing"
)

// UtilsLinkChecked struct.
type UtilsLinkChecked struct {
	Link   string `json:"link"` // Link URL
	Status string `json:"status"`
}

// UtilsLinkStats struct.
type UtilsLinkStats struct {
	Key   string       `json:"key"` // Link key (characters after vk.cc/)
	Stats []UtilsStats `json:"stats"`
}

// UtilsLinkStatsExtended struct.
type UtilsLinkStatsExtended struct {
	Key   string               `json:"key"` // Link key (characters after vk.cc/)
	Stats []UtilsStatsExtended `json:"stats"`
}

// UtilsShortLink struct.
type UtilsShortLink struct {
	AccessKey string `json:"access_key"` // Access key for private stats
	Key       string `json:"key"`        // Link key (characters after vk.cc/)
	ShortURL  string `json:"short_url"`  // Short link URL
	URL       string `json:"url"`        // Full URL
}

// UtilsStats struct.
type UtilsStats struct {
	Timestamp int `json:"timestamp"` // Start time
	Views     int `json:"views"`     // Total views number
}

// UtilsStatsCity struct.
type UtilsStatsCity struct {
	CityID int `json:"city_id"` // City ID
	Views  int `json:"views"`   // Views number
}

// UtilsStatsCountry struct.
type UtilsStatsCountry struct {
	CountryID int `json:"country_id"` // Country ID
	Views     int `json:"views"`      // Views number
}

// UtilsStatsExtended struct.
type UtilsStatsExtended struct {
	Cities    []UtilsStatsCity    `json:"cities"`
	Countries []UtilsStatsCountry `json:"countries"`
	SexAge    []UtilsStatsSexAge  `json:"sex_age"`
	Timestamp int                 `json:"timestamp"` // Start time
	Views     int                 `json:"views"`     // Total views number
}

// UtilsStatsSexAge struct.
type UtilsStatsSexAge struct {
	AgeRange string `json:"age_range"` // Age denotation
	Female   int    `json:"female"`    //  Views by female users
	Male     int    `json:"male"`      //  Views by male users
}
